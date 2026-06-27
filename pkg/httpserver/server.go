package httpserver

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ServerConfig struct {
	Port            int           `env:"SERVER_PORT" envDefault:"8080"`
	ReadTimeout     time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"30s"`
	WriteTimeout    time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"30s"`
	ShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT" envDefault:"10s"`
}

type Server struct {
	srv *http.Server
	cfg ServerConfig
}

func NewServer(cfg ServerConfig, router *gin.Engine) *Server {
	return &Server{
		srv: &http.Server{
			Addr:         ":" + strconv.Itoa(cfg.Port),
			Handler:      router,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.WriteTimeout + 5*time.Second,
		},
		cfg: cfg,
	}
}

func (s *Server) Start() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		zap.L().Info("server starting", zap.String("addr", s.srv.Addr))
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("server failed", zap.Error(err))
		}
	}()

	<-quit
	zap.L().Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		zap.L().Error("server shutdown error", zap.Error(err))
	}
	zap.L().Info("server stopped")
}
