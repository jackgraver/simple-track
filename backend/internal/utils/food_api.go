package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// USDA FoodData Central search endpoint.
// Docs: https://fdc.nal.usda.gov/api-guide
const fdcSearchURL = "https://api.nal.usda.gov/fdc/v1/foods/search"

// USDA nutrient ids we surface to clients (USDA "nutrientId" values).
const (
	nutrientEnergyKcal        = 1008 // Energy
	nutrientEnergyAtwaterGen  = 2047 // Energy (Atwater General Factors)
	nutrientEnergyAtwaterSpec = 2048 // Energy (Atwater Specific Factors)
	nutrientProtein           = 1003
	nutrientCarbs             = 1005
	nutrientFat               = 1004
	nutrientFiber             = 1079
	nutrientSugar             = 2000
)

const (
	minQueryLen     = 2
	maxQueryLen     = 100
	defaultPageSize = 5
	cacheTTL        = 10 * time.Minute
	cacheMaxEntries = 500
	upstreamTimeout = 8 * time.Second

	upstreamLogEvery = 25
)

// FoodSummary is the trimmed-down food row returned to clients.
// Macros are reported per serving_size (serving_size_unit). When FDC omits
// serving size or unit, we assume 100 g (common for non-branded / legacy rows).
type FoodSummary struct {
	FdcID       int     `json:"fdc_id"`
	Name        string  `json:"name"`
	Brand       string  `json:"brand,omitempty"`
	ServingSize float64 `json:"serving_size"`
	ServingUnit string  `json:"serving_size_unit"`
	Calories    float64 `json:"calories"`
	Protein     float64 `json:"protein_g"`
	Carbs       float64 `json:"carbs_g"`
	Fat         float64 `json:"fat_g"`
	Fiber       float64 `json:"fiber_g"`
	Sugar       float64 `json:"sugar_g"`
}

// ErrQueryTooShort is returned when the trimmed query is shorter than minQueryLen.
// We prefer this server-side gate over relying purely on client debounce so a
// noisy client (or a different client) can't pile DEMO_KEY-eating requests
// onto the upstream.
var ErrQueryTooShort = fmt.Errorf("query must be at least %d characters", minQueryLen)

// ErrUpstreamSaturated is returned when the upstream limiter (or USDA itself
// via 429) prevents a request from going through and we have nothing to serve
// from cache.
var ErrUpstreamSaturated = errors.New("upstream saturated, try again shortly")

// ErrCallerRateLimited is returned when a single caller exceeds their per-key
// budget. Hand the corresponding 429 back to the client.
var ErrCallerRateLimited = errors.New("caller rate limited")

// FoodAPIClient wraps the USDA FDC search endpoint with:
//   - response caching keyed by normalized query
//   - an in-process upstream token bucket (so the whole process never
//     out-paces USDA, even under burst)
//   - a per-caller token bucket (so one user can't burn the whole budget)
//   - an atomic counter for observability
//
// Safe for concurrent use.
type FoodAPIClient struct {
	httpClient *http.Client
	apiKey     string

	upstream *tokenBucket

	cacheMu sync.RWMutex
	cache   map[string]cacheEntry

	callersMu sync.Mutex
	callers   map[string]*tokenBucket

	requestCount uint64
}

type cacheEntry struct {
	results   []FoodSummary
	expiresAt time.Time
}

// NewFoodAPIClient builds a client. USDA_API_KEY is read from env; if missing
// we fall back to the public DEMO_KEY which is heavily rate limited
// (~30 requests/hour, ~50/day per IP), so a real key should be set in any
// shared environment.
func NewFoodAPIClient() *FoodAPIClient {
	apiKey := os.Getenv("USDA_API_KEY")
	if apiKey == "" {
		apiKey = "DEMO_KEY"
		log.Println("[food_api] USDA_API_KEY not set; falling back to DEMO_KEY (very low rate limits)")
	}
	return &FoodAPIClient{
		httpClient: &http.Client{Timeout: upstreamTimeout},
		apiKey:     apiKey,
		// Conservative upstream budget: burst 10, sustained 2/s. With caching
		// and a min-query-length filter, this comfortably fits a real USDA
		// 1000/hour key and just barely fits DEMO_KEY for light dev use.
		upstream: newTokenBucket(10, 2),
		cache:    make(map[string]cacheEntry),
		callers:  make(map[string]*tokenBucket),
	}
}

// AllowCaller returns false if the given caller key (e.g. user/IP) is over
// their per-caller budget. Returning false should map to HTTP 429.
func (c *FoodAPIClient) AllowCaller(key string) bool {
	if key == "" {
		return true
	}
	c.callersMu.Lock()
	bucket, ok := c.callers[key]
	if !ok {
		// Burst 8, sustained 1/s feels reasonable for an autocomplete UI
		// that is also debounced client-side.
		bucket = newTokenBucket(8, 1)
		c.callers[key] = bucket
	}
	if len(c.callers) > 1024 {
		c.evictIdleCallersLocked()
	}
	c.callersMu.Unlock()
	return bucket.allow()
}

// RequestCount returns the total upstream USDA requests issued since process
// start. Useful for /metrics-style observability.
func (c *FoodAPIClient) RequestCount() uint64 {
	return atomic.LoadUint64(&c.requestCount)
}

// SearchFoods returns up to defaultPageSize short-form foods for the query.
// Repeated identical queries within cacheTTL are served from memory.
func (c *FoodAPIClient) SearchFoods(ctx context.Context, rawQuery string) ([]FoodSummary, error) {
	query := normalizeQuery(rawQuery)
	if len(query) < minQueryLen {
		return nil, ErrQueryTooShort
	}
	if len(query) > maxQueryLen {
		query = query[:maxQueryLen]
	}

	if hit, ok := c.lookupCache(query, false); ok {
		return hit, nil
	}

	if !c.upstream.allow() {
		// Last-ditch: serve a stale cache entry rather than 429-ing the user.
		if hit, ok := c.lookupCache(query, true); ok {
			return hit, nil
		}
		return nil, ErrUpstreamSaturated
	}

	results, err := c.fetchUpstream(ctx, query)
	if err != nil {
		return nil, err
	}
	c.storeCache(query, results)
	return results, nil
}

func (c *FoodAPIClient) fetchUpstream(ctx context.Context, query string) ([]FoodSummary, error) {
	body, err := json.Marshal(fdcSearchRequest{
		Query:           query,
		DataType:        []string{"Foundation", "SR Legacy"},
		PageSize:        defaultPageSize,
		SortBy:          "dataType.keyword",
		SortOrder:       "asc",
		RequireAllWords: true,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal search request: %w", err)
	}

	endpoint := fdcSearchURL + "?api_key=" + url.QueryEscape(c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "simpletracker/1.0")

	n := atomic.AddUint64(&c.requestCount, 1)
	if n%upstreamLogEvery == 0 {
		log.Printf("[food_api] upstream USDA requests since startup: %d", n)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("usda fdc search: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusTooManyRequests {
		return nil, ErrUpstreamSaturated
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("usda fdc search: status %d", res.StatusCode)
	}

	var parsed fdcSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode usda response: %w", err)
	}

	out := make([]FoodSummary, 0, len(parsed.Foods))
	for _, f := range parsed.Foods {
		out = append(out, summarizeFood(f))
	}
	return out, nil
}

func (c *FoodAPIClient) lookupCache(query string, ignoreExpiry bool) ([]FoodSummary, bool) {
	c.cacheMu.RLock()
	entry, ok := c.cache[query]
	c.cacheMu.RUnlock()
	if !ok {
		return nil, false
	}
	if !ignoreExpiry && time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.results, true
}

func (c *FoodAPIClient) storeCache(query string, results []FoodSummary) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	if len(c.cache) >= cacheMaxEntries {
		now := time.Now()
		for k, e := range c.cache {
			if now.After(e.expiresAt) {
				delete(c.cache, k)
			}
		}
		if len(c.cache) >= cacheMaxEntries {
			for k := range c.cache {
				delete(c.cache, k)
				break
			}
		}
	}
	c.cache[query] = cacheEntry{
		results:   results,
		expiresAt: time.Now().Add(cacheTTL),
	}
}

func (c *FoodAPIClient) evictIdleCallersLocked() {
	cutoff := time.Now().Add(-30 * time.Minute)
	for k, b := range c.callers {
		b.mu.Lock()
		idle := b.last.Before(cutoff) && b.tokens >= b.capacity
		b.mu.Unlock()
		if idle {
			delete(c.callers, k)
		}
	}
}

func normalizeQuery(q string) string {
	q = strings.ToLower(strings.TrimSpace(q))
	return strings.Join(strings.Fields(q), " ")
}

func summarizeFood(f fdcFood) FoodSummary {
	brand := strings.TrimSpace(f.BrandName)
	if brand == "" {
		brand = strings.TrimSpace(f.BrandOwner)
	}

	servingSize := f.ServingSize
	servingUnit := strings.ToLower(strings.TrimSpace(f.ServingSizeUnit))
	if servingSize <= 0 || servingUnit == "" {
		servingSize = 100
		servingUnit = "g"
	}

	var kcal1008, kcal2047, kcal2048 float64
	sum := FoodSummary{
		FdcID:       f.FdcID,
		Name:        strings.TrimSpace(f.Description),
		Brand:       brand,
		ServingSize: servingSize,
		ServingUnit: servingUnit,
	}
	for _, n := range f.FoodNutrients {
		switch n.NutrientID {
		case nutrientEnergyKcal:
			kcal1008 = n.Value
		case nutrientEnergyAtwaterGen:
			kcal2047 = n.Value
		case nutrientEnergyAtwaterSpec:
			kcal2048 = n.Value
		case nutrientProtein:
			sum.Protein = n.Value
		case nutrientCarbs:
			sum.Carbs = n.Value
		case nutrientFat:
			sum.Fat = n.Value
		case nutrientFiber:
			sum.Fiber = n.Value
		case nutrientSugar:
			sum.Sugar = n.Value
		}
	}
	sum.Calories = pickEnergyKcal(kcal1008, kcal2047, kcal2048)
	return sum
}

// pickEnergyKcal prefers label Energy (1008), then Atwater Specific (2048),
// then Atwater General (2047). If the preferred id is absent or zero, falls
// through so rows that only report Atwater energy still get calories.
func pickEnergyKcal(kcal1008, kcal2047, kcal2048 float64) float64 {
	if kcal1008 > 0 {
		return kcal1008
	}
	if kcal2048 > 0 {
		return kcal2048
	}
	if kcal2047 > 0 {
		return kcal2047
	}
	if kcal1008 != 0 {
		return kcal1008
	}
	if kcal2048 != 0 {
		return kcal2048
	}
	if kcal2047 != 0 {
		return kcal2047
	}
	return 0
}

// --- USDA wire types (only the fields we read) ---

type fdcSearchRequest struct {
	Query           string   `json:"query"`
	DataType        []string `json:"dataType,omitempty"`
	PageSize        int      `json:"pageSize,omitempty"`
	SortBy          string   `json:"sortBy,omitempty"`
	SortOrder       string   `json:"sortOrder,omitempty"`
	RequireAllWords bool     `json:"requireAllWords,omitempty"`
}

type fdcSearchResponse struct {
	Foods []fdcFood `json:"foods"`
}

type fdcFood struct {
	FdcID           int           `json:"fdcId"`
	Description     string        `json:"description"`
	BrandOwner      string        `json:"brandOwner"`
	BrandName       string        `json:"brandName"`
	ServingSize     float64       `json:"servingSize"`
	ServingSizeUnit string        `json:"servingSizeUnit"`
	FoodNutrients   []fdcNutrient `json:"foodNutrients"`
}

type fdcNutrient struct {
	NutrientID int     `json:"nutrientId"`
	Value      float64 `json:"value"`
	UnitName   string  `json:"unitName"`
}

// --- token bucket ---

// tokenBucket is a tiny lock-protected token bucket used for both the
// upstream-wide limiter and the per-caller limiter. Not export-worthy on its
// own; kept private to this file.
type tokenBucket struct {
	mu       sync.Mutex
	capacity float64
	tokens   float64
	refill   float64
	last     time.Time
}

func newTokenBucket(capacity int, perSecond float64) *tokenBucket {
	return &tokenBucket{
		capacity: float64(capacity),
		tokens:   float64(capacity),
		refill:   perSecond,
		last:     time.Now(),
	}
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(b.last).Seconds()
	b.last = now
	b.tokens += elapsed * b.refill
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}
