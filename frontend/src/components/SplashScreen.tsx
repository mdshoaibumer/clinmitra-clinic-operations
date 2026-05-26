import toothLogo from '@/assets/tooth-logo.avif'

export default function SplashScreen() {
  return (
    <div className="flex flex-col items-center justify-center h-screen bg-background">
      <div className="flex flex-col items-center gap-5 animate-in">
        {/* Logo */}
        <div className="h-20 w-20 rounded-2xl bg-primary/8 flex items-center justify-center shadow-soft ring-1 ring-primary/10">
          <img
            src={toothLogo}
            alt="Clinmitra Dental"
            className="h-12 w-12 object-contain"
          />
        </div>

        {/* App Name */}
        <div className="text-center">
          <h1 className="text-2xl font-heading font-bold text-foreground tracking-tight">
            Clinmitra Dental
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            Clinic Management System
          </p>
        </div>

        {/* Loading bar */}
        <div className="w-48 h-1 bg-muted rounded-full overflow-hidden mt-3">
          <div className="h-full w-1/2 bg-primary rounded-full animate-pulse-soft" 
               style={{ animation: 'loading-bar 1.5s ease-in-out infinite' }} />
        </div>
      </div>

      {/* Footer */}
      <p className="absolute bottom-6 text-xs text-muted-foreground/60">
        v1.1.1
      </p>

      <style>{`
        @keyframes loading-bar {
          0% { transform: translateX(-100%); width: 40%; }
          50% { transform: translateX(60%); width: 60%; }
          100% { transform: translateX(200%); width: 40%; }
        }
      `}</style>
    </div>
  )
}
