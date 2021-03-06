# grpc-client

grpc-client, works with running server

can be used to draw text over source image, result is saved in working directory

flags:

    -a          specify server address and port (default: "localhost:50051")
    -s          source image url or base64 encoded image, only jpeg and png are supported (required)
    -t          text to draw over source, "@" will be replaced with "\n", up to 3 lines of text supported
    -o          output file name, saves file in working dir, output is always png (default: "out.png")
    -h,--help   prints help message
    --fonts     prints fonts file names available on server

drawing options (optional):

    -i,         use font with index i, font list can be obtained with --fonts flag
    --fscale    relative font scale, higher values will give bigger glyphs (values: 1-4)
    --fcolor    rgb hex color as string, i.e. "ffffff", also supports "random" value, which will give random color
    --ocolor    same as above, but for outline
    --oscale    same as font scale, but for outline
    --nooutline disable outline drawing
    --distort   add noise to glyph coordinates

running command:

    go run . -s "https://avatars.githubusercontent.com/u/89552583?v=4" -t "test@TEST@test" --fcolor 0000f0 --ocolor random --oscale 4 --fscale 3 -i 6 --distort

will produce image:

![Example](./out.png)
