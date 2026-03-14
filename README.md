# Interview-IQ

A LeetCode-style interview prep tracker that scores your problem-solving sessions and uses time-based decay to surface what you need to review — not what you already know cold.

**Free tier** — log up to 20 problems, full scoring and decay, skill radar, problem history. No credit card required.
**Pro tier** — unlimited problem logging, AI-powered recommendations, and priority support. Payments via Stripe.

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql&logoColor=white)
![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=black)
![TypeScript](https://img.shields.io/badge/TypeScript-strict-3178C6?logo=typescript&logoColor=white)
![Clerk](https://img.shields.io/badge/Auth-Clerk-6C47FF?logo=clerk&logoColor=white)
![Railway](https://img.shields.io/badge/Backend-Railway-0B0D0E?logo=railway&logoColor=white)
![Vercel](https://img.shields.io/badge/Frontend-Vercel-000000?logo=vercel&logoColor=white)

---

## What It Does

When you solve a LeetCode problem, you log it: how many attempts it took, whether you peeked at the solution, whether you reached an optimal or brute-force solution, and any notes on your approach. Interview-IQ assigns a score (0–100), then decays that score over time as you'd naturally start forgetting the pattern.

Across all your logged problems, the app aggregates strength scores per category (arrays, graphs, dynamic programming, etc.) and visualizes them as both a radar chart and a bar chart. Your weakest category is surfaced automatically. A persistent **Retry Panel** ranks the problems most worth revisiting right now, and the AI-powered **Recommendations** page generates targeted practice suggestions using GPT-4o-mini based on your actual history.

**Core loop:**
1. Solve a problem on LeetCode
2. Search for it by name in the auto-complete log form — categories and difficulty auto-fill
3. Add how many attempts it took, whether you peeked, solution type, and optional notes
4. Score is computed and stored
5. Over days/weeks, your scores decay — the radar chart shifts to reflect what you've retained
6. The Retry Panel and AI Recommendations tell you exactly what to practice next

---

## Screenshots

> Dashboard — skill radar + category bar chart + weakest-category alert + AI recommendation popover

> Problem list — server-side search, multi-filter sidebar, paginated table with decayed scores and deduplication badges

> Problem detail — score history chart, attempt history table, notes, and decay breakdown

> Recommendations — AI-generated problem suggestions per category with difficulty, description, and rationale

---

## Scoring Model

### Raw Score (computed once at log time)

| Condition | Effect |
|---|---|
| Base score | 100 |
| Each extra attempt beyond the first | −10 (capped at −40) |
| Looked at the solution | −25 |
| Brute-force solution only | −15 |
| Absolute minimum | 5 |

Examples:
- 1 attempt, optimal solution, no peek → **100**
- 3 attempts, no peek, optimal → **80**
- 5+ attempts, no peek → **60** (penalty capped)
- 1 attempt, peeked → **75**
- 1 attempt, brute-force only → **85**
- 5+ attempts, peeked → **35**

### Decay (applied at read time, never stored)

The raw score decays linearly after a 7-day grace period, flooring at 40% of the original. This simulates the forgetting curve — a problem you solved three weeks ago is worth less than one you nailed yesterday.

| Days since solving | Effect on a 100-point score |
|---|---|
| 0–7 | 100 (no decay) |
| 14 | 93 |
| 21 | 86 |
| 30 | 77 |
| 60+ | 40 (floor) |

A background cron job runs daily at 10 PM EST to persist the decayed scores to the database, keeping the `DecayAllProblems` calculation cheap at read time.

### Category Strength

Per-category strength is computed from the **latest attempt** at each unique problem. Re-solving a problem always updates your strength in that category, never drags it down with old attempts.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.26 |
| HTTP router | Chi v5 |
| Database | PostgreSQL 16 |
| DB access | `database/sql` + `lib/pq` (no ORM) |
| Authentication | Clerk (RS256 JWT verification) |
| AI recommendations | OpenAI GPT-4o-mini |
| Rate limiting | Token-bucket per IP + per user (`golang.org/x/time/rate`) |
| Scheduled jobs | Background goroutine (daily decay cron) |
| Payments | Stripe (pro tier, live) |
| Frontend | React 19 + Vite + TypeScript (strict) |
| UI components | ShadCN/UI |
| Charts | Recharts (radar + bar chart + line chart) |
| HTTP client | Axios |
| Routing | React Router |
| Error monitoring | Sentry (backend + frontend) |
| Product analytics | PostHog |
| Backend hosting | Railway |
| Frontend hosting | Vercel |

---

## Features

### Backend
- **Clerk authentication** — Clerk handles identity; the backend verifies RS256 JWTs and upserts users on first sign-in
- **LeetCode problem catalog** — a seeded `leetcode_problems` table with full-text GIN index powers typeahead search
- **Multi-category tagging** — each problem can be tagged with multiple categories (`categories TEXT[]`)
- **Solution type** — log whether you reached an optimal or brute-force solution; brute-force carries a −15 point penalty
- **Notes** — optional free-text field per attempt to record approach, edge cases, or learnings
- **Server-side search, filter, and pagination** — the `/api/problems` endpoint supports `q`, `categories`, `difficulties`, `score_min`, `score_max`, `date_from`, `date_to`, `limit`, and `offset` query params
- **AI recommendations** — `GET /api/recommendations` calls GPT-4o-mini with a structured prompt built from your actual problem history; auto-selects categories below 60 strength (or weakest if all ≥ 60); excludes already-attempted problems with score ≥ 75
- **Daily decay cron** — background goroutine runs `DecayAllProblems` nightly at 10 PM EST so stored `decayed_score` stays fresh
- **Rate limiting** — token-bucket limiters applied per IP (unauthenticated routes) and per user ID (authenticated routes); idle entries are pruned in the background
- **Input sanitization** — all string inputs stripped and length-validated before reaching the service layer
- **Score by latest attempt** — category stats are computed from the most recent attempt per problem, not an average across all attempts
- **Decay at read time** — `ApplyDecay` is never persisted for ad-hoc reads; it is computed fresh on every direct read

### Frontend
- **Clerk-powered auth** — sign-in and sign-up handled by Clerk's hosted UI; no custom login forms to maintain
- **ShadCN/UI sidebar layout** — collapsible sidebar with dark/light mode toggle (persisted preference)
- **Skill radar chart** — all 21 categories shown even if strength is 0, so gaps are visible
- **Category bar chart** — precise per-category strength values alongside the radar
- **Weakest category banner** — dashboard surfaces your lowest-strength category with 3 targeted problem recommendations
- **AI recommendation popover** — one-click AI recommendations from the dashboard toolbar, plus a dedicated Recommendations page with category filter and per-category result cards (problem name, difficulty, description, rationale)
- **Retry Panel** — fixed right sidebar ranking your top 8 problems to revisit, scored by `(100 − score) × (100 − category_weakness) / 100`; links directly to each problem's detail page
- **LeetCode typeahead** — log form searches the problem catalog as you type (600 ms debounce); selecting a problem auto-fills difficulty and categories
- **Multi-category selector** — log a problem against one or more categories with a badge-based picker
- **Notes field** — optional textarea on the log form for recording your approach or edge cases
- **Problem detail page** — per-problem view with decay breakdown, score history line chart, full attempt history table, and saved notes
- **Problem list with filters** — filter by name, category, difficulty, score range (preset buckets), and date range (preset windows); paginated at 20 per page
- **Problem deduplication badges** — "Latest" and "Earlier attempt" tags on repeated problem entries in the list
- **Decay tooltip** — hovering a decayed score shows the original score, decay amount (−X in red), and time since solved
- **Loading skeleton + progress bar** — feedback on every data fetch
- **LeetCode-style difficulty badges** — color-coded Easy / Medium / Hard badges
- **Dark/light mode** — persisted theme preference with a single toggle

---

## Tiers

| Feature | Free | Pro |
|---|---|---|
| Problems logged | Up to 20 | Unlimited |
| Scoring & decay | ✓ | ✓ |
| Skill radar & bar chart | ✓ | ✓ |
| Problem detail & history | ✓ | ✓ |
| Retry Panel | ✓ | ✓ |
| AI Recommendations | — | ✓ |
| Priority support | — | ✓ |

Pro subscriptions are billed via Stripe and managed through the Stripe billing portal. Users can upgrade, downgrade, or cancel at any time — access continues until the end of the billing period.

---

## API Reference

All protected endpoints require a Clerk-issued JWT in the `Authorization: Bearer <token>` header.

### Problems

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/problems` | Clerk JWT | List problems with decayed scores (filterable, paginated) |
| `POST` | `/api/problems` | Clerk JWT | Log a new problem attempt |
| `GET` | `/api/problems/{problemID}` | Clerk JWT | Get a single problem by ID |

#### Query parameters for `GET /api/problems`

| Parameter | Type | Description |
|---|---|---|
| `q` | string | Full-text search on problem name |
| `categories` | string (comma-separated) | Filter by one or more categories |
| `difficulties` | string (comma-separated) | `easy`, `medium`, `hard` |
| `score_min` | int | Minimum decayed score (0–100) |
| `score_max` | int | Maximum decayed score (0–100) |
| `date_from` | string | ISO 8601 date, inclusive lower bound on `solved_at` |
| `date_to` | string | ISO 8601 date, inclusive upper bound on `solved_at` |
| `limit` | int | Results per page (default: 20, max: 100) |
| `offset` | int | Zero-based offset for pagination |

### Categories

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/categories/stats` | Clerk JWT | Per-category strength scores (0–100) |
| `GET` | `/api/categories/weakest` | Clerk JWT | Weakest category + 3 recommended problems |

### Recommendations

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/recommendations` | Clerk JWT | AI-generated problem recommendations per category |

#### Query parameters for `GET /api/recommendations`

| Parameter | Type | Description |
|---|---|---|
| `category` | string (repeatable) | Limit to specific categories; omit to auto-select weak ones |
| `from` | string | ISO 8601 date lower bound for practice history context |
| `to` | string | ISO 8601 date upper bound for practice history context |
| `limit` | int | Recommendations per category (1–10, default: 3) |

### LeetCode Problem Search

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/leetcode-problems/search` | Clerk JWT | Typeahead search against the LeetCode catalog |

### Billing

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/billing/status` | Clerk JWT | Returns subscription tier, problem count, and limit (0 = unlimited for Pro) |
| `POST` | `/api/billing/checkout` | Clerk JWT | Creates a Stripe Checkout Session; body: `{"plan": "monthly"\|"annual"}` |
| `POST` | `/api/billing/portal` | Clerk JWT | Creates a Stripe Billing Portal session for managing/cancelling a subscription |
| `POST` | `/api/webhooks/stripe` | None (Stripe signature) | Stripe webhook handler — processes subscription lifecycle events |

### Health & Docs

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/health` | None | Health check — returns `200` if DB is reachable, `503` otherwise |
| `GET` | `/swagger/*` | None | Interactive API documentation (Swagger UI) |

### Example: Log a Problem

```http
POST /api/problems
Authorization: Bearer <clerk-jwt>
Content-Type: application/json

{
  "name": "Two Sum",
  "categories": ["array", "hash-map"],
  "difficulty": "easy",
  "attempts": 1,
  "looked_at_solution": false,
  "solution_type": "optimal",
  "time_taken_mins": 12,
  "notes": "Used a hash map to store complement → index. O(n) time, O(n) space."
}
```

```json
{
  "id": 42,
  "name": "Two Sum",
  "categories": ["array", "hash-map"],
  "difficulty": "easy",
  "attempts": 1,
  "looked_at_solution": false,
  "solution_type": "optimal",
  "time_taken_mins": 12,
  "score": 100,
  "decayed_score": 100,
  "notes": "Used a hash map to store complement → index. O(n) time, O(n) space.",
  "solved_at": "2026-03-08T14:30:00Z",
  "created_at": "2026-03-08T14:30:00Z"
}
```

### Valid Categories (21 total)

```
array, string, hash-map, two-pointers, sliding-window, binary-search,
stack, queue, linked-list, tree, trie, graph, advanced-graphs, heap,
dp, dp-2d, backtracking, greedy, intervals, math, bit-manipulation, other
```

### Standard Error Shape

```json
{ "error": "human-readable message" }
```

---

## Getting Started

### Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Node.js 20+ and pnpm](https://pnpm.io/installation)
- [Docker](https://www.docker.com/) (for local PostgreSQL)
- A free [Clerk](https://clerk.com) account — create an app and grab your keys
- An [OpenAI](https://platform.openai.com) API key (for AI recommendations)
- A [Stripe](https://stripe.com) account (for billing; test mode is fine locally)

### 1. Clone and set up environment variables

```bash
git clone https://github.com/apgupta3091/interview-iq.git
cd interview-iq
```

**Backend** — create `backend/.env`:
```bash
PORT=8080
APP_ENV=development
DATABASE_URL=postgres://interviewiq:interviewiq_secret@localhost:5432/interviewiq?sslmode=disable
CLERK_SECRET_KEY=sk_test_...
OPENAI_API_KEY=sk-...
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PRICE_MONTHLY=price_...
STRIPE_PRICE_ANNUAL=price_...
FRONTEND_URL=http://localhost:5173
# SENTRY_DSN=  (optional locally)
```

**Frontend** — create `frontend/.env.local`:
```bash
VITE_API_URL=http://localhost:8080
VITE_CLERK_PUBLISHABLE_KEY=pk_test_...
VITE_STRIPE_PUBLISHABLE_KEY=pk_test_...
# VITE_SENTRY_DSN=  (optional locally)
# VITE_POSTHOG_KEY=  (optional locally)
```

### 2. Start the database

```bash
make db-up
```

### 3. Run the backend

```bash
make backend
```

Migrations run automatically on first start. The API will be available at `http://localhost:8080`.

### 4. Run the frontend

```bash
cd frontend
pnpm install
pnpm dev
```

The app will be available at `http://localhost:5173`.

---

## Development

```bash
make dev        # Start DB + run server together
make test       # Run all backend tests
make test-v     # Run tests with verbose output
make lint       # go vet all packages
make tidy       # Tidy go.mod/go.sum
make db-reset   # Wipe DB volume and restart fresh
```

### Seed realistic data (dev only)

A seed script generates 79 realistic problem attempts across all categories to populate the dashboard during development:

```bash
cd backend
SEED_USER_ID=1 go run ./cmd/seed
```

---

## Deployment

The production stack runs on:
- **Backend + Postgres** → [Railway](https://railway.app) (Go service + managed Postgres, auto-deploys from `main` via GitHub Actions)
- **Frontend** → [Vercel](https://vercel.com) (auto-deploys from `main` on push)
- **Domain + DNS** → Cloudflare (`interview-iq.com`)

### Production environment variables

**Railway (backend):**

| Variable | Notes |
|---|---|
| `DATABASE_URL` | Auto-injected by Railway Postgres |
| `PORT` | `8080` |
| `APP_ENV` | `production` |
| `CLERK_SECRET_KEY` | Production Clerk secret |
| `OPENAI_API_KEY` | Production OpenAI key |
| `STRIPE_SECRET_KEY` | Live Stripe secret |
| `STRIPE_WEBHOOK_SECRET` | From Stripe → Webhooks |
| `STRIPE_PRICE_MONTHLY` | Live monthly price ID |
| `STRIPE_PRICE_ANNUAL` | Live annual price ID |
| `FRONTEND_URL` | `https://interview-iq.com` |
| `SENTRY_DSN` | Backend Sentry DSN |

**Vercel (frontend):**

| Variable | Notes |
|---|---|
| `VITE_API_URL` | `https://api.interview-iq.com` |
| `VITE_CLERK_PUBLISHABLE_KEY` | Production Clerk publishable key |
| `VITE_STRIPE_PUBLISHABLE_KEY` | Live Stripe publishable key |
| `VITE_SENTRY_DSN` | Frontend Sentry DSN |
| `VITE_POSTHOG_KEY` | PostHog project key |

**GitHub Secrets (for CI):**

| Secret | Notes |
|---|---|
| `RAILWAY_TOKEN` | Railway API token (Account → API Tokens) |
| `VITE_CLERK_PUBLISHABLE_KEY` | Used during frontend CI build |

---

## Project Structure

```
interview-iq/
├── backend/
│   ├── Dockerfile                      # Multi-stage Alpine build for Railway
│   ├── cmd/
│   │   └── server/main.go              # Entry point: env, DI wiring, router setup
│   ├── migrations/
│   │   ├── 001_init.sql                # users + problems schema
│   │   ├── 002_clerk_auth.sql          # clerk_user_id column; nullable email
│   │   ├── 002_multi_category.sql      # categories TEXT[] replaces single category
│   │   ├── 003_leetcode_problems.sql   # leetcode_problems catalog + GIN index
│   │   ├── 004_solution_type.sql       # solution_type column (none/brute_force/optimal)
│   │   ├── 005_nullable_email.sql      # make email nullable for Clerk-only sign-in
│   │   ├── 006_original_score.sql      # original_score column; backfill + recompute decayed scores
│   │   ├── 007_notes.sql               # notes TEXT column on problems
│   │   └── 008_billing.sql             # stripe_customer_id + subscription_tier on users
│   └── internal/
│       ├── handlers/                   # HTTP layer — thin, no business logic
│       │   ├── problems.go             # List, Log, GetByID
│       │   ├── categories.go           # GetStats, GetWeakest
│       │   ├── recommendations.go      # AI-powered recommendations
│       │   ├── leetcode.go             # LeetCode catalog search
│       │   └── helpers.go             # writeJSON, writeError
│       ├── service/                    # Validation + business logic
│       │   ├── problem_service.go
│       │   ├── category_service.go
│       │   └── recommendation_service.go  # OpenAI GPT-4o-mini integration
│       ├── repository/                 # SQL only — no business logic
│       │   ├── user_repo.go
│       │   ├── problem_repo.go         # Includes DecayAllProblems
│       │   ├── category_repo.go
│       │   └── leetcode_repo.go
│       ├── models/                     # Domain types + scoring functions
│       │   ├── types.go
│       │   └── score.go               # ComputeScore, ApplyDecay
│       ├── middleware/
│       │   ├── auth.go                # ClerkAuthenticate: verifies RS256 JWT, upserts user
│       │   └── rate_limit.go          # Per-IP and per-user token-bucket limiters
│       └── cron/
│           └── decay.go               # Daily decay cron (10 PM EST)
└── frontend/
    └── src/
        ├── pages/
        │   ├── Dashboard.tsx           # Radar + bar chart + weakest banner + AI popover
        │   ├── ProblemList.tsx         # Paginated table with filter sidebar + dedup badges
        │   ├── ProblemDetail.tsx       # Score history, attempt table, notes, decay breakdown
        │   ├── LogProblem.tsx          # Log form: typeahead, multi-category, notes
        │   ├── Recommendations.tsx     # AI recommendations page
        │   ├── Pricing.tsx             # Upgrade page with plan comparison + Stripe checkout
        │   ├── PrivacyPolicy.tsx       # Privacy policy (/privacy)
        │   └── Terms.tsx               # Terms of service (/terms)
        ├── components/
        │   ├── AppLayout.tsx           # Layout shell (sidebar + retry panel)
        │   ├── AppSidebar.tsx          # Left nav + theme toggle + sign out
        │   ├── RetryPanel.tsx          # Right sidebar: prioritized retry queue
        │   ├── ProblemFilters.tsx      # Filter controls (search, category, difficulty, date, score)
        │   ├── CategoryRadarChart.tsx
        │   ├── CategoryBarChart.tsx
        │   └── ui/                    # ShadCN auto-generated (do not hand-edit)
        ├── lib/
        │   ├── api/                   # Per-resource Axios wrappers
        │   │   ├── client.ts          # Axios instance + Clerk JWT interceptor + 401 handler
        │   │   ├── problems.ts
        │   │   ├── categories.ts
        │   │   ├── leetcode.ts
        │   │   └── recommendations.ts
        │   └── constants.ts           # CATEGORIES list (21 values)
        ├── hooks/                     # Custom React hooks
        ├── types/api.ts               # TypeScript types mirroring API responses
        └── main.tsx                   # Routes + ClerkProvider
```

The backend follows a strict 3-layer architecture: **handlers** call **services**, services call **repositories**. Layers communicate via interfaces and are independently testable.

---

## Architecture Notes

### Authentication flow (Clerk)

1. The frontend obtains a short-lived JWT from Clerk via `window.Clerk.session.getToken()`
2. The Axios interceptor attaches it as `Authorization: Bearer <token>` on every request
3. The `ClerkAuthenticate` middleware verifies the RS256 JWT against Clerk's JWKS endpoint
4. On first sign-in, the middleware upserts a row in `users` keyed on `clerk_user_id` and injects the internal integer `user_id` into the request context
5. All downstream handlers read `userID` from context — never from the request body

### Score model invariant

`ComputeScore` is called **once at write time** and stored in the DB. `ApplyDecay` is called **at read time only** and is never persisted for individual reads. The nightly cron job persists decayed scores in bulk so the `decayed_score` column stays current without per-request computation overhead.

### AI recommendations

`RecommendationService` builds a structured prompt from the user's actual problem history and posts it to the OpenAI chat completions API (`gpt-4o-mini`, JSON response format). The response is post-filtered to exclude problems the user has already attempted with a score ≥ 75. Categories below 60 strength are auto-selected if no explicit category filter is provided; if all categories are ≥ 60, the weakest is used.

### Retry Panel ranking

Problems are ranked by: `(100 − score) × (100 − weakest_category_strength) / 100`. Only problems with a score below 80 are eligible. The top 8 are shown. This surfaces problems at the intersection of personal weakness and category weakness.

### Rate limiting

Two independent token-bucket limiters run in the middleware chain:

| Limiter | Scope | Sustained | Burst |
|---|---|---|---|
| IP limiter | All routes | 60 req/min | 20 |
| User limiter | Authenticated routes | 120 req/min | 40 |

Idle limiter entries are pruned in a background goroutine to prevent unbounded memory growth.

---

## License

MIT
