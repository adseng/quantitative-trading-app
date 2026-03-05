package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type row struct {
	id       int
	name     string
	sigAcc   float64
	signal   int
	total    int
	correct  int
	avgScore float64
	strength float64
}

func main() {
	path := findLatestExcel()
	if path == "" {
		fmt.Println("No batch_test_log*.xlsx found in docs/test/")
		os.Exit(1)
	}
	fmt.Printf("Reading: %s\n\n", path)

	f, err := excelize.OpenFile(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	rows, _ := f.GetRows("Sheet1")
	fmt.Printf("Total rows: %d (including header)\n\n", len(rows))
	if len(rows) <= 1 {
		fmt.Println("No data rows.")
		return
	}

	header := rows[0]
	idx := map[string]int{}
	for i, h := range header {
		idx[h] = i
	}

	get := func(r []string, col string) string {
		i, ok := idx[col]
		if !ok || i >= len(r) {
			return ""
		}
		return r[i]
	}

	var data []row
	for i := 1; i < len(rows); i++ {
		r := rows[i]
		id, _ := strconv.Atoi(get(r, "ID"))
		sigAccStr := strings.TrimSuffix(get(r, "信号正确率"), "%")
		sigAcc, _ := strconv.ParseFloat(sigAccStr, 64)
		signal, _ := strconv.Atoi(get(r, "有效预测"))
		total, _ := strconv.Atoi(get(r, "总数"))
		correct, _ := strconv.Atoi(get(r, "正确数"))
		avgScore, _ := strconv.ParseFloat(get(r, "平均净分"), 64)
		strength, _ := strconv.ParseFloat(get(r, "信号强度"), 64)

		data = append(data, row{
			id: id, name: get(r, "名称"),
			sigAcc: sigAcc, signal: signal, total: total, correct: correct,
			avgScore: avgScore, strength: strength,
		})
	}

	sort.Slice(data, func(i, j int) bool { return data[i].sigAcc > data[j].sigAcc })

	printSection := func(title string, list []row, max int) {
		fmt.Printf("\n=== %s ===\n", title)
		for i := 0; i < max && i < len(list); i++ {
			r := list[i]
			rate := 0.0
			if r.total > 0 {
				rate = float64(r.signal) / float64(r.total) * 100
			}
			fmt.Printf("#%-3d %-40s SigAcc=%5.1f%%  Signal=%6d(%4.1f%%)  Str=%5.3f  Avg=%+.3f\n",
				r.id, r.name, r.sigAcc, r.signal, rate, r.strength, r.avgScore)
		}
	}

	printSection("TOP 30 by Signal Accuracy", data, 30)

	var f10, f30 []row
	for _, r := range data {
		rate := float64(r.signal) / float64(r.total)
		if rate > 0.10 {
			f10 = append(f10, r)
		}
		if rate > 0.30 {
			f30 = append(f30, r)
		}
	}
	printSection("TOP 30 (signal rate > 10%)", f10, 30)
	printSection("TOP 30 (signal rate > 30%)", f30, 30)

	fmt.Printf("\n=== BOTTOM 10 ===\n")
	for i := len(data) - 1; i >= 0 && i >= len(data)-10; i-- {
		r := data[i]
		rate := float64(r.signal) / float64(r.total) * 100
		fmt.Printf("#%-3d %-40s SigAcc=%5.1f%%  Signal=%6d(%4.1f%%)\n",
			r.id, r.name, r.sigAcc, r.signal, rate)
	}

	fmt.Printf("\n=== ALL (%d results, sorted by SigAcc desc) ===\n", len(data))
	for _, r := range data {
		rate := float64(r.signal) / float64(r.total) * 100
		fmt.Printf("#%-3d %-40s SigAcc=%5.1f%%  Signal=%6d(%4.1f%%)  Str=%5.3f  Avg=%+.3f\n",
			r.id, r.name, r.sigAcc, r.signal, rate, r.strength, r.avgScore)
	}
}

func findLatestExcel() string {
	dir := filepath.Join("docs", "test")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	latest := ""
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "batch_test_log") && strings.HasSuffix(name, ".xlsx") {
			if name > latest {
				latest = name
			}
		}
	}
	if latest == "" {
		return ""
	}
	return filepath.Join(dir, latest)
}
