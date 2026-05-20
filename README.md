<div align="center">

# 🦷 ClinMitra Dental

**Smart Clinic Management for Indian Dental Practices**

[![CI](https://github.com/mdshoaibumer/clinmitra-clinic-operations/actions/workflows/ci.yml/badge.svg)](https://github.com/mdshoaibumer/clinmitra-clinic-operations/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/mdshoaibumer/clinmitra-clinic-operations?include_prereleases)](https://github.com/mdshoaibumer/clinmitra-clinic-operations/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/mdshoaibumer/clinmitra-clinic-operations)](go.mod)
[![License](https://img.shields.io/badge/license-proprietary-blue)](#license)

A modern, offline-first desktop application purpose-built for single-doctor dental clinics in India. No internet required. No monthly fees. Your data stays on your computer.

[Download Latest Release](https://github.com/mdshoaibumer/clinmitra-clinic-operations/releases/latest) · [Report Bug](https://github.com/mdshoaibumer/clinmitra-clinic-operations/issues) · [Request Feature](https://github.com/mdshoaibumer/clinmitra-clinic-operations/issues)

</div>

---

## ✨ Features

| Feature | Description |
|---------|-------------|
| **Patient Management** | Full patient records, treatment history, medical notes |
| **Appointments** | Daily/weekly calendar view, drag scheduling, conflict detection |
| **Invoicing & Billing** | GST-compliant invoices, payment tracking, outstanding balances |
| **Dashboard** | Daily/monthly revenue, appointment stats, quick actions |
| **Treatment Catalog** | Customizable treatment list with pricing |
| **Backup & Restore** | Local + cloud drive (OneDrive/Google Drive) auto-backup |
| **Auto Updates** | In-app update notification when new versions are released |
| **Multi-user Auth** | Admin/staff roles with session management |

## 🖥️ Screenshots

> *Coming soon*

## 🏗️ Tech Stack

| Layer | Technology |
|-------|------------|
| Desktop Framework | [Wails v2](https://wails.io) |
| Backend | Go 1.22 |
| Frontend | React 18 + TypeScript + Tailwind CSS |
| Database | SQLite (WAL mode) via GORM |
| State Management | Zustand |
| Build & Packaging | NSIS Installer |

## 📦 Installation

### For Users

1. Go to [Releases](https://github.com/mdshoaibumer/clinmitra-clinic-operations/releases/latest)
2. Download `ClinmitraDental-x.x.x-Setup.exe`
3. Run the installer
4. Launch ClinMitra Dental from Start Menu

### For Developers

#### Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Clone the repo
git clone https://github.com/mdshoaibumer/clinmitra-clinic-operations.git
cd clinmitra-clinic-operations

# Install frontend dependencies
cd frontend && npm install && cd ..

# Run in development mode (hot reload)
wails dev

# Build production binary
wails build

# Build with NSIS installer
wails build -nsis
```

## 🧪 Testing

```bash
# Run all Go tests
go test ./...

# Run with race detection + coverage
go test -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Run frontend E2E tests (requires app running)
cd frontend && npx playwright test
```

## 📁 Project Structure

```
├── .github/workflows/     # CI/CD pipelines
├── build/                 # Build assets (icons, manifests)
├── frontend/              # React + TypeScript frontend
│   ├── src/
│   │   ├── components/    # Reusable UI components
│   │   ├── pages/         # Route-level pages
│   │   ├── store/         # Zustand state stores
│   │   ├── lib/           # Utilities & validators
│   │   └── types/         # TypeScript type definitions
│   └── tests/             # Playwright E2E tests
├── internal/              # Go backend (not importable externally)
│   ├── app/               # Application wiring & lifecycle
│   ├── auth/              # Session management & password hashing
│   ├── config/            # Configuration & paths
│   ├── db/                # Database init, migrations, seeding
│   ├── handler/           # Wails-bound handlers (controller layer)
│   ├── models/            # GORM models & domain types
│   ├── repository/        # Data access layer
│   ├── service/           # Business logic layer
│   └── utils/             # Shared utilities & error types
├── main.go                # Application entry point
├── wails.json             # Wails project configuration
└── ARCHITECTURE.md        # Detailed architecture documentation
```

## 🚀 Release Process

Releases are automated via GitHub Actions. To create a new release:

```bash
# 1. Update version in internal/config/config.go
#    Version: "1.2.0"

# 2. Update version in wails.json
#    "productVersion": "1.2.0"

# 3. Commit the version bump
git add -A
git commit -m "chore: bump version to 1.2.0"

# 4. Create and push a tag
git tag v1.2.0
git push origin main --tags
```

This triggers the Release workflow which:
- Builds the Windows application with NSIS installer
- Creates a GitHub Release with auto-generated changelog
- Uploads the installer as a downloadable asset

The app's built-in update checker will notify existing users about the new version.

## 🔒 Security

- All passwords hashed with bcrypt (cost 12)
- Session-based auth with configurable timeout
- Login rate limiting with account lockout
- Input validation at all system boundaries
- SQLite database stored in user's AppData with restrictive permissions
- No network calls except update checks (optional)

## 🗺️ Roadmap

- [ ] SMS/WhatsApp appointment reminders
- [ ] Multi-clinic support
- [ ] Data export (Excel/PDF reports)
- [ ] Database encryption at rest
- [ ] Cloud sync between devices
- [ ] Mobile companion app

## 📄 License

This project is proprietary software. All rights reserved.

---

<div align="center">
Made with ❤️ for Indian dental clinics
</div>
