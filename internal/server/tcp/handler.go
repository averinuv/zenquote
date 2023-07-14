package tcp

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"time"

	"go.uber.org/zap"

	"zenquote/internal/pow"
	storage "zenquote/internal/redis"
)

const (
	getChallengeCmd   = "get_challenge"
	checkSolutionCmd  = "check_solution"
	hashcashStoreTtl  = 30 * time.Minute // TTL for the hashcash data in the repository
	hashcashRandInter = 1 << 30
)

var zenQuotes = []string{
	"Do not dwell in the past, do not dream of the future, concentrate the mind on the present moment.",
	"Do not let the behavior of others destroy your inner peace.",
	"The trouble is, you think you have time.",
	"Peace comes from within. Do not seek it without.",
	"Your work is to discover your world and then with all your heart give yourself to it.",
}

type Request struct {
	Cmd  string `json:"cmd"`
	Data string `json:"data,omitempty"`
}

// KeyValueRepo repository for storing hashcash data
type KeyValueRepo interface {
	Store(ctx context.Context, key string, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

type Handler struct {
	logger *zap.Logger
	repo   KeyValueRepo
}

func NewHandler(logger *zap.Logger, store *storage.RedisStorage) *Handler {
	return &Handler{logger: logger, repo: store}
}

func (h *Handler) Handle(ctx context.Context, conn net.Conn, reqBytes []byte) {
	remoteAddr, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	var r Request
	err := json.Unmarshal(reqBytes, &r)
	if err != nil {
		h.logger.Error("unmarshal request failed", zap.ByteString("reqBody", reqBytes))
		return
	}

	switch r.Cmd {
	// generate a Proof of Work challenge
	case getChallengeCmd:
		hcRand := rand.Intn(hashcashRandInter)
		hc := pow.NewHashcash(remoteAddr, hcRand)

		hcBytes, err := json.Marshal(hc)
		if err != nil {
			h.logger.Error("marshal hashcash failed", zap.String("remoteAddr", remoteAddr), zap.Error(err))
			return
		}

		err = h.repo.Store(ctx, remoteAddr, string(hcBytes), hashcashStoreTtl)
		if err != nil {
			h.logger.Error("repo store failed", zap.String("remoteAddr", remoteAddr), zap.Error(err))
			return
		}

		challenge := hc.ToString()
		_, err = conn.Write([]byte(challenge + "\n"))
		if err != nil {
			h.logger.Error("repo store failed", zap.String("remoteAddr", remoteAddr), zap.Error(err))
			return
		}

		h.logger.Info(
			"Challenge sent",
			zap.String("challenge", challenge),
			zap.Any("hashcash", hc),
		)

	// check solution Proof of Work challenge and return zen quote
	case checkSolutionCmd:
		h.logger.Info("check solution request", zap.String("remoteAddr", remoteAddr), zap.Any("req", r))

		// validate the request by checking for hashcash in repo
		hcStr, err := h.repo.Get(ctx, remoteAddr)
		if err != nil || len(hcStr) == 0 {
			h.logger.Error("no hashcash found in repo",
				zap.Error(err),
				zap.String("remoteAddr", remoteAddr),
				zap.Any("req", r),
			)
			return
		}

		// create Hashcash from received string
		hc, err := pow.NewHashcashFromString(r.Data)
		if err != nil {
			h.logger.Error(
				"unmarshal req hashcash failed",
				zap.Error(err),
				zap.String("remoteAddr", remoteAddr),
				zap.Any("req", r),
			)
			return
		}

		// validate solution
		if hc.ValidateSolution() {
			h.logger.Info(
				"Correct PoW response received",
				zap.String("remoteAddr", remoteAddr),
				zap.Any("request", r),
			)

			// send zen quote message after correct PoW response
			zenQuote := zenQuotes[rand.Intn(len(zenQuotes))]
			zenQuoteResp := []byte(fmt.Sprintf("%s\n", zenQuote))
			_, err = conn.Write(zenQuoteResp)
			if err != nil {
				h.logger.Error("write zen quote failed", zap.String("remoteAddr", remoteAddr), zap.Error(err))
				return
			}

			// remove the hashcash from the cache after successful response
			err = h.repo.Delete(ctx, remoteAddr)
			if err != nil {
				h.logger.Error(
					"remove hashcash from storage failed",
					zap.String("remoteAddr", remoteAddr),
					zap.Error(err),
				)
			}
		} else {
			h.logger.Info(
				"Incorrect PoW response received",
				zap.String("remoteAddr", remoteAddr),
				zap.Any("request", r),
			)

			hc.Counter++
		}
	default:
		h.logger.Error("Unknown cmd", zap.String("cmd", r.Cmd))
	}
}
