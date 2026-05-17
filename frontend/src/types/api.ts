// API request/response types

export interface AuthResponse {
  user: {
    id: string
    username: string
    fullName: string
    role: string
  }
  loggedIn: boolean
}

export interface PatientListResponse {
  patients: import('./models').Patient[]
  total: number
  page: number
  pageSize: number
}

export interface InvoiceListResponse {
  invoices: import('./models').Invoice[]
  total: number
  page: number
  pageSize: number
}

export interface DashboardStats {
  todayAppointments: number
  totalPatients: number
  todayRevenue: number // paise
  monthRevenue: number // paise
  totalOutstanding: number // paise
  patientsThisMonth: number
}

export interface DailyReport {
  date: string
  totalCollection: number // paise
  payments: PaymentSummary[]
}

export interface PaymentSummary {
  invoiceNumber: string
  patientName: string
  amount: number // paise
  method: string
}

export interface MonthlyReport {
  year: number
  month: number
  totalRevenue: number // paise
  totalInvoiced: number // paise
  totalOutstanding: number // paise
}

export interface SetupInput {
  clinicName: string
  doctorName: string
  doctorQualification: string
  address: string
  city: string
  state: string
  pincode: string
  phone: string
  email: string
  gstin: string
  gstEnabled: boolean
  invoicePrefix: string
  adminUsername: string
  adminPassword: string
  adminFullName: string
}

export interface CreatePatientInput {
  name: string
  phone: string
  email: string
  gender: string
  age: number
  dateOfBirth: string
  address: string
  city: string
  bloodGroup: string
  medicalHistory: string
  allergies: string
  notes: string
}

export interface CreateAppointmentInput {
  patientId: string
  date: string
  startTime: string
  endTime: string
  duration: number
  purpose: string
  notes: string
}

export interface CreateInvoiceInput {
  patientId: string
  items: InvoiceItemInput[]
  discountPercent: number
  discountAmount: number
  notes: string
}

export interface InvoiceItemInput {
  treatmentId: string
  description: string
  quantity: number
  unitPrice: number
  toothNumber: string
}

export interface RecordPaymentInput {
  invoiceId: string
  amount: number
  method: string
  paymentDate: string
  reference: string
  notes: string
}

export interface BackupInfo {
  fileName: string
  filePath: string
  size: number
  createdAt: string
}

export interface CloudDriveInfo {
  provider: 'google_drive' | 'onedrive' | 'dropbox'
  path: string
  available: boolean
}
