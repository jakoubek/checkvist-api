package checkvist

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAPITime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, got APITime)
	}{
		{
			name:    "Checkvist API format",
			input:   `"2026/01/14 16:07:31 +0000"`,
			wantErr: false,
			check: func(t *testing.T, got APITime) {
				if got.Year() != 2026 || got.Month() != 1 || got.Day() != 14 {
					t.Errorf("date mismatch: got %v", got.Time)
				}
				if got.Hour() != 16 || got.Minute() != 7 || got.Second() != 31 {
					t.Errorf("time mismatch: got %v", got.Time)
				}
			},
		},
		{
			name:    "RFC3339 format",
			input:   `"2026-01-14T10:00:00Z"`,
			wantErr: false,
			check: func(t *testing.T, got APITime) {
				if got.Year() != 2026 || got.Month() != 1 || got.Day() != 14 {
					t.Errorf("date mismatch: got %v", got.Time)
				}
				if got.Hour() != 10 || got.Minute() != 0 {
					t.Errorf("time mismatch: got %v", got.Time)
				}
			},
		},
		{
			name:    "RFC3339 with timezone offset",
			input:   `"2026-01-14T10:00:00+02:00"`,
			wantErr: false,
			check: func(t *testing.T, got APITime) {
				if got.Year() != 2026 || got.Month() != 1 || got.Day() != 14 {
					t.Errorf("date mismatch: got %v", got.Time)
				}
			},
		},
		{
			name:    "empty string",
			input:   `""`,
			wantErr: false,
			check: func(t *testing.T, got APITime) {
				if !got.IsZero() {
					t.Error("expected zero time for empty string")
				}
			},
		},
		{
			name:    "null",
			input:   `null`,
			wantErr: false,
			check: func(t *testing.T, got APITime) {
				if !got.IsZero() {
					t.Error("expected zero time for null")
				}
			},
		},
		{
			name:    "invalid format",
			input:   `"not a date"`,
			wantErr: true,
			check:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got APITime
			err := json.Unmarshal([]byte(tt.input), &got)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestAPITime_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    APITime
		expected string
	}{
		{
			name:     "normal time",
			input:    NewAPITime(time.Date(2026, 1, 14, 10, 30, 0, 0, time.UTC)),
			expected: `"2026-01-14T10:30:00Z"`,
		},
		{
			name:     "zero time",
			input:    APITime{},
			expected: `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(got) != tt.expected {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestAPITime_InStruct(t *testing.T) {
	// Test unmarshaling a struct with APITime fields using real API format
	jsonData := `{
		"id": 1,
		"name": "Test Checklist",
		"updated_at": "2026/01/14 16:07:31 +0000"
	}`

	type testStruct struct {
		ID        int     `json:"id"`
		Name      string  `json:"name"`
		UpdatedAt APITime `json:"updated_at"`
	}

	var result testStruct
	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if result.UpdatedAt.Year() != 2026 {
		t.Errorf("expected year 2026, got %d", result.UpdatedAt.Year())
	}
	if result.UpdatedAt.Month() != 1 {
		t.Errorf("expected month 1, got %d", result.UpdatedAt.Month())
	}
	if result.UpdatedAt.Day() != 14 {
		t.Errorf("expected day 14, got %d", result.UpdatedAt.Day())
	}
}

func TestNewAPITime(t *testing.T) {
	now := time.Now()
	apiTime := NewAPITime(now)

	if !apiTime.Time.Equal(now) {
		t.Errorf("NewAPITime() did not preserve time: got %v, want %v", apiTime.Time, now)
	}
}
