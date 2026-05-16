import { formatCurrency, formatDate } from '@/lib/utils'
import { amountToWords } from '@/lib/numberToWords'
import type { Invoice, ClinicSettings } from '@/types/models'
import toothLogo from '@/assets/tooth-logo.avif'

interface InvoicePrintViewProps {
  invoice: Invoice
  clinic: ClinicSettings | null
}

export default function InvoicePrintView({ invoice, clinic }: InvoicePrintViewProps) {
  return (
    <div className="print:block hidden w-full print-container" aria-hidden="true">
      {/* 
        MyBillBook Style Dental Invoice Format
        Replicated from: https://mybillbook.in/s/bill-format/dental-clinic/
      */}
      <div className="max-w-[210mm] mx-auto p-4 font-sans text-[11px] text-black bg-white border-[2px] border-black">
        
        {/* Header: Clinic Info and Logo/Branding */}
        <div className="flex border-b-[2px] border-black h-28">
          <div className="flex-1 p-3 flex flex-col justify-center">
            <p className="font-bold text-xl">{clinic?.clinicName?.toUpperCase() || 'CLINIC NAME'}</p>
            <p className="font-semibold text-sm text-gray-700">
              {clinic?.doctorName?.toUpperCase().startsWith('DR.') ? clinic.doctorName.toUpperCase() : `DR. ${clinic?.doctorName?.toUpperCase() || 'CLINIC HEAD'}`}
            </p>
            {clinic?.doctorQualification && (
              <p className="text-[9px] font-bold text-gray-600 tracking-wider">
                {clinic.doctorQualification.toUpperCase()}
              </p>
            )}
            <p className="leading-tight mt-1 text-[10px]">{clinic?.address || 'Clinic Address'}</p>
            <div className="flex gap-4 mt-1 font-semibold text-[10px]">
              <p>Phone No.: {clinic?.phone}</p>
              <p>Email: {clinic?.email}</p>
            </div>
          </div>
          <div className="w-1/4 p-2 flex items-center justify-center border-l-[2px] border-black">
            {clinic?.logoBase64 ? (
              <img src={clinic.logoBase64} alt="Logo" className="max-h-24 max-w-full object-contain grayscale" />
            ) : (
              <div className="flex flex-col items-center">
                <img src={toothLogo} alt="Dental Care" className="w-20 h-20 object-contain opacity-40 grayscale" />
                <span className="text-[8px] font-black text-black uppercase tracking-[1px] mt-1">Dental Care</span>
              </div>
            )}
          </div>
        </div>

        {/* Section: Patient Details */}
        <div className="bg-gray-200 border-b-[2px] border-black px-3 py-0.5 font-bold uppercase tracking-wider text-[10px]">
          Patient Details:
        </div>
        <div className="flex border-b-[2px] border-black">
          <div className="w-1/2 p-3 border-r-[2px] border-black space-y-1.5">
            <div className="flex"><span className="w-20 font-bold">Name:</span> <span className="uppercase">{invoice.patient?.name}</span></div>
            <div className="flex"><span className="w-20 font-bold">Address:</span> <span>{invoice.patient?.city || '-'}</span></div>
            <div className="flex mt-4"><span className="w-20 font-bold">Phone No.:</span> <span>{invoice.patient?.phone}</span></div>
            <div className="flex"><span className="w-20 font-bold">Email ID:</span> <span>{invoice.patient?.email || '-'}</span></div>
          </div>
          <div className="w-1/2 p-3 space-y-1.5">
            <div className="flex"><span className="w-28 font-bold">Gender:</span> <span>{invoice.patient?.gender || '-'}</span></div>
            <div className="flex"><span className="w-28 font-bold">Age:</span> <span>{invoice.patient?.age || '-'}</span></div>
            <div className="flex"><span className="w-28 font-bold">Invoice No:</span> <span className="font-bold underline">{invoice.invoiceNumber}</span></div>
            <div className="flex"><span className="w-28 font-bold">Date:</span> <span>{formatDate(invoice.invoiceDate)}</span></div>
            <div className="flex"><span className="w-28 font-bold">Next Consultancy:</span> <span>-</span></div>
          </div>
        </div>

        {/* Section: Patient Observation */}
        <div className="bg-gray-200 border-b-[2px] border-black px-3 py-0.5 font-bold uppercase tracking-wider text-[10px]">
          Patient Observation:
        </div>
        <div className="h-20 border-b-[2px] border-black p-2 text-gray-400 italic">
          Enter clinical observations, symptoms, or treatment plan notes...
        </div>

        {/* Treatment Table */}
        <table className="w-full border-collapse">
          <thead>
            <tr className="bg-gray-200 border-b-[2px] border-black text-[10px] font-bold">
              <th className="border-r border-black p-1 text-center w-10">SR.</th>
              <th className="border-r border-black p-1 text-left">Description</th>
              <th className="border-r border-black p-1 text-right w-24">Price</th>
              <th className="p-1 text-right w-24">Amount</th>
            </tr>
          </thead>
          <tbody>
            {(invoice.items || []).map((item, i) => (
              <tr key={item.id} className="border-b border-black">
                <td className="border-r border-black p-2 text-center text-[10px]">{i + 1}</td>
                <td className="border-r border-black p-2 font-bold uppercase text-[10px]">{item.description}</td>
                <td className="border-r border-black p-2 text-right">{formatCurrency(item.unitPrice)}</td>
                <td className="p-2 text-right font-bold">{formatCurrency(item.amount)}</td>
              </tr>
            ))}
            {/* Empty rows to match style */}
            {[...Array(Math.max(0, 4 - (invoice.items?.length || 0)))].map((_, i) => (
              <tr key={`empty-${i}`} className="border-b border-black h-8">
                <td className="border-r border-black"></td>
                <td className="border-r border-black"></td>
                <td className="border-r border-black"></td>
                <td></td>
              </tr>
            ))}
            <tr className="bg-gray-200 border-b-[2px] border-black font-black">
              <td className="border-r border-black p-1 text-center" colSpan={3}>Total</td>
              <td className="p-1 text-right">{formatCurrency(invoice.subTotal)}</td>
            </tr>
          </tbody>
        </table>

        {/* Bottom Section */}
        <div className="flex min-h-[160px]">
          <div className="w-2/3 border-r-[2px] border-black flex flex-col">
            <div className="p-3 border-b border-black flex-1">
              <p className="font-bold text-[9px] mb-1">Total Amount In Words:</p>
              <p className="font-bold uppercase italic text-[11px] leading-tight">
                {amountToWords(invoice.totalAmount)}
              </p>
            </div>
            <div className="bg-gray-200 border-b border-black px-3 py-0.5 font-bold text-[9px]">
              Payment Info:
            </div>
            <div className="p-2 text-[9px] space-y-0.5 flex-1">
              <p><span className="font-bold w-20 inline-block">Account No.:</span> {clinic?.bankAccount || '-'}</p>
              <p><span className="font-bold w-20 inline-block">Account Name:</span> {clinic?.accountName || clinic?.clinicName}</p>
              <p><span className="font-bold w-20 inline-block">Bank Name:</span> {clinic?.bankName || '-'}</p>
              <p><span className="font-bold w-20 inline-block">IFSC/Bank Code:</span> {clinic?.ifscCode || '-'}</p>
              <p><span className="font-bold w-20 inline-block">UPI ID:</span> {clinic?.upiId || '-'}</p>
            </div>
            <div className="bg-gray-200 border-y border-black px-3 py-0.5 font-bold text-[9px]">
              Terms and Conditions:
            </div>
            <div className="p-2 text-[8px] leading-tight text-gray-600">
              <p>1. This is a computer-generated tax invoice.</p>
              <p>2. Fees once paid are non-refundable.</p>
              <p>3. Please bring this invoice for your follow-up visits.</p>
            </div>
          </div>
          <div className="w-1/3 flex flex-col">
            <div className="p-2 space-y-1.5 border-b border-black flex-1">
              <div className="flex justify-between font-bold"><span>Sub Total:</span> <span>{formatCurrency(invoice.subTotal)}</span></div>
              <div className="flex justify-between"><span>Discount:</span> <span>{invoice.discountAmount > 0 ? `-${formatCurrency(invoice.discountAmount)}` : '0.00'}</span></div>
              <div className="flex justify-between text-[9px]"><span>Tax Rate:</span> <span>{clinic?.gstEnabled ? `${clinic.gstRate}%` : '0%'}</span></div>
              <div className="flex justify-between text-[9px]"><span>CGST:</span> <span>{formatCurrency(invoice.cgstAmount)}</span></div>
              <div className="flex justify-between text-[9px]"><span>SGST:</span> <span>{formatCurrency(invoice.sgstAmount)}</span></div>
              <div className="flex justify-between font-black text-sm border-t-2 border-black pt-2 bg-gray-100 p-1">
                <span>Total Amount:</span>
                <span>{formatCurrency(invoice.totalAmount)}</span>
              </div>
            </div>
            <div className="h-20 flex flex-col items-center justify-end p-2 pb-1">
              <div className="w-full border-t border-black text-center font-bold text-[9px] pt-1">
                Clinic Seal & Signature
              </div>
            </div>
          </div>
        </div>

        {/* Product Branding Footer */}
        <div className="text-center py-1 mt-1 border-t border-black">
          <p className="text-[8px] font-bold uppercase tracking-[3px] text-gray-800">
            Powered by ClinMitra Dental — Smart Clinic Management
          </p>
          <p className="text-[6px] text-gray-500 italic mt-0.5">
            Excellence in Clinic Digitization — www.clinmitra.in
          </p>
        </div>
      </div>
    </div>
  )
}
