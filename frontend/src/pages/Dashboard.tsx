import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'sonner'
import axios from 'axios'
import { AlertTriangle, TrendingUp } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import SkillRadar from '@/components/SkillRadar'
import { api } from '@/lib/api'
import type { CategoryStats, WeakestResult, ApiError } from '@/types/api'

function strengthColor(s: number) {
  if (s >= 70) return 'text-emerald-500'
  if (s >= 40) return 'text-amber-500'
  return 'text-red-500'
}

function strengthBg(s: number) {
  if (s >= 70) return 'bg-emerald-500'
  if (s >= 40) return 'bg-amber-500'
  return 'bg-red-500'
}

function StrengthBar({ value }: { value: number }) {
  const [width, setWidth] = useState(0)
  useEffect(() => {
    const t = setTimeout(() => setWidth(Math.min(value, 100)), 50)
    return () => clearTimeout(t)
  }, [value])

  return (
    <div className="w-full bg-muted rounded-full h-1.5 mt-2 overflow-hidden">
      <div
        className={`h-1.5 rounded-full transition-all duration-700 ease-out ${strengthBg(value)}`}
        style={{ width: `${width}%` }}
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
      .then(([s, w]) => { setStats(s); setWeakest(w) })
      .catch((err) => {
        if (axios.isAxiosError(err)) {
          const msg = (err.response?.data as ApiError)?.error ?? 'Failed to load dashboard'
          if (err.response?.status !== 404) toast.error(msg)
        } else {
          toast.error('Unexpected error')
        }
      })
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <div className="space-y-4 animate-pulse">
        <div className="h-7 w-32 bg-muted rounded" />
        <div className="h-24 bg-muted rounded-lg" />
        <div className="h-96 bg-muted rounded-lg" />
      </div>
    )
  }

  if (stats.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-32 text-center animate-fade-up">
        <div className="w-12 h-12 rounded-xl bg-muted flex items-center justify-center mb-4">
          <TrendingUp className="w-6 h-6 text-muted-foreground" />
        </div>
        <p className="text-base font-semibold mb-1">No data yet</p>
        <p className="text-sm text-muted-foreground">
          <Link to="/problems/new" className="text-primary hover:underline underline-offset-4">
            Log your first problem
          </Link>{' '}
          to see your skill radar.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-fade-up">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-sm text-muted-foreground mt-1">Your interview prep skill overview</p>
      </div>

      {/* Weakest category */}
      {weakest?.category && (
        <div className="rounded-lg border border-amber-500/20 bg-amber-500/5 p-4">
          <div className="flex gap-3">
            <AlertTriangle className="w-4 h-4 text-amber-500 mt-0.5 shrink-0" />
            <div>
              <p className="text-sm font-medium">
                Focus on{' '}
                <span className="text-amber-500 font-semibold">{weakest.category}</span>
                {' '}—{' '}
                <span className={strengthColor(weakest.strength ?? 0)}>
                  {Math.round(weakest.strength ?? 0)}% strength
                </span>
              </p>
              <ul className="mt-2 space-y-1">
                {weakest.recommendations?.map((r) => (
                  <li key={r} className="text-sm text-muted-foreground flex items-center gap-2">
                    <span className="w-1 h-1 rounded-full bg-muted-foreground/50 shrink-0" />
                    {r}
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>
      )}

      {/* Radar chart */}
      <Card className="border-border/60">
        <CardHeader className="pb-0">
          <CardTitle className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
            Skill Radar
          </CardTitle>
        </CardHeader>
        <CardContent className="pt-2">
          <SkillRadar stats={stats} />
        </CardContent>
      </Card>

      {/* Category grid */}
      <div>
        <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wider mb-3">
          Category Breakdown
        </h2>
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">
          {stats.map((s) => (
            <Card key={s.category} className="border-border/60 hover:border-border transition-colors">
              <CardContent className="p-4">
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm font-medium truncate pr-2">{s.category}</span>
                  <span className={`text-sm font-bold tabular-nums shrink-0 ${strengthColor(s.strength ?? 0)}`}>
                    {Math.round(s.strength ?? 0)}%
                  </span>
                </div>
                <StrengthBar value={s.strength ?? 0} />
                <p className="text-xs text-muted-foreground mt-2">
                  {s.problem_count} {s.problem_count === 1 ? 'problem' : 'problems'}
                </p>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </div>
  )
}
