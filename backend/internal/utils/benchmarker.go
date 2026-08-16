package utils

import (
	"be-simpletracker/internal/env"
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	benchmarkLogFile      = "benchmark.log"
	benchmarkLogChanSize  = 4096
	benchmarkLogWriteBuf  = 64 * 1024
	benchmarkLogFlushTick = 500 * time.Millisecond
)

type benchmarkLogEntry struct {
	Route     string  `json:"route"`
	Method    string  `json:"method"`
	LatencyMs float64 `json:"latency_ms"`
	Timestamp string  `json:"timestamp"`
}

var benchmarkLogCh chan benchmarkLogEntry
var benchmarkLogOnce sync.Once

func startBenchmarkLogWriter() {
	benchmarkLogOnce.Do(func() {
		benchmarkLogCh = make(chan benchmarkLogEntry, benchmarkLogChanSize)
		go runBenchmarkLogWriter()
	})
}

// runBenchmarkLogWriter owns the benchmark log file for the lifetime of the
// process: a single open handle, a buffered writer to coalesce syscalls, and
// a periodic flush so data still reaches disk under low traffic.
func runBenchmarkLogWriter() {
	f, err := os.OpenFile(benchmarkLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("benchmark log: failed to open %s: %v", benchmarkLogFile, err)
		for range benchmarkLogCh {
		}
		return
	}
	defer f.Close()

	bw := bufio.NewWriterSize(f, benchmarkLogWriteBuf)
	enc := json.NewEncoder(bw)
	defer bw.Flush()

	ticker := time.NewTicker(benchmarkLogFlushTick)
	defer ticker.Stop()

	writeEntry := func(e benchmarkLogEntry) {
		if err := enc.Encode(e); err != nil {
			log.Printf("benchmark log: write error: %v", err)
		}
	}

	for {
		select {
		case entry, ok := <-benchmarkLogCh:
			if !ok {
				return
			}
			writeEntry(entry)
			drained := false
			for !drained {
				select {
				case more, ok := <-benchmarkLogCh:
					if !ok {
						return
					}
					writeEntry(more)
				default:
					drained = true
				}
			}
		case <-ticker.C:
			if err := bw.Flush(); err != nil {
				log.Printf("benchmark log: flush error: %v", err)
			}
		}
	}
}

const benchmarkRoutePath = "/benchmark"

// BenchmarkMiddleware records per-route timing and registers GET /benchmark for stats.
// Query: none → all routes; path=… exact; q=… substring; sort=path|avg|time|total (default path asc); order=asc|desc (time sorts default desc).
func BenchmarkMiddleware(router *gin.Engine) gin.HandlerFunc {
	if env.IsProduction() {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	startBenchmarkLogWriter()
	benchmarker := &Benchmarker{Benchmarks: make(map[string]*Benchmark)}

	router.GET(benchmarkRoutePath, func(c *gin.Context) {
		exact := strings.TrimSpace(c.Query("path"))
		q := strings.TrimSpace(c.Query("q"))
		sortBy := strings.ToLower(strings.TrimSpace(c.Query("sort")))
		order := strings.ToLower(strings.TrimSpace(c.Query("order")))
		routes := benchmarker.ListStats(exact, q, sortBy, order)
		c.JSON(http.StatusOK, gin.H{
			"routes": routes,
			"count":  len(routes),
		})
	})

	return func(c *gin.Context) {
		if c.FullPath() == benchmarkRoutePath {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()
		elapsed := time.Since(start)

		path := c.FullPath()
		if path == "" {
			return
		}

		latency := float64(elapsed.Nanoseconds()) / 1e6
		benchmarker.AddBenchmark(path, latency)
		select {
		case benchmarkLogCh <- benchmarkLogEntry{
			Route:     path,
			Method:    c.Request.Method,
			LatencyMs: latency,
			Timestamp: start.UTC().Format(time.RFC3339Nano),
		}:
		default:
		}
	}
}

type Benchmark struct {
	Path        string  `json:"path"`
	TotalHits   int     `json:"totalHits"`
	TotalTimeMs float64 `json:"totalTimeMs"`
	AverageMs   float64 `json:"averageMs"`
}

type Benchmarker struct {
	mu         sync.RWMutex
	Benchmarks map[string]*Benchmark
}

func (b *Benchmarker) AddBenchmark(path string, durationMs float64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	existing := b.Benchmarks[path]
	if existing != nil {
		existing.TotalHits++
		existing.TotalTimeMs += durationMs
		existing.AverageMs = existing.TotalTimeMs / float64(existing.TotalHits)
		return
	}

	b.Benchmarks[path] = &Benchmark{
		Path:        path,
		TotalHits:   1,
		TotalTimeMs: durationMs,
		AverageMs:   durationMs,
	}
}

func statsRow(be *Benchmark) gin.H {
	return gin.H{
		"path":        be.Path,
		"totalHits":   be.TotalHits,
		"totalTimeMs": be.TotalTimeMs,
		"averageMs":   be.AverageMs,
	}
}

// ListStats returns benchmark rows filtered by exact path and/or substring q.
// sortBy: empty or "path" → by path; "avg", "average", "time" → averageMs; "total", "sum" → totalTimeMs.
// order: "asc"|"ascending" vs "desc"|"descending"; defaults: path asc, time metrics desc (slowest / largest first).
func (b *Benchmarker) ListStats(exactPath, q, sortBy, order string) []gin.H {
	b.mu.RLock()
	defer b.mu.RUnlock()

	paths := make([]string, 0, len(b.Benchmarks))
	for p := range b.Benchmarks {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	ql := strings.ToLower(q)
	var rows []*Benchmark
	for _, p := range paths {
		be := b.Benchmarks[p]
		if be == nil {
			continue
		}
		if exactPath != "" && be.Path != exactPath {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(be.Path), ql) {
			continue
		}
		rows = append(rows, be)
	}

	if sortBy == "" {
		sortBy = "path"
	}
	if order == "" {
		if sortBy == "path" {
			order = "asc"
		} else {
			order = "desc"
		}
	}
	asc := order == "asc" || order == "ascending"

	switch sortBy {
	case "avg", "average", "time":
		sort.Slice(rows, func(i, j int) bool {
			if asc {
				return rows[i].AverageMs < rows[j].AverageMs
			}
			return rows[i].AverageMs > rows[j].AverageMs
		})
	case "total", "sum":
		sort.Slice(rows, func(i, j int) bool {
			if asc {
				return rows[i].TotalTimeMs < rows[j].TotalTimeMs
			}
			return rows[i].TotalTimeMs > rows[j].TotalTimeMs
		})
	default:
		sort.Slice(rows, func(i, j int) bool {
			if asc {
				return rows[i].Path < rows[j].Path
			}
			return rows[i].Path > rows[j].Path
		})
	}

	out := make([]gin.H, 0, len(rows))
	for _, be := range rows {
		out = append(out, statsRow(be))
	}
	return out
}
