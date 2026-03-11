import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import axios from 'axios'
import { ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid,
  Tooltip, ResponsiveContainer,
} from 'recharts'
import { api } from '@/lib/api'
import type { Problem, ApiError } from '@/types/api'

const DIFFICULTY_STYLES: Record<string, string> = {
  easy:   'bg-emerald-500/10 text-emerald-500 border border-emerald-500/20',
  medium: 'bg-amber-500/10  text-amber-500  border border-amber-500/20',
  hard:   'bg-red-500/10    text-red-500    border border-red-500/20',
}

function scoreColor(s: number) {
  if (s >= 70) return 'text-emerald-500'
  if (s >= 40) return 'text-amber-500'
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

export default function ProblemDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [problem, setProblem] = useState<Problem | null>(null)
  const [history, setHistory] = useState<Problem[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!id) return
    const numId = parseInt(id, 10)
    if (isNaN(numId)) { navigate('/problems'); return }

    setLoading(true)
    api.problems.getById(numId)
      .then(async (p) => {
        setProblem(p)
        // Fetch all attempts with the same name for score history
        const result = await api.problems.listFiltered({ q: p.name, limit: 100 })
        const exact = (result.problems ?? [])
          .filter((h) => h.name === p.name)
          .sort((a, b) =>
            new Date(a.solved_at ?? '').getTime() - new Date(b.solved_at ?? '').getTime(),
          )
        setHistory(exact)
      })
      .catch((err: unknown) => {
        if (axios.isAxiosError(err) && err.response?.status === 404) {
          toast.error('Problem not found')
          navigate('/problems')
        } else {
          const msg = axios.isAxiosError(err)
            ? ((err.response?.data as ApiError)?.error ?? 'Failed to load problem')
            : 'Failed to load problem'
          toast.error(msg)
        }
      })
      .finally(() => setLoading(false))
  }, [id, navigate])

  if (loading) {
    return (
      <div className="space-y-4 animate-pulse">
        <div className="h-8 bg-muted rounded w-1/3" />
        <div className="h-48 bg-muted rounded" />
        <div className="h-48 bg-muted rounded" />
      </div>
    )
  }

  if (!problem) return null

  const decayDelta = (problem.original_score ?? 0) - (problem.score ?? 0)
  const isDecayed = decayDelta > 0
  const solvedDate = problem.solved_at ? new Date(problem.solved_at) : null

  const chartData = history.map((h, i) => ({
    label: h.solved_at ? new Date(h.solved_at).toLocaleDateString() : `#${i + 1}`,
    score: h.score ?? 0,
  }))

  return (
    <div className="space-y-6 animate-fade-up">
      {/* Header */}
      <div className="space-y-3">
        <Button
          variant="ghost"
          size="sm"
          className="-ml-2 text-muted-foreground"
          onClick={() => navigate('/problems')}
        >
          <ArrowLeft className="w-4 h-4 mr-1" />
          Problems
        </Button>
        <h1 className="text-2xl font-bold tracking-tight">{problem.name}</h1>
        <div className="flex flex-wrap items-center gap-2">
          <span className={`text-xs px-2 py-0.5 rounded-md font-medium ${DIFFICULTY_STYLES[problem.difficulty ?? ''] ?? ''}`}>
            {problem.difficulty}
          </span>
          {(problem.categories ?? []).map((cat) => (
            <span key={cat} className="text-xs bg-muted px-2 py-0.5 rounded-md text-muted-foreground">
              {cat}
            </span>
          ))}
        </div>
      </div>

      {/* Score + Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <Card className="border-border/60">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Score</CardTitle>
          </CardHeader>
          <CardContent>
            <p className={`text-5xl font-bold font-mono ${scoreColor(problem.score ?? 0)}`}>
              {problem.score}
            </p>
            {isDecayed && (
              <div className="mt-2 space-y-0.5 text-sm">
                <p className="text-muted-foreground">
                  Original: <span className="text-foreground font-mono">{problem.original_score}</span>
                </p>
                <p className="text-red-500 font-mono">−{decayDelta} decay</p>
                {solvedDate && (
                  <p className="text-muted-foreground text-xs">{timeSince(solvedDate)}</p>
                )}
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="border-border/60">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Details</CardTitle>
          </CardHeader>
          <CardContent>
            <dl className="grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
              <div>
                <dt className="text-muted-foreground text-xs">Attempts</dt>
                <dd className="font-medium">{problem.attempts}</dd>
              </div>
              <div>
                <dt className="text-muted-foreground text-xs">Time taken</dt>
                <dd className="font-medium">{problem.time_taken_mins} min</dd>
              </div>
              <div>
                <dt className="text-muted-foreground text-xs">Solution type</dt>
                <dd className="font-medium">
                  {problem.solution_type === 'brute_force'
                    ? 'Brute force'
                    : problem.solution_type === 'optimal'
                    ? 'Optimal'
                    : '—'}
                </dd>
              </div>
              <div>
                <dt className="text-muted-foreground text-xs">Peeked at solution</dt>
                <dd className={`font-medium ${problem.looked_at_solution ? 'text-amber-500' : ''}`}>
                  {problem.looked_at_solution ? 'Yes' : 'No'}
                </dd>
              </div>
              <div className="col-span-2">
                <dt className="text-muted-foreground text-xs">Solved at</dt>
                <dd className="font-medium">
                  {solvedDate
                    ? solvedDate.toLocaleDateString(undefined, { year: 'numeric', month: 'long', day: 'numeric' })
                    : '—'}
                </dd>
              </div>
            </dl>
          </CardContent>
        </Card>
      </div>

      {/* Notes */}
      <Card className="border-border/60">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium text-muted-foreground">Your Notes</CardTitle>
        </CardHeader>
        <CardContent>
          {problem.notes ? (
            <p className="text-sm whitespace-pre-wrap">{problem.notes}</p>
          ) : (
            <p className="text-sm text-muted-foreground/50 italic">No notes recorded for this attempt.</p>
          )}
        </CardContent>
      </Card>

      {/* Score History */}
      <div className="space-y-3">
        <h2 className="text-base font-semibold">All Attempts for {problem.name}</h2>
        {history.length <= 1 ? (
          <p className="text-sm text-muted-foreground">This is your only recorded attempt.</p>
        ) : (
          <>
            <Card className="border-border/60">
              <CardContent className="pt-4">
                <ResponsiveContainer width="100%" height={200}>
                  <LineChart data={chartData} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
                    <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
                    <XAxis dataKey="label" tick={{ fontSize: 11 }} />
                    <YAxis domain={[0, 100]} tick={{ fontSize: 11 }} />
                    <Tooltip
                      contentStyle={{ fontSize: 12 }}
                      formatter={(value: number) => [value, 'Score']}
                    />
                    <Line
                      type="monotone"
                      dataKey="score"
                      stroke="hsl(var(--primary))"
                      strokeWidth={2}
                      dot={{ r: 3 }}
                    />
                  </LineChart>
                </ResponsiveContainer>
              </CardContent>
            </Card>

            <div className="rounded-lg border border-border/60 overflow-hidden">
              <table className="w-full text-sm">
                <thead>
                  <tr className="bg-muted/30 text-muted-foreground">
                    <th className="px-4 py-2 text-left font-medium">Date</th>
                    <th className="px-4 py-2 text-center font-medium">Attempts</th>
                    <th className="px-4 py-2 text-center font-medium">Time</th>
                    <th className="px-4 py-2 text-center font-medium">Peeked</th>
                    <th className="px-4 py-2 text-right font-medium">Score</th>
                  </tr>
                </thead>
                <tbody>
                  {history.map((h) => (
                    <tr key={h.id} className="border-t border-border/60 hover:bg-muted/20">
                      <td className="px-4 py-2 text-muted-foreground">
                        {h.solved_at ? new Date(h.solved_at).toLocaleDateString() : '—'}
                      </td>
                      <td className="px-4 py-2 text-center">{h.attempts}</td>
                      <td className="px-4 py-2 text-center">{h.time_taken_mins} min</td>
                      <td className="px-4 py-2 text-center">
                        <span className={h.looked_at_solution ? 'text-amber-500' : 'text-muted-foreground'}>
                          {h.looked_at_solution ? 'Yes' : 'No'}
                        </span>
                      </td>
                      <td className={`px-4 py-2 text-right font-mono font-medium ${scoreColor(h.score ?? 0)}`}>
                        {h.score}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
