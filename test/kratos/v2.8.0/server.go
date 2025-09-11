package main

import (
	"flag"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	server "kratos/v2.5.2/pkg"
	"kratos/v2.5.2/pkg/biz"
	"kratos/v2.5.2/pkg/conf"
	"kratos/v2.5.2/pkg/service"
	"os"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name = "opentelemetry-kratos-server"
	// Version is the version of the compiled software.
	Version = "v1"
	// flagconf is the config flag.
	flagconf string

	id = "opentelemetry-id"
)

func init() {
	flag.StringVar(&flagconf, "conf", "pkg/configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{
			"agent": "opentelemetry-go",
		}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func wireApp(confServer *conf.Server, confData *conf.Data, logger log.Logger) (*kratos.App, func(), error) {
	greeterUsecase := biz.NewGreeterUsecase(logger)
	greeterService := service.NewGreeterService(greeterUsecase)
	grpcServer := server.NewGRPCServer(confServer, greeterService)
	httpServer := server.NewHTTPServer(confServer, greeterService)
	app := newApp(logger, grpcServer, httpServer)
	return app, func() {}, nil
}

func startup() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
