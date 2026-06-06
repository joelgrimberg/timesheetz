package workschedule

import (
	"testing"
	"time"
)

func TestDefault_WeeklyTotalIs36(t *testing.T) {
	if got := Default().WeeklyTotal(); got != 36 {
		t.Errorf("Default().WeeklyTotal() = %d, want 36", got)
	}
}

func TestIsZero(t *testing.T) {
	if !(Schedule{}).IsZero() {
		t.Error("empty Schedule should be Zero")
	}
	if Default().IsZero() {
		t.Error("Default schedule should not be Zero")
	}
}

func TestExpectedHoursForMonth(t *testing.T) {
	s := Default()

	tests := []struct {
		name  string
		year  int
		month time.Month
		want  int
	}{
		// June 2026 starts on a Monday, 30 days.
		// Mon ×5 + Tue ×5 + Wed ×4 + Fri ×4 = 18 working days × 9 = 162.
		{"June 2026 (Mon start, 30 days)", 2026, time.June, 162},

		// May 2026 starts on Friday, 31 days.
		// Friday-start case the user called out.
		// Fri: 1, 8, 15, 22, 29 = 5
		// Sat/Sun off
		// Mon: 4, 11, 18, 25 = 4
		// Tue: 5, 12, 19, 26 = 4
		// Wed: 6, 13, 20, 27 = 4
		// Thu off
		// Total: (5+4+4+4) × 9 = 17 × 9 = 153.
		{"May 2026 (Fri start, 31 days)", 2026, time.May, 153},

		// February 2026: 28 days, starts on Sunday.
		// Mon: 2,9,16,23 = 4
		// Tue: 3,10,17,24 = 4
		// Wed: 4,11,18,25 = 4
		// Fri: 6,13,20,27 = 4
		// = 16 × 9 = 144.
		{"February 2026 (Sun start, 28 days)", 2026, time.February, 144},

		// October 2026 starts on Thursday (non-working day).
		// Thu off, Fri 2,9,16,23,30 = 5
		// Mon: 5,12,19,26 = 4; Tue: 6,13,20,27 = 4; Wed: 7,14,21,28 = 4
		// = (5+4+4+4) × 9 = 17 × 9 = 153.
		{"October 2026 (Thu start, 31 days)", 2026, time.October, 153},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpectedHoursForMonth(tt.year, tt.month, s)
			if got != tt.want {
				t.Errorf("ExpectedHoursForMonth(%d, %s) = %d, want %d",
					tt.year, tt.month, got, tt.want)
			}
		})
	}
}

func TestExpectedHoursForMonth_CustomSchedule(t *testing.T) {
	// 40-hour week: 8 per day Mon–Fri.
	s := Schedule{
		time.Monday:    8,
		time.Tuesday:   8,
		time.Wednesday: 8,
		time.Thursday:  8,
		time.Friday:    8,
	}

	// June 2026: weekdays Mon–Fri.
	// Mon ×5 + Tue ×5 + Wed ×4 + Thu ×4 + Fri ×4 = 22 × 8 = 176.
	got := ExpectedHoursForMonth(2026, time.June, s)
	if got != 176 {
		t.Errorf("40h-week schedule on June 2026 = %d, want 176", got)
	}
}
