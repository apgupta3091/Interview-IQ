import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom'
import { Toaster } from '@/components/ui/sonner'
import { AuthProvider, useAuth } from '@/hooks/useAuth'
import Navbar from '@/components/Navbar'
import Login from '@/pages/Login'
import Register from '@/pages/Register'
import LogProblem from '@/pages/LogProblem'
import './index.css'

// Placeholder pages — replaced in subsequent steps
function Dashboard() { return <div className="p-8 text-lg font-semibold">Dashboard</div> }
function ProblemList() { return <div className="p-8 text-lg font-semibold">Problems</div> }

// Wraps authenticated pages: checks auth, renders navbar + page content.
function AppLayout() {
  const { isAuthenticated } = useAuth()
  if (!isAuthenticated) return <Navigate to="/login" replace />
  return (
    <>
      <Navbar />
      <main className="max-w-5xl mx-auto px-4 py-8">
        <Outlet />
      </main>
    </>
  )
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route element={<AppLayout />}>
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/problems" element={<ProblemList />} />
            <Route path="/problems/new" element={<LogProblem />} />
          </Route>
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
        </Routes>
        <Toaster />
      </AuthProvider>
    </BrowserRouter>
  </StrictMode>,
)
