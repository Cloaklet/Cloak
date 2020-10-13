//go:generate go run bundler.go

package i18n

import (
	"fmt"
	"github.com/tidwall/gjson"
	"strings"
)

var l localizer

// C is a channel through which the locale changes will be sent.
// Other parts of the application can receive locale changes from this channel and refresh their UI.
var C = make(chan string)

type localizer struct {
	data          string
	currentLocale string
}

// translate translates given key
func (l *localizer) translate(key string) string {
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

// setLocale sets current language
func (l *localizer) setLocale(lang string) error {
	if gjson.Get(l.data, lang).Exists() {
		l.currentLocale = lang
		return nil
	}
	return fmt.Errorf("language %s not supported", lang)
}

// T translates given key
func T(key string) string {
	return l.translate(key)
}

// SetLocale sets current language
func SetLocale(lang string) error {
	err := l.setLocale(lang)
	if err == nil {
		C <- lang
	}
	return err
}

// GetCurrentLocale returns current effective locale
func GetCurrentLocale() string {
	if l.currentLocale != "" {
		return l.currentLocale
	}
	return "en"
}
