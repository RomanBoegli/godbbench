package benchmark

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"text/template"
	"time"
)

// Bencher is the interface a benchmark has to impelement.
type Bencher interface {
	Setup()
	Cleanup(bool)
	Benchmarks() []Benchmark
	Exec(string)
}

// BenchType determines if the particular benchmark should be run several times or only once.
type BenchType int

const (
	// TypeLoop executes the benchmark several times.
	TypeLoop BenchType = iota
	// TypeOnce executes the benchmark once.
	TypeOnce BenchType = iota
)

// Benchmark contains the benchmark name, its db statement and its type.
type Benchmark struct {
	Name      string
	Type      BenchType
	IterRatio float64
	Parallel  bool
	Stmt      string
}

// Result encapsulates the metrics of a benchmark run
type Result struct {
	Min                 time.Duration
	Max                 time.Duration
	TotalExecutionTime  time.Duration
	Start               time.Time
	End                 time.Time
	Duration            time.Duration
	TotalExecutionCount uint64
}

// Avg calculates the results average
func (r Result) Avg() time.Duration {
	if r.TotalExecutionCount == 0 {
		return 0
	}
	return time.Duration(int64(r.TotalExecutionTime) / int64(r.TotalExecutionCount))
}

// bencherExecutor is responsible for running the benchmark, keeping track
// of metrics as the execution goes
type bencherExecutor struct {
	result Result
	mux    sync.Mutex
}

// Run executes the benchmark.
func Run(bencher Bencher, b Benchmark, iter, threads int) Result {
	t := template.New(b.Name)
	t, err := t.Parse(b.Stmt)
	if err != nil {
		log.Fatalf("failed to parse template: %v", err)
	}

	executor := bencherExecutor{
		result: Result{
			Start: time.Now(),
		},
	}

	switch b.Type {
	case TypeOnce:
		if b.Parallel {
			go executor.once(bencher, t)
		} else {
			executor.once(bencher, t)
		}
	case TypeLoop:
		_iter := int(math.Max((float64(iter) * b.IterRatio), 1.0))
		if b.Parallel {
			go executor.loop(bencher, t, _iter, threads)
		} else {
			executor.loop(bencher, t, _iter, threads)
		}
	}

	executor.result.End = time.Now()
	executor.result.Duration = time.Since(executor.result.Start)

	return executor.result
}

// loop runs the benchmark concurrently several times.
func (b *bencherExecutor) loop(bencher Bencher, t *template.Template, iterations, threads int) {
	wg := &sync.WaitGroup{}
	wg.Add(threads)
	defer wg.Wait()

	// start as many routines as specified
	for routine := 0; routine < threads; routine++ {
		// calculate the amount of iterations to execute in this routine
		from := ((iterations / threads) * routine) + 1
		to := (iterations / threads) * (routine + 1)

		// Add the remainder of iterations to the last routine.
		if routine == threads-1 {
			remainder := iterations - to
			to += remainder
		}

		// start the routine
		go func(gofrom, togo int) {
			defer wg.Done()
			// notify channel for SIGINT (ctrl-c)
			sigchan := make(chan os.Signal, 1)
			signal.Notify(sigchan, os.Interrupt)

			for i := gofrom; i <= togo; i++ {
				select {
				case <-sigchan:
					// got SIGINT, stop benchmarking
					return
				default:
					// build and execute the statement
					stmt := buildStmt(t, i)
					now := time.Now()
					bencher.Exec(stmt)
					b.collectStats(now)
				}
			}
		}(from, to)
	}
}

func (b *bencherExecutor) collectStats(start time.Time) {
	b.mux.Lock()
	defer b.mux.Unlock()

	b.result.TotalExecutionCount++

	durTime := time.Since(start)

	b.result.TotalExecutionTime += durTime

	if durTime > b.result.Max {
		b.result.Max = durTime
	}

	if durTime < b.result.Min || b.result.Min == 0 {
		b.result.Min = durTime
	}
}

// once runs the benchmark a single time.
func (b *bencherExecutor) once(bencher Bencher, t *template.Template) {
	stmt := buildStmt(t, 1)
	defer b.collectStats(time.Now())
	bencher.Exec(stmt)
}

// buildStmt parses the given template with variables and functions to a pure DB statement.
func buildStmt(t *template.Template, i int) string {
	sb := &strings.Builder{}

	data := struct {
		Iter             int
		RandIntBetween   func(int, int) int
		RandFloatBetween func(float64, float64) float64
		RandId           func(string, string) string
		Seed             func(int64)
		RandInt63        func() int64
		RandInt63n       func(int64) int64
		RandFloat32      func() float32
		RandFloat64      func() float64
		RandExpFloat64   func() float64
		RandNormFloat64  func() float64
		RandString       func(int, int) string
		RandDate         func() string
	}{
		Iter:             i,
		RandIntBetween:   RandInt,
		RandFloatBetween: RandFloat64Between,
		RandId:           GetRandId,
		Seed:             rand.Seed,
		RandInt63:        rand.Int63,
		RandInt63n:       rand.Int63n,
		RandFloat32:      rand.Float32,
		RandFloat64:      rand.Float64,
		RandExpFloat64:   rand.ExpFloat64,
		RandNormFloat64:  rand.NormFloat64,
		RandString:       RandStringBytes,
		RandDate:         RandDate,
	}
	if err := t.Execute(sb, data); err != nil {
		log.Fatalf("failed to execute template: %v", err)
	}
	return sb.String()
}

func RandInt(min int, max int) int {
	return rand.Intn(max-min) + min
}

func RandFloat64Between(min float64, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func GetRandId(entity string, language string) string {
	switch strings.ToLower(language) {
	case "mysql":
		return fmt.Sprintf("(SELECT %sId FROM GoBench.%s ORDER BY RAND() LIMIT 1)", entity, entity)
	case "postgres":
		return fmt.Sprintf("(SELECT %sId FROM GoBench.%s ORDER BY RANDOM() LIMIT 1)", entity, entity)
	case "cypher":
		return fmt.Sprintf("(MATCH (x:%s) RETURN x.%sId, rand() as rand ORDER BY rand ASC LIMIT 1)", entity, entity)
	default:
		return "1"
	}

}

func RandStringBytes(min int, max int) string {
	var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	n := rand.Intn(max-min) + min
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func RandDate() string {
	min := time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2023, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min
	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0).Format("2006-01-02")
}
