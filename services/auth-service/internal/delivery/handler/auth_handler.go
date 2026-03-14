package grpc

import (
	"context"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/bagasss3/toko/pb/auth"
	"github.com/bagasss3/toko/services/auth-service/internal/model"
)

type AuthGRPCHandler struct {
	pb.UnimplementedAuthServiceServer
	authUsecase model.AuthUsecase
}

func NewAuthGRPCHandler(authUsecase model.AuthUsecase) *AuthGRPCHandler {
	return &AuthGRPCHandler{authUsecase: authUsecase}
}

func (h *AuthGRPCHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	tokens, err := h.authUsecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		log.WithError(err).Error("login failed")
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}

	return &pb.LoginResponse{
		AccessToken: tokens.AccessToken,
	}, nil
}

func (h *AuthGRPCHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, err := h.authUsecase.Register(ctx, req.Email, req.Password)
	if err != nil {
		log.WithError(err).Error("register failed")
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.RegisterResponse{
		Id:    user.ID,
		Email: user.Email,
	}, nil
}

func (h *AuthGRPCHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	claims, err := h.authUsecase.ValidateToken(ctx, req.Token)
	if err != nil {
		log.WithError(err).Error("validate token failed")
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}

	return &pb.ValidateTokenResponse{
		UserId: claims.UserID,
		Email:  claims.Email,
	}, nil
}
