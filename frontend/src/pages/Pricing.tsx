import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { ArrowLeft, Check, Zap } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { api } from '@/lib/api'

const FREE_FEATURES = [
  'Log up to 20 problems',
  'Score decay tracking',
  'Category radar chart',
  'Weakest category detection',
]

const PRO_FEATURES = [
  'Unlimited problem logging',
  'AI-powered study recommendations',
  'Advanced search & filters',
  'Export to CSV / JSON',
  'Streak tracking',
  'Priority support',
]

export default function Pricing() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState<'monthly' | 'annual' | null>(null)

  async function upgrade(plan: 'monthly' | 'annual') {
    setLoading(plan)
    try {
      const { url } = await api.billing.createCheckout(plan)
      window.location.href = url
    } catch {
      toast.error('Failed to start checkout — please try again.')
    } finally {
      setLoading(null)
    }
  }

  return (
    <div className="max-w-3xl mx-auto animate-fade-up space-y-8">
      <div className="space-y-1">
        <Button
          variant="ghost"
          size="sm"
          className="-ml-2 text-muted-foreground"
          onClick={() => navigate(-1)}
        >
          <ArrowLeft className="w-4 h-4 mr-1" />
          Back
        </Button>
        <h1 className="text-2xl font-bold tracking-tight">Upgrade to Pro</h1>
        <p className="text-sm text-muted-foreground">
          Unlimited logging and AI-powered study recommendations.
        </p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        {/* Free tier */}
        <div className="rounded-xl border border-border/60 p-6 space-y-4">
          <div>
            <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Free</p>
            <p className="text-3xl font-bold mt-1">$0</p>
          </div>
          <ul className="space-y-2">
            {FREE_FEATURES.map((f) => (
              <li key={f} className="flex items-center gap-2 text-sm text-muted-foreground">
                <Check className="w-4 h-4 shrink-0 text-muted-foreground/50" />
                {f}
              </li>
            ))}
          </ul>
          <Button variant="outline" className="w-full" disabled>
            Current plan
          </Button>
        </div>

        {/* Pro tier */}
        <div className="rounded-xl border border-primary/40 bg-primary/5 p-6 space-y-4 relative">
          <div className="absolute top-3 right-3">
            <span className="text-[10px] font-semibold bg-primary text-primary-foreground px-2 py-0.5 rounded-full uppercase tracking-wide">
              Popular
            </span>
          </div>
          <div>
            <p className="text-xs font-medium text-primary uppercase tracking-wider flex items-center gap-1">
              <Zap className="w-3 h-3" /> Pro
            </p>
            <div className="mt-1">
              <span className="text-3xl font-bold">$7</span>
              <span className="text-sm text-muted-foreground">/month</span>
            </div>
            <p className="text-xs text-muted-foreground mt-0.5">or $60/year — save 29%</p>
          </div>
          <ul className="space-y-2">
            {PRO_FEATURES.map((f) => (
              <li key={f} className="flex items-center gap-2 text-sm">
                <Check className="w-4 h-4 shrink-0 text-primary" />
                {f}
              </li>
            ))}
          </ul>
          <div className="space-y-2 pt-1">
            <Button
              className="w-full"
              onClick={() => upgrade('monthly')}
              disabled={loading !== null}
            >
              {loading === 'monthly' ? 'Loading…' : 'Subscribe monthly — $7/mo'}
            </Button>
            <Button
              variant="outline"
              className="w-full"
              onClick={() => upgrade('annual')}
              disabled={loading !== null}
            >
              {loading === 'annual' ? 'Loading…' : 'Subscribe annually — $60/yr'}
            </Button>
          </div>
        </div>
      </div>

      <p className="text-xs text-center text-muted-foreground">
        Secure payment via Stripe. Cancel anytime from your{' '}
        <Link to="/dashboard" className="underline underline-offset-2">
          dashboard
        </Link>
        .
      </p>
    </div>
  )
}
