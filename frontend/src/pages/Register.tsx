import { SignUp } from '@clerk/react'

export default function Register() {
  return (
    <div className="min-h-screen flex items-center justify-center">
      <SignUp routing="path" path="/register" fallbackRedirectUrl="/dashboard" />
    </div>
  )
}
