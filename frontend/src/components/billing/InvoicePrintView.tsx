import { formatCurrency, formatDate } from '@/lib/utils'
import type { Invoice } from '@/types/models'
import type { ClinicSettings } from '@/types/models'

interface InvoicePrintViewProps {
  invoice: Invoice
  clinic: ClinicSettings | null
}

export default function InvoicePrintView({ invoice, clinic }: InvoicePrintViewProps) {
  return (
    <div className="hidden print:block print:p-0 print:m-0" aria-hidden="true" data-print-only="true">
      <div className="max-w-[210mm] mx-auto p-8 font-sans text-sm">
        {/* Clinic Header */}
        <div className="text-center border-b-2 border-black pb-4 mb-4">
          {clinic?.logoBase64 && (
            <div className="flex justify-center mb-2">
              <img
                src={clinic.logoBase64}
                alt="Clinic Logo"
                className="max-h-16 max-w-48 object-contain"
              />
            </div>
          )}
          <h1 className="text-xl font-bold uppercase">{clinic?.clinicName || 'Clinic'}</h1>
          {clinic?.doctorName && <p className="text-sm">Dr. {clinic.doctorName}</p>}
          {clinic?.address && <p className="text-xs text-gray-600">{clinic.address}</p>}
          <div className="flex justify-center gap-4 text-xs text-gray-600 mt-1">
            {clinic?.phone && <span>Ph: {clinic.phone}</span>}
            {clinic?.email && <span>Email: {clinic.email}</span>}
          </div>
          {clinic?.gstin && clinic.gstEnabled && (
            <p className="text-xs mt-1">GSTIN: {clinic.gstin}</p>
          )}
        </div>

        {/* Invoice Title */}
        <div className="text-center mb-4">
          <h2 className="text-lg font-bold">TAX INVOICE</h2>
        </div>

        {/* Invoice Details Row */}
        <div className="flex justify-between mb-4 text-sm">
          <div>
            <p><span className="font-medium">Invoice No:</span> {invoice.invoiceNumber}</p>
            <p><span className="font-medium">Date:</span> {formatDate(invoice.invoiceDate)}</p>
          </div>
          <div className="text-right">
            <p><span className="font-medium">Patient:</span> {invoice.patient?.name}</p>
            <p><span className="font-medium">Phone:</span> {invoice.patient?.phone}</p>
          </div>
        </div>

        {/* Line Items Table */}
        <table className="w-full border-collapse border border-gray-400 mb-4 text-sm">
          <thead>
            <tr className="bg-gray-100">
              <th className="border border-gray-400 px-2 py-1 text-left w-8">#</th>
              <th className="border border-gray-400 px-2 py-1 text-left">Description</th>
              <th className="border border-gray-400 px-2 py-1 text-center w-12">Tooth</th>
              <th className="border border-gray-400 px-2 py-1 text-center w-12">Qty</th>
              <th className="border border-gray-400 px-2 py-1 text-right w-24">Rate</th>
              <th className="border border-gray-400 px-2 py-1 text-right w-24">Amount</th>
            </tr>
          </thead>
          <tbody>
            {invoice.items?.map((item, idx) => (
              <tr key={item.id}>
                <td className="border border-gray-400 px-2 py-1">{idx + 1}</td>
                <td className="border border-gray-400 px-2 py-1">{item.description}</td>
                <td className="border border-gray-400 px-2 py-1 text-center">{item.toothNumber || '-'}</td>
                <td className="border border-gray-400 px-2 py-1 text-center">{item.quantity}</td>
                <td className="border border-gray-400 px-2 py-1 text-right">{formatCurrency(item.unitPrice)}</td>
                <td className="border border-gray-400 px-2 py-1 text-right">{formatCurrency(item.amount)}</td>
              </tr>
            ))}
          </tbody>
        </table>

        {/* Totals */}
        <div className="flex justify-end">
          <div className="w-64 text-sm">
            <div className="flex justify-between py-1">
              <span>Subtotal:</span>
              <span>{formatCurrency(invoice.subTotal)}</span>
            </div>
            {invoice.discountAmount > 0 && (
              <div className="flex justify-between py-1">
                <span>Discount ({invoice.discountPercent}%):</span>
                <span>-{formatCurrency(invoice.discountAmount)}</span>
              </div>
            )}
            {invoice.cgstAmount > 0 && (
              <>
                <div className="flex justify-between py-1">
                  <span>CGST:</span>
                  <span>{formatCurrency(invoice.cgstAmount)}</span>
                </div>
                <div className="flex justify-between py-1">
                  <span>SGST:</span>
                  <span>{formatCurrency(invoice.sgstAmount)}</span>
                </div>
              </>
            )}
            <div className="flex justify-between py-1 border-t-2 border-black font-bold text-base">
              <span>Total:</span>
              <span>{formatCurrency(invoice.totalAmount)}</span>
            </div>
            {invoice.paidAmount > 0 && (
              <div className="flex justify-between py-1 text-green-700">
                <span>Paid:</span>
                <span>{formatCurrency(invoice.paidAmount)}</span>
              </div>
            )}
            {invoice.balanceAmount > 0 && (
              <div className="flex justify-between py-1 font-medium">
                <span>Balance Due:</span>
                <span>{formatCurrency(invoice.balanceAmount)}</span>
              </div>
            )}
          </div>
        </div>

        {/* Payment Info */}
        {invoice.payments && invoice.payments.length > 0 && (
          <div className="mt-4 pt-4 border-t text-xs">
            <p className="font-medium mb-1">Payment History:</p>
            {invoice.payments.map((p) => (
              <p key={p.id}>
                {formatDate(p.paymentDate)} — {formatCurrency(p.amount)} via {p.method}
                {p.reference && ` (Ref: ${p.reference})`}
              </p>
            ))}
          </div>
        )}

        {/* Footer */}
        <div className="mt-8 pt-4 border-t text-xs text-center text-gray-500">
          <p>Thank you for visiting {clinic?.clinicName || 'our clinic'}!</p>
          <p className="mt-1">This is a computer-generated invoice.</p>
        </div>
      </div>
    </div>
  )
}
