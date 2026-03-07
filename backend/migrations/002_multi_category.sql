-- Migrate problems from a single category TEXT column to categories TEXT[].
-- Each existing problem gets its original category as a one-element array.
ALTER TABLE problems ADD COLUMN categories TEXT[] NOT NULL DEFAULT '{}';
UPDATE problems SET categories = ARRAY[category];
ALTER TABLE problems DROP COLUMN category;
DROP INDEX IF EXISTS idx_problems_category;
CREATE INDEX idx_problems_categories ON problems USING GIN(categories);
