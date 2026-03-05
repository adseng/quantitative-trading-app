package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"quantitative-trading-app/internal/binance"
	"quantitative-trading-app/internal/config"
	"quantitative-trading-app/internal/datafile"
)

func main() {
	_ = config.Load()
	binance.InitClient()

	outFile := datafile.DefaultPath

	if info, err := os.Stat(outFile); err == nil && info.Size() > 0 {
		fmt.Printf("%s already exists (%d bytes), skip fetching.\n", outFile, info.Size())
		fmt.Println("Delete the file to re-fetch.")
		return
	}

	fmt.Println("Fetching BTCUSDT 15m klines: 100 rounds × 1000 per round ...")

	cancel := make(chan struct{})
	klines, err := binance.FetchKlines("BTCUSDT", "15m", 0, &binance.FetchKlinesOpts{
		PerReq:  1000,
		Chunks:  100,
		DelayMs: 600,
		ProgressFn: func(round, totalRounds, fetched int) {
			fmt.Printf("  round %d/%d, total fetched: %d\n", round, totalRounds, fetched)
		},
		CancelCh: cancel,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Fetched %d klines, writing to %s ...\n", len(klines), outFile)

	_ = os.MkdirAll(filepath.Dir(outFile), 0o755)

	f, err := os.Create(outFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create file error: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, kl := range klines {
		if err := enc.Encode(kl); err != nil {
			fmt.Fprintf(os.Stderr, "write error: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Done. %d klines saved to %s\n", len(klines), outFile)
}
