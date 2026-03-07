import { Link, useLocation, useNavigate } from 'react-router-dom'
import { LayoutDashboard, ListChecks, PlusCircle, LogOut, Sun, Moon } from 'lucide-react'
import { useTheme } from 'next-themes'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar'
import { useAuth } from '@/hooks/useAuth'

const NAV_ITEMS = [
  { to: '/dashboard',    label: 'Dashboard',   icon: LayoutDashboard },
  { to: '/problems',     label: 'Problems',    icon: ListChecks },
  { to: '/problems/new', label: 'Log Problem', icon: PlusCircle },
]

export default function AppSidebar() {
  const { pathname } = useLocation()
  const { logout, email } = useAuth()
  const navigate = useNavigate()
  const { theme, setTheme } = useTheme()

  function handleLogout() {
    logout()
    navigate('/login', { replace: true })
  }

  return (
    <Sidebar>
      <SidebarHeader className="px-4 py-3">
        <span className="font-semibold tracking-tight text-sm">Interview IQ</span>
        {email && <p className="text-xs text-muted-foreground truncate">{email}</p>}
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Navigation</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {NAV_ITEMS.map(({ to, label, icon: Icon }) => (
                <SidebarMenuItem key={to}>
                  <SidebarMenuButton asChild isActive={pathname === to}>
                    <Link to={to}>
                      <Icon />
                      <span>{label}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter className="p-2">
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}>
              {theme === 'dark' ? <Sun /> : <Moon />}
              <span>{theme === 'dark' ? 'Light mode' : 'Dark mode'}</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem>
            <SidebarMenuButton onClick={handleLogout}>
              <LogOut />
              <span>Sign out</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
  )
}
