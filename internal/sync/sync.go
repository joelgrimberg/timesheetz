// Package sync provides bidirectional synchronization between SQLite and PostgreSQL databases
package sync

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
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
		{"buffer_hours", s.syncBufferHours},
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

	// Tombstone pass: reconcile deletes before the upsert pass so we don't
	// re-insert a row that was just deleted on the other side.
	localTs, err := s.getTombstonesFromDB(s.localDB, "sqlite", db.TombstoneTableClients)
	if err != nil {
		return fmt.Errorf("failed to get local tombstones: %w", err)
	}
	remoteTs, err := s.getTombstonesFromDB(s.remoteDB, "postgres", db.TombstoneTableClients)
	if err != nil {
		return fmt.Errorf("failed to get remote tombstones: %w", err)
	}
	rec, err := s.reconcileTombstones(
		db.TombstoneTableClients,
		localTs, remoteTs,
		func(key string) (string, bool) {
			c, ok := localMap[key]
			return c.UpdatedAt, ok
		},
		func(key string) (string, bool) {
			c, ok := remoteMap[key]
			return c.UpdatedAt, ok
		},
		func(key string) error {
			_, err := s.localDB.Exec(`DELETE FROM clients WHERE name = ?`, key)
			delete(localMap, key)
			return err
		},
		func(key string) error {
			_, err := s.remoteDB.Exec(`DELETE FROM clients WHERE name = $1`, key)
			delete(remoteMap, key)
			return err
		},
	)
	if err != nil {
		return err
	}

	// Push local -> remote
	if direction == SyncBidirectional || direction == SyncPushOnly {
		for name, local := range localMap {
			if rec.isKilled(name) {
				continue
			}
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
			if rec.isKilled(name) {
				continue
			}
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

	// Tombstone pass.
	localTs, err := s.getTombstonesFromDB(s.localDB, "sqlite", db.TombstoneTableClientRates)
	if err != nil {
		return fmt.Errorf("failed to get local rate tombstones: %w", err)
	}
	remoteTs, err := s.getTombstonesFromDB(s.remoteDB, "postgres", db.TombstoneTableClientRates)
	if err != nil {
		return fmt.Errorf("failed to get remote rate tombstones: %w", err)
	}
	rec, err := s.reconcileTombstones(
		db.TombstoneTableClientRates,
		localTs, remoteTs,
		func(key string) (string, bool) {
			r, ok := localRateMap[key]
			return r.UpdatedAt, ok
		},
		func(key string) (string, bool) {
			r, ok := remoteRateMap[key]
			return r.UpdatedAt, ok
		},
		func(key string) error {
			// key = "clientName|effectiveDate"; resolve clientId via the
			// local client map and delete by (client_id, effective_date).
			name, date, ok := splitRateKey(key)
			if !ok {
				return nil
			}
			cid, ok := localClientMap[name]
			if !ok {
				delete(localRateMap, key)
				return nil
			}
			_, err := s.localDB.Exec(`DELETE FROM client_rates WHERE client_id = ? AND effective_date = ?`, cid, date)
			delete(localRateMap, key)
			return err
		},
		func(key string) error {
			name, date, ok := splitRateKey(key)
			if !ok {
				return nil
			}
			cid, ok := remoteClientMap[name]
			if !ok {
				delete(remoteRateMap, key)
				return nil
			}
			_, err := s.remoteDB.Exec(`DELETE FROM client_rates WHERE client_id = $1 AND effective_date = $2`, cid, date)
			delete(remoteRateMap, key)
			return err
		},
	)
	if err != nil {
		return err
	}

	// Push local -> remote
	if direction == SyncBidirectional || direction == SyncPushOnly {
		for key, local := range localRateMap {
			if rec.isKilled(key) {
				continue
			}
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
			if rec.isKilled(key) {
				continue
			}
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

// splitRateKey splits a "clientName|effectiveDate" key back into its parts.
// Returns ok=false when the key is malformed (shouldn't happen given the
// data layer is the only thing writing these).
func splitRateKey(key string) (name, date string, ok bool) {
	i := strings.Index(key, "|")
	if i < 0 {
		return "", "", false
	}
	return key[:i], key[i+1:], true
}

// splitTrainingKey splits a "date|trainingName" key back into its parts.
func splitTrainingKey(key string) (date, name string, ok bool) {
	i := strings.Index(key, "|")
	if i < 0 {
		return "", "", false
	}
	return key[:i], key[i+1:], true
}

// parseBufferKey parses a "YYYY-MM" key back into year and month.
func parseBufferKey(key string) (year, month int, ok bool) {
	i := strings.Index(key, "-")
	if i < 0 {
		return 0, 0, false
	}
	y, err := strconv.Atoi(key[:i])
	if err != nil {
		return 0, 0, false
	}
	m, err := strconv.Atoi(key[i+1:])
	if err != nil {
		return 0, 0, false
	}
	return y, m, true
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

	// Tombstone pass.
	localTs, err := s.getTombstonesFromDB(s.localDB, "sqlite", db.TombstoneTableTimesheet)
	if err != nil {
		return fmt.Errorf("failed to get local timesheet tombstones: %w", err)
	}
	remoteTs, err := s.getTombstonesFromDB(s.remoteDB, "postgres", db.TombstoneTableTimesheet)
	if err != nil {
		return fmt.Errorf("failed to get remote timesheet tombstones: %w", err)
	}
	rec, err := s.reconcileTombstones(
		db.TombstoneTableTimesheet,
		localTs, remoteTs,
		func(key string) (string, bool) {
			e, ok := localMap[key]
			return e.UpdatedAt, ok
		},
		func(key string) (string, bool) {
			e, ok := remoteMap[key]
			return e.UpdatedAt, ok
		},
		func(key string) error {
			_, err := s.localDB.Exec(`DELETE FROM timesheet WHERE date = ?`, key)
			delete(localMap, key)
			return err
		},
		func(key string) error {
			_, err := s.remoteDB.Exec(`DELETE FROM timesheet WHERE date = $1`, key)
			delete(remoteMap, key)
			return err
		},
	)
	if err != nil {
		return err
	}

	// Push local -> remote
	if direction == SyncBidirectional || direction == SyncPushOnly {
		for date, local := range localMap {
			if rec.isKilled(date) {
				continue
			}
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
			if rec.isKilled(date) {
				continue
			}
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

	// Tombstone pass.
	localTs, err := s.getTombstonesFromDB(s.localDB, "sqlite", db.TombstoneTableTrainingBudget)
	if err != nil {
		return fmt.Errorf("failed to get local training tombstones: %w", err)
	}
	remoteTs, err := s.getTombstonesFromDB(s.remoteDB, "postgres", db.TombstoneTableTrainingBudget)
	if err != nil {
		return fmt.Errorf("failed to get remote training tombstones: %w", err)
	}
	rec, err := s.reconcileTombstones(
		db.TombstoneTableTrainingBudget,
		localTs, remoteTs,
		func(key string) (string, bool) {
			e, ok := localMap[key]
			return e.UpdatedAt, ok
		},
		func(key string) (string, bool) {
			e, ok := remoteMap[key]
			return e.UpdatedAt, ok
		},
		func(key string) error {
			date, name, ok := splitTrainingKey(key)
			if !ok {
				return nil
			}
			_, err := s.localDB.Exec(`DELETE FROM training_budget WHERE date = ? AND training_name = ?`, date, name)
			delete(localMap, key)
			return err
		},
		func(key string) error {
			date, name, ok := splitTrainingKey(key)
			if !ok {
				return nil
			}
			_, err := s.remoteDB.Exec(`DELETE FROM training_budget WHERE date = $1 AND training_name = $2`, date, name)
			delete(remoteMap, key)
			return err
		},
	)
	if err != nil {
		return err
	}

	// Push local -> remote
	if direction == SyncBidirectional || direction == SyncPushOnly {
		for key, local := range localMap {
			if rec.isKilled(key) {
				continue
			}
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
			if rec.isKilled(key) {
				continue
			}
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

// syncBufferHours synchronizes the buffer_hours table. The unique key is
// (year, month), so we map on that composite to detect inserts vs. updates.
func (s *SyncService) syncBufferHours(direction SyncDirection, stats *SyncStats) error {
	type key struct{ year, month int }

	localEntries, err := s.getBufferHoursFromDB(s.localDB, "sqlite")
	if err != nil {
		return fmt.Errorf("failed to get local buffer hours: %w", err)
	}
	remoteEntries, err := s.getBufferHoursFromDB(s.remoteDB, "postgres")
	if err != nil {
		return fmt.Errorf("failed to get remote buffer hours: %w", err)
	}

	localMap := make(map[key]db.BufferEntry, len(localEntries))
	for _, e := range localEntries {
		localMap[key{e.Year, e.Month}] = e
	}
	remoteMap := make(map[key]db.BufferEntry, len(remoteEntries))
	for _, e := range remoteEntries {
		remoteMap[key{e.Year, e.Month}] = e
	}

	// Tombstone pass. Keys are encoded as "YYYY-MM" strings; we parse them
	// back to the (year, month) struct keys our maps use.
	localTs, err := s.getTombstonesFromDB(s.localDB, "sqlite", db.TombstoneTableBufferHours)
	if err != nil {
		return fmt.Errorf("failed to get local buffer tombstones: %w", err)
	}
	remoteTs, err := s.getTombstonesFromDB(s.remoteDB, "postgres", db.TombstoneTableBufferHours)
	if err != nil {
		return fmt.Errorf("failed to get remote buffer tombstones: %w", err)
	}
	rec, err := s.reconcileTombstones(
		db.TombstoneTableBufferHours,
		localTs, remoteTs,
		func(tk string) (string, bool) {
			y, m, ok := parseBufferKey(tk)
			if !ok {
				return "", false
			}
			e, found := localMap[key{y, m}]
			return e.UpdatedAt, found
		},
		func(tk string) (string, bool) {
			y, m, ok := parseBufferKey(tk)
			if !ok {
				return "", false
			}
			e, found := remoteMap[key{y, m}]
			return e.UpdatedAt, found
		},
		func(tk string) error {
			y, m, ok := parseBufferKey(tk)
			if !ok {
				return nil
			}
			_, err := s.localDB.Exec(`DELETE FROM buffer_hours WHERE year = ? AND month = ?`, y, m)
			delete(localMap, key{y, m})
			return err
		},
		func(tk string) error {
			y, m, ok := parseBufferKey(tk)
			if !ok {
				return nil
			}
			_, err := s.remoteDB.Exec(`DELETE FROM buffer_hours WHERE year = $1 AND month = $2`, y, m)
			delete(remoteMap, key{y, m})
			return err
		},
	)
	if err != nil {
		return err
	}

	if direction == SyncBidirectional || direction == SyncPushOnly {
		for k, local := range localMap {
			if rec.isKilled(db.TombstoneKeyBufferHours(k.year, k.month)) {
				continue
			}
			remote, exists := remoteMap[k]
			if !exists {
				if err := s.insertBufferHoursToRemote(local); err != nil {
					return fmt.Errorf("failed to insert buffer %d-%02d to remote: %w", k.year, k.month, err)
				}
				stats.RecordsPushed++
			} else if local.UpdatedAt > remote.UpdatedAt {
				if err := s.updateBufferHoursInRemote(local, remote.Id); err != nil {
					return fmt.Errorf("failed to update buffer %d-%02d in remote: %w", k.year, k.month, err)
				}
				stats.RecordsPushed++
			}
		}
	}

	if direction == SyncBidirectional || direction == SyncPullOnly {
		for k, remote := range remoteMap {
			if rec.isKilled(db.TombstoneKeyBufferHours(k.year, k.month)) {
				continue
			}
			local, exists := localMap[k]
			if !exists {
				if err := s.insertBufferHoursToLocal(remote); err != nil {
					return fmt.Errorf("failed to insert buffer %d-%02d to local: %w", k.year, k.month, err)
				}
				stats.RecordsPulled++
			} else if remote.UpdatedAt > local.UpdatedAt {
				if err := s.updateBufferHoursInLocal(remote, local.Id); err != nil {
					return fmt.Errorf("failed to update buffer %d-%02d in local: %w", k.year, k.month, err)
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

	// Tombstone pass. Keys are the year encoded as a decimal string.
	localTs, err := s.getTombstonesFromDB(s.localDB, "sqlite", db.TombstoneTableVacationCarryover)
	if err != nil {
		return fmt.Errorf("failed to get local vacation tombstones: %w", err)
	}
	remoteTs, err := s.getTombstonesFromDB(s.remoteDB, "postgres", db.TombstoneTableVacationCarryover)
	if err != nil {
		return fmt.Errorf("failed to get remote vacation tombstones: %w", err)
	}
	rec, err := s.reconcileTombstones(
		db.TombstoneTableVacationCarryover,
		localTs, remoteTs,
		func(tk string) (string, bool) {
			y, err := strconv.Atoi(tk)
			if err != nil {
				return "", false
			}
			e, ok := localMap[y]
			return e.UpdatedAt, ok
		},
		func(tk string) (string, bool) {
			y, err := strconv.Atoi(tk)
			if err != nil {
				return "", false
			}
			e, ok := remoteMap[y]
			return e.UpdatedAt, ok
		},
		func(tk string) error {
			y, err := strconv.Atoi(tk)
			if err != nil {
				return nil
			}
			_, err = s.localDB.Exec(`DELETE FROM vacation_carryover WHERE year = ?`, y)
			delete(localMap, y)
			return err
		},
		func(tk string) error {
			y, err := strconv.Atoi(tk)
			if err != nil {
				return nil
			}
			_, err = s.remoteDB.Exec(`DELETE FROM vacation_carryover WHERE year = $1`, y)
			delete(remoteMap, y)
			return err
		},
	)
	if err != nil {
		return err
	}

	// Push local -> remote
	if direction == SyncBidirectional || direction == SyncPushOnly {
		for year, local := range localMap {
			if rec.isKilled(db.TombstoneKeyVacationCarryover(year)) {
				continue
			}
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
			if rec.isKilled(db.TombstoneKeyVacationCarryover(year)) {
				continue
			}
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
