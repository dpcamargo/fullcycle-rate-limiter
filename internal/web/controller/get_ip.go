package controller

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func GetIP(c echo.Context) error {
	req := c.Request()

	ip := req.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = strings.Split(req.RemoteAddr, ":")[0]
	} else {
		ip = strings.Split(ip, ",")[0]
	}
	return c.String(http.StatusOK, "Your IP address is: "+ip)
}
