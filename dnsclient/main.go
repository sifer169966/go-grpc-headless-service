package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sifer169966/go-grpc-client-lb/dnsclient/config"
	"github.com/sifer169966/go-grpc-client-lb/server/apis/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	config.Init()
	srv := echo.New()
	serverHost := fmt.Sprintf("dns:///%s:%s", config.Get().GRPClient.ServerHost, config.Get().GRPClient.ServerPort)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	opts = append(opts, grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`))
	conn, err := grpc.Dial(serverHost, opts...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := pb.NewDeviceInteractionServiceClient(conn)
	srv.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "dnsclient service is running.")
	})
	srv.GET("/try", func(c echo.Context) error {
		id := uuid.NewString()
		log.Printf("ID=%s STATE=%s TARGET=%s\n", id, conn.GetState().String(), conn.Target())
		err := tryClient(id, client)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, "success")
	})
	err = srv.Start(":" + config.Get().App.RESTPort)
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func tryClient(id string, c pb.DeviceInteractionServiceClient) error {
	payload := &pb.CreateDeviceInteractionRequest{
		Timestamp: timestamppb.New(time.Now()),
		Localtion: &pb.GeoLocation{
			Latitude:  "90",
			Longitude: "-120",
		},
		Devices: []*pb.Device{
			{
				Id:   id,
				Name: "Monkey D Luffer",
			},
		},
	}
	log.Println("calling to grpc server with id= ", id)
	_, err := c.CreateDeviceInteraction(context.Background(), payload)
	if err != nil {
		log.Printf("error: %s from id=%s\n", err.Error(), id)
		return err
	}
	log.Println(" success response from grpc server with id=", id)
	return nil
}
