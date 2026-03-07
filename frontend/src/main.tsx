import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Toaster } from '@/components/ui/sonner'
import './index.css'

// Placeholder pages — filled in subsequent steps
function Login() { return <div className="p-8 text-lg font-semibold">Login</div> }
function Register() { return <div className="p-8 text-lg font-semibold">Register</div> }
function Dashboard() { return <div className="p-8 text-lg font-semibold">Dashboard</div> }
function LogProblem() { return <div className="p-8 text-lg font-semibold">Log Problem</div> }
function ProblemList() { return <div className="p-8 text-lg font-semibold">Problems</div> }

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/problems/new" element={<LogProblem />} />
        <Route path="/problems" element={<ProblemList />} />
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
      </Routes>
      <Toaster />
    </BrowserRouter>
  </StrictMode>,
)
