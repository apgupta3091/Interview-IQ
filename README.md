# Interview-IQ

A LeetCode-style interview prep tracker that scores your problem-solving sessions and uses time-based decay to surface what you need to review — not what you already know cold.

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql&logoColor=white)
![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=black)
![TypeScript](https://img.shields.io/badge/TypeScript-strict-3178C6?logo=typescript&logoColor=white)
![Clerk](https://img.shields.io/badge/Auth-Clerk-6C47FF?logo=clerk&logoColor=white)

---

## What It Does

When you solve a LeetCode problem, you log it: how many attempts it took, whether you peeked at the solution, and whether you reached an optimal or brute-force solution. Interview-IQ assigns a score (0–100), then decays that score over time as you'd naturally start forgetting the pattern.

Across all your logged problems, the app aggregates strength scores per category (arrays, graphs, dynamic programming, etc.) and visualizes them as both a radar chart and a bar chart. Your weakest category is surfaced automatically with targeted problem recommendations so you know exactly where to focus next.

**Core loop:**
1. Solve a problem on LeetCode
2. Search for it by name in the auto-complete log form — categories and difficulty auto-fill
3. Add how many attempts it took and whether you peeked at the solution
4. Score is computed and stored
5. Over days/weeks, your scores decay — the radar chart shifts to reflect what you've retained
6. The app tells you your weakest category and suggests what to practice next

---

## Screenshots

> Dashboard — skill radar + category bar chart with weakest-category alert

> Problem list — server-side search, multi-filter sidebar, paginated table with decayed scores

---

## Scoring Model

### Raw Score (computed once at log time)

| Condition | Effect |
|---|---|
| Base score | 100 |
| Each extra attempt beyond the first | −10 (capped at −40) |
| Looked at the solution | −25 |
| Brute-force solution only | −10 |
| Absolute minimum | 5 |

Examples:
- 1 attempt, optimal solution, no peek → **100**
- 3 attempts, no peek, no solution type → **80**
- 5+ attempts, no peek → **60** (penalty capped)
- 1 attempt, peeked → **75**
- 1 attempt, brute-force only → **90**
- 5+ attempts, peeked → **35**

### Decay (applied at read time, never stored)

The raw score decays linearly after a 3-day grace period, flooring at 30% of the original. This simulates the forgetting curve — a problem you solved three weeks ago is worth less than one you nailed yesterday.

| Days since solving | Effect on a 100-point score |
|---|---|
| 0–3 | 100 (no decay) |
| 10 | 86 |
| 21 | 64 |
| 30+ | 30 (floor) |

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
| Rate limiting | Token-bucket per IP + per user (`golang.org/x/time/rate`) |
| Frontend | React 19 + Vite + TypeScript (strict) |
| UI components | ShadCN/UI |
| Charts | Recharts (radar + bar chart) |
| HTTP client | Axios |
| Routing | React Router |

---

## Features

### Backend
- **Clerk authentication** — Clerk handles identity; the backend verifies RS256 JWTs and upserts users on first sign-in
- **LeetCode problem catalog** — a seeded `leetcode_problems` table with full-text GIN index powers typeahead search
- **Multi-category tagging** — each problem can be tagged with multiple categories (`categories TEXT[]`)
- **Solution type** — log whether you reached an optimal or brute-force solution; scores reflect the difference
- **Server-side search, filter, and pagination** — the `/api/problems` endpoint supports `q`, `categories`, `difficulties`, `score_min`, `score_max`, `date_from`, `date_to`, `page`, and `page_size` query params
- **Rate limiting** — token-bucket limiters applied per IP (unauthenticated routes) and per user ID (authenticated routes); idle entries are pruned in the background
- **Input sanitization** — all string inputs stripped and length-validated before reaching the service layer
- **Score by latest attempt** — category stats are computed from the most recent attempt per problem, not an average across all attempts
- **Decay at read time** — `ApplyDecay` is never persisted; it is computed fresh on every read

### Frontend
- **Clerk-powered auth** — sign-in and sign-up handled by Clerk's hosted UI; no custom login forms to maintain
- **ShadCN/UI sidebar layout** — collapsible sidebar with dark/light mode toggle
- **Skill radar chart** — all 17+ categories shown even if strength is 0, so gaps are visible
- **Category bar chart** — precise per-category strength values alongside the radar
- **Weakest category banner** — dashboard surfaces your lowest-strength category with 3 targeted problem recommendations
- **LeetCode typeahead** — log form searches the problem catalog as you type; selecting a problem auto-fills difficulty
- **Multi-category selector** — log a problem against one or more categories with a badge-based picker
- **Problem list with filters** — filter by name, category, difficulty, score range, and date range; filters apply on demand (not on every keystroke); paginated at 20 per page
- **Loading skeleton + progress bar** — feedback on every data fetch
- **LeetCode-style difficulty badges** — color-coded Easy / Medium / Hard badges consistent with leetcode.com conventions
- **Dark/light mode** — persisted theme preference with a single toggle

---

## API Reference

All protected endpoints require a Clerk-issued JWT in the `Authorization: Bearer <token>` header.

### Problems

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/problems` | Clerk JWT | List problems with decayed scores (filterable, paginated) |
| `POST` | `/api/problems` | Clerk JWT | Log a new problem attempt |

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
| `page` | int | Page number (1-indexed, default: 1) |
| `page_size` | int | Results per page (default: 20, max: 100) |

### Categories

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/categories/stats` | Clerk JWT | Per-category strength scores (0–100) |
| `GET` | `/api/categories/weakest` | Clerk JWT | Weakest category + 3 recommended problems |

### LeetCode Problem Search

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/leetcode-problems/search` | Clerk JWT | Typeahead search against the LeetCode catalog |

### Health

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/health` | None | Health check |

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
  "time_taken_mins": 12
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
  "solved_at": "2026-03-08T14:30:00Z",
  "created_at": "2026-03-08T14:30:00Z"
}
```

### Valid Categories

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

### 1. Clone and set up environment variables

```bash
git clone https://github.com/apgupta3091/interview-iq.git
cd interview-iq
```

**Backend** — create `backend/.env`:
```bash
PORT=8080
DATABASE_URL=postgres://interviewiq:interviewiq_secret@localhost:5432/interviewiq?sslmode=disable
CLERK_SECRET_KEY=sk_test_...
```

**Frontend** — create `frontend/.env.local`:
```bash
VITE_API_URL=http://localhost:8080
VITE_CLERK_PUBLISHABLE_KEY=pk_test_...
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

## Project Structure

```
interview-iq/
├── backend/
│   ├── cmd/
│   │   └── server/main.go          # Entry point: env, DI wiring, router setup
│   ├── migrations/
│   │   ├── 001_init.sql            # users + problems schema
│   │   ├── 002_clerk_auth.sql      # clerk_user_id column; nullable email
│   │   ├── 002_multi_category.sql  # categories TEXT[] replaces single category
│   │   ├── 003_leetcode_problems.sql # leetcode_problems catalog + GIN index
│   │   ├── 004_solution_type.sql   # solution_type column (none/brute_force/optimal)
│   │   └── 005_nullable_email.sql  # make email nullable for Clerk-only sign-in
│   └── internal/
│       ├── handlers/               # HTTP layer — thin, no business logic
│       ├── service/                # Validation + business logic
│       ├── repository/             # SQL only — no business logic
│       ├── models/                 # Domain types + scoring functions
│       └── middleware/
│           ├── auth.go             # ClerkAuthenticate: verifies RS256 JWT, upserts user
│           └── rate_limit.go       # Per-IP and per-user token-bucket limiters
└── frontend/
    └── src/
        ├── pages/
        │   ├── Dashboard.tsx       # Radar + bar chart + weakest category banner
        │   ├── ProblemList.tsx     # Paginated table with filter sidebar
        │   └── LogProblem.tsx      # Log form with LeetCode typeahead + multi-category picker
        ├── components/
        │   ├── CategoryRadarChart.tsx
        │   ├── CategoryBarChart.tsx
        │   ├── ProblemFilters.tsx
        │   └── AppSidebar.tsx
        ├── hooks/                  # Custom React hooks
        ├── types/api.ts            # TypeScript types mirroring API responses
        └── lib/api.ts              # Axios instance with Clerk JWT interceptor
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

`ComputeScore` is called **once at write time** and stored in the DB. `ApplyDecay` is called **at read time only** and is never persisted. This keeps historical scores stable while letting the displayed value reflect how much you've retained.

### Rate limiting

Two independent token-bucket limiters run in the middleware chain:

| Limiter | Scope | Burst |
|---|---|---|
| IP limiter | Unauthenticated and authenticated routes | Configurable |
| User limiter | Authenticated routes only (keyed on internal user ID) | Configurable |

Idle limiter entries are pruned in a background goroutine to prevent unbounded memory growth.

---

## License

MIT
