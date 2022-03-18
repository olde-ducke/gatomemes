package main

import (
	"context"
	"log"
	"net"
	"os"

	pb "github.com/olde-ducke/gatomemes/src/drawtext"
	"github.com/olde-ducke/gatomemes/src/gatomemes"
	"google.golang.org/grpc"
)

var logger = log.New(os.Stdout, "\x1b[32m[RPC] \x1b[0m", log.LstdFlags)

type server struct {
	pb.UnimplementedDrawTextServer
}

func (s *server) Draw(ctx context.Context, in *pb.DrawRequest) (*pb.DrawReply, error) {
	logger.Printf("received src: \"%s\" text: \"%s\"", in.GetSrc(), in.GetText())
	reply := &pb.DrawReply{Reply: "available fonts: " + os.Getenv("PROJECTFONTS")}
	data, err := gatomemes.GetNewFromSrc(in.GetSrc(), in.GetText())
	if err != nil {
		return reply, err
	}
	reply.Data = data
	return reply, nil
}

func grpcServerRun() {
	lis, err := net.Listen("tcp", os.Getenv("GRPCADDR"))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterDrawTextServer(s, &server{})
	logger.Printf("grcp server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
