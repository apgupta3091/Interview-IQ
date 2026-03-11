import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ThemeProvider } from 'next-themes'
import { ClerkProvider } from '@clerk/react'
import { Toaster } from '@/components/ui/sonner'
import AppLayout from '@/components/AppLayout'
import Login from '@/pages/Login'
import Register from '@/pages/Register'
import LogProblem from '@/pages/LogProblem'
import ProblemList from '@/pages/ProblemList'
import ProblemDetail from '@/pages/ProblemDetail'
import Dashboard from '@/pages/Dashboard'
import Recommendations from '@/pages/Recommendations'
import './index.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ClerkProvider
      publishableKey={import.meta.env.VITE_CLERK_PUBLISHABLE_KEY}
      afterSignOutUrl="/"
    >
      <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route element={<AppLayout />}>
              <Route path="/dashboard" element={<Dashboard />} />
              <Route path="/problems" element={<ProblemList />} />
              <Route path="/problems/:id" element={<ProblemDetail />} />
              <Route path="/problems/new" element={<LogProblem />} />
              <Route path="/recommendations" element={<Recommendations />} />
            </Route>
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
          </Routes>
          <Toaster />
        </BrowserRouter>
      </ThemeProvider>
    </ClerkProvider>
  </StrictMode>,
)
