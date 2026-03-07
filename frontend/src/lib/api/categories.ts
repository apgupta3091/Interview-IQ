import type { CategoryStats, WeakestResult } from '@/types/api'
import client from './client'

export const categories = {
  stats: () =>
    client.get<CategoryStats[]>('/api/categories/stats').then((r) => r.data),

  weakest: () =>
    client.get<WeakestResult>('/api/categories/weakest').then((r) => r.data),
}
