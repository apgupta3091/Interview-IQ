import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'sonner'
import axios from 'axios'
import { Plus, FileX } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { api } from '@/lib/api'
import type { Problem, ApiError } from '@/types/api'

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

export default function ProblemList() {
  const [problems, setProblems] = useState<Problem[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.problems.list()
      .then(setProblems)
      .catch((err) => {
        if (axios.isAxiosError(err)) {
          const msg = (err.response?.data as ApiError)?.error ?? 'Failed to load problems'
          toast.error(msg)
        } else {
          toast.error('Unexpected error')
        }
      })
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <div className="space-y-3 animate-pulse">
        <div className="h-7 w-28 bg-muted rounded" />
        {[...Array(5)].map((_, i) => (
          <div key={i} className="h-12 bg-muted rounded" />
        ))}
      </div>
    )
  }

  return (
    <div className="space-y-5 animate-fade-up">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Problems</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {problems.length} problem{problems.length !== 1 ? 's' : ''} logged
          </p>
        </div>
        <Button asChild size="sm">
          <Link to="/problems/new">
            <Plus className="w-4 h-4 mr-1" />
            Log problem
          </Link>
        </Button>
      </div>

      {problems.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-32 text-center">
          <div className="w-12 h-12 rounded-xl bg-muted flex items-center justify-center mb-4">
            <FileX className="w-6 h-6 text-muted-foreground" />
          </div>
          <p className="text-base font-semibold mb-1">No problems yet</p>
          <p className="text-sm text-muted-foreground">
            <Link to="/problems/new" className="text-primary hover:underline underline-offset-4">
              Log your first problem
            </Link>{' '}
            to get started.
          </p>
        </div>
      ) : (
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
                  <TableCell className="font-medium">{p.name}</TableCell>
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
                    {p.solution_type === 'optimal' && (
                      <span className="text-xs font-medium text-emerald-500">Optimal</span>
                    )}
                    {p.solution_type === 'brute_force' && (
                      <span className="text-xs font-medium text-amber-500">Brute force</span>
                    )}
                    {(!p.solution_type || p.solution_type === 'none') && (
                      <span className="text-xs text-muted-foreground">—</span>
                    )}
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
      )}
    </div>
  )
}
