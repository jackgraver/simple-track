package utils

import (
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const benchmarkRoutePath = "/benchmark"

// BenchmarkMiddleware records per-route timing and registers GET /benchmark for stats.
// Query: none → all routes; path=… exact; q=… substring; sort=path|avg|time|total (default path asc); order=asc|desc (time sorts default desc).
func BenchmarkMiddleware(router *gin.Engine) gin.HandlerFunc {
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

		benchmarker.AddBenchmark(path, float64(elapsed.Nanoseconds())/1e6)
	}
}

type Benchmark struct {
	Path         string  `json:"path"`
	TotalHits    int     `json:"totalHits"`
	TotalTimeMs  float64 `json:"totalTimeMs"`
	AverageMs    float64 `json:"averageMs"`
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
		"path":         be.Path,
		"totalHits":    be.TotalHits,
		"totalTimeMs":  be.TotalTimeMs,
		"averageMs":    be.AverageMs,
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
