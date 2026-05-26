import { useNavigate } from 'react-router-dom'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { formatCurrency } from '@/lib/utils'
import { Users, Receipt, Calendar, IndianRupee, Plus, ArrowRight, TrendingUp } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { useToast } from '@/components/ui/use-toast'
import PageTransition from '@/components/PageTransition'
import { DashboardSkeleton } from '@/components/ui/skeletons'
import { useEffect } from 'react'

export default function Dashboard() {
  const navigate = useNavigate()
  const { toast } = useToast()

  const { data: stats, isLoading: statsLoading, error: statsError } = useQuery({
    queryKey: ['dashboardStats'],
    queryFn: () => window.go.handler.DashboardHandler.GetDashboardStats(),
  })

  const { data: todayAppointments = [], isLoading: aptLoading, error: aptError } = useQuery({
    queryKey: ['todayAppointments'],
    queryFn: () => window.go.handler.AppointmentHandler.GetTodayAppointments(),
  })

  useEffect(() => {
    if (statsError || aptError) {
      toast({
        variant: "destructive",
        title: "Error fetching dashboard data",
        description: "Please check your database connection or try again.",
      })
    }
  }, [statsError, aptError, toast])

  const loading = statsLoading || aptLoading

  if (loading) {
    return <DashboardSkeleton />
  }

  // Empty state: show onboarding guidance when clinic has no data
  const isEmpty = stats && stats.totalPatients === 0 && stats.todayAppointments === 0

  if (isEmpty) {
    return (
      <PageTransition className="space-y-6">
        <h1 className="text-2xl font-heading font-bold text-foreground">Welcome to Clinmitra Dental!</h1>

        <Card className="border-primary/20 bg-primary/5 shadow-card">
          <CardContent className="py-8">
            <div className="text-center space-y-4">
              <div className="inline-flex items-center justify-center w-16 h-16 bg-primary/10 rounded-full">
                <Users className="h-8 w-8 text-primary" />
              </div>
              <h2 className="text-xl font-heading font-semibold">Get Started in 3 Steps</h2>
              <p className="text-muted-foreground max-w-md mx-auto">
                Your clinic management system is ready. Follow these steps to start managing your practice efficiently.
              </p>
            </div>
          </CardContent>
        </Card>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card className="card-interactive cursor-pointer border-border/60" onClick={() => navigate('/patients?action=new')}>
            <CardContent className="pt-6">
              <div className="flex items-center gap-3 mb-3">
                <div className="h-8 w-8 rounded-full bg-primary/10 text-primary flex items-center justify-center text-sm font-bold">1</div>
                <h3 className="font-heading font-semibold">Register Your First Patient</h3>
              </div>
              <p className="text-sm text-muted-foreground mb-3">
                Add patient details — name, phone, and medical history. You can search them instantly later.
              </p>
              <Button size="sm" variant="outline" className="w-full">
                <Plus className="h-4 w-4 mr-2" /> Add Patient <ArrowRight className="h-4 w-4 ml-auto" />
              </Button>
            </CardContent>
          </Card>

          <Card className="card-interactive cursor-pointer border-border/60" onClick={() => navigate('/appointments?action=new')}>
            <CardContent className="pt-6">
              <div className="flex items-center gap-3 mb-3">
                <div className="h-8 w-8 rounded-full bg-accent/10 text-accent flex items-center justify-center text-sm font-bold">2</div>
                <h3 className="font-heading font-semibold">Book an Appointment</h3>
              </div>
              <p className="text-sm text-muted-foreground mb-3">
                Schedule patient visits. View all appointments at a glance on the calendar.
              </p>
              <Button size="sm" variant="outline" className="w-full">
                <Calendar className="h-4 w-4 mr-2" /> Book Appointment <ArrowRight className="h-4 w-4 ml-auto" />
              </Button>
            </CardContent>
          </Card>

          <Card className="card-interactive cursor-pointer border-border/60" onClick={() => navigate('/billing?action=new')}>
            <CardContent className="pt-6">
              <div className="flex items-center gap-3 mb-3">
                <div className="h-8 w-8 rounded-full bg-secondary/10 text-secondary flex items-center justify-center text-sm font-bold">3</div>
                <h3 className="font-heading font-semibold">Create an Invoice</h3>
              </div>
              <p className="text-sm text-muted-foreground mb-3">
                Generate professional invoices with treatments, GST, and print them instantly.
              </p>
              <Button size="sm" variant="outline" className="w-full">
                <Receipt className="h-4 w-4 mr-2" /> Create Invoice <ArrowRight className="h-4 w-4 ml-auto" />
              </Button>
            </CardContent>
          </Card>
        </div>

        <Card>
          <CardContent className="py-4">
            <p className="text-sm text-muted-foreground text-center">
              <strong>Tip:</strong> Use <kbd className="px-1.5 py-0.5 bg-muted rounded text-xs font-mono">Ctrl+K</kbd> for quick search, <kbd className="px-1.5 py-0.5 bg-muted rounded text-xs font-mono">Ctrl+N</kbd> for new patient, <kbd className="px-1.5 py-0.5 bg-muted rounded text-xs font-mono">Ctrl+B</kbd> for new invoice.
            </p>
          </CardContent>
        </Card>
      </PageTransition>
    )
  }

  return (
    <PageTransition className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-heading font-bold text-foreground">Dashboard</h1>
        <p className="text-sm text-muted-foreground">
          {new Date().toLocaleDateString('en-IN', { weekday: 'long', day: 'numeric', month: 'long' })}
        </p>
      </div>

      {/* Stats Grid — KPI Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="shadow-card border-border/60 hover:shadow-card-hover transition-shadow duration-200">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Today's Appointments</CardTitle>
            <div className="h-9 w-9 rounded-lg bg-primary/10 flex items-center justify-center">
              <Calendar className="h-4.5 w-4.5 text-primary" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-heading font-bold text-foreground">{stats?.todayAppointments || 0}</div>
            <p className="text-xs text-muted-foreground mt-1">scheduled for today</p>
          </CardContent>
        </Card>

        <Card className="shadow-card border-border/60 hover:shadow-card-hover transition-shadow duration-200">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Total Patients</CardTitle>
            <div className="h-9 w-9 rounded-lg bg-accent/10 flex items-center justify-center">
              <Users className="h-4.5 w-4.5 text-accent" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-heading font-bold text-foreground">{stats?.totalPatients || 0}</div>
            <p className="text-xs text-muted-foreground mt-1">registered patients</p>
          </CardContent>
        </Card>

        <Card className="shadow-card border-border/60 hover:shadow-card-hover transition-shadow duration-200">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Today's Revenue</CardTitle>
            <div className="h-9 w-9 rounded-lg bg-success/10 flex items-center justify-center">
              <IndianRupee className="h-4.5 w-4.5 text-success" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-heading font-bold text-foreground">{formatCurrency(stats?.todayRevenue || 0)}</div>
            <p className="text-xs text-muted-foreground mt-1">collected today</p>
          </CardContent>
        </Card>

        <Card className="shadow-card border-border/60 hover:shadow-card-hover transition-shadow duration-200">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Outstanding Dues</CardTitle>
            <div className="h-9 w-9 rounded-lg bg-destructive/10 flex items-center justify-center">
              <Receipt className="h-4.5 w-4.5 text-destructive" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-heading font-bold text-destructive">{formatCurrency(stats?.totalOutstanding || 0)}</div>
            <p className="text-xs text-muted-foreground mt-1">pending collection</p>
          </CardContent>
        </Card>
      </div>

      {/* Monthly Revenue */}
      <Card className="shadow-card border-border/60">
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-lg font-heading">This Month's Revenue</CardTitle>
          <TrendingUp className="h-5 w-5 text-success" />
        </CardHeader>
        <CardContent>
          <div className="text-3xl font-heading font-bold text-success">{formatCurrency(stats?.monthRevenue || 0)}</div>
          <p className="text-sm text-muted-foreground mt-1">
            {new Date().toLocaleDateString('en-IN', { month: 'long', year: 'numeric' })}
          </p>
        </CardContent>
      </Card>

      {/* Today's Appointments */}
      <Card className="shadow-card border-border/60">
        <CardHeader>
          <CardTitle className="text-lg font-heading">Today's Appointments</CardTitle>
        </CardHeader>
        <CardContent>
          {todayAppointments.length === 0 ? (
            <div className="text-center py-8">
              <Calendar className="h-10 w-10 text-muted-foreground/40 mx-auto mb-3" />
              <p className="text-muted-foreground text-sm">No appointments scheduled for today.</p>
            </div>
          ) : (
            <div className="space-y-2">
              {todayAppointments.map((apt) => (
                <div key={apt.id} className="flex items-center justify-between p-3 bg-muted/40 rounded-lg border border-border/40 hover:bg-muted/60 transition-colors duration-150">
                  <div>
                    <p className="font-medium text-foreground">{apt.patient?.name || 'Unknown Patient'}</p>
                    <p className="text-sm text-muted-foreground">{apt.purpose || 'General checkup'}</p>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-medium text-foreground">{apt.startTime} - {apt.endTime}</p>
                    <p className="text-xs text-muted-foreground capitalize">{apt.status}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </PageTransition>
  )
}
