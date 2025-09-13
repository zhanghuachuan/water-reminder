package trpcservices

import (
	"context"

	"github.com/zhanghuachuan/water-reminder/api/proto"
	"github.com/zhanghuachuan/water-reminder/internal/services"
	"trpc.group/trpc-go/trpc-go/server"
)

type AuthService struct {
	proto.UnimplementedAuthServiceServer
}

func RegisterAuthService(s server.Service, svr *AuthService) {
	s.Register(&server.ServiceDesc{
		ServiceName: "water_reminder.auth.AuthService",
		HandlerType: (*AuthService)(nil),
		Methods: []server.Method{
			{
				Name: "Register",
				Func: svr.Register,
			},
			{
				Name: "Login",
				Func: svr.Login,
			},
		},
	}, svr)
}

func (s *AuthService) Register(svr interface{}, ctx context.Context, f server.FilterFunc) (interface{}, error) {
	req := &proto.RegisterRequest{}
	_, err := f(req)
	if err != nil {
		return nil, err
	}
	user, err := services.Register(req.Username, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	resp := &proto.RegisterResponse{
		UserId:   uint32(user.ID),
		Username: user.Username,
		Email:    user.Email,
	}
	return resp, nil
}

// mustEmbedUnimplementedAuthServiceServer 是protobuf生成的接口实现方法
// 用于向后兼容，即使未使用也不应删除
func (s *AuthService) mustEmbedUnimplementedAuthServiceServer() {}

func (s *AuthService) Login(svr interface{}, ctx context.Context, f server.FilterFunc) (interface{}, error) {
	req := &proto.LoginRequest{}
	_, err := f(req)
	if err != nil {
		return nil, err
	}
	user, err := services.Login(req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	resp := &proto.LoginResponse{
		UserId:   uint32(user.ID),
		Username: user.Username,
		Email:    user.Email,
	}
	return resp, nil
}
