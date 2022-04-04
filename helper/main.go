package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-gota/gota/dataframe"
	"github.com/spf13/pflag"
)

func main() {
	var (
		defaultFlags = pflag.NewFlagSet("default", pflag.ExitOnError)

		mergeCsvFlags = pflag.NewFlagSet("mergecsv", pflag.ExitOnError)
		rootDir       = mergeCsvFlags.String("rootDir", "../helper", "path to folder with csv files to be merged")
		targetFile    = mergeCsvFlags.String("targetFile", "../helper/merged.csv", "target file path for merged csv")

		createChartFlags = pflag.NewFlagSet("createcharts", pflag.ExitOnError)
		dataFile         = createChartFlags.String("dataFile", "../helper/merged.csv", "path to source data file, assumes headers")
	)

	defaultFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Available subcommands:\n\tmergecsv|createcharts\n")
		fmt.Fprintf(os.Stderr, "\tUse 'subcommand --help' for all flags of the specified command.\n")
		fmt.Fprintf(os.Stderr, "Generic flags for all subcommands:\n")
		defaultFlags.PrintDefaults()
	}

	// No comamnd given. Print usage help and exit.
	if len(os.Args) < 2 {
		defaultFlags.Usage()
		os.Exit(1)
	}

	var system = os.Args[1]
	switch system {
	case "mergecsv":
		if err := mergeCsvFlags.Parse(os.Args[2:]); err != nil {
			log.Fatalf("failed to parse postgres flags: %v", err)
		}
		MergeKnownCsv(*rootDir, *targetFile)
	case "createcharts":
		if err := createChartFlags.Parse(os.Args[2:]); err != nil {
			log.Fatalf("failed to parse postgres flags: %v", err)
		}
		CreateCharts(*dataFile)
	default:
		if err := defaultFlags.Parse(os.Args[1:]); err != nil {
			log.Fatalf("failed to parse default flags: %v", err)
		}
		defaultFlags.Usage()
		os.Exit(1)
	}
}

func MergeKnownCsv(rootDir string, targetFile string) {

	files, err := ioutil.ReadDir(rootDir)
	if err != nil {
		log.Fatal(err)
	}

	headers := []string{"system", "multiplicity", "name", "executions", "total (μs)", "avg (μs)", "min (μs)", "max (μs)", "ops/s", "μs/op"}
	allrecords := [][]string{headers}

	for _, file := range files {
		filename := file.Name()
		if filepath.Ext(filename) == ".csv" && filepath.Base(filename) != filepath.Base(targetFile) {

			_file, err := os.Open(filename)
			if err != nil {
				fmt.Println(err)
			}
			reader := csv.NewReader(_file)
			records, _ := reader.ReadAll()
			headerrow := records[0]
			isgood := reflect.DeepEqual(headerrow, headers)

			if isgood {
				allrecords = append(allrecords, records[1:]...)
			} else {
				fmt.Printf("File with wrong structure: %v\n", filename)
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
	fmt.Printf("Generated merge file: %v\n", targetFile)
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

	// total

	for c1, metric := range []string{"total (μs)", "avg (μs)", "ops/s", "μs/op"} {
		for c2, mult := range mults {
			bar := getBasicBarChart(fmt.Sprintf("Chart %v.%v", c1+1, c2), fmt.Sprintf("%v with %v iterations", metric, mult))
			bar.SetXAxis(names)
			for _, system := range systems {
				data := df.
					Filter(dataframe.F{0, "system", "==", system}).
					Filter(dataframe.F{1, "multiplicity", "==", mult}).
					Select([]string{metric}).Records()[1:]
				adsf := generateBarItems(data)
				bar.AddSeries(system, adsf)
			}
			bar.SetSeriesOptions(
				charts.WithBarChartOpts(opts.BarChart{Type: "bar", BarGap: "10%", BarCategoryGap: "30%", RoundCap: true}),
				charts.WithLabelOpts(opts.Label{Show: true, Position: "top"}),
			)
			page.AddCharts(bar)
		}
	}

	page.SetLayout(components.PageFlexLayout)

	f, err := os.Create("Charts.html")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	page.Render(io.MultiWriter(f))

	fmt.Printf("Charts created in: %v\n", f)
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
