import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ThemeProvider } from 'next-themes'
import { ClerkProvider } from '@clerk/react'
import * as Sentry from '@sentry/react'
import posthog from 'posthog-js'
import { Toaster } from '@/components/ui/sonner'
import AppLayout from '@/components/AppLayout'
import Login from '@/pages/Login'
import Register from '@/pages/Register'
import LogProblem from '@/pages/LogProblem'
import ProblemList from '@/pages/ProblemList'
import ProblemDetail from '@/pages/ProblemDetail'
import Dashboard from '@/pages/Dashboard'
import Recommendations from '@/pages/Recommendations'
// import Pricing from '@/pages/Pricing'
// Payments removed — Pricing page disabled. Re-enable when billing is added back.
import PrivacyPolicy from '@/pages/PrivacyPolicy'
import Terms from '@/pages/Terms'
import './index.css'

Sentry.init({
  dsn: import.meta.env.VITE_SENTRY_DSN,
  environment: import.meta.env.MODE,
  integrations: [Sentry.browserTracingIntegration()],
  tracesSampleRate: 0.1,
})

if (import.meta.env.VITE_POSTHOG_KEY) {
  posthog.init(import.meta.env.VITE_POSTHOG_KEY, {
    api_host: import.meta.env.VITE_POSTHOG_HOST ?? 'https://app.posthog.com',
    person_profiles: 'identified_only',
  })
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Sentry.ErrorBoundary fallback={<p>Something went wrong.</p>}>
      <ClerkProvider
        publishableKey={import.meta.env.VITE_CLERK_PUBLISHABLE_KEY}
        afterSignOutUrl="/"
      >
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          <BrowserRouter>
            <Routes>
              <Route path="/login" element={<Login />} />
              <Route path="/register" element={<Register />} />
              <Route path="/privacy" element={<PrivacyPolicy />} />
              <Route path="/terms" element={<Terms />} />
              <Route element={<AppLayout />}>
                <Route path="/dashboard" element={<Dashboard />} />
                <Route path="/problems" element={<ProblemList />} />
                <Route path="/problems/:id" element={<ProblemDetail />} />
                <Route path="/problems/new" element={<LogProblem />} />
                <Route path="/recommendations" element={<Recommendations />} />
                {/* Payments removed — pricing route disabled. Re-enable when billing is added back. */}
                {/* <Route path="/pricing" element={<Pricing />} /> */}
              </Route>
              <Route path="/" element={<Navigate to="/dashboard" replace />} />
            </Routes>
            <Toaster />
          </BrowserRouter>
        </ThemeProvider>
      </ClerkProvider>
    </Sentry.ErrorBoundary>
  </StrictMode>,
)
