package pow

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSolveChallenge(t *testing.T) {
	t.Parallel()

	hc, _ := NewHashcash("test")

	err := hc.SolveChallenge()
	assert.Nil(t, err, "Expected no error in solving the challenge")

	valid := hc.ValidateSolution()
	assert.True(t, valid, "Expected solution to be valid")
}

func TestValidateSolution(t *testing.T) {
	t.Parallel()

	hashcash, _ := NewHashcash("test")

	hashcash.Counter = 123
	if invalid := hashcash.ValidateSolution(); invalid {
		t.Errorf("Expected solution to be invalid, but got valid")
	}

	err := hashcash.SolveChallenge()
	assert.Nil(t, err, "Expected no error in solving the challenge")

	valid := hashcash.ValidateSolution()
	if !valid {
		t.Errorf("Expected solution to be valid, but got invalid")
	}
}

func TestNewHashcashFromString(t *testing.T) {
	t.Parallel()

	hcStr := "1:20:1625075186:test::123456:0"

	hashcash, err := NewHashcashFromString(hcStr)
	assert.Nil(t, err, "Expected no error in creating Hashcash from string")

	assert.Equal(t, 1, hashcash.Version)
	assert.Equal(t, 20, hashcash.Bits)
	assert.Equal(t, time.Unix(1625075186, 0), hashcash.Date)
	assert.Equal(t, "test", hashcash.Resource)
	assert.Equal(t, 123456, int(hashcash.Rand))
	assert.Equal(t, 0, hashcash.Counter)
}

func TestToString(t *testing.T) {
	t.Parallel()

	hashcash := Hashcash{
		Version:  1,
		Bits:     20,
		Date:     time.Unix(1625075186, 0),
		Resource: "test",
		Ext:      "",
		Rand:     123456,
		Counter:  0,
	}

	expected := "1:20:1625075186:test::123456:0"
	actual := hashcash.ToString()

	assert.Equal(t, expected, actual, "Expected string representation to match")
}

func TestNewRandomForPOW(t *testing.T) {
	t.Parallel()

	rand1, err1 := newRandomForPOW()
	rand2, err2 := newRandomForPOW()

	if err1 != nil {
		t.Errorf("newRandomForPOW() should not return an error, got: %v", err1)
	}

	if err2 != nil {
		t.Errorf("newRandomForPOW() should not return an error, got: %v", err2)
	}

	if rand1.Cmp(rand2) == 0 {
		t.Errorf("newRandomForPOW() should generate unique random numbers, got: %v and %v", rand1, rand2)
	}

	zero := big.NewInt(0)
	if rand1.Cmp(zero) < 0 || rand1.Cmp(big.NewInt(hashcashRandInter)) >= 0 {
		t.Errorf("newRandomForPOW() should generate a number in range [0, hashcashRandInter), got: %v", rand1)
	}

	if rand2.Cmp(zero) < 0 || rand2.Cmp(big.NewInt(hashcashRandInter)) >= 0 {
		t.Errorf("newRandomForPOW() should generate a number in range [0, hashcashRandInter), got: %v", rand2)
	}
}
