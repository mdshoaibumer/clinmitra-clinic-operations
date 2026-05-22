/**
 * Wails Mock Layer (Persistent)
 */

if (typeof window !== 'undefined') {
  const w = window as Window & { go: Record<string, unknown> };
  w.go = w.go || {};

  const getSession = (key: string, def: unknown): unknown => {
    const val = sessionStorage.getItem(key);
    return val ? JSON.parse(val) : def;
  };
  const setSession = (key: string, val: unknown) => sessionStorage.setItem(key, JSON.stringify(val));

  const handlers = {
    AuthHandler: {
      Login: async (username: string, password: string) => {
        if (username === 'admin' && password === 'password123') {
          const res = {
            user: { id: 'admin-1', username: 'admin', fullName: 'System Admin', role: 'admin' },
            loggedIn: true,
          };
          setSession('_auth', res);
          return res;
        }
        if (username === 'sql-inj') {
          throw new Error('[INTERNAL_ERROR] SQL Exception near "OR"');
        }
        throw new Error('[UNAUTHORIZED] Invalid credentials');
      },
      Logout: async () => {
        sessionStorage.removeItem('_auth');
        return null;
      },
      GetCurrentUser: async () => getSession('_auth', { loggedIn: false }),
      ChangePassword: async (oldPassword: string, _newPassword: string) => {
        if (oldPassword !== 'password123') throw new Error('[UNAUTHORIZED] Current password is incorrect');
        return null;
      },
    },
    SettingsHandler: {
      GetClinicSettings: async () => {
        const logo = getSession('_clinicLogo', '');
        return {
          id: 'clinic-1',
          clinicName: 'Clinmitra Test Clinic',
          doctorName: 'Dr. Test',
          doctorQualification: 'BDS, MDS',
          phone: '9876543210',
          address: '123 Test Street',
          city: 'Mumbai',
          state: 'Maharashtra',
          pincode: '400001',
          email: 'test@clinic.com',
          gstEnabled: true,
          gstin: '27AAAAA0000A1Z5',
          invoicePrefix: 'TEST',
          gstRate: 18,
          logoBase64: logo,
          logoPath: '',
          setupComplete: true,
          autoBackup: true,
          backupPath: './backups',
          bankAccount: '1234567890',
          accountName: 'Clinmitra Test Clinic',
          bankName: 'Test Bank',
          ifscCode: 'TEST0001234',
          upiId: 'test@upi',
        };
      },
      UpdateClinicSettings: async (settings: unknown) => settings,
      UploadLogo: async (base64Data: string) => {
        setSession('_clinicLogo', base64Data);
        return null;
      },
      RemoveLogo: async () => {
        setSession('_clinicLogo', '');
        return null;
      },
      CompleteSetup: async (data: { phone?: string }) => {
        if (data.phone === '9999999999') throw new Error('[DUPLICATE] Database Error: Unique constraint failed');
        setSession('_setupComplete', true);
        return null;
      },
      IsSetupComplete: async () => getSession('_setupComplete', false),
      ListTreatments: async () => {
        const treatments = getSession('_treatments', [
          { id: 't-1', name: 'Root Canal', code: 'RC', category: 'Endodontics', defaultPrice: 500000 },
          { id: 't-2', name: 'Cleaning', code: 'CLN', category: 'Preventive', defaultPrice: 100000 },
        ]);
        return treatments;
      },
      CreateTreatment: async (name: string, code: string, category: string, _desc: string, price: number) => {
        const treatments = getSession('_treatments', [
          { id: 't-1', name: 'Root Canal', code: 'RC', category: 'Endodontics', defaultPrice: 500000 },
          { id: 't-2', name: 'Cleaning', code: 'CLN', category: 'Preventive', defaultPrice: 100000 },
        ]) as Array<Record<string, unknown>>;
        const newTreatment = { id: `t-${Date.now()}`, name, code, category, defaultPrice: price, isActive: true };
        setSession('_treatments', [...treatments, newTreatment]);
        return newTreatment;
      },
      UpdateTreatment: async (id: string, name: string, code: string, category: string, _desc: string, price: number) => {
        const treatments = getSession('_treatments', []) as Array<Record<string, unknown>>;
        const updated = treatments.map(t => t.id === id ? { ...t, name, code, category, defaultPrice: price } : t);
        setSession('_treatments', updated);
        return null;
      },
      DeleteTreatment: async (id: string) => {
        const treatments = getSession('_treatments', [
          { id: 't-1', name: 'Root Canal', code: 'RC', category: 'Endodontics', defaultPrice: 500000 },
          { id: 't-2', name: 'Cleaning', code: 'CLN', category: 'Preventive', defaultPrice: 100000 },
        ]) as Array<Record<string, unknown>>;
        setSession('_treatments', treatments.filter(t => t.id !== id));
        return null;
      },
    },
    PatientHandler: {
      CreatePatient: async (input: { name: string; phone: string }) => {
        const patients = getSession('_patients', []) as Array<Record<string, unknown>>;
        const newPatient = { id: `p-${Date.now()}`, ...input };
        setSession('_patients', [...patients, newPatient]);
        return newPatient;
      },
      ListPatients: async (_page: number, _pageSize: number, search: string) => {
        let patients = getSession('_patients', []) as Array<Record<string, string>>;
        if (search) {
          patients = patients.filter(p => p.name.includes(search) || p.phone.includes(search));
        }
        // Add some dummy ones if empty to test search
        if (patients.length === 0 && !search) {
          patients = [{ id: 'p-fixed', name: 'Fixed Patient', phone: '1234567890' }];
        }
        return { patients, total: patients.length, page: _page, pageSize: _pageSize };
      },
      GetPatient: async (id: string) => ({
        id,
        name: 'John Doe',
        phone: '9876543210',
        email: 'john@example.com',
        gender: 'male',
        age: 35,
        bloodGroup: 'O+',
        city: 'Mumbai',
        address: '123 Main St',
        medicalHistory: 'None',
        allergies: 'None',
        notes: '',
        createdAt: '2025-01-01T00:00:00Z',
      }),
      GetPatientHistory: async (_patientId: string) => [
        {
          id: 'pt-1',
          treatmentDate: '2025-03-15',
          toothNumber: '14',
          treatment: { name: 'Root Canal', defaultPrice: 500000 },
        },
        {
          id: 'pt-2',
          treatmentDate: '2025-02-10',
          toothNumber: '',
          treatment: { name: 'Cleaning', defaultPrice: 100000 },
        },
      ],
      UpdatePatient: async (id: string, input: Record<string, unknown>) => ({ id, ...input }),
      DeletePatient: async (_id: string) => null,
    },
    InvoiceHandler: {
      CreateInvoice: async (input: Record<string, unknown>) => ({
        id: 'inv-123',
        invoiceNumber: 'TEST-2605-0001',
        totalAmount: 100000,
        balanceAmount: 100000,
        status: 'issued',
        ...input,
      }),
      ListInvoices: async (_page: number, _pageSize: number, _status: string) => ({
        invoices: [
          {
            id: 'inv-list-1',
            invoiceNumber: 'TEST-2605-0001',
            invoiceDate: '2025-05-16',
            totalAmount: 500000,
            balanceAmount: 200000,
            status: 'partial',
            patient: { name: 'John Doe' },
          },
          {
            id: 'inv-list-2',
            invoiceNumber: 'TEST-2605-0002',
            invoiceDate: '2025-05-15',
            totalAmount: 100000,
            balanceAmount: 0,
            status: 'paid',
            patient: { name: 'Jane Smith' },
          },
        ],
        total: 2,
      }),
      GetInvoice: async (id: string) => ({
        id,
        invoiceNumber: 'TEST-2605-0001',
        invoiceDate: '2025-05-16',
        subTotal: 500000,
        discountAmount: 0,
        cgstAmount: 0,
        sgstAmount: 0,
        totalAmount: 500000,
        paidAmount: 300000,
        balanceAmount: 200000,
        status: 'partial',
        patient: { name: 'John Doe', phone: '9876543210' },
        items: [
          { id: 'item-1', description: 'Root Canal', quantity: 1, unitPrice: 500000, amount: 500000, toothNumber: '14' },
        ],
        payments: [
          { id: 'pay-1', amount: 300000, method: 'cash', paymentDate: '2025-05-16', reference: '' },
        ],
      }),
      RecordPayment: async (input: Record<string, unknown>) => {
        return { id: `pay-${Date.now()}`, ...input };
      },
      VoidInvoice: async (_id: string, _reason: string) => null,
    },
    AppointmentHandler: {
      BookAppointment: async () => ({ id: 'appt-123' }),
      GetTodayAppointments: async () => [
        {
          id: 'appt-today-1',
          patient: { name: 'John Doe' },
          purpose: 'Root Canal Follow-up',
          startTime: '09:00',
          endTime: '09:30',
          duration: 30,
          status: 'scheduled',
          notes: '',
        },
        {
          id: 'appt-today-2',
          patient: { name: 'Jane Smith' },
          purpose: 'Cleaning',
          startTime: '10:00',
          endTime: '10:30',
          duration: 30,
          status: 'completed',
          notes: '',
        },
      ],
      GetAppointmentsByDate: async (_date: string) => {
        const appointments = getSession('_appointments', [
          {
            id: 'appt-1',
            patient: { name: 'John Doe' },
            purpose: 'Root Canal',
            startTime: '09:00',
            endTime: '09:30',
            duration: 30,
            status: 'scheduled',
            notes: 'First visit',
          },
        ]) as Array<Record<string, unknown>>;
        return appointments;
      },
      CreateAppointment: async (input: Record<string, unknown>) => {
        const appointments = getSession('_appointments', [
          {
            id: 'appt-1',
            patient: { name: 'John Doe' },
            purpose: 'Root Canal',
            startTime: '09:00',
            endTime: '09:30',
            duration: 30,
            status: 'scheduled',
            notes: 'First visit',
          },
        ]) as Array<Record<string, unknown>>;
        const patient = (getSession('_patients', [{ id: 'p-fixed', name: 'Fixed Patient', phone: '1234567890' }]) as Array<Record<string, string>>)
          .find(p => p.id === input.patientId) || { name: 'Selected Patient' };
        const newAppt = {
          id: `appt-${Date.now()}`,
          patient,
          purpose: input.purpose || 'General',
          startTime: input.startTime,
          endTime: input.endTime,
          duration: input.duration || 30,
          status: 'scheduled',
          notes: input.notes || '',
        };
        setSession('_appointments', [...appointments, newAppt]);
        return newAppt;
      },
      CancelAppointment: async (id: string, _reason: string) => {
        const defaultAppts = [
          {
            id: 'appt-1',
            patient: { name: 'John Doe' },
            purpose: 'Root Canal',
            startTime: '09:00',
            endTime: '09:30',
            duration: 30,
            status: 'scheduled',
            notes: 'First visit',
          },
        ];
        const appointments = getSession('_appointments', defaultAppts) as Array<Record<string, unknown>>;
        setSession('_appointments', appointments.map(a => a.id === id ? { ...a, status: 'cancelled' } : a));
        return null;
      },
      CompleteAppointment: async (id: string) => {
        const defaultAppts = [
          {
            id: 'appt-1',
            patient: { name: 'John Doe' },
            purpose: 'Root Canal',
            startTime: '09:00',
            endTime: '09:30',
            duration: 30,
            status: 'scheduled',
            notes: 'First visit',
          },
        ];
        const appointments = getSession('_appointments', defaultAppts) as Array<Record<string, unknown>>;
        setSession('_appointments', appointments.map(a => a.id === id ? { ...a, status: 'completed' } : a));
        return null;
      },
      GetWeekAppointments: async () => [],
    },
    DashboardHandler: {
      GetDashboardStats: async () => ({
        todayAppointments: 5,
        totalPatients: 100,
        todayRevenue: 500000,
        monthRevenue: 15000000,
        totalOutstanding: 200000,
      }),
      GetDailyReport: async (date: string) => ({
        date,
        totalCollection: 350000,
        payments: [
          { invoiceNumber: 'TEST-0001', patientName: 'John Doe', amount: 200000, method: 'cash' },
          { invoiceNumber: 'TEST-0002', patientName: 'Jane Smith', amount: 150000, method: 'upi' },
        ],
      }),
      GetMonthlyReport: async (year: number, month: number) => ({
        year,
        month,
        totalRevenue: 15000000,
        totalInvoiced: 18000000,
        totalOutstanding: 3000000,
      }),
    },
    BackupHandler: {
      ListBackups: async () => [
        { fileName: 'clinmitra_backup_20250516_180000.db', filePath: '/backups/clinmitra_backup_20250516_180000.db', size: 1024000, createdAt: '2025-05-16T18:00:00Z' },
        { fileName: 'clinmitra_backup_20250515_180000.db', filePath: '/backups/clinmitra_backup_20250515_180000.db', size: 1020000, createdAt: '2025-05-15T18:00:00Z' },
      ],
      CreateBackup: async () => ({ fileName: 'clinmitra_backup_20250516_manual.db', filePath: '/backups/clinmitra_backup_20250516_manual.db', size: 1024000, createdAt: new Date().toISOString() }),
      RestoreFromBackup: async (_filename: string) => null,
      VerifyBackup: async (_filePath: string) => true,
      GetAutoBackupPath: async () => './backups',
      DetectCloudDrives: async () => [
        { provider: 'google_drive', path: 'G:\\My Drive', available: true },
        { provider: 'onedrive', path: 'C:\\Users\\User\\OneDrive', available: true },
      ],
      CreateCloudBackup: async () => ({ fileName: 'clinmitra_backup_cloud.db', filePath: 'G:\\My Drive\\ClinMitra Backups\\clinmitra_backup_cloud.db', size: 1024000, createdAt: new Date().toISOString() }),
    },
    WhatsAppHandler: {
      GetWhatsAppTemplates: async () => ({
        welcomeTemplate: 'Hello {{patient_name}}! Welcome to {{clinic_name}}. Dr. {{doctor_name}} and our team look forward to caring for you. Call us at {{clinic_phone}}.',
        invoiceTemplate: 'Hi {{patient_name}}, your invoice {{invoice_number}} dated {{invoice_date}}: Total ₹{{total_amount}}, Paid ₹{{paid_amount}}, Balance ₹{{balance_amount}}. Method: {{payment_method}}. Thank you!',
        enabled: true,
      }),
      PrepareWelcomeMessage: async (patient: { name: string; phone: string }) => ({
        phone: '91' + patient.phone.replace(/\D/g, '').slice(-10),
        message: `Hello ${patient.name}! Welcome to Mock Clinic. Dr. Mock and our team look forward to caring for you.`,
        whatsAppUrl: `whatsapp://send?phone=91${patient.phone.replace(/\D/g, '').slice(-10)}&text=Hello+${encodeURIComponent(patient.name)}`,
        webUrl: `https://wa.me/91${patient.phone.replace(/\D/g, '').slice(-10)}?text=Hello+${encodeURIComponent(patient.name)}`,
        isDesktopPresent: false,
      }),
      PrepareInvoiceMessage: async (invoiceID: string, paymentMethod: string) => ({
        phone: '919876543210',
        message: `Hi Patient, invoice ${invoiceID}: Total ₹1000.00. Method: ${paymentMethod}. Thank you!`,
        whatsAppUrl: `whatsapp://send?phone=919876543210&text=Invoice`,
        webUrl: `https://wa.me/919876543210?text=Invoice`,
        isDesktopPresent: false,
      }),
      IsWhatsAppInstalled: async () => false,
      SendViaWhatsApp: async (_url: string) => { console.log('Mock: Would open WhatsApp URL'); },
    },
  };

  (w.go as Record<string, unknown>).handler = handlers;
  console.log('Wails Mock Layer Initialized');
}

export {}
