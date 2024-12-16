package middleware

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

func LoggingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		fmt.Printf("Request method: %s, path: %s\n", c.Request().Method, c.Request().URL.Path)
		return next(c)
	}
}
