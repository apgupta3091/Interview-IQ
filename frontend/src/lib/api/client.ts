import axios from 'axios'

// Extend the Window interface to access the Clerk global.
declare global {
  interface Window {
    Clerk?: {
      session?: {
        getToken(): Promise<string | null>
      }
      signOut?: () => Promise<void>
    }
  }
}

const client = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? 'http://localhost:8080',
})

// Fetch the Clerk session JWT and attach it to every request.
client.interceptors.request.use(async (config) => {
  const token = await window.Clerk?.session?.getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// On 401, sign out via Clerk and redirect to login.
client.interceptors.response.use(
  (res) => res,
  async (error) => {
    if (error.response?.status === 401) {
      await window.Clerk?.signOut?.()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default client
