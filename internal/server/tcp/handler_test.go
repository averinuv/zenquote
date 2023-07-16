package tcp

import (
	"bytes"
	"context"
	"testing"
	"time"
	"zenquote/api"
	"zenquote/internal/pow"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

type MockRepo struct {
	StoreFunc  func(ctx context.Context, key string, value string, ttl time.Duration) error
	GetFunc    func(ctx context.Context, key string) (string, error)
	DeleteFunc func(ctx context.Context, key string) error
}

func (mr *MockRepo) Store(ctx context.Context, key string, value string, ttl time.Duration) error {
	if mr.StoreFunc != nil {
		return mr.StoreFunc(ctx, key, value, ttl)
	}

	return nil
}

func (mr *MockRepo) Get(ctx context.Context, key string) (string, error) {
	if mr.GetFunc != nil {
		return mr.GetFunc(ctx, key)
	}

	return "", nil
}

func (mr *MockRepo) Delete(ctx context.Context, key string) error {
	if mr.DeleteFunc != nil {
		return mr.DeleteFunc(ctx, key)
	}

	return nil
}

type MockZenquoteRepo struct {
	GetRandomFunc func(ctx context.Context) (string, error)
}

func (zr *MockZenquoteRepo) GetRandom(ctx context.Context) (string, error) {
	if zr.GetRandomFunc != nil {
		return zr.GetRandomFunc(ctx)
	}

	return "", nil
}

func TestHandleGetChallenge(t *testing.T) {
	t.Parallel()

	repo := &MockRepo{
		StoreFunc: func(ctx context.Context, key string, value string, ttl time.Duration) error {
			assert.NotEmpty(t, value)

			return nil
		},
		GetFunc:    nil,
		DeleteFunc: nil,
	}

	zenRepo := &MockZenquoteRepo{
		GetRandomFunc: nil,
	}

	handler := NewHandler(nil, repo, zenRepo)

	req := &Request{
		Request: &api.Request{
			Cmd:  api.Command_GET_CHALLENGE,
			Data: "",
		},
		ClientIP: "",
	}
	writer := &bytes.Buffer{}

	handler.Handle(context.Background(), writer, req)

	resp := &api.Response{
		Status:   0,
		Response: nil,
	}
	_ = proto.Unmarshal(writer.Bytes(), resp)

	assert.Equal(t, api.Response_SUCCESS, resp.Status)
}

func TestHandleCheckSolutionValid(t *testing.T) {
	t.Parallel()

	repo := &MockRepo{
		StoreFunc: func(ctx context.Context, key string, value string, ttl time.Duration) error {
			assert.NotEmpty(t, value)

			return nil
		},
		GetFunc: func(ctx context.Context, key string) (string, error) {
			return "resource", nil
		},
		DeleteFunc: nil,
	}

	zenRepo := &MockZenquoteRepo{
		GetRandomFunc: func(ctx context.Context) (string, error) {
			return "some random Zen quote", nil
		},
	}

	hc, _ := pow.NewHashcash("127.0.0.1")
	if err := hc.SolveChallenge(); err != nil {
		t.Fatalf("Failed to solve hashcash challenge: %s", err)
	}

	validHcString := hc.ToString()

	req := &Request{
		Request:  &api.Request{Cmd: api.Command_CHECK_SOLUTION, Data: validHcString},
		ClientIP: "127.0.0.1",
	}
	writer := &bytes.Buffer{}

	handler := NewHandler(nil, repo, zenRepo)
	handler.handleCheckSolution(context.Background(), writer, req)

	resp := &api.Response{
		Status:   0,
		Response: nil,
	}
	_ = proto.Unmarshal(writer.Bytes(), resp)

	assert.Equal(t, api.Response_SUCCESS, resp.Status)
	assert.Equal(t, "some random Zen quote", resp.GetData())
}
