package api

import (
	"time"
	"timesheet/internal/db"
)

// ClientAdapter adapts the API Client to implement the DataLayer interface
type ClientAdapter struct {
	client *Client
}

// NewClientAdapter creates a new adapter for the API client
func NewClientAdapter(client *Client) *ClientAdapter {
	return &ClientAdapter{client: client}
}

func (a *ClientAdapter) GetAllTimesheetEntries(year int, month time.Month) ([]db.TimesheetEntry, error) {
	return a.client.GetAllTimesheetEntries(year, month)
}

func (a *ClientAdapter) GetTimesheetEntryByDate(date string) (db.TimesheetEntry, error) {
	return a.client.GetTimesheetEntryByDate(date)
}

func (a *ClientAdapter) AddTimesheetEntry(entry db.TimesheetEntry) error {
	return a.client.AddTimesheetEntry(entry)
}

func (a *ClientAdapter) UpdateTimesheetEntry(entry db.TimesheetEntry) error {
	return a.client.UpdateTimesheetEntry(entry)
}

func (a *ClientAdapter) DeleteTimesheetEntryByDate(date string) error {
	return a.client.DeleteTimesheetEntryByDate(date)
}

func (a *ClientAdapter) DeleteTimesheetEntry(id string) error {
	return a.client.DeleteTimesheetEntry(id)
}

func (a *ClientAdapter) GetLastClientName() (string, error) {
	return a.client.GetLastClientName()
}

func (a *ClientAdapter) GetTrainingEntriesForYear(year int) ([]db.TimesheetEntry, error) {
	return a.client.GetTrainingEntriesForYear(year)
}

func (a *ClientAdapter) GetVacationEntriesForYear(year int) ([]db.TimesheetEntry, error) {
	return a.client.GetVacationEntriesForYear(year)
}

func (a *ClientAdapter) GetVacationHoursForYear(year int) (int, error) {
	return a.client.GetVacationHoursForYear(year)
}

func (a *ClientAdapter) GetTrainingBudgetEntriesForYear(year int) ([]db.TrainingBudgetEntry, error) {
	return a.client.GetTrainingBudgetEntriesForYear(year)
}

func (a *ClientAdapter) AddTrainingBudgetEntry(entry db.TrainingBudgetEntry) error {
	return a.client.AddTrainingBudgetEntry(entry)
}

func (a *ClientAdapter) UpdateTrainingBudgetEntry(entry db.TrainingBudgetEntry) error {
	return a.client.UpdateTrainingBudgetEntry(entry)
}

func (a *ClientAdapter) DeleteTrainingBudgetEntry(id int) error {
	return a.client.DeleteTrainingBudgetEntry(id)
}

func (a *ClientAdapter) GetTrainingBudgetEntry(id int) (db.TrainingBudgetEntry, error) {
	return a.client.GetTrainingBudgetEntry(id)
}

func (a *ClientAdapter) GetTrainingBudgetEntryByDate(date string) (db.TrainingBudgetEntry, error) {
	return a.client.GetTrainingBudgetEntryByDate(date)
}

func (a *ClientAdapter) Ping() error {
	return a.client.Ping()
}

