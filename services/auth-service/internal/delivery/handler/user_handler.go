package grpc

import (
	"context"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	apperrors "github.com/bagasss3/toko/packages/errors"
	"github.com/bagasss3/toko/services/auth-service/internal/model"
	pb "github.com/bagasss3/toko/services/auth-service/pb/auth"
)

func (h *AuthGRPCHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	var addressReq *model.AddressRequest
	if req.Address != nil {
		addressReq = &model.AddressRequest{
			ReceiverName: req.Address.ReceiverName,
			Phone:        req.Address.Phone,
			AddressLine:  req.Address.AddressLine,
			City:         req.Address.City,
			Province:     req.Address.Province,
			PostalCode:   req.Address.PostalCode,
			IsDefault:    req.Address.IsDefault,
		}
	}

	registerReq := model.RegisterRequest{
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Address: addressReq,
	}
	registerReq.Password = req.Password

	user, err := h.userUsecase.Register(ctx, registerReq)
	if err != nil {
		log.WithError(err).Error("register failed")
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &pb.RegisterResponse{
		Id:    user.ID.String(),
		Name:  user.Name,
		Email: user.Email,
		Role:  string(user.Role),
		Phone: user.Phone,
	}, nil
}

func (h *AuthGRPCHandler) CreateAdmin(ctx context.Context, req *pb.CreateAdminRequest) (*pb.UserResponse, error) {
	createdBy, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		return nil, err
	}

	registerReq := model.RegisterRequest{
		Name:  req.Name,
		Email: req.Email,
		Phone: req.Phone,
		Role:  model.RoleAdmin,
	}
	registerReq.Password = req.Password

	user, err := h.userUsecase.CreateAdmin(ctx, registerReq, createdBy)
	if err != nil {
		log.WithError(err).Error("create admin failed")
		if err == apperrors.ErrUnauthorized {
			return nil, status.Errorf(codes.PermissionDenied, "%s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &pb.UserResponse{
		Id:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		Role:      string(user.Role),
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt.String(),
	}, nil
}
