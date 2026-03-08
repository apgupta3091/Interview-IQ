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
export type LeetCodeProblemSuggestion = Schemas['handlers.leetCodeProblemSuggestion'];
