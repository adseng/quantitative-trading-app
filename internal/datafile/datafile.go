package datafile

import (
	"encoding/json"
	"os"

	"quantitative-trading-app/internal/factor"
)

const DefaultPath = "docs/btcusdt100w.txt"

// LoadKlines reads klines from the JSONL data file.
func LoadKlines(path string) ([]*factor.KLine, error) {
	if path == "" {
		path = DefaultPath
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var klines []*factor.KLine
	dec := json.NewDecoder(f)
	for dec.More() {
		var kl factor.KLine
		if err := dec.Decode(&kl); err != nil {
			return nil, err
		}
		klines = append(klines, &kl)
	}
	return klines, nil
}
