import axios from 'axios'

// Extend the Window interface to access the Clerk global.
declare global {
  interface Window {
    Clerk?: {
      session?: {
        getToken(): Promise<string | null>
      }
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

export default client
