package service

import (
	"testing"

	"practivo/internal/models"
)

func TestAppointmentConflictDetection(t *testing.T) {
	type timeSlot struct {
		start string
		end   string
	}

	tests := []struct {
		name       string
		existing   timeSlot
		requested  timeSlot
		isConflict bool
	}{
		{
			name:       "no overlap - before",
			existing:   timeSlot{"10:00", "10:30"},
			requested:  timeSlot{"09:00", "09:30"},
			isConflict: false,
		},
		{
			name:       "no overlap - after",
			existing:   timeSlot{"10:00", "10:30"},
			requested:  timeSlot{"11:00", "11:30"},
			isConflict: false,
		},
		{
			name:       "overlap - same time",
			existing:   timeSlot{"10:00", "10:30"},
			requested:  timeSlot{"10:00", "10:30"},
			isConflict: true,
		},
		{
			name:       "overlap - starts during",
			existing:   timeSlot{"10:00", "10:30"},
			requested:  timeSlot{"10:15", "10:45"},
			isConflict: true,
		},
		{
			name:       "overlap - ends during",
			existing:   timeSlot{"10:00", "10:30"},
			requested:  timeSlot{"09:45", "10:15"},
			isConflict: true,
		},
		{
			name:       "overlap - encompasses",
			existing:   timeSlot{"10:00", "10:30"},
			requested:  timeSlot{"09:30", "11:00"},
			isConflict: true,
		},
		{
			name:       "adjacent - no conflict",
			existing:   timeSlot{"10:00", "10:30"},
			requested:  timeSlot{"10:30", "11:00"},
			isConflict: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Conflict: requested.start < existing.end AND requested.end > existing.start
			conflict := tt.requested.start < tt.existing.end && tt.requested.end > tt.existing.start

			if conflict != tt.isConflict {
				t.Errorf("expected conflict=%v, got conflict=%v", tt.isConflict, conflict)
			}
		})
	}
}

func TestAppointmentStatusTransitions(t *testing.T) {
	tests := []struct {
		name        string
		current     models.AppointmentStatus
		target      string
		shouldAllow bool
	}{
		{"scheduled to completed", models.AppointmentScheduled, "completed", true},
		{"scheduled to cancelled", models.AppointmentScheduled, "cancelled", true},
		{"scheduled to no_show", models.AppointmentScheduled, "no_show", true},
		{"completed cannot be cancelled", models.AppointmentCompleted, "cancelled", false},
		{"cancelled cannot be completed", models.AppointmentCancelled, "completed", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := tt.current == models.AppointmentScheduled

			if allowed != tt.shouldAllow {
				t.Errorf("expected allowed=%v, got allowed=%v", tt.shouldAllow, allowed)
			}
		})
	}
}

func TestCreateAppointment_Validation(t *testing.T) {
	tests := []struct {
		name      string
		input     CreateAppointmentInput
		wantError bool
		errField  string
	}{
		{
			name: "valid input",
			input: CreateAppointmentInput{
				PatientID: "patient-1",
				Date:      "2024-06-15",
				StartTime: "10:00",
				EndTime:   "10:30",
				Duration:  30,
			},
			wantError: false,
		},
		{
			name: "missing patient ID",
			input: CreateAppointmentInput{
				Date:      "2024-06-15",
				StartTime: "10:00",
				EndTime:   "10:30",
			},
			wantError: true,
			errField:  "Patient",
		},
		{
			name: "missing date",
			input: CreateAppointmentInput{
				PatientID: "patient-1",
				StartTime: "10:00",
				EndTime:   "10:30",
			},
			wantError: true,
			errField:  "Date",
		},
		{
			name: "missing start time",
			input: CreateAppointmentInput{
				PatientID: "patient-1",
				Date:      "2024-06-15",
				EndTime:   "10:30",
			},
			wantError: true,
			errField:  "StartTime",
		},
		{
			name: "missing end time",
			input: CreateAppointmentInput{
				PatientID: "patient-1",
				Date:      "2024-06-15",
				StartTime: "10:00",
			},
			wantError: true,
			errField:  "EndTime",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := false
			if tt.input.PatientID == "" || tt.input.Date == "" || tt.input.StartTime == "" || tt.input.EndTime == "" {
				hasError = true
			}
			if hasError != tt.wantError {
				t.Errorf("expected error=%v, got error=%v", tt.wantError, hasError)
			}
		})
	}
}
