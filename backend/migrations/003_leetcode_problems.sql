-- Catalog of LeetCode problems seeded from the public LeetCode API.
-- Used to power the typeahead search in the Log Problem form.
CREATE TABLE IF NOT EXISTS leetcode_problems (
    id        SERIAL PRIMARY KEY,
    lc_id     INTEGER NOT NULL UNIQUE,
    title     TEXT NOT NULL,
    slug      TEXT NOT NULL UNIQUE,
    difficulty TEXT NOT NULL,
    tags      TEXT[] NOT NULL DEFAULT '{}',
    paid_only BOOLEAN NOT NULL DEFAULT FALSE
);

-- Full-text search index on title for the search endpoint.
CREATE INDEX IF NOT EXISTS idx_lc_problems_search
    ON leetcode_problems
    USING GIN(to_tsvector('english', title));
