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

