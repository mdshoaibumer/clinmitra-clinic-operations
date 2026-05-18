import toothLogo from '@/assets/tooth-logo.avif'

export default function SplashScreen() {
  return (
    <div className="flex flex-col items-center justify-center h-screen bg-gradient-to-br from-blue-50 via-white to-blue-50">
      <div className="flex flex-col items-center gap-6 animate-in fade-in duration-500">
        {/* Logo */}
        <div className="relative">
          <div className="h-20 w-20 rounded-2xl bg-primary/10 flex items-center justify-center shadow-sm">
            <img
              src={toothLogo}
              alt="Clinmitra Dental"
              className="h-12 w-12 object-contain"
            />
          </div>
        </div>

        {/* App Name */}
        <div className="text-center">
          <h1 className="text-2xl font-bold text-primary tracking-tight">
            Clinmitra Dental
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            Clinic Management System
          </p>
        </div>

        {/* Loading Indicator */}
        <div className="flex items-center gap-1.5 mt-4">
          <div className="h-1.5 w-1.5 rounded-full bg-primary/60 animate-bounce [animation-delay:0ms]" />
          <div className="h-1.5 w-1.5 rounded-full bg-primary/60 animate-bounce [animation-delay:150ms]" />
          <div className="h-1.5 w-1.5 rounded-full bg-primary/60 animate-bounce [animation-delay:300ms]" />
        </div>
      </div>

      {/* Footer */}
      <p className="absolute bottom-6 text-xs text-muted-foreground/60">
        v1.0.0
      </p>
    </div>
  )
}
