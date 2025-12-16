package node

import (
	"encoding/json"
	"os"
	"time"
)

// Config defines your blockchain configuration
type Config struct {
	ChainID          string    `json:"chain_id"`
	EpochDurationSec int       `json:"epoch_duration_sec"`
	InitialSupply    int       `json:"initial_supply"`
	Timestamp        time.Time `json:"timestamp"`
}

// LoadConfig loads config from a JSON file
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
