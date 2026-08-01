package ui

import (
	"testing"
	"timesheet/internal/db"
)

// TestSyncComplete_OnlyRefreshesActiveTab verifies that a SyncCompleteMsg
// only rebuilds the model for the currently active tab. Rebuilding all
// tabs every sync interval (~15s) was ~7× more work than needed when the
// user is looking at one tab. Inactive tabs still catch synced data via
// the existing < / > tab-switch refresh path.
//
// The test asserts this by mutating the DB between initial model builds
// and the sync-complete message: the active tab must reflect the change
// (rebuild happened) while inactive tabs must not (rebuild skipped).
func TestSyncComplete_OnlyRefreshesActiveTab(t *testing.T) {
	if err := db.InitializeDatabase(":memory:"); err != nil {
		t.Fatalf("init db: %v", err)
	}
	defer db.Close()

	// Seed one client so all Initial*Model calls succeed.
	if _, err := db.AddClient(db.Client{Name: "Acme", IsActive: true}); err != nil {
		t.Fatal(err)
	}

	m := AppModel{
		ActiveMode:     TimesheetMode,
		TimesheetModel: InitialTimesheetModel(),
		OverviewModel:  InitialOverviewModel(),
		ClientsModel:   InitialClientsModel(),
	}

	inactiveRowsBefore := len(m.ClientsModel.table.Rows())

	// Mutate the DB — add a new client. After SyncComplete, the active
	// tab (Timesheet) would be rebuilt; ClientsModel (inactive) must NOT
	// be rebuilt and therefore must not see the new row yet.
	if _, err := db.AddClient(db.Client{Name: "Beta", IsActive: true}); err != nil {
		t.Fatal(err)
	}

	updated, _ := m.Update(SyncCompleteMsg{})
	got := updated.(AppModel)

	if got.ActiveMode != TimesheetMode {
		t.Fatalf("ActiveMode changed unexpectedly: %v", got.ActiveMode)
	}

	if got.syncStatus == "" || got.syncStatus == "Sync error" {
		t.Errorf("expected a healthy sync status, got %q", got.syncStatus)
	}

	if inactiveRowsAfter := len(got.ClientsModel.table.Rows()); inactiveRowsAfter != inactiveRowsBefore {
		t.Errorf("ClientsModel rebuilt despite being inactive: %d rows before, %d after",
			inactiveRowsBefore, inactiveRowsAfter)
	}
}

// TestSyncComplete_RefreshesActiveClientsTab is the mirror test: when the
// user IS on the ClientsMode tab, a sync must refresh it so newly-synced
// clients appear immediately (no user action required).
func TestSyncComplete_RefreshesActiveClientsTab(t *testing.T) {
	if err := db.InitializeDatabase(":memory:"); err != nil {
		t.Fatalf("init db: %v", err)
	}
	defer db.Close()

	if _, err := db.AddClient(db.Client{Name: "Acme", IsActive: true}); err != nil {
		t.Fatal(err)
	}

	m := AppModel{
		ActiveMode:   ClientsMode,
		ClientsModel: InitialClientsModel(),
	}
	rowsBefore := len(m.ClientsModel.table.Rows())

	if _, err := db.AddClient(db.Client{Name: "Beta", IsActive: true}); err != nil {
		t.Fatal(err)
	}

	updated, _ := m.Update(SyncCompleteMsg{})
	got := updated.(AppModel)

	if rowsAfter := len(got.ClientsModel.table.Rows()); rowsAfter != rowsBefore+1 {
		t.Errorf("active ClientsModel was not refreshed on sync: %d rows before, %d after (expected %d)",
			rowsBefore, rowsAfter, rowsBefore+1)
	}
}
