# syntax=docker/dockerfile:1

FROM golang:1.18-alpine
 
WORKDIR /app

COPY . .
 
RUN go mod download
RUN go build -o /gatomemes ./src/
 
CMD [ "/gatomemes" ]
