package grpc

import (
	"github.com/bagasss3/toko/services/auth-service/internal/model"
	pb "github.com/bagasss3/toko/services/auth-service/pb/auth"
)

type AuthGRPCHandler struct {
	pb.UnimplementedAuthServiceServer
	authUsecase    model.AuthUsecase
	userUsecase    model.UserUsecase
	addressUsecase model.AddressUsecase
}

func NewAuthGRPCHandler(
	authUsecase model.AuthUsecase,
	userUsecase model.UserUsecase,
	addressUsecase model.AddressUsecase,
) *AuthGRPCHandler {
	return &AuthGRPCHandler{
		authUsecase:    authUsecase,
		userUsecase:    userUsecase,
		addressUsecase: addressUsecase,
	}
}

var _ pb.AuthServiceServer = (*AuthGRPCHandler)(nil)
