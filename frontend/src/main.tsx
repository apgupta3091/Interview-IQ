import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Toaster } from '@/components/ui/sonner'
import { AuthProvider, useAuth } from '@/hooks/useAuth'
import type { ReactNode } from 'react'
import './index.css'

// Placeholder pages — filled in subsequent steps
function Login() { return <div className="p-8 text-lg font-semibold">Login</div> }
function Register() { return <div className="p-8 text-lg font-semibold">Register</div> }
function Dashboard() { return <div className="p-8 text-lg font-semibold">Dashboard</div> }
function LogProblem() { return <div className="p-8 text-lg font-semibold">Log Problem</div> }
function ProblemList() { return <div className="p-8 text-lg font-semibold">Problems</div> }

// Redirects unauthenticated users to /login.
function ProtectedRoute({ children }: { children: ReactNode }) {
  const { isAuthenticated } = useAuth()
  return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
          <Route path="/problems/new" element={<ProtectedRoute><LogProblem /></ProtectedRoute>} />
          <Route path="/problems" element={<ProtectedRoute><ProblemList /></ProtectedRoute>} />
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
        </Routes>
        <Toaster />
      </AuthProvider>
    </BrowserRouter>
  </StrictMode>,
)
