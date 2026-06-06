// Package workschedule models a recurring weekly working pattern (hours per
// weekday) and computes how many hours a given month is expected to contain.
package workschedule

import "time"

// Schedule holds the expected hours for each weekday, indexed by time.Weekday
// (Sunday=0 .. Saturday=6).
type Schedule [7]int

// Default returns the built-in default schedule: 9 hours on Mon/Tue/Wed/Fri,
// 0 on Thu/Sat/Sun (36 hours/week).
func Default() Schedule {
	return Schedule{
		time.Sunday:    0,
		time.Monday:    9,
		time.Tuesday:   9,
		time.Wednesday: 9,
		time.Thursday:  0,
		time.Friday:    9,
		time.Saturday:  0,
	}
}

// IsZero reports whether every weekday is zero. Used to detect an unset
// schedule loaded from config so callers can fall back to Default().
func (s Schedule) IsZero() bool {
	for _, h := range s {
		if h != 0 {
			return false
		}
	}
	return true
}

// WeeklyTotal returns the sum of hours across the week.
func (s Schedule) WeeklyTotal() int {
	total := 0
	for _, h := range s {
		total += h
	}
	return total
}

// ExpectedHoursForMonth walks every day in the given month and sums the
// schedule's hours for each day's weekday. Independent of how time was
// actually logged — this is the target.
func ExpectedHoursForMonth(year int, month time.Month, s Schedule) int {
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local)

	total := 0
	for day := firstDay; !day.After(lastDay); day = day.AddDate(0, 0, 1) {
		total += s[day.Weekday()]
	}
	return total
}
