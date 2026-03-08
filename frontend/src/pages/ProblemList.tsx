import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'sonner'
import axios from 'axios'
import { Plus, FileX, ChevronLeft, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { api } from '@/lib/api'
import type { Problem, ApiError, ProblemListResponse } from '@/types/api'
import ProblemFilters from '@/components/ProblemFilters'

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

export default function ProblemList() {
  const [nameSearch, setNameSearch]                   = useState('')
  const [debouncedSearch, setDebouncedSearch]         = useState('')
  const [selectedCategories, setSelectedCategories]   = useState<string[]>([])
  const [selectedDifficulties, setSelectedDifficulties] = useState<string[]>([])
  const [scoreMin, setScoreMin]   = useState('')
  const [scoreMax, setScoreMax]   = useState('')
  const [dateFrom, setDateFrom]   = useState('')
  const [dateTo, setDateTo]       = useState('')
  const [offset, setOffset]       = useState(0)

  const [result, setResult]   = useState<ProblemListResponse>({ problems: [], total: 0, limit: PAGE_SIZE, offset: 0 })
  const [loading, setLoading] = useState(true)

  // Debounce the name search by 400 ms; reset to page 1 when the value settles.
  useEffect(() => {
    const id = setTimeout(() => {
      setDebouncedSearch(nameSearch)
      setOffset(0)
    }, 400)
    return () => clearTimeout(id)
  }, [nameSearch])

  // Fetch whenever filters or pagination change.
  useEffect(() => {
    setLoading(true)
    api.problems.listFiltered({
      q:          debouncedSearch || undefined,
      category:   selectedCategories.length ? selectedCategories : undefined,
      difficulty: selectedDifficulties.length ? selectedDifficulties : undefined,
      score_min:  scoreMin !== '' ? Number(scoreMin) : undefined,
      score_max:  scoreMax !== '' ? Number(scoreMax) : undefined,
      from:       dateFrom || undefined,
      to:         dateTo   || undefined,
      limit:      PAGE_SIZE,
      offset,
    })
      .then(setResult)
      .catch((err) => {
        if (axios.isAxiosError(err)) {
          const msg = (err.response?.data as ApiError)?.error ?? 'Failed to load problems'
          toast.error(msg)
        } else {
          toast.error('Unexpected error')
        }
      })
      .finally(() => setLoading(false))
  }, [debouncedSearch, selectedCategories, selectedDifficulties, scoreMin, scoreMax, dateFrom, dateTo, offset])

  const hasFilters = !!(
    nameSearch || selectedCategories.length || selectedDifficulties.length ||
    scoreMin || scoreMax || dateFrom || dateTo
  )

  const clearFilters = () => {
    setNameSearch('')
    setSelectedCategories([])
    setSelectedDifficulties([])
    setScoreMin('')
    setScoreMax('')
    setDateFrom('')
    setDateTo('')
    setOffset(0)
  }

  const problems = result.problems ?? []
  const total    = result.total    ?? 0
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
            {loading ? '…' : `${total} problem${total !== 1 ? 's' : ''}${hasFilters ? ' matched' : ' logged'}`}
          </p>
        </div>
        <Button asChild size="sm">
          <Link to="/problems/new">
            <Plus className="w-4 h-4 mr-1" />
            Log problem
          </Link>
        </Button>
      </div>

      {/* Filters — each setter also resets to page 1 so there's no separate
          "reset offset" effect competing with the fetch effect on mount. */}
      <ProblemFilters
        nameSearch={nameSearch}           onNameSearch={setNameSearch}
        dateFrom={dateFrom}               onDateFrom={(v) => { setDateFrom(v); setOffset(0) }}
        dateTo={dateTo}                   onDateTo={(v) => { setDateTo(v); setOffset(0) }}
        selectedCategories={selectedCategories}     onCategoriesChange={(v) => { setSelectedCategories(v); setOffset(0) }}
        selectedDifficulties={selectedDifficulties} onDifficultiesChange={(v) => { setSelectedDifficulties(v); setOffset(0) }}
        scoreMin={scoreMin}               onScoreMin={(v) => { setScoreMin(v); setOffset(0) }}
        scoreMax={scoreMax}               onScoreMax={(v) => { setScoreMax(v); setOffset(0) }}
        hasFilters={hasFilters}           onClear={clearFilters}
      />

      {/* Loading skeleton */}
      {loading ? (
        <div className="space-y-3 animate-pulse">
          {[...Array(5)].map((_, i) => <div key={i} className="h-12 bg-muted rounded" />)}
        </div>
      ) : problems.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-32 text-center">
          <div className="w-12 h-12 rounded-xl bg-muted flex items-center justify-center mb-4">
            <FileX className="w-6 h-6 text-muted-foreground" />
          </div>
          <p className="text-base font-semibold mb-1">
            {hasFilters ? 'No problems match your filters' : 'No problems yet'}
          </p>
          <p className="text-sm text-muted-foreground">
            {hasFilters ? (
              <button onClick={clearFilters} className="text-primary hover:underline underline-offset-4">
                Clear filters
              </button>
            ) : (
              <Link to="/problems/new" className="text-primary hover:underline underline-offset-4">
                Log your first problem
              </Link>
            )}{' '}
            {!hasFilters && 'to get started.'}
          </p>
        </div>
      ) : (
        <>
          <div className="rounded-lg border border-border/60 overflow-hidden">
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
                  <TableHead className="text-right font-medium">Decayed</TableHead>
                  <TableHead className="text-right font-medium">Solved</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {problems.map((p) => (
                  <TableRow key={p.id} className="hover:bg-muted/20 transition-colors">
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
                      {p.solution_type === 'optimal'    && <span className="text-xs font-medium text-emerald-500">Optimal</span>}
                      {p.solution_type === 'brute_force' && <span className="text-xs font-medium text-amber-500">Brute force</span>}
                      {(!p.solution_type || p.solution_type === 'none') && <span className="text-xs text-muted-foreground">—</span>}
                    </TableCell>
                    <TableCell className={`text-right font-mono text-sm font-medium ${scoreColor(p.score ?? 0)}`}>
                      {p.score}
                    </TableCell>
                    <TableCell className={`text-right font-mono text-sm font-medium ${scoreColor(p.decayed_score ?? 0)}`}>
                      {Math.round(p.decayed_score ?? 0)}
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
        </>
      )}
    </div>
  )
}
