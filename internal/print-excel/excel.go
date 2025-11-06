package printExcel

import (
	"fmt"
	"os"
	"timesheet/internal/config"

	"github.com/xuri/excelize/v2"
)

type TimesheetRow struct {
	Date          string
	ClientName    string
	ClientHours   float64
	TrainingHours float64
	VacationHours float64
	IdleHours     float64
	HolidayHours  float64
	SickHours     float64
}

func TimesheetToExcel(timesheetData []TimesheetRow) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Get user configuration
	name, company, freeSpeech, err := config.GetUserConfig()
	if err != nil {
		// Use default values if config cannot be read
		name = "Unknown User"
		company = "Unknown Company"
		freeSpeech = "Free Speech"
	}

	// Use Sheet1 instead of creating a new sheet
	sheetName := "Sheet1"

	logoPath := "assets/logo.jpg"
	if _, err := os.Stat(logoPath); os.IsNotExist(err) {
		logoPath = "docs/images/unicorn.jpg" // Fallback image
	}
	if _, err := os.Stat(logoPath); err == nil {
		f.AddPicture(sheetName, "A1", logoPath, nil)
	}

	if err != nil {
		fmt.Println(err)
	}

	// Add user information at the top
	f.SetCellValue(sheetName, "E1", "Name:")
	f.SetCellValue(sheetName, "F1", name)
	f.SetCellValue(sheetName, "E2", "Company:")
	f.SetCellValue(sheetName, "F2", company)
	f.SetCellValue(sheetName, "E3", "Free Speech:")
	f.SetCellValue(sheetName, "F3", freeSpeech)

	// Make the labels bold
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "0000FF", Style: 3},
			{Type: "top", Color: "00FF00", Style: 4},
			{Type: "bottom", Color: "FFFF00", Style: 5},
			{Type: "right", Color: "FF0000", Style: 6},
			{Type: "diagonalDown", Color: "A020F0", Style: 7},
			{Type: "diagonalUp", Color: "A020F0", Style: 8},
		},
		Font: &excelize.Font{
			Bold: true,
		},
	})
	f.SetCellStyle(sheetName, "A1", "A3", headerStyle)

	// Set header row
	headers := []string{
		"Date", "Client Name", "Client Hours", "Training Hours", "Vacation Hours",
		"Idle Hours", "Holiday Hours", "Sick Hours",
	}
	for i, header := range headers {
		cell := fmt.Sprintf("%s5", string(rune('A'+i)))
		f.SetCellValue(sheetName, cell, header)
	}

	// Add data starting from row 6
	for rowIndex, data := range timesheetData {
		excelRow := rowIndex + 6 // Start from row 6

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", excelRow), data.Date)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", excelRow), data.ClientName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", excelRow), data.ClientHours)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", excelRow), data.TrainingHours)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", excelRow), data.VacationHours)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", excelRow), data.IdleHours)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", excelRow), data.HolidayHours)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", excelRow), data.SickHours)
	}

	// Calculate totals for each column
	totalClientHours := 0.0
	totalTrainingHours := 0.0
	totalVacationHours := 0.0
	totalIdleHours := 0.0
	totalHolidayHours := 0.0
	totalSickHours := 0.0

	for _, data := range timesheetData {
		totalClientHours += data.ClientHours
		totalTrainingHours += data.TrainingHours
		totalVacationHours += data.VacationHours
		totalIdleHours += data.IdleHours
		totalHolidayHours += data.HolidayHours
		totalSickHours += data.SickHours
	}

	// Add totals row
	totalRowIndex := len(timesheetData) + 6 // Row after the last data row

	// Style for total row (bold)
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
	})

	// Add "Total" label and values
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", totalRowIndex), "Total")
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", totalRowIndex), totalClientHours)
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", totalRowIndex), totalTrainingHours)
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", totalRowIndex), totalVacationHours)
	f.SetCellValue(sheetName, fmt.Sprintf("F%d", totalRowIndex), totalIdleHours)
	f.SetCellValue(sheetName, fmt.Sprintf("G%d", totalRowIndex), totalHolidayHours)
	f.SetCellValue(sheetName, fmt.Sprintf("H%d", totalRowIndex), totalSickHours)

	// Apply bold style to the total row
	for _, col := range []string{"A", "B", "C", "D", "E", "F", "G", "H"} {
		cellRef := fmt.Sprintf("%s%d", col, totalRowIndex)
		f.SetCellStyle(sheetName, cellRef, cellRef, style)
	}

	// Save spreadsheet
	if err := f.SaveAs("Timesheet.xlsx"); err != nil {
		return fmt.Errorf("failed to save excel file: %w", err)
	}

	return nil
}
