package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/actions-go/toolkit/core"
	"github.com/panagiotisptr/cov-diff/cov"
	"github.com/panagiotisptr/cov-diff/diff"
	"github.com/panagiotisptr/cov-diff/files"
	"github.com/panagiotisptr/cov-diff/interval"
)

var path = flag.String("path", "", "path to the git repository")
var coverageFile = flag.String("coverprofile", "", "location of the coverage file")
var diffFile = flag.String("diff", "", "location of the diff file")
var moduleName = flag.String("module", "", "the name of module")
var ignoreMain = flag.String("ignore-main", "", "ignore main package")

func emptyValAndActionInputSet(val string, input string) bool {
	return val == "" && os.Getenv(
		fmt.Sprintf("INPUT_%s", strings.ToUpper(input)),
	) != ""
}

func getActionInput(input string) string {
	return os.Getenv(
		fmt.Sprintf("INPUT_%s", strings.ToUpper(input)),
	)
}

func populateFlagsFromActionEnvs() {
	if emptyValAndActionInputSet(*path, "path") {
		*path = getActionInput("path")
	}
	if emptyValAndActionInputSet(*coverageFile, "coverprofile") {
		*coverageFile = getActionInput("coverprofile")
	}
	if emptyValAndActionInputSet(*diffFile, "diff") {
		*diffFile = getActionInput("diff")
	}
	if emptyValAndActionInputSet(*moduleName, "module") {
		*moduleName = getActionInput("module")
	}
	if emptyValAndActionInputSet(*ignoreMain, "ignore-main") {
		*ignoreMain = getActionInput("ignore-main")
	}
}

func buildMissingMessage(missingLines map[string][]interval.Interval) string {
	var md strings.Builder
	md.WriteString("| File | Missing Line Intervals |\n")
	md.WriteString("|------|------------------------|\n")

	// Fill the table with data from the map
	for file, intervals := range missingLines {
		md.WriteString(fmt.Sprintf("| %s |", file))
		first := true
		for _, interval := range intervals {
			if !first {
				md.WriteString(", ")
			}
			md.WriteString(fmt.Sprintf("%d-%d", interval.Start, interval.End))
			first = false
		}
		md.WriteString(" |\n")
	}
	return md.String()
}

func main() {
	flag.Parse()
	populateFlagsFromActionEnvs()

	if *coverageFile == "" {
		log.Fatal("missing coverage file")
	}

	diffBytes, err := os.ReadFile(*diffFile)
	if err != nil {
		log.Fatal("failed to read diff file: ", err)
	}

	diffIntervals, err := diff.GetFilesIntervalsFromDiff(diffBytes)
	if err != nil {
		log.Fatal(err)
	}
	// de-allocate diffBytes
	diffBytes = nil

	covFileBytes, err := os.ReadFile(*coverageFile)
	if err != nil {
		log.Fatal(err)
	}

	coverIntervals, err := cov.GetFilesIntervalsFromCoverage(covFileBytes)
	if err != nil {
		log.Fatal(err)
	}
	// de-allocate covFileBytes
	covFileBytes = nil

	total := 0
	covered := 0

	missingLines := map[string][]interval.Interval{}

	for filename, di := range diffIntervals {
		fileBytes, err := os.ReadFile(filepath.Join(*path, filename))
		if err != nil {
			log.Fatal(err)
		}
		fi, err := files.GetIntervalsFromFile(fileBytes, *ignoreMain == "true")
		if err != nil {
			log.Fatal(err)
		}

		// intervals that changed and are parts of the code we care about
		measuredIntervals := interval.Union(di, fi)
		total += interval.Sum(measuredIntervals)

		fullFilename := filepath.Join(*moduleName, filename)
		ci, ok := coverIntervals[fullFilename]
		if !ok {
			fmt.Println("no coverage data for", fullFilename)
			missingLines[fullFilename] = measuredIntervals
			continue
		}

		coveredMeasuredIntervals := interval.Union(measuredIntervals, ci)
		covered += interval.Sum(coveredMeasuredIntervals)

		missingLines[fullFilename] = interval.SubtractIntervals(measuredIntervals, ci)
	}

	percentCoverage := 100
	if total != 0 {
		percentCoverage = (100 * covered) / total
	}

	missingLinesMessage := buildMissingMessage(missingLines)

	fmt.Printf("Coverage on new lines: %d%%\n", percentCoverage)
	if getActionInput("coverprofile") != "" {
		core.SetOutput("covdiff", fmt.Sprintf("%d", percentCoverage))
		core.SetOutput("missing-lines", missingLinesMessage)
	}
}
