import { useEffect, useState } from 'react'
import { formatCurrency, formatDate, getTodayDate } from '@/lib/utils'
import { exportToCSV, formatCurrencyForExport } from '@/lib/exportCSV'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { DatePicker, MonthPicker } from '@/components/ui/date-picker'
import { Download } from 'lucide-react'
import type { DailyReport, MonthlyReport } from '@/types/api'

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
      setDailyReport(result)
    } catch {
      setDailyReport(null)
    }
    setIsLoading(false)
  }

  const fetchMonthlyReport = async () => {
    setIsLoading(true)
    try {
      const [yearStr, monthStr] = selectedMonth.split('-')
      const year = parseInt(yearStr, 10)
      const month = parseInt(monthStr, 10)
      const result = await window.go.handler.DashboardHandler.GetMonthlyReport(year, month)
      setMonthlyReport(result)
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
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Reports</h1>
        {reportType === 'daily' && dailyReport && dailyReport.payments.length > 0 && (
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              exportToCSV(
                dailyReport.payments.map(p => ({
                  invoiceNumber: p.invoiceNumber,
                  patientName: p.patientName,
                  method: p.method,
                  amount: formatCurrencyForExport(p.amount),
                })),
                `daily-report-${dailyReport.date}`,
                [
                  { key: 'invoiceNumber', header: 'Invoice #' },
                  { key: 'patientName', header: 'Patient' },
                  { key: 'method', header: 'Method' },
                  { key: 'amount', header: 'Amount' },
                ]
              )
            }}
          >
            <Download className="h-4 w-4 mr-2" /> Export CSV
          </Button>
        )}
        {reportType === 'monthly' && monthlyReport && (
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              exportToCSV(
                [{
                  month: `${monthlyReport.year}-${String(monthlyReport.month).padStart(2, '0')}`,
                  totalRevenue: formatCurrencyForExport(monthlyReport.totalRevenue),
                  totalInvoiced: formatCurrencyForExport(monthlyReport.totalInvoiced),
                  outstanding: formatCurrencyForExport(monthlyReport.totalOutstanding),
                }],
                `monthly-report-${monthlyReport.year}-${String(monthlyReport.month).padStart(2, '0')}`,
                [
                  { key: 'month', header: 'Month' },
                  { key: 'totalRevenue', header: 'Total Revenue' },
                  { key: 'totalInvoiced', header: 'Total Invoiced' },
                  { key: 'outstanding', header: 'Outstanding' },
                ]
              )
            }}
          >
            <Download className="h-4 w-4 mr-2" /> Export CSV
          </Button>
        )}
      </div>

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
              <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">Total Payments</CardTitle></CardHeader>
              <CardContent><p className="text-2xl font-bold">{dailyReport.payments.length}</p></CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">Total Collection</CardTitle></CardHeader>
              <CardContent><p className="text-2xl font-bold text-green-600">{formatCurrency(dailyReport.totalCollection)}</p></CardContent>
            </Card>
          </div>

          {dailyReport.payments.length > 0 && (
            <Card>
              <CardHeader><CardTitle className="text-lg">Payment Details</CardTitle></CardHeader>
              <CardContent>
                <table className="w-full">
                  <thead className="border-b">
                    <tr>
                      <th className="pb-2 text-left text-sm font-medium">Invoice</th>
                      <th className="pb-2 text-left text-sm font-medium">Patient</th>
                      <th className="pb-2 text-left text-sm font-medium">Method</th>
                      <th className="pb-2 text-right text-sm font-medium">Amount</th>
                    </tr>
                  </thead>
                  <tbody>
                    {dailyReport.payments.map((payment, idx) => (
                      <tr key={idx} className="border-b">
                        <td className="py-2 text-sm font-mono">{payment.invoiceNumber}</td>
                        <td className="py-2 text-sm">{payment.patientName}</td>
                        <td className="py-2 text-sm capitalize">{payment.method}</td>
                        <td className="py-2 text-sm text-right">{formatCurrency(payment.amount)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </CardContent>
            </Card>
          )}
        </div>
      )}

      {/* Monthly Report */}
      {reportType === 'monthly' && monthlyReport && (
        <div className="space-y-4">
          <h2 className="text-lg font-semibold">Monthly Report — {monthlyReport.year}/{String(monthlyReport.month).padStart(2, '0')}</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Card>
              <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">Total Revenue</CardTitle></CardHeader>
              <CardContent><p className="text-2xl font-bold">{formatCurrency(monthlyReport.totalRevenue)}</p></CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">Total Invoiced</CardTitle></CardHeader>
              <CardContent><p className="text-2xl font-bold">{formatCurrency(monthlyReport.totalInvoiced)}</p></CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">Outstanding</CardTitle></CardHeader>
              <CardContent><p className="text-2xl font-bold text-red-600">{formatCurrency(monthlyReport.totalOutstanding)}</p></CardContent>
            </Card>
          </div>
        </div>
      )}
    </div>
  )
}
