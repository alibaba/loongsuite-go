package main

import (
	"context"
	"os"

	v1 "github.com/alibaba/loongsuite-go-agent/example/kratos-demo/api/message"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// server implements the MessageService gRPC server
type server struct {
	v1.UnimplementedMessageServiceServer
}

// GetWeatherMessage handles weather message requests and returns weather information for the specified city
func (s *server) GetWeatherMessage(ctx context.Context, request *v1.GetWeatherMessageRequest) (*v1.WeatherMessage, error) {
	// Validate input parameters
	if request.City == "" {
		return nil, status.Errorf(codes.InvalidArgument, "City name cannot be empty")
	}

	// Static weather data for demonstration purposes
	// In production, this would typically fetch data from external weather APIs
	weatherMap := map[string]string{
		"HangZhou":  "Hangzhou: ☀️ Sunny skies throughout the day. High: 32°C, Low: 25°C. Wind: Light breeze at 5 km/h.",
		"ShangHai":  "Shanghai: ⛅ Partly cloudy with occasional sunshine. High: 31°C, Low: 26°C. Wind: Southeast at 15 km/h.",
		"BeiJing":   "Beijing: 🌤️ Sunny in the morning, becoming cloudy in the afternoon. High: 30°C, Low: 22°C. Wind: North at 10 km/h.",
		"ShenZhen":  "Shenzhen: ⛈️ Thunderstorms expected in the afternoon. High: 33°C, Low: 28°C. Wind: South at 25 km/h.",
		"GuangZhou": "Guangzhou: 🌦️ Scattered showers throughout the day. High: 32°C, Low: 27°C. Wind: Southeast at 15 km/h.",
		"NewYork":   "New York: 🌥️ Mostly cloudy with a chance of evening rain. High: 25°C, Low: 18°C. Wind: West at 12 km/h.",
		"London":    "London: 🌧️ Light rain expected all day. High: 18°C, Low: 12°C. Wind: Southwest at 18 km/h.",
		"Tokyo":     "Tokyo: ☀️ Clear and sunny conditions. High: 28°C, Low: 20°C. Wind: Light and variable.",
		"Sydney":    "Sydney: 🌞 Beautiful sunny day. High: 30°C, Low: 22°C. Wind: Northeast at 10 km/h.",
		"Paris":     "Paris: ☁️ Overcast skies with no rain expected. High: 22°C, Low: 15°C. Wind: Northwest at 8 km/h.",
	}

	// Look up weather data for the requested city
	content, exists := weatherMap[request.City]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "Weather forecast not available for city: %s", request.City)
	}

	// Return the weather message
	return &v1.WeatherMessage{
		Content: content,
	}, nil
}

func main() {
	// Initialize structured logger with tracing support
	logger := log.NewStdLogger(os.Stdout)
	logger = log.With(logger, "trace_id", tracing.TraceID())
	logger = log.With(logger, "span_id", tracing.SpanID())
	log := log.NewHelper(logger)

	// Create server instance
	s := &server{}

	// Configure gRPC server with middleware chain
	grpcSrv := grpc.NewServer(
		grpc.Address(":8081"), // Listen on port 8081
		grpc.Middleware(
			middleware.Chain(
				recovery.Recovery(),    // Panic recovery middleware
				tracing.Server(),       // OpenTelemetry tracing middleware
				logging.Server(logger), // Request logging middleware
			),
		),
	)

	// Register the message service implementation
	v1.RegisterMessageServiceServer(grpcSrv, s)

	// Create and configure the Kratos application
	app := kratos.New(
		kratos.Name("message_service"),
		kratos.Server(grpcSrv),
	)

	// Start the server and handle any startup errors
	if err := app.Run(); err != nil {
		log.Error(err)
	}
}
