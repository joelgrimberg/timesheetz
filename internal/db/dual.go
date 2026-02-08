package db

import (
	"fmt"
	"reflect"
	"time"
	"timesheet/internal/logging"
)

// DualLayer implements DataLayer by coordinating both local DB and remote API
// In dual mode, writes go to both, reads are compared for validation
type DualLayer struct {
	local  DataLayer
	remote DataLayer
}

// NewDualLayer creates a new dual mode data layer
func NewDualLayer(local DataLayer, remote DataLayer) *DualLayer {
	return &DualLayer{
		local:  local,
		remote: remote,
	}
}

// compareEntries compares two slices of entries and logs differences
func (d *DualLayer) compareEntries(local, remote []TimesheetEntry, operation string) {
	if len(local) != len(remote) {
		logging.Log("DUAL MODE: %s - Entry count mismatch: local=%d, remote=%d", operation, len(local), len(remote))
		return
	}

	for i := range local {
		if !reflect.DeepEqual(local[i], remote[i]) {
			logging.Log("DUAL MODE: %s - Entry mismatch at index %d: local=%+v, remote=%+v", operation, i, local[i], remote[i])
		}
	}
}

// compareTrainingBudgetEntries compares two slices of training budget entries
func (d *DualLayer) compareTrainingBudgetEntries(local, remote []TrainingBudgetEntry, operation string) {
	if len(local) != len(remote) {
		logging.Log("DUAL MODE: %s - Training budget entry count mismatch: local=%d, remote=%d", operation, len(local), len(remote))
		return
	}

	for i := range local {
		if !reflect.DeepEqual(local[i], remote[i]) {
			logging.Log("DUAL MODE: %s - Training budget entry mismatch at index %d: local=%+v, remote=%+v", operation, i, local[i], remote[i])
		}
	}
}

// GetAllTimesheetEntries reads from both sources and compares
func (d *DualLayer) GetAllTimesheetEntries(year int, month time.Month) ([]TimesheetEntry, error) {
	localEntries, localErr := d.local.GetAllTimesheetEntries(year, month)
	remoteEntries, remoteErr := d.remote.GetAllTimesheetEntries(year, month)

	// If both succeed, compare
	if localErr == nil && remoteErr == nil {
		d.compareEntries(localEntries, remoteEntries, "GetAllTimesheetEntries")
		// Return local entries (primary source)
		return localEntries, nil
	}

	// If only one succeeds, log warning and return that one
	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteEntries, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localEntries, nil
	}

	// Both failed
	return nil, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

// GetTimesheetEntryByDate reads from both sources and compares
func (d *DualLayer) GetTimesheetEntryByDate(date string) (TimesheetEntry, error) {
	localEntry, localErr := d.local.GetTimesheetEntryByDate(date)
	remoteEntry, remoteErr := d.remote.GetTimesheetEntryByDate(date)

	// If both succeed, compare
	if localErr == nil && remoteErr == nil {
		if !reflect.DeepEqual(localEntry, remoteEntry) {
			logging.Log("DUAL MODE: GetTimesheetEntryByDate - Entry mismatch for date %s: local=%+v, remote=%+v", date, localEntry, remoteEntry)
		}
		return localEntry, nil
	}

	// If only one succeeds, log warning and return that one
	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteEntry, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localEntry, nil
	}

	// Both failed
	return TimesheetEntry{}, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

// AddTimesheetEntry writes to both sources
func (d *DualLayer) AddTimesheetEntry(entry TimesheetEntry) error {
	logging.Log("DUAL MODE: AddTimesheetEntry - Writing to BOTH local DB and remote API...")
	localErr := d.local.AddTimesheetEntry(entry)
	remoteErr := d.remote.AddTimesheetEntry(entry)

	if localErr != nil {
		logging.Log("DUAL MODE: Local DB write failed: %v", localErr)
	} else {
		logging.Log("DUAL MODE: Local DB write succeeded")
	}
	if remoteErr != nil {
		logging.Log("DUAL MODE: Remote API write failed: %v", remoteErr)
	} else {
		logging.Log("DUAL MODE: Remote API write succeeded")
	}

	// If both fail, return error
	if localErr != nil && remoteErr != nil {
		return fmt.Errorf("both local and remote writes failed: local=%v, remote=%v", localErr, remoteErr)
	}

	// If at least one succeeds, validate by reading back
	if localErr == nil && remoteErr == nil {
		// Read back from both to validate
		localRead, _ := d.local.GetTimesheetEntryByDate(entry.Date)
		remoteRead, _ := d.remote.GetTimesheetEntryByDate(entry.Date)
		if !reflect.DeepEqual(localRead, remoteRead) {
			logging.Log("DUAL MODE: AddTimesheetEntry validation failed - entries differ after write")
		}
	}

	// Return local error if it exists, otherwise remote error (or nil)
	if localErr != nil {
		return fmt.Errorf("local write failed: %w", localErr)
	}
	return remoteErr
}

// UpdateTimesheetEntry writes to both sources
func (d *DualLayer) UpdateTimesheetEntry(entry TimesheetEntry) error {
	localErr := d.local.UpdateTimesheetEntry(entry)
	remoteErr := d.remote.UpdateTimesheetEntry(entry)

	if localErr != nil {
		logging.Log("DUAL MODE: Local DB update failed: %v", localErr)
	}
	if remoteErr != nil {
		logging.Log("DUAL MODE: Remote API update failed: %v", remoteErr)
	}

	// If both fail, return error
	if localErr != nil && remoteErr != nil {
		return fmt.Errorf("both local and remote updates failed: local=%v, remote=%v", localErr, remoteErr)
	}

	// If at least one succeeds, validate by reading back
	if localErr == nil && remoteErr == nil {
		localRead, _ := d.local.GetTimesheetEntryByDate(entry.Date)
		remoteRead, _ := d.remote.GetTimesheetEntryByDate(entry.Date)
		if !reflect.DeepEqual(localRead, remoteRead) {
			logging.Log("DUAL MODE: UpdateTimesheetEntry validation failed - entries differ after update")
		}
	}

	// Return local error if it exists, otherwise remote error (or nil)
	if localErr != nil {
		return fmt.Errorf("local update failed: %w", localErr)
	}
	return remoteErr
}

// UpdateTimesheetEntryById writes to both sources
func (d *DualLayer) UpdateTimesheetEntryById(id string, data map[string]any) error {
	localErr := d.local.UpdateTimesheetEntryById(id, data)
	remoteErr := d.remote.UpdateTimesheetEntryById(id, data)

	if localErr != nil {
		logging.Log("DUAL MODE: Local DB update by ID failed: %v", localErr)
	}
	if remoteErr != nil {
		logging.Log("DUAL MODE: Remote API update by ID failed: %v", remoteErr)
	}

	// If both fail, return error
	if localErr != nil && remoteErr != nil {
		return fmt.Errorf("both local and remote updates failed: local=%v, remote=%v", localErr, remoteErr)
	}

	// Return local error if it exists, otherwise remote error (or nil)
	if localErr != nil {
		return fmt.Errorf("local update failed: %w", localErr)
	}
	return remoteErr
}

// DeleteTimesheetEntryByDate deletes from both sources
func (d *DualLayer) DeleteTimesheetEntryByDate(date string) error {
	localErr := d.local.DeleteTimesheetEntryByDate(date)
	remoteErr := d.remote.DeleteTimesheetEntryByDate(date)

	if localErr != nil {
		logging.Log("DUAL MODE: Local DB delete failed: %v", localErr)
	}
	if remoteErr != nil {
		logging.Log("DUAL MODE: Remote API delete failed: %v", remoteErr)
	}

	// If both fail, return error
	if localErr != nil && remoteErr != nil {
		return fmt.Errorf("both local and remote deletes failed: local=%v, remote=%v", localErr, remoteErr)
	}

	// Return local error if it exists, otherwise remote error (or nil)
	if localErr != nil {
		return fmt.Errorf("local delete failed: %w", localErr)
	}
	return remoteErr
}

// DeleteTimesheetEntry deletes from both sources
func (d *DualLayer) DeleteTimesheetEntry(id string) error {
	localErr := d.local.DeleteTimesheetEntry(id)
	remoteErr := d.remote.DeleteTimesheetEntry(id)

	if localErr != nil {
		logging.Log("DUAL MODE: Local DB delete failed: %v", localErr)
	}
	if remoteErr != nil {
		logging.Log("DUAL MODE: Remote API delete failed: %v", remoteErr)
	}

	// If both fail, return error
	if localErr != nil && remoteErr != nil {
		return fmt.Errorf("both local and remote deletes failed: local=%v, remote=%v", localErr, remoteErr)
	}

	// Return local error if it exists, otherwise remote error (or nil)
	if localErr != nil {
		return fmt.Errorf("local delete failed: %w", localErr)
	}
	return remoteErr
}

// GetLastClientName reads from both sources and compares
func (d *DualLayer) GetLastClientName() (string, error) {
	localName, localErr := d.local.GetLastClientName()
	remoteName, remoteErr := d.remote.GetLastClientName()

	// If both succeed, compare
	if localErr == nil && remoteErr == nil {
		if localName != remoteName {
			logging.Log("DUAL MODE: GetLastClientName - Mismatch: local=%s, remote=%s", localName, remoteName)
		}
		return localName, nil
	}

	// If only one succeeds, log warning and return that one
	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteName, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localName, nil
	}

	// Both failed
	return "", fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

// GetTrainingEntriesForYear reads from both sources and compares
func (d *DualLayer) GetTrainingEntriesForYear(year int) ([]TimesheetEntry, error) {
	localEntries, localErr := d.local.GetTrainingEntriesForYear(year)
	remoteEntries, remoteErr := d.remote.GetTrainingEntriesForYear(year)

	// If both succeed, compare
	if localErr == nil && remoteErr == nil {
		d.compareEntries(localEntries, remoteEntries, "GetTrainingEntriesForYear")
		return localEntries, nil
	}

	// If only one succeeds, log warning and return that one
	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteEntries, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localEntries, nil
	}

	// Both failed
	return nil, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

// GetVacationEntriesForYear reads from both sources and compares
func (d *DualLayer) GetVacationEntriesForYear(year int) ([]TimesheetEntry, error) {
	localEntries, localErr := d.local.GetVacationEntriesForYear(year)
	remoteEntries, remoteErr := d.remote.GetVacationEntriesForYear(year)

	// If both succeed, compare
	if localErr == nil && remoteErr == nil {
		d.compareEntries(localEntries, remoteEntries, "GetVacationEntriesForYear")
		return localEntries, nil
	}

	// If only one succeeds, log warning and return that one
	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteEntries, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localEntries, nil
	}

	// Both failed
	return nil, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

// GetVacationHoursForYear reads from both sources and compares
func (d *DualLayer) GetVacationHoursForYear(year int) (int, error) {
	localHours, localErr := d.local.GetVacationHoursForYear(year)
	remoteHours, remoteErr := d.remote.GetVacationHoursForYear(year)

	// If both succeed, compare
	if localErr == nil && remoteErr == nil {
		if localHours != remoteHours {
			logging.Log("DUAL MODE: GetVacationHoursForYear - Mismatch for year %d: local=%d, remote=%d", year, localHours, remoteHours)
		}
		return localHours, nil
	}

	// If only one succeeds, log warning and return that one
	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteHours, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localHours, nil
	}

	// Both failed
	return 0, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

// GetTrainingBudgetEntriesForYear reads from both sources and compares
func (d *DualLayer) GetTrainingBudgetEntriesForYear(year int) ([]TrainingBudgetEntry, error) {
	localEntries, localErr := d.local.GetTrainingBudgetEntriesForYear(year)
	remoteEntries, remoteErr := d.remote.GetTrainingBudgetEntriesForYear(year)

	// If both succeed, compare
	if localErr == nil && remoteErr == nil {
		d.compareTrainingBudgetEntries(localEntries, remoteEntries, "GetTrainingBudgetEntriesForYear")
		return localEntries, nil
	}

	// If only one succeeds, log warning and return that one
	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteEntries, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localEntries, nil
	}

	// Both failed
	return nil, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

// AddTrainingBudgetEntry writes to both sources
func (d *DualLayer) AddTrainingBudgetEntry(entry TrainingBudgetEntry) error {
	localErr := d.local.AddTrainingBudgetEntry(entry)
	remoteErr := d.remote.AddTrainingBudgetEntry(entry)

	if localErr != nil {
		logging.Log("DUAL MODE: Local DB write failed: %v", localErr)
	}
	if remoteErr != nil {
		logging.Log("DUAL MODE: Remote API write failed: %v", remoteErr)
	}

	// If both fail, return error
	if localErr != nil && remoteErr != nil {
		return fmt.Errorf("both local and remote writes failed: local=%v, remote=%v", localErr, remoteErr)
	}

	// Return local error if it exists, otherwise remote error (or nil)
	if localErr != nil {
		return fmt.Errorf("local write failed: %w", localErr)
	}
	return remoteErr
}

// UpdateTrainingBudgetEntry writes to both sources
func (d *DualLayer) UpdateTrainingBudgetEntry(entry TrainingBudgetEntry) error {
	localErr := d.local.UpdateTrainingBudgetEntry(entry)
	remoteErr := d.remote.UpdateTrainingBudgetEntry(entry)

	if localErr != nil {
		logging.Log("DUAL MODE: Local DB update failed: %v", localErr)
	}
	if remoteErr != nil {
		logging.Log("DUAL MODE: Remote API update failed: %v", remoteErr)
	}

	// If both fail, return error
	if localErr != nil && remoteErr != nil {
		return fmt.Errorf("both local and remote updates failed: local=%v, remote=%v", localErr, remoteErr)
	}

	// Return local error if it exists, otherwise remote error (or nil)
	if localErr != nil {
		return fmt.Errorf("local update failed: %w", localErr)
	}
	return remoteErr
}

// DeleteTrainingBudgetEntry deletes from both sources
func (d *DualLayer) DeleteTrainingBudgetEntry(id int) error {
	localErr := d.local.DeleteTrainingBudgetEntry(id)
	remoteErr := d.remote.DeleteTrainingBudgetEntry(id)

	if localErr != nil {
		logging.Log("DUAL MODE: Local DB delete failed: %v", localErr)
	}
	if remoteErr != nil {
		logging.Log("DUAL MODE: Remote API delete failed: %v", remoteErr)
	}

	// If both fail, return error
	if localErr != nil && remoteErr != nil {
		return fmt.Errorf("both local and remote deletes failed: local=%v, remote=%v", localErr, remoteErr)
	}

	// Return local error if it exists, otherwise remote error (or nil)
	if localErr != nil {
		return fmt.Errorf("local delete failed: %w", localErr)
	}
	return remoteErr
}

// GetTrainingBudgetEntry reads from both sources and compares
func (d *DualLayer) GetTrainingBudgetEntry(id int) (TrainingBudgetEntry, error) {
	localEntry, localErr := d.local.GetTrainingBudgetEntry(id)
	remoteEntry, remoteErr := d.remote.GetTrainingBudgetEntry(id)

	// If both succeed, compare
	if localErr == nil && remoteErr == nil {
		if !reflect.DeepEqual(localEntry, remoteEntry) {
			logging.Log("DUAL MODE: GetTrainingBudgetEntry - Entry mismatch for id %d: local=%+v, remote=%+v", id, localEntry, remoteEntry)
		}
		return localEntry, nil
	}

	// If only one succeeds, log warning and return that one
	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteEntry, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localEntry, nil
	}

	// Both failed
	return TrainingBudgetEntry{}, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

// GetTrainingBudgetEntryByDate reads from both sources and compares
func (d *DualLayer) GetTrainingBudgetEntryByDate(date string) (TrainingBudgetEntry, error) {
	localEntry, localErr := d.local.GetTrainingBudgetEntryByDate(date)
	remoteEntry, remoteErr := d.remote.GetTrainingBudgetEntryByDate(date)

	// If both succeed, compare
	if localErr == nil && remoteErr == nil {
		if !reflect.DeepEqual(localEntry, remoteEntry) {
			logging.Log("DUAL MODE: GetTrainingBudgetEntryByDate - Entry mismatch for date %s: local=%+v, remote=%+v", date, localEntry, remoteEntry)
		}
		return localEntry, nil
	}

	// If only one succeeds, log warning and return that one
	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteEntry, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localEntry, nil
	}

	// Both failed
	return TrainingBudgetEntry{}, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

// Ping checks both sources
func (d *DualLayer) Ping() error {
	localErr := d.local.Ping()
	remoteErr := d.remote.Ping()

	if localErr != nil {
		logging.Log("DUAL MODE: Local DB ping failed: %v", localErr)
	}
	if remoteErr != nil {
		logging.Log("DUAL MODE: Remote API ping failed: %v", remoteErr)
	}

	// If both fail, return error
	if localErr != nil && remoteErr != nil {
		return fmt.Errorf("both local and remote pings failed: local=%v, remote=%v", localErr, remoteErr)
	}

	// Return nil if at least one succeeds
	return nil
}

// compareClients compares two slices of clients
func (d *DualLayer) compareClients(local, remote []Client, operation string) {
	if len(local) != len(remote) {
		logging.Log("DUAL MODE: %s - Client count mismatch: local=%d, remote=%d", operation, len(local), len(remote))
		return
	}

	for i := range local {
		if !reflect.DeepEqual(local[i], remote[i]) {
			logging.Log("DUAL MODE: %s - Client mismatch at index %d: local=%+v, remote=%+v", operation, i, local[i], remote[i])
		}
	}
}

// compareClientRates compares two slices of client rates
func (d *DualLayer) compareClientRates(local, remote []ClientRate, operation string) {
	if len(local) != len(remote) {
		logging.Log("DUAL MODE: %s - Client rate count mismatch: local=%d, remote=%d", operation, len(local), len(remote))
		return
	}

	for i := range local {
		if !reflect.DeepEqual(local[i], remote[i]) {
			logging.Log("DUAL MODE: %s - Client rate mismatch at index %d: local=%+v, remote=%+v", operation, i, local[i], remote[i])
		}
	}
}

// Client Operations

func (d *DualLayer) GetAllClients() ([]Client, error) {
	localClients, localErr := d.local.GetAllClients()
	remoteClients, remoteErr := d.remote.GetAllClients()

	if localErr == nil && remoteErr == nil {
		d.compareClients(localClients, remoteClients, "GetAllClients")
		return localClients, nil
	}

	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteClients, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localClients, nil
	}

	return nil, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

func (d *DualLayer) GetActiveClients() ([]Client, error) {
	localClients, localErr := d.local.GetActiveClients()
	remoteClients, remoteErr := d.remote.GetActiveClients()

	if localErr == nil && remoteErr == nil {
		d.compareClients(localClients, remoteClients, "GetActiveClients")
		return localClients, nil
	}

	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteClients, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localClients, nil
	}

	return nil, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

func (d *DualLayer) GetClientById(id int) (Client, error) {
	localClient, localErr := d.local.GetClientById(id)
	remoteClient, remoteErr := d.remote.GetClientById(id)

	if localErr == nil && remoteErr == nil {
		if !reflect.DeepEqual(localClient, remoteClient) {
			logging.Log("DUAL MODE: GetClientById - Client mismatch for id %d: local=%+v, remote=%+v", id, localClient, remoteClient)
		}
		return localClient, nil
	}

	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteClient, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localClient, nil
	}

	return Client{}, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

func (d *DualLayer) GetClientByName(name string) (Client, error) {
	localClient, localErr := d.local.GetClientByName(name)
	remoteClient, remoteErr := d.remote.GetClientByName(name)

	if localErr == nil && remoteErr == nil {
		if !reflect.DeepEqual(localClient, remoteClient) {
			logging.Log("DUAL MODE: GetClientByName - Client mismatch for name %s: local=%+v, remote=%+v", name, localClient, remoteClient)
		}
		return localClient, nil
	}

	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteClient, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localClient, nil
	}

	return Client{}, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

func (d *DualLayer) AddClient(client Client) (int, error) {
	localId, localErr := d.local.AddClient(client)
	remoteId, remoteErr := d.remote.AddClient(client)

	if localErr != nil {
		logging.Log("DUAL MODE: Local DB write failed: %v", localErr)
	}
	if remoteErr != nil {
		logging.Log("DUAL MODE: Remote API write failed: %v", remoteErr)
	}

	if localErr != nil && remoteErr != nil {
		return 0, fmt.Errorf("both local and remote writes failed: local=%v, remote=%v", localErr, remoteErr)
	}

	// Return local ID if successful, otherwise remote ID
	if localErr == nil {
		return localId, nil
	}
	return remoteId, remoteErr
}

func (d *DualLayer) UpdateClient(client Client) error {
	localErr := d.local.UpdateClient(client)
	remoteErr := d.remote.UpdateClient(client)

	if localErr != nil {
		logging.Log("DUAL MODE: Local DB update failed: %v", localErr)
	}
	if remoteErr != nil {
		logging.Log("DUAL MODE: Remote API update failed: %v", remoteErr)
	}

	if localErr != nil && remoteErr != nil {
		return fmt.Errorf("both local and remote updates failed: local=%v, remote=%v", localErr, remoteErr)
	}

	if localErr != nil {
		return fmt.Errorf("local update failed: %w", localErr)
	}
	return remoteErr
}

func (d *DualLayer) DeleteClient(id int) error {
	localErr := d.local.DeleteClient(id)
	remoteErr := d.remote.DeleteClient(id)

	if localErr != nil {
		logging.Log("DUAL MODE: Local DB delete failed: %v", localErr)
	}
	if remoteErr != nil {
		logging.Log("DUAL MODE: Remote API delete failed: %v", remoteErr)
	}

	if localErr != nil && remoteErr != nil {
		return fmt.Errorf("both local and remote deletes failed: local=%v, remote=%v", localErr, remoteErr)
	}

	if localErr != nil {
		return fmt.Errorf("local delete failed: %w", localErr)
	}
	return remoteErr
}

func (d *DualLayer) DeactivateClient(id int) error {
	localErr := d.local.DeactivateClient(id)
	remoteErr := d.remote.DeactivateClient(id)

	if localErr != nil {
		logging.Log("DUAL MODE: Local DB deactivate failed: %v", localErr)
	}
	if remoteErr != nil {
		logging.Log("DUAL MODE: Remote API deactivate failed: %v", remoteErr)
	}

	if localErr != nil && remoteErr != nil {
		return fmt.Errorf("both local and remote deactivates failed: local=%v, remote=%v", localErr, remoteErr)
	}

	if localErr != nil {
		return fmt.Errorf("local deactivate failed: %w", localErr)
	}
	return remoteErr
}

// Client Rate Operations

func (d *DualLayer) GetClientRates(clientId int) ([]ClientRate, error) {
	localRates, localErr := d.local.GetClientRates(clientId)
	remoteRates, remoteErr := d.remote.GetClientRates(clientId)

	if localErr == nil && remoteErr == nil {
		d.compareClientRates(localRates, remoteRates, "GetClientRates")
		return localRates, nil
	}

	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteRates, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localRates, nil
	}

	return nil, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

func (d *DualLayer) GetClientRateById(id int) (ClientRate, error) {
	localRate, localErr := d.local.GetClientRateById(id)
	remoteRate, remoteErr := d.remote.GetClientRateById(id)

	if localErr == nil && remoteErr == nil {
		if !reflect.DeepEqual(localRate, remoteRate) {
			logging.Log("DUAL MODE: GetClientRateById - Rate mismatch for id %d: local=%+v, remote=%+v", id, localRate, remoteRate)
		}
		return localRate, nil
	}

	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteRate, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localRate, nil
	}

	return ClientRate{}, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

func (d *DualLayer) AddClientRate(rate ClientRate) error {
	localErr := d.local.AddClientRate(rate)
	remoteErr := d.remote.AddClientRate(rate)

	if localErr != nil {
		logging.Log("DUAL MODE: Local DB write failed: %v", localErr)
	}
	if remoteErr != nil {
		logging.Log("DUAL MODE: Remote API write failed: %v", remoteErr)
	}

	if localErr != nil && remoteErr != nil {
		return fmt.Errorf("both local and remote writes failed: local=%v, remote=%v", localErr, remoteErr)
	}

	if localErr != nil {
		return fmt.Errorf("local write failed: %w", localErr)
	}
	return remoteErr
}

func (d *DualLayer) UpdateClientRate(rate ClientRate) error {
	localErr := d.local.UpdateClientRate(rate)
	remoteErr := d.remote.UpdateClientRate(rate)

	if localErr != nil {
		logging.Log("DUAL MODE: Local DB update failed: %v", localErr)
	}
	if remoteErr != nil {
		logging.Log("DUAL MODE: Remote API update failed: %v", remoteErr)
	}

	if localErr != nil && remoteErr != nil {
		return fmt.Errorf("both local and remote updates failed: local=%v, remote=%v", localErr, remoteErr)
	}

	if localErr != nil {
		return fmt.Errorf("local update failed: %w", localErr)
	}
	return remoteErr
}

func (d *DualLayer) DeleteClientRate(id int) error {
	localErr := d.local.DeleteClientRate(id)
	remoteErr := d.remote.DeleteClientRate(id)

	if localErr != nil {
		logging.Log("DUAL MODE: Local DB delete failed: %v", localErr)
	}
	if remoteErr != nil {
		logging.Log("DUAL MODE: Remote API delete failed: %v", remoteErr)
	}

	if localErr != nil && remoteErr != nil {
		return fmt.Errorf("both local and remote deletes failed: local=%v, remote=%v", localErr, remoteErr)
	}

	if localErr != nil {
		return fmt.Errorf("local delete failed: %w", localErr)
	}
	return remoteErr
}

func (d *DualLayer) GetClientRateForDate(clientId int, date string) (ClientRate, error) {
	localRate, localErr := d.local.GetClientRateForDate(clientId, date)
	remoteRate, remoteErr := d.remote.GetClientRateForDate(clientId, date)

	if localErr == nil && remoteErr == nil {
		if !reflect.DeepEqual(localRate, remoteRate) {
			logging.Log("DUAL MODE: GetClientRateForDate - Rate mismatch for client %d on %s: local=%+v, remote=%+v", clientId, date, localRate, remoteRate)
		}
		return localRate, nil
	}

	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteRate, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localRate, nil
	}

	return ClientRate{}, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

func (d *DualLayer) GetClientRateByName(clientName string, date string) (float64, error) {
	localRate, localErr := d.local.GetClientRateByName(clientName, date)
	remoteRate, remoteErr := d.remote.GetClientRateByName(clientName, date)

	if localErr == nil && remoteErr == nil {
		if localRate != remoteRate {
			logging.Log("DUAL MODE: GetClientRateByName - Rate mismatch for %s on %s: local=%.2f, remote=%.2f", clientName, date, localRate, remoteRate)
		}
		return localRate, nil
	}

	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteRate, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localRate, nil
	}

	return 0.0, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

// Earnings Operations

func (d *DualLayer) CalculateEarningsForYear(year int) (EarningsOverview, error) {
	localEarnings, localErr := d.local.CalculateEarningsForYear(year)
	remoteEarnings, remoteErr := d.remote.CalculateEarningsForYear(year)

	if localErr == nil && remoteErr == nil {
		// Compare totals
		if localEarnings.TotalHours != remoteEarnings.TotalHours || localEarnings.TotalEarnings != remoteEarnings.TotalEarnings {
			logging.Log("DUAL MODE: CalculateEarningsForYear - Earnings mismatch for year %d: local(hours=%d, earnings=%.2f), remote(hours=%d, earnings=%.2f)",
				year, localEarnings.TotalHours, localEarnings.TotalEarnings, remoteEarnings.TotalHours, remoteEarnings.TotalEarnings)
		}
		return localEarnings, nil
	}

	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteEarnings, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localEarnings, nil
	}

	return EarningsOverview{}, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

func (d *DualLayer) CalculateEarningsSummaryForYear(year int) (EarningsOverview, error) {
	localEarnings, localErr := d.local.CalculateEarningsSummaryForYear(year)
	remoteEarnings, remoteErr := d.remote.CalculateEarningsSummaryForYear(year)

	if localErr == nil && remoteErr == nil {
		// Compare totals
		if localEarnings.TotalHours != remoteEarnings.TotalHours || localEarnings.TotalEarnings != remoteEarnings.TotalEarnings {
			logging.Log("DUAL MODE: CalculateEarningsSummaryForYear - Earnings mismatch for year %d: local(hours=%d, earnings=%.2f), remote(hours=%d, earnings=%.2f)",
				year, localEarnings.TotalHours, localEarnings.TotalEarnings, remoteEarnings.TotalHours, remoteEarnings.TotalEarnings)
		}
		return localEarnings, nil
	}

	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteEarnings, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localEarnings, nil
	}

	return EarningsOverview{}, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

func (d *DualLayer) CalculateEarningsForMonth(year int, month int) (EarningsOverview, error) {
	localEarnings, localErr := d.local.CalculateEarningsForMonth(year, month)
	remoteEarnings, remoteErr := d.remote.CalculateEarningsForMonth(year, month)

	if localErr == nil && remoteErr == nil {
		// Compare totals
		if localEarnings.TotalHours != remoteEarnings.TotalHours || localEarnings.TotalEarnings != remoteEarnings.TotalEarnings {
			logging.Log("DUAL MODE: CalculateEarningsForMonth - Earnings mismatch for %d/%d: local(hours=%d, earnings=%.2f), remote(hours=%d, earnings=%.2f)",
				year, month, localEarnings.TotalHours, localEarnings.TotalEarnings, remoteEarnings.TotalHours, remoteEarnings.TotalEarnings)
		}
		return localEarnings, nil
	}

	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteEarnings, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localEarnings, nil
	}

	return EarningsOverview{}, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

func (d *DualLayer) GetClientWithRates(clientId int) (ClientWithRates, error) {
	localData, localErr := d.local.GetClientWithRates(clientId)
	remoteData, remoteErr := d.remote.GetClientWithRates(clientId)

	if localErr == nil && remoteErr == nil {
		if !reflect.DeepEqual(localData, remoteData) {
			logging.Log("DUAL MODE: GetClientWithRates - Data mismatch for client %d", clientId)
		}
		return localData, nil
	}

	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteData, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localData, nil
	}

	return ClientWithRates{}, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

// Vacation Carryover Operations

func (d *DualLayer) GetVacationCarryoverForYear(year int) (VacationCarryover, error) {
	localCarryover, localErr := d.local.GetVacationCarryoverForYear(year)
	remoteCarryover, remoteErr := d.remote.GetVacationCarryoverForYear(year)

	if localErr == nil && remoteErr == nil {
		if !reflect.DeepEqual(localCarryover, remoteCarryover) {
			logging.Log("DUAL MODE: GetVacationCarryoverForYear - Mismatch for year %d: local=%+v, remote=%+v",
				year, localCarryover, remoteCarryover)
		}
		return localCarryover, nil
	}

	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteCarryover, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localCarryover, nil
	}

	return VacationCarryover{}, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}

func (d *DualLayer) SetVacationCarryover(carryover VacationCarryover) error {
	localErr := d.local.SetVacationCarryover(carryover)
	remoteErr := d.remote.SetVacationCarryover(carryover)

	if localErr != nil {
		logging.Log("DUAL MODE: Local DB write failed: %v", localErr)
	}
	if remoteErr != nil {
		logging.Log("DUAL MODE: Remote API write failed: %v", remoteErr)
	}

	if localErr != nil && remoteErr != nil {
		return fmt.Errorf("both local and remote writes failed: local=%v, remote=%v", localErr, remoteErr)
	}

	if localErr != nil {
		return fmt.Errorf("local write failed: %w", localErr)
	}
	return remoteErr
}

func (d *DualLayer) DeleteVacationCarryover(year int) error {
	localErr := d.local.DeleteVacationCarryover(year)
	remoteErr := d.remote.DeleteVacationCarryover(year)

	if localErr != nil {
		logging.Log("DUAL MODE: Local DB delete failed: %v", localErr)
	}
	if remoteErr != nil {
		logging.Log("DUAL MODE: Remote API delete failed: %v", remoteErr)
	}

	if localErr != nil && remoteErr != nil {
		return fmt.Errorf("both local and remote deletes failed: local=%v, remote=%v", localErr, remoteErr)
	}

	if localErr != nil {
		return fmt.Errorf("local delete failed: %w", localErr)
	}
	return remoteErr
}

func (d *DualLayer) GetVacationSummaryForYear(year int) (VacationSummary, error) {
	localSummary, localErr := d.local.GetVacationSummaryForYear(year)
	remoteSummary, remoteErr := d.remote.GetVacationSummaryForYear(year)

	if localErr == nil && remoteErr == nil {
		if !reflect.DeepEqual(localSummary, remoteSummary) {
			logging.Log("DUAL MODE: GetVacationSummaryForYear - Mismatch for year %d", year)
		}
		return localSummary, nil
	}

	if localErr != nil && remoteErr == nil {
		logging.Log("DUAL MODE: Local DB failed, using remote: %v", localErr)
		return remoteSummary, nil
	}
	if localErr == nil && remoteErr != nil {
		logging.Log("DUAL MODE: Remote API failed, using local: %v", remoteErr)
		return localSummary, nil
	}

	return VacationSummary{}, fmt.Errorf("both local and remote failed: local=%v, remote=%v", localErr, remoteErr)
}
