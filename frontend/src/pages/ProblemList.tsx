import { useEffect, useRef, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import axios from 'axios'
import { Plus, FileX, ChevronLeft, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { api } from '@/lib/api'
import type { Problem, ApiError, ProblemListResponse } from '@/types/api'
import ProblemFilters, {
  SCORE_RANGE_OPTIONS,
} from '@/components/ProblemFilters'
import type { DateRangeValue, ScoreRangeValue } from '@/components/ProblemFilters'

const PAGE_SIZE = 20

const DIFFICULTY_STYLES: Record<string, string> = {
  easy:   'bg-emerald-500/10 text-emerald-500 border border-emerald-500/20',
  medium: 'bg-amber-500/10  text-amber-500  border border-amber-500/20',
  hard:   'bg-red-500/10    text-red-500    border border-red-500/20',
}

function scoreColor(score: number) {
  if (score >= 70) return 'text-emerald-500'
  if (score >= 40) return 'text-amber-500'
  return 'text-red-500'
}

function timeSince(date: Date): string {
  const days = Math.floor((Date.now() - date.getTime()) / 86_400_000)
  if (days === 0) return 'today'
  if (days === 1) return '1 day ago'
  if (days < 30) return `${days} days ago`
  const months = Math.floor(days / 30)
  if (months === 1) return '1 month ago'
  if (months < 12) return `${months} months ago`
  const years = Math.floor(months / 12)
  return years === 1 ? '1 year ago' : `${years} years ago`
}

// Within the current page (sorted newest-first), mark the first occurrence of
// each problem name as "Latest" when duplicates exist on this page.
function buildAttemptMeta(problems: Problem[]) {
  const nameCounts: Record<string, number> = {}
  for (const p of problems) nameCounts[p.name ?? ''] = (nameCounts[p.name ?? ''] ?? 0) + 1
  const seen = new Set<string>()
  const latestIds = new Set<number>()
  for (const p of problems) {
    const name = p.name ?? ''
    if (!seen.has(name)) { seen.add(name); if (p.id !== undefined) latestIds.add(p.id) }
  }
  return { nameCounts, latestIds }
}

// ---------------------------------------------------------------------------
// Filter state
// ---------------------------------------------------------------------------

type FilterState = {
  nameSearch: string
  categories: string[]
  difficulties: string[]
  scoreRange: ScoreRangeValue
  dateRange: DateRangeValue
}

const EMPTY_FILTERS: FilterState = {
  nameSearch: '', categories: [], difficulties: [],
  scoreRange: '', dateRange: '',
}

function hasAnyFilter(f: FilterState) {
  return !!(f.nameSearch || f.categories.length || f.difficulties.length ||
            f.scoreRange || f.dateRange)
}

/** Compute the ISO date string for the start of a date range preset. */
function dateRangeToFrom(range: DateRangeValue): string | undefined {
  if (!range) return undefined
  const now = new Date()
  const days: Record<string, number> = {
    day: 1, week: 7, '2weeks': 14, month: 30, '3months': 90,
  }
  const d = days[range]
  if (!d) return undefined
  const from = new Date(now)
  from.setDate(from.getDate() - d)
  return from.toISOString().slice(0, 10)
}

// ---------------------------------------------------------------------------

export default function ProblemList() {
  const navigate = useNavigate()
  // draft  — what the filter controls show (not yet applied to the query)
  const [draft, setDraft]     = useState<FilterState>(EMPTY_FILTERS)
  // applied — committed filters; the fetch effect depends on this value
  const [applied, setApplied] = useState<FilterState>(EMPTY_FILTERS)
  const [offset, setOffset]   = useState(0)

  const [result, setResult]           = useState<ProblemListResponse>({ problems: [], total: 0, limit: PAGE_SIZE, offset: 0 })
  const [loading, setLoading]         = useState(true)
  const [isInitialLoad, setIsInitialLoad] = useState(true)
  // Monotonically-increasing counter to discard stale responses.
  const fetchIdRef = useRef(0)

  // Fetch whenever applied filters or page offset change.
  useEffect(() => {
    const id = ++fetchIdRef.current
    setLoading(true)
    const scoreOpt = applied.scoreRange
      ? SCORE_RANGE_OPTIONS.find((o) => o.value === applied.scoreRange)
      : undefined
    api.problems.listFiltered({
      q:          applied.nameSearch || undefined,
      category:   applied.categories.length  ? applied.categories  : undefined,
      difficulty: applied.difficulties.length ? applied.difficulties : undefined,
      score_min:  scoreOpt ? scoreOpt.min : undefined,
      score_max:  scoreOpt ? scoreOpt.max : undefined,
      from:       dateRangeToFrom(applied.dateRange),
      to:         undefined,
      limit:      PAGE_SIZE,
      offset,
    })
      .then((data) => {
        if (id !== fetchIdRef.current) return
        setResult(data)
        setIsInitialLoad(false)
      })
      .catch((err) => {
        if (id !== fetchIdRef.current) return
        if (axios.isAxiosError(err)) {
          const msg = (err.response?.data as ApiError)?.error ?? 'Failed to load problems'
          toast.error(msg)
        } else {
          toast.error('Unexpected error')
        }
      })
      .finally(() => {
        if (id !== fetchIdRef.current) return
        setLoading(false)
      })
  }, [applied, offset])

  // Instant-apply handlers: update both draft and applied immediately.
  const handleNameSearch = (v: string) => {
    setDraft((d) => ({ ...d, nameSearch: v }))
    setApplied((a) => ({ ...a, nameSearch: v }))
    setOffset(0)
  }
  const handleCategoriesChange = (v: string[]) => {
    setDraft((d) => ({ ...d, categories: v }))
    setApplied((a) => ({ ...a, categories: v }))
    setOffset(0)
  }
  const handleDifficultiesChange = (v: string[]) => {
    setDraft((d) => ({ ...d, difficulties: v }))
    setApplied((a) => ({ ...a, difficulties: v }))
    setOffset(0)
  }

  const handleDateRangeChange = (v: DateRangeValue) => {
    setDraft((d) => ({ ...d, dateRange: v }))
    setApplied((a) => ({ ...a, dateRange: v }))
    setOffset(0)
  }

  const handleScoreRangeChange = (v: ScoreRangeValue) => {
    setDraft((d) => ({ ...d, scoreRange: v }))
    setApplied((a) => ({ ...a, scoreRange: v }))
    setOffset(0)
  }

  // Clear resets both draft and applied (new object reference triggers a refetch).
  const clearFilters = () => {
    setDraft(EMPTY_FILTERS)
    setApplied({ ...EMPTY_FILTERS })
    setOffset(0)
  }

  const hasApplied       = hasAnyFilter(applied)
  const hasDraftOrActive = hasAnyFilter(draft) || hasApplied

  const problems    = result.problems ?? []
  const total       = result.total    ?? 0
  const totalPages  = Math.ceil(total / PAGE_SIZE)
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1
  const { nameCounts, latestIds } = buildAttemptMeta(problems)

  return (
    <div className="space-y-5 animate-fade-up">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Problems</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {loading ? '…' : `${total} problem${total !== 1 ? 's' : ''}${hasApplied ? ' matched' : ' logged'}`}
          </p>
        </div>
        <Button asChild size="sm">
          <Link to="/problems/new">
            <Plus className="w-4 h-4 mr-1" />
            Log problem
          </Link>
        </Button>
      </div>

      {/* Filters */}
      <ProblemFilters
        nameSearch={draft.nameSearch}             onNameSearch={handleNameSearch}
        dateRange={applied.dateRange}             onDateRangeChange={handleDateRangeChange}
        selectedCategories={draft.categories}     onCategoriesChange={handleCategoriesChange}
        selectedDifficulties={draft.difficulties} onDifficultiesChange={handleDifficultiesChange}
        scoreRange={applied.scoreRange}           onScoreRangeChange={handleScoreRangeChange}
        hasFilters={hasDraftOrActive}
        onClear={clearFilters}
      />

      {/* Indeterminate loading bar — visible on subsequent fetches only */}
      <div className={`relative h-0.5 rounded-full overflow-hidden bg-border transition-opacity duration-300 ${loading && !isInitialLoad ? 'opacity-100' : 'opacity-0'}`}>
        <div className="absolute inset-y-0 left-0 w-1/4 bg-primary rounded-full animate-progress-slide" />
      </div>

      {/* Skeleton — initial load only */}
      {isInitialLoad && loading ? (
        <div className="space-y-3 animate-pulse">
          {[...Array(5)].map((_, i) => <div key={i} className="h-12 bg-muted rounded" />)}
        </div>
      ) : problems.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-32 text-center">
          <div className="w-12 h-12 rounded-xl bg-muted flex items-center justify-center mb-4">
            <FileX className="w-6 h-6 text-muted-foreground" />
          </div>
          <p className="text-base font-semibold mb-1">
            {hasApplied ? 'No problems match your filters' : 'No problems yet'}
          </p>
          <p className="text-sm text-muted-foreground">
            {hasApplied ? (
              <button onClick={clearFilters} className="text-primary hover:underline underline-offset-4">
                Clear filters
              </button>
            ) : (
              <Link to="/problems/new" className="text-primary hover:underline underline-offset-4">
                Log your first problem
              </Link>
            )}{' '}
            {!hasApplied && 'to get started.'}
          </p>
        </div>
      ) : (
        <div className={`space-y-4 transition-opacity duration-200 ${loading ? 'opacity-50 pointer-events-none select-none' : 'opacity-100'}`}>

          {/* ── Mobile list (no table, no min-width) ── */}
          <div className="sm:hidden rounded-lg border border-border/60 divide-y divide-border/60">
            {problems.map((p) => (
              <div
                key={p.id}
                className="flex items-center justify-between px-4 py-3 hover:bg-muted/20 cursor-pointer"
                onClick={() => navigate(`/problems/${p.id}`)}
              >
                <div className="min-w-0 mr-3">
                  <p className="text-sm font-medium truncate">{p.name}</p>
                  <div className="flex items-center gap-1.5 mt-0.5">
                    <span className={`text-xs px-1.5 py-0.5 rounded-md font-medium ${DIFFICULTY_STYLES[p.difficulty ?? ''] ?? ''}`}>
                      {p.difficulty}
                    </span>
                    {nameCounts[p.name ?? ''] > 1 && latestIds.has(p.id ?? -1) && (
                      <span className="text-xs font-medium text-emerald-500 bg-emerald-500/10 border border-emerald-500/20 px-1.5 py-0.5 rounded">
                        Latest
                      </span>
                    )}
                  </div>
                </div>
                <span className={`font-mono text-sm font-semibold shrink-0 ${scoreColor(p.score ?? 0)}`}>
                  {p.score}
                </span>
              </div>
            ))}
          </div>

          {/* ── Desktop table ── */}
          <div className="hidden sm:block rounded-lg border border-border/60 overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow className="bg-muted/30 hover:bg-muted/30">
                  <TableHead className="font-medium">Problem</TableHead>
                  <TableHead className="font-medium">Category</TableHead>
                  <TableHead className="font-medium">Difficulty</TableHead>
                  <TableHead className="text-center font-medium">Attempts</TableHead>
                  <TableHead className="text-center font-medium">Peeked</TableHead>
                  <TableHead className="text-center font-medium">Solution</TableHead>
                  <TableHead className="text-right font-medium">Score</TableHead>
                  <TableHead className="text-right font-medium">Solved</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {problems.map((p) => (
                  <TableRow
                    key={p.id}
                    className="hover:bg-muted/20 transition-colors cursor-pointer"
                    onClick={() => navigate(`/problems/${p.id}`)}
                  >
                    <TableCell className="font-medium">
                      <span>{p.name}</span>
                      {nameCounts[p.name ?? ''] > 1 && latestIds.has(p.id ?? -1) && (
                        <span className="ml-2 text-xs font-medium text-emerald-500 bg-emerald-500/10 border border-emerald-500/20 px-1.5 py-0.5 rounded">
                          Latest
                        </span>
                      )}
                      {nameCounts[p.name ?? ''] > 1 && !latestIds.has(p.id ?? -1) && (
                        <span className="ml-2 text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded">
                          Earlier attempt
                        </span>
                      )}
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {(p.categories ?? []).map((cat) => (
                          <span key={cat} className="text-xs bg-muted px-2 py-0.5 rounded-md text-muted-foreground">
                            {cat}
                          </span>
                        ))}
                      </div>
                    </TableCell>
                    <TableCell>
                      <span className={`text-xs px-2 py-0.5 rounded-md font-medium ${DIFFICULTY_STYLES[p.difficulty ?? ''] ?? ''}`}>
                        {p.difficulty}
                      </span>
                    </TableCell>
                    <TableCell className="text-center text-muted-foreground">{p.attempts}</TableCell>
                    <TableCell className="text-center">
                      <span className={p.looked_at_solution ? 'text-amber-500 text-xs' : 'text-muted-foreground text-xs'}>
                        {p.looked_at_solution ? 'Yes' : 'No'}
                      </span>
                    </TableCell>
                    <TableCell className="text-center">
                      {p.solution_type === 'optimal'     && <span className="text-xs font-medium text-emerald-500">Optimal</span>}
                      {p.solution_type === 'brute_force' && <span className="text-xs font-medium text-amber-500">Brute force</span>}
                      {(!p.solution_type || p.solution_type === 'none') && <span className="text-xs text-muted-foreground">—</span>}
                    </TableCell>
                    <TableCell className="text-right">
                      {(p.original_score ?? 0) - (p.score ?? 0) > 0 ? (
                        <TooltipProvider delayDuration={150}>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <span className={`font-mono text-sm font-medium cursor-help underline decoration-dotted underline-offset-2 ${scoreColor(p.score ?? 0)}`}>
                                {p.score}
                              </span>
                            </TooltipTrigger>
                            <TooltipContent side="top" className="font-mono text-xs space-y-0.5">
                              <p>Original score: <span className="text-foreground">{p.original_score}</span></p>
                              <p>Decay: <span className="text-red-500">−{(p.original_score ?? 0) - (p.score ?? 0)}</span></p>
                              <p>Last solved: <span className="text-muted-foreground">{p.solved_at ? timeSince(new Date(p.solved_at)) : '—'}</span></p>
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      ) : (
                        <span className={`font-mono text-sm font-medium ${scoreColor(p.score ?? 0)}`}>
                          {p.score}
                        </span>
                      )}
                    </TableCell>
                    <TableCell className="text-right text-xs text-muted-foreground">
                      {p.solved_at ? new Date(p.solved_at).toLocaleDateString() : '—'}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between text-sm text-muted-foreground pt-1">
              <span>
                Page {currentPage} of {totalPages} · {total} result{total !== 1 ? 's' : ''}
              </span>
              <div className="flex gap-1">
                <Button
                  variant="outline" size="sm"
                  disabled={offset === 0}
                  onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
                  className="h-8 w-8 p-0"
                >
                  <ChevronLeft className="w-4 h-4" />
                </Button>
                <Button
                  variant="outline" size="sm"
                  disabled={offset + PAGE_SIZE >= total}
                  onClick={() => setOffset(offset + PAGE_SIZE)}
                  className="h-8 w-8 p-0"
                >
                  <ChevronRight className="w-4 h-4" />
                </Button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
