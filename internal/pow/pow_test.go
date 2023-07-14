package pow_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"zenquote/internal/pow"
)

func TestSolveChallenge(t *testing.T) {
	hc := pow.NewHashcash("test", 123456)

	err := hc.SolveChallenge()
	assert.Nil(t, err, "Expected no error in solving the challenge")

	valid := hc.ValidateSolution()
	assert.True(t, valid, "Expected solution to be valid")
}

func TestValidateSolution(t *testing.T) {
	hc := pow.NewHashcash("test", 123456)

	hc.Counter = 123
	invalid := hc.ValidateSolution()
	if invalid {
		t.Errorf("Expected solution to be invalid, but got valid")
	}

	err := hc.SolveChallenge()
	assert.Nil(t, err, "Expected no error in solving the challenge")

	valid := hc.ValidateSolution()
	if !valid {
		t.Errorf("Expected solution to be valid, but got invalid")
	}
}

func TestNewHashcashFromString(t *testing.T) {
	hcStr := "1:20:1625075186:test:123456:0"

	hc, err := pow.NewHashcashFromString(hcStr)
	assert.Nil(t, err, "Expected no error in creating Hashcash from string")

	assert.Equal(t, 1, hc.Version)
	assert.Equal(t, 20, hc.Bits)
	assert.Equal(t, time.Unix(1625075186, 0), hc.Date)
	assert.Equal(t, "test", hc.Resource)
	assert.Equal(t, 123456, hc.Rand)
	assert.Equal(t, 0, hc.Counter)
}

func TestToString(t *testing.T) {
	hc := pow.Hashcash{
		Version:  1,
		Bits:     20,
		Date:     time.Unix(1625075186, 0),
		Resource: "test",
		Rand:     123456,
		Counter:  0,
	}

	expected := "1:20:1625075186:test:123456:0"
	actual := hc.ToString()

	assert.Equal(t, expected, actual, "Expected string representation to match")
}
