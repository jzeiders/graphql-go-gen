package report

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/jzeiders/graphql-go-gen/benchmark/internal/runner"
)

type Reporter struct {
	jsonOutput bool
	outputPath string
}

func NewReporter(jsonOutput bool, outputPath string) *Reporter {
	return &Reporter{
		jsonOutput: jsonOutput,
		outputPath: outputPath,
	}
}

type JSONReport struct {
	Timestamp   time.Time             `json:"timestamp"`
	System      SystemInfo            `json:"system"`
	Benchmarks  []BenchmarkReport     `json:"benchmarks"`
	Summary     Summary               `json:"summary"`
}

type SystemInfo struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	CPUCount     int    `json:"cpu_count"`
	GoVersion    string `json:"go_version"`
}

type BenchmarkReport struct {
	Name               string        `json:"name"`
	FileCount          int           `json:"file_count"`
	TagCount           int           `json:"tag_count"`
	TotalLOC           int           `json:"total_loc"`
	SetupTimeMs        int64         `json:"setup_time_ms"`
	GenerationTimeMs   int64         `json:"generation_time_ms"`
	MemoryUsedBytes    uint64        `json:"memory_used_bytes"`
	FilesPerSecond     float64       `json:"files_per_second"`
	TagsPerSecond      float64       `json:"tags_per_second"`
	LOCPerSecond       float64       `json:"loc_per_second"`
	ErrorCount         int           `json:"error_count"`
	Errors             []string      `json:"errors,omitempty"`
}

type Summary struct {
	TotalFiles         int           `json:"total_files"`
	TotalTags          int           `json:"total_tags"`
	TotalLOC           int           `json:"total_loc"`
	TotalGenerationMs  int64         `json:"total_generation_ms"`
	AverageFilesPerSec float64       `json:"average_files_per_second"`
	AverageTagsPerSec  float64       `json:"average_tags_per_second"`
}

func (r *Reporter) Generate(results []*runner.BenchmarkResult) error {
	if r.jsonOutput {
		return r.generateJSON(results)
	}
	return r.generateTable(results)
}

func (r *Reporter) generateTable(results []*runner.BenchmarkResult) error {
	// Header
	fmt.Println("\n" + strings.Repeat("=", 120))
	fmt.Println("BENCHMARK RESULTS")
	fmt.Println(strings.Repeat("=", 120))
	fmt.Printf("Generated at: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Header
	fmt.Fprintln(w, "Test Set\tFiles\tTags\tLOC\tSetup\tGeneration\tFiles/s\tTags/s\tMemory\tStatus")
	fmt.Fprintln(w, strings.Repeat("-", 110))

	// Data rows
	var totalFiles, totalTags, totalLOC int
	var totalGenTime time.Duration

	for _, r := range results {
		status := "✅ Success"
		if len(r.Errors) > 0 {
			status = fmt.Sprintf("❌ %d errors", len(r.Errors))
		}

		genSeconds := r.GenerationTime.Seconds()
		filesPerSec := float64(r.FileCount) / genSeconds
		tagsPerSec := float64(r.TagCount) / genSeconds

		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%v\t%v\t%.1f\t%.1f\t%s\t%s\n",
			r.Name,
			r.FileCount,
			r.TagCount,
			r.TotalLOC,
			r.SetupTime.Round(time.Millisecond),
			r.GenerationTime.Round(time.Millisecond),
			filesPerSec,
			tagsPerSec,
			formatBytes(r.MemoryUsed),
			status,
		)

		totalFiles += r.FileCount
		totalTags += r.TagCount
		totalLOC += r.TotalLOC
		totalGenTime += r.GenerationTime
	}

	// Summary row
	fmt.Fprintln(w, strings.Repeat("-", 110))
	avgFilesPerSec := float64(totalFiles) / totalGenTime.Seconds()
	avgTagsPerSec := float64(totalTags) / totalGenTime.Seconds()

	fmt.Fprintf(w, "TOTAL\t%d\t%d\t%d\t\t%v\t%.1f\t%.1f\t\t\n",
		totalFiles,
		totalTags,
		totalLOC,
		totalGenTime.Round(time.Millisecond),
		avgFilesPerSec,
		avgTagsPerSec,
	)

	w.Flush()

	// Print errors if any
	for _, r := range results {
		if len(r.Errors) > 0 {
			fmt.Printf("\n⚠️  Errors for %s:\n", r.Name)
			for _, err := range r.Errors {
				fmt.Printf("  - %v\n", err)
			}
		}
	}

	// Performance insights
	fmt.Println("\n" + strings.Repeat("=", 120))
	fmt.Println("PERFORMANCE INSIGHTS")
	fmt.Println(strings.Repeat("-", 120))

	fmt.Printf("Average Processing Speed:\n")
	fmt.Printf("  - Files: %.2f files/second\n", avgFilesPerSec)
	fmt.Printf("  - Tags: %.2f tags/second\n", avgTagsPerSec)
	fmt.Printf("  - LOC: %.0f lines/second\n", float64(totalLOC)/totalGenTime.Seconds())

	// Find fastest and slowest
	if len(results) > 1 {
		var fastest, slowest *runner.BenchmarkResult
		for _, r := range results {
			if len(r.Errors) == 0 {
				if fastest == nil || r.GenerationTime < fastest.GenerationTime {
					fastest = r
				}
				if slowest == nil || r.GenerationTime > slowest.GenerationTime {
					slowest = r
				}
			}
		}

		if fastest != nil && slowest != nil && fastest != slowest {
			fmt.Printf("\nFastest: %s (%.2fs)\n", fastest.Name, fastest.GenerationTime.Seconds())
			fmt.Printf("Slowest: %s (%.2fs)\n", slowest.Name, slowest.GenerationTime.Seconds())

			speedup := slowest.GenerationTime.Seconds() / fastest.GenerationTime.Seconds()
			fmt.Printf("Speed difference: %.2fx\n", speedup)
		}
	}

	fmt.Println(strings.Repeat("=", 120))

	return nil
}

func (r *Reporter) generateJSON(results []*runner.BenchmarkResult) error {
	report := JSONReport{
		Timestamp: time.Now(),
		System:    getSystemInfo(),
		Benchmarks: make([]BenchmarkReport, len(results)),
	}

	var totalFiles, totalTags, totalLOC int
	var totalGenTimeMs int64

	for i, res := range results {
		genSeconds := res.GenerationTime.Seconds()

		br := BenchmarkReport{
			Name:             res.Name,
			FileCount:        res.FileCount,
			TagCount:         res.TagCount,
			TotalLOC:         res.TotalLOC,
			SetupTimeMs:      res.SetupTime.Milliseconds(),
			GenerationTimeMs: res.GenerationTime.Milliseconds(),
			MemoryUsedBytes:  res.MemoryUsed,
			FilesPerSecond:   float64(res.FileCount) / genSeconds,
			TagsPerSecond:    float64(res.TagCount) / genSeconds,
			LOCPerSecond:     float64(res.TotalLOC) / genSeconds,
			ErrorCount:       len(res.Errors),
		}

		if len(res.Errors) > 0 {
			br.Errors = make([]string, len(res.Errors))
			for j, err := range res.Errors {
				br.Errors[j] = err.Error()
			}
		}

		report.Benchmarks[i] = br

		totalFiles += res.FileCount
		totalTags += res.TagCount
		totalLOC += res.TotalLOC
		totalGenTimeMs += res.GenerationTime.Milliseconds()
	}

	totalGenSeconds := float64(totalGenTimeMs) / 1000.0
	report.Summary = Summary{
		TotalFiles:         totalFiles,
		TotalTags:          totalTags,
		TotalLOC:           totalLOC,
		TotalGenerationMs:  totalGenTimeMs,
		AverageFilesPerSec: float64(totalFiles) / totalGenSeconds,
		AverageTagsPerSec:  float64(totalTags) / totalGenSeconds,
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	if r.outputPath != "" {
		return os.WriteFile(r.outputPath, data, 0644)
	}

	fmt.Println(string(data))
	return nil
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func getSystemInfo() SystemInfo {
	return SystemInfo{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		CPUCount:     runtime.NumCPU(),
		GoVersion:    runtime.Version(),
	}
}