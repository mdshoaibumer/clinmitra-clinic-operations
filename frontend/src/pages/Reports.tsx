import { useEffect, useState } from 'react'
import { formatCurrency, formatDate, getTodayDate } from '@/lib/utils'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { DatePicker, MonthPicker } from '@/components/ui/date-picker'

interface DailyReport {
  date: string
  totalInvoices: number
  totalAmount: number
  totalCollected: number
  cashAmount: number
  upiAmount: number
  cardAmount: number
}

interface MonthlyReport {
  month: string
  totalInvoices: number
  totalRevenue: number
  totalCollected: number
  outstandingAmount: number
  newPatients: number
  totalAppointments: number
}

export default function Reports() {
  const [reportType, setReportType] = useState<'daily' | 'monthly'>('daily')
  const [selectedDate, setSelectedDate] = useState(getTodayDate())
  const [selectedMonth, setSelectedMonth] = useState(getTodayDate().substring(0, 7))
  const [dailyReport, setDailyReport] = useState<DailyReport | null>(null)
  const [monthlyReport, setMonthlyReport] = useState<MonthlyReport | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const fetchDailyReport = async () => {
    setIsLoading(true)
    try {
      const result = await window.go.handler.DashboardHandler.GetDailyReport(selectedDate)
      setDailyReport(result as DailyReport)
    } catch {
      setDailyReport(null)
    }
    setIsLoading(false)
  }

  const fetchMonthlyReport = async () => {
    setIsLoading(true)
    try {
      const result = await window.go.handler.DashboardHandler.GetMonthlyReport(selectedMonth)
      setMonthlyReport(result as MonthlyReport)
    } catch {
      setMonthlyReport(null)
    }
    setIsLoading(false)
  }

  useEffect(() => {
    if (reportType === 'daily') fetchDailyReport()
    else fetchMonthlyReport()
  }, [reportType, selectedDate, selectedMonth])

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Reports</h1>

      {/* Report Type Tabs */}
      <div className="flex gap-2">
        <Button variant={reportType === 'daily' ? 'default' : 'outline'} onClick={() => setReportType('daily')}>
          Daily Report
        </Button>
        <Button variant={reportType === 'monthly' ? 'default' : 'outline'} onClick={() => setReportType('monthly')}>
          Monthly Report
        </Button>
      </div>

      {/* Date Selector */}
      <div className="flex gap-4 items-end">
        {reportType === 'daily' ? (
          <div className="space-y-2">
            <Label>Date</Label>
            <DatePicker value={selectedDate} onChange={(date) => setSelectedDate(date)} />
          </div>
        ) : (
          <div className="space-y-2">
            <Label>Month</Label>
            <MonthPicker value={selectedMonth} onChange={(month) => setSelectedMonth(month)} />
          </div>
        )}
      </div>

      {isLoading && <p className="text-muted-foreground">Loading report...</p>}

      {/* Daily Report */}
      {reportType === 'daily' && dailyReport && (
        <div className="space-y-4">
          <h2 className="text-lg font-semibold">Daily Collection Report — {formatDate(dailyReport.date)}</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Card>
              <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">Total Invoices</CardTitle></CardHeader>
              <CardContent><p className="text-2xl font-bold">{dailyReport.totalInvoices}</p></CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">Billed Amount</CardTitle></CardHeader>
              <CardContent><p className="text-2xl font-bold">{formatCurrency(dailyReport.totalAmount)}</p></CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">Collected</CardTitle></CardHeader>
              <CardContent><p className="text-2xl font-bold text-green-600">{formatCurrency(dailyReport.totalCollected)}</p></CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader><CardTitle className="text-lg">Collection by Method</CardTitle></CardHeader>
            <CardContent>
              <div className="space-y-3">
                <div className="flex justify-between items-center">
                  <span className="text-sm">Cash</span>
                  <span className="font-medium">{formatCurrency(dailyReport.cashAmount)}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm">UPI</span>
                  <span className="font-medium">{formatCurrency(dailyReport.upiAmount)}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm">Card</span>
                  <span className="font-medium">{formatCurrency(dailyReport.cardAmount)}</span>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Monthly Report */}
      {reportType === 'monthly' && monthlyReport && (
        <div className="space-y-4">
          <h2 className="text-lg font-semibold">Monthly Report — {monthlyReport.month}</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Card>
              <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">Total Revenue</CardTitle></CardHeader>
              <CardContent><p className="text-2xl font-bold">{formatCurrency(monthlyReport.totalRevenue)}</p></CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">Total Collected</CardTitle></CardHeader>
              <CardContent><p className="text-2xl font-bold text-green-600">{formatCurrency(monthlyReport.totalCollected)}</p></CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">Outstanding</CardTitle></CardHeader>
              <CardContent><p className="text-2xl font-bold text-red-600">{formatCurrency(monthlyReport.outstandingAmount)}</p></CardContent>
            </Card>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Card>
              <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">Total Invoices</CardTitle></CardHeader>
              <CardContent><p className="text-2xl font-bold">{monthlyReport.totalInvoices}</p></CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">New Patients</CardTitle></CardHeader>
              <CardContent><p className="text-2xl font-bold">{monthlyReport.newPatients}</p></CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">Appointments</CardTitle></CardHeader>
              <CardContent><p className="text-2xl font-bold">{monthlyReport.totalAppointments}</p></CardContent>
            </Card>
          </div>
        </div>
      )}
    </div>
  )
}
