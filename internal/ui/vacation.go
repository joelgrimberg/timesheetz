package ui

import (
    "fmt"
    "time"
    "timesheet/internal/api"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/charmbracelet/bubbles/table"
)

// VacationModel represents the vacation view
type VacationModel struct {
    entries     []api.VacationEntry
    yearlyTarget int
    totalHours  int
    remaining   int
    year        int
    ready       bool
}

// InitialVacationModel creates a new vacation model
func InitialVacationModel() VacationModel {
    return VacationModel{
        year:  time.Now().Year(),
        ready: true,
    }
}

// Init initializes the model
func (m VacationModel) Init() tea.Cmd {
    return m.fetchVacationData
}

// Update handles messages and updates the model
func (m VacationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "r":
            return m, m.fetchVacationData
        case "left", "h":
            // Decrease year
            m.year--
            return m, m.fetchVacationData
        case "right", "l":
            // Increase year
            m.year++
            return m, m.fetchVacationData
        }
    case vacationDataMsg:
        m.entries = msg.entries
        m.yearlyTarget = msg.yearlyTarget
        m.totalHours = msg.totalHours
        m.remaining = msg.remaining
        m.ready = true
    }
    return m, nil
}

// View renders the model
func (m VacationModel) View() string {
    if !m.ready {
        return "Loading vacation data..."
    }

    var s string

    // Show the year as title
    yearTitle := fmt.Sprintf("Vacation %d", m.year)
    s += titleStyle.Render(yearTitle) + "\n"

    // Create columns for the table
    columns := []table.Column{
        {Title: "Date", Width: 12},
        {Title: "Hours", Width: 8},
    }

    // Create the table
    t := table.New(
        table.WithColumns(columns),
        table.WithFocused(true),
        table.WithHeight(15),
    )

    // Set styles
    tableStyles := table.DefaultStyles()
    tableStyles.Header = tableStyles.Header.
        BorderStyle(lipgloss.NormalBorder()).
        BorderForeground(lipgloss.Color("240")).
        BorderBottom(true).
        Bold(false)
    tableStyles.Selected = tableStyles.Selected.
        Foreground(lipgloss.Color("229")).
        Background(lipgloss.Color("57")).
        Bold(false)
    t.SetStyles(tableStyles)

    // Convert entries to table rows
    var rows []table.Row
    for _, entry := range m.entries {
        rows = append(rows, table.Row{
            entry.Date,
            fmt.Sprintf("%d", entry.Hours),
        })
    }

    // Add total row
    rows = append(rows, table.Row{
        "Total",
        fmt.Sprintf("%d/%d", m.totalHours, m.yearlyTarget),
    })

    t.SetRows(rows)

    // Get the table view
    tableView := t.View()

    // Render the table with baseStyle
    s += baseStyle.Render(tableView) + "\n"

    // Add help text
    s += helpStyle.Render("Controls: r: Refresh  ←/→: Change Year  q: Quit\n")

    return s
}

// vacationDataMsg is sent when vacation data is fetched
type vacationDataMsg struct {
    entries     []api.VacationEntry
    yearlyTarget int
    totalHours  int
    remaining   int
}

// errorMsg is sent when an error occurs
type errorMsg struct {
    err error
}

// fetchVacationData fetches vacation data from the API
func (m VacationModel) fetchVacationData() tea.Msg {
    data, err := api.GetVacation(m.year)
    if err != nil {
        return errorMsg{err}
    }
    return vacationDataMsg{
        entries:     data.Entries,
        yearlyTarget: data.YearlyTarget,
        totalHours:  data.TotalHours,
        remaining:   data.Remaining,
    }
} 