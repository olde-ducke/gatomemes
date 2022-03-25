package main

func main() {
	go grpcServerRun()
	httpServerRun()
}
