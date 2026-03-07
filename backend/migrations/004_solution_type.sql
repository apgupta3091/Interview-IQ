-- Add solution_type to record whether the user reached an optimal or brute-force solution.
-- 'none'        — default; user did not categorise their approach
-- 'brute_force' — working solution but not optimal; carries a score penalty
-- 'optimal'     — fully optimal solution; no additional penalty
ALTER TABLE problems
    ADD COLUMN IF NOT EXISTS solution_type TEXT NOT NULL DEFAULT 'none'
        CHECK (solution_type IN ('none', 'brute_force', 'optimal'));
