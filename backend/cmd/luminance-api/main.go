package main

import (
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/luminance/backend/internal/server"
)

func main() {
	logger := log.NewStdLogger(os.Stdout)

	httpSrv := server.NewHTTPServer()
	grpcSrv := server.NewGRPCServer()

	app := kratos.New(
		kratos.Name("luminance-api"),
		kratos.Logger(logger),
		kratos.Server(httpSrv, grpcSrv),
	)

	if err := app.Run(); err != nil {
		log.NewHelper(logger).Fatalf("failed to run luminance-api: %v", err)
	}
}
