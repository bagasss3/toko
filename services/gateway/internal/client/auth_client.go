package client

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/bagasss3/toko/services/auth-service/pb/auth"
	"github.com/bagasss3/toko/services/gateway/internal/model"
)

type AuthClient struct {
	client pb.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthClient(addr string) (*AuthClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth-service: %w", err)
	}

	log.Infof("connected to auth-service at %s", addr)

	return &AuthClient{
		client: pb.NewAuthServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *AuthClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *AuthClient) Login(ctx context.Context, email, password string) (string, error) {
	res, err := c.client.Login(ctx, &pb.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", err
	}

	return res.AccessToken, nil
}

func (c *AuthClient) Register(ctx context.Context, email, password string) (*model.TokenClaims, error) {
	res, err := c.client.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	return &model.TokenClaims{
		UserID: res.Id,
		Email:  res.Email,
	}, nil
}

func (c *AuthClient) ValidateToken(ctx context.Context, token string) (*model.TokenClaims, error) {
	res, err := c.client.ValidateToken(ctx, &pb.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		return nil, err
	}

	return &model.TokenClaims{
		UserID: res.UserId,
		Email:  res.Email,
	}, nil
}
