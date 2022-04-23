package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/RomanBoegli/godbbench/benchmark"
	"github.com/RomanBoegli/godbbench/databases"
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
	hheaders = []string{"system", "multiplicity", "name", "executions", "total (μs)", "arithMean (μs)", "geoMean (μs)", "min (μs)", "max (μs)", "ops/s", "μs/op"}
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
		charttype        = createChartFlags.String("charttype", "line", "default is 'line', alternative is 'bar'")
	)

	defaultFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Available subcommands:\n\tmysql | postgres | neo4j | mergecsv | createcharts\n")
		fmt.Fprintf(os.Stderr, "\tUse 'subcommand --help' for all flags of the specified command.\n")
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
		CreateCharts(*dataFile, *charttype)
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
			μsPerOp := float64(results.Duration.Microseconds())

			// execution in ns/op for mode loop
			if b.Type == benchmark.TypeLoop {
				μsPerOp /= float64(int64(*iter))
			}

			summary = append(summary, []string{
				system,
				fmt.Sprint(*iter),
				b.Name,
				fmt.Sprint(results.TotalExecutionCount),
				fmt.Sprint(results.Duration.Microseconds()),
				fmt.Sprint(results.ArithMean().Microseconds()),
				fmt.Sprint(results.GeoMean().Microseconds()),
				fmt.Sprint(results.Min.Microseconds()),
				fmt.Sprint(results.Max.Microseconds()),
				fmt.Sprint(int64(float64(results.TotalExecutionCount) / (results.Duration.Seconds()))),
				fmt.Sprint(int64(μsPerOp))})

			// Don't sleep after the last benchmark
			if i != len(benchmarks)-1 {
				time.Sleep(*sleep)
			}
		}
	}

	// write results to csv
	if *writecsv != "" {
		path := filepath.Dir(*writecsv)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err := os.MkdirAll(path, os.ModePerm)
			if err != nil {
				log.Fatalln("failed to create folder", err)
			}
		}
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

			fmt.Printf("%v (%vx) took: %vμs\narithMean: %vμs, geoMean: %vμs\nmin: %vμs, max: %vμs\nops/s: %v, μs/op: %v\n\n", y...)
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

func CreateCharts(dataFile string, charttype string) {

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

	for c1, name := range names {
		for c2, metric := range []string{"arithMean (μs)", "geoMean (μs)", "ops/s", "μs/op"} {
			chart := getBasicChart(fmt.Sprintf("Chart %v.%v: %v", c1+1, c2, name), "", "multiplicity", metric)
			chart.SetXAxis(mults)
			for _, system := range systems {
				data := df.
					Filter(dataframe.F{0, "system", "==", system}).
					Filter(dataframe.F{1, "name", "==", name}).
					Select([]string{metric}).Records()
				if len(data) != 0 {
					chart.AddSeries(system, generateBarItems(data))
				}
			}
			chart.SetSeriesOptions(
				charts.WithBarChartOpts(opts.BarChart{Type: charttype, BarGap: "10%", BarCategoryGap: "30%", RoundCap: true}),
				charts.WithLineChartOpts(opts.LineChart{Smooth: true}),
				charts.WithLabelOpts(opts.Label{Show: true, Position: "top"}),
			)
			page.AddCharts(chart)
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

func getBasicChart(title string, subtitle string, xAxisLabel string, yAxisLabel string) *charts.Bar {

	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{PageTitle: "Charts", Width: "1100px", Height: "450px"}),
		charts.WithTitleOpts(opts.Title{Title: title, Subtitle: subtitle}),
		charts.WithLegendOpts(opts.Legend{Show: true, Y: "30", SelectedMode: "multiple", ItemWidth: 20}),
		charts.WithColorsOpts(opts.Colors{"#E16F0C", "#318BFF", "#23B12A"}),
		charts.WithYAxisOpts(opts.YAxis{AxisLabel: &opts.AxisLabel{Show: true, Formatter: "{value}"}}),
		//charts.WithXAxisOpts(opts.XAxis{AxisLabel: &opts.AxisLabel{Show: true, Rotate: 0, FontSize: "9", Interval: "0"}}), // has a bug
		charts.WithToolboxOpts(opts.Toolbox{Show: true, Right: "10%", Feature: &opts.ToolBoxFeature{
			SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{Show: true, Title: "Download", Type: "png"},
			DataView:    &opts.ToolBoxFeatureDataView{Show: true, Title: "Data", Lang: []string{"raw data", "go back", "refresh"}},
			DataZoom:    &opts.ToolBoxFeatureDataZoom{Show: true},
		}}),
		charts.WithXAxisOpts(opts.XAxis{Name: xAxisLabel}),
		charts.WithYAxisOpts(opts.YAxis{Name: yAxisLabel}),
		charts.WithDataZoomOpts(opts.DataZoom{Type: "slider", Start: 0, End: 100}),
	)

	return bar
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
