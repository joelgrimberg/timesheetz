package ui

import (
	"testing"
	"timesheet/internal/db"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func TestTrainingBudgetKeyMap_EditBinding(t *testing.T) {
	keyMap := DefaultTrainingBudgetKeyMap()

	// Test that edit binding exists and has correct keys
	if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}, keyMap.Edit) {
		t.Error("Expected 'e' key to match Edit binding")
	}

	if !key.Matches(tea.KeyMsg{Type: tea.KeyEnter}, keyMap.Edit) {
		t.Error("Expected 'enter' key to match Edit binding")
	}
}

func TestInitialTrainingBudgetFormModelForEdit(t *testing.T) {
	entry := db.TrainingBudgetEntry{
		Id:               123,
		Date:             "2024-03-15",
		Training_name:    "Go Training Course",
		Hours:            8,
		Cost_without_vat: 500.50,
	}

	form := InitialTrainingBudgetFormModelForEdit(entry)

	// Verify isEditing flag is set
	if !form.isEditing {
		t.Error("Expected isEditing to be true")
	}

	// Verify entryID is set
	if form.entryID != 123 {
		t.Errorf("Expected entryID to be 123, got %d", form.entryID)
	}

	// Verify form fields are pre-filled
	if form.inputs[0].Value() != "2024-03-15" {
		t.Errorf("Expected date to be '2024-03-15', got '%s'", form.inputs[0].Value())
	}

	if form.inputs[1].Value() != "Go Training Course" {
		t.Errorf("Expected training name to be 'Go Training Course', got '%s'", form.inputs[1].Value())
	}

	if form.inputs[2].Value() != "500.50" {
		t.Errorf("Expected cost to be '500.50', got '%s'", form.inputs[2].Value())
	}
}

func TestInitialTrainingBudgetFormModel_NotEditing(t *testing.T) {
	form := InitialTrainingBudgetFormModel()

	// Verify isEditing flag is false by default
	if form.isEditing {
		t.Error("Expected isEditing to be false for new form")
	}

	// Verify entryID is 0 by default
	if form.entryID != 0 {
		t.Errorf("Expected entryID to be 0, got %d", form.entryID)
	}
}

func TestTrainingBudgetFormModel_ViewTitle(t *testing.T) {
	// Test add mode title
	addForm := InitialTrainingBudgetFormModel()
	addView := addForm.View()
	if !containsString(addView, "Add Training Budget Entry") {
		t.Error("Expected 'Add Training Budget Entry' in add form view")
	}

	// Test edit mode title
	entry := db.TrainingBudgetEntry{
		Id:               1,
		Date:             "2024-03-15",
		Training_name:    "Test",
		Cost_without_vat: 100.0,
	}
	editForm := InitialTrainingBudgetFormModelForEdit(entry)
	editView := editForm.View()
	if !containsString(editView, "Edit Training Budget Entry") {
		t.Error("Expected 'Edit Training Budget Entry' in edit form view")
	}
}

func TestTrainingBudgetModel_EditEntryCmd(t *testing.T) {
	// Create a minimal model without database access
	model := TrainingBudgetModel{}

	entry := db.TrainingBudgetEntry{
		Id:               42,
		Date:             "2024-03-20",
		Training_name:    "Test Training",
		Cost_without_vat: 250.0,
	}

	cmd := model.editEntryCmd(entry)
	msg := cmd()

	editMsg, ok := msg.(EditTrainingBudgetMsg)
	if !ok {
		t.Fatalf("Expected EditTrainingBudgetMsg, got %T", msg)
	}

	if editMsg.Entry.Id != 42 {
		t.Errorf("Expected entry ID 42, got %d", editMsg.Entry.Id)
	}

	if editMsg.Entry.Training_name != "Test Training" {
		t.Errorf("Expected 'Test Training', got '%s'", editMsg.Entry.Training_name)
	}
}

func TestTrainingBudgetModel_AddEntryCmd(t *testing.T) {
	// Create a minimal model without database access
	model := TrainingBudgetModel{}

	cmd := model.addEntryCmd()
	msg := cmd()

	_, ok := msg.(AddTrainingBudgetMsg)
	if !ok {
		t.Fatalf("Expected AddTrainingBudgetMsg, got %T", msg)
	}
}

func TestTrainingBudgetKeyMap_FullHelp_ContainsEdit(t *testing.T) {
	keyMap := DefaultTrainingBudgetKeyMap()
	fullHelp := keyMap.FullHelp()

	// Find the Edit binding in full help
	found := false
	for _, group := range fullHelp {
		for _, binding := range group {
			if binding.Help().Key == keyMap.Edit.Help().Key {
				found = true
				break
			}
		}
	}

	if !found {
		t.Error("Expected Edit binding to be in FullHelp")
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
