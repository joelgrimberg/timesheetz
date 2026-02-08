// Package sync provides bidirectional synchronization between SQLite and PostgreSQL databases
package sync

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"timesheet/internal/db"
	"timesheet/internal/logging"
)

// SyncService handles synchronization between local SQLite and remote PostgreSQL
type SyncService struct {
	localDB  *sql.DB
	remoteDB *sql.DB
	mu       sync.Mutex

	// Sync state
	lastSyncTime time.Time
	syncInterval time.Duration
	stopChan     chan struct{}
	running      bool

	// Stats
	lastSyncStats SyncStats
}

// SyncStats contains statistics about the last sync operation
type SyncStats struct {
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
	TablesProcessed int
	RecordsPushed   int
	RecordsPulled   int
	Errors          []string
}

// SyncDirection indicates the direction of sync
type SyncDirection int

const (
	SyncBidirectional SyncDirection = iota
	SyncPushOnly                    // Local -> Remote
	SyncPullOnly                    // Remote -> Local
)

// NewSyncService creates a new sync service
func NewSyncService(localDB, remoteDB *sql.DB, interval time.Duration) *SyncService {
	return &SyncService{
		localDB:      localDB,
		remoteDB:     remoteDB,
		syncInterval: interval,
		stopChan:     make(chan struct{}),
	}
}

// Start begins background synchronization
func (s *SyncService) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	logging.Log("Starting background sync service (interval: %v)", s.syncInterval)

	go func() {
		// Initial sync
		s.Sync(SyncBidirectional)

		ticker := time.NewTicker(s.syncInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.Sync(SyncBidirectional)
			case <-s.stopChan:
				logging.Log("Sync service stopped")
				return
			}
		}
	}()
}

// Stop halts background synchronization
func (s *SyncService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		close(s.stopChan)
		s.running = false
	}
}

// IsRunning returns whether the sync service is running
func (s *SyncService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// GetLastSyncTime returns the time of the last successful sync
func (s *SyncService) GetLastSyncTime() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastSyncTime
}

// GetLastSyncStats returns statistics from the last sync
func (s *SyncService) GetLastSyncStats() SyncStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastSyncStats
}

// Sync performs synchronization between databases
func (s *SyncService) Sync(direction SyncDirection) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats := SyncStats{
		StartTime: time.Now(),
	}

	logging.Log("Starting sync...")

	// Sync each table
	tables := []struct {
		name     string
		syncFunc func(SyncDirection, *SyncStats) error
	}{
		{"clients", s.syncClients},
		{"client_rates", s.syncClientRates},
		{"timesheet", s.syncTimesheet},
		{"training_budget", s.syncTrainingBudget},
		{"vacation_carryover", s.syncVacationCarryover},
	}

	for _, table := range tables {
		if err := table.syncFunc(direction, &stats); err != nil {
			errMsg := fmt.Sprintf("Error syncing %s: %v", table.name, err)
			stats.Errors = append(stats.Errors, errMsg)
			logging.Log("%s", errMsg)
		} else {
			stats.TablesProcessed++
		}
	}

	stats.EndTime = time.Now()
	stats.Duration = stats.EndTime.Sub(stats.StartTime)

	s.lastSyncTime = time.Now()
	s.lastSyncStats = stats

	logging.Log("Sync completed in %v (pushed: %d, pulled: %d, errors: %d)",
		stats.Duration, stats.RecordsPushed, stats.RecordsPulled, len(stats.Errors))

	if len(stats.Errors) > 0 {
		return fmt.Errorf("sync completed with %d errors", len(stats.Errors))
	}
	return nil
}

// syncClients synchronizes the clients table
func (s *SyncService) syncClients(direction SyncDirection, stats *SyncStats) error {
	// Get all clients from both databases
	localClients, err := s.getClientsFromDB(s.localDB, "sqlite")
	if err != nil {
		return fmt.Errorf("failed to get local clients: %w", err)
	}

	remoteClients, err := s.getClientsFromDB(s.remoteDB, "postgres")
	if err != nil {
		return fmt.Errorf("failed to get remote clients: %w", err)
	}

	// Build maps for comparison (by name since that's the unique key)
	localMap := make(map[string]clientRecord)
	for _, c := range localClients {
		localMap[c.Name] = c
	}

	remoteMap := make(map[string]clientRecord)
	for _, c := range remoteClients {
		remoteMap[c.Name] = c
	}

	// Push local -> remote
	if direction == SyncBidirectional || direction == SyncPushOnly {
		for name, local := range localMap {
			remote, exists := remoteMap[name]
			if !exists {
				// Insert new record to remote
				if err := s.insertClientToRemote(local); err != nil {
					return fmt.Errorf("failed to insert client %s to remote: %w", name, err)
				}
				stats.RecordsPushed++
			} else if local.UpdatedAt > remote.UpdatedAt {
				// Update remote with local data
				if err := s.updateClientInRemote(local, remote.Id); err != nil {
					return fmt.Errorf("failed to update client %s in remote: %w", name, err)
				}
				stats.RecordsPushed++
			}
		}
	}

	// Pull remote -> local
	if direction == SyncBidirectional || direction == SyncPullOnly {
		for name, remote := range remoteMap {
			local, exists := localMap[name]
			if !exists {
				// Insert new record to local
				if err := s.insertClientToLocal(remote); err != nil {
					return fmt.Errorf("failed to insert client %s to local: %w", name, err)
				}
				stats.RecordsPulled++
			} else if remote.UpdatedAt > local.UpdatedAt {
				// Update local with remote data
				if err := s.updateClientInLocal(remote, local.Id); err != nil {
					return fmt.Errorf("failed to update client %s in local: %w", name, err)
				}
				stats.RecordsPulled++
			}
		}
	}

	return nil
}

// syncClientRates synchronizes the client_rates table
func (s *SyncService) syncClientRates(direction SyncDirection, stats *SyncStats) error {
	// First, we need a mapping of client names to IDs in both databases
	localClientMap, err := s.getClientIdMap(s.localDB, "sqlite")
	if err != nil {
		return fmt.Errorf("failed to get local client map: %w", err)
	}

	remoteClientMap, err := s.getClientIdMap(s.remoteDB, "postgres")
	if err != nil {
		return fmt.Errorf("failed to get remote client map: %w", err)
	}

	// Get all rates from both databases
	localRates, err := s.getClientRatesFromDB(s.localDB, "sqlite")
	if err != nil {
		return fmt.Errorf("failed to get local rates: %w", err)
	}

	remoteRates, err := s.getClientRatesFromDB(s.remoteDB, "postgres")
	if err != nil {
		return fmt.Errorf("failed to get remote rates: %w", err)
	}

	// Build reverse map (ID -> name) for lookups
	localIdToName := make(map[int]string)
	for name, id := range localClientMap {
		localIdToName[id] = name
	}

	remoteIdToName := make(map[int]string)
	for name, id := range remoteClientMap {
		remoteIdToName[id] = name
	}

	// Create composite key for rates: clientName + effectiveDate
	localRateMap := make(map[string]clientRateRecord)
	for _, r := range localRates {
		clientName := localIdToName[r.ClientId]
		key := fmt.Sprintf("%s|%s", clientName, r.EffectiveDate)
		localRateMap[key] = r
	}

	remoteRateMap := make(map[string]clientRateRecord)
	for _, r := range remoteRates {
		clientName := remoteIdToName[r.ClientId]
		key := fmt.Sprintf("%s|%s", clientName, r.EffectiveDate)
		remoteRateMap[key] = r
	}

	// Push local -> remote
	if direction == SyncBidirectional || direction == SyncPushOnly {
		for key, local := range localRateMap {
			clientName := localIdToName[local.ClientId]
			remoteClientId, ok := remoteClientMap[clientName]
			if !ok {
				continue // Client doesn't exist in remote yet
			}

			remote, exists := remoteRateMap[key]
			if !exists {
				if err := s.insertClientRateToRemote(local, remoteClientId); err != nil {
					return fmt.Errorf("failed to insert rate to remote: %w", err)
				}
				stats.RecordsPushed++
			} else if local.UpdatedAt > remote.UpdatedAt {
				if err := s.updateClientRateInRemote(local, remote.Id, remoteClientId); err != nil {
					return fmt.Errorf("failed to update rate in remote: %w", err)
				}
				stats.RecordsPushed++
			}
		}
	}

	// Pull remote -> local
	if direction == SyncBidirectional || direction == SyncPullOnly {
		for key, remote := range remoteRateMap {
			clientName := remoteIdToName[remote.ClientId]
			localClientId, ok := localClientMap[clientName]
			if !ok {
				continue // Client doesn't exist in local yet
			}

			local, exists := localRateMap[key]
			if !exists {
				if err := s.insertClientRateToLocal(remote, localClientId); err != nil {
					return fmt.Errorf("failed to insert rate to local: %w", err)
				}
				stats.RecordsPulled++
			} else if remote.UpdatedAt > local.UpdatedAt {
				if err := s.updateClientRateInLocal(remote, local.Id, localClientId); err != nil {
					return fmt.Errorf("failed to update rate in local: %w", err)
				}
				stats.RecordsPulled++
			}
		}
	}

	return nil
}

// syncTimesheet synchronizes the timesheet table
func (s *SyncService) syncTimesheet(direction SyncDirection, stats *SyncStats) error {
	localEntries, err := s.getTimesheetFromDB(s.localDB, "sqlite")
	if err != nil {
		return fmt.Errorf("failed to get local timesheet: %w", err)
	}

	remoteEntries, err := s.getTimesheetFromDB(s.remoteDB, "postgres")
	if err != nil {
		return fmt.Errorf("failed to get remote timesheet: %w", err)
	}

	// Use date as the unique key (one entry per date)
	localMap := make(map[string]timesheetRecord)
	for _, e := range localEntries {
		localMap[e.Date] = e
	}

	remoteMap := make(map[string]timesheetRecord)
	for _, e := range remoteEntries {
		remoteMap[e.Date] = e
	}

	// Push local -> remote
	if direction == SyncBidirectional || direction == SyncPushOnly {
		for date, local := range localMap {
			remote, exists := remoteMap[date]
			if !exists {
				if err := s.insertTimesheetToRemote(local); err != nil {
					return fmt.Errorf("failed to insert timesheet %s to remote: %w", date, err)
				}
				stats.RecordsPushed++
			} else if local.UpdatedAt > remote.UpdatedAt {
				if err := s.updateTimesheetInRemote(local, remote.Id); err != nil {
					return fmt.Errorf("failed to update timesheet %s in remote: %w", date, err)
				}
				stats.RecordsPushed++
			}
		}
	}

	// Pull remote -> local
	if direction == SyncBidirectional || direction == SyncPullOnly {
		for date, remote := range remoteMap {
			local, exists := localMap[date]
			if !exists {
				if err := s.insertTimesheetToLocal(remote); err != nil {
					return fmt.Errorf("failed to insert timesheet %s to local: %w", date, err)
				}
				stats.RecordsPulled++
			} else if remote.UpdatedAt > local.UpdatedAt {
				if err := s.updateTimesheetInLocal(remote, local.Id); err != nil {
					return fmt.Errorf("failed to update timesheet %s in local: %w", date, err)
				}
				stats.RecordsPulled++
			}
		}
	}

	return nil
}

// syncTrainingBudget synchronizes the training_budget table
func (s *SyncService) syncTrainingBudget(direction SyncDirection, stats *SyncStats) error {
	localEntries, err := s.getTrainingBudgetFromDB(s.localDB, "sqlite")
	if err != nil {
		return fmt.Errorf("failed to get local training budget: %w", err)
	}

	remoteEntries, err := s.getTrainingBudgetFromDB(s.remoteDB, "postgres")
	if err != nil {
		return fmt.Errorf("failed to get remote training budget: %w", err)
	}

	// Use date + training_name as composite key
	localMap := make(map[string]trainingBudgetRecord)
	for _, e := range localEntries {
		key := fmt.Sprintf("%s|%s", e.Date, e.TrainingName)
		localMap[key] = e
	}

	remoteMap := make(map[string]trainingBudgetRecord)
	for _, e := range remoteEntries {
		key := fmt.Sprintf("%s|%s", e.Date, e.TrainingName)
		remoteMap[key] = e
	}

	// Push local -> remote
	if direction == SyncBidirectional || direction == SyncPushOnly {
		for key, local := range localMap {
			remote, exists := remoteMap[key]
			if !exists {
				if err := s.insertTrainingBudgetToRemote(local); err != nil {
					return fmt.Errorf("failed to insert training budget to remote: %w", err)
				}
				stats.RecordsPushed++
			} else if local.UpdatedAt > remote.UpdatedAt {
				if err := s.updateTrainingBudgetInRemote(local, remote.Id); err != nil {
					return fmt.Errorf("failed to update training budget in remote: %w", err)
				}
				stats.RecordsPushed++
			}
		}
	}

	// Pull remote -> local
	if direction == SyncBidirectional || direction == SyncPullOnly {
		for key, remote := range remoteMap {
			local, exists := localMap[key]
			if !exists {
				if err := s.insertTrainingBudgetToLocal(remote); err != nil {
					return fmt.Errorf("failed to insert training budget to local: %w", err)
				}
				stats.RecordsPulled++
			} else if remote.UpdatedAt > local.UpdatedAt {
				if err := s.updateTrainingBudgetInLocal(remote, local.Id); err != nil {
					return fmt.Errorf("failed to update training budget in local: %w", err)
				}
				stats.RecordsPulled++
			}
		}
	}

	return nil
}

// syncVacationCarryover synchronizes the vacation_carryover table
func (s *SyncService) syncVacationCarryover(direction SyncDirection, stats *SyncStats) error {
	localEntries, err := s.getVacationCarryoverFromDB(s.localDB, "sqlite")
	if err != nil {
		return fmt.Errorf("failed to get local vacation carryover: %w", err)
	}

	remoteEntries, err := s.getVacationCarryoverFromDB(s.remoteDB, "postgres")
	if err != nil {
		return fmt.Errorf("failed to get remote vacation carryover: %w", err)
	}

	// Use year as unique key
	localMap := make(map[int]db.VacationCarryover)
	for _, e := range localEntries {
		localMap[e.Year] = e
	}

	remoteMap := make(map[int]db.VacationCarryover)
	for _, e := range remoteEntries {
		remoteMap[e.Year] = e
	}

	// Push local -> remote
	if direction == SyncBidirectional || direction == SyncPushOnly {
		for year, local := range localMap {
			remote, exists := remoteMap[year]
			if !exists {
				if err := s.insertVacationCarryoverToRemote(local); err != nil {
					return fmt.Errorf("failed to insert vacation carryover %d to remote: %w", year, err)
				}
				stats.RecordsPushed++
			} else if local.UpdatedAt > remote.UpdatedAt {
				if err := s.updateVacationCarryoverInRemote(local, remote.Id); err != nil {
					return fmt.Errorf("failed to update vacation carryover %d in remote: %w", year, err)
				}
				stats.RecordsPushed++
			}
		}
	}

	// Pull remote -> local
	if direction == SyncBidirectional || direction == SyncPullOnly {
		for year, remote := range remoteMap {
			local, exists := localMap[year]
			if !exists {
				if err := s.insertVacationCarryoverToLocal(remote); err != nil {
					return fmt.Errorf("failed to insert vacation carryover %d to local: %w", year, err)
				}
				stats.RecordsPulled++
			} else if remote.UpdatedAt > local.UpdatedAt {
				if err := s.updateVacationCarryoverInLocal(remote, local.Id); err != nil {
					return fmt.Errorf("failed to update vacation carryover %d in local: %w", year, err)
				}
				stats.RecordsPulled++
			}
		}
	}

	return nil
}
