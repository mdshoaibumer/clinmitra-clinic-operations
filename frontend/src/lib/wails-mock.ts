/**
 * Wails Mock Layer (Persistent)
 * This file populates window.go with mock implementations of the Go handlers.
 * It allows the frontend to run in a standard browser for E2E testing.
 */

export {};

if (typeof window !== 'undefined') {
  // @ts-ignore
  window.go = window.go || {};
  // @ts-ignore
  window.go.handler = window.go.handler || {};

  const getSession = (key: string, def: any) => {
    const val = sessionStorage.getItem(key);
    return val ? JSON.parse(val) : def;
  };
  const setSession = (key: string, val: any) => sessionStorage.setItem(key, JSON.stringify(val));

  const handlers = {
    AuthHandler: {
      Login: async (username: string, password: string): Promise<any> => {
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
      Logout: async (): Promise<null> => {
        sessionStorage.removeItem('_auth');
        return null;
      },
      GetCurrentUser: async (): Promise<any> => getSession('_auth', { loggedIn: false }),
      ChangePassword: async (): Promise<null> => null,
    },
    SettingsHandler: {
      GetClinicSettings: async (): Promise<any> => ({
        clinicName: 'Clinmitra Test Clinic',
        doctorName: 'Dr. Test',
        phone: '9876543210',
        gstEnabled: true,
        invoicePrefix: 'TEST',
        gstRate: 18,
      }),
      UpdateClinicSettings: async (settings: any): Promise<any> => settings,
      CompleteSetup: async (data: any): Promise<null> => {
        if (data.phone === '9999999999') throw new Error('[DUPLICATE] Database Error: Unique constraint failed');
        setSession('_setupComplete', true);
        return null;
      },
      IsSetupComplete: async (): Promise<boolean> => getSession('_setupComplete', false),
      ListTreatments: async (): Promise<any[]> => [
        { id: 't-1', name: 'Root Canal', defaultPrice: 500000 },
        { id: 't-2', name: 'Cleaning', defaultPrice: 100000 },
      ],
    },
    PatientHandler: {
      CreatePatient: async (input: any): Promise<any> => {
        const patients = getSession('_patients', []);
        const newPatient = { id: `p-${Date.now()}`, ...input };
        setSession('_patients', [...patients, newPatient]);
        return newPatient;
      },
      ListPatients: async (page: number, pageSize: number, search: string): Promise<any> => {
        let patients = getSession('_patients', []);
        if (search) {
          patients = patients.filter((p: any) => p.name.includes(search) || p.phone.includes(search));
        }
        if (patients.length === 0 && !search) {
          patients = [{ id: 'p-fixed', name: 'Fixed Patient', phone: '1234567890' }];
        }
        return { patients, total: patients.length, page, pageSize };
      },
      GetPatient: async (id: string): Promise<any> => ({ id, name: 'John Doe', phone: '9876543210' }),
    },
    InvoiceHandler: {
      CreateInvoice: async (input: any): Promise<any> => ({
        id: 'inv-123',
        invoiceNumber: 'TEST-2605-0001',
        totalAmount: 100000,
        balanceAmount: 100000,
        status: 'issued',
        ...input,
      }),
      ListInvoices: async (): Promise<any> => ({ invoices: [], total: 0 }),
      RecordPayment: async (): Promise<any> => ({ id: 'pay-123' }),
    },
    AppointmentHandler: {
      BookAppointment: async (): Promise<any> => ({ id: 'appt-123' }),
      GetTodayAppointments: async (): Promise<any[]> => [],
    },
    DashboardHandler: {
      GetDashboardStats: async (): Promise<any> => ({
        todayAppointments: 5,
        totalPatients: 100,
        todayRevenue: 500000,
        monthRevenue: 15000000,
        totalOutstanding: 200000,
      }),
    },
  };

  // @ts-ignore
  window.go.handler = handlers;
  
  console.log('Wails Mock Layer Initialized');
}
