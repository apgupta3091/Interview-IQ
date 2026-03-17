-- problem_notes: standalone notes per user+problem_name, aggregated across all attempts.
-- Notes are keyed by (user_id, LOWER(TRIM(problem_name))) so all attempts for
-- the same problem share one note pool.
CREATE TABLE IF NOT EXISTS problem_notes (
    id           SERIAL PRIMARY KEY,
    user_id      INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_name TEXT        NOT NULL,  -- stored normalised: LOWER(TRIM(name))
    content      TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_problem_notes_user_name ON problem_notes(user_id, problem_name);

-- Migrate any existing per-attempt notes into the new table so no data is lost.
INSERT INTO problem_notes (user_id, problem_name, content, created_at, updated_at)
SELECT user_id, LOWER(TRIM(name)), notes, created_at, created_at
FROM problems
WHERE notes IS NOT NULL AND TRIM(notes) != ''
ON CONFLICT DO NOTHING;
