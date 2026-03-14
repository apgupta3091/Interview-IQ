import { useEffect, useState } from 'react'
import { api } from '@/lib/api'

type TierState = 'loading' | 'free' | 'pro'

export function useBillingTier(): TierState {
  const [tier, setTier] = useState<TierState>('loading')

  useEffect(() => {
    api.billing
      .getStatus()
      .then((s) => setTier(s.tier === 'pro' ? 'pro' : 'free'))
      .catch(() => setTier('free'))
  }, [])

  return tier
}
