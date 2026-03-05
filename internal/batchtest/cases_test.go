package batchtest

import (
	"fmt"
	"testing"
)

func TestGenerateTestCases(t *testing.T) {
	cases := GenerateTestCases()
	fmt.Printf("Generated %d test cases\n", len(cases))
	if len(cases) != 200 {
		t.Errorf("expected 200 cases, got %d", len(cases))
	}
	seen := make(map[int]bool)
	for _, c := range cases {
		if seen[c.ID] {
			t.Errorf("duplicate ID: %d", c.ID)
		}
		seen[c.ID] = true
	}
}
