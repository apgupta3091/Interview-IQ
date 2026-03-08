import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '@clerk/react'
import { SidebarInset, SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'
import AppSidebar from '@/components/AppSidebar'

export default function AppLayout() {
  const { isSignedIn, isLoaded } = useAuth()
  if (!isLoaded) return null
  if (!isSignedIn) return <Navigate to="/login" replace />
  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-13 items-center gap-2 border-b px-4 bg-background/80 backdrop-blur-sm sticky top-0 z-10">
          <SidebarTrigger className="text-muted-foreground hover:text-foreground" />
        </header>
        <div className="p-6 w-full">
          <Outlet />
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
