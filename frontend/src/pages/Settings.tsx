import { useEffect, useState } from 'react'
import { useSettingsStore } from '@/store/settingsStore'
import { useAuthStore } from '@/store/authStore'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { changePasswordSchema, type ChangePasswordFormData } from '@/lib/validators'
import { formatCurrency } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { Treatment } from '@/types/models'
import { Download, Upload, Plus, Trash2, Image, X } from 'lucide-react'

export default function Settings() {
  const { clinic: settings, treatments, isLoading, fetchSettings, fetchTreatments, updateSettings } = useSettingsStore()
  const { changePassword } = useAuthStore()
  const [activeTab, setActiveTab] = useState<'clinic' | 'treatments' | 'password' | 'backup'>('clinic')
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  // Clinic form state
  const [clinicName, setClinicName] = useState('')
  const [clinicPhone, setClinicPhone] = useState('')
  const [clinicAddress, setClinicAddress] = useState('')
  const [clinicEmail, setClinicEmail] = useState('')
  const [gstEnabled, setGstEnabled] = useState(false)
  const [gstin, setGstin] = useState('')
  const [gstRate, setGstRate] = useState(0)
  const [logoPreview, setLogoPreview] = useState<string>('')

  // Treatment form
  const [showTreatmentForm, setShowTreatmentForm] = useState(false)
  const [newTreatmentName, setNewTreatmentName] = useState('')
  const [newTreatmentCode, setNewTreatmentCode] = useState('')
  const [newTreatmentPrice, setNewTreatmentPrice] = useState('')
  const [newTreatmentCategory, setNewTreatmentCategory] = useState('')

  // Password form
  const { register: regPwd, handleSubmit: submitPwd, formState: { errors: pwdErrors }, reset: resetPwd } = useForm<ChangePasswordFormData>({
    resolver: zodResolver(changePasswordSchema),
  })

  // Backup state
  const [backups, setBackups] = useState<string[]>([])

  useEffect(() => {
    fetchSettings()
    fetchTreatments()
  }, [fetchSettings, fetchTreatments])

  useEffect(() => {
    if (settings) {
      setClinicName(settings.clinicName || '')
      setClinicPhone(settings.phone || '')
      setClinicAddress(settings.address || '')
      setClinicEmail(settings.email || '')
      setGstEnabled(settings.gstEnabled || false)
      setGstin(settings.gstin || '')
      setGstRate(settings.gstRate || 0)
      setLogoPreview(settings.logoBase64 || '')
    }
  }, [settings])

  useEffect(() => {
    if (activeTab === 'backup') loadBackups()
  }, [activeTab])

  const loadBackups = async () => {
    try {
      const list = await window.go.handler.BackupHandler.ListBackups()
      setBackups(list || [])
    } catch { setBackups([]) }
  }

  const handleSaveClinic = async () => {
    setError('')
    setMessage('')
    try {
      await updateSettings({
        ...settings!,
        clinicName,
        phone: clinicPhone,
        address: clinicAddress,
        email: clinicEmail,
        gstEnabled,
        gstin,
        gstRate,
      })
      setMessage('Settings saved successfully.')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to save settings')
    }
  }

  const handleLogoUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    // Validate file type
    if (!['image/png', 'image/jpeg', 'image/jpg', 'image/webp'].includes(file.type)) {
      setError('Please select a PNG, JPG, or WebP image.')
      return
    }

    // Validate size (max 512KB)
    if (file.size > 512 * 1024) {
      setError('Logo must be less than 512KB.')
      return
    }

    const reader = new FileReader()
    reader.onload = async (event) => {
      const base64 = event.target?.result as string
      try {
        await window.go.handler.SettingsHandler.UploadLogo(base64)
        setLogoPreview(base64)
        setMessage('Logo uploaded successfully.')
        setError('')
        fetchSettings()
      } catch (err: unknown) {
        setError(err instanceof Error ? err.message : 'Failed to upload logo')
      }
    }
    reader.readAsDataURL(file)
    // Reset input so the same file can be re-selected
    e.target.value = ''
  }

  const handleRemoveLogo = async () => {
    try {
      await window.go.handler.SettingsHandler.RemoveLogo()
      setLogoPreview('')
      setMessage('Logo removed.')
      setError('')
      fetchSettings()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to remove logo')
    }
  }

  const handleAddTreatment = async () => {
    if (!newTreatmentName || !newTreatmentPrice) return
    try {
      await window.go.handler.SettingsHandler.CreateTreatment(
        newTreatmentName,
        newTreatmentCode,
        newTreatmentCategory,
        '',
        Math.round(parseFloat(newTreatmentPrice) * 100),
      )
      setNewTreatmentName('')
      setNewTreatmentCode('')
      setNewTreatmentPrice('')
      setNewTreatmentCategory('')
      setShowTreatmentForm(false)
      fetchTreatments()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to add treatment')
    }
  }

  const handleDeleteTreatment = async (id: string) => {
    if (!confirm('Remove this treatment from catalog?')) return
    try {
      await window.go.handler.SettingsHandler.DeleteTreatment(id)
      fetchTreatments()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to delete treatment')
    }
  }

  const handleChangePassword = async (data: ChangePasswordFormData) => {
    setError('')
    setMessage('')
    try {
      await changePassword(data.oldPassword, data.newPassword)
      setMessage('Password changed successfully.')
      resetPwd()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to change password')
    }
  }

  const handleCreateBackup = async () => {
    setError('')
    setMessage('')
    try {
      const path = await window.go.handler.BackupHandler.CreateBackup()
      setMessage(`Backup created: ${path}`)
      loadBackups()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Backup failed')
    }
  }

  const handleRestore = async (filename: string) => {
    if (!confirm('Restore from this backup? The application will restart.')) return
    try {
      await window.go.handler.BackupHandler.RestoreFromBackup(filename)
      setMessage('Restore complete. Please restart the application.')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Restore failed')
    }
  }

  const tabs = [
    { key: 'clinic', label: 'Clinic' },
    { key: 'treatments', label: 'Treatments' },
    { key: 'password', label: 'Password' },
    { key: 'backup', label: 'Backup' },
  ] as const

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Settings</h1>

      {message && <div className="p-3 text-sm text-green-600 bg-green-50 rounded-md">{message}</div>}
      {error && <div className="p-3 text-sm text-red-600 bg-red-50 rounded-md">{error}</div>}

      {/* Tabs */}
      <div className="flex gap-2 border-b pb-2">
        {tabs.map(tab => (
          <Button
            key={tab.key}
            variant={activeTab === tab.key ? 'default' : 'ghost'}
            size="sm"
            onClick={() => { setActiveTab(tab.key); setMessage(''); setError('') }}
          >
            {tab.label}
          </Button>
        ))}
      </div>

      {/* Clinic Settings */}
      {activeTab === 'clinic' && (
        <Card>
          <CardHeader><CardTitle>Clinic Information</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Clinic Name</Label>
                <Input value={clinicName} onChange={(e) => setClinicName(e.target.value)} />
              </div>
              <div className="space-y-2">
                <Label>Phone</Label>
                <Input value={clinicPhone} onChange={(e) => setClinicPhone(e.target.value)} />
              </div>
              <div className="space-y-2">
                <Label>Email</Label>
                <Input value={clinicEmail} onChange={(e) => setClinicEmail(e.target.value)} />
              </div>
              <div className="space-y-2">
                <Label>Address</Label>
                <Input value={clinicAddress} onChange={(e) => setClinicAddress(e.target.value)} />
              </div>
            </div>

            <div className="border-t pt-4 space-y-4">
              <h3 className="font-medium">Clinic Logo</h3>
              <p className="text-sm text-muted-foreground">Upload a logo to display on printed invoices. Recommended: PNG or JPG, max 512KB.</p>
              <div className="flex items-center gap-4">
                {logoPreview ? (
                  <div className="relative">
                    <img src={logoPreview} alt="Clinic Logo" className="h-16 max-w-48 object-contain border rounded p-1" />
                    <button
                      onClick={handleRemoveLogo}
                      className="absolute -top-2 -right-2 bg-red-500 text-white rounded-full p-0.5 hover:bg-red-600"
                      title="Remove logo"
                    >
                      <X className="h-3 w-3" />
                    </button>
                  </div>
                ) : (
                  <div className="h-16 w-32 border-2 border-dashed rounded flex items-center justify-center text-muted-foreground">
                    <Image className="h-6 w-6" />
                  </div>
                )}
                <div>
                  <label className="cursor-pointer">
                    <Button variant="outline" size="sm" asChild>
                      <span><Upload className="h-3 w-3 mr-1" /> {logoPreview ? 'Change' : 'Upload'}</span>
                    </Button>
                    <input
                      type="file"
                      accept="image/png,image/jpeg,image/jpg,image/webp"
                      className="hidden"
                      onChange={handleLogoUpload}
                    />
                  </label>
                </div>
              </div>
            </div>

            <div className="border-t pt-4 space-y-4">
              <h3 className="font-medium">GST Configuration</h3>
              <div className="flex items-center gap-3">
                <input type="checkbox" checked={gstEnabled} onChange={(e) => setGstEnabled(e.target.checked)} className="h-4 w-4" />
                <Label>Enable GST</Label>
              </div>
              {gstEnabled && (
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>GSTIN</Label>
                    <Input value={gstin} onChange={(e) => setGstin(e.target.value)} />
                  </div>
                  <div className="space-y-2">
                    <Label>GST Rate %</Label>
                    <Input type="number" value={gstRate} onChange={(e) => setGstRate(parseFloat(e.target.value) || 0)} />
                  </div>
                </div>
              )}
            </div>

            <Button onClick={handleSaveClinic}>Save Settings</Button>
          </CardContent>
        </Card>
      )}

      {/* Treatments */}
      {activeTab === 'treatments' && (
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>Treatment Catalog</CardTitle>
            <Button size="sm" onClick={() => setShowTreatmentForm(!showTreatmentForm)}>
              <Plus className="h-4 w-4 mr-1" /> Add
            </Button>
          </CardHeader>
          <CardContent className="space-y-4">
            {showTreatmentForm && (
              <div className="p-4 border rounded-md space-y-3">
                <div className="grid grid-cols-4 gap-3">
                  <Input placeholder="Name *" value={newTreatmentName} onChange={(e) => setNewTreatmentName(e.target.value)} />
                  <Input placeholder="Code" value={newTreatmentCode} onChange={(e) => setNewTreatmentCode(e.target.value)} />
                  <Input type="number" placeholder="Price (₹)" value={newTreatmentPrice} onChange={(e) => setNewTreatmentPrice(e.target.value)} />
                  <Input placeholder="Category" value={newTreatmentCategory} onChange={(e) => setNewTreatmentCategory(e.target.value)} />
                </div>
                <div className="flex gap-2">
                  <Button size="sm" onClick={handleAddTreatment}>Save</Button>
                  <Button size="sm" variant="outline" onClick={() => setShowTreatmentForm(false)}>Cancel</Button>
                </div>
              </div>
            )}

            {isLoading ? (
              <p className="text-muted-foreground">Loading...</p>
            ) : treatments.length === 0 ? (
              <p className="text-muted-foreground">No treatments in catalog.</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="border-b bg-muted/50">
                    <tr>
                      <th className="px-3 py-2 text-left text-sm font-medium">Code</th>
                      <th className="px-3 py-2 text-left text-sm font-medium">Name</th>
                      <th className="px-3 py-2 text-left text-sm font-medium">Category</th>
                      <th className="px-3 py-2 text-right text-sm font-medium">Price</th>
                      <th className="px-3 py-2 text-right text-sm font-medium">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {treatments.map((t: Treatment) => (
                      <tr key={t.id} className="border-b">
                        <td className="px-3 py-2 text-sm font-mono">{t.code || '-'}</td>
                        <td className="px-3 py-2 text-sm">{t.name}</td>
                        <td className="px-3 py-2 text-sm">{t.category || '-'}</td>
                        <td className="px-3 py-2 text-sm text-right">{formatCurrency(t.defaultPrice)}</td>
                        <td className="px-3 py-2 text-right">
                          <Button variant="ghost" size="icon" onClick={() => handleDeleteTreatment(t.id)}>
                            <Trash2 className="h-4 w-4 text-red-500" />
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Change Password */}
      {activeTab === 'password' && (
        <Card>
          <CardHeader><CardTitle>Change Password</CardTitle></CardHeader>
          <CardContent>
            <form onSubmit={submitPwd(handleChangePassword)} className="space-y-4 max-w-md">
              <div className="space-y-2">
                <Label>Current Password</Label>
                <Input type="password" {...regPwd('oldPassword')} />
                {pwdErrors.oldPassword && <p className="text-sm text-red-500">{pwdErrors.oldPassword.message}</p>}
              </div>
              <div className="space-y-2">
                <Label>New Password</Label>
                <Input type="password" {...regPwd('newPassword')} />
                {pwdErrors.newPassword && <p className="text-sm text-red-500">{pwdErrors.newPassword.message}</p>}
              </div>
              <div className="space-y-2">
                <Label>Confirm New Password</Label>
                <Input type="password" {...regPwd('confirmPassword')} />
                {pwdErrors.confirmPassword && <p className="text-sm text-red-500">{pwdErrors.confirmPassword.message}</p>}
              </div>
              <Button type="submit">Change Password</Button>
            </form>
          </CardContent>
        </Card>
      )}

      {/* Backup */}
      {activeTab === 'backup' && (
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>Database Backup</CardTitle>
            <Button onClick={handleCreateBackup}>
              <Download className="h-4 w-4 mr-2" /> Create Backup
            </Button>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Backups are stored locally. A backup is automatically created when the application closes.
            </p>
            {backups.length === 0 ? (
              <p className="text-muted-foreground">No backups found.</p>
            ) : (
              <div className="space-y-2">
                {backups.map((backup) => (
                  <div key={backup} className="flex items-center justify-between p-3 bg-muted rounded-md">
                    <span className="text-sm font-mono">{backup}</span>
                    <Button variant="outline" size="sm" onClick={() => handleRestore(backup)}>
                      <Upload className="h-3 w-3 mr-1" /> Restore
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  )
}
