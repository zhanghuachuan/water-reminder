package trpcservices

import (
	"context"

	"github.com/zhanghuachuan/water-reminder/api/proto"
	"github.com/zhanghuachuan/water-reminder/internal/services"
	"trpc.group/trpc-go/trpc-go/server"
)

type AuthService struct{}

func RegisterAuthService(s server.Service, svr *AuthService) {
	proto.RegisterAuthServiceService(s, svr)
}

func (s *AuthService) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	user, err := services.Register(req.Username, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &proto.RegisterResponse{
		UserId:   uint32(user.ID),
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	user, err := services.Login(req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &proto.LoginResponse{
		UserId:   uint32(user.ID),
		Username: user.Username,
		Email:    user.Email,
	}, nil
}
