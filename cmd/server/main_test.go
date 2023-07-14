package main

import (
	"testing"

	"go.uber.org/fx"
)

func TestApp(t *testing.T) {
	err := fx.ValidateApp(options...)
	if err != nil {
		t.Fatalf("error starting app: %v", err)
	}
}
