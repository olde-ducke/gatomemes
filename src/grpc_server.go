package main

import (
	"context"
	"log"
	"net"
	"os"

	pb "github.com/olde-ducke/gatomemes/src/drawtext"
	"github.com/olde-ducke/gatomemes/src/gatomemes"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const maxMsgSize = 1024 * 1024 * 20

var logger = log.New(os.Stdout, "\x1b[32m[RPC] \x1b[0m", log.LstdFlags)

type server struct {
	pb.UnimplementedDrawTextServer
}

func (s *server) Draw(ctx context.Context, in *pb.DrawRequest) (*pb.DrawReply, error) {
	logger.Printf("received src: \"%s\" text: \"%s\"", in.GetSrc(), in.GetText())
	data, err := gatomemes.CreateNewFromSrc(in.GetSrc(), in.GetText(),
		&gatomemes.Options{
			FontIndex:      in.GetIndex(),
			FontScale:      in.GetFontScale(),
			FontColor:      in.GetFontColor(),
			OutlineColor:   in.GetOutlineColor(),
			OutlineScale:   in.GetOutlineScale(),
			DisableOutline: in.GetDisableOutline(),
			Distort:        in.GetDistort(),
		})
	if err != nil {
		return nil, err
	}
	return &pb.DrawReply{Data: data}, nil
}

func (s *server) GetFontNames(ctx context.Context, in *emptypb.Empty) (*pb.TextReply, error) {
	logger.Println("received request for font names")
	return &pb.TextReply{Filenames: "available fonts:\n" + os.Getenv("APP_FONTS")}, nil
}

func grpcServerRun() {
	lis, err := net.Listen("tcp", os.Getenv("GRPC_ADDR"))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(grpc.MaxSendMsgSize(maxMsgSize))
	pb.RegisterDrawTextServer(s, &server{})
	logger.Printf("grcp server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
