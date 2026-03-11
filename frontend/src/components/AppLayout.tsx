import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '@clerk/react'
import { SidebarProvider, SidebarTrigger, useSidebar } from '@/components/ui/sidebar'
import AppSidebar from '@/components/AppSidebar'
import RetryPanel from '@/components/RetryPanel'

// Must match the RetryPanel's w-56 (14rem)
const RETRY_PANEL_WIDTH = '14rem'

function AppShell() {
  const { open } = useSidebar()
  return (
    <div
      className="flex flex-1 flex-col min-h-screen transition-[margin-left] duration-200 ease-linear"
      style={{
        marginLeft: open ? 'var(--sidebar-width, 16rem)' : '0',
        marginRight: RETRY_PANEL_WIDTH,
      }}
    >
      <header className="flex h-13 items-center gap-2 border-b px-4 bg-background/80 backdrop-blur-sm sticky top-0 z-10">
        <SidebarTrigger className="text-muted-foreground hover:text-foreground" />
      </header>
      <main className="p-6">
        <Outlet />
      </main>
    </div>
  )
}

export default function AppLayout() {
  const { isSignedIn, isLoaded } = useAuth()
  if (!isLoaded) return null
  if (!isSignedIn) return <Navigate to="/login" replace />
  return (
    <SidebarProvider>
      <AppSidebar />
      <AppShell />
      <RetryPanel />
    </SidebarProvider>
  )
}
