import { useEffect, useState } from 'react'
import type { CategoryStats } from '@/types/api'

type Props = { stats: CategoryStats[] }

function barColor(s: number) {
  if (s >= 70) return 'bg-emerald-500'
  if (s >= 40) return 'bg-amber-500'
  return 'bg-red-500'
}

function scoreColor(s: number) {
  if (s >= 70) return 'text-emerald-500'
  if (s >= 40) return 'text-amber-500'
  return 'text-red-500'
}

export default function CategoryBars({ stats }: Props) {
  const [animate, setAnimate] = useState(false)

  useEffect(() => {
    const t = setTimeout(() => setAnimate(true), 50)
    return () => clearTimeout(t)
  }, [])

  const sorted = [...stats].sort((a, b) => (b.strength ?? 0) - (a.strength ?? 0))

  return (
    <div className="space-y-2.5">
      {sorted.map((s) => {
        const strength = Math.round(s.strength ?? 0)
        return (
          <div key={s.category} className="flex items-center gap-3">
            <span className="w-28 shrink-0 text-sm text-right text-muted-foreground truncate">
              {s.category}
            </span>
            <div className="flex-1 bg-muted rounded-full h-2 overflow-hidden">
              <div
                className={`h-2 rounded-full transition-all duration-700 ease-out ${barColor(strength)}`}
                style={{ width: animate ? `${strength}%` : '0%' }}
              />
            </div>
            <span className={`w-10 shrink-0 text-sm font-semibold tabular-nums text-right ${scoreColor(strength)}`}>
              {strength}%
            </span>
            <span className="w-16 shrink-0 text-xs text-muted-foreground text-right hidden sm:block">
              {s.problem_count} {s.problem_count === 1 ? 'problem' : 'problems'}
            </span>
          </div>
        )
      })}
    </div>
  )
}
