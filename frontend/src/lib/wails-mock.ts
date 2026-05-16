/**
 * Wails Mock Layer (Persistent)
 */

if (typeof window !== 'undefined') {
  // @ts-ignore
  window.go = window.go || {};
  // @ts-ignore
  window.go.handler = window.go.handler || {};

  const getSession = (key, def) => {
    const val = sessionStorage.getItem(key);
    return val ? JSON.parse(val) : def;
  };
  const setSession = (key, val) => sessionStorage.setItem(key, JSON.stringify(val));

  const handlers = {
    AuthHandler: {
      Login: async (username, password) => {
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
      ChangePassword: async () => null,
    },
    SettingsHandler: {
      GetClinicSettings: async () => ({
        clinicName: 'Clinmitra Test Clinic',
        doctorName: 'Dr. Test',
        phone: '9876543210',
        gstEnabled: true,
        invoicePrefix: 'TEST',
        gstRate: 18,
      }),
      UpdateClinicSettings: async (settings) => settings,
      CompleteSetup: async (data) => {
        if (data.phone === '9999999999') throw new Error('[DUPLICATE] Database Error: Unique constraint failed');
        setSession('_setupComplete', true);
        return null;
      },
      IsSetupComplete: async () => getSession('_setupComplete', false),
      ListTreatments: async () => [
        { id: 't-1', name: 'Root Canal', defaultPrice: 500000 },
        { id: 't-2', name: 'Cleaning', defaultPrice: 100000 },
      ],
    },
    PatientHandler: {
      CreatePatient: async (input) => {
        const patients = getSession('_patients', []);
        const newPatient = { id: `p-${Date.now()}`, ...input };
        setSession('_patients', [...patients, newPatient]);
        return newPatient;
      },
      ListPatients: async (page, pageSize, search) => {
        let patients = getSession('_patients', []);
        if (search) {
          patients = patients.filter(p => p.name.includes(search) || p.phone.includes(search));
        }
        // Add some dummy ones if empty to test search
        if (patients.length === 0 && !search) {
          patients = [{ id: 'p-fixed', name: 'Fixed Patient', phone: '1234567890' }];
        }
        return { patients, total: patients.length, page, pageSize };
      },
      GetPatient: async (id) => ({ id, name: 'John Doe', phone: '9876543210' }),
    },
    InvoiceHandler: {
      CreateInvoice: async (input) => ({
        id: 'inv-123',
        invoiceNumber: 'TEST-2605-0001',
        totalAmount: 100000,
        balanceAmount: 100000,
        status: 'issued',
        ...input,
      }),
      ListInvoices: async () => ({ invoices: [], total: 0 }),
      RecordPayment: async () => ({ id: 'pay-123' }),
    },
    AppointmentHandler: {
      BookAppointment: async () => ({ id: 'appt-123' }),
      GetTodayAppointments: async () => [],
    },
    DashboardHandler: {
      GetDashboardStats: async () => ({
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
