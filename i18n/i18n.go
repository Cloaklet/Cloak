//go:generate go run bundle.go

package i18n

import (
	"fmt"
)

var l localizer

// C is a channel through which the locale changes will be sent.
// Other parts of the application can receive locale changes from this channel and refresh their UI.
var C = make(chan string)

type localizer struct {
	data          map[string]map[string]string // language => {key => string}
	currentLocale string
}

// translate translates given key
func (l *localizer) translate(key string) string {
	// Fallback to en
	locale := l.currentLocale
	if locale == "" {
		locale = "en"
	}
	if locale, ok := l.data[locale]; ok {
		if str, ok := locale[key]; ok {
			return str
		}
	}
	return ""
}

// setLocale sets current language
func (l *localizer) setLocale(lang string) error {
	if _, ok := l.data[lang]; ok {
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
