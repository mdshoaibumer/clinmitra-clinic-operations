# Practivo — Foundation Architecture Blueprint

## SECTION 1 — PRODUCT ARCHITECTURE

### Overall System Architecture

Practivo is a single-process desktop application built on the Wails v2 framework. The Go backend and React frontend run as a unified binary. The backend owns all business logic, data access, and file I/O. The frontend is a presentation layer rendered in a native webview.

```
┌─────────────────────────────────────────────────────────────────┐
│                        Practivo.exe                              │
│                                                                 │
│  ┌───────────────────────────┐  ┌────────────────────────────┐  │
│  │       Go Backend          │  │     React Frontend         │  │
│  │                           │  │                            │  │
│  │  ┌─────────────────────┐  │  │  ┌──────────────────────┐  │  │
│  │  │   Handler Layer     │◄─┼──┼──│   Wails Bindings     │  │  │
│  │  │   (Wails Bound)     │──┼──┼─►│   (TypeScript)       │  │  │
│  │  └────────┬────────────┘  │  │  └──────────────────────┘  │  │
│  │           │               │  │                            │  │
│  │  ┌────────▼────────────┐  │  │  ┌──────────────────────┐  │  │
│  │  │   Service Layer     │  │  │  │   Zustand Stores     │  │  │
│  │  │   (Business Logic)  │  │  │  │   (UI State)         │  │  │
│  │  └────────┬────────────┘  │  │  └──────────────────────┘  │  │
│  │           │               │  │                            │  │
│  │  ┌────────▼────────────┐  │  │  ┌──────────────────────┐  │  │
│  │  │   Repository Layer  │  │  │  │   React Components   │  │  │
│  │  │   (Data Access)     │  │  │  │   (Presentation)     │  │  │
│  │  └────────┬────────────┘  │  │  └──────────────────────┘  │  │
│  │           │               │  │                            │  │
│  │  ┌────────▼────────────┐  │  └────────────────────────────┘  │
│  │  │   SQLite Database   │  │                                  │
│  │  │   (Local File)      │  │                                  │
│  │  └─────────────────────┘  │                                  │
│  └───────────────────────────┘                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Why Offline-First

Indian dental clinics, particularly single-doctor practices, operate in environments with unreliable internet. Power cuts, ISP outages, and slow connections are normal. A cloud-dependent system would be unusable during patient visits. Offline-first guarantees:

- Zero downtime from network issues
- Instant responsiveness (no API latency)
- Full data ownership (no vendor lock-in)
- No recurring SaaS costs for budget-conscious clinics
- HIPAA/DISHA compliance simplified (data never leaves premises)

### Desktop Architecture Decisions

| Decision | Rationale |
|----------|-----------|
| Wails v2 over Electron | 10x smaller binary (~15MB vs 150MB+), native performance, Go backend |
| Single binary distribution | No runtime dependencies, simple install |
| SQLite over PostgreSQL | Zero-config, single-file DB, perfect for single-user/few-user scenarios |
| Go backend | Type safety, compiled binary, excellent concurrency for backup operations |
| React frontend | Largest ecosystem, best component libraries for complex UIs |

### Backend/Frontend Communication in Wails

Wails binds Go struct methods directly to the frontend as TypeScript functions. No HTTP server, no REST API, no WebSocket layer.

```
Frontend calls:    window.go.main.PatientHandler.GetPatient(id)
Backend receives:  func (h *PatientHandler) GetPatient(id string) (*PatientResponse, error)
```

Communication is synchronous from the frontend's perspective (async/await). Wails generates TypeScript bindings automatically from Go struct method signatures. This eliminates API versioning, serialization bugs, and network error handling.

Events flow backend → frontend via Wails runtime events for real-time notifications (appointment reminders, backup progress).

### Local Database Strategy

- Single SQLite file: `%APPDATA%/Practivo/Practivo.db`
- WAL mode enabled for concurrent read performance
- Foreign keys enforced at database level
- All migrations versioned and applied on startup
- Database file encrypted at rest using SQLCipher (future enhancement; MVP uses OS-level file permissions)

### Future Migration Path to SaaS/Cloud Sync

Architecture is designed for eventual migration:

1. **Repository pattern** isolates all data access — swap SQLite for PostgreSQL by implementing same interface
2. **UUID primary keys** — no auto-increment conflicts during sync
3. **Timestamps on all records** — `created_at`, `updated_at`, `deleted_at` enable conflict resolution
4. **Soft deletes everywhere** — sync requires knowing what was deleted
5. **Audit log** — provides full change history for conflict resolution
6. **Service layer** — business logic is transport-agnostic; can be exposed via HTTP API later

Migration path: Local-only → Local + cloud backup → Eventual consistency sync → Full cloud with offline cache

### Why SQLite is Suitable Initially

- Single clinic = 1-5 concurrent users maximum
- SQLite handles millions of rows; a clinic generates ~10K-50K records/year
- WAL mode supports concurrent readers with single writer (sufficient for reception + doctor)
- No DBA required, no server process, no connection pooling
- Backup is literally copying one file
- Proven in production: every iPhone, every Android device, every browser uses SQLite

### Updates/Installer Strategy

- **Installer**: NSIS-based `.exe` installer via Wails build system
- **Auto-update**: Embed version in binary; on startup (if internet available), check GitHub releases API for newer version
- **Update flow**: Download new installer → prompt user → user runs installer (overwrites binary, preserves data directory)
- **Database migrations**: Applied automatically on startup; migration version tracked in `migrations` table
- **Rollback**: Pre-migration backup created automatically before applying new migrations

### Security Model

- Application runs under the OS user's permissions
- Database file stored in user-specific `%APPDATA%` directory
- Application-level authentication (username/password) controls access to features
- bcrypt-hashed passwords with configurable cost factor
- Session token stored in Go backend memory (not browser localStorage)
- Sensitive data (patient records) protected by application login, not OS permissions alone
- Audit log captures all data mutations with user attribution

---

## SECTION 2 — COMPLETE PROJECT STRUCTURE

```
Practivo/
├── main.go                          # Wails application entry point
├── app.go                           # Wails app struct, lifecycle hooks
├── wails.json                       # Wails project configuration
├── go.mod
├── go.sum
├── build/                           # Wails build assets
│   ├── appicon.png
│   ├── windows/
│   │   ├── icon.ico
│   │   ├── installer/
│   │   │   └── project.nsi          # NSIS installer script
│   │   └── wails.exe.manifest
│   └── README.md
│
├── cmd/
│   └── migrate/
│       └── main.go                  # Standalone migration runner (dev tool)
│
├── internal/
│   ├── app/
│   │   └── app.go                   # Application container, DI wiring
│   │
│   ├── db/
│   │   ├── database.go              # SQLite connection, GORM setup, WAL config
│   │   ├── migrations.go            # Migration runner
│   │   └── seed.go                  # Default data seeding (treatment catalog)
│   │
│   ├── models/
│   │   ├── base.go                  # BaseModel (ID, timestamps, soft delete)
│   │   ├── user.go                  # User, Role
│   │   ├── patient.go               # Patient
│   │   ├── appointment.go           # Appointment
│   │   ├── treatment.go             # Treatment catalog + performed treatments
│   │   ├── invoice.go               # Invoice, InvoiceItem
│   │   ├── payment.go               # Payment
│   │   ├── clinic.go                # ClinicSettings
│   │   └── audit.go                 # AuditLog
│   │
│   ├── repository/
│   │   ├── interfaces.go            # All repository interfaces
│   │   ├── user_repo.go
│   │   ├── patient_repo.go
│   │   ├── appointment_repo.go
│   │   ├── treatment_repo.go
│   │   ├── invoice_repo.go
│   │   ├── payment_repo.go
│   │   ├── clinic_repo.go
│   │   └── audit_repo.go
│   │
│   ├── service/
│   │   ├── auth_service.go
│   │   ├── patient_service.go
│   │   ├── appointment_service.go
│   │   ├── invoice_service.go
│   │   ├── dashboard_service.go
│   │   ├── backup_service.go
│   │   ├── settings_service.go
│   │   └── audit_service.go
│   │
│   ├── handler/
│   │   ├── auth_handler.go          # Wails-bound methods for auth
│   │   ├── patient_handler.go
│   │   ├── appointment_handler.go
│   │   ├── invoice_handler.go
│   │   ├── dashboard_handler.go
│   │   ├── backup_handler.go
│   │   ├── settings_handler.go
│   │   └── report_handler.go
│   │
│   ├── auth/
│   │   ├── password.go              # bcrypt hashing/verification
│   │   ├── session.go               # In-memory session management
│   │   └── rbac.go                  # Role-based access control
│   │
│   ├── backup/
│   │   ├── backup.go                # SQLite backup (file copy + integrity check)
│   │   ├── restore.go               # Restore from backup file
│   │   └── scheduler.go             # Auto-backup scheduler
│   │
│   ├── billing/
│   │   ├── calculator.go            # Invoice totals, GST, discounts
│   │   ├── numbering.go             # Invoice number generation
│   │   └── pdf.go                   # PDF generation for invoices
│   │
│   ├── utils/
│   │   ├── validator.go             # Common validation helpers
│   │   ├── formatter.go             # Date, currency formatting
│   │   └── errors.go                # Application error types
│   │
│   └── config/
│       └── config.go                # App configuration, paths, defaults
│
├── migrations/
│   ├── 001_initial_schema.sql
│   ├── 002_seed_treatments.sql
│   └── README.md
│
├── frontend/
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   ├── tailwind.config.ts
│   ├── postcss.config.js
│   ├── vite.config.ts
│   ├── components.json               # shadcn/ui config
│   │
│   └── src/
│       ├── main.tsx                   # React entry point
│       ├── App.tsx                    # Root component, router setup
│       │
│       ├── components/
│       │   ├── ui/                    # shadcn/ui primitives (button, input, etc.)
│       │   ├── layout/
│       │   │   ├── AppLayout.tsx      # Main app shell (sidebar + content)
│       │   │   ├── Sidebar.tsx
│       │   │   ├── Header.tsx
│       │   │   └── PageContainer.tsx
│       │   ├── shared/
│       │   │   ├── DataTable.tsx      # Reusable data table
│       │   │   ├── SearchInput.tsx
│       │   │   ├── ConfirmDialog.tsx
│       │   │   ├── LoadingSpinner.tsx
│       │   │   ├── EmptyState.tsx
│       │   │   └── ErrorBoundary.tsx
│       │   ├── patients/
│       │   │   ├── PatientForm.tsx
│       │   │   ├── PatientCard.tsx
│       │   │   └── PatientHistory.tsx
│       │   ├── billing/
│       │   │   ├── InvoiceForm.tsx
│       │   │   ├── InvoicePreview.tsx
│       │   │   ├── PaymentForm.tsx
│       │   │   └── InvoiceTable.tsx
│       │   ├── appointments/
│       │   │   ├── AppointmentForm.tsx
│       │   │   ├── AppointmentCalendar.tsx
│       │   │   └── AppointmentList.tsx
│       │   └── dashboard/
│       │       ├── StatsCard.tsx
│       │       ├── TodayAppointments.tsx
│       │       └── RevenueChart.tsx
│       │
│       ├── pages/
│       │   ├── SetupWizard.tsx
│       │   ├── Login.tsx
│       │   ├── Dashboard.tsx
│       │   ├── Patients.tsx
│       │   ├── PatientDetail.tsx
│       │   ├── Billing.tsx
│       │   ├── InvoiceDetail.tsx
│       │   ├── Appointments.tsx
│       │   ├── Reports.tsx
│       │   ├── Settings.tsx
│       │   └── NotFound.tsx
│       │
│       ├── layouts/
│       │   ├── AuthLayout.tsx         # Layout for login/setup (no sidebar)
│       │   └── MainLayout.tsx         # Layout for authenticated pages
│       │
│       ├── hooks/
│       │   ├── useAuth.ts
│       │   ├── usePatients.ts
│       │   ├── useInvoices.ts
│       │   ├── useAppointments.ts
│       │   ├── useDebounce.ts
│       │   └── useNotification.ts
│       │
│       ├── store/
│       │   ├── authStore.ts
│       │   ├── patientStore.ts
│       │   ├── invoiceStore.ts
│       │   ├── appointmentStore.ts
│       │   ├── settingsStore.ts
│       │   └── uiStore.ts            # Sidebar state, theme, notifications
│       │
│       ├── types/
│       │   ├── models.ts             # Mirrors Go models (auto-generated by Wails)
│       │   ├── api.ts                # Request/response types
│       │   └── common.ts             # Shared utility types
│       │
│       ├── lib/
│       │   ├── utils.ts              # cn() helper, formatters
│       │   ├── constants.ts          # App constants, enum maps
│       │   └── validators.ts         # Zod schemas
│       │
│       ├── services/
│       │   ├── patientService.ts     # Wraps Wails bindings for patients
│       │   ├── invoiceService.ts
│       │   ├── appointmentService.ts
│       │   ├── authService.ts
│       │   └── backupService.ts
│       │
│       └── routes/
│           ├── index.tsx             # Route definitions
│           ├── ProtectedRoute.tsx    # Auth guard
│           └── SetupGuard.tsx        # Redirects to setup if not configured
```

### Folder Responsibilities

| Folder | Responsibility |
|--------|---------------|
| `internal/models/` | GORM struct definitions. Single source of truth for data shape. No business logic. |
| `internal/repository/` | Data access only. Raw CRUD + queries. No business rules. Interface-driven for testability. |
| `internal/service/` | Business logic orchestration. Validation, authorization checks, multi-repo coordination. |
| `internal/handler/` | Wails-bound methods. Input sanitization, response formatting. Thin — delegates to services. |
| `internal/auth/` | Password hashing, session lifecycle, RBAC enforcement. |
| `internal/backup/` | Database backup/restore, integrity verification, scheduled backups. |
| `internal/billing/` | Invoice calculation engine, number generation, PDF rendering. Isolated from persistence. |
| `internal/db/` | Database connection, migration execution, GORM configuration. |
| `internal/config/` | Application paths, defaults, runtime configuration. |
| `frontend/src/services/` | Thin wrappers over Wails-generated bindings. Adds error normalization. |
| `frontend/src/store/` | Zustand stores. Client-side state. Never duplicates server state unnecessarily. |
| `frontend/src/hooks/` | Custom React hooks combining store access + service calls. |
| `frontend/src/components/ui/` | shadcn/ui primitives. Never modified directly. |
| `frontend/src/components/shared/` | App-specific reusable components. |

---

## SECTION 3 — MVP FEATURE SCOPE

### Included in MVP (v1.0)

| Feature | Scope |
|---------|-------|
| **Setup Wizard** | Clinic name, address, GST number, logo upload, admin account creation. One-time flow. |
| **Authentication** | Login/logout, password change. Single role initially (admin). Lock after 5 failed attempts. |
| **Patient Management** | Register, edit, search, list. Fields: name, age, gender, phone, address, medical history notes, allergies. |
| **Visit History** | Per-patient list of treatments performed with dates and notes. |
| **Treatment Catalog** | Predefined list of dental procedures with default prices. Editable. |
| **Billing/Invoices** | Create invoice linked to patient + treatments. Auto-calculate totals + GST. Sequential numbering. |
| **Payments** | Record full/partial payments against invoices. Track pending dues. |
| **Invoice Printing** | Print-optimized invoice view. Browser print dialog (Wails supports `window.print()`). |
| **Appointment Scheduling** | Create/edit/cancel appointments. Today's view. Week view. Status: scheduled/completed/cancelled. |
| **Dashboard** | Today's appointments, patients seen today, revenue today/this month, pending dues total. |
| **Basic Reports** | Daily collection report, monthly revenue summary, outstanding dues list. |
| **Backup/Restore** | Manual backup to user-chosen directory. Restore from backup file. Auto-backup on app close. |
| **Settings** | Edit clinic info, manage treatment catalog, change password. |

### Explicitly Excluded from MVP

| Excluded | Reason |
|----------|--------|
| AI features | Complexity; zero value for v1 target clinics |
| Multi-branch sync | Requires server infrastructure; out of scope for single-clinic target |
| Cloud sync | Defeats offline-first promise; Phase 2+ |
| Insurance/TPA | Indian dental clinics rarely process insurance for routine procedures |
| Mobile app | Desktop-first; mobile adds platform complexity |
| Advanced analytics | Basic reports sufficient for single-doctor clinic |
| Multi-language | English + basic Hindi labels sufficient for MVP |
| SMS/WhatsApp reminders | Requires internet + third-party integration |
| Inventory management | Nice-to-have; not core billing/patient flow |
| X-ray/image storage | Storage complexity; DICOM not needed for small clinics |
| Multi-user roles | Single admin sufficient for 1-2 person clinics |
| Prescription module | Out of scope; dentists use separate prescription pads |

### Why These Decisions Are Correct

The target user is a single-doctor Indian dental clinic with 1-2 staff. Their current system is a paper register or Excel sheet. The MVP must be:

1. **Simpler than their current workflow** — or they won't adopt it
2. **Instantly useful** — patient lookup + billing covers 90% of daily needs
3. **Zero learning curve for core tasks** — register patient, create bill, print
4. **Reliable** — offline, fast, never loses data

Every excluded feature adds onboarding friction without solving the core problem: "I need to find patient records and generate bills quickly."

---

## SECTION 4 — DATABASE DESIGN

### Table List & Relationships

```
clinic_settings (1 row)
    │
users ──────────────┐
    │                │ (created_by FK on most tables)
    │                │
patients ────────────┤
    │                │
    ├── appointments │
    │                │
    ├── invoices ────┤
    │       │        │
    │       ├── invoice_items
    │       │        │
    │       └── payments
    │
    └── patient_treatments
            │
        treatments (catalog)

audit_logs (references all entities by type + ID)
migrations (schema version tracking)
```

### Relationship Diagram

```
clinic_settings
  PK: id (UUID)

users
  PK: id (UUID)
  UNIQUE: username

patients
  PK: id (UUID)
  FK: created_by → users.id
  INDEX: phone, name (for search)

treatments (catalog)
  PK: id (UUID)
  -- Predefined dental procedures with default pricing

appointments
  PK: id (UUID)
  FK: patient_id → patients.id
  FK: created_by → users.id
  INDEX: appointment_date, status

invoices
  PK: id (UUID)
  FK: patient_id → patients.id
  FK: created_by → users.id
  UNIQUE: invoice_number
  INDEX: invoice_date, status

invoice_items
  PK: id (UUID)
  FK: invoice_id → invoices.id
  FK: treatment_id → treatments.id (nullable — custom line items allowed)

payments
  PK: id (UUID)
  FK: invoice_id → invoices.id
  FK: received_by → users.id
  INDEX: payment_date

patient_treatments (visit history)
  PK: id (UUID)
  FK: patient_id → patients.id
  FK: treatment_id → treatments.id
  FK: invoice_id → invoices.id (nullable)
  INDEX: treatment_date

audit_logs
  PK: id (UUID)
  FK: user_id → users.id
  INDEX: entity_type + entity_id, created_at
  -- No soft delete on audit logs (immutable)

migrations
  PK: version (integer)
  -- Tracks applied migration versions
```

### UUID Strategy

- All primary keys are UUIDs (v4) generated in Go using `google/uuid`
- Stored as TEXT in SQLite (36 chars)
- Rationale: No conflicts during future sync, no sequential ID leakage, globally unique
- Trade-off accepted: Slightly larger than INTEGER PKs, but negligible at clinic scale

### Timestamp Strategy

- All tables (except `migrations`) include: `created_at`, `updated_at`
- Stored as ISO 8601 TEXT in SQLite (`2024-01-15T10:30:00Z`)
- Go side uses `time.Time`; GORM handles serialization
- All timestamps in UTC internally; formatted to local timezone in frontend
- `deleted_at` nullable timestamp for soft-deleted records

### Indexing Strategy

- Primary keys: automatic index
- Foreign keys: explicit index on every FK column
- Search columns: `patients.phone`, `patients.name` (case-insensitive via COLLATE NOCASE)
- Date columns used in range queries: `appointments.appointment_date`, `invoices.invoice_date`, `payments.payment_date`
- Composite index: `audit_logs(entity_type, entity_id)` for entity history lookup
- Invoice number: unique index

SQLite indexes are B-tree. At clinic scale (<100K records), most queries return in <1ms even without perfect indexing. Optimize only after profiling.

### Soft Delete Strategy

- GORM's built-in `gorm.DeletedAt` field on all entities except `audit_logs` and `migrations`
- Soft-deleted records excluded from all default queries via GORM's automatic `WHERE deleted_at IS NULL`
- Hard delete never exposed in application; only via direct DB maintenance
- Rationale: Healthcare data should never be permanently destroyed; supports future sync; supports audit compliance

### Audit Logging Strategy

- Every create/update/delete operation writes to `audit_logs`
- Schema: `id, user_id, action (CREATE|UPDATE|DELETE), entity_type, entity_id, old_value (JSON), new_value (JSON), created_at`
- `old_value` and `new_value` stored as JSON text — enables showing "what changed"
- Audit writes happen in the same transaction as the data mutation (consistency guarantee)
- Audit logs are append-only (no update, no delete)
- Retention: indefinite (at clinic scale, audit logs add negligible storage)

---

## SECTION 5 — SERVICE LAYER DESIGN

### AuthService

**Responsibilities:**
- User login/logout
- Password hashing and verification
- Session creation and validation
- Failed login attempt tracking
- Account lockout after threshold
- Password change

**Exposed Methods:**
```
Login(username, password) → (Session, error)
Logout(sessionToken) → error
ChangePassword(userID, oldPassword, newPassword) → error
ValidateSession(sessionToken) → (User, error)
CreateInitialAdmin(username, password) → error
```

**Dependencies:** UserRepository, AuditService

**Validation Rules:**
- Username: 3-50 chars, alphanumeric + underscore
- Password: minimum 6 chars (not enterprise; balancing security with clinic staff usability)
- Lock account after 5 consecutive failed attempts for 15 minutes

**Business Logic:**
- bcrypt cost factor 12
- Session expires after 8 hours (clinic working day)
- Only one active session per user (new login invalidates previous)

---

### PatientService

**Responsibilities:**
- Patient CRUD operations
- Patient search (by name, phone)
- Patient visit history retrieval
- Duplicate detection (same phone number)

**Exposed Methods:**
```
CreatePatient(input) → (Patient, error)
UpdatePatient(id, input) → (Patient, error)
GetPatient(id) → (Patient, error)
ListPatients(page, search) → ([]Patient, total, error)
GetPatientHistory(patientID) → ([]Treatment, error)
DeletePatient(id) → error
```

**Dependencies:** PatientRepository, AuditService

**Validation Rules:**
- Name: required, 2-100 chars
- Phone: required, 10 digits (Indian mobile)
- Age: 0-120
- Gender: M/F/Other
- Duplicate phone triggers warning (not hard block — family members share phones)

**Business Logic:**
- Cannot delete patient with unpaid invoices
- Search is case-insensitive, partial match on name and exact match on phone

---

### InvoiceService

**Responsibilities:**
- Invoice creation with line items
- Invoice total calculation (subtotal, GST, discount, grand total)
- Invoice numbering
- Payment recording
- Payment status tracking
- Invoice PDF generation trigger
- Outstanding dues calculation

**Exposed Methods:**
```
CreateInvoice(patientID, items, discount) → (Invoice, error)
GetInvoice(id) → (Invoice, error)
ListInvoices(filters) → ([]Invoice, total, error)
RecordPayment(invoiceID, amount, method) → (Payment, error)
GetOutstandingDues(patientID) → (amount, error)
GetInvoicePDF(id) → ([]byte, error)
VoidInvoice(id, reason) → error
```

**Dependencies:** InvoiceRepository, PaymentRepository, PatientRepository, BillingCalculator, InvoiceNumbering, AuditService

**Validation Rules:**
- At least one line item required
- All amounts in paise (positive integers)
- Payment cannot exceed remaining balance
- Cannot modify invoice after payment recorded (void and recreate)
- Discount cannot exceed subtotal

**Business Logic:**
- Invoice status: draft → issued → partial → paid → void
- Invoice number format: `DF-{YYMM}-{SEQ}` (e.g., DF-2401-0042)
- GST calculated per item based on clinic's GST configuration
- Void requires reason (audit trail)

---

### AppointmentService

**Responsibilities:**
- Appointment CRUD
- Schedule conflict detection
- Today's appointment list
- Status management
- Appointment reminders (future: desktop notification)

**Exposed Methods:**
```
CreateAppointment(patientID, datetime, duration, notes) → (Appointment, error)
UpdateAppointment(id, input) → (Appointment, error)
CancelAppointment(id, reason) → error
CompleteAppointment(id) → error
GetTodayAppointments() → ([]Appointment, error)
GetWeekAppointments(startDate) → ([]Appointment, error)
CheckConflict(datetime, duration) → (bool, error)
```

**Dependencies:** AppointmentRepository, PatientRepository, AuditService

**Validation Rules:**
- Appointment date must be today or future (for creation)
- Duration: 15-180 minutes in 15-minute increments
- Cannot have overlapping appointments (single doctor)
- Cannot cancel completed appointment

**Business Logic:**
- Default duration: 30 minutes
- Conflict check: new appointment overlaps if it starts during another or another starts during it
- Past appointments auto-marked as "no-show" if not completed (nightly check unnecessary for MVP; manual)

---

### DashboardService

**Responsibilities:**
- Aggregate statistics for dashboard display
- Today's metrics
- Monthly summary
- Pending actions

**Exposed Methods:**
```
GetDashboardStats() → (DashboardStats, error)
GetRevenueStats(period) → (RevenueStats, error)
GetDailyReport(date) → (DailyReport, error)
GetMonthlyReport(year, month) → (MonthlyReport, error)
GetOutstandingDuesList() → ([]DueEntry, error)
```

**Dependencies:** InvoiceRepository, AppointmentRepository, PaymentRepository, PatientRepository

**Business Logic:**
- Stats computed on-demand (no materialized views; SQLite aggregations are fast at this scale)
- Revenue = sum of payments received (not invoiced amounts)
- "Today" = local timezone date boundaries

---

### BackupService

**Responsibilities:**
- Manual database backup to user-specified location
- Automatic backup on application close
- Backup integrity verification
- Restore from backup file
- Backup history tracking

**Exposed Methods:**
```
CreateBackup(destinationPath) → (BackupInfo, error)
RestoreFromBackup(filePath) → error
ListBackups() → ([]BackupInfo, error)
VerifyBackup(filePath) → (bool, error)
GetAutoBackupPath() → (string, error)
SetAutoBackupPath(path) → error
```

**Dependencies:** Database connection, ClinicRepository (for settings), filesystem

**Validation Rules:**
- Destination must be writable
- Restore file must pass SQLite integrity check (`PRAGMA integrity_check`)
- Cannot restore while other operations in progress

**Business Logic:**
- Backup is SQLite's `.backup` API (consistent snapshot, no locking)
- Backup filename: `Practivo_backup_{YYYYMMDD_HHmmss}.db`
- Auto-backup keeps last 7 daily backups (configurable)
- Restore creates a backup of current DB before overwriting (safety net)
- After restore, application restarts to reinitialize connections

---

### SettingsService

**Responsibilities:**
- Clinic configuration management
- Setup wizard completion
- Treatment catalog management
- Application preferences

**Exposed Methods:**
```
GetClinicSettings() → (ClinicSettings, error)
UpdateClinicSettings(input) → error
IsSetupComplete() → (bool, error)
CompleteSetup(input) → error
ListTreatments() → ([]Treatment, error)
CreateTreatment(input) → (Treatment, error)
UpdateTreatment(id, input) → error
DeleteTreatment(id) → error
```

**Dependencies:** ClinicRepository, TreatmentRepository, AuditService

**Business Logic:**
- Setup wizard runs exactly once (first launch)
- Cannot delete treatment that's referenced in invoices (soft delete only)
- Clinic settings is a single-row table (upsert pattern)

---

### AuditService

**Responsibilities:**
- Recording all data mutations
- Audit trail retrieval
- Entity history reconstruction

**Exposed Methods:**
```
LogAction(userID, action, entityType, entityID, oldValue, newValue) → error
GetEntityHistory(entityType, entityID) → ([]AuditLog, error)
GetUserActivity(userID, dateRange) → ([]AuditLog, error)
```

**Dependencies:** AuditRepository

**Business Logic:**
- Never fails silently — if audit write fails, the parent transaction rolls back
- JSON serialization of old/new values (only changed fields for updates)
- Called by other services, not directly by handlers

---

## SECTION 6 — FRONTEND ARCHITECTURE

### Routing Architecture

Using React Router v6 with layout-based routing:

```
/setup              → SetupWizard (no layout, full-page)
/login              → Login (AuthLayout)
/dashboard          → Dashboard (MainLayout)
/patients           → Patient list (MainLayout)
/patients/:id       → Patient detail + history (MainLayout)
/billing            → Invoice list (MainLayout)
/billing/new        → Create invoice (MainLayout)
/billing/:id        → Invoice detail (MainLayout)
/appointments       → Appointment schedule (MainLayout)
/reports            → Reports page (MainLayout)
/settings           → Settings page (MainLayout)
```

**Route Guards:**
- `SetupGuard`: If setup not complete → redirect to `/setup`
- `ProtectedRoute`: If not authenticated → redirect to `/login`
- On app load: check setup status → check session → route accordingly

### Layout System

```
AuthLayout
├── Centered card container
├── App logo
└── Single content slot (login form / setup wizard)

MainLayout
├── Sidebar (collapsible)
│   ├── Clinic name/logo
│   ├── Navigation links
│   └── User info + logout
├── Header
│   ├── Page title (breadcrumb)
│   ├── Quick search
│   └── Notification area
└── Content area
    └── PageContainer (max-width, padding, scroll)
```

### Authentication Flow

```
App Start
    │
    ▼
Check: IsSetupComplete() ──── No ───► /setup
    │
   Yes
    │
    ▼
Check: ValidateSession() ──── Invalid ───► /login
    │
   Valid
    │
    ▼
Store user in authStore ───► /dashboard
```

- Session token held in Go backend memory (not exposed to frontend)
- Frontend calls `ValidateSession()` on mount — Go checks in-memory session map
- On login success, Go creates session and frontend stores `{ user, isAuthenticated }` in Zustand
- No token in localStorage — Wails binding calls are implicitly "authenticated" if Go says so
- Logout clears Go session + Zustand store

### Zustand Store Structure

```typescript
// authStore
{
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (username, password) => Promise<void>
  logout: () => Promise<void>
  checkSession: () => Promise<void>
}

// patientStore
{
  patients: Patient[]
  totalCount: number
  currentPatient: Patient | null
  searchQuery: string
  page: number
  isLoading: boolean
  fetchPatients: () => Promise<void>
  setSearch: (query) => void
  setPage: (page) => void
}

// invoiceStore
{
  invoices: Invoice[]
  currentInvoice: Invoice | null
  filters: InvoiceFilters
  isLoading: boolean
  fetchInvoices: () => Promise<void>
  createInvoice: (data) => Promise<Invoice>
}

// appointmentStore
{
  todayAppointments: Appointment[]
  weekAppointments: Appointment[]
  selectedDate: Date
  isLoading: boolean
  fetchToday: () => Promise<void>
  fetchWeek: (startDate) => Promise<void>
}

// settingsStore
{
  clinic: ClinicSettings | null
  treatments: Treatment[]
  isSetupComplete: boolean
  fetchSettings: () => Promise<void>
  checkSetup: () => Promise<boolean>
}

// uiStore
{
  sidebarCollapsed: boolean
  notifications: Notification[]
  toggleSidebar: () => void
  addNotification: (notification) => void
  removeNotification: (id) => void
}
```

### Shared Component Strategy

- All `shadcn/ui` components live in `components/ui/` — never modified
- App-level shared components in `components/shared/` — DataTable, SearchInput, ConfirmDialog, etc.
- Feature-specific components in `components/{feature}/` — PatientForm, InvoicePreview, etc.
- No "god components" — each component does one thing
- Props-driven; minimal internal state (state lives in stores)

### Form Handling Strategy

- React Hook Form for all forms
- Zod schemas for validation (shared between frontend validation and type inference)
- Pattern: `useForm()` → `zodResolver(schema)` → `onSubmit` calls service → service calls Wails binding
- Form errors displayed inline below fields
- Submit button disabled during submission
- Optimistic UI not needed (local calls are <50ms)

### Error Handling Approach

- Wails binding errors surface as rejected promises
- Services catch errors and normalize to `{ code, message }` shape
- Components use try/catch in submit handlers
- Global error boundary catches unhandled React errors → shows "something went wrong" with retry
- Toast notifications for operation results (success/failure)
- No silent failures

### Loading State Strategy

- Each store has an `isLoading` boolean
- Page-level skeleton loaders (not spinners) for initial data fetch
- Button-level loading state during form submission
- No loading states for cached data (re-fetch in background, show stale data)

### Notification/Toast Strategy

- Zustand `uiStore` manages notification queue
- Toast component renders from queue (bottom-right position)
- Types: success, error, warning, info
- Auto-dismiss after 4 seconds (errors persist until dismissed)
- Used for: "Patient created", "Invoice saved", "Backup complete", etc.

---

## SECTION 7 — SECURITY ARCHITECTURE

### Password Hashing Strategy

- Algorithm: bcrypt
- Cost factor: 12 (≈250ms on modern hardware; acceptable for local app)
- Library: `golang.org/x/crypto/bcrypt`
- No password stored in plaintext anywhere (DB, logs, memory after verification)
- Password change requires old password verification

### Session Handling

- Sessions stored in Go memory (map[string]Session)
- Session token: crypto/rand generated, 32 bytes, hex-encoded
- Session expires after 8 hours or on logout
- No session persistence across app restarts (user re-authenticates on launch)
- Frontend never sees or stores session token — authentication is implicit via Wails binding context

**Why not localStorage:** Wails webview has access to localStorage, but storing auth state there is inappropriate because:
1. Dev tools can inspect it
2. No server to validate tokens against (Go backend IS the server)
3. Go session map is the single source of truth — frontend asks "am I authenticated?" via binding call

### Role-Based Access Control

MVP has single role (admin/owner). Architecture supports future roles:

```
Roles: admin, doctor, receptionist
Permissions: patients.read, patients.write, invoices.create, settings.manage, etc.
```

RBAC enforcement at service layer (not handler layer) — every service method checks permission before executing.

### Failed Login Protection

- Track consecutive failed attempts per username in memory
- After 5 failures: lock account for 15 minutes
- Lock is in-memory (resets on app restart — acceptable for local desktop)
- Audit log records all failed login attempts
- No CAPTCHA needed (physical access to machine required)

### Database Backup Safety

- Backup files are plain SQLite databases (portable)
- Backup integrity verified with `PRAGMA integrity_check` before restore
- Current database backed up before restore (safety net)
- Backup files should be stored on external drive / USB by user (recommended in setup wizard)
- Future: encrypted backup option (AES-256-GCM wrapping the SQLite file)

### Sensitive Data Handling

- Patient data: stored as-is in SQLite (no field-level encryption in MVP)
- Justification: Database file access requires OS user login; field-level encryption adds complexity without meaningful security for single-user desktop app
- Future: SQLCipher for full-database encryption
- No sensitive data in application logs
- Medical history stored as free text (no structured diagnosis codes in MVP)

### Local File Storage Strategy

```
%APPDATA%/Practivo/
├── Practivo.db              # Main database
├── backups/                 # Auto-backup location
│   ├── Practivo_backup_20240115_093000.db
│   └── ...
├── config.json              # Non-sensitive app config (window size, theme)
└── logs/                    # Application logs (no PII)
    └── Practivo_2024-01-15.log
```

- All paths derived from `os.UserConfigDir()` (cross-platform compatible)
- No data stored in installation directory
- Directory created on first launch with restrictive permissions

### Invoice Tamper Prevention

- Invoices are immutable after creation (no edit, only void + recreate)
- Invoice number is sequential with no gaps (void doesn't release number)
- Audit log captures invoice creation with full details
- Hash-based integrity checking (future enhancement):
  - SHA-256 hash of invoice data stored in `invoice.integrity_hash`
  - Verified on display; tampering detected if hash mismatch
- For MVP: rely on audit log + immutability

### Audit Logging

- All mutations logged: who, what, when, old value, new value
- Audit log is append-only (no UPDATE/DELETE on audit_logs table)
- Logged in same transaction as mutation (atomic — if audit fails, mutation rolls back)
- Captures: login attempts, patient changes, invoice creation, payment recording, setting changes
- Retention: permanent (tiny storage cost at clinic scale)

---

## SECTION 8 — BILLING SYSTEM DESIGN

### Invoice Numbering Strategy

Format: `DF-{YYMM}-{SEQUENCE}`

Examples: `DF-2401-0001`, `DF-2401-0042`, `DF-2402-0001`

- Prefix `DF` = Practivo (configurable in settings for clinic branding)
- `YYMM` = year + month (resets sequence monthly for manageable numbers)
- Sequence = zero-padded 4-digit sequential within month
- Stored in `clinic_settings.last_invoice_number` — atomic increment via DB transaction
- Voided invoices retain their number (no reuse)
- Guarantees: unique, sequential, human-readable, sortable

### GST Handling

- Clinic's GSTIN stored in `clinic_settings`
- GST rate configurable per treatment (default: 18% for dental services)
- Invoice shows: subtotal, GST amount, total
- GST calculation: per line item, then sum (not on total — per Indian GST rules)
- If clinic is not GST registered: GST fields hidden, no tax calculation
- GST breakup on invoice: CGST (9%) + SGST (9%) for intra-state (covers 99% of dental clinics)

### Payment Tracking

- Multiple payments allowed per invoice
- Payment methods: Cash, UPI, Card, Bank Transfer, Other
- Each payment records: amount, method, date, received_by (user)
- Invoice status derived from payments:
  - `issued`: no payments
  - `partial`: sum(payments) < invoice total
  - `paid`: sum(payments) >= invoice total
  - `void`: manually voided

### Partial Payments

- Common in Indian dental clinics (patient pays across visits)
- Invoice shows: total, paid, balance
- Payment history visible on invoice detail
- Dashboard shows total outstanding across all patients
- Patient detail shows their total pending dues

### Pending Dues

- Calculated: sum of (invoice.total - sum(payments)) for all non-void invoices per patient
- Dashboard widget: total outstanding across clinic
- Report: list of patients with pending dues, sorted by amount
- No automated reminders in MVP (no internet required)

### Print Workflow

1. User clicks "Print" on invoice detail page
2. Frontend renders print-optimized view (hidden on screen, shown for print CSS)
3. Calls `window.print()` (Wails supports native print dialog)
4. Print layout: A5 size (half A4 — standard for Indian clinic bills)
5. Contains: clinic header, patient info, itemized treatments, totals, payment summary, footer

Alternative: Generate PDF → save/print from OS.

### PDF Export Strategy

- Go library: `jung-kurt/gofpdf` or `unidoc/unipdf`
- Template: defined in Go code (not HTML-to-PDF — better control)
- Layout: A5 portrait, clinic branding, structured invoice data
- Generated on-demand when user clicks "Download PDF"
- Saved to user-chosen location via Wails file dialog
- Not stored in database (regenerated from data each time)

### Money Handling Strategy

**All monetary values stored as integer paise (1 rupee = 100 paise).**

| Storage | Display |
|---------|---------|
| `15000` | ₹150.00 |
| `50075` | ₹500.75 |
| `100` | ₹1.00 |

**Why integer paise instead of float rupees:**

1. **Floating point is fundamentally broken for money.** `0.1 + 0.2 = 0.30000000000000004` in every language. This causes:
   - Totals that don't add up (₹499.99 instead of ₹500.00)
   - Rounding errors that compound across thousands of invoices
   - GST calculations that are off by paisa (tax authorities don't accept this)

2. **Integer arithmetic is exact.** `15000 + 50075 = 65075` — always. No rounding. No surprises.

3. **Industry standard.** Stripe uses cents. Razorpay uses paise. Every serious payment system uses smallest currency unit as integer.

4. **SQLite has no decimal type.** It would store floats as IEEE 754 doubles — guaranteed precision loss.

5. **Comparison is safe.** `amount == 0` works. With floats, you need epsilon comparisons.

**Implementation rules:**
- All calculations in paise (integer math)
- Division (e.g., GST split) uses `math.Round()` then cast to int64
- Frontend formats for display: `(amount / 100).toFixed(2)`
- API boundary: always paise in, paise out
- Database column type: INTEGER

---

## SECTION 9 — IMPLEMENTATION ROADMAP

### Phase 1: Architecture + Setup (Week 1)

**Objective:** Bootable Wails application with database connection and migration system.

**Files/Modules:**
- `main.go`, `app.go`, `wails.json`
- `internal/db/database.go`, `internal/db/migrations.go`
- `internal/config/config.go`
- `internal/models/base.go`
- `migrations/001_initial_schema.sql`
- `frontend/` scaffolding (Vite + React + Tailwind + shadcn)
- `go.mod` with all dependencies

**Implementation Order:**
1. `wails init` project scaffold
2. Go module setup with dependencies
3. Config system (app data paths)
4. Database connection + WAL mode
5. Migration runner
6. Initial schema migration
7. Frontend scaffold (Vite, Tailwind, shadcn/ui setup)
8. Verify Wails builds and runs

**Complexity:** Low-Medium

**Risks:**
- Wails v2 + modernc.org/sqlite CGo-free build compatibility
- GORM + SQLite type mapping for UUIDs

**Testing:**
- Database opens in WAL mode
- Migrations apply idempotently
- Wails builds successfully on Windows
- Frontend renders in webview

---

### Phase 2: Authentication + Setup Wizard (Week 2)

**Objective:** First-run setup wizard and login system.

**Files/Modules:**
- `internal/models/user.go`, `internal/models/clinic.go`
- `internal/repository/user_repo.go`, `internal/repository/clinic_repo.go`
- `internal/auth/password.go`, `internal/auth/session.go`
- `internal/service/auth_service.go`, `internal/service/settings_service.go`
- `internal/handler/auth_handler.go`, `internal/handler/settings_handler.go`
- `frontend/src/pages/SetupWizard.tsx`, `frontend/src/pages/Login.tsx`
- `frontend/src/store/authStore.ts`, `frontend/src/store/settingsStore.ts`
- `frontend/src/routes/` (guards, routing)

**Implementation Order:**
1. User + ClinicSettings models
2. Password hashing utilities
3. Session management
4. Auth service + handler
5. Settings service (setup completion check)
6. Frontend routing with guards
7. Setup wizard UI (multi-step form)
8. Login page UI
9. Auth store integration

**Complexity:** Medium

**Risks:**
- Session management without traditional HTTP (Wails context binding)
- Ensuring setup wizard runs exactly once

**Testing:**
- Password hash/verify roundtrip
- Login success/failure flows
- Session expiry
- Account lockout after failed attempts
- Setup wizard completion persists
- Route guards redirect correctly

---

### Phase 3: Patient Management (Week 3)

**Objective:** Full patient CRUD with search and visit history.

**Files/Modules:**
- `internal/models/patient.go`, `internal/models/treatment.go`
- `internal/repository/patient_repo.go`, `internal/repository/treatment_repo.go`
- `internal/service/patient_service.go`
- `internal/handler/patient_handler.go`
- `frontend/src/pages/Patients.tsx`, `frontend/src/pages/PatientDetail.tsx`
- `frontend/src/components/patients/`
- `frontend/src/store/patientStore.ts`
- `frontend/src/hooks/usePatients.ts`
- `migrations/002_seed_treatments.sql`

**Implementation Order:**
1. Patient + Treatment models
2. Patient repository (CRUD + search)
3. Treatment catalog seeding
4. Patient service with validation
5. Patient handler (Wails bindings)
6. Patient list page with search
7. Patient creation form
8. Patient detail page with history
9. DataTable shared component

**Complexity:** Medium

**Risks:**
- Search performance with LIKE queries (mitigated: COLLATE NOCASE index)
- Phone number validation for Indian numbers

**Testing:**
- Patient CRUD operations
- Search by name (partial) and phone (exact)
- Duplicate phone warning
- Pagination correctness
- Cannot delete patient with invoices

---

### Phase 4: Billing System (Week 4-5)

**Objective:** Invoice creation, payment recording, print/PDF.

**Files/Modules:**
- `internal/models/invoice.go`, `internal/models/payment.go`
- `internal/repository/invoice_repo.go`, `internal/repository/payment_repo.go`
- `internal/service/invoice_service.go`
- `internal/billing/calculator.go`, `internal/billing/numbering.go`, `internal/billing/pdf.go`
- `internal/handler/invoice_handler.go`
- `frontend/src/pages/Billing.tsx`, `frontend/src/pages/InvoiceDetail.tsx`
- `frontend/src/components/billing/`
- `frontend/src/store/invoiceStore.ts`

**Implementation Order:**
1. Invoice + InvoiceItem + Payment models
2. Invoice numbering system
3. Billing calculator (subtotal, GST, total)
4. Invoice repository
5. Payment repository
6. Invoice service (create, record payment, void)
7. Invoice handler
8. Invoice creation form (select patient, add items)
9. Invoice detail + payment recording UI
10. Print layout (CSS print stylesheet)
11. PDF generation

**Complexity:** High (most business logic lives here)

**Risks:**
- Integer arithmetic edge cases (rounding during GST split)
- Invoice number atomicity under concurrent access (unlikely but guard against)
- Print layout cross-printer compatibility
- PDF library selection and font embedding for ₹ symbol

**Testing:**
- Invoice total calculation (multiple items + GST)
- Partial payment tracking
- Invoice status transitions
- Cannot overpay invoice
- Invoice numbering sequential with no gaps
- Void doesn't release number
- Print layout renders correctly
- PDF generates with correct data

---

### Phase 5: Appointments + Reports (Week 6)

**Objective:** Appointment scheduling with conflict detection, basic reports.

**Files/Modules:**
- `internal/models/appointment.go`
- `internal/repository/appointment_repo.go`
- `internal/service/appointment_service.go`, `internal/service/dashboard_service.go`
- `internal/handler/appointment_handler.go`, `internal/handler/report_handler.go`
- `frontend/src/pages/Appointments.tsx`, `frontend/src/pages/Dashboard.tsx`, `frontend/src/pages/Reports.tsx`
- `frontend/src/components/appointments/`, `frontend/src/components/dashboard/`
- `frontend/src/store/appointmentStore.ts`

**Implementation Order:**
1. Appointment model
2. Appointment repository
3. Conflict detection logic
4. Appointment service
5. Dashboard service (aggregations)
6. Appointment handler + dashboard handler
7. Appointment list/calendar UI
8. Dashboard page with stats
9. Reports page (daily collection, monthly revenue, outstanding dues)

**Complexity:** Medium

**Risks:**
- Timezone handling for appointment times
- Conflict detection edge cases (exactly adjacent appointments)
- Report query performance (mitigated: indexes on date columns)

**Testing:**
- Appointment CRUD
- Conflict detection (overlapping times)
- Status transitions (scheduled → completed, scheduled → cancelled)
- Dashboard stats accuracy
- Report totals match individual records

---

### Phase 6: Backup/Restore + Packaging (Week 7)

**Objective:** Data safety and distributable installer.

**Files/Modules:**
- `internal/backup/backup.go`, `internal/backup/restore.go`, `internal/backup/scheduler.go`
- `internal/service/backup_service.go`
- `internal/handler/backup_handler.go`
- `internal/service/audit_service.go`
- `internal/handler/` (audit integration across all handlers)
- `build/windows/installer/project.nsi`
- `frontend/src/pages/Settings.tsx` (backup section)

**Implementation Order:**
1. Backup utility (SQLite backup API)
2. Restore utility with integrity check
3. Auto-backup on app close
4. Backup service + handler
5. Audit service implementation
6. Integrate audit logging into all existing services
7. Settings page (backup/restore UI)
8. NSIS installer configuration
9. Build pipeline (Wails build for Windows)
10. Installer testing on clean Windows machine

**Complexity:** Medium-High

**Risks:**
- SQLite backup while database is in use (WAL mode handles this)
- Restore requires app restart (UX consideration)
- NSIS installer configuration for first-time setup
- Windows Defender false positives on unsigned exe

**Testing:**
- Backup creates valid SQLite file
- Restore overwrites and restarts correctly
- Integrity check catches corrupted files
- Auto-backup fires on close
- Installer works on clean Windows 10/11
- Uninstaller removes app but preserves data directory
- Audit logs capture all operations

---

### Cross-Cutting Concerns (Throughout All Phases)

| Concern | Strategy |
|---------|----------|
| Error handling | Custom error types in Go; normalized error responses to frontend |
| Logging | Structured logging (zerolog) to file; rotate daily; no PII in logs |
| Testing | Repository tests with in-memory SQLite; service tests with mocked repos |
| Code generation | Wails generates TypeScript bindings after each backend change |
| Code style | `gofmt` + `golangci-lint` for Go; ESLint + Prettier for TypeScript |

---

## Summary of Key Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Framework | Wails v2 | Small binary, Go backend, native webview |
| Database | SQLite (WAL mode) | Zero-config, single-file, perfect for 1-5 users |
| ORM | GORM | Mature, migrations, soft deletes built-in |
| Primary keys | UUID v4 | Future sync compatibility, no sequential leakage |
| Money storage | Integer paise | Exact arithmetic, no floating point errors |
| Auth storage | Go in-memory session | No localStorage, no cookies, Wails context |
| State management | Zustand | Simple, TypeScript-first, no boilerplate |
| Forms | React Hook Form + Zod | Performant, schema-validated, type-safe |
| UI components | shadcn/ui | Accessible, customizable, Tailwind-native |
| Architecture | Service/Repository | Testable, swappable, SaaS-migration ready |
| Soft deletes | GORM DeletedAt | Healthcare data retention, sync compatibility |
| Invoice numbers | DF-YYMM-SEQ | Human-readable, sequential, monthly reset |
| PDF generation | Go-side library | Full control, no browser dependency |
| Backup | SQLite backup API | Consistent snapshot, no downtime |

This architecture supports the immediate goal (production-quality offline dental clinic software) while maintaining a clear upgrade path to multi-tenant SaaS when the market demands it.
