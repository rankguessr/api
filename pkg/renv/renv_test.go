package renv

import (
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	err := os.Setenv("PORT", "8080")
	if err != nil {
		t.Fatalf("failed to set environment variable: %v", err)
	}

	type Config struct {
		Port string `env:"PORT,required"`
	}

	cfg := &Config{}
	err = Parse(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Fatalf("expected PORT to be '8080', got '%s'", cfg.Port)
	}
}

func TestMissing(t *testing.T) {
	os.Unsetenv("PORT")
	type Config struct {
		Port string `env:"PORT,required"`
	}

	cfg := &Config{}
	err := Parse(cfg)
	if err == nil {
		t.Fatal("expected error for missing required variable")
	}

	if err.Error() != "PORT environment variable is required but not set" {
		t.Fatalf("unexpected error: %v", err)
	}
}
