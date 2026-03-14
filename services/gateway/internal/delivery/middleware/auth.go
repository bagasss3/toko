package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/bagasss3/toko/services/gateway/internal/client"
	"github.com/bagasss3/toko/services/gateway/internal/model"
)

const (
	AuthorizationHeader = "Authorization"
	AuthorizationBearer = "Bearer"
	UserClaimsKey       = "user_claims"
)

func Auth(authClient *client.AuthClient) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get(AuthorizationHeader)
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != AuthorizationBearer {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			claims, err := authClient.ValidateToken(c.Request().Context(), parts[1])
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			c.Set(UserClaimsKey, claims)
			return next(c)
		}
	}
}

func GetClaims(c echo.Context) *model.TokenClaims {
	return c.Get(UserClaimsKey).(*model.TokenClaims)
}