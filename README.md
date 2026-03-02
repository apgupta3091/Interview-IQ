# Interview-IQ

A LeetCode-style interview prep tracker that scores your problem-solving sessions and uses time-based decay to surface what you need to review — not what you already know cold.

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql&logoColor=white)
![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=black)
![TypeScript](https://img.shields.io/badge/TypeScript-strict-3178C6?logo=typescript&logoColor=white)

---

## What It Does

When you solve a LeetCode problem, you log it: how many attempts it took and whether you peeked at the solution. Interview-IQ assigns a score (0–100), then decays that score over time as you'd naturally start forgetting the pattern.

Across all your logged problems, the app aggregates strength scores per category (arrays, graphs, dynamic programming, etc.) and visualizes them as a radar chart. The weakest category is surfaced with targeted problem recommendations so you know exactly where to focus next.

**Core loop:**
1. Solve a problem on LeetCode
2. Log it in Interview-IQ (attempts, category, difficulty, did you peek?)
3. Score is computed and stored
4. Over days/weeks, your scores decay — the radar chart shifts to reflect what you've retained
5. The app tells you your weakest category and suggests what to practice

---

## Scoring Model

### Raw Score (computed once at log time)

| Condition | Effect |
|---|---|
| Base score | 100 |
| Each extra attempt beyond the first | −10 (capped at −40) |
| Looked at the solution | −25 |
| Absolute minimum | 5 |

Examples:
- 1 attempt, no peek → **100**
- 3 attempts, no peek → **80**
- 5+ attempts, no peek → **60** (penalty capped)
- 1 attempt, peeked → **75**
- 5+ attempts, peeked → **35**

### Decay (applied at read time, never stored)

The raw score decays linearly after a 3-day grace period, flooring at 30% of the original.

| Days since solving | Effect on a 100-point score |
|---|---|
| 0–3 | 100 (no decay) |
| 10 | 86 |
| 21 | 64 |
| 30+ | 30 (floor) |

This simulates the forgetting curve — a problem you solved three weeks ago is worth less than one you solved yesterday, even if you nailed it both times.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.26 |
| HTTP router | Chi v5 |
| Database | PostgreSQL 16 |
| DB access | `database/sql` + `lib/pq` (no ORM) |
| Auth | JWT (HS256, 7-day expiry) + bcrypt |
| API docs | Swagger (auto-generated from annotations) |
| Frontend | React 19 + Vite + TypeScript (strict) |
| UI components | ShadCN/UI |
| Charts | Recharts (radar chart) |
| HTTP client | Axios |
| Routing | React Router |

---

## API Reference

### Authentication

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/api/auth/register` | None | Register a new account |
| `POST` | `/api/auth/login` | None | Log in, receive JWT |

### Problems

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/problems` | JWT | List all logged problems with decayed scores |
| `POST` | `/api/problems` | JWT | Log a new problem attempt |

### Categories

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/categories/stats` | JWT | Per-category strength scores (0–100) |
| `GET` | `/api/categories/weakest` | JWT | Weakest category + 3 recommended problems |

### Health

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/health` | None | Health check |

Interactive docs are available at `/swagger/index.html` when the server is running.

### Example: Log a Problem

```http
POST /api/problems
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Two Sum",
  "category": "hash-map",
  "difficulty": "easy",
  "attempts": 1,
  "looked_at_solution": false,
  "time_taken_mins": 12
}
```

```json
{
  "id": 42,
  "name": "Two Sum",
  "category": "hash-map",
  "difficulty": "easy",
  "attempts": 1,
  "looked_at_solution": false,
  "time_taken_mins": 12,
  "score": 100,
  "decayed_score": 100,
  "solved_at": "2026-03-02T14:30:00Z",
  "created_at": "2026-03-02T14:30:00Z"
}
```

### Valid Categories

```
array, string, hash-map, two-pointers, sliding-window, binary-search,
stack, queue, linked-list, tree, graph, heap, dp, backtracking,
greedy, math, other
```

---

## Getting Started

### Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Docker](https://www.docker.com/) (for local PostgreSQL)

### Setup

```bash
# Clone the repo
git clone https://github.com/apgupta3091/interview-iq.git
cd interview-iq

# Start the database
make db-up

# Run the backend (auto-runs migrations on first start)
make backend
```

The API will be available at `http://localhost:8080`.

### Environment Variables

Create a `.env` file in `backend/` (or export these in your shell):

```bash
PORT=8080
DATABASE_URL=postgres://interviewiq:interviewiq_secret@localhost:5432/interviewiq?sslmode=disable
JWT_SECRET=change-this-in-production
```

The defaults above work out of the box with `make db-up`. Override `JWT_SECRET` in any non-local environment.

---

## Development

```bash
make dev        # Start DB + run server together
make test       # Run all tests
make test-v     # Run tests with verbose output
make lint       # go vet all packages
make tidy       # Tidy go.mod/go.sum
make db-reset   # Wipe DB volume and restart fresh
```

---

## Project Structure

```
interview-iq/
├── backend/
│   ├── cmd/server/main.go          # Entry point: env, DI wiring, router setup
│   ├── migrations/001_init.sql     # users + problems schema
│   └── internal/
│       ├── handlers/               # HTTP layer — thin, no business logic
│       ├── service/                # Validation + business logic
│       ├── repository/             # SQL only — no business logic
│       ├── models/                 # Domain types + scoring functions
│       ├── middleware/auth.go      # JWT validation, injects userID into context
│       └── auth/jwt.go             # Token generation and parsing
└── frontend/                       # React 19 + Vite (in progress)
    └── src/
        ├── pages/                  # Route-level components
        ├── components/             # Reusable UI components
        ├── hooks/                  # useAuth, useProblems, useCategoryStats
        ├── types/api.ts            # TypeScript types mirroring API responses
        └── lib/api.ts              # Axios instance with JWT interceptor
```

The backend follows a strict 3-layer architecture: handlers call services, services call repositories. Layers communicate via interfaces and are independently testable.

---

## License

MIT
