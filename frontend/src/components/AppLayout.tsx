import { useState } from 'react'
import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '@clerk/react'
import { RefreshCw } from 'lucide-react'
import { SidebarProvider, SidebarTrigger, useSidebar } from '@/components/ui/sidebar'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Button } from '@/components/ui/button'
import AppSidebar from '@/components/AppSidebar'
import RetryPanel from '@/components/RetryPanel'
import UpgradeBanner from '@/components/UpgradeBanner'
import { useIsMobile } from '@/hooks/use-mobile'

// Must match the RetryPanel's w-56 (14rem)
const RETRY_PANEL_WIDTH = '14rem'

function AppShell() {
  const { open, isMobile: sidebarIsMobile } = useSidebar()
  const isMobile = useIsMobile()
  const [retryOpen, setRetryOpen] = useState(false)

  return (
    <>
      <div
        className="flex flex-1 flex-col min-h-screen transition-[margin-left] duration-200 ease-linear"
        style={{
          // On mobile the sidebar renders as an overlay Sheet — no left margin needed.
          marginLeft: (!sidebarIsMobile && open) ? 'var(--sidebar-width, 16rem)' : '0',
          marginRight: isMobile ? 0 : RETRY_PANEL_WIDTH,
        }}
      >
        <header className="flex h-13 items-center justify-between gap-2 border-b px-4 bg-background/80 backdrop-blur-sm sticky top-0 z-10">
          <SidebarTrigger className="text-muted-foreground hover:text-foreground" />
          <Button
            variant="ghost"
            size="icon"
            className="md:hidden h-8 w-8 text-muted-foreground hover:text-foreground"
            onClick={() => setRetryOpen(true)}
            aria-label="Open retry list"
          >
            <RefreshCw className="w-4 h-4" />
          </Button>
        </header>
        <UpgradeBanner />
        <main className="p-4 sm:p-6">
          <Outlet />
        </main>
      </div>

      {isMobile ? (
        <Sheet open={retryOpen} onOpenChange={setRetryOpen}>
          <SheetContent side="right" className="p-0 flex flex-col w-72">
            <SheetHeader className="px-4 py-4 border-b flex-row items-center gap-2 space-y-0">
              <RefreshCw className="w-4 h-4 text-primary" />
              <SheetTitle className="text-sm font-semibold">Retry List</SheetTitle>
            </SheetHeader>
            <RetryPanel asSheet />
          </SheetContent>
        </Sheet>
      ) : (
        <RetryPanel />
      )}
    </>
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
    </SidebarProvider>
  )
}
