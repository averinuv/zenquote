package tcp

import (
	"net"
	"testing"
	"zenquote/api"
	"zenquote/internal/config"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"google.golang.org/protobuf/proto"
)

func TestNewRequest(t *testing.T) {
	t.Parallel()

	reqBody := &api.Request{Cmd: 0, Data: "test_data"}

	tests := []struct {
		name    string
		reqBody *api.Request
		wantErr bool
	}{
		{
			name:    "Success",
			reqBody: reqBody,
			wantErr: false,
		},
		{
			name:    "UnmarshalError",
			reqBody: &api.Request{Cmd: 0, Data: string([]byte{0xff, 0xfe, 0xfd})},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		ttCopy := tt
		t.Run(ttCopy.name, func(t *testing.T) {
			t.Parallel()

			reqBytes, _ := proto.Marshal(ttCopy.reqBody)
			conn, _ := net.Pipe()

			_, err := NewRequest(conn, reqBytes)
			if (err != nil) != ttCopy.wantErr {
				t.Errorf("NewRequest() error = %v, wantErr %v", err, ttCopy.wantErr)

				return
			}
		})
	}
}

func TestReadFromConnection(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		TCP: config.TCP{
			Host:             "",
			Port:             0,
			ReqTimeout:       0,
			MaxReqSizeBytes:  2048,
			MaxReqPerSession: 0,
		},
		Redis:  config.Redis{Host: "", Port: 0},
		Logger: config.Logger{Level: "", Encoding: "", Colored: false, Tags: nil},
	}

	server := NewServer(cfg, zap.NewNop(), nil)

	conn1, conn2 := net.Pipe()
	defer func(conn1 net.Conn) {
		_ = conn1.Close()
	}(conn1)
	defer func(conn2 net.Conn) {
		_ = conn2.Close()
	}(conn2)

	go func() {
		_, _ = conn2.Write([]byte("test data"))
		_ = conn2.Close()
	}()

	scanner := server.readFromConnection(conn1)
	scanner.Scan() // call Scan() before checking the size of scanner buffer

	assert.NotNil(t, scanner)
	assert.Equal(t, cfg.TCP.MaxReqSizeBytes, cap(scanner.Bytes()))

	if err := scanner.Err(); err != nil {
		t.Error("Failed to scan from connection")
	}

	assert.Equal(t, "test data", scanner.Text())
}

func TestIsValidReqSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     []byte
		maxSize  int
		expected bool
	}{
		{name: "valid size", data: make([]byte, 1024), maxSize: 2048, expected: true},
		{name: "invalid size", data: make([]byte, 2049), maxSize: 2048, expected: false},
	}

	for _, tc := range tests {
		tcCopy := tc
		t.Run(tcCopy.name, func(t *testing.T) {
			t.Parallel()

			// Create a connection for each test case
			conn1, conn2 := net.Pipe()
			defer func(conn1 net.Conn) {
				_ = conn1.Close()
			}(conn1)
			defer func(conn2 net.Conn) {
				_ = conn2.Close()
			}(conn2)

			cfg := config.Config{
				TCP: config.TCP{
					Host:             "",
					Port:             0,
					ReqTimeout:       0,
					MaxReqSizeBytes:  2048,
					MaxReqPerSession: 0,
				},
				Redis:  config.Redis{Host: "", Port: 0},
				Logger: config.Logger{Level: "", Encoding: "", Colored: false, Tags: nil},
			}

			server := NewServer(cfg, zap.NewNop(), nil)

			// Write data to connection and close it
			go func() {
				_, _ = conn2.Write(tcCopy.data)
				_ = conn2.Close()
			}()

			// Read data from connection
			buf := make([]byte, tcCopy.maxSize+1)
			n, _ := conn1.Read(buf)

			assert.Equal(t, tcCopy.expected, server.validateReqSize(buf[:n], conn1))
		})
	}
}

func TestIsValidReqLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		reqCount int
		maxCount int
		expected bool
	}{
		{
			name:     "valid request count",
			reqCount: 5,
			maxCount: 10,
			expected: true,
		},
		{
			name:     "invalid request count",
			reqCount: 10,
			maxCount: 5,
			expected: false,
		},
	}

	for _, tc := range tests {
		tcCopy := tc
		t.Run(tcCopy.name, func(t *testing.T) {
			t.Parallel()
			serverConn, clientConn := net.Pipe()

			cfg := config.Config{
				TCP: config.TCP{
					Host:             "",
					Port:             0,
					ReqTimeout:       0,
					MaxReqSizeBytes:  2048,
					MaxReqPerSession: tcCopy.maxCount,
				},
				Redis:  config.Redis{Host: "", Port: 0},
				Logger: config.Logger{Level: "", Encoding: "", Colored: false, Tags: nil},
			}

			server := NewServer(cfg, zap.NewNop(), nil)

			// Create a separate goroutine to handle potential server writes.
			go func() {
				// Use a simple protocol: server writes, then client reads.
				for i := 0; i < tcCopy.reqCount; i++ {
					_, _ = clientConn.Read(make([]byte, 1024))
				}
				// Once done, close the client side of the connection.
				_ = clientConn.Close()
			}()

			// Assert our condition.
			assert.Equal(t, tcCopy.expected, server.validateReqLimit(tcCopy.reqCount, serverConn))

			// Clean up the server side of the connection.
			_ = serverConn.Close()
		})
	}
}
