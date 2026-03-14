import { client } from './client'

export const billing = {
  createCheckout: async (plan: 'monthly' | 'annual'): Promise<{ url: string }> => {
    const { data } = await client.post<{ url: string }>('/api/billing/checkout', { plan })
    return data
  },

  createPortal: async (): Promise<{ url: string }> => {
    const { data } = await client.post<{ url: string }>('/api/billing/portal')
    return data
  },

  getStatus: async (): Promise<{ tier: string; problem_count: number; problem_limit: number }> => {
    const { data } = await client.get<{ tier: string; problem_count: number; problem_limit: number }>('/api/billing/status')
    return data
  },
}
