package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bagasss3/toko/services/gateway/internal/model"
)

// ErrorHandler returns a custom error handler for echo
func ErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		// Handle gRPC errors
		if st, ok := status.FromError(err); ok {
			handleGRPCError(st, c)
			return
		}

		// Handle echo HTTP errors
		if he, ok := err.(*echo.HTTPError); ok {
			handleHTTPError(he, c)
			return
		}

		// Handle our custom error response
		if er, ok := err.(model.ErrorResponse); ok {
			c.JSON(er.Status, er)
			return
		}

		// Default internal error
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Code:    model.ErrCodeInternal,
			Message: "Internal server error",
		})
	}
}

func handleGRPCError(st *status.Status, c echo.Context) {
	var statusCode int
	var code string
	message := st.Message()

	switch st.Code() {
	case codes.InvalidArgument:
		statusCode = http.StatusBadRequest
		code = model.ErrCodeInvalidRequest
	case codes.Unauthenticated:
		statusCode = http.StatusUnauthorized
		code = model.ErrCodeUnauthorized
	case codes.PermissionDenied:
		statusCode = http.StatusForbidden
		code = model.ErrCodeForbidden
	case codes.NotFound:
		statusCode = http.StatusNotFound
		code = model.ErrCodeNotFound
	case codes.AlreadyExists:
		statusCode = http.StatusConflict
		code = model.ErrCodeConflict
	case codes.FailedPrecondition:
		statusCode = http.StatusBadRequest
		code = model.ErrCodeValidation
	case codes.Unavailable:
		statusCode = http.StatusServiceUnavailable
		code = model.ErrCodeServiceUnavailable
	default:
		statusCode = http.StatusInternalServerError
		code = model.ErrCodeInternal
		message = "Internal server error"
	}

	c.JSON(statusCode, model.ErrorResponse{
		Status:  statusCode,
		Code:    code,
		Message: message,
	})
}

func handleHTTPError(he *echo.HTTPError, c echo.Context) {
	code := model.ErrCodeInternal
	message := he.Message.(string)

	switch he.Code {
	case http.StatusBadRequest:
		code = model.ErrCodeInvalidRequest
	case http.StatusUnauthorized:
		code = model.ErrCodeUnauthorized
	case http.StatusForbidden:
		code = model.ErrCodeForbidden
	case http.StatusNotFound:
		code = model.ErrCodeNotFound
	case http.StatusConflict:
		code = model.ErrCodeConflict
	case http.StatusUnprocessableEntity:
		code = model.ErrCodeValidation
	}

	c.JSON(he.Code, model.ErrorResponse{
		Status:  he.Code,
		Code:    code,
		Message: message,
	})
}
