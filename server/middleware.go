package server

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// CheckRuntimeDeps is a labstack/echo middleware.
// It checks for necessary external runtime dependencies, and block API requests if they are not all met.
func (s *ApiServer) CheckRuntimeDeps(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Always allow `GET /options` since it's the initial request sent by the UI.
		// TODO Need a more elegant way to do this
		if c.Path() == "/api/options" && c.Request().Method == http.MethodGet {
			return next(c)
		}

		if !s.fuseAvailable {
			return ErrMissingFuse
		}
		if s.cmd == "" {
			return ErrMissingGocryptfsBinary
		}
		// All check passes, call next handler
		return next(c)
	}
}
