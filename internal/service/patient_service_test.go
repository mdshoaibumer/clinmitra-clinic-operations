package service

import (
	"testing"

	"practivo/internal/utils"
)

func TestCreatePatient_Validation(t *testing.T) {
	tests := []struct {
		name      string
		input     CreatePatientInput
		wantError bool
		errMsg    string
	}{
		{
			name: "valid patient",
			input: CreatePatientInput{
				Name:   "Ramesh Kumar",
				Phone:  "9876543210",
				Gender: "male",
				Age:    35,
			},
			wantError: false,
		},
		{
			name: "missing name",
			input: CreatePatientInput{
				Phone:  "9876543210",
				Gender: "male",
			},
			wantError: true,
			errMsg:    "Name is required",
		},
		{
			name: "name too short",
			input: CreatePatientInput{
				Name:   "R",
				Phone:  "9876543210",
				Gender: "male",
			},
			wantError: true,
			errMsg:    "at least 2",
		},
		{
			name: "missing phone",
			input: CreatePatientInput{
				Name:   "Ramesh Kumar",
				Gender: "male",
			},
			wantError: true,
			errMsg:    "Phone is required",
		},
		{
			name: "invalid phone - too short",
			input: CreatePatientInput{
				Name:   "Ramesh Kumar",
				Phone:  "12345",
				Gender: "male",
			},
			wantError: true,
			errMsg:    "Invalid Indian phone",
		},
		{
			name: "invalid phone - starts with 5",
			input: CreatePatientInput{
				Name:   "Ramesh Kumar",
				Phone:  "5876543210",
				Gender: "male",
			},
			wantError: true,
			errMsg:    "Invalid Indian phone",
		},
		{
			name: "invalid age - negative",
			input: CreatePatientInput{
				Name:   "Ramesh Kumar",
				Phone:  "9876543210",
				Gender: "male",
				Age:    -5,
			},
			wantError: true,
			errMsg:    "Age",
		},
		{
			name: "invalid age - too old",
			input: CreatePatientInput{
				Name:   "Ramesh Kumar",
				Phone:  "9876543210",
				Gender: "male",
				Age:    130,
			},
			wantError: true,
			errMsg:    "Age",
		},
		{
			name: "valid with +91 prefix phone",
			input: CreatePatientInput{
				Name:   "Suresh Patel",
				Phone:  "+919876543210",
				Gender: "male",
				Age:    45,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePatientInput(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("expected error=%v, got error=%v (err: %v)", tt.wantError, err != nil, err)
			}
		})
	}
}

// validatePatientInput mirrors the validation logic in PatientService.CreatePatient
func validatePatientInput(input CreatePatientInput) error {
	if err := utils.ValidateRequired("Name", input.Name); err != nil {
		return err
	}
	if err := utils.ValidateRequired("Phone", input.Phone); err != nil {
		return err
	}
	if err := utils.ValidateMinLength("Name", input.Name, 2); err != nil {
		return err
	}

	cleanedPhone := utils.CleanPhone(input.Phone)
	if err := utils.ValidatePhone(cleanedPhone); err != nil {
		return err
	}

	if input.Age != 0 {
		if err := utils.ValidateAge(input.Age); err != nil {
			return err
		}
	}

	return nil
}

func TestPhoneCleaningAndValidation(t *testing.T) {
	tests := []struct {
		name      string
		phone     string
		wantError bool
	}{
		{"valid 10 digit", "9876543210", false},
		{"with +91", "+919876543210", false},
		{"with 91", "919876543210", false},
		{"with spaces", "98765 43210", false},
		{"starts with 6", "6876543210", false},
		{"starts with 7", "7876543210", false},
		{"starts with 8", "8876543210", false},
		{"starts with 5 invalid", "5876543210", true},
		{"too short", "98765", true},
		{"too long", "98765432101", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaned := utils.CleanPhone(tt.phone)
			err := utils.ValidatePhone(cleaned)
			if (err != nil) != tt.wantError {
				t.Errorf("phone=%q: expected error=%v, got error=%v (cleaned=%q)", tt.phone, tt.wantError, err != nil, cleaned)
			}
		})
	}
}
