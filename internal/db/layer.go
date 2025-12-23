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

	// Vacation carryover operations
	GetVacationCarryoverForYear(year int) (VacationCarryover, error)
	SetVacationCarryover(carryover VacationCarryover) error
	DeleteVacationCarryover(year int) error
	GetVacationSummaryForYear(year int) (VacationSummary, error)

	// Training budget operations
	GetTrainingBudgetEntriesForYear(year int) ([]TrainingBudgetEntry, error)
	AddTrainingBudgetEntry(entry TrainingBudgetEntry) error
	UpdateTrainingBudgetEntry(entry TrainingBudgetEntry) error
	DeleteTrainingBudgetEntry(id int) error
	GetTrainingBudgetEntry(id int) (TrainingBudgetEntry, error)
	GetTrainingBudgetEntryByDate(date string) (TrainingBudgetEntry, error)

	// Client operations
	GetAllClients() ([]Client, error)
	GetActiveClients() ([]Client, error)
	GetClientById(id int) (Client, error)
	GetClientByName(name string) (Client, error)
	AddClient(client Client) (int, error)
	UpdateClient(client Client) error
	DeleteClient(id int) error
	DeactivateClient(id int) error

	// Client rate operations
	GetClientRates(clientId int) ([]ClientRate, error)
	GetClientRateById(id int) (ClientRate, error)
	AddClientRate(rate ClientRate) error
	UpdateClientRate(rate ClientRate) error
	DeleteClientRate(id int) error
	GetClientRateForDate(clientId int, date string) (ClientRate, error)
	GetClientRateByName(clientName string, date string) (float64, error)

	// Earnings operations
	CalculateEarningsForYear(year int) (EarningsOverview, error)
	CalculateEarningsSummaryForYear(year int) (EarningsOverview, error)
	CalculateEarningsForMonth(year int, month int) (EarningsOverview, error)
	GetClientWithRates(clientId int) (ClientWithRates, error)

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

func (l *LocalDBLayer) GetVacationCarryoverForYear(year int) (VacationCarryover, error) {
	return GetVacationCarryoverForYear(year)
}

func (l *LocalDBLayer) SetVacationCarryover(carryover VacationCarryover) error {
	return SetVacationCarryover(carryover)
}

func (l *LocalDBLayer) DeleteVacationCarryover(year int) error {
	return DeleteVacationCarryover(year)
}

func (l *LocalDBLayer) GetVacationSummaryForYear(year int) (VacationSummary, error) {
	return GetVacationSummaryForYear(year)
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

// Client operations

func (l *LocalDBLayer) GetAllClients() ([]Client, error) {
	return GetAllClients()
}

func (l *LocalDBLayer) GetActiveClients() ([]Client, error) {
	return GetActiveClients()
}

func (l *LocalDBLayer) GetClientById(id int) (Client, error) {
	return GetClientById(id)
}

func (l *LocalDBLayer) GetClientByName(name string) (Client, error) {
	return GetClientByName(name)
}

func (l *LocalDBLayer) AddClient(client Client) (int, error) {
	return AddClient(client)
}

func (l *LocalDBLayer) UpdateClient(client Client) error {
	return UpdateClient(client)
}

func (l *LocalDBLayer) DeleteClient(id int) error {
	return DeleteClient(id)
}

func (l *LocalDBLayer) DeactivateClient(id int) error {
	return DeactivateClient(id)
}

// Client rate operations

func (l *LocalDBLayer) GetClientRates(clientId int) ([]ClientRate, error) {
	return GetClientRates(clientId)
}

func (l *LocalDBLayer) GetClientRateById(id int) (ClientRate, error) {
	return GetClientRateById(id)
}

func (l *LocalDBLayer) AddClientRate(rate ClientRate) error {
	return AddClientRate(rate)
}

func (l *LocalDBLayer) UpdateClientRate(rate ClientRate) error {
	return UpdateClientRate(rate)
}

func (l *LocalDBLayer) DeleteClientRate(id int) error {
	return DeleteClientRate(id)
}

func (l *LocalDBLayer) GetClientRateForDate(clientId int, date string) (ClientRate, error) {
	return GetClientRateForDate(clientId, date)
}

func (l *LocalDBLayer) GetClientRateByName(clientName string, date string) (float64, error) {
	return GetClientRateByName(clientName, date)
}

// Earnings operations

func (l *LocalDBLayer) CalculateEarningsForYear(year int) (EarningsOverview, error) {
	return CalculateEarningsForYear(year)
}

func (l *LocalDBLayer) CalculateEarningsSummaryForYear(year int) (EarningsOverview, error) {
	return CalculateEarningsSummaryForYear(year)
}

func (l *LocalDBLayer) CalculateEarningsForMonth(year int, month int) (EarningsOverview, error) {
	return CalculateEarningsForMonth(year, month)
}

func (l *LocalDBLayer) GetClientWithRates(clientId int) (ClientWithRates, error) {
	return GetClientWithRates(clientId)
}

