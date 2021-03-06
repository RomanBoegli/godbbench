package benchmark

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var (
	// ErrNoMode is raised when tehre is no mode token after \benchmark.
	ErrNoMode = errors.New("failed to parse \\benchmark line, missing mode")
	// ErrNoName is raised when there is no token after \name.
	ErrNoName = errors.New("missing name after \\name token")
)

// Helper function to determine the benchmark name.
func getName(benchmark Benchmark, start, line int) string {
	switch benchmark.Type {
	case TypeLoop:
		if benchmark.Name != "" {
			return "(loop) " + benchmark.Name
		}
		return fmt.Sprintf("(loop) line %v-%v", start, line-1)
	case TypeOnce:
		if benchmark.Name != "" {
			return "(once) " + benchmark.Name
		}
		return fmt.Sprintf("(once) line %v-%v", start, line-1)
	}
	return "" // shouldn't happen
}

// ParseScript parses a benchmark script and returns the benchmarks.
func ParseScript(r io.Reader) ([]Benchmark, error) {
	var (
		scanner    = bufio.NewScanner(r)
		loopStart  = 1             // line the current loop mode started
		lineN      = 1             // current line number
		benchmarks = []Benchmark{} // the result
		curBench   = Benchmark{Type: TypeLoop, Parallel: false, IterRatio: 1.0}
	)

	// Helper function to append a new loop benchmark
	flushLoop := func() {
		if curBench.Stmt != "" {
			curBench.Stmt = strings.TrimSuffix(curBench.Stmt, "\n")
			curBench.Name = getName(curBench, loopStart, lineN)
			benchmarks = append(benchmarks, curBench)

			// Start new empty benchmark
			curBench = Benchmark{IterRatio: 1.0}
		}
	}

	// Parse each line of the script file
	for ; scanner.Scan(); lineN++ {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines.
		if strings.HasPrefix(line, "--") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") || line == "" {
			continue
		}

		// Parse '\benchmark' command.
		if strings.HasPrefix(line, "\\benchmark") {
			tokens := strings.Split(line, " ")

			// remove '\benchmark' entry from tokens
			tokens = tokens[1:]

			if len(tokens) <= 0 {
				// line does only consist of the token '\benchmark', we need more info
				return []Benchmark{}, ErrNoMode
			}

			// parse benchmark mode 'once' or 'loop'
			switch tokens[0] {
			case "once":
				if curBench.Type == TypeLoop {
					flushLoop()
				}
				curBench.Type = TypeOnce
				loopStart = lineN + 1
			case "loop":
				flushLoop()
				curBench.Type = TypeLoop
				loopStart = lineN + 1
			default:
				return []Benchmark{}, fmt.Errorf("failed to parse mode, neither 'once' nor 'loop': %v", tokens[0])
			}

			if len(tokens) > 1 && !strings.HasPrefix(tokens[1], "\\") {
				// custom execution count ratio specified
				if curBench.Type == TypeLoop {
					ratio, _ := strconv.ParseFloat(tokens[1], 32)
					if ratio > 0.0 && ratio <= 1.0 {
						curBench.IterRatio = ratio
					}
				}
				tokens = tokens[2:]
			} else {
				tokens = tokens[1:]
			}

			// Parse remaining tokens
			for _, t := range tokens {
				// reminder: can't change 'tokens' inside the range, e.g. 'cutting' with tokens[2:]
				// so we have to iterate even the token after \name, which could be skipped otherwise.
				switch t {
				case "\\parallel":
					curBench.Parallel = true
				case "\\name":
					if len(tokens) < 2 {
						return []Benchmark{}, ErrNoName
					}
					curBench.Name = tokens[1]
				}
			}

			// don't append '\benchmark' line
			continue
		}

		// Neither a '\benchmark' nor '\name' command line.
		// Should be an SQL statement line.
		// Append the line either as benchmark type once or loop
		curBench.Stmt += line + "\n"
	}

	// reached the end of the file, append remaining loop statements to benchmark
	if curBench.Stmt != "" {
		curBench.Stmt = strings.TrimSuffix(curBench.Stmt, "\n")
		curBench.Name = getName(curBench, loopStart, lineN)
		benchmarks = append(benchmarks, curBench)
	}

	return benchmarks, nil
}
