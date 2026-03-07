import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'sonner'
import axios from 'axios'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { api } from '@/lib/api'
import type { Problem, ApiError } from '@/types/api'

const DIFFICULTY_VARIANT: Record<string, 'default' | 'secondary' | 'destructive'> = {
  easy: 'default',
  medium: 'secondary',
  hard: 'destructive',
}

function scoreColor(score: number) {
  if (score >= 70) return 'text-green-600'
  if (score >= 40) return 'text-yellow-600'
  return 'text-red-600'
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
    return <p className="text-muted-foreground">Loading…</p>
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Problems</h1>
        <Button asChild size="sm">
          <Link to="/problems/new">+ Log problem</Link>
        </Button>
      </div>

      {problems.length === 0 ? (
        <p className="text-muted-foreground py-12 text-center">
          No problems logged yet.{' '}
          <Link to="/problems/new" className="underline underline-offset-4 hover:text-primary">
            Log your first one.
          </Link>
        </p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Category</TableHead>
              <TableHead>Difficulty</TableHead>
              <TableHead className="text-center">Attempts</TableHead>
              <TableHead className="text-center">Peeked</TableHead>
              <TableHead className="text-right">Score</TableHead>
              <TableHead className="text-right">Decayed</TableHead>
              <TableHead className="text-right">Solved</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {problems.map((p) => (
              <TableRow key={p.id}>
                <TableCell className="font-medium">{p.name}</TableCell>
                <TableCell>{p.category}</TableCell>
                <TableCell>
                  <Badge variant={DIFFICULTY_VARIANT[p.difficulty ?? ''] ?? 'default'}>
                    {p.difficulty}
                  </Badge>
                </TableCell>
                <TableCell className="text-center">{p.attempts}</TableCell>
                <TableCell className="text-center">{p.looked_at_solution ? 'Yes' : 'No'}</TableCell>
                <TableCell className={`text-right font-mono ${scoreColor(p.score ?? 0)}`}>
                  {p.score}
                </TableCell>
                <TableCell className={`text-right font-mono ${scoreColor(p.decayed_score ?? 0)}`}>
                  {Math.round(p.decayed_score ?? 0)}
                </TableCell>
                <TableCell className="text-right text-muted-foreground text-sm">
                  {p.solved_at ? new Date(p.solved_at).toLocaleDateString() : '—'}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </div>
  )
}
