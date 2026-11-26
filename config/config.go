package config

/*
This file was written with AI assistance (Claude Sonnet 5.5)
*/

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const (
	defaultFilePort = 4000
	defaultControlPort = 4001
)

type Config struct {
	FilePort    int
	ControlPort int
}

// name - of parameter; def - default
func portFromEnv(name string, def int) (int, error) {
	v := os.Getenv(name)
	if v == "" {
		return def, nil
	}

	p, err := strconv.Atoi(v)
	if err != nil || p < 1 || p > 65535 {
		return 0, fmt.Errorf("%s must be an integer between 1 and 65535, got %q", name, v)
	}
	return p, nil
}

// Load reads the env file (if present) and returns validated settings.
// Variables already set in the real environment take precedence over the file.
func Load(envFile string) (Config, error) {
	// A missing file is fine: defaults and real environment variables still apply
	if err := godotenv.Load(envFile); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Config{}, fmt.Errorf("reading %s: %w", envFile, err)
	}

	filePort, err := portFromEnv("FILE_PORT", defaultFilePort)
	if err != nil {
		return Config{}, err
	}

	controlPort, err := portFromEnv("CONTROL_PORT", defaultControlPort)
	if err != nil {
		return Config{}, err
	}

	if filePort == controlPort {
		return Config{}, fmt.Errorf("FILE_PORT and CONTROL_PORT must differ (both are %d)", filePort)
	}

	return Config{FilePort: filePort, ControlPort: controlPort}, nil
}
