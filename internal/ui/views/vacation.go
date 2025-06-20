package views

import (
    "fmt"
    "strconv"
    "time"

    "github.com/charmbracelet/bubbles/table"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "timesheet/internal/api"
)

// VacationView represents the vacation tracking view
type VacationView struct {
    table      table.Model
    inputs     []textinput.Model
    year       int
    yearSelect int
    ready      bool
}

// NewVacationView creates a new vacation view
func NewVacationView() *VacationView {
    // Create text inputs
    inputs := make([]textinput.Model, 3)
    inputs[0] = textinput.New()
    inputs[0].Placeholder = "Date"
    inputs[0].Width = 20

    inputs[1] = textinput.New()
    inputs[1].Placeholder = "Hours"
    inputs[1].Width = 10

    inputs[2] = textinput.New()
    inputs[2].Placeholder = "Notes"
    inputs[2].Width = 30

    // Create table
    columns := []table.Column{
        {Title: "Date", Width: 20},
        {Title: "Hours", Width: 10},
        {Title: "Category", Width: 15},
        {Title: "Notes", Width: 30},
    }

    t := table.New(
        table.WithColumns(columns),
        table.WithFocused(true),
        table.WithHeight(10),
    )

    s := table.DefaultStyles()
    s.Header = s.Header.
        BorderStyle(lipgloss.NormalBorder()).
        BorderForeground(lipgloss.Color("240"))
    s.Selected = s.Selected.
        Foreground(lipgloss.Color("229")).
        Background(lipgloss.Color("57")).
        Bold(true)
    t.SetStyles(s)

    return &VacationView{
        table:      t,
        inputs:     inputs,
        year:       time.Now().Year(),
        yearSelect: 1,
    }
}

// Init initializes the view
func (v VacationView) Init() tea.Cmd {
    return nil
}

// Update handles messages and updates the view
func (v VacationView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "tab":
            // Handle tab navigation between inputs
            for i := range v.inputs {
                if v.inputs[i].Focused() {
                    v.inputs[i].Blur()
                    if i < len(v.inputs)-1 {
                        v.inputs[i+1].Focus()
                    } else {
                        v.table.Focus()
                    }
                    break
                }
            }
        case "enter":
            if v.table.Focused() {
                // Handle table selection
            } else {
                // Add new vacation entry
                v.addVacation()
            }
        case "1", "2", "3":
            if msg.String() == "1" {
                v.year = time.Now().Year() - 1
            } else if msg.String() == "2" {
                v.year = time.Now().Year()
            } else {
                v.year = time.Now().Year() + 1
            }
            v.refreshTable()
        }
    }

    // Update inputs
    for i := range v.inputs {
        var cmd tea.Cmd
        v.inputs[i], cmd = v.inputs[i].Update(msg)
        cmds = append(cmds, cmd)
    }

    // Update table
    var cmd tea.Cmd
    v.table, cmd = v.table.Update(msg)
    cmds = append(cmds, cmd)

    return v, tea.Batch(cmds...)
}

// View renders the view
func (v VacationView) View() string {
    if !v.ready {
        return "Initializing..."
    }

    // Create year selector
    yearSelector := fmt.Sprintf("Year: %d [1: Previous] [2: Current] [3: Next]", v.year)

    // Create form
    form := lipgloss.JoinVertical(
        lipgloss.Left,
        v.inputs[0].View(),
        v.inputs[1].View(),
        v.inputs[2].View(),
        "[Enter] Add | [Tab] Navigate | [Ctrl+C] Quit",
    )

    // Combine all elements
    return lipgloss.JoinVertical(
        lipgloss.Left,
        yearSelector,
        v.table.View(),
        form,
    )
}

// addVacation adds a new vacation entry
func (v *VacationView) addVacation() {
    date := v.inputs[0].Value()
    hoursStr := v.inputs[1].Value()
    notes := v.inputs[2].Value()

    hours, err := strconv.Atoi(hoursStr)
    if err != nil {
        return
    }

    entry := api.VacationEntry{
        Date:     date,
        Hours:    hours,
        Category: "Vacation",
        Notes:    notes,
    }

    if err := api.CreateVacation(entry); err != nil {
        return
    }

    // Clear inputs
    for i := range v.inputs {
        v.inputs[i].Reset()
    }

    v.refreshTable()
}

// refreshTable updates the table with current data
func (v *VacationView) refreshTable() {
    data, err := api.GetVacation(v.year)
    if err != nil {
        return
    }

    // Convert entries to table rows
    rows := make([]table.Row, len(data.Entries))
    for i, entry := range data.Entries {
        rows[i] = table.Row{
            entry.Date,
            fmt.Sprintf("%d", entry.Hours),
            entry.Category,
            entry.Notes,
        }
    }

    // Add summary row
    rows = append(rows, table.Row{
        "Total",
        fmt.Sprintf("%d/%d", data.TotalHours, data.YearlyTarget),
        "Remaining",
        fmt.Sprintf("%d", data.Remaining),
    })

    v.table.SetRows(rows)
} 