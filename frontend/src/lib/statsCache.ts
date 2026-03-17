import type { CategoryStats, WeakestResult } from '@/types/api'

// Module-level cache — survives route changes within the same session.
// Invalidated when a new problem is logged so charts reflect updated scores.
let stats: CategoryStats[] | null = null
let weakest: WeakestResult | null = null

export const statsCache = {
  get(): { stats: CategoryStats[]; weakest: WeakestResult | null } | null {
    if (stats === null) return null
    return { stats, weakest }
  },
  set(s: CategoryStats[], w: WeakestResult | null) {
    stats = s
    weakest = w
  },
  invalidate() {
    stats = null
    weakest = null
  },
}
