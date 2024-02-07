package handler

import (
	"context"
	"log"

	"github.com/sifer169966/go-grpc-client-lb/server/apis/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ pb.DeviceInteractionServiceServer = (*grpcHandler)(nil)

type grpcHandler struct {
	pb.UnimplementedDeviceInteractionServiceServer
}

func NewGRPC() *grpcHandler {
	return &grpcHandler{}
}

func (hdl *grpcHandler) CreateDeviceInteraction(_ context.Context, in *pb.CreateDeviceInteractionRequest) (*emptypb.Empty, error) {
	log.Printf("got request with payload {timestamp = %v, Localtion = %+v, Devices = %+v}\n", in.Timestamp, in.Localtion, in.Devices)
	return &emptypb.Empty{}, nil
}
