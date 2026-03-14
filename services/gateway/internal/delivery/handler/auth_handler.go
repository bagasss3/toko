package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/bagasss3/toko/services/gateway/internal/client"
	"github.com/bagasss3/toko/services/gateway/internal/delivery/middleware"
	"github.com/bagasss3/toko/services/gateway/internal/model"
)

type AuthHandler struct {
	authClient *client.AuthClient
}

func NewAuthHandler(authClient *client.AuthClient) *AuthHandler {
	return &AuthHandler{authClient: authClient}
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req model.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	accessToken, err := h.authClient.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	return c.JSON(http.StatusOK, model.LoginResponse{
		AccessToken: accessToken,
	})
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req model.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	claims, err := h.authClient.Register(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, model.UserResponse{
		UserID: claims.UserID,
		Email:  claims.Email,
	})
}

func (h *AuthHandler) Me(c echo.Context) error {
	claims := middleware.GetClaims(c)

	return c.JSON(http.StatusOK, model.UserResponse{
		UserID: claims.UserID,
		Email:  claims.Email,
	})
}
