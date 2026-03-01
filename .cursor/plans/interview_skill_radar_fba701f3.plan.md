---
name: Interview Skill Radar
overview: "Build a full-stack Interview Skill Radar app using Go (Chi), PostgreSQL, React, and ShadCN. Features: problem logging, score engine, category strength aggregation, time-decay, weakest category detection, and a radar chart visualization. Auth via JWT."
todos:
  - id: step-1
    content: Docker Compose for PostgreSQL (docker-compose.yml)
    status: completed
  - id: step-2
    content: Go module init, folder structure, Chi router, /health endpoint
    status: pending
  - id: step-3
    content: DB connection pool + migrations/001_init.sql (users + problems tables)
    status: pending
  - id: step-4
    content: "Auth handlers: POST /api/auth/register and POST /api/auth/login with bcrypt + JWT"
    status: pending
  - id: step-5
    content: JWT middleware protecting /api/* routes
    status: pending
  - id: step-6
    content: Score engine (score.go) + decay formula helper
    status: pending
  - id: step-7
    content: POST /api/problems — log problem, compute + store score
    status: pending
  - id: step-8
    content: GET /api/problems — list problems with live decayed score
    status: pending
  - id: step-9
    content: GET /api/categories/stats — per-category strength %
    status: pending
  - id: step-10
    content: GET /api/categories/weakest — weakest category + static recommendations
    status: pending
  - id: step-11
    content: "Frontend: Vite + React + ShadCN + Recharts + React Router init"
    status: pending
  - id: step-12
    content: Axios client with JWT interceptor + auth context
    status: pending
  - id: step-13
    content: Login + Register pages
    status: pending
  - id: step-14
    content: Log Problem page (form with 6 fields)
    status: pending
  - id: step-15
    content: Problems list page (table with decayed scores)
    status: pending
  - id: step-16
    content: "Dashboard page: category strength cards + weakest category banner"
    status: pending
  - id: step-17
    content: Recharts RadarChart on Dashboard (one axis per category)
    status: pending
  - id: step-18
    content: ShadCN sidebar nav layout + loading states + error toasts + empty states
    status: pending
isProject: false
---

# Interview Skill Radar — Build Plan

## Stack

- **Backend:** Go 1.22+, Chi router, `database/sql` + `lib/pq` (raw SQL, no ORM)
- **Frontend:** React + Vite, ShadCN/UI, Recharts (radar chart), Axios
- **Database:** PostgreSQL (via Docker Compose for local dev)
- **Auth:** bcrypt password hashing + JWT (golang-jwt/jwt)

## Project Structure

```
interview-iq/
  backend/
    cmd/server/main.go
    internal/
      db/         # DB pool init
      auth/       # JWT helpers, bcrypt
      handlers/   # HTTP handlers
      middleware/ # JWT auth middleware
      models/     # Go structs + score formula
    migrations/   # Raw .sql files
    go.mod
  frontend/
    src/
      pages/      # Login, Register, Dashboard, LogProblem
      components/ # RadarChart, CategoryCard, ProblemTable
      lib/        # axios client, auth helpers
    package.json
  docker-compose.yml
```

## Score Formula

```
score = 100
score -= (attempts - 1) * 10   // -10 per extra attempt, e.g. 3 attempts = -20
if looked_at_solution: score -= 25
score = max(score, 5)          // floor at 5
```

## Decay Formula (applied at read time)

```
decayed_score = score * (0.9 ^ (days_since_solved / 7))
// ~10% decay per week of not revisiting
```

## Category Strength

```
strength% = avg(decayed_score) across all problems in that category
```

## DB Schema (2 tables)

- `users` — id, email, password_hash, created_at
- `problems` — id, user_id, name, category, difficulty, attempts, looked_at_solution, time_taken_mins, score, solved_at, created_at

## Static Recommendation Map

Hardcoded Go map: category → 3 problem names. Shown when that category is weakest. No AI needed.

---

## Steps

### Phase 1: Scaffolding & Database

- **Step 1** — Docker Compose for PostgreSQL + `docker-compose.yml`
- **Step 2** — Go module init, folder structure, Chi router wired, `/health` endpoint
- **Step 3** — DB connection pool (`internal/db`), migration files (`migrations/001_init.sql`)

### Phase 2: Auth Backend

- **Step 4** — `users` table handlers: `POST /api/auth/register`, `POST /api/auth/login` (bcrypt + JWT response)
- **Step 5** — JWT middleware (`internal/middleware/auth.go`) protecting all `/api/` routes except auth

### Phase 3: Problem API

- **Step 6** — Score engine function (`internal/models/score.go`) + decay helper
- **Step 7** — `POST /api/problems` — log a problem, compute + store score
- **Step 8** — `GET /api/problems` — list all problems for authed user (with live decayed score)
- **Step 9** — `GET /api/categories/stats` — return per-category strength % (avg decayed score)
- **Step 10** — `GET /api/categories/weakest` — return weakest category + 3 static problem recommendations

### Phase 4: Frontend Scaffolding

- **Step 11** — Vite + React project init, install ShadCN, Recharts, Axios, React Router
- **Step 12** — Axios client with JWT interceptor (`src/lib/api.ts`), auth context/store

### Phase 5: Auth Pages

- **Step 13** — Login page + Register page using ShadCN form components, store JWT in localStorage

### Phase 6: Core App Pages

- **Step 14** — Log Problem page: form with all 6 fields (name, category, difficulty, attempts, looked_at_solution, time_taken), POST to backend
- **Step 15** — Problems list page: table of logged problems with decayed score shown per row
- **Step 16** — Dashboard page: category strength cards (% per category), weakest category banner with 3 recommendations

### Phase 7: Radar Chart

- **Step 17** — Skill Radar chart using Recharts `RadarChart` — one axis per category, value = strength %; sits on Dashboard

### Phase 8: Polish

- **Step 18** — ShadCN sidebar nav layout wrapping all pages, loading states, error toasts, empty states
