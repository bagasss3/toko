package grpc

import (
	"context"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	apperrors "github.com/bagasss3/toko/packages/errors"
	pb "github.com/bagasss3/toko/services/auth-service/pb/auth"
)


func (h *AuthGRPCHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	tokens, err := h.authUsecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		log.WithError(err).Error("login failed")
		if err == apperrors.ErrInvalidCredentials {
			return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &pb.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (h *AuthGRPCHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	claims, err := h.authUsecase.ValidateToken(ctx, req.Token)
	if err != nil {
		log.WithError(err).Error("validate token failed")
		return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
	}

	return &pb.ValidateTokenResponse{
		UserId: claims.UserID.String(),
		Email:  claims.Email,
		Role:   string(claims.Role),
	}, nil
}
