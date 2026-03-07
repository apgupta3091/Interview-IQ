import type { LeetCodeProblemSuggestion } from '@/types/api'
import client from './client'

export const leetcodeProblems = {
  search: (q: string, limit = 10) =>
    client
      .get<LeetCodeProblemSuggestion[]>(
        `/api/leetcode-problems/search?q=${encodeURIComponent(q)}&limit=${limit}`,
      )
      .then((r) => r.data),
}
