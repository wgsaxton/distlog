package main

import (
	"fmt"

	api "github.com/wgsaxton/distlog/api/v1"
)

var testingv api.LogServer = (*grpcServer)(nil)

type grpcServer struct {
	api.UnimplementedLogServer
	*Config
}

type Config struct {
	CommitLog string
}

func main() {
	fmt.Println(*testingv)
}
