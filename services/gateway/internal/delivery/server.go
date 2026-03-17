package delivery

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/bagasss3/toko/services/gateway/internal/client"
	"github.com/bagasss3/toko/services/gateway/internal/delivery/handler"
	"github.com/bagasss3/toko/services/gateway/internal/delivery/middleware"
)

type customValidator struct{ validator *validator.Validate }

func (cv *customValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}

type Server struct {
	echo       *echo.Echo
	authClient *client.AuthClient
}

func NewServer(authClient *client.AuthClient) *Server {
	e := echo.New()
	e.Validator = &customValidator{validator: validator.New()}
	e.HTTPErrorHandler = middleware.ErrorHandler()
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.RequestID())

	s := &Server{
		echo:       e,
		authClient: authClient,
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	authHandler := handler.NewAuthHandler(s.authClient)

	v1 := s.echo.Group("/api/v1")

	// public
	auth := v1.Group("/auth")
	auth.POST("/login", authHandler.Login)
	auth.POST("/register", authHandler.Register)

	// protected
	protected := v1.Group("")
	protected.Use(middleware.Auth(s.authClient))
	protected.GET("/me", authHandler.Me)
}

func (s *Server) Start(addr string) error {
	return s.echo.Start(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
