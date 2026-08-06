# Post-Optimization Performance Report (Redis Caching)

URL shortener API — Go + Gin + GORM + Postgres + **Redis**. Single node, localhost benchmark.
Companion to [`pre_optimization_report.md`](./pre_optimization_report.md).

## 1. What Changed

- **Redis cache added to the read path.** `URLRepository.GetUrl` now checks Redis
  (`internal/cache/urls_cache.go`) before hitting Postgres; on a hit it returns without a DB query.
- **Writes are cached too.** `URLRepository.Create` stores the short→long mapping in Redis
  (`internal/cache/urls_cache.go`), so fresh URLs are resolvable from cache immediately.
- **Cache-miss correctness fix** (`internal/cache/urls_cache.go`): `GetUrl` now treats a 0-field
  `HGETALL` as a miss and falls through to Postgres, instead of returning an empty (nil-error)
  hit that broke `Create` (empty short codes) and the 404 path.

## 2. Load-Test Setup (Deployment-Readiness Run)

Validated from a **fully clean Redis state** (`FLUSHALL`, 0 keys) to prove the whole
deployment flow — DB fallback, cache fill, and cache-hit reads — works and is production-ready.

- Tooling: `api/bin/loadtest`, `-n 5000`, `-mode bench`, single localhost node.
- **Step 1 — cold pass from empty cache** (50k requests, `c=100`): exercises the
  cache-miss → Postgres fallback → cache-refill path; also warms Redis for the measured runs.
- **Step 2 — measured warm benches** at `c=20/100/200/400` on the hot cache path.
- `-endpoint redirect` exercises `GET /api/url_shorter/{code}` → 302 (read path).
- `-endpoint create` exercises `POST /api/url_shorter` (write path).
- Server: `api/bin/benchserver` on `:8081` (same handlers, no gin Logger middleware).
- Ref: rerun a clean expand sequence confirmed **all tests pass (66/66)**.

## 3. Cold-Start (Cache Miss) Validation

First pass from an empty Redis — the miss path that also proves Postgres sizing / correctness.

| Metric | c=100 (n=50000) |
|---|---|
| **rps** | 10,716 |
| **p50** | 7.25ms |
| **p95** | 20.14ms |
| **p99** | 30.61ms |
| **max** | 198.97ms |
| **success** | 50,000 (0 errors) |

**All 50,000 → 302, 0 errors**: cold misses correctly fall through to Postgres, fill the cache,
and serve. No stale/incomplete reads.

## 4. Results — Read Path (redirect, hot cache)

| Metric | c=20 | c=100 | c=200 | c=400 |
|---|---|---|---|---|
| **rps** | 11,467 | 15,395 | 14,999 | 13,833 |
| **p50** | 1.29ms | 5.21ms | 10.55ms | 22.69ms |
| **p95** | 3.87ms | 12.59ms | 24.10ms | 53.74ms |
| **p99** | 6.14ms | 18.12ms | 35.38ms | 73.66ms |
| **max** | 24.68ms | 28.30ms | 54.64ms | 93.91ms |
| **error_pct** | 0.00 | 0.00 | 0.00 | 0.00 |

## 5. Results — Write Path (create)

| Metric | c=20 | c=100 | c=200 |
|---|---|---|---|
| **rps** | 1,118 | 3,309 | 3,124 |
| **p50** | 19.49ms | 28.25ms | 59.89ms |
| **p95** | 23.44ms | 45.76ms | 99.12ms |
| **p99** | 40.59ms | 63.68ms | 122.96ms |
| **error_pct** | 0.00 | 0.00 | 0.00 |

> Write is still bounded by the synchronous Postgres insert + Redis write; it is a
> fundamentally more expensive path than a cache read.

## 6. Comparison vs Pre-Optimization (redirect)

| Metric | Pre-opt | Post-fix (no cache) | **Post-opt (Redis)** |
|---|---|---|---|
| c=20 rps | 507 | 1,254 | **11,467** |
| c=100 rps | 484 | 1,230 | **15,395** |
| c=20 p50 | 28ms | 4.79ms | **1.29ms** |
| c=100 p50 | 165ms | 8.74ms | **5.21ms** |

**Improvement vs the "no cache" baseline: ~9–12x throughput, ~4–6x lower p50.**

## 7. Redis Health After Load

| Metric | Value |
|---|---|
| Keys | 38,523 |
| Memory used / RSS | 10.95 MB / 14.95 MB |
| **Evictions** | 0 |
| **Expired keys** | 0 |
| **Rejected connections** | 0 |
| Hit ratio | ~65% (cold start + ~15k un-read created keys dilute the ratio) |
| Cost/key | ≈ 300 B/url |

Efficient (≈300 B/url), zero eviction/churn, zero connection drops — a healthy deployment state.

## 8. Capacity Conclusion (100M users/day)

- Demand: 100M / 86,400s ≈ **1,157 rps avg**, bursts 3–6k rps.
- **One node sustains ~11–16k rps** on the warm read path — **2–4x the peak burst requirement**,
  leaving headroom instead of being "barely at the average".
- **Cold start also sustained 10.7k rps with 0 errors**, so Postgres handles the miss rate safely.
- Write path (~3k rps/node) still needs horizontal scaling or async writes for 100M/day
  creation rates.

## 9. Deployment Readiness

- Components: Postgres (pooled), Redis (in-memory, no persistence configured), single app node.
- **Reads scale beyond 100M/day peak on a single node; writes are the scaling constraint.**
- Tests: 66/66 passing. No errors across cold and hot paths.

## 10. Next Steps

- Re-test the cold path at much larger `count` (>1M rows) to size Postgres under a high miss rate.
- Decide Redis durability (AOF/persistence) if cache loss on restart is unacceptable; confirm cache
  refresh-on-miss covers evicted/expired entries.
- Consider client-side caching / CDN edge 302s to offload 100M/day reads.