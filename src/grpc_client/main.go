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
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

const maxMsgSize = 1024 * 1024 * 20

var (
	addr      = flag.String("a", "localhost:50051", "")
	src       = flag.String("s", "", "")
	text      = flag.String("t", "", "")
	out       = flag.String("o", "out.png", "")
	names     = flag.Bool("fonts", false, "")
	i         = flag.Int64("i", 0, "")
	fScale    = flag.Int64("fscale", 0, "")
	fColor    = flag.String("fcolor", "", "")
	oColor    = flag.String("ocolor", "", "")
	oScale    = flag.Int64("oscale", 0, "")
	noOutline = flag.Bool("nooutline", false, "")
	distort   = flag.Bool("distort", false, "")
)

func usage() {
	fmt.Println(`
grpc-client, works with running server

can be used to draw text over source image, result is saved in working directory

FLAGS:
 -a             specify server address and port (default: "localhost:50051")
 -s             source image url or base64 encoded image, only jpeg and png are supported (required)
 -t             text to draw over source, "@" will be replaced with "\n", up to 3 lines of text supported
 -o             output file name, saves file in working dir, output is always png (default: "out.png")
 -h,--help      prints help message
    --fonts     prints fonts file names available on server

DRAWING OPTIONS (optional):
 -i,            use font with index i, font list can be obtained with --fonts flag
	--fscale	relative font scale, higher values will give bigger glyphs (values: 1-4)
    --fcolor	rgb hex color as string, i.e. "ffffff", also supports "random" value, which will give random color
    --ocolor	same as above, but for outline
    --oscale	same as font scale, but for outline
    --nooutline	disable outline drawing
    --distort	add noise to glyph coordinates
	`)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize)))
	if err != nil {
		log.Fatalf("connection failed with error: %v", err)
	}
	defer conn.Close()

	client := pb.NewDrawTextClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if *names {
		reply, err := client.GetFontNames(ctx, &emptypb.Empty{})
		if err != nil {
			log.Fatalf("request failed with error: %v", err)
		}
		fmt.Println(reply.GetFilenames())
		os.Exit(0)
	}

	reply, err := client.Draw(ctx, &pb.DrawRequest{
		Src:            *src,
		Text:           *text,
		Index:          *i,
		FontScale:      *fScale,
		FontColor:      *fColor,
		OutlineColor:   *oColor,
		OutlineScale:   *oScale,
		DisableOutline: *noOutline,
		Distort:        *distort,
	})
	if err != nil {
		log.Fatalf("request failed with error: %v", err)
	}

	err = os.WriteFile(*out, reply.GetData(), fs.ModePerm)
	if err != nil {
		log.Fatalf("saving file %s failed with error: %v", *out, err)
	}
}
