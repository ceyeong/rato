package auth

import (
	"context"

	"grng.dev/guff/pb"
)

// Service :
type Service struct {
}

// Register :
func (s *Service) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	//todo implement
	return nil, nil
}

// Login :
func (s *Service) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	//todo implement
	return nil, nil
}
