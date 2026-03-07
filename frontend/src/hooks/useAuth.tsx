import { createContext, useContext, useState, useCallback } from 'react'
import type { ReactNode } from 'react'

type AuthState = {
  token: string | null
  email: string | null
}

type AuthContextValue = AuthState & {
  login: (token: string, email: string) => void
  logout: () => void
  isAuthenticated: boolean
}

const AuthContext = createContext<AuthContextValue | null>(null)

function readInitialState(): AuthState {
  return {
    token: localStorage.getItem('token'),
    email: localStorage.getItem('email'),
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>(readInitialState)

  const login = useCallback((token: string, email: string) => {
    localStorage.setItem('token', token)
    localStorage.setItem('email', email)
    setState({ token, email })
  }, [])

  const logout = useCallback(() => {
    localStorage.removeItem('token')
    localStorage.removeItem('email')
    setState({ token: null, email: null })
  }, [])

  return (
    <AuthContext.Provider
      value={{ ...state, login, logout, isAuthenticated: state.token !== null }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider')
  return ctx
}
