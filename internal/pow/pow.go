package pow

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// Constants related to hashcash.
const (
	difficulty        = 3       // difficulty level for hashcash computation
	version           = 1       // hashcash ver
	maxIterations     = 1 << 30 // maximum number of iterations for solve challenge
	hcStringParts     = 6       // expected number of parts in a hashcash string
	hashcashRandInter = 1 << 30
)

// String indexes for hashcash parts.
const (
	strVersionIdx = iota
	strBitsIdx
	strDateIdx
	strResourceIdx
	_ // strExtIdx
	strRandIdx
	strCounterIdx
)

var (
	ErrMaxIterationsExceeded = errors.New("maximum number of iterations exceeded")
	ErrInvalidHashcashString = errors.New("invalid input format")
)

type Hashcash struct {
	Version  int
	Bits     int
	Date     time.Time
	Resource string
	Ext      string
	Rand     int64
	Counter  int
}

func NewHashcash(resource string) (*Hashcash, error) {
	hcRand, err := newRandomForPOW()
	if err != nil {
		return nil, fmt.Errorf("make random failed: %w", err)
	}

	return &Hashcash{
		Version:  version,
		Bits:     difficulty,
		Date:     time.Now().UTC(),
		Resource: resource,
		Ext:      "",
		Rand:     hcRand.Int64(),
		Counter:  0,
	}, nil
}

// NewRandomForPOW generates a random integer for use in the Hashcash proof-of-work algorithm.
// The random integer is used as part of the data that must be hashed in the proof-of-work challenge.
func newRandomForPOW() (*big.Int, error) {
	hcRand, err := rand.Int(rand.Reader, big.NewInt(hashcashRandInter))
	if err != nil {
		return nil, fmt.Errorf("make random int failed: %w", err)
	}

	return hcRand, nil
}

// NewHashcashFromString parses the input string and returns the created Hashcash structure
// or an error if the input format is invalid.
// The string should have the format "version:bits:date:resource:ext:rand:counter".
func NewHashcashFromString(hcStr string) (*Hashcash, error) {
	parts := strings.Split(hcStr, ":")
	if len(parts) < hcStringParts {
		return nil, fmt.Errorf("%w: expected at least 6 parts, got %d", ErrInvalidHashcashString, len(parts))
	}

	verPart := parts[strVersionIdx]
	bitsPart := parts[strBitsIdx]
	datePart := parts[strDateIdx]
	resourcePart := parts[strResourceIdx]
	randPart := parts[strRandIdx]
	counterPart := parts[strCounterIdx]

	ver, err := strconv.Atoi(verPart)
	if err != nil {
		return nil, fmt.Errorf("parse version %s failed: %w", verPart, err)
	}

	bits, err := strconv.Atoi(bitsPart)
	if err != nil {
		return nil, fmt.Errorf("parse bits %s failed: %w", bitsPart, err)
	}

	dateTimestamp, err := strconv.ParseInt(datePart, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse date %s failed: %w", datePart, err)
	}

	date := time.Unix(dateTimestamp, 0)

	hcRand, err := strconv.ParseInt(randPart, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse rand %s failed: %w", randPart, err)
	}

	counter, err := strconv.Atoi(counterPart)
	if err != nil {
		return nil, fmt.Errorf("parse counter %s failed: %w", counterPart, err)
	}

	hashcash := &Hashcash{
		Version:  ver,
		Bits:     bits,
		Date:     date,
		Resource: resourcePart,
		Ext:      "",
		Rand:     hcRand,
		Counter:  counter,
	}

	return hashcash, nil
}

// ToString returns a string representation of the Hashcash structure.
// String format "version:bits:date:resource:ext:rand:counter".
func (h *Hashcash) ToString() string {
	return fmt.Sprintf("%d:%d:%d:%s:%s:%d:%d",
		h.Version,
		h.Bits,
		h.Date.Unix(),
		h.Resource,
		h.Ext,
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
