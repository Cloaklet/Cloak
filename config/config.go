package config

import (
	"Cloak/extension"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"gopkg.in/ini.v1"
)

// Configurator is a type which allows setting Key-value config pairs and callbacks that reacts to them.
type Configurator struct {
	callbacks map[string]Callback
	data      map[string]string
	filePath  string
	ini       *ini.File
	rw        sync.RWMutex
}

// Callback is a type of function that accepts a config value.
type Callback func(v string) error

var logger zerolog.Logger

func init() {
	logger = extension.GetLogger("config")
}

// NewConfigurator creates a new Configurator instance.
func NewConfigurator(iniPath string) (*Configurator, error) {
	// Check parent directory of `iniPath`
	iniDirPath := filepath.Dir(iniPath)
	info, err := os.Stat(iniDirPath)
	if err != nil {
		logger.Warn().
			Err(err).
			Str("iniPath", iniPath).
			Msg("Failed to read directory of INI file")
		return nil, err
	}
	if !info.IsDir() {
		logger.Warn().
			Str("iniPath", iniPath).
			Str("iniDirPath", iniDirPath).
			Interface("info", info).
			Msg("Directory of INI file is not a real directory")
		return nil, fmt.Errorf("invalid ini file path")
	}

	// Check INI file path
	info, _ = os.Stat(iniPath)
	if info != nil && info.IsDir() {
		logger.Warn().
			Err(err).
			Str("iniPath", iniPath).
			Interface("info", info).
			Msg("INI file path is a directory")
		return nil, fmt.Errorf("invalid ini file path")
	}

	return &Configurator{
		callbacks: make(map[string]Callback),
		data:      make(map[string]string),
		filePath:  iniPath,
	}, nil
}

// SetCallbacks sets multiple config key callbacks in batch.
func (c *Configurator) SetCallbacks(cbs map[string]Callback) {
	c.rw.Lock()
	defer c.rw.Unlock()

	for k, fn := range cbs {
		c.callbacks[k] = fn
	}
}

// SetCallback sets a single config key callback.
func (c *Configurator) SetCallback(keyPath string, cb Callback) {
	c.rw.Lock()
	defer c.rw.Unlock()

	c.callbacks[keyPath] = cb
}

// Load loads configuration data from INI file.
func (c *Configurator) Load() error {
	c.rw.Lock()
	defer c.rw.Unlock()

	// Allow non-existing INI file path, assume empty config
	if _, err := os.Stat(c.filePath); err != nil && os.IsNotExist(err) {
		logger.Info().
			Str("filePath", c.filePath).
			Msg("INI file not exists, init empty config")
		if err := ini.Empty().SaveTo(c.filePath); err != nil {
			logger.Error().
				Err(err).
				Str("filePath", c.filePath).
				Msg("Failed to save empty INI file")
			return err
		}
	}

	iniFile, err := ini.Load(c.filePath)
	if err != nil {
		logger.Warn().
			Err(err).
			Str("filePath", c.filePath).
			Msg("Failed to load config from INI file")
		return err
	}

	DefaultSectionPrefix := fmt.Sprintf("%s.", ini.DefaultSection)
	c.ini = iniFile
	data := c.loadSections(ini.DefaultSection, c.ini.Sections())
	for k, v := range data {
		key := k
		value := v

		if strings.HasPrefix(key, DefaultSectionPrefix) {
			key = strings.TrimPrefix(key, DefaultSectionPrefix)
		}

		c.data[key] = value
		if cb, ok := c.callbacks[key]; ok {
			if err := cb(value); err != nil {
				logger.Warn().
					Err(err).
					Str("key", key).
					Str("value", value).
					Msg("Failed to call callback when loading settings key")
			}
		}
	}

	return nil
}

func (c *Configurator) loadKeys(sectionPath string, section *ini.Section) map[string]string {
	data := make(map[string]string)
	for _, key := range section.Keys() {
		data[fmt.Sprintf("%s.%s", sectionPath, key.Name())] = key.Value()
	}
	return data
}

func (c *Configurator) loadSections(sectionPath string, sections []*ini.Section) map[string]string {
	data := make(map[string]string)

	for _, section := range sections {
		sectionName := section.Name()
		var fullSectionPath string

		if sectionPath == ini.DefaultSection || sectionPath == "" {
			fullSectionPath = sectionName
		} else {
			fullSectionPath = fmt.Sprintf("%s.%s", sectionPath, sectionName)
		}
		for k, v := range c.loadKeys(fullSectionPath, section) {
			data[k] = v
		}

		for k, v := range c.loadSections(fullSectionPath, section.ChildSections()) {
			data[k] = v
		}
	}

	return data
}

func (c *Configurator) persistKey(key, value string) error {
	if key == "" {
		return fmt.Errorf("empty key not allowed")
	}

	keySegs := strings.Split(key, ".")
	var section *ini.Section

	if len(keySegs) == 1 {
		section = c.ini.Section(ini.DefaultSection)
	} else {
		section = c.ini.Section(keySegs[0])
		keySegs = keySegs[1:]
	}

	for i, seg := range keySegs {
		// Section
		if i < len(keySegs)-1 {
			section = c.ini.Section(seg)
		} else {
			// Key
			section.Key(seg).SetValue(value)
		}
	}

	return c.ini.SaveTo(c.filePath)
}

// Set sets a key-value pair.
func (c *Configurator) Set(key, value string) error {
	// Only update if value differs.
	if cv, ok := c.data[key]; ok && cv == value {
		return nil
	}

	c.rw.Lock()
	defer c.rw.Unlock()

	// TODO If persistKey failed, roll back changes.
	c.data[key] = value
	if err := c.persistKey(key, value); err != nil {
		logger.Warn().
			Err(err).
			Str("key", key).
			Str("value", value).
			Msg("Failed to persist settings after setting new key-value")
		return err
	}

	if cb, ok := c.callbacks[key]; ok {
		return cb(value)
	}

	return nil
}

// Get gets value gor given key.
func (c *Configurator) Get(key string) string {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return c.data[key]
}

// All returns current configuration KV pairs
func (c *Configurator) All() map[string]string {
	c.rw.RLock()
	defer c.rw.RUnlock()

	kvs := make(map[string]string)
	for k, v := range c.data {
		kvs[k] = v
	}
	return kvs
}
