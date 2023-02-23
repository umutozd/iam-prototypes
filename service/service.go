package service

import (
	"context"

	"github.com/umutozd/iam-prototypes/pb"
)

func (s *Server) GetFoo(ctx context.Context, in *pb.GetFooReq) (*pb.Foo, error) {
	return &pb.Foo{
		Id:    in.Name,
		Count: 42,
	}, nil
}

func (s *Server) UpdateFoo(ctx context.Context, in *pb.UpdateFooReq) (*pb.Foo, error) {
	return &pb.Foo{
		Id:    in.FooId,
		Count: in.Count,
	}, nil
}
