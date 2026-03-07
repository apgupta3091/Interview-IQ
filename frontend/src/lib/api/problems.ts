import type { Problem, LogProblemRequest } from '@/types/api'
import client from './client'

export const problems = {
  list: () =>
    client.get<Problem[]>('/api/problems').then((r) => r.data),

  log: (body: LogProblemRequest) =>
    client.post<Problem>('/api/problems', body).then((r) => r.data),
}
