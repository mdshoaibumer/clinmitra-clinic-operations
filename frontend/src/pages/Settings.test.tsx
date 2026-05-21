import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import Settings from './Settings'

// Mock stores
const mockFetchSettings = vi.fn()
const mockFetchTreatments = vi.fn()
const mockUpdateSettings = vi.fn()
const mockChangePassword = vi.fn()

vi.mock('@/store/settingsStore', () => ({
  useSettingsStore: () => ({
    clinic: null,
    treatments: [],
    isLoading: false,
    fetchSettings: mockFetchSettings,
    fetchTreatments: mockFetchTreatments,
    updateSettings: mockUpdateSettings,
  }),
}))

vi.mock('@/store/authStore', () => ({
  useAuthStore: () => ({
    changePassword: mockChangePassword,
  }),
}))

// Mock toast
const mockToast = vi.fn()
vi.mock('@/components/ui/use-toast', () => ({
  useToast: () => ({ toast: mockToast }),
}))

// Mock window.go
const mockCheckForUpdate = vi.fn()
const mockDownloadAndInstallUpdate = vi.fn()
const mockListBackups = vi.fn().mockResolvedValue([])
const mockDetectCloudDrives = vi.fn().mockResolvedValue([])

beforeEach(() => {
  vi.clearAllMocks()
  ;(window as any).go = {
    handler: {
      UpdateHandler: {
        CheckForUpdate: mockCheckForUpdate,
        DownloadAndInstallUpdate: mockDownloadAndInstallUpdate,
      },
      BackupHandler: {
        ListBackups: mockListBackups,
        DetectCloudDrives: mockDetectCloudDrives,
        CreateBackup: vi.fn(),
        CreateCloudBackup: vi.fn(),
        RestoreFromBackup: vi.fn(),
        VerifyBackup: vi.fn(),
        GetAutoBackupPath: vi.fn(),
      },
    },
  }
})

describe('Settings - About Tab', () => {
  const navigateToAbout = () => {
    render(<Settings />)
    fireEvent.click(screen.getByRole('button', { name: /about/i }))
  }

  it('renders About tab with app info', () => {
    navigateToAbout()
    expect(screen.getByText('About ClinMitra Dental')).toBeInTheDocument()
    expect(screen.getByText('ClinMitra Dental')).toBeInTheDocument()
    expect(screen.getByText('1.0.0')).toBeInTheDocument()
  })

  it('shows Check for Updates button', () => {
    navigateToAbout()
    expect(screen.getByRole('button', { name: /check for updates/i })).toBeInTheDocument()
  })

  it('calls CheckForUpdate when button clicked', async () => {
    mockCheckForUpdate.mockResolvedValue({
      available: false,
      currentVersion: '1.0.0',
      latestVersion: '1.0.0',
      downloadURL: '',
      releaseNotes: '',
    })

    navigateToAbout()
    fireEvent.click(screen.getByRole('button', { name: /check for updates/i }))

    await waitFor(() => {
      expect(mockCheckForUpdate).toHaveBeenCalledTimes(1)
    })
  })

  it('shows up-to-date message when no update available', async () => {
    mockCheckForUpdate.mockResolvedValue({
      available: false,
      currentVersion: '1.0.0',
      latestVersion: '1.0.0',
      downloadURL: '',
      releaseNotes: '',
    })

    navigateToAbout()
    fireEvent.click(screen.getByRole('button', { name: /check for updates/i }))

    await waitFor(() => {
      expect(screen.getByText(/you're up to date/i)).toBeInTheDocument()
    })
  })

  it('shows toast on successful up-to-date check', async () => {
    mockCheckForUpdate.mockResolvedValue({
      available: false,
      currentVersion: '1.0.0',
      latestVersion: '1.0.0',
      downloadURL: '',
      releaseNotes: '',
    })

    navigateToAbout()
    fireEvent.click(screen.getByRole('button', { name: /check for updates/i }))

    await waitFor(() => {
      expect(mockToast).toHaveBeenCalledWith(
        expect.objectContaining({ title: 'Up to date' })
      )
    })
  })

  it('shows update available UI when newer version exists', async () => {
    mockCheckForUpdate.mockResolvedValue({
      available: true,
      currentVersion: '1.0.0',
      latestVersion: '1.1.0',
      downloadURL: 'https://github.com/test/releases/download/v1.1.0/installer.exe',
      releaseNotes: 'Bug fixes and improvements',
    })

    navigateToAbout()
    fireEvent.click(screen.getByRole('button', { name: /check for updates/i }))

    await waitFor(() => {
      expect(screen.getByText(/update available: v1\.1\.0/i)).toBeInTheDocument()
      expect(screen.getByText('Bug fixes and improvements')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /download & install/i })).toBeInTheDocument()
    })
  })

  it('calls DownloadAndInstallUpdate with correct URL', async () => {
    const downloadURL = 'https://github.com/test/releases/download/v1.1.0/installer.exe'
    mockCheckForUpdate.mockResolvedValue({
      available: true,
      currentVersion: '1.0.0',
      latestVersion: '1.1.0',
      downloadURL,
      releaseNotes: '',
    })
    mockDownloadAndInstallUpdate.mockResolvedValue(undefined)

    navigateToAbout()
    fireEvent.click(screen.getByRole('button', { name: /check for updates/i }))

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /download & install/i })).toBeInTheDocument()
    })

    fireEvent.click(screen.getByRole('button', { name: /download & install/i }))

    await waitFor(() => {
      expect(mockDownloadAndInstallUpdate).toHaveBeenCalledWith(downloadURL)
    })
  })

  it('shows toast on successful download', async () => {
    mockCheckForUpdate.mockResolvedValue({
      available: true,
      currentVersion: '1.0.0',
      latestVersion: '1.1.0',
      downloadURL: 'https://example.com/installer.exe',
      releaseNotes: '',
    })
    mockDownloadAndInstallUpdate.mockResolvedValue(undefined)

    navigateToAbout()
    fireEvent.click(screen.getByRole('button', { name: /check for updates/i }))

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /download & install/i })).toBeInTheDocument()
    })

    fireEvent.click(screen.getByRole('button', { name: /download & install/i }))

    await waitFor(() => {
      expect(mockToast).toHaveBeenCalledWith(
        expect.objectContaining({ title: 'Update downloaded' })
      )
    })
  })

  it('shows destructive toast on check failure', async () => {
    mockCheckForUpdate.mockRejectedValue(new Error('Network error'))

    navigateToAbout()
    fireEvent.click(screen.getByRole('button', { name: /check for updates/i }))

    await waitFor(() => {
      expect(mockToast).toHaveBeenCalledWith(
        expect.objectContaining({ variant: 'destructive', title: 'Update check failed' })
      )
    })
  })

  it('shows destructive toast on download failure', async () => {
    mockCheckForUpdate.mockResolvedValue({
      available: true,
      currentVersion: '1.0.0',
      latestVersion: '1.1.0',
      downloadURL: 'https://example.com/installer.exe',
      releaseNotes: '',
    })
    mockDownloadAndInstallUpdate.mockRejectedValue(new Error('Download interrupted'))

    navigateToAbout()
    fireEvent.click(screen.getByRole('button', { name: /check for updates/i }))

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /download & install/i })).toBeInTheDocument()
    })

    fireEvent.click(screen.getByRole('button', { name: /download & install/i }))

    await waitFor(() => {
      expect(mockToast).toHaveBeenCalledWith(
        expect.objectContaining({ variant: 'destructive', title: 'Update failed' })
      )
    })
  })

  it('disables button while checking for updates', async () => {
    // Never-resolving promise to keep loading state
    mockCheckForUpdate.mockReturnValue(new Promise(() => {}))

    navigateToAbout()
    fireEvent.click(screen.getByRole('button', { name: /check for updates/i }))

    await waitFor(() => {
      const btn = screen.getByRole('button', { name: /checking/i })
      expect(btn).toBeDisabled()
    })
  })

  it('disables button while downloading', async () => {
    mockCheckForUpdate.mockResolvedValue({
      available: true,
      currentVersion: '1.0.0',
      latestVersion: '1.1.0',
      downloadURL: 'https://example.com/installer.exe',
      releaseNotes: '',
    })
    mockDownloadAndInstallUpdate.mockReturnValue(new Promise(() => {}))

    navigateToAbout()
    fireEvent.click(screen.getByRole('button', { name: /check for updates/i }))

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /download & install/i })).toBeInTheDocument()
    })

    fireEvent.click(screen.getByRole('button', { name: /download & install/i }))

    await waitFor(() => {
      const btn = screen.getByRole('button', { name: /downloading/i })
      expect(btn).toBeDisabled()
    })
  })
})
