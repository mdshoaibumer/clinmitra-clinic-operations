import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { usePatientStore } from '@/store/patientStore'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { patientSchema, type PatientFormData } from '@/lib/validators'
import { GENDER_OPTIONS, BLOOD_GROUP_OPTIONS, INDIAN_STATES, INDIAN_CITIES } from '@/lib/constants'
import { useDebounce } from '@/lib/useDebounce'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Plus, Search } from 'lucide-react'

export default function Patients() {
  const navigate = useNavigate()
  const { patients, totalCount, searchQuery, page, isLoading, fetchPatients, setSearch, setPage } = usePatientStore()
  const [showForm, setShowForm] = useState(false)
  const [formError, setFormError] = useState('')
  const [localSearch, setLocalSearch] = useState(searchQuery)
  const [selectedState, setSelectedState] = useState('')
  const debouncedSearch = useDebounce(localSearch, 300)

  const { register, handleSubmit, formState: { errors }, reset } = useForm<PatientFormData>({
    resolver: zodResolver(patientSchema),
    defaultValues: { gender: 'male' },
  })

  useEffect(() => {
    setSearch(debouncedSearch)
  }, [debouncedSearch, setSearch])

  useEffect(() => {
    fetchPatients()
  }, [fetchPatients, page, searchQuery])

  const onSubmit = async (data: PatientFormData) => {
    setFormError('')
    try {
      await usePatientStore.getState().createPatient({
        name: data.name,
        phone: data.phone,
        email: data.email || '',
        gender: data.gender,
        age: data.age || 0,
        dateOfBirth: data.dateOfBirth || '',
        address: data.address || '',
        city: data.city || '',
        bloodGroup: data.bloodGroup || '',
        medicalHistory: data.medicalHistory || '',
        allergies: data.allergies || '',
        notes: data.notes || '',
      })
      reset()
      setShowForm(false)
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to create patient'
      setFormError(message)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Patients</h1>
        <Button onClick={() => setShowForm(!showForm)}>
          <Plus className="h-4 w-4 mr-2" /> New Patient
        </Button>
      </div>

      {/* New Patient Form */}
      {showForm && (
        <Card>
          <CardHeader>
            <CardTitle>Register New Patient</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              {formError && (
                <div className="p-3 text-sm text-red-600 bg-red-50 rounded-md">{formError}</div>
              )}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="space-y-2">
                  <Label>Name *</Label>
                  <Input {...register('name')} placeholder="Patient full name" />
                  {errors.name && <p className="text-sm text-red-500">{errors.name.message}</p>}
                </div>
                <div className="space-y-2">
                  <Label>Phone *</Label>
                  <Input {...register('phone')} placeholder="10-digit mobile" />
                  {errors.phone && <p className="text-sm text-red-500">{errors.phone.message}</p>}
                </div>
                <div className="space-y-2">
                  <Label>Gender *</Label>
                  <select {...register('gender')} className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                    {GENDER_OPTIONS.map(opt => (
                      <option key={opt.value} value={opt.value}>{opt.label}</option>
                    ))}
                  </select>
                </div>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <div className="space-y-2">
                  <Label>Age</Label>
                  <Input type="number" {...register('age', { valueAsNumber: true })} />
                </div>
                <div className="space-y-2">
                  <Label>Email</Label>
                  <Input type="email" {...register('email')} />
                </div>
                <div className="space-y-2">
                  <Label>Blood Group</Label>
                  <select {...register('bloodGroup')} className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                    {BLOOD_GROUP_OPTIONS.map(opt => (
                      <option key={opt.value} value={opt.value}>{opt.label}</option>
                    ))}
                  </select>
                </div>
                <div className="space-y-2">
                  <Label>State</Label>
                  <select
                    value={selectedState}
                    onChange={(e) => setSelectedState(e.target.value)}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  >
                    <option value="">Select state...</option>
                    {INDIAN_STATES.map(s => (
                      <option key={s} value={s}>{s}</option>
                    ))}
                  </select>
                </div>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="space-y-2">
                  <Label>City</Label>
                  <select {...register('city')} className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                    <option value="">Select city...</option>
                    {(selectedState ? (INDIAN_CITIES[selectedState] || []) : Object.values(INDIAN_CITIES).flat()).map(c => (
                      <option key={c} value={c}>{c}</option>
                    ))}
                  </select>
                </div>
                <div className="space-y-2">
                  <Label>Medical History</Label>
                  <Input {...register('medicalHistory')} placeholder="Any known conditions" />
                </div>
                <div className="space-y-2">
                  <Label>Allergies</Label>
                  <Input {...register('allergies')} placeholder="Known allergies" />
                </div>
              </div>
              <div className="flex gap-2">
                <Button type="submit">Save Patient</Button>
                <Button type="button" variant="outline" onClick={() => setShowForm(false)}>Cancel</Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
        <Input
          placeholder="Search by name or phone..."
          value={localSearch}
          onChange={(e) => setLocalSearch(e.target.value)}
          className="pl-10"
        />
      </div>

      {/* Patient List */}
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="p-6 text-center text-muted-foreground">Loading...</div>
          ) : patients.length === 0 ? (
            <div className="p-6 text-center text-muted-foreground">
              {searchQuery ? 'No patients found matching your search.' : 'No patients registered yet.'}
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="border-b bg-muted/50">
                  <tr>
                    <th className="px-4 py-3 text-left text-sm font-medium">Name</th>
                    <th className="px-4 py-3 text-left text-sm font-medium">Phone</th>
                    <th className="px-4 py-3 text-left text-sm font-medium">Gender</th>
                    <th className="px-4 py-3 text-left text-sm font-medium">Age</th>
                    <th className="px-4 py-3 text-left text-sm font-medium">City</th>
                    <th className="px-4 py-3 text-left text-sm font-medium">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {patients.map((patient) => (
                    <tr key={patient.id} className="border-b hover:bg-muted/30 cursor-pointer" onClick={() => navigate(`/patients/${patient.id}`)}>
                      <td className="px-4 py-3 text-sm font-medium">{patient.name}</td>
                      <td className="px-4 py-3 text-sm">{patient.phone}</td>
                      <td className="px-4 py-3 text-sm capitalize">{patient.gender}</td>
                      <td className="px-4 py-3 text-sm">{patient.age || '-'}</td>
                      <td className="px-4 py-3 text-sm">{patient.city || '-'}</td>
                      <td className="px-4 py-3 text-sm">
                        <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); navigate(`/patients/${patient.id}`) }}>
                          View
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

      {/* Pagination */}
      {totalCount > 20 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Showing {(page - 1) * 20 + 1} to {Math.min(page * 20, totalCount)} of {totalCount}
          </p>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>Previous</Button>
            <Button variant="outline" size="sm" disabled={page * 20 >= totalCount} onClick={() => setPage(page + 1)}>Next</Button>
          </div>
        </div>
      )}
    </div>
  )
}
