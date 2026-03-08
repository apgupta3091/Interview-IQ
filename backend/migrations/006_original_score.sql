ALTER TABLE problems ADD COLUMN original_score INTEGER NOT NULL DEFAULT 0;

-- Backfill: the existing score column held the raw value before this migration.
UPDATE problems SET original_score = score;

-- Recompute score = ApplyDecay(original_score, solved_at) for all existing rows.
-- Constants match score.go: 7-day grace, 1.0 pt/day decay, 0.40 floor.
UPDATE problems
SET score = GREATEST(
    CEIL(original_score * 0.40),
    ROUND(
        original_score::numeric
        - GREATEST(
            EXTRACT(EPOCH FROM (NOW() - solved_at)) / 86400.0 - 7,
            0
          ) * 1.0
    )
)::integer;
