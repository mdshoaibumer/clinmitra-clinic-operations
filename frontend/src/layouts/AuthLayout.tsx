import { ReactNode } from 'react'

interface AuthLayoutProps {
  children: ReactNode
}

export default function AuthLayout({ children }: AuthLayoutProps) {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-blue-100">
      <div className="w-full max-w-md px-4">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-primary">Clinmitra Dental</h1>
          <p className="text-muted-foreground mt-2">Simple, modern dental clinic management software for Indian clinics.</p>
        </div>
        {children}
      </div>
    </div>
  )
}
