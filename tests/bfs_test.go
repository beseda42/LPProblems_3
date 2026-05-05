package tests

import (
	"LPProblems/eka"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindMaxFlow_InputFiles(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "qpbo_problem_3",
			path: filepath.Join("..", "tests", "input", "qpbo_problem_3.cleanrescale.bq.max"),
		},
		{
			name: "qpbo_problem_15",
			path: filepath.Join("..", "tests", "input", "qpbo_problem_15.cleanrescale.bq.max"),
		},
		{
			name: "qpbo_problem_99",
			path: filepath.Join("..", "tests", "input", "qpbo_problem_99.cleanrescale.bq.max"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if _, err := os.Stat(tc.path); err != nil {
				t.Skipf("fixture is not available: %v", err)
			}

			maxFlow, diff, err := eka.FindMaxFlow(tc.path)
			if err != nil {
				t.Fatalf("FindMaxFlow(%q): %v", tc.path, err)
			}

			if maxFlow < 0 {
				t.Fatalf("maxFlow = %d, want non-negative", maxFlow)
			}

			outDir := "output"
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				t.Fatalf("mkdir output: %v", err)
			}
			outName := filepath.Base(tc.path) + ".txt"
			outPath := filepath.Join(outDir, outName)
			out := formatMaxFlowOutput(maxFlow, diff)
			if err := os.WriteFile(outPath, []byte(out), 0o644); err != nil {
				t.Fatalf("write %q: %v", outPath, err)
			}

			t.Logf("maxFlow for %s: %d", tc.name, maxFlow)
		})
	}
}

// formatMaxFlowOutput — первая строка: max flow; далее по строке на каждую дугу u→v с ненулевой
// разностью capacity (начальный граф − остаточный): "u v значение", как 0 5 -1480.
func formatMaxFlowOutput(maxFlow int, diff [][]int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "max flow: %d\n", maxFlow)
	n := len(diff)
	for u := 0; u < n; u++ {
		for v := 0; v < n; v++ {
			d := diff[u][v]
			if d == 0 {
				continue
			}
			fmt.Fprintf(&b, "%d %d %d\n", u, v, d)
		}
	}
	return b.String()
}
