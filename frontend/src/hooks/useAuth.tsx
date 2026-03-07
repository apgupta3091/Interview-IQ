import { createContext, useContext, useState, useCallback } from 'react';
import type { ReactNode } from 'react';

type AuthState = {
  token: string | null;
  userId: number | null;
  email: string | null;
};

type AuthContextValue = AuthState & {
  login: (token: string, userId: number, email: string) => void;
  logout: () => void;
  isAuthenticated: boolean;
};

const AuthContext = createContext<AuthContextValue | null>(null);

function readInitialState(): AuthState {
  const token = localStorage.getItem('token');
  const userId = localStorage.getItem('userId');
  const email = localStorage.getItem('email');
  return {
    token,
    userId: userId ? Number(userId) : null,
    email,
  };
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>(readInitialState);

  const login = useCallback((token: string, userId: number, email: string) => {
    localStorage.setItem('token', token);
    localStorage.setItem('userId', String(userId));
    localStorage.setItem('email', email);
    setState({ token, userId, email });
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('token');
    localStorage.removeItem('userId');
    localStorage.removeItem('email');
    setState({ token: null, userId: null, email: null });
  }, []);

  return (
    <AuthContext.Provider
      value={{ ...state, login, logout, isAuthenticated: state.token !== null }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider');
  return ctx;
}
