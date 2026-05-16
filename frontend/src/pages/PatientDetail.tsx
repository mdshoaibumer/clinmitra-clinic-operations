import { useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { usePatientStore } from '@/store/patientStore'
import { formatDate, formatCurrency } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import PatientAvatar from '@/components/PatientAvatar'
import { ArrowLeft, Receipt, Calendar } from 'lucide-react'

export default function PatientDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { currentPatient, patientHistory, isLoading, fetchPatient, fetchPatientHistory } = usePatientStore()

  useEffect(() => {
    if (id) {
      fetchPatient(id)
      fetchPatientHistory(id)
    }
  }, [id, fetchPatient, fetchPatientHistory])

  if (isLoading || !currentPatient) {
    return <div className="text-muted-foreground">Loading patient details...</div>
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" onClick={() => navigate('/patients')}>
          <ArrowLeft className="h-5 w-5" />
        </Button>
        <PatientAvatar name={currentPatient.name} size="lg" />
        <div>
          <h1 className="text-2xl font-bold">{currentPatient.name}</h1>
          <p className="text-sm text-muted-foreground">Registered {formatDate(currentPatient.createdAt)}</p>
        </div>
      </div>

      {/* Patient Info */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Personal Information</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <InfoRow label="Phone" value={currentPatient.phone} />
            <InfoRow label="Email" value={currentPatient.email} />
            <InfoRow label="Gender" value={currentPatient.gender} />
            <InfoRow label="Age" value={currentPatient.age ? `${currentPatient.age} years` : '-'} />
            <InfoRow label="Blood Group" value={currentPatient.bloodGroup} />
            <InfoRow label="City" value={currentPatient.city} />
            <InfoRow label="Address" value={currentPatient.address} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Medical Information</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <InfoRow label="Medical History" value={currentPatient.medicalHistory} />
            <InfoRow label="Allergies" value={currentPatient.allergies} />
            <InfoRow label="Notes" value={currentPatient.notes} />
            <InfoRow label="Registered" value={formatDate(currentPatient.createdAt)} />
          </CardContent>
        </Card>
      </div>

      {/* Quick Actions */}
      <div className="flex gap-3">
        <Button onClick={() => navigate(`/billing?patientId=${currentPatient.id}`)}>
          <Receipt className="h-4 w-4 mr-2" /> Create Invoice
        </Button>
        <Button variant="outline" onClick={() => navigate(`/appointments?patientId=${currentPatient.id}`)}>
          <Calendar className="h-4 w-4 mr-2" /> Book Appointment
        </Button>
      </div>

      {/* Treatment History */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Treatment History</CardTitle>
        </CardHeader>
        <CardContent>
          {patientHistory.length === 0 ? (
            <p className="text-muted-foreground text-sm">No treatment history yet.</p>
          ) : (
            <div className="space-y-2">
              {patientHistory.map((record) => (
                <div key={record.id} className="flex items-center justify-between p-3 bg-muted rounded-md">
                  <div>
                    <p className="font-medium text-sm">{record.treatment?.name || 'Treatment'}</p>
                    <p className="text-xs text-muted-foreground">
                      {formatDate(record.treatmentDate)}
                      {record.toothNumber && ` • Tooth: ${record.toothNumber}`}
                    </p>
                  </div>
                  <div className="text-right text-sm">
                    {record.treatment && formatCurrency(record.treatment.defaultPrice)}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: string | undefined }) {
  return (
    <div className="flex justify-between text-sm">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-medium capitalize">{value || '-'}</span>
    </div>
  )
}
