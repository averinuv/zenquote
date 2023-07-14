package tcp

import (
	"bufio"
	"context"
	"fmt"
	"net"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"zenquote/internal/config"
)

type Server struct {
	cfg       config.Tcp
	logger    *zap.Logger
	listener  net.Listener
	handler   *Handler
	closeChan chan struct{}
}

func NewServer(cfg config.Config, logger *zap.Logger, handler *Handler) *Server {
	return &Server{cfg: cfg.Tcp, logger: logger, handler: handler}
}

func (s *Server) Start(_ context.Context, stop fx.Shutdowner) {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	s.logger.Info("starting server", zap.String("addr", addr))

	var err error
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		s.logger.Error("error while starting tcp server", zap.Error(err))
		_ = stop.Shutdown()
	}

	for {
		select {
		case <-s.closeChan:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				s.logger.Error("accept connection failed", zap.Error(err))
				continue
			}

			_ = conn
			go s.handleConnection(conn)
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ReqTimeout)
	defer cancel()

	// Create connection scanner
	scanner := bufio.NewScanner(conn)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, s.cfg.MaxReqSizeBytes)

	// Scan connection
	for scanner.Scan() {
		if len(scanner.Text()) >= s.cfg.MaxReqSizeBytes {
			s.logger.Error("request to large", zap.String("addr", conn.RemoteAddr().String()))
			return
		}

		s.handler.Handle(ctx, conn, scanner.Bytes())
	}

	if err := scanner.Err(); err != nil {
		s.logger.Error("reading from connection failed", zap.Error(err))
	}
}

func (s *Server) Shutdown() {
	s.logger.Info("stopping server")
	close(s.closeChan)
	err := s.listener.Close()
	if err != nil {
		s.logger.Error("close tcp listener failed", zap.Error(err))
	}
}
