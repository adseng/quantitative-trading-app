package datafile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"quantitative-trading-app/internal/market"
)

const DefaultDir = "docs/db"

// LoadKlines reads klines from the JSONL data file.
func LoadKlines(path string) ([]*market.KLine, error) {
	if path == "" {
		path = DefaultKlinePath("15m", 10000)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var klines []*market.KLine
	dec := json.NewDecoder(f)
	for dec.More() {
		var kl market.KLine
		if err := dec.Decode(&kl); err != nil {
			return nil, err
		}
		klines = append(klines, &kl)
	}
	return klines, nil
}

func SaveKlines(path string, klines []*market.KLine) error {
	if path == "" {
		return fmt.Errorf("empty output path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, kl := range klines {
		if err := enc.Encode(kl); err != nil {
			return err
		}
	}
	return nil
}

func DefaultKlinePath(interval string, count int) string {
	if interval == "" {
		interval = "15m"
	}
	if count <= 0 {
		count = 10000
	}
	return filepath.ToSlash(filepath.Join(DefaultDir, fmt.Sprintf("data-k-%s-%d.txt", interval, count)))
}
