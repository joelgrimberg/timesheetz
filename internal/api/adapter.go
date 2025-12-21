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

// Client operations

func (a *ClientAdapter) GetAllClients() ([]db.Client, error) {
	return a.client.GetAllClients()
}

func (a *ClientAdapter) GetActiveClients() ([]db.Client, error) {
	return a.client.GetActiveClients()
}

func (a *ClientAdapter) GetClientById(id int) (db.Client, error) {
	return a.client.GetClientById(id)
}

func (a *ClientAdapter) GetClientByName(name string) (db.Client, error) {
	return a.client.GetClientByName(name)
}

func (a *ClientAdapter) AddClient(client db.Client) (int, error) {
	return a.client.AddClient(client)
}

func (a *ClientAdapter) UpdateClient(client db.Client) error {
	return a.client.UpdateClient(client)
}

func (a *ClientAdapter) DeleteClient(id int) error {
	return a.client.DeleteClient(id)
}

func (a *ClientAdapter) DeactivateClient(id int) error {
	return a.client.DeactivateClient(id)
}

// Client rate operations

func (a *ClientAdapter) GetClientRates(clientId int) ([]db.ClientRate, error) {
	return a.client.GetClientRates(clientId)
}

func (a *ClientAdapter) GetClientRateById(id int) (db.ClientRate, error) {
	return a.client.GetClientRateById(id)
}

func (a *ClientAdapter) AddClientRate(rate db.ClientRate) error {
	return a.client.AddClientRate(rate)
}

func (a *ClientAdapter) UpdateClientRate(rate db.ClientRate) error {
	return a.client.UpdateClientRate(rate)
}

func (a *ClientAdapter) DeleteClientRate(id int) error {
	return a.client.DeleteClientRate(id)
}

func (a *ClientAdapter) GetClientRateForDate(clientId int, date string) (db.ClientRate, error) {
	return a.client.GetClientRateForDate(clientId, date)
}

func (a *ClientAdapter) GetClientRateByName(clientName string, date string) (float64, error) {
	return a.client.GetClientRateByName(clientName, date)
}

// Earnings operations

func (a *ClientAdapter) CalculateEarningsForYear(year int) (db.EarningsOverview, error) {
	return a.client.CalculateEarningsForYear(year)
}

func (a *ClientAdapter) CalculateEarningsSummaryForYear(year int) (db.EarningsOverview, error) {
	return a.client.CalculateEarningsSummaryForYear(year)
}

func (a *ClientAdapter) CalculateEarningsForMonth(year int, month int) (db.EarningsOverview, error) {
	return a.client.CalculateEarningsForMonth(year, month)
}

func (a *ClientAdapter) GetClientWithRates(clientId int) (db.ClientWithRates, error) {
	return a.client.GetClientWithRates(clientId)
}

