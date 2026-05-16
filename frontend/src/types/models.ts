// Models matching Go backend structures

export interface User {
  id: string
  username: string
  fullName: string
  role: 'admin' | 'doctor' | 'receptionist'
  lastLoginAt?: string
}

export interface ClinicSettings {
  id: string
  clinicName: string
  doctorName: string
  address: string
  city: string
  state: string
  pincode: string
  phone: string
  email: string
  gstin: string
  gstEnabled: boolean
  gstRate: number
  invoicePrefix: string
  logoPath: string
  logoBase64: string
  setupComplete: boolean
  autoBackup: boolean
  backupPath: string
}

export interface Patient {
  id: string
  name: string
  phone: string
  email: string
  gender: 'male' | 'female' | 'other'
  age: number
  dateOfBirth: string
  address: string
  city: string
  bloodGroup: string
  medicalHistory: string
  allergies: string
  notes: string
  createdBy: string
  createdAt: string
  updatedAt: string
}

export interface Treatment {
  id: string
  name: string
  code: string
  defaultPrice: number // paise
  category: string
  description: string
  isActive: boolean
}

export interface Appointment {
  id: string
  patientId: string
  appointmentDate: string
  startTime: string
  endTime: string
  duration: number
  status: 'scheduled' | 'completed' | 'cancelled' | 'no_show'
  purpose: string
  notes: string
  cancelReason: string
  createdBy: string
  createdAt: string
  patient?: Patient
}

export interface Invoice {
  id: string
  invoiceNumber: string
  patientId: string
  invoiceDate: string
  subTotal: number
  discountAmount: number
  discountPercent: number
  taxableAmount: number
  cgstAmount: number
  sgstAmount: number
  totalAmount: number
  paidAmount: number
  balanceAmount: number
  status: 'issued' | 'partial' | 'paid' | 'void'
  notes: string
  voidReason: string
  createdBy: string
  createdAt: string
  patient?: Patient
  items?: InvoiceItem[]
  payments?: Payment[]
}

export interface InvoiceItem {
  id: string
  invoiceId: string
  treatmentId: string
  description: string
  quantity: number
  unitPrice: number // paise
  amount: number // paise
  toothNumber: string
  treatment?: Treatment
}

export interface Payment {
  id: string
  invoiceId: string
  amount: number // paise
  method: 'cash' | 'upi' | 'card' | 'bank_transfer' | 'other'
  paymentDate: string
  reference: string
  notes: string
  receivedBy: string
  createdAt: string
}

export interface PatientTreatment {
  id: string
  patientId: string
  treatmentId: string
  invoiceId: string
  treatmentDate: string
  toothNumber: string
  notes: string
  performedBy: string
  treatment?: Treatment
}

export interface AuditLog {
  id: string
  userId: string
  action: string
  entityType: string
  entityId: string
  oldValue: string
  newValue: string
  createdAt: string
}

export interface BackupInfo {
  fileName: string
  filePath: string
  size: number
  createdAt: string
}
