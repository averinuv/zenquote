package pow

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	difficulty    = 3
	version       = 1
	maxIterations = 1 << 30
)

var ErrMaxIterationsExceeded = errors.New("maximum number of iterations exceeded")

type Hashcash struct {
	Version  int
	Bits     int
	Date     time.Time
	Resource string
	Ext      string
	Rand     int
	Counter  int
}

func NewHashcash(resource string, rand int) *Hashcash {
	return &Hashcash{
		Version:  version,
		Bits:     difficulty,
		Date:     time.Now().UTC(),
		Resource: resource,
		Rand:     rand,
		Counter:  0,
	}
}

// NewHashcashFromString parses the input string and returns the created Hashcash structure
// or an error if the input format is invalid.
// The string should have the format "version:bits:date:resource:rand:counter".
func NewHashcashFromString(hcStr string) (*Hashcash, error) {
	parts := strings.Split(hcStr, ":")
	if len(parts) < 6 {
		return nil, errors.New("invalid input format")
	}

	v, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, errors.New("invalid version")
	}

	bits, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, errors.New("invalid bits")
	}

	dateTimestamp, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, errors.New("invalid date")
	}
	date := time.Unix(dateTimestamp, 0)

	rand, err := strconv.Atoi(parts[4])
	if err != nil {
		return nil, errors.New("invalid rand")
	}

	counter, err := strconv.Atoi(parts[5])
	if err != nil {
		return nil, errors.New("invalid counter")
	}

	hashcash := &Hashcash{
		Version:  v,
		Bits:     bits,
		Date:     date,
		Resource: parts[3],
		Rand:     rand,
		Counter:  counter,
	}

	return hashcash, nil
}

// ToString returns a string representation of the Hashcash structure.
// Example:
//
//	"1:20:1625075186:example.com:123456:0"
func (h *Hashcash) ToString() string {
	return fmt.Sprintf("%d:%d:%d:%s:%d:%d",
		h.Version,
		h.Bits,
		h.Date.Unix(),
		h.Resource,
		h.Rand,
		h.Counter,
	)
}

// ValidateSolution checks if the solution for the Hashcash challenge is valid.
// It constructs the challenge by combining the Hashcash and counter, calculates the hash,
// and checks if it has the required leading zeros. Returns true if valid; otherwise, false.
func (h *Hashcash) ValidateSolution() bool {
	challenge := fmt.Sprintf("%s%d", h.ToString(), h.Counter)
	hash := sha256.Sum256([]byte(challenge))
	hashStr := hex.EncodeToString(hash[:])
	return strings.HasPrefix(hashStr, strings.Repeat("0", difficulty))
}

// SolveChallenge tries to find a valid solution for the challenge by incrementing the counter and computing the hash.
// Returns nil if a valid solution is found
// or ErrMaxIterationsExceeded if the maximum number of iterations is reached without finding a solution.
func (h *Hashcash) SolveChallenge() error {
	for i := 0; i < maxIterations; i++ {
		hash := sha256.Sum256([]byte(h.ToString() + strconv.Itoa(h.Counter)))
		hashStr := hex.EncodeToString(hash[:])
		if strings.HasPrefix(hashStr, strings.Repeat("0", difficulty)) {
			return nil
		}
		h.Counter++
	}
	return ErrMaxIterationsExceeded
}
