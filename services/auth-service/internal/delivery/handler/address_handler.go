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

func (h *AuthGRPCHandler) GetMyAddresses(ctx context.Context, req *pb.GetMyAddressesRequest) (*pb.AddressesResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id")
	}

	addresses, err := h.addressUsecase.GetMyAddresses(ctx, userID)
	if err != nil {
		log.WithError(err).Error("get my addresses failed")
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	pbAddresses := make([]*pb.Address, len(addresses))
	for i, addr := range addresses {
		pbAddresses[i] = toPBAddress(addr)
	}

	return &pb.AddressesResponse{Addresses: pbAddresses}, nil
}

func (h *AuthGRPCHandler) GetAddressByID(ctx context.Context, req *pb.GetAddressByIDRequest) (*pb.AddressResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id")
	}

	addressID, err := uuid.Parse(req.AddressId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address id")
	}

	address, err := h.addressUsecase.GetByID(ctx, userID, addressID)
	if err != nil {
		log.WithError(err).Error("get address by id failed")
		switch err {
		case apperrors.ErrAddressNotFound:
			return nil, status.Errorf(codes.NotFound, "%s", err.Error())
		case apperrors.ErrNotOwner:
			return nil, status.Errorf(codes.PermissionDenied, "%s", err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "%s", err.Error())
		}
	}

	return &pb.AddressResponse{Address: toPBAddress(address)}, nil
}

func (h *AuthGRPCHandler) CreateAddress(ctx context.Context, req *pb.CreateAddressRequest) (*pb.AddressResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id")
	}

	addressReq := model.AddressRequest{
		ReceiverName: req.Address.ReceiverName,
		Phone:        req.Address.Phone,
		AddressLine:  req.Address.AddressLine,
		City:         req.Address.City,
		Province:     req.Address.Province,
		PostalCode:   req.Address.PostalCode,
		IsDefault:    req.Address.IsDefault,
	}

	address, err := h.addressUsecase.Create(ctx, userID, addressReq)
	if err != nil {
		log.WithError(err).Error("create address failed")
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &pb.AddressResponse{Address: toPBAddress(address)}, nil
}

func (h *AuthGRPCHandler) UpdateAddress(ctx context.Context, req *pb.UpdateAddressRequest) (*pb.AddressResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id")
	}

	addressID, err := uuid.Parse(req.AddressId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address id")
	}

	addressReq := model.AddressRequest{
		ReceiverName: req.Address.ReceiverName,
		Phone:        req.Address.Phone,
		AddressLine:  req.Address.AddressLine,
		City:         req.Address.City,
		Province:     req.Address.Province,
		PostalCode:   req.Address.PostalCode,
		IsDefault:    req.Address.IsDefault,
	}

	address, err := h.addressUsecase.Update(ctx, userID, addressID, addressReq)
	if err != nil {
		log.WithError(err).Error("update address failed")
		switch err {
		case apperrors.ErrAddressNotFound:
			return nil, status.Errorf(codes.NotFound, "%s", err.Error())
		case apperrors.ErrNotOwner:
			return nil, status.Errorf(codes.PermissionDenied, "%s", err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "%s", err.Error())
		}
	}

	return &pb.AddressResponse{Address: toPBAddress(address)}, nil
}

func (h *AuthGRPCHandler) DeleteAddress(ctx context.Context, req *pb.DeleteAddressRequest) (*pb.DeleteAddressResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id")
	}

	addressID, err := uuid.Parse(req.AddressId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address id")
	}

	err = h.addressUsecase.Delete(ctx, userID, addressID)
	if err != nil {
		log.WithError(err).Error("delete address failed")
		switch err {
		case apperrors.ErrAddressNotFound:
			return nil, status.Errorf(codes.NotFound, "%s", err.Error())
		case apperrors.ErrNotOwner:
			return nil, status.Errorf(codes.PermissionDenied, "%s", err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "%s", err.Error())
		}
	}

	return &pb.DeleteAddressResponse{Success: true, Message: "address deleted"}, nil
}

func (h *AuthGRPCHandler) SetDefaultAddress(ctx context.Context, req *pb.SetDefaultAddressRequest) (*pb.AddressResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id")
	}

	addressID, err := uuid.Parse(req.AddressId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address id")
	}

	err = h.addressUsecase.SetDefault(ctx, userID, addressID)
	if err != nil {
		log.WithError(err).Error("set default address failed")
		switch err {
		case apperrors.ErrAddressNotFound:
			return nil, status.Errorf(codes.NotFound, "%s", err.Error())
		case apperrors.ErrNotOwner:
			return nil, status.Errorf(codes.PermissionDenied, "%s", err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "%s", err.Error())
		}
	}

	// Get updated address
	address, err := h.addressUsecase.GetByID(ctx, userID, addressID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &pb.AddressResponse{Address: toPBAddress(address)}, nil
}

// toPBAddress converts model.Address to pb.Address
func toPBAddress(addr *model.Address) *pb.Address {
	return &pb.Address{
		Id:           addr.ID.String(),
		UserId:       addr.UserID.String(),
		ReceiverName: addr.ReceiverName,
		Phone:        addr.Phone,
		AddressLine:  addr.AddressLine,
		City:         addr.City,
		Province:     addr.Province,
		PostalCode:   addr.PostalCode,
		IsDefault:    addr.IsDefault,
		CreatedAt:    addr.CreatedAt.String(),
	}
}
