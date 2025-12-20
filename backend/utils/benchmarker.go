package utils

import (
	"time"

	"github.com/gin-gonic/gin"
)

func BenchmarkMiddleware(router *gin.Engine) gin.HandlerFunc {
	benchmarker := Benchmarker{Benchmarks: make(map[string]*Benchmark)}

	router.GET("/benchmark", func(c *gin.Context) {	
		c.JSON(200, benchmarker.GetBenchmarksStats("/mealplan/today"))
	})

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		elapsed := time.Since(start)
		benchmarker.AddBenchmark(c.FullPath(), elapsed.Seconds())
	}
}

type Benchmark struct {
	Path      string
	TotalHits int
	TotalTime float64
}

type Benchmarker struct {
	Benchmarks map[string]*Benchmark
}

func (b *Benchmarker) AddBenchmark(path string, totalTime float64) {
	benchark := b.GetBenchmarks(path)
	if benchark != nil {
		benchark.TotalHits += 1
		benchark.TotalTime += totalTime
	} else {
		benchmark := Benchmark{
			Path:      path,
			TotalHits: 1,
			TotalTime: totalTime,
		}
		b.Benchmarks[path] = &benchmark
	}
}

func (b *Benchmarker) GetBenchmarks(Path string) *Benchmark {
	return b.Benchmarks[Path]
}

func (b *Benchmarker) GetBenchmarksStats(Path string) map[string]interface{} {
	benchark := b.GetBenchmarks(Path)
	if benchark == nil {
		return nil
	}

	return map[string]interface{}{
		"path":      benchark.Path,
		"totalHits": benchark.TotalHits,
		"totalTime": benchark.TotalTime,
		"average":   benchark.TotalTime / float64(benchark.TotalHits),
	}
}