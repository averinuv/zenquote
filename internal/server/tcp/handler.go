package tcp

import (
	"context"
	"io"
	"time"
	"zenquote/api"
	"zenquote/internal/pow"

	"google.golang.org/protobuf/proto"

	"go.uber.org/zap"
)

const (
	hashcashStoreTTL = 30 * time.Minute // TTL for the hashcash data in the repository
)

type HashcashRepo interface {
	Store(ctx context.Context, resource string, hashcash string, ttl time.Duration) error
	Get(ctx context.Context, resource string) (string, error)
	Delete(ctx context.Context, resource string) error
}

type ZenquoteRepo interface {
	GetRandom(ctx context.Context) (string, error)
}

type Handler struct {
	logger       *zap.Logger
	repo         HashcashRepo
	zenquoteRepo ZenquoteRepo
}

func NewHandler(logger *zap.Logger, store HashcashRepo, zenquoteRepo ZenquoteRepo) *Handler {
	return &Handler{logger: logger, repo: store, zenquoteRepo: zenquoteRepo}
}

func (h *Handler) Handle(ctx context.Context, respWriter io.Writer, req *Request) {
	switch req.GetCmd() {
	case api.Command_GET_CHALLENGE:
		h.handleGetChallenge(ctx, respWriter, req)
	case api.Command_CHECK_SOLUTION:
		h.handleCheckSolution(ctx, respWriter, req)
	default:
		h.respondWithErr(respWriter, "unknown command", zap.String("cmd", req.GetCmd().String()))
	}
}

// Generate a Proof of Work challenge.
func (h *Handler) handleGetChallenge(ctx context.Context, respWriter io.Writer, req *Request) {
	hashcash, err := pow.NewHashcash(req.ClientIP)
	if err != nil {
		h.respondWithErr(respWriter, "new hashcash failed", zap.String("clientIP", req.ClientIP))

		return
	}

	err = h.repo.Store(ctx, req.ClientIP, hashcash.ToString(), hashcashStoreTTL)
	if err != nil {
		h.respondWithErr(respWriter, "repo store failed")

		return
	}

	challenge := hashcash.ToString()
	h.respondWithSuccess(respWriter, challenge)
}

// Check solution Proof of Work challenge and return zen quote.
func (h *Handler) handleCheckSolution(ctx context.Context, respWriter io.Writer, req *Request) {
	// validate the request by checking for hashcash in repo
	hcStr, err := h.repo.Get(ctx, req.ClientIP)
	if err != nil || len(hcStr) == 0 {
		h.respondWithErr(respWriter, "no hashcash found", zap.Error(err), zap.Any("req", req))

		return
	}

	// create Hashcash from received string
	hashcash, err := pow.NewHashcashFromString(req.GetData())
	if err != nil {
		h.respondWithErr(respWriter, "new hashcash from str failed", zap.Error(err), zap.Any("req", req))

		return
	}

	// validate solution
	if !hashcash.ValidateSolution() {
		h.respondWithErr(respWriter, "challenge solution invalid", zap.Any("req", req))
	}

	// remove the hashcash from the cache
	if err = h.repo.Delete(ctx, req.ClientIP); err != nil {
		h.logger.Error("remove hashcash from storage failed", zap.Error(err), zap.Any("req", req))
	}

	// send zen quote
	quote, err := h.zenquoteRepo.GetRandom(ctx)
	if err != nil {
		h.respondWithErr(respWriter, "get random zen quote failed", zap.Error(err), zap.Any("req", req))

		return
	}

	h.respondWithSuccess(respWriter, quote)
}

func (h *Handler) respondWithSuccess(respWriter io.Writer, msg string) {
	response := &api.Response{
		Status: api.Response_SUCCESS,
		Response: &api.Response_Data{
			Data: msg,
		},
	}

	data, err := proto.Marshal(response)
	if err != nil {
		h.logger.Error("failed to marshal response", zap.Error(err), zap.String("msg", msg))

		return
	}

	data = append(data, '\n')
	if _, err := respWriter.Write(data); err != nil {
		h.logger.Error("failed to write response", zap.Error(err), zap.String("msg", msg))

		return
	}
}

func (h *Handler) respondWithErr(respWriter io.Writer, msg string, logData ...zap.Field) {
	h.logger.Error(msg, logData...)

	response := &api.Response{
		Status: api.Response_FAILURE,
		Response: &api.Response_Error{
			Error: msg,
		},
	}

	data, err := proto.Marshal(response)
	if err != nil {
		h.logger.Error("failed to marshal error response", zap.Error(err))

		return
	}

	data = append(data, '\n')
	if _, err = respWriter.Write(data); err != nil {
		h.logger.Error("failed to write error response", zap.Error(err))

		return
	}
}
