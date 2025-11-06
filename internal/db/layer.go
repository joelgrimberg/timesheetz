package db

import (
	"time"
)

// DataLayer defines the interface for data access operations
// Both local DB and remote API client implement this interface
type DataLayer interface {
	// Timesheet operations
	GetAllTimesheetEntries(year int, month time.Month) ([]TimesheetEntry, error)
	GetTimesheetEntryByDate(date string) (TimesheetEntry, error)
	AddTimesheetEntry(entry TimesheetEntry) error
	UpdateTimesheetEntry(entry TimesheetEntry) error
	DeleteTimesheetEntryByDate(date string) error
	DeleteTimesheetEntry(id string) error
	GetLastClientName() (string, error)

	// Training operations
	GetTrainingEntriesForYear(year int) ([]TimesheetEntry, error)
	GetVacationEntriesForYear(year int) ([]TimesheetEntry, error)
	GetVacationHoursForYear(year int) (int, error)

	// Training budget operations
	GetTrainingBudgetEntriesForYear(year int) ([]TrainingBudgetEntry, error)
	AddTrainingBudgetEntry(entry TrainingBudgetEntry) error
	UpdateTrainingBudgetEntry(entry TrainingBudgetEntry) error
	DeleteTrainingBudgetEntry(id int) error
	GetTrainingBudgetEntry(id int) (TrainingBudgetEntry, error)
	GetTrainingBudgetEntryByDate(date string) (TrainingBudgetEntry, error)

	// Health check
	Ping() error
}

// LocalDBLayer wraps the existing DB functions to implement DataLayer
type LocalDBLayer struct{}

func (l *LocalDBLayer) GetAllTimesheetEntries(year int, month time.Month) ([]TimesheetEntry, error) {
	return GetAllTimesheetEntries(year, month)
}

func (l *LocalDBLayer) GetTimesheetEntryByDate(date string) (TimesheetEntry, error) {
	return GetTimesheetEntryByDate(date)
}

func (l *LocalDBLayer) AddTimesheetEntry(entry TimesheetEntry) error {
	return AddTimesheetEntry(entry)
}

func (l *LocalDBLayer) UpdateTimesheetEntry(entry TimesheetEntry) error {
	return UpdateTimesheetEntry(entry)
}

func (l *LocalDBLayer) DeleteTimesheetEntryByDate(date string) error {
	return DeleteTimesheetEntryByDate(date)
}

func (l *LocalDBLayer) DeleteTimesheetEntry(id string) error {
	return DeleteTimesheetEntry(id)
}

func (l *LocalDBLayer) GetLastClientName() (string, error) {
	return GetLastClientName()
}

func (l *LocalDBLayer) GetTrainingEntriesForYear(year int) ([]TimesheetEntry, error) {
	return GetTrainingEntriesForYear(year)
}

func (l *LocalDBLayer) GetVacationEntriesForYear(year int) ([]TimesheetEntry, error) {
	return GetVacationEntriesForYear(year)
}

func (l *LocalDBLayer) GetVacationHoursForYear(year int) (int, error) {
	return GetVacationHoursForYear(year)
}

func (l *LocalDBLayer) GetTrainingBudgetEntriesForYear(year int) ([]TrainingBudgetEntry, error) {
	return GetTrainingBudgetEntriesForYear(year)
}

func (l *LocalDBLayer) AddTrainingBudgetEntry(entry TrainingBudgetEntry) error {
	return AddTrainingBudgetEntry(entry)
}

func (l *LocalDBLayer) UpdateTrainingBudgetEntry(entry TrainingBudgetEntry) error {
	return UpdateTrainingBudgetEntry(entry)
}

func (l *LocalDBLayer) DeleteTrainingBudgetEntry(id int) error {
	return DeleteTrainingBudgetEntry(id)
}

func (l *LocalDBLayer) GetTrainingBudgetEntry(id int) (TrainingBudgetEntry, error) {
	return GetTrainingBudgetEntry(id)
}

func (l *LocalDBLayer) GetTrainingBudgetEntryByDate(date string) (TrainingBudgetEntry, error) {
	return GetTrainingBudgetEntryByDate(date)
}

func (l *LocalDBLayer) Ping() error {
	return Ping()
}

