import type { AuthResponse, AuthRequest } from '@/types/api'
import client from './client'

export const auth = {
  login: (body: AuthRequest) =>
    client.post<AuthResponse>('/api/auth/login', body).then((r) => r.data),

  register: (body: AuthRequest) =>
    client.post<AuthResponse>('/api/auth/register', body).then((r) => r.data),
}
