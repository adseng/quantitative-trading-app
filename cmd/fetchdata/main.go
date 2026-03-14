package main

import (
	"flag"
	"fmt"
	"os"

	"quantitative-trading-app/internal/binance"
	"quantitative-trading-app/internal/config"
	"quantitative-trading-app/internal/datafile"
)

func main() {
	_ = config.Load()
	binance.InitClient()

	symbol := flag.String("symbol", config.Get(config.KeySymbol, "BTCUSDT"), "trading symbol")
	interval := flag.String("interval", "15m", "kline interval, e.g. 1m/5m/15m/1h")
	count := flag.Int("count", 10000, "number of klines to fetch")
	out := flag.String("out", "", "output file path")
	force := flag.Bool("force", false, "overwrite output file when it already exists")
	flag.Parse()

	outFile := *out
	if outFile == "" {
		outFile = datafile.DefaultKlinePath(*interval, *count)
	}

	if info, err := os.Stat(outFile); err == nil && info.Size() > 0 && !*force {
		fmt.Printf("%s already exists (%d bytes), skip fetching.\n", outFile, info.Size())
		fmt.Println("Use -force to overwrite the file.")
		return
	}

	fmt.Printf("Fetching %s %s klines: %d bars\n", *symbol, *interval, *count)

	klines, err := binance.FetchKlines(*symbol, *interval, int64(*count))
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Fetched %d klines, writing to %s ...\n", len(klines), outFile)
	if err := datafile.SaveKlines(outFile, klines); err != nil {
		fmt.Fprintf(os.Stderr, "write error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Done. %d klines saved to %s\n", len(klines), outFile)
}
