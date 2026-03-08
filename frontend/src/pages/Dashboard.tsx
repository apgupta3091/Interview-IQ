import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'sonner'
import axios from 'axios'
import { AlertTriangle, Loader2, Sparkles, TrendingUp } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import CategoryBarChart from '@/components/CategoryBarChart'
import CategoryRadarChart from '@/components/CategoryRadarChart'
import { api } from '@/lib/api'
import type { CategoryRec, CategoryStats, WeakestResult, ApiError } from '@/types/api'

function strengthColor(s: number) {
  if (s >= 70) return 'text-emerald-500'
  if (s >= 40) return 'text-amber-500'
  return 'text-red-500'
}

function difficultyClass(d: string) {
  if (d === 'easy') return 'text-emerald-500'
  if (d === 'hard') return 'text-red-500'
  return 'text-amber-500'
}

export default function Dashboard() {
  const [stats, setStats] = useState<CategoryStats[]>([])
  const [weakest, setWeakest] = useState<WeakestResult | null>(null)
  const [loading, setLoading] = useState(true)

  // AI popover state
  const [popoverOpen, setPopoverOpen] = useState(false)
  const [aiLoading, setAiLoading] = useState(false)
  const [aiRec, setAiRec] = useState<CategoryRec | null>(null)
  const [aiError, setAiError] = useState<string | null>(null)

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

  function handleAIClick() {
    if (!weakest?.category) return
    setPopoverOpen(true)
    setAiRec(null)
    setAiError(null)
    setAiLoading(true)
    api.recommendations
      .get({ category: weakest.category, limit: 3 })
      .then((result) => {
        setAiRec(result.categories[0] ?? null)
      })
      .catch((err) => {
        const msg = axios.isAxiosError(err)
          ? ((err.response?.data as ApiError)?.error ?? 'Failed to get recommendations')
          : 'Unexpected error'
        setAiError(msg)
      })
      .finally(() => setAiLoading(false))
  }

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
          <div className="flex items-start justify-between gap-3">
            <div className="flex gap-3 min-w-0">
              <AlertTriangle className="w-4 h-4 text-amber-500 mt-0.5 shrink-0" />
              <div className="min-w-0">
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

            {/* AI recommendations popover */}
            <Popover open={popoverOpen} onOpenChange={setPopoverOpen}>
              <PopoverTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6 shrink-0 text-amber-500 hover:text-amber-600 hover:bg-amber-500/10"
                  onClick={handleAIClick}
                  title="Get AI recommendations"
                >
                  <Sparkles className="w-3.5 h-3.5" />
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-80 p-4" align="end">
                {aiLoading && (
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <Loader2 className="w-4 h-4 animate-spin" />
                    Getting AI recommendations…
                  </div>
                )}
                {!aiLoading && aiError && (
                  <p className="text-sm text-red-500">{aiError}</p>
                )}
                {!aiLoading && aiRec && (
                  <div className="space-y-3">
                    <p className="text-xs text-muted-foreground italic leading-relaxed">
                      {aiRec.focus_note}
                    </p>
                    <ul className="space-y-2">
                      {aiRec.problems.map((p) => (
                        <li key={p.name} className="space-y-0.5">
                          <div className="flex items-center gap-1.5">
                            <span className="text-sm font-medium leading-tight">{p.name}</span>
                            <span className={`text-xs font-medium ${difficultyClass(p.difficulty)}`}>
                              {p.difficulty}
                            </span>
                          </div>
                          <p className="text-xs text-muted-foreground leading-relaxed">
                            {p.description}
                          </p>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </PopoverContent>
            </Popover>
          </div>
        </div>
      )}

      {/* Two-column charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">

        {/* Radar — shape at a glance */}
        <Card className="border-border/60">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
              Skill Radar
            </CardTitle>
          </CardHeader>
          <CardContent>
            <CategoryRadarChart stats={stats} />
          </CardContent>
        </Card>

        {/* Bar chart — precise values */}
        <Card className="border-border/60">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
              Category Strength
            </CardTitle>
          </CardHeader>
          <CardContent>
            <CategoryBarChart stats={stats} />
          </CardContent>
        </Card>

      </div>
    </div>
  )
}
