package tcp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"zenquote/api"
	"zenquote/internal/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Request struct {
	*api.Request
	ClientIP string
}

func NewRequest(conn net.Conn, reqBytes []byte) (*Request, error) {
	var reqBody api.Request
	if err := proto.Unmarshal(reqBytes, &reqBody); err != nil {
		return nil, fmt.Errorf("unmarshal request failed: %w", err)
	}

	clientIP, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	return &Request{Request: &reqBody, ClientIP: clientIP}, nil
}

type Server struct {
	cfg       config.TCP
	logger    *zap.Logger
	listener  net.Listener
	handler   *Handler
	closeChan chan struct{}
}

func NewServer(cfg config.Config, logger *zap.Logger, handler *Handler) *Server {
	return &Server{
		cfg:       cfg.TCP,
		logger:    logger,
		handler:   handler,
		listener:  nil,
		closeChan: make(chan struct{}),
	}
}

// Start Configure and start TCP server.
func (s *Server) Start(_ context.Context, stop fx.Shutdowner) {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	s.logger.Info("starting server", zap.String("addr", addr))

	var err error
	if s.listener, err = net.Listen("tcp", addr); err != nil {
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

			go func(conn net.Conn) {
				// Create a context with a timeout
				ctx, cancelCtx := context.WithTimeout(context.Background(), s.cfg.ReqTimeout)
				defer cancelCtx()

				// Set conn deadline
				d, _ := ctx.Deadline()
				if err = conn.SetDeadline(d); err != nil {
					s.logger.Error("set connection deadline failed", zap.Error(err))
					cancelCtx()

					return
				}

				s.handleConn(ctx, conn)
			}(conn)
		}
	}
}

func (s *Server) handleConn(ctx context.Context, conn net.Conn) {
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	reqCount := 0
	scanner := s.readFromConnection(conn)

	for scanner.Scan() {
		reqCount++

		// Validate request
		if !s.validateReqSize(scanner.Bytes(), conn) || !s.validateReqLimit(reqCount, conn) {
			break
		}

		// Create Request
		req, err := NewRequest(conn, scanner.Bytes())
		if err != nil {
			s.logger.Error("failed to create request", zap.Error(err))

			return
		}

		s.handler.Handle(ctx, conn, req)
	}

	if err := scanner.Err(); err != nil {
		s.logger.Error("reading from connection failed", zap.Error(err))
	}
}

func (s *Server) readFromConnection(conn net.Conn) *bufio.Scanner {
	scanner := bufio.NewScanner(conn)
	buf := make([]byte, 0, s.cfg.MaxReqSizeBytes)
	scanner.Buffer(buf, s.cfg.MaxReqSizeBytes)

	return scanner
}

func (s *Server) validateReqSize(data []byte, respWrite io.Writer) bool {
	if len(data) >= s.cfg.MaxReqSizeBytes {
		respBytes, _ := proto.Marshal(&api.Response{
			Status:   api.Response_SUCCESS,
			Response: &api.Response_Error{Error: "request too large"},
		})
		respBytes = append(respBytes, '\n')
		_, _ = respWrite.Write(respBytes)

		return false
	}

	return true
}

func (s *Server) validateReqLimit(reqCount int, respWrite io.Writer) bool {
	if reqCount > s.cfg.MaxReqPerSession {
		respBytes, _ := proto.Marshal(&api.Response{
			Status:   api.Response_FAILURE,
			Response: &api.Response_Error{Error: "session request limit exceeded"},
		})
		respBytes = append(respBytes, '\n')
		_, _ = respWrite.Write(respBytes)

		return false
	}

	return true
}

// Shutdown stops the server and cleans up any resources it was using.
// Calling Shutdown multiple times or while the server is already stopped will cause a runtime panic.
// Ensure that Shutdown is called exactly once when the server is no longer needed.
func (s *Server) Shutdown() {
	s.logger.Info("stopping server")
	close(s.closeChan)

	if err := s.listener.Close(); err != nil {
		s.logger.Error("close tcp listener failed", zap.Error(err))
	}
}
