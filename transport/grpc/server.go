package grpc

import (
	ragv1 "Api_Rag/gen/rag/v1"
	"Api_Rag/services"
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// InterceptorLogger adapta slog para logging interceptor
func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

// NewGRPCServer configura o servidor gRPC com middlewares de logging, recovery e reflection
func NewGRPCServer(albumService *services.AlbumService) *grpc.Server {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p any) (err error) {
			logger.Error("Pânico recuperado no handler gRPC", "panic", p)
			return status.Errorf(codes.Internal, "erro interno do servidor")
		}),
	}

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
		logging.WithDurationField(func(duration time.Duration) logging.Fields {
			return logging.Fields{"grpc.time_ms", duration.Milliseconds()}
		}),
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(recoveryOpts...),
			logging.UnaryServerInterceptor(InterceptorLogger(logger), loggingOpts...),
		),
		grpc.ChainStreamInterceptor(
			recovery.StreamServerInterceptor(recoveryOpts...),
			logging.StreamServerInterceptor(InterceptorLogger(logger), loggingOpts...),
		),
	)

	// Registra o serviço de álbuns
	albumServer := NewAlbumGrpcServer(albumService)
	ragv1.RegisterAlbumServiceServer(server, albumServer)

	// Health check padrão
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	// Habilita Reflection para Evans, Postman, grpcurl
	reflection.Register(server)

	return server
}
