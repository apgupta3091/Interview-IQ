import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom'
import { ThemeProvider } from 'next-themes'
import { Toaster } from '@/components/ui/sonner'
import { AuthProvider, useAuth } from '@/hooks/useAuth'
import { SidebarProvider, SidebarTrigger, useSidebar } from '@/components/ui/sidebar'
import AppSidebar from '@/components/AppSidebar'
import Login from '@/pages/Login'
import Register from '@/pages/Register'
import LogProblem from '@/pages/LogProblem'
import ProblemList from '@/pages/ProblemList'
import Dashboard from '@/pages/Dashboard'
import './index.css'

// Inner shell — must be inside SidebarProvider to access useSidebar.
function AppShell() {
  const { open } = useSidebar()
  return (
    <div
      className="flex flex-1 flex-col min-h-screen transition-[margin-left] duration-200 ease-linear"
      style={{ marginLeft: open ? 'var(--sidebar-width, 16rem)' : '0' }}
    >
      <header className="flex h-13 items-center gap-2 border-b px-4 bg-background/80 backdrop-blur-sm sticky top-0 z-10">
        <SidebarTrigger className="text-muted-foreground hover:text-foreground" />
      </header>
      <main className="p-6 max-w-6xl">
        <Outlet />
      </main>
    </div>
  )
}

// Wraps authenticated pages: checks auth, renders sidebar + page content.
function AppLayout() {
  const { isAuthenticated } = useAuth()
  if (!isAuthenticated) return <Navigate to="/login" replace />
  return (
    <SidebarProvider>
      <AppSidebar />
      <AppShell />
    </SidebarProvider>
  )
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
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
    </ThemeProvider>
  </StrictMode>,
)
