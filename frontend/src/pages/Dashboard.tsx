import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'sonner'
import axios from 'axios'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import SkillRadar from '@/components/SkillRadar'
import { api } from '@/lib/api'
import type { CategoryStats, WeakestResult, ApiError } from '@/types/api'

function strengthColor(strength: number) {
  if (strength >= 70) return 'text-green-600'
  if (strength >= 40) return 'text-yellow-600'
  return 'text-red-500'
}

function StrengthBar({ value }: { value: number }) {
  return (
    <div className="w-full bg-muted rounded-full h-2 mt-2">
      <div
        className="h-2 rounded-full bg-primary transition-all"
        style={{ width: `${Math.min(value, 100)}%` }}
      />
    </div>
  )
}

export default function Dashboard() {
  const [stats, setStats] = useState<CategoryStats[]>([])
  const [weakest, setWeakest] = useState<WeakestResult | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    Promise.all([api.categories.stats(), api.categories.weakest()])
      .then(([s, w]) => {
        setStats(s)
        setWeakest(w)
      })
      .catch((err) => {
        if (axios.isAxiosError(err)) {
          const msg = (err.response?.data as ApiError)?.error ?? 'Failed to load dashboard'
          // 404 means no problems logged yet — not a real error
          if (err.response?.status !== 404) toast.error(msg)
        } else {
          toast.error('Unexpected error')
        }
      })
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <p className="text-muted-foreground">Loading…</p>

  if (stats.length === 0) {
    return (
      <div className="py-24 text-center space-y-3">
        <p className="text-xl font-semibold">No data yet</p>
        <p className="text-muted-foreground">
          <Link to="/problems/new" className="underline underline-offset-4 hover:text-primary">
            Log your first problem
          </Link>{' '}
          to see your skill radar.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      <h1 className="text-2xl font-semibold">Dashboard</h1>

      {/* Weakest category banner */}
      {weakest?.category && (
        <Card className="border-yellow-400 bg-yellow-50 dark:bg-yellow-950">
          <CardHeader className="pb-2">
            <div className="flex items-center gap-2">
              <CardTitle className="text-base">Focus area</CardTitle>
              <Badge variant="secondary">{weakest.category}</Badge>
            </div>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground mb-2">
              Your weakest category — strength{' '}
              <span className={`font-semibold ${strengthColor(weakest.strength ?? 0)}`}>
                {Math.round(weakest.strength ?? 0)}%
              </span>
              . Try these problems:
            </p>
            <ul className="list-disc list-inside space-y-1">
              {weakest.recommendations?.map((r) => (
                <li key={r} className="text-sm">{r}</li>
              ))}
            </ul>
          </CardContent>
        </Card>
      )}

      {/* Skill radar chart */}
      <Card>
        <CardHeader className="pb-0">
          <CardTitle className="text-base">Skill radar</CardTitle>
        </CardHeader>
        <CardContent>
          <SkillRadar stats={stats} />
        </CardContent>
      </Card>

      {/* Category strength grid */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
        {stats.map((s) => (
          <Card key={s.category}>
            <CardContent className="pt-4 pb-3">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium truncate">{s.category}</span>
                <span className={`text-sm font-semibold tabular-nums ${strengthColor(s.strength ?? 0)}`}>
                  {Math.round(s.strength ?? 0)}%
                </span>
              </div>
              <StrengthBar value={s.strength ?? 0} />
              <p className="text-xs text-muted-foreground mt-1.5">
                {s.problem_count} {s.problem_count === 1 ? 'problem' : 'problems'}
              </p>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}
