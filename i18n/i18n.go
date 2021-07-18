package i18n

import (
	_ "embed"
	"fmt"
	"github.com/tidwall/gjson"
	"strings"
	"sync"
)

//go:embed locales.json
var data string

var l Localizer
var once sync.Once

func init() {
	if !gjson.Valid(data) {
		panic("malformed i18n data")
	}
}

// Localizer is a type which can translates JSON key path to string in given locale.
// It also features a channel through which locale change can be monitored.
type Localizer struct {
	data          string
	currentLocale string
	Ch            chan string
}

// GetLocalizer returns the global localizer (translator)
func GetLocalizer() *Localizer {
	once.Do(func() {
		l = Localizer{
			data:          data,
			currentLocale: "en",
			Ch:            make(chan string, 1),
		}
	})
	return &l
}

// T translates given key
func (l *Localizer) T(key string) string {
	// Fallback to en
	locale := l.currentLocale
	if locale == "" {
		locale = "en"
	}
	result := gjson.Get(l.data, strings.Join([]string{locale, key}, "."))
	if result.Exists() {
		return result.String()
	}
	return ""
}

// SetLocale sets current language
func (l *Localizer) SetLocale(lang string) error {
	if gjson.Get(l.data, lang).Exists() {
		l.currentLocale = lang
		l.Ch <- lang
		return nil
	}
	return fmt.Errorf("language %s not supported", lang)
}

// GetCurrentLocale returns current effective locale
func (l *Localizer) GetCurrentLocale() string {
	return l.currentLocale
}
