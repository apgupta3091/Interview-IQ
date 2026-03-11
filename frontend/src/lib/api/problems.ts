import type { Problem, LogProblemRequest, ProblemListResponse } from '@/types/api'
import client from './client'

export type { ProblemListResponse }

export type ProblemListParams = {
  q?: string
  category?: string[]
  difficulty?: string[]
  score_min?: number
  score_max?: number
  from?: string
  to?: string
  limit?: number
  offset?: number
}

export const problems = {
  listFiltered: (params: ProblemListParams = {}) => {
    const sp = new URLSearchParams()
    if (params.q) sp.set('q', params.q)
    params.category?.forEach((c) => sp.append('category', c))
    params.difficulty?.forEach((d) => sp.append('difficulty', d))
    if (params.score_min !== undefined) sp.set('score_min', String(params.score_min))
    if (params.score_max !== undefined) sp.set('score_max', String(params.score_max))
    if (params.from) sp.set('from', params.from)
    if (params.to) sp.set('to', params.to)
    if (params.limit !== undefined) sp.set('limit', String(params.limit))
    if (params.offset !== undefined) sp.set('offset', String(params.offset))
    return client
      .get<ProblemListResponse>(`/api/problems?${sp.toString()}`)
      .then((r) => r.data)
  },

  log: (body: LogProblemRequest) =>
    client.post<Problem>('/api/problems', body).then((r) => r.data),

  getById: (id: number): Promise<Problem> =>
    client.get<Problem>(`/api/problems/${id}`).then((r) => r.data),
}
