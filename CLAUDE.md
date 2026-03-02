# Interview-IQ Coding Standards

> **Purpose**: This document provides coding standards for the Interview-IQ platform. It is designed for AI LLMs, code reviews, and maintaining consistency across the codebase.

**Last Updated**: 2026-03-02
**Version**: 1.0.0

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Repository Structure](#repository-structure)
3. [Universal Standards](#universal-standards)
4. [Backend Standards](#backend-standards)
5. [Frontend Standards](#frontend-standards)
6. [API Design](#api-design)
7. [Code Review Checklist](#code-review-checklist)

---

## Project Overview

**Interview-IQ** is a LeetCode-style interview preparation tracker. Users log coding problems they've solved and the app:

- Scores each attempt based on number of attempts and whether the user peeked at the solution
- Applies time-based score decay at read time (simulating forgetting)
- Aggregates per-category strength scores across all solved problems
- Identifies the user's weakest category and recommends problems to focus on
- Visualizes skill levels across 17 categories via a radar/spider chart

### Technology Stack

| Layer | Technology |
|---|---|
| Backend language | Go 1.26 |
| HTTP router | Chi v5 |
| Database | PostgreSQL 16 |
| DB access | `database/sql` + `lib/pq` (no ORM) |
| Authentication | JWT (HS256, 7-day expiry) + bcrypt |
| Frontend language | TypeScript (strict) |
| Frontend framework | React 19 + Vite |
| UI components | ShadCN/UI |
| Charts | Recharts |
| HTTP client | Axios |
| Routing | React Router |

---

## Repository Structure

```
interview-iq/
├── CLAUDE.md                   # This file
├── README.md
├── docker-compose.yml          # Local PostgreSQL service
├── backend/
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   └── server/
│   │       └── main.go         # Entry point: env vars, DI wiring, router setup
│   ├── internal/
│   │   ├── auth/
│   │   │   └── jwt.go          # Token generation and parsing (HS256, 7-day)
│   │   ├── db/
│   │   │   ├── db.go           # Connection pool (max 25 open, 5 idle)
│   │   │   └── migrate.go      # SQL migration file runner
│   │   ├── handlers/
│   │   │   ├── auth.go         # Register + Login handlers
│   │   │   ├── problems.go     # ListProblems + LogProblem handlers
│   │   │   ├── categories.go   # GetStats + GetWeakest handlers
│   │   │   └── helpers.go      # writeJSON / writeError
│   │   ├── middleware/
│   │   │   └── auth.go         # JWT auth middleware + UserIDFromContext
│   │   ├── models/
│   │   │   ├── types.go        # Plain Go structs (Problem, CategoryStats, etc.)
│   │   │   ├── score.go        # ComputeScore + ApplyDecay business logic
│   │   │   └── score_test.go   # Unit tests for scoring and decay
│   │   ├── repository/
│   │   │   ├── errors.go       # ErrNotFound, ErrDuplicate sentinel errors
│   │   │   ├── user_repo.go    # UserRepository interface + sql implementation
│   │   │   ├── problem_repo.go # ProblemRepository interface + sql implementation
│   │   │   └── category_repo.go# CategoryRepository interface + sql implementation
│   │   └── service/
│   │       ├── errors.go       # ValidationError, ErrEmailTaken, ErrInvalidCredentials
│   │       ├── auth_service.go # AuthService interface + implementation
│   │       ├── problem_service.go
│   │       └── category_service.go
│   └── migrations/
│       └── 001_init.sql        # users + problems tables
└── frontend/                   # To be scaffolded (Vite + React)
    ├── src/
    │   ├── pages/              # Route-level components (Login, Register, Dashboard, etc.)
    │   ├── components/         # Reusable UI components
    │   │   └── ui/             # ShadCN auto-generated components (do not edit manually)
    │   ├── lib/
    │   │   └── api.ts          # Axios instance with JWT interceptor
    │   ├── hooks/              # Custom React hooks
    │   ├── types/              # Shared TypeScript types matching API responses
    │   └── main.tsx
    ├── index.html
    ├── vite.config.ts
    └── tsconfig.json
```

---

## Universal Standards

### Naming Conventions

#### Variables and Functions

- **Format**: `camelCase`
- **Examples**: `userID`, `solvedAt`, `lookedAtSolution`, `computeScore`

#### Go Exported Names

- **Format**: `PascalCase`
- **Examples**: `AuthService`, `ProblemRepository`, `ComputeScore`, `ApplyDecay`

#### Database Tables and Columns

- **Format**: `lower_case_with_underscores`
- **Examples**: `users`, `problems`, `user_id`, `looked_at_solution`, `time_taken_mins`
- **Never use**: capital letters or spaces in database identifiers

#### Constants and Enums

- **Format**: `SCREAMING_SNAKE_CASE` in TypeScript/JSON; Go uses `camelCase` unexported constants per idiomatic Go
- **Go examples**: `baseScore`, `attemptPenalty`, `decayGraceDays` (unexported package-level constants)
- **API/TS examples**: `AUTH_INVALID_TOKEN`, `PROBLEM_INVALID_CATEGORY`

#### URLs and Routes

- **Format**: `dash-separated-words`, plural resource names
- **Examples**: `/api/problems`, `/api/categories/stats`, `/api/categories/weakest`
- **Dynamic parameters**: `:camelCase` (Go Chi) / `:camelCase` (React Router)

### Comments and Documentation

- All non-trivial logic must have explanatory comments
- Required for: scoring/decay math, middleware context wiring, SQL queries with non-obvious filters
- The scoring and decay model constants (`score.go`) must always have inline comments explaining the rationale

### TODOs

```go
// TODO(#issue): Description of what needs to be done and why
```

- Must reference an associated GitHub issue number
- Include the reason, not just the task

### Example / Test Data

- Use placeholder emails like `alice@example.com`, `bob@example.com` in tests
- Never use real user data or real domain names you don't own

---

## Backend Standards

> **Stack**: Go, Chi v5, PostgreSQL, `database/sql`, `lib/pq`, JWT, bcrypt

### Architecture — 3-Layer Pattern

The backend follows a strict 3-layer architecture. **Never skip layers or mix concerns.**

```
Request
  └─> Handler       (decode input, call service, write response)
        └─> Service   (validation, business logic, orchestration)
              └─> Repository  (SQL queries, data access only)
```

Each layer communicates via **interfaces**, making layers independently testable.

```go
// Layer 1: Repository defines the interface AND the sql implementation
type ProblemRepository interface {
    Insert(ctx context.Context, p InsertProblemParams) (models.Problem, error)
    ListByUser(ctx context.Context, userID int) ([]models.Problem, error)
}

// Layer 2: Service accepts the interface, not the concrete type
type ProblemService interface {
    Log(ctx context.Context, userID int, input LogProblemInput) (models.Problem, error)
    List(ctx context.Context, userID int) ([]models.Problem, error)
}

// Layer 3: Handler holds the service interface
type ProblemHandler struct {
    Service service.ProblemService
}
```

### Handler Pattern

Handlers must be thin. Their only job: parse input → call service → write response.

```go
func (h *ProblemHandler) LogProblem(w http.ResponseWriter, r *http.Request) {
    // 1. Get authenticated user from context
    userID := middleware.UserIDFromContext(r.Context())

    // 2. Decode request body
    var req logProblemRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    // 3. Delegate all logic to service
    p, err := h.Service.Log(r.Context(), userID, service.LogProblemInput{...})
    if err != nil {
        // 4. Map domain errors to HTTP status codes
        var ve service.ValidationError
        if errors.As(err, &ve) {
            writeError(w, http.StatusBadRequest, ve.Message)
            return
        }
        writeError(w, http.StatusInternalServerError, "failed to log problem")
        return
    }

    // 5. Write response
    writeJSON(w, http.StatusCreated, toProblemResponse(p))
}
```

**Always use `writeJSON` / `writeError`** from `handlers/helpers.go`. Never call `json.NewEncoder` or `w.WriteHeader` directly in a handler.

### Service Pattern

Services contain all validation and business logic. Keep repositories free of business rules.

```go
// Define a sentinel error for each service error type
var ErrEmailTaken = errors.New("email already registered")

// Use a typed ValidationError for user-input errors
type ValidationError struct{ Message string }
func (e ValidationError) Error() string { return e.Message }

// Map repository errors to service errors — never let repository sentinel
// errors (ErrNotFound, ErrDuplicate) leak out of the service layer
func (s *authService) Register(ctx context.Context, email, password string) (string, string, error) {
    if email == "" || len(password) < 8 {
        return "", "", ValidationError{Message: "email required and password must be at least 8 characters"}
    }
    userID, err := s.users.Create(ctx, email, hash)
    if errors.Is(err, repository.ErrDuplicate) {
        return "", "", ErrEmailTaken  // translate, don't leak
    }
    ...
}
```

### Repository Pattern

Repositories contain only SQL. No validation, no business logic, no score computation.

- Always use `QueryRowContext` / `QueryContext` / `ExecContext` (never the non-context variants)
- Always `defer rows.Close()` immediately after a successful `QueryContext`
- Check `rows.Err()` after iterating
- Return `repository.ErrNotFound` when `sql.ErrNoRows` is encountered
- Return `repository.ErrDuplicate` when a unique-constraint violation occurs (check `pq.Error` code `23505`)

```go
func (r *sqlUserRepo) GetByEmail(ctx context.Context, email string) (int, string, error) {
    var id int
    var hash string
    err := r.db.QueryRowContext(ctx,
        `SELECT id, password_hash FROM users WHERE email = $1`, email,
    ).Scan(&id, &hash)
    if errors.Is(err, sql.ErrNoRows) {
        return 0, "", repository.ErrNotFound
    }
    if err != nil {
        return 0, "", err
    }
    return id, hash, nil
}
```

### Models

Plain Go structs with no ORM tags. Live in `internal/models/`.

- `types.go` — domain structs (`Problem`, `CategoryStats`, `WeakestResult`, `CategoryRawScore`)
- `score.go` — pure functions (`ComputeScore`, `ApplyDecay`); no DB access, no HTTP concerns

### Score and Decay Model

**Critical rule**: `ComputeScore` is called **once at write time** and stored in the DB. `ApplyDecay` is called **at read time only** and must never be persisted. This separation is intentional.

```go
// score.go constants — document any changes with a comment explaining the tradeoff
const (
    baseScore         = 100
    attemptPenalty    = 10   // -10 per extra attempt beyond the first
    maxAttemptPenalty = 40   // cap: minimum from attempts alone = 60
    solutionPenalty   = 25   // peeking at a solution is a significant penalty
    minScore          = 5    // logging any attempt is worth something

    decayGraceDays  = 3    // problems solved ≤3 days ago have no decay
    decayPerDay     = 2.0  // points lost per day after the grace period
    decayFloorRatio = 0.30 // you never fully forget: floor at 30% of original
)
```

### Error Handling

- Always wrap errors with context: `fmt.Errorf("users.Create: %w", err)`
- Never discard errors with `_` except for `defer rows.Close()` and body close
- Use `errors.As` for typed errors, `errors.Is` for sentinel errors
- Never let `*pq.Error` or `sql.ErrNoRows` escape the repository layer

### Authentication

- JWT: HS256, 7-day expiry, `user_id` claim. Generated in `internal/auth/jwt.go`
- Passwords: bcrypt at `bcrypt.DefaultCost`
- Middleware: `middleware.Authenticate` validates the `Authorization: Bearer <token>` header and injects `userID` into context via a typed `contextKey`
- Retrieve user in handlers with: `userID := middleware.UserIDFromContext(r.Context())`

### Environment Variables

| Variable | Default (dev only) | Notes |
|---|---|---|
| `PORT` | `8080` | HTTP listen port |
| `DATABASE_URL` | `postgres://interviewiq:interviewiq_secret@localhost:5432/interviewiq?sslmode=disable` | Full DSN |
| `JWT_SECRET` | `dev-secret-change-in-production` | Must be overridden in staging/prod |

- Always read from `os.Getenv` with a fallback only for local development
- Never commit real secrets; use environment-specific secrets in deployment

### Database Migrations

- Migration files live in `backend/migrations/` and are named `NNN_description.sql`
- Migrations run automatically on startup via `db.RunMigrations`
- Never edit a migration file that has already been applied to any environment
- New schema changes always get a new numbered migration file

### Logging

- Use Chi's built-in `chimiddleware.Logger` for request logging (already wired in `main.go`)
- For structured application logging in new code, use `log/slog` with context-aware methods:
  ```go
  slog.InfoContext(ctx, "problem logged", "user_id", userID, "problem_id", p.ID)
  slog.ErrorContext(ctx, "failed to insert problem", "user_id", userID, "error", err)
  ```
- Never use `log.Fatal` / `log.Fatalf` inside handlers or services — only in `main.go` startup

### Testing

- Use table-driven tests with a named `cases` slice of structs
- Use `t.Run(tc.name, ...)` for subtests
- Use only the standard `testing` package — no test framework dependencies
- Services should be tested with mock/stub repository implementations (define the interface, inject a fake)
- See `models/score_test.go` as the reference pattern:

```go
func TestComputeScore(t *testing.T) {
    cases := []struct {
        name             string
        attempts         int
        lookedAtSolution bool
        want             int
    }{
        {"perfect: 1 attempt, no solution", 1, false, 100},
        {"5 attempts, no solution (cap at -40)", 5, false, 60},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            got := ComputeScore(tc.attempts, tc.lookedAtSolution)
            if got != tc.want {
                t.Errorf("ComputeScore(%d, %v) = %d, want %d", tc.attempts, tc.lookedAtSolution, got, tc.want)
            }
        })
    }
}
```

### Development Commands

```bash
# Start local Postgres
docker compose up -d

# Run the backend (from backend/)
go run ./cmd/server

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./internal/models/...

# Check for common issues
go vet ./...
```

### Validated Categories (17 total)

The following are the only valid `category` values for problems:

```
array, string, hash-map, two-pointers, sliding-window, binary-search,
stack, queue, linked-list, tree, graph, heap, dp, backtracking,
greedy, math, other
```

Category validation lives in the service layer, not the handler or repository.

---

## Frontend Standards

> **Stack**: React 19, Vite, TypeScript (strict), ShadCN/UI, Recharts, Axios, React Router

### Project Setup Conventions

- Path alias: `@/` maps to `src/` (configured in `vite.config.ts` and `tsconfig.json`)
- All ShadCN components live in `src/components/ui/` — **do not manually edit these files**
- Custom components live in `src/components/` (one component per file, PascalCase filename)
- Pages (route-level components) live in `src/pages/`
- Custom hooks live in `src/hooks/` (e.g., `useAuth`, `useProblems`, `useCategoryStats`)
- Shared TypeScript types live in `src/types/` and must mirror the API response shapes

### TypeScript

- Strict mode must be enabled (`"strict": true` in `tsconfig.json`)
- Always use `type` imports for type-only usage:
  ```typescript
  import type { Problem, CategoryStats } from '@/types/api';
  ```
- Never use `any`. Use `unknown` and narrow with type guards when the shape is genuinely unknown
- Define response types in `src/types/api.ts` that match the backend JSON exactly:
  ```typescript
  export type Problem = {
    id: number;
    name: string;
    category: string;
    difficulty: 'easy' | 'medium' | 'hard';
    attempts: number;
    looked_at_solution: boolean;
    time_taken_mins: number;
    score: number;
    decayed_score: number;
    solved_at: string;       // ISO 8601
    created_at: string;
  };
  ```

### ShadCN/UI

- Always prefer ShadCN components over raw HTML elements for UI
- Install new components with: `npx shadcn@latest add <component>`
- Do not edit files inside `src/components/ui/` — regenerate them instead
- Compose ShadCN primitives to build feature-specific components in `src/components/`

### Axios API Client

The Axios instance with JWT interceptor lives in `src/lib/api.ts`:

```typescript
import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? 'http://localhost:8080',
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export default api;
```

- All API calls must go through this instance — never use `fetch` directly
- Store the JWT in `localStorage` under the key `token`
- On 401 responses, clear the token and redirect to `/login`

### React Patterns

- Use functional components with hooks exclusively — no class components
- Colocate component state with the component that owns it
- Lift state up only when genuinely shared between siblings
- Keep components under 200 lines; extract sub-components or hooks when larger
- No `console.log` in committed code

### Radar Chart (Recharts)

The per-category radar chart is the core visualization. Use Recharts `RadarChart`:

```typescript
import { RadarChart, PolarGrid, PolarAngleAxis, Radar, ResponsiveContainer } from 'recharts';

// Data shape expected by the chart
type RadarDatum = {
  category: string;
  strength: number; // 0–100, the decayed average score for that category
};
```

- Normalize `strength` to a 0–100 scale before passing to the chart
- Show all 17 categories on the chart even if `strength` is 0 (makes gaps visible)
- Use `ResponsiveContainer` so the chart is fluid

### React Router

- Define all routes in `src/main.tsx` (or a dedicated `src/routes.tsx`)
- Use `<Navigate>` for redirects, never `window.location.href`
- Protect authenticated routes with a wrapper component that checks `localStorage` for a token

### Environment Variables

| Variable | Purpose |
|---|---|
| `VITE_API_URL` | Base URL for the Go backend (default: `http://localhost:8080`) |

- All Vite env vars must start with `VITE_`
- Never commit `.env` files with real values; commit `.env.example` with placeholder values

---

## API Design

### Route Table

| Method | Path | Auth | Handler |
|---|---|---|---|
| `GET` | `/health` | None | Health check |
| `POST` | `/api/auth/register` | None | `AuthHandler.Register` |
| `POST` | `/api/auth/login` | None | `AuthHandler.Login` |
| `GET` | `/api/problems` | JWT | `ProblemHandler.ListProblems` |
| `POST` | `/api/problems` | JWT | `ProblemHandler.LogProblem` |
| `GET` | `/api/categories/stats` | JWT | `CategoryHandler.GetStats` |
| `GET` | `/api/categories/weakest` | JWT | `CategoryHandler.GetWeakest` |

### Standard Error Response

All errors use this shape:

```json
{
  "error": "human-readable message"
}
```

When adding new endpoints, maintain this shape. Do not add fields like `code` or `details` unless the frontend specifically needs to branch on them.

### Error Domain Prefixes (for future machine-readable codes)

If the API ever expands to return structured error codes, prefix by domain:

- `AUTH_` — registration, login, token issues
- `PROBLEM_` — invalid category, difficulty, missing fields
- `CATEGORY_` — stats, weakest category errors

### Pagination

- List endpoints default to returning all records for a single user (volume is low per user)
- If any endpoint grows beyond ~500 records per user, add `limit`/`offset` query params

---

## Code Review Checklist

### Universal

- [ ] Variables `camelCase`, DB columns `lower_case_with_underscores`, constants follow Go/TS conventions
- [ ] URL routes use `dash-separated-words` plural nouns
- [ ] Non-trivial logic has explanatory comments
- [ ] TODOs reference a GitHub issue number
- [ ] No secrets or credentials in code

### Backend (Go)

- [ ] **Architecture layers respected**: handlers call services, services call repositories — no skipping
- [ ] **Handler is thin**: only decodes input, calls service, maps errors, writes response
- [ ] **Service owns validation**: all input validation in service, not handler or repository
- [ ] **Repository is pure SQL**: no business logic, no score computation
- [ ] **Error wrapping**: all errors wrapped with `fmt.Errorf("context: %w", err)`
- [ ] **Repository errors translated**: `ErrNotFound` / `ErrDuplicate` never escape to handlers
- [ ] **Context passed**: all DB calls use `*Context` variants
- [ ] **`rows.Close()` deferred**: immediately after successful `QueryContext`
- [ ] **`rows.Err()` checked**: after row iteration loop
- [ ] **Score model not persisted on read**: `ApplyDecay` is never saved to the DB
- [ ] **Table-driven tests**: new tests follow the `[]struct{ name, ... }` + `t.Run` pattern
- [ ] **No `log.Fatal` outside `main.go`**
- [ ] **New migration file**: schema changes have a new `NNN_description.sql`

### Frontend (React / TypeScript)

- [ ] **No `any`**: strict TypeScript used throughout
- [ ] **`type` imports**: type-only imports use `import type`
- [ ] **API types defined**: response types in `src/types/api.ts` mirror backend JSON
- [ ] **ShadCN preferred**: raw HTML not used where a ShadCN component exists
- [ ] **`src/components/ui/` not hand-edited**: ShadCN files regenerated, not modified
- [ ] **Axios instance used**: no bare `fetch` calls
- [ ] **JWT interceptor wired**: all authenticated requests go through `src/lib/api.ts`
- [ ] **No `console.log`**: removed before committing
- [ ] **Components ≤200 lines**: larger components split into sub-components or hooks
- [ ] **Radar chart shows all 17 categories**: even those with strength 0

### Security

- [ ] JWT validated on all protected routes
- [ ] Passwords hashed with bcrypt — plaintext never stored or logged
- [ ] User can only access their own problems (filter by `userID` from JWT, not from request body)
- [ ] Input validated before any DB operation
- [ ] No sensitive data in error messages returned to client

### Performance

- [ ] No N+1 queries: `ListByUser` fetches all problems in one query, not one per problem
- [ ] `ApplyDecay` computed in Go at read time, not in SQL
- [ ] DB indexes exist on `problems(user_id)` and `problems(category)` (see `001_init.sql`)
- [ ] React renders don't cause unnecessary re-fetches (memoize or use proper hook dependencies)

---

## Additional Resources

### Go

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Chi Documentation](https://github.com/go-chi/chi)
- [lib/pq Error Codes](https://pkg.go.dev/github.com/lib/pq#hdr-Error_Handling)

### Frontend

- [ShadCN/UI Documentation](https://ui.shadcn.com)
- [Recharts Documentation](https://recharts.org)
- [React Router Documentation](https://reactrouter.com)
- [Vite Documentation](https://vite.dev)

---

**For questions about these standards, refer to the architecture diagrams in `.cursor/plans/` or open a GitHub discussion.**
