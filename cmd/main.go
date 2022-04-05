package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/RomanBoegli/gobench/benchmark"
	"github.com/RomanBoegli/gobench/databases"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/pflag"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-gota/gota/dataframe"
)

var (
	hheaders = []string{"system", "multiplicity", "name", "executions", "total (μs)", "avg (μs)", "min (μs)", "max (μs)", "ops/s", "μs/op"}
)

func main() {
	var (
		// Default set of flags, available for all subcommands (benchmark options).
		defaultFlags = pflag.NewFlagSet("defaults", pflag.ExitOnError)
		iter         = defaultFlags.Int("iter", 1000, "how many iterations should be run")
		threads      = defaultFlags.Int("threads", 25, "max. number of green threads (iter >= threads > 0)")
		sleep        = defaultFlags.Duration("sleep", 0, "how long to pause after each single benchmark (valid units: ns, us, ms, s, m, h)")
		nosetup      = defaultFlags.Bool("nosetup", false, "initialize database and tables, e.g. when running own scripts")
		nocleanstart = defaultFlags.Bool("nocleanstart", false, "make a cleanup before setup")
		keep         = defaultFlags.Bool("keep", false, "keep benchmark data")
		runBench     = defaultFlags.String("run", "all", "only run the specified benchmarks, e.g. \"inserts deletes\"")
		scriptname   = defaultFlags.String("script", "", "custom sql file to execute")
		writecsv     = defaultFlags.String("writecsv", "", "write result to csv file")

		// Connection flags, applicable for most databases.
		connFlags = pflag.NewFlagSet("conn", pflag.ExitOnError)
		host      = connFlags.String("host", "localhost", "address of the server")
		port      = connFlags.Int("port", 0, "port of the server (0 -> db defaults)")
		user      = connFlags.String("user", "root", "user name to connect with the server")
		pass      = connFlags.String("pass", "root", "password to connect with the server")

		// Max. connections, applicable for most databases (not neo4j).
		maxconnsFlags = pflag.NewFlagSet("conns", pflag.ExitOnError)
		maxconns      = maxconnsFlags.Int("conns", 0, "max. number of open connections")

		// Flag sets for each database. DB specific flags are set in the switch statement below.
		mysqlFlags    = pflag.NewFlagSet("mysql", pflag.ExitOnError)
		postgresFlags = pflag.NewFlagSet("postgres", pflag.ExitOnError)
		neo4jFlags    = pflag.NewFlagSet("neo4j", pflag.ExitOnError)

		// Flags to merge result csv files
		mergeCsvFlags = pflag.NewFlagSet("mergecsv", pflag.ExitOnError)
		rootDir       = mergeCsvFlags.String("rootDir", "../tmp", "path to folder with csv files to be merged")
		targetFile    = mergeCsvFlags.String("targetFile", "../tmp/merged.csv", "target file path for merged csv")

		// Flags to generate charts
		createChartFlags = pflag.NewFlagSet("createcharts", pflag.ExitOnError)
		dataFile         = createChartFlags.String("dataFile", "../tmp/merged.csv", "path to source data file, assumes headers")
	)

	defaultFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Available subcommands:\n\tmysql | postgres | neo4j | mergecsv | createcharts\n")
		fmt.Fprintf(os.Stderr, "\tUse 'subcommand --help' for all flags of the specified command.\n")
		fmt.Fprintf(os.Stderr, "Generic flags for all subcommands:\n")
		defaultFlags.PrintDefaults()
	}

	// No comamnd given. Print usage help and exit.
	if len(os.Args) < 2 {
		defaultFlags.Usage()
		os.Exit(1)
	}

	var bencher benchmark.Bencher
	system := os.Args[1]

	switch system {
	case "postgres":
		postgresFlags.AddFlagSet(defaultFlags)
		postgresFlags.AddFlagSet(connFlags)
		postgresFlags.AddFlagSet(maxconnsFlags)
		if err := postgresFlags.Parse(os.Args[2:]); err != nil {
			log.Fatalf("failed to parse postgres flags: %v", err)
		}
		bencher = databases.NewPostgres(*host, *port, *user, *pass, *maxconns)
	case "mysql":
		mysqlFlags.AddFlagSet(defaultFlags)
		mysqlFlags.AddFlagSet(connFlags)
		mysqlFlags.AddFlagSet(maxconnsFlags)
		if err := mysqlFlags.Parse(os.Args[2:]); err != nil {
			log.Fatalf("failed to parse mysql flags: %v", err)
		}
		bencher = databases.NewMySQL(*host, *port, *user, *pass, *maxconns)
	case "neo4j":
		neo4jFlags.AddFlagSet(defaultFlags)
		neo4jFlags.AddFlagSet(connFlags)
		if err := neo4jFlags.Parse(os.Args[2:]); err != nil {
			log.Fatalf("failed to parse neo4j flags: %v", err)
		}
		bencher = databases.NewNeo4J(*host, *port, *user, *pass)
	case "mergecsv":
		if err := mergeCsvFlags.Parse(os.Args[2:]); err != nil {
			log.Fatalf("failed to parse postgres flags: %v", err)
		}
		MergeKnownCsv(*rootDir, *targetFile)
		os.Exit(0)
	case "createcharts":
		if err := createChartFlags.Parse(os.Args[2:]); err != nil {
			log.Fatalf("failed to parse postgres flags: %v", err)
		}
		CreateCharts(*dataFile)
		os.Exit(0)
	default:
		if err := defaultFlags.Parse(os.Args[1:]); err != nil {
			log.Fatalf("failed to parse default flags: %v", err)
		}

		// Command not recognized. Print usage help and exit.
		defaultFlags.Usage()
		os.Exit(1)
	}

	// clean old data when cleanstart flag is set
	if !*nocleanstart {
		bencher.Cleanup(false)
		fmt.Println("cleaned data")
		// os.Exit(0)
	}

	// setup database
	if !*nosetup {
		bencher.Setup()
	}

	// cleanup benchmark data when flag is not set
	if !*keep {
		defer bencher.Cleanup(true)
	}

	// we need at least one thread
	if *threads == 0 {
		*threads = 1
		fmt.Println("increased to 1 thread")
	}

	// can't have more threads than iterations
	if *threads > *iter {
		*threads = *iter
	}

	var benchmarks []benchmark.Benchmark

	// If a script was specified, overwrite built-in benchmarks.
	if *scriptname != "" {
		dat, err := ioutil.ReadFile(*scriptname)
		if err != nil {
			log.Fatalf("failed to read file: %v", err)
		}
		buf := bytes.NewBuffer(dat)
		benchmarks, err = benchmark.ParseScript(buf)
		if err != nil {
			log.Fatalf("failed to parse script: %v\n", err)
		}
	} else {
		// Otherwise use built-in benchmarks.
		benchmarks = bencher.Benchmarks()
	}

	// split benchmark names when "-run 'bench0 bench1 ...'" flag was used
	toRun := strings.Split(*runBench, " ")

	startTotal := time.Now()

	// notify channel for SIGINT (ctrl-c)
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)

	summary := [][]string{hheaders}

	for i, b := range benchmarks {
		select {
		case <-sigchan:
			// got SIGINT, stop benchmarking
			printTotal(startTotal)
			// using os.Exit(130) instead of return won't
			// run deferred funcs (e.g. b.Cleanup())
			return
		default:
			// check if we want to run this particular benchmark
			if !contains(toRun, "all") && !contains(toRun, b.Name) {
				continue
			}

			// run the particular benchmark
			results := benchmark.Run(bencher, b, *iter, *threads)

			// execution in ms for mode once
			msPerOp := results.Duration.Milliseconds()

			// execution in ns/op for mode loop
			if b.Type == benchmark.TypeLoop {
				msPerOp /= int64(*iter)
			}

			summary = append(summary, []string{
				system,
				fmt.Sprint(*iter),
				b.Name,
				fmt.Sprint(results.TotalExecutionCount),
				fmt.Sprint(results.Duration.Milliseconds()),
				fmt.Sprint(results.Avg().Milliseconds()),
				fmt.Sprint(results.Min.Milliseconds()),
				fmt.Sprint(results.Max.Milliseconds()),
				fmt.Sprint(float64(results.TotalExecutionCount) / (results.Duration.Seconds())),
				fmt.Sprint(msPerOp)})

			// Don't sleep after the last benchmark
			if i != len(benchmarks)-1 {
				time.Sleep(*sleep)
			}
		}
	}

	// write results to csv
	if *writecsv != "" {

		f, err := os.Create(*writecsv)
		if err != nil {
			log.Fatalln("failed to open file", err)
		}

		w := csv.NewWriter(f)
		err = w.WriteAll(summary) // calls Flush internally
		if err != nil {
			log.Fatal(err)
		}
		f.Close()
		fmt.Printf("Results written to: %v\n", *writecsv)

	} else {

		for _, record := range summary[1:] {
			y := make([]interface{}, len(record)-2)
			for i, v := range record[2:] {
				y[i] = v
			}

			fmt.Printf("%v (%vx) took: %vms\navg: %vms, min: %vms, max: %vms\nops/s: %v\nms/op: %v\n\n", y...)
		}
	}

	printTotal(startTotal)
}

func printTotal(startTotal time.Time) {
	fmt.Printf("elapsed time: %v\n", time.Since(startTotal))
}

func contains(options []string, want string) bool {
	for _, o := range options {
		if o == want {
			return true
		}
	}
	return false
}

func MergeKnownCsv(rootDir string, targetFile string) {

	files, err := ioutil.ReadDir(rootDir)
	if err != nil {
		log.Fatal(err)
	}

	allrecords := [][]string{hheaders}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".csv" && filepath.Base(file.Name()) != filepath.Base(targetFile) {
			fileToMerge := fmt.Sprintf("%v/%v", filepath.Clean(rootDir), file.Name())
			_file, err := os.Open(fileToMerge)
			if err != nil {
				fmt.Println(err)
			}
			reader := csv.NewReader(_file)
			records, _ := reader.ReadAll()
			if len(records) > 1 {
				headerrow := records[0]
				isgood := reflect.DeepEqual(headerrow, hheaders)

				if isgood {
					allrecords = append(allrecords, records[1:]...)
					fmt.Printf("Merging:\t%v\n", fileToMerge)
				} else {
					fmt.Printf("Bad structure:\t%v\n", fileToMerge)
				}
			}
		}
	}

	f, err := os.Create(targetFile)
	if err != nil {
		fmt.Println("failed to open file", err)
	}

	w := csv.NewWriter(f)
	err = w.WriteAll(allrecords) // calls Flush internally
	if err != nil {
		fmt.Println(err)
	}
	f.Close()
	fmt.Printf("Result:  \t%v\n", targetFile)
}

func CreateCharts(dataFile string) {

	csvfile, err := os.Open(dataFile)
	if err != nil {
		log.Fatal(err)
	}

	df := dataframe.ReadCSV(csvfile)
	if len(df.Records()) <= 1 {
		fmt.Println("specified file has no data")
		os.Exit(1)
	}

	systems := unique(df.Select([]string{"system"}).Records())
	mults := unique(df.Select([]string{"multiplicity"}).Records())
	names := unique(df.Select([]string{"name"}).Records())

	page := components.NewPage()

	for c1, metric := range []string{"total (μs)", "avg (μs)", "ops/s", "μs/op"} {
		for c2, mult := range mults {
			bar := getBasicBarChart(fmt.Sprintf("Chart %v.%v", c1+1, c2), fmt.Sprintf("%v with %v iterations", metric, mult))
			bar.SetXAxis(names)
			for _, system := range systems {
				data := df.
					Filter(dataframe.F{0, "system", "==", system}).
					Filter(dataframe.F{1, "multiplicity", "==", mult}).
					Select([]string{metric}).Records()[1:]
				if len(data) != 0 {
					bar.AddSeries(system, generateBarItems(data))
				}
			}
			bar.SetSeriesOptions(
				charts.WithBarChartOpts(opts.BarChart{Type: "bar", BarGap: "10%", BarCategoryGap: "30%", RoundCap: true}),
				charts.WithLabelOpts(opts.Label{Show: true, Position: "top"}),
			)
			page.AddCharts(bar)
		}
	}

	page.SetLayout(components.PageFlexLayout)

	html := fmt.Sprintf("%v/%v", filepath.Dir(dataFile), "charts.html")
	f, err := os.Create(html)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	page.Render(io.MultiWriter(f))

	fmt.Printf("Charts created in: %v\n", html)
}

func getBasicBarChart(title string, subtitle string) *charts.Bar {

	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{PageTitle: "Charts", Width: "1300px", Height: "500px"}),
		charts.WithTitleOpts(opts.Title{Title: title, Subtitle: subtitle}),
		charts.WithLegendOpts(opts.Legend{Show: true, Y: "30"}),
		charts.WithColorsOpts(opts.Colors{"blue", "red", "green", "purple", "orange", "brown", "yellow", "black"}),
		charts.WithYAxisOpts(opts.YAxis{AxisLabel: &opts.AxisLabel{Show: true, Formatter: "{value}"}}),
		charts.WithXAxisOpts(opts.XAxis{AxisLabel: &opts.AxisLabel{Show: true, Rotate: 20}}),
		charts.WithToolboxOpts(opts.Toolbox{Show: true, Right: "10%", Feature: &opts.ToolBoxFeature{
			SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{Show: true, Title: "Download", Type: "png"},
			DataView:    &opts.ToolBoxFeatureDataView{Show: true, Title: "Data", Lang: []string{"raw data", "go back", "refresh"}},
			DataZoom:    &opts.ToolBoxFeatureDataZoom{Show: true},
		}}),
		//charts.WithDataZoomOpts(opts.DataZoom{Type: "slider", Start: 0, End: 100}),
	)

	return bar
}

func barSample() *charts.Bar {

	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{PageTitle: "Charts", Width: "900px", Height: "500px"}),
		charts.WithTitleOpts(opts.Title{Title: "Chart Title", Subtitle: "Any subtitle or description"}),
		charts.WithLegendOpts(opts.Legend{Show: true, Y: "20"}),
		charts.WithColorsOpts(opts.Colors{"blue", "red", "green"}),
		charts.WithYAxisOpts(opts.YAxis{AxisLabel: &opts.AxisLabel{Show: true, Formatter: "{value}"}}),
		charts.WithToolboxOpts(opts.Toolbox{Show: true, Right: "10%", Feature: &opts.ToolBoxFeature{
			SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{Show: true, Title: "Download", Type: "png"},
			DataView:    &opts.ToolBoxFeatureDataView{Show: true, Title: "Data", Lang: []string{"raw data", "go back", "refresh"}},
			DataZoom:    &opts.ToolBoxFeatureDataZoom{Show: true},
		}}),
		//charts.WithDataZoomOpts(opts.DataZoom{Type: "slider", Start: 0, End: 100}),
	)
	bar.SetXAxis([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
		AddSeries("mysql", generateRandomBarItems()).
		AddSeries("postgres", generateRandomBarItems()).
		AddSeries("neo4j", generateRandomBarItems()).
		SetSeriesOptions(
			charts.WithBarChartOpts(opts.BarChart{Type: "bar", BarGap: "10%", BarCategoryGap: "30%", RoundCap: true}),
			charts.WithLabelOpts(opts.Label{Show: true, Position: "top"}),
		)

	return bar
}

// generate random data for bar chart
func generateRandomBarItems() []opts.BarData {
	items := make([]opts.BarData, 0)
	for i := 0; i < 7; i++ {
		items = append(items, opts.BarData{Value: rand.Intn(30000)})
	}
	return items
}

func generateBarItems(table [][]string) []opts.BarData {
	items := make([]opts.BarData, 0)

	for _, a := range table[1:] {
		for _, b := range a {
			items = append(items, opts.BarData{Name: a[0], Value: b})
		}
	}

	return items
}

func unique(table [][]string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, stringSlice := range table[1:] {
		for _, entry := range stringSlice {
			if _, value := keys[entry]; !value {
				keys[entry] = true
				list = append(list, entry)
			}
		}
	}
	return list
}
