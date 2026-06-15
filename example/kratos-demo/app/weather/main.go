// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"os"
	"time"

	messagev1 "kratos-demo/api/message"
	v1 "kratos-demo/api/weather"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	grpcx "google.golang.org/grpc"
)

// server implements the WeatherService HTTP server
type server struct {
	v1.UnimplementedWeatherServiceServer
}

// GetWeatherMessage handles HTTP requests for weather messages by calling the MessageService via gRPC
func (s *server) GetWeatherMessage(ctx context.Context, in *v1.GetWeatherMessageRequest) (*v1.GetWeatherMessageResponse, error) {
	// Create gRPC connection to MessageService
	// NOTE: In production, use connection pooling and proper service discovery
	conn, err := grpc.DialInsecure(ctx,
		grpc.WithEndpoint("127.0.0.1:8081"), // Connect to MessageService on port 8081
		grpc.WithMiddleware(
			recovery.Recovery(), // Handle panics in client calls
			tracing.Client(),    // Add tracing to outgoing requests
		),
		grpc.WithTimeout(2*time.Second), // Set reasonable timeout
		// Enable stats handler for distributed tracing
		grpc.WithOptions(grpcx.WithStatsHandler(&tracing.ClientHandler{})),
	)
	if err != nil {
		return nil, err
	}

	// Create MessageService client
	messageClient := messagev1.NewMessageServiceClient(conn)

	// Call MessageService to get weather information
	res, err := messageClient.GetWeatherMessage(ctx, &messagev1.GetWeatherMessageRequest{
		City: in.City,
	})
	if err != nil {
		return nil, err
	}

	// Return the weather message response
	return &v1.GetWeatherMessageResponse{
		Message: res.Content,
	}, nil
}

func main() {
	// Initialize structured logger with tracing support
	logger := log.NewStdLogger(os.Stdout)
	logger = log.With(logger, "trace_id", tracing.TraceID())
	logger = log.With(logger, "span_id", tracing.SpanID())
	log := log.NewHelper(logger)

	// Configure HTTP server with middleware chain
	httpSrv := http.NewServer(
		http.Address(":8080"), // Listen on port 8080
		http.Middleware(
			recovery.Recovery(),    // Panic recovery middleware
			tracing.Server(),       // OpenTelemetry tracing middleware
			logging.Server(logger), // Request logging middleware
		),
	)

	// Create server instance and register HTTP handlers
	s := &server{}
	v1.RegisterWeatherServiceHTTPServer(httpSrv, s)

	// Create and configure the Kratos application
	app := kratos.New(
		kratos.Name("weather_service"),
		kratos.Server(httpSrv),
	)

	// Start the server and handle any startup errors
	if err := app.Run(); err != nil {
		log.Error(err)
	}
}
