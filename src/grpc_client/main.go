package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"

	pb "github.com/olde-ducke/gatomemes/src/drawtext"
	"google.golang.org/grpc"
)

var (
	addr = flag.String("a", "localhost:50051", "the grpc server address to connect to")
	out  = flag.String("o", "out.png", "output file name, will be saved at the working dir")
	src  = flag.String("s", "", "(required) original image source or base64 encoded image, only "+
		"jpeg and png formats are supported")
	text = flag.String("t", "", "(required) text separated by next \"@\" symbol, 3 lines (top, "+
		"middle, bottom) will be drawn in the centre of the source image")
)

func usage() {
	fmt.Println("test")
}

func main() {
	flag.Usage = usage
	flag.Parse()
	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("connection failed with error: %v", err)
	}
	defer conn.Close()

	client := pb.NewDrawTextClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	reply, err := client.Draw(ctx, &pb.DrawRequest{
		Src:  *src,
		Text: *text,
	})
	if err != nil {
		log.Fatalf("request failed with error: %v", err)
	}
	fmt.Println(reply.GetReply())

	err = os.WriteFile(*out, reply.GetData(), fs.ModePerm)
	if err != nil {
		log.Fatalf("saving file %s failed with error: %v", *out, err)
	}
}
