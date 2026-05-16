import { formatCurrency } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { Patient, Treatment } from '@/types/models'
import type { InvoiceItemInput } from '@/types/api'

interface InvoiceConfirmProps {
  patient: Patient | undefined
  items: (InvoiceItemInput & { key: string })[]
  discount: number
  treatments: Treatment[]
  subtotal: number
  onConfirm: () => void
  onBack: () => void
}

export default function InvoiceConfirmDialog({
  patient,
  items,
  discount,
  treatments,
  subtotal,
  onConfirm,
  onBack,
}: InvoiceConfirmProps) {
  const discountAmount = Math.round(subtotal * discount / 100)
  const total = subtotal - discountAmount

  return (
    <Card className="border-primary/30 bg-primary/5">
      <CardHeader>
        <CardTitle className="text-lg">Confirm Invoice</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Patient Info */}
        <div className="p-3 bg-white rounded-md">
          <p className="text-sm text-muted-foreground">Patient</p>
          <p className="font-medium">{patient?.name || 'Unknown'}</p>
          <p className="text-sm text-muted-foreground">{patient?.phone}</p>
        </div>

        {/* Items Summary */}
        <div className="p-3 bg-white rounded-md">
          <p className="text-sm text-muted-foreground mb-2">Items ({items.length})</p>
          <div className="space-y-1">
            {items.map((item) => {
              const treatment = treatments.find(t => t.id === item.treatmentId)
              return (
                <div key={item.key} className="flex justify-between text-sm">
                  <span>{item.description || treatment?.name || 'Custom Item'}</span>
                  <span className="font-medium">{formatCurrency(Math.round(item.unitPrice * 100) * item.quantity)}</span>
                </div>
              )
            })}
          </div>
        </div>

        {/* Totals */}
        <div className="p-3 bg-white rounded-md space-y-1 text-sm">
          <div className="flex justify-between">
            <span>Subtotal</span>
            <span>{formatCurrency(subtotal)}</span>
          </div>
          {discount > 0 && (
            <div className="flex justify-between text-red-600">
              <span>Discount ({discount}%)</span>
              <span>-{formatCurrency(discountAmount)}</span>
            </div>
          )}
          <div className="flex justify-between font-bold text-lg border-t pt-1">
            <span>Total</span>
            <span>{formatCurrency(total)}</span>
          </div>
        </div>

        {/* Actions */}
        <div className="flex gap-2 pt-2">
          <Button onClick={onConfirm} className="flex-1">
            Confirm & Create Invoice
          </Button>
          <Button variant="outline" onClick={onBack}>
            Go Back & Edit
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
