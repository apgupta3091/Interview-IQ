// Payments removed — UpgradeBanner disabled. All users have full access.
// Re-enable when billing is added back.

export default function UpgradeBanner() {
  return null
}

/*
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Zap } from 'lucide-react'
import { api } from '@/lib/api'

const FREE_LIMIT = 20
const WARN_AT = 18

export default function UpgradeBanner() {
  const [state, setState] = useState<{ tier: string; count: number } | null>(null)

  useEffect(() => {
    api.billing
      .getStatus()
      .then((s) => setState({ tier: s.tier, count: s.problem_count }))
      .catch(() => {})
  }, [])

  if (!state || state.tier !== 'free' || state.count < WARN_AT) return null

  const atLimit = state.count >= FREE_LIMIT

  return (
    <div className="px-6 pt-4">
      <div className="flex items-center justify-between gap-3 rounded-lg border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm">
        <div className="flex items-center gap-2 min-w-0">
          <Zap className="w-4 h-4 text-amber-500 shrink-0" />
          <span className="text-amber-700 dark:text-amber-300 truncate">
            {atLimit
              ? "You've reached the 20-problem free limit."
              : `You've logged ${state.count}/20 free problems.`}
          </span>
        </div>
        <Link
          to="/pricing"
          className="shrink-0 text-xs font-semibold text-amber-700 dark:text-amber-300 underline underline-offset-2 hover:no-underline"
        >
          Upgrade to Pro
        </Link>
      </div>
    </div>
  )
}
*/
