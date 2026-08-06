package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/DanielJohn17/url-shortner/api/internal/cache"
	"github.com/DanielJohn17/url-shortner/api/internal/config"
	"github.com/DanielJohn17/url-shortner/api/internal/storage"
	"github.com/DanielJohn17/url-shortner/api/internal/urls"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomCode() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(b)
}

func main() {
	mode := flag.String("mode", "bench", "seed | bench")
	count := flag.Int("count", 10000, "number of rows to seed")
	endpoint := flag.String("endpoint", "redirect", "redirect | create")
	n := flag.Int("n", 100000, "total requests")
	c := flag.Int("c", 200, "concurrency")
	rps := flag.Int("rps", 0, "rate limit in req/s (0 = unlimited)")
	tag := flag.String("tag", "", "label for the result")
	seedFile := flag.String("seedfile", "", "optional file of seeded codes (one per line), reused from seed step")

	flag.Parse()

	switch *mode {
	case "seed":
		seed(*count)
	case "bench":
		runBench(*endpoint, *n, *c, *rps, *tag, *seedFile)
	default:
		fmt.Fprintln(os.Stderr, "unknown mode", *mode)
		os.Exit(2)
	}
}

func connect() *gorm.DB {
	db, err := storage.NewDatabase(storage.DBConfig{
		DBHost:     config.Env.DBHost,
		DBUser:     config.Env.DBUser,
		DBPassword: config.Env.DBPassword,
		DBName:     config.Env.DBName,
		DBPort:     config.Env.DBPort,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "db connect failed:", err)
		os.Exit(1)
	}
	return db
}

func connectCache() *cache.URLCacheRepository {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Env.RedisAddr,
		Password: config.Env.RedisPassword,
		DB:       int(config.Env.RedisDB),
		Protocol: int(config.Env.RedisProtocol),
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		fmt.Fprintln(os.Stderr, "redis connect failed:", err)
		os.Exit(1)
	}
	return cache.NewUrlCacheRepository(rdb)
}

func seed(count int) {
	db := connect()
	repo := urls.NewURLRepository(db, connectCache())
	ctx := context.Background()

	seen := map[string]bool{}
	start := time.Now()
	for len(seen) < count {
		code := randomCode()
		if seen[code] {
			continue
		}
		seen[code] = true
		u := &urls.URL{
			ShortUrl: code,
			LongUrl:  fmt.Sprintf("https://example.com/seed/%s", code),
		}
		if _, err := repo.Create(ctx, u); err != nil {
			continue
		}
	}
	fmt.Printf("seed_done count=%d rows=%d elapsed=%s\n", count, len(seen), time.Since(start))
}

func runBench(endpoint string, total, conc, rateRPS int, tag, seedFile string) {
	base := os.Getenv("LOADTEST_BASE")
	if base == "" {
		base = "http://localhost:8080"
	}

	var codes []string
	if endpoint == "redirect" {
		if seedFile != "" {
			codes = loadCodes(seedFile)
		} else {
			codes = loadCodesFromDB()
		}
		if len(codes) == 0 {
			fmt.Fprintln(os.Stderr, "no seeded codes found")
			os.Exit(1)
		}
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        1000,
			MaxIdleConnsPerHost: 1000,
			MaxConnsPerHost:     0,
		},
	}

	// rate limiter: a channel delivering one token per request at the target rate
	var limiter chan struct{}
	if rateRPS > 0 {
		limiter = make(chan struct{}, conc*10)
		go func() {
			ticker := time.NewTicker(time.Second / time.Duration(rateRPS))
			defer ticker.Stop()
			for range ticker.C {
				select {
				case limiter <- struct{}{}:
				default:
				}
			}
		}()
	}

	jobs := make(chan int)
	var wg sync.WaitGroup
	start := time.Now()
	merged := &resultCollector{lats: make([]time.Duration, 0, total), status: map[int]int{}}

	worker := func(wid int) {
		defer wg.Done()
		loc := &resultCollector{lats: make([]time.Duration, 0, total/conc + 1000), status: map[int]int{}}
		for range jobs {
			if limiter != nil {
				<-limiter
			}
			reqWall := time.Now()
			code := ""
			if endpoint == "redirect" {
				code = codes[rand.Intn(len(codes))]
			}
			resp, err := doRequest(client, base, endpoint, code, wid)
			dur := time.Since(reqWall)
			if err != nil {
				loc.status[0]++ // 0 = transport/error
			} else {
				loc.status[resp.StatusCode]++
			}
			loc.lats = append(loc.lats, dur)
			if resp != nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}
		merged.mu.Lock()
		for k, v := range loc.status {
			merged.status[k] += v
		}
		merged.lats = append(merged.lats, loc.lats...)
		merged.mu.Unlock()
	}

	for w := 0; w < conc; w++ {
		wg.Add(1)
		go worker(w)
	}
	for i := 0; i < total; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	elapsed := time.Since(start)

	report(endpoint, tag, total, conc, rateRPS, elapsed, merged.lats, merged.status)
}

type resultCollector struct {
	mu     sync.Mutex
	lats   []time.Duration
	status map[int]int
}

func doRequest(client *http.Client, base, endpoint, code string, wid int) (*http.Response, error) {
	var path string
	var body io.Reader
	method := http.MethodGet
	if endpoint == "redirect" {
		path = base + "/api/url_shorter/" + code
	} else {
		method = http.MethodPost
		path = base + "/api/url_shorter"
		pl := fmt.Sprintf(`{"long_url":"https://loadtest.example.com/%d/%d"}`, wid, rand.Int63())
		body = strings.NewReader(pl)
	}
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}

func report(endpoint, tag string, total, conc, rateRPS int, elapsed time.Duration, lats []time.Duration, status map[int]int) {
	sort.Slice(lats, func(i, j int) bool { return lats[i] < lats[j] })
	pct := func(p float64) time.Duration {
		if len(lats) == 0 {
			return 0
		}
		idx := int(float64(len(lats)) * p)
		if idx >= len(lats) {
			idx = len(lats) - 1
		}
		return lats[idx]
	}

	ok := 0
	for k, v := range status {
		if k >= 200 && k < 400 {
			ok += v
		}
	}
	errs := total - ok

	fmt.Println("===RESULT===")
	fmt.Printf("tag=%s\n", tag)
	fmt.Printf("endpoint=%s\n", endpoint)
	fmt.Printf("total=%d concurrency=%d rate_limit_rps=%d\n", total, conc, rateRPS)
	fmt.Printf("elapsed_s=%.3f\n", elapsed.Seconds())
	fmt.Printf("achieved_rps=%.1f\n", float64(total)/elapsed.Seconds())
	fmt.Printf("p50_ms=%.3f\n", float64(pct(0.5))/float64(time.Millisecond))
	fmt.Printf("p95_ms=%.3f\n", float64(pct(0.95))/float64(time.Millisecond))
	fmt.Printf("p99_ms=%.3f\n", float64(pct(0.99))/float64(time.Millisecond))
	var max time.Duration
	if len(lats) > 0 {
		max = lats[len(lats)-1]
	}
	fmt.Printf("max_ms=%.3f\n", float64(max)/float64(time.Millisecond))
	fmt.Printf("success=%d errors=%d error_pct=%.2f\n", ok, errs, float64(errs)/float64(total)*100)
	b, _ := json.Marshal(status)
	fmt.Printf("status_dist=%s\n", string(b))
}

func connectRaw() *gorm.DB { return connect() }

func loadCodesFromDB() []string {
	db := connectRaw()
	var all []urls.URL
	if err := db.Find(&all).Error; err != nil {
		fmt.Fprintln(os.Stderr, "failed to load codes:", err)
		os.Exit(1)
	}
	codes := make([]string, 0, len(all))
	for _, u := range all {
		codes = append(codes, u.ShortUrl)
	}
	return codes
}

func loadCodes(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read seedfile:", err)
		os.Exit(1)
	}
	var codes []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			codes = append(codes, line)
		}
	}
	return codes
}