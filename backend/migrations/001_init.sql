CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS problems (
    id                  SERIAL PRIMARY KEY,
    user_id             INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name                TEXT NOT NULL,
    category            TEXT NOT NULL,
    difficulty          TEXT NOT NULL CHECK (difficulty IN ('easy', 'medium', 'hard')),
    attempts            INTEGER NOT NULL DEFAULT 1,
    looked_at_solution  BOOLEAN NOT NULL DEFAULT FALSE,
    time_taken_mins     INTEGER NOT NULL DEFAULT 0,
    score               INTEGER NOT NULL DEFAULT 100,
    solved_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_problems_user_id ON problems(user_id);
CREATE INDEX IF NOT EXISTS idx_problems_category ON problems(category);
