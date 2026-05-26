import { ReactNode } from 'react'
import toothLogo from '@/assets/tooth-logo.avif'

interface AuthLayoutProps {
  children: ReactNode
}

export default function AuthLayout({ children }: AuthLayoutProps) {
  return (
    <div className="min-h-screen flex items-center justify-center bg-background relative overflow-hidden">
      {/* Subtle background pattern */}
      <div className="absolute inset-0 opacity-30">
        <div className="absolute top-0 left-0 w-96 h-96 bg-primary/5 rounded-full -translate-x-1/2 -translate-y-1/2" />
        <div className="absolute bottom-0 right-0 w-80 h-80 bg-accent/5 rounded-full translate-x-1/3 translate-y-1/3" />
      </div>

      <div className="w-full max-w-md px-4 relative z-10">
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center h-16 w-16 rounded-2xl bg-white shadow-card mb-4 ring-1 ring-border">
            <img src={toothLogo} alt="Clinmitra Dental" className="h-10 w-10 object-contain" />
          </div>
          <h1 className="text-3xl font-heading font-bold text-foreground">Clinmitra Dental</h1>
          <p className="text-muted-foreground mt-2 text-sm">
            Simple, modern clinic management for Indian dental practices.
          </p>
        </div>
        {children}
      </div>
    </div>
  )
}
