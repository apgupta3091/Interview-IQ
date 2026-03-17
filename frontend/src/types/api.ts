/**
 * Ergonomic type aliases extracted from the auto-generated schema.
 * Run `pnpm gen:types` to regenerate schema.ts from the backend Swagger spec.
 * Do not add hand-written types here — add Swagger annotations in Go instead.
 */
import type { components } from './schema';

type Schemas = components['schemas'];

export type Problem                   = Schemas['handlers.problemResponse'];
export type LogProblemRequest         = Schemas['handlers.logProblemRequest'];
export type ProblemListResponse       = Schemas['handlers.listProblemsResponse'];
export type CategoryStats             = Schemas['handlers.categoryStatsResponse'];
export type WeakestResult             = Schemas['handlers.weakestResponse'];
export type ApiError                  = Schemas['handlers.errorResponse'];
// Manually defined — leetCodeProblemSuggestion is unexported in Go and not in the generated schema
export type LeetCodeProblemSuggestion = {
  lc_id: number;
  title: string;
  slug: string;
  difficulty: string;
  tags: string[];
};

// --- Problem Notes (hand-written — separate from per-attempt notes) ---

export type ProblemNote = {
  id: number;
  problem_name: string;
  content: string;
  created_at: string;  // ISO 8601
  updated_at: string;  // ISO 8601
};

export type NoteListResponse = {
  notes: ProblemNote[];
};

// --- Recommendations (hand-written until Swagger annotations are added) ---

export type ProblemRec = {
  name: string;
  difficulty: string;
  description: string;
  reason: string;
};

export type CategoryRec = {
  category: string;
  strength: number;
  focus_note: string;
  problems: ProblemRec[];
};

export type RecommendationResult = {
  categories: CategoryRec[];
};

export type RecommendationParams = {
  category?: string;
  limit?: number;
};
