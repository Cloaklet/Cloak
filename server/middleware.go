package server

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// CheckRuntimeDeps is a labstack/echo middleware.
// It checks for necessary external runtime dependencies, and block API requests if they are not all met.
func (s *ApiServer) CheckRuntimeDeps(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !s.fuseAvailable {
			return c.JSON(http.StatusOK, echo.Map{
				"code": ERR_CODE_MISSING_FUSE,
				"msg":  ERR_MSG_MISSING_FUSE,
			})
		}
		if s.cmd == "" {
			return c.JSON(http.StatusOK, echo.Map{
				"code": ERR_CODE_MISSING_GOCRYPTFS_BINARY,
				"msg":  ERR_MSG_MISSING_GOCRYPTFS_BINARY,
			})
		}
		// All check passes, call next handler
		return next(c)
	}
}
