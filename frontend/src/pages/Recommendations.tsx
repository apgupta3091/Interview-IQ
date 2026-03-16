import { useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'sonner'
import axios from 'axios'
import { Lock, Loader2, Sparkles, X, Rocket } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { CATEGORIES } from '@/lib/constants'
import { api } from '@/lib/api'
import { useBillingTier } from '@/hooks/useBillingTier'
import type { ApiError, CategoryRec, RecommendationParams } from '@/types/api'

const COMING_SOON = true

function strengthBadgeClass(strength: number) {
  if (strength >= 70) return 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
  if (strength >= 40) return 'bg-amber-500/10 text-amber-600 dark:text-amber-400'
  return 'bg-red-500/10 text-red-600 dark:text-red-400'
}

function difficultyClass(d: string) {
  if (d === 'easy') return 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
  if (d === 'hard') return 'bg-red-500/10 text-red-600 dark:text-red-400'
  return 'bg-amber-500/10 text-amber-600 dark:text-amber-400'
}

function CategoryCard({ rec }: { rec: CategoryRec }) {
  return (
    <Card className="border-border/60">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between gap-2 flex-wrap">
          <CardTitle className="text-base font-semibold capitalize">{rec.category}</CardTitle>
          <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${strengthBadgeClass(rec.strength)}`}>
            {Math.round(rec.strength)}% strength
          </span>
        </div>
        <p className="text-sm text-muted-foreground italic leading-relaxed mt-1">
          {rec.focus_note}
        </p>
      </CardHeader>
      <CardContent className="pt-0 space-y-3">
        {rec.problems.map((p) => (
          <div key={p.name} className="flex flex-col gap-0.5 rounded-lg border border-border/50 p-3">
            <div className="flex items-center gap-2 flex-wrap">
              <span className="text-sm font-medium">{p.name}</span>
              <span className={`text-xs font-medium px-1.5 py-0.5 rounded capitalize ${difficultyClass(p.difficulty)}`}>
                {p.difficulty}
              </span>
            </div>
            <p className="text-xs text-muted-foreground leading-relaxed">{p.description}</p>
            {p.reason && (
              <p className="text-xs text-muted-foreground/70 leading-relaxed border-t border-border/40 pt-1.5 mt-1">
                <span className="font-medium text-muted-foreground">Why: </span>{p.reason}
              </p>
            )}
          </div>
        ))}
      </CardContent>
    </Card>
  )
}

function LoadingSkeleton() {
  return (
    <div className="grid grid-cols-1 gap-4">
      {[1].map((i) => (
        <Card key={i} className="border-border/60">
          <CardHeader className="pb-3 space-y-2">
            <div className="flex items-center justify-between">
              <Skeleton className="h-5 w-24" />
              <Skeleton className="h-5 w-20" />
            </div>
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-3/4" />
          </CardHeader>
          <CardContent className="pt-0 space-y-2">
            {[1, 2, 3].map((j) => (
              <div key={j} className="rounded-lg border border-border/50 p-3 space-y-1.5">
                <div className="flex items-center gap-2">
                  <Skeleton className="h-4 w-40" />
                  <Skeleton className="h-4 w-12" />
                </div>
                <Skeleton className="h-3 w-full" />
              </div>
            ))}
          </CardContent>
        </Card>
      ))}
    </div>
  )
}

export default function Recommendations() {
  if (COMING_SOON) {
    return (
      <div className="space-y-6 animate-fade-up">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Recommendations</h1>
          <p className="text-sm text-muted-foreground mt-1">
            AI-powered problem suggestions tailored to your weak spots
          </p>
        </div>
        <div className="flex flex-col items-center justify-center py-24 gap-4 text-center rounded-lg border border-border/60">
          <div className="w-12 h-12 rounded-xl bg-muted flex items-center justify-center">
            <Rocket className="w-5 h-5 text-muted-foreground" />
          </div>
          <div className="space-y-1">
            <p className="text-base font-semibold">Coming Soon</p>
            <p className="text-sm text-muted-foreground max-w-xs">
              AI-powered recommendations are on their way. Check back soon!
            </p>
          </div>
        </div>
      </div>
    )
  }

  const tier = useBillingTier()

  // Draft form state (not yet applied)
  const [draftCategory, setDraftCategory] = useState('')
  const [draftLimit, setDraftLimit] = useState('3')

  // Results state
  const [loading, setLoading] = useState(false)
  const [results, setResults] = useState<CategoryRec[] | null>(null)
  const [hasFetched, setHasFetched] = useState(false)

  function handleGet() {
    const limit = Math.min(10, Math.max(1, parseInt(draftLimit, 10) || 3))
    const params: RecommendationParams = {
      limit,
      ...(draftCategory && { category: draftCategory }),
    }

    setLoading(true)
    setHasFetched(true)
    api.recommendations
      .get(params)
      .then((result) => {
        setResults(result.categories)
      })
      .catch((err) => {
        const msg = axios.isAxiosError(err)
          ? ((err.response?.data as ApiError)?.error ?? 'Failed to get recommendations')
          : 'Unexpected error'
        toast.error(msg)
        setResults(null)
      })
      .finally(() => setLoading(false))
  }

  function handleClear() {
    setDraftCategory('')
    setDraftLimit('3')
    setResults(null)
    setHasFetched(false)
  }

  const hasFilters = !!draftCategory || draftLimit !== '3'

  if (tier === 'free') {
    return (
      <div className="space-y-6 animate-fade-up">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Recommendations</h1>
          <p className="text-sm text-muted-foreground mt-1">
            AI-powered problem suggestions tailored to your weak spots
          </p>
        </div>
        <div className="flex flex-col items-center justify-center py-24 gap-4 text-center rounded-lg border border-border/60">
          <div className="w-12 h-12 rounded-xl bg-muted flex items-center justify-center">
            <Lock className="w-5 h-5 text-muted-foreground" />
          </div>
          <div className="space-y-1">
            <p className="text-base font-semibold">Pro feature</p>
            <p className="text-sm text-muted-foreground max-w-xs">
              AI-powered recommendations are available on the Pro plan.
            </p>
          </div>
          <Button asChild size="sm">
            <Link to="/pricing">Upgrade to Pro</Link>
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-fade-up">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Recommendations</h1>
        <p className="text-sm text-muted-foreground mt-1">
          AI-powered problem suggestions tailored to your weak spots
        </p>
      </div>

      {/* Form */}
      <div className="rounded-lg border border-border/60 p-4 space-y-3">
        <div className="flex flex-wrap items-center gap-2">
          <Select value={draftCategory} onValueChange={setDraftCategory}>
            <SelectTrigger className="h-8 text-xs w-40">
              <SelectValue placeholder="Select a category" />
            </SelectTrigger>
            <SelectContent>
              {CATEGORIES.map((c) => (
                <SelectItem key={c} value={c} className="text-xs capitalize">
                  {c}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <div className="flex items-center gap-1">
            <span className="text-xs text-muted-foreground">Per category</span>
            <Input
              type="number"
              value={draftLimit}
              onChange={(e) => setDraftLimit(e.target.value)}
              className="h-8 text-xs w-16"
              min={1}
              max={10}
            />
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button size="sm" className="h-8 text-xs gap-1" onClick={handleGet} disabled={loading || !draftCategory}>
            {loading ? (
              <Loader2 className="w-3.5 h-3.5 animate-spin" />
            ) : (
              <Sparkles className="w-3.5 h-3.5" />
            )}
            Get Recommendations
          </Button>
          {hasFilters && (
            <Button
              variant="ghost"
              size="sm"
              className="h-8 text-xs text-muted-foreground gap-1"
              onClick={handleClear}
              disabled={loading}
            >
              <X className="w-3.5 h-3.5" />
              Clear
            </Button>
          )}
        </div>
      </div>

      {/* Results */}
      {loading && <LoadingSkeleton />}

      {!loading && hasFetched && results && results.length > 0 && (
        <div className="space-y-4">
          {results.map((rec) => (
            <CategoryCard key={rec.category} rec={rec} />
          ))}
        </div>
      )}

      {!loading && hasFetched && results && results.length === 0 && (
        <div className="text-center py-16 text-sm text-muted-foreground">
          No recommendations returned. Try selecting different categories.
        </div>
      )}

      {!hasFetched && (
        <div className="text-center py-16 text-sm text-muted-foreground">
          Select a category and click <span className="font-medium text-foreground">Get Recommendations</span> to see AI-powered suggestions.
        </div>
      )}
    </div>
  )
}
