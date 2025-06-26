package middleware

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

// Define tipo de usuário
const (
	RoleAdmin    = "admin"
	RoleManager  = "manager"
	RoleEmployee = "employee"
)

// Middleware de autorização baseado em header "X-User-Role"
func RoleRequired(requiredRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role := c.Request().Header.Get("X-User-Role")
			role = strings.ToLower(role)

			for _, r := range requiredRoles {
				if role == r {
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "Access denied",
			})
		}
	}
}
