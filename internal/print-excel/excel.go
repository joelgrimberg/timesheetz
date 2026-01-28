package printExcel

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
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

type excelTranslations struct {
	Headers        []string
	HoursTotal     string
	Month          string
	Year           string
	Client         string
	Project        string
	NameConsultant string
	HoursReport    string
	FilePrefix     string // "Urensheet" or "Timesheet"
	FileIntern     string // "intern" or "internal"
	MonthAbbrevs   []string
}

func getTranslations(lang string) excelTranslations {
	if lang == "nl" {
		return excelTranslations{
			Headers:        []string{"Dag", "Gewerkt", "Overwerk", "Ziekte", "Verlof", "Feestdag", "Beschikbaar", "Opleiding", "Overig", "Stand-By", "Kilometers", "Toelichting"},
			HoursTotal:     "Uren totaal",
			Month:          "Maand",
			Year:           "Jaar",
			Client:         "Klant",
			Project:        "Project",
			NameConsultant: "Naam Consultant",
			HoursReport:    "Urenverantwoording",
			FilePrefix:     "Urensheet",
		FileIntern:     "intern",
			MonthAbbrevs:   []string{"jan", "feb", "mrt", "apr", "mei", "jun", "jul", "aug", "sep", "okt", "nov", "dec"},
		}
	}
	return excelTranslations{
		Headers:        []string{"Day", "Worked", "Overtime", "Sick", "Leave", "Holiday", "Available", "Training", "Other", "Stand-By", "Kilometers", "Notes"},
		HoursTotal:     "Hours total",
		Month:          "Month",
		Year:           "Year",
		Client:         "Client",
		Project:        "Project",
		NameConsultant: "Name Consultant",
		HoursReport:    "Hours report",
		FilePrefix:     "Timesheet",
		FileIntern:     "internal",
		MonthAbbrevs:   []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
	}
}

func TimesheetToExcel(timesheetData []TimesheetRow, year int, month time.Month) (string, error) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Get user configuration
	name, company, _, err := config.GetUserConfig()
	if err != nil {
		name = "Unknown User"
		company = "Unknown Company"
	}

	// Get client name from first entry (or empty)
	clientName := ""
	if len(timesheetData) > 0 {
		clientName = timesheetData[0].ClientName
	}

	sheetName := "Sheet1"

	// Hide gridlines
	showGridLines := false
	f.SetSheetView(sheetName, 0, &excelize.ViewOptions{
		ShowGridLines: &showGridLines,
	})

	// Add logo if it exists at ~/.config/timesheetz/logo.png
	if homeDir, err := os.UserHomeDir(); err == nil {
		logoPath := filepath.Join(homeDir, ".config", "timesheetz", "logo.png")
		if _, err := os.Stat(logoPath); err == nil {
			f.AddPicture(sheetName, "A1", logoPath, &excelize.GraphicOptions{
				ScaleX:  0.5,
				ScaleY:  0.5,
				Positioning: "oneCell",
			})
		}
	}

	// Load translations
	lang := config.GetExportLanguage()
	t := getTranslations(lang)

	// Build a map of day -> data for quick lookup
	dayData := make(map[int]TimesheetRow)
	for _, row := range timesheetData {
		// Parse day from date (format: 2006-01-02)
		t, err := time.Parse("2006-01-02", row.Date)
		if err == nil {
			dayData[t.Day()] = row
		}
	}

	// Define fonts
	defaultFont := &excelize.Font{Family: "Tahoma", Size: 12}
	boldFont := &excelize.Font{Family: "Tahoma", Size: 12, Bold: true}

	// Set column widths (base width * 1.5)
	f.SetColWidth(sheetName, "A", "A", 3)          // Spacing column
	f.SetColWidth(sheetName, "B", "B", 13.5)       // Dag (1.5x wider)
	f.SetColWidth(sheetName, "C", "C", 15)         // Gewerkt
	f.SetColWidth(sheetName, "D", "D", 15)         // Overwerk
	f.SetColWidth(sheetName, "E", "E", 12)         // Ziekte
	f.SetColWidth(sheetName, "F", "F", 12)         // Verlof
	f.SetColWidth(sheetName, "G", "G", 15)         // Feestdag
	f.SetColWidth(sheetName, "H", "H", 18)         // Beschikbaar
	f.SetColWidth(sheetName, "I", "I", 15)         // Opleiding
	f.SetColWidth(sheetName, "J", "J", 12)         // Overig
	f.SetColWidth(sheetName, "K", "K", 15)         // Stand-By
	f.SetColWidth(sheetName, "L", "L", 18)         // Kilometers
	f.SetColWidth(sheetName, "M", "M", 18)         // Toelichting
	f.SetColWidth(sheetName, "N", "N", 30)         // Header info column

	// Style for header info text
	infoStyle, _ := f.NewStyle(&excelize.Style{Font: defaultFont})
	infoBoldStyle, _ := f.NewStyle(&excelize.Style{Font: boldFont})

	// Header section (matching deTesters format)
	f.SetCellValue(sheetName, "N3", fmt.Sprintf("%s", company))
	f.SetCellStyle(sheetName, "N3", "N3", infoBoldStyle)
	f.SetCellValue(sheetName, "N5", t.HoursReport)
	f.SetCellStyle(sheetName, "N5", "N5", infoBoldStyle)
	f.SetCellValue(sheetName, "N7", fmt.Sprintf("%s : %d", t.Month, month))
	f.SetCellStyle(sheetName, "N7", "N7", infoStyle)
	f.SetCellValue(sheetName, "N8", fmt.Sprintf("%s : %d", t.Year, year))
	f.SetCellStyle(sheetName, "N8", "N8", infoStyle)
	f.SetCellValue(sheetName, "N10", fmt.Sprintf("%s : %s", t.Client, clientName))
	f.SetCellStyle(sheetName, "N10", "N10", infoStyle)
	f.SetCellValue(sheetName, "N11", fmt.Sprintf("%s :", t.Project))
	f.SetCellStyle(sheetName, "N11", "N11", infoStyle)
	f.SetCellValue(sheetName, "B14", fmt.Sprintf("%s:", t.NameConsultant))
	f.SetCellStyle(sheetName, "B14", "B14", infoBoldStyle)
	f.SetCellValue(sheetName, "E14", name)
	f.SetCellStyle(sheetName, "E14", "E14", infoStyle)

	// Header border styles (rows 17-19) - same color as footer (#027B8D)
	headerBorderColor := "027B8D"

	// Row 17 styles (top row)
	styleTopLeft, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "top", Color: headerBorderColor, Style: 1},
			{Type: "left", Color: headerBorderColor, Style: 1},
		},
	})
	styleTop, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "top", Color: headerBorderColor, Style: 1},
		},
	})
	styleTopRight, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "top", Color: headerBorderColor, Style: 1},
			{Type: "right", Color: headerBorderColor, Style: 1},
		},
	})

	// Row 18 styles (middle row with text) - centered
	centerAlign := &excelize.Alignment{Horizontal: "center"}
	styleLeft, _ := f.NewStyle(&excelize.Style{
		Font:      boldFont,
		Alignment: centerAlign,
		Border: []excelize.Border{
			{Type: "left", Color: headerBorderColor, Style: 1},
		},
	})
	styleMiddle, _ := f.NewStyle(&excelize.Style{
		Font:      boldFont,
		Alignment: centerAlign,
	})
	styleRight, _ := f.NewStyle(&excelize.Style{
		Font:      boldFont,
		Alignment: centerAlign,
		Border: []excelize.Border{
			{Type: "right", Color: headerBorderColor, Style: 1},
		},
	})

	// Row 19 styles (bottom row)
	styleBottomLeft, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "bottom", Color: headerBorderColor, Style: 1},
			{Type: "left", Color: headerBorderColor, Style: 1},
		},
	})
	styleBottom, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "bottom", Color: headerBorderColor, Style: 1},
		},
	})
	styleBottomRight, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "bottom", Color: headerBorderColor, Style: 1},
			{Type: "right", Color: headerBorderColor, Style: 1},
		},
	})

	// Apply styles to row 17 (top of header)
	f.SetCellStyle(sheetName, "B17", "B17", styleTopLeft)
	f.SetCellStyle(sheetName, "C17", "L17", styleTop)
	f.SetCellStyle(sheetName, "M17", "M17", styleTopRight)

	// Column headers (row 18)
	for i, header := range t.Headers {
		cell := fmt.Sprintf("%s18", string(rune('B'+i)))
		f.SetCellValue(sheetName, cell, header)
	}
	f.SetCellStyle(sheetName, "B18", "B18", styleLeft)
	f.SetCellStyle(sheetName, "C18", "L18", styleMiddle)
	f.SetCellStyle(sheetName, "M18", "M18", styleRight)

	// Apply styles to row 19 (bottom of header)
	f.SetCellStyle(sheetName, "B19", "B19", styleBottomLeft)
	f.SetCellStyle(sheetName, "C19", "L19", styleBottom)
	f.SetCellStyle(sheetName, "M19", "M19", styleBottomRight)

	// Set row height for header rows (1.5x default of 15 = 22.5)
	rowHeight := 22.5
	f.SetRowHeight(sheetName, 17, rowHeight)
	f.SetRowHeight(sheetName, 18, rowHeight)
	f.SetRowHeight(sheetName, 19, rowHeight)

	// Get number of days in month
	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()

	// Table spans from row 20 to totalRow+1 (daysInMonth + 21)
	firstDataRow := 20

	// Border color for table
	borderColor := "027B8D"

	// Weekend background fill (light grey)
	weekendFill := &excelize.Fill{Type: "pattern", Color: []string{"D9D9D9"}, Pattern: 1}

	// Border styles for data table - outer border only
	// Top row styles
	dataTopLeft, _ := f.NewStyle(&excelize.Style{
		Font:      defaultFont,
		Alignment: centerAlign,
		Border: []excelize.Border{
			{Type: "top", Color: borderColor, Style: 1},
			{Type: "left", Color: borderColor, Style: 1},
		},
	})
	dataTop, _ := f.NewStyle(&excelize.Style{
		Font:      defaultFont,
		Alignment: centerAlign,
		Border: []excelize.Border{
			{Type: "top", Color: borderColor, Style: 1},
		},
	})
	dataTopRight, _ := f.NewStyle(&excelize.Style{
		Font:      defaultFont,
		Alignment: centerAlign,
		Border: []excelize.Border{
			{Type: "top", Color: borderColor, Style: 1},
			{Type: "right", Color: borderColor, Style: 1},
		},
	})

	// Top row styles for weekends (with grey background)
	dataTopLeftWeekend, _ := f.NewStyle(&excelize.Style{
		Font:      defaultFont,
		Alignment: centerAlign,
		Fill:      *weekendFill,
		Border: []excelize.Border{
			{Type: "top", Color: borderColor, Style: 1},
			{Type: "left", Color: borderColor, Style: 1},
		},
	})
	dataTopWeekend, _ := f.NewStyle(&excelize.Style{
		Font:      defaultFont,
		Alignment: centerAlign,
		Fill:      *weekendFill,
		Border: []excelize.Border{
			{Type: "top", Color: borderColor, Style: 1},
		},
	})
	dataTopRightWeekend, _ := f.NewStyle(&excelize.Style{
		Font:      defaultFont,
		Alignment: centerAlign,
		Fill:      *weekendFill,
		Border: []excelize.Border{
			{Type: "top", Color: borderColor, Style: 1},
			{Type: "right", Color: borderColor, Style: 1},
		},
	})

	// Middle row styles
	dataLeft, _ := f.NewStyle(&excelize.Style{
		Font:      defaultFont,
		Alignment: centerAlign,
		Border: []excelize.Border{
			{Type: "left", Color: borderColor, Style: 1},
		},
	})
	dataMiddle, _ := f.NewStyle(&excelize.Style{
		Font:      defaultFont,
		Alignment: centerAlign,
	})
	dataRight, _ := f.NewStyle(&excelize.Style{
		Font:      defaultFont,
		Alignment: centerAlign,
		Border: []excelize.Border{
			{Type: "right", Color: borderColor, Style: 1},
		},
	})

	// Middle row styles for weekends (with grey background)
	dataLeftWeekend, _ := f.NewStyle(&excelize.Style{
		Font:      defaultFont,
		Alignment: centerAlign,
		Fill:      *weekendFill,
		Border: []excelize.Border{
			{Type: "left", Color: borderColor, Style: 1},
		},
	})
	dataMiddleWeekend, _ := f.NewStyle(&excelize.Style{
		Font:      defaultFont,
		Alignment: centerAlign,
		Fill:      *weekendFill,
	})
	dataRightWeekend, _ := f.NewStyle(&excelize.Style{
		Font:      defaultFont,
		Alignment: centerAlign,
		Fill:      *weekendFill,
		Border: []excelize.Border{
			{Type: "right", Color: borderColor, Style: 1},
		},
	})

	// Bottom row styles (bold for totals)
	dataBottomLeft, _ := f.NewStyle(&excelize.Style{
		Font:      boldFont,
		Alignment: centerAlign,
		Border: []excelize.Border{
			{Type: "bottom", Color: borderColor, Style: 1},
			{Type: "left", Color: borderColor, Style: 1},
		},
	})
	dataBottom, _ := f.NewStyle(&excelize.Style{
		Font:      boldFont,
		Alignment: centerAlign,
		Border: []excelize.Border{
			{Type: "bottom", Color: borderColor, Style: 1},
		},
	})
	dataBottomRight, _ := f.NewStyle(&excelize.Style{
		Font:      boldFont,
		Alignment: centerAlign,
		Border: []excelize.Border{
			{Type: "bottom", Color: borderColor, Style: 1},
			{Type: "right", Color: borderColor, Style: 1},
		},
	})

	// Footer top row styles (thin top line + left/right, bold for label)
	footerTopLeft, _ := f.NewStyle(&excelize.Style{
		Font:      boldFont,
		Alignment: centerAlign,
		Border: []excelize.Border{
			{Type: "top", Color: borderColor, Style: 1},
			{Type: "left", Color: borderColor, Style: 1},
		},
	})
	footerTop, _ := f.NewStyle(&excelize.Style{
		Font:      boldFont,
		Alignment: centerAlign,
		Border: []excelize.Border{
			{Type: "top", Color: borderColor, Style: 1},
		},
	})
	footerTopRight, _ := f.NewStyle(&excelize.Style{
		Font:      boldFont,
		Alignment: centerAlign,
		Border: []excelize.Border{
			{Type: "top", Color: borderColor, Style: 1},
			{Type: "right", Color: borderColor, Style: 1},
		},
	})

	// Middle row styles for totals rows (bold, left/right only)
	totalLeft, _ := f.NewStyle(&excelize.Style{
		Font:      boldFont,
		Alignment: centerAlign,
		Border: []excelize.Border{
			{Type: "left", Color: borderColor, Style: 1},
		},
	})
	totalMiddle, _ := f.NewStyle(&excelize.Style{
		Font:      boldFont,
		Alignment: centerAlign,
	})
	totalRight, _ := f.NewStyle(&excelize.Style{
		Font:      boldFont,
		Alignment: centerAlign,
		Border: []excelize.Border{
			{Type: "right", Color: borderColor, Style: 1},
		},
	})

	// Totals
	var totalGewerkt, totalOverwerk, totalZiekte, totalVerlof, totalFeestdag float64
	var totalBeschikbaar, totalOpleiding, totalOverig, totalStandBy, totalKilometers float64

	// Data rows - one row per day of month (starting at row 20)
	for day := 1; day <= daysInMonth; day++ {
		excelRow := day + 19 // Day 1 = row 20

		// Set row height (1.5x)
		f.SetRowHeight(sheetName, excelRow, rowHeight)

		// Check if this day is a weekend (Saturday=6, Sunday=0)
		date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		isWeekend := date.Weekday() == time.Saturday || date.Weekday() == time.Sunday

		// Day number
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", excelRow), day)

		// Apply appropriate border style based on position and weekend
		if excelRow == firstDataRow {
			// First row - top border
			if isWeekend {
				f.SetCellStyle(sheetName, fmt.Sprintf("B%d", excelRow), fmt.Sprintf("B%d", excelRow), dataTopLeftWeekend)
				f.SetCellStyle(sheetName, fmt.Sprintf("C%d", excelRow), fmt.Sprintf("L%d", excelRow), dataTopWeekend)
				f.SetCellStyle(sheetName, fmt.Sprintf("M%d", excelRow), fmt.Sprintf("M%d", excelRow), dataTopRightWeekend)
			} else {
				f.SetCellStyle(sheetName, fmt.Sprintf("B%d", excelRow), fmt.Sprintf("B%d", excelRow), dataTopLeft)
				f.SetCellStyle(sheetName, fmt.Sprintf("C%d", excelRow), fmt.Sprintf("L%d", excelRow), dataTop)
				f.SetCellStyle(sheetName, fmt.Sprintf("M%d", excelRow), fmt.Sprintf("M%d", excelRow), dataTopRight)
			}
		} else {
			// Middle rows - left/right border only
			if isWeekend {
				f.SetCellStyle(sheetName, fmt.Sprintf("B%d", excelRow), fmt.Sprintf("B%d", excelRow), dataLeftWeekend)
				f.SetCellStyle(sheetName, fmt.Sprintf("C%d", excelRow), fmt.Sprintf("L%d", excelRow), dataMiddleWeekend)
				f.SetCellStyle(sheetName, fmt.Sprintf("M%d", excelRow), fmt.Sprintf("M%d", excelRow), dataRightWeekend)
			} else {
				f.SetCellStyle(sheetName, fmt.Sprintf("B%d", excelRow), fmt.Sprintf("B%d", excelRow), dataLeft)
				f.SetCellStyle(sheetName, fmt.Sprintf("C%d", excelRow), fmt.Sprintf("L%d", excelRow), dataMiddle)
				f.SetCellStyle(sheetName, fmt.Sprintf("M%d", excelRow), fmt.Sprintf("M%d", excelRow), dataRight)
			}
		}

		// Fill in data if we have it for this day
		if data, ok := dayData[day]; ok {
			if data.ClientHours > 0 {
				f.SetCellValue(sheetName, fmt.Sprintf("C%d", excelRow), data.ClientHours)
				totalGewerkt += data.ClientHours
			}
			// Overwerk (overtime) - we don't track this, leave empty
			if data.SickHours > 0 {
				f.SetCellValue(sheetName, fmt.Sprintf("E%d", excelRow), data.SickHours)
				totalZiekte += data.SickHours
			}
			if data.VacationHours > 0 {
				f.SetCellValue(sheetName, fmt.Sprintf("F%d", excelRow), data.VacationHours)
				totalVerlof += data.VacationHours
			}
			if data.HolidayHours > 0 {
				f.SetCellValue(sheetName, fmt.Sprintf("G%d", excelRow), data.HolidayHours)
				totalFeestdag += data.HolidayHours
			}
			if data.IdleHours > 0 {
				f.SetCellValue(sheetName, fmt.Sprintf("H%d", excelRow), data.IdleHours)
				totalBeschikbaar += data.IdleHours
			}
			if data.TrainingHours > 0 {
				f.SetCellValue(sheetName, fmt.Sprintf("I%d", excelRow), data.TrainingHours)
				totalOpleiding += data.TrainingHours
			}
		}
	}

	// Footer section - 3 rows like header
	// Row 1: top border only (empty)
	// Row 2: content (Uren totaal + values)
	// Row 3: bottom border only (empty)
	footerRow1 := daysInMonth + 20
	footerRow2 := footerRow1 + 1 // Content row
	footerRow3 := footerRow1 + 2

	// Set row height for footer rows (1.5x)
	f.SetRowHeight(sheetName, footerRow1, rowHeight)
	f.SetRowHeight(sheetName, footerRow2, rowHeight)
	f.SetRowHeight(sheetName, footerRow3, rowHeight)

	// Calculate grand total (sum of all hour categories)
	grandTotal := totalGewerkt + totalOverwerk + totalZiekte + totalVerlof + totalFeestdag + totalBeschikbaar + totalOpleiding + totalOverig + totalStandBy

	// Set hours total label in footerRow1 (top row of footer)
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", footerRow1), t.HoursTotal)

	// Set content in middle row (footerRow2) - values aligned under their header columns
	// B=grandTotal, C=Gewerkt, D=Overwerk, E=Ziekte, F=Verlof, G=Feestdag, H=Beschikbaar, I=Opleiding, J=Overig, K=Stand-By, L=Kilometers, M=Toelichting
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", footerRow2), grandTotal)
	if totalGewerkt > 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", footerRow2), totalGewerkt)
	}
	if totalOverwerk > 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", footerRow2), totalOverwerk)
	}
	if totalZiekte > 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", footerRow2), totalZiekte)
	}
	if totalVerlof > 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", footerRow2), totalVerlof)
	}
	if totalFeestdag > 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", footerRow2), totalFeestdag)
	}
	if totalBeschikbaar > 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", footerRow2), totalBeschikbaar)
	}
	if totalOpleiding > 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", footerRow2), totalOpleiding)
	}
	if totalOverig > 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", footerRow2), totalOverig)
	}
	if totalStandBy > 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", footerRow2), totalStandBy)
	}
	if totalKilometers > 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", footerRow2), totalKilometers)
	}

	// Apply styles to footer rows - 3 rows like header
	// Row 1 (top of footer) - thin top border + left/right
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", footerRow1), fmt.Sprintf("B%d", footerRow1), footerTopLeft)
	f.SetCellStyle(sheetName, fmt.Sprintf("C%d", footerRow1), fmt.Sprintf("L%d", footerRow1), footerTop)
	f.SetCellStyle(sheetName, fmt.Sprintf("M%d", footerRow1), fmt.Sprintf("M%d", footerRow1), footerTopRight)

	// Row 2 (content) - left/right only
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", footerRow2), fmt.Sprintf("B%d", footerRow2), totalLeft)
	f.SetCellStyle(sheetName, fmt.Sprintf("C%d", footerRow2), fmt.Sprintf("L%d", footerRow2), totalMiddle)
	f.SetCellStyle(sheetName, fmt.Sprintf("M%d", footerRow2), fmt.Sprintf("M%d", footerRow2), totalRight)

	// Row 3 (bottom of footer) - bottom border + left/right
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", footerRow3), fmt.Sprintf("B%d", footerRow3), dataBottomLeft)
	f.SetCellStyle(sheetName, fmt.Sprintf("C%d", footerRow3), fmt.Sprintf("L%d", footerRow3), dataBottom)
	f.SetCellStyle(sheetName, fmt.Sprintf("M%d", footerRow3), fmt.Sprintf("M%d", footerRow3), dataBottomRight)

	// Generate filename with month and year
	monthAbbrev := t.MonthAbbrevs[month-1]
	companyClean := strings.ReplaceAll(company, " ", "")
	filename := fmt.Sprintf("%s_%s_%s_%s_%d.xlsx", t.FilePrefix, companyClean, t.FileIntern, monthAbbrev, year)
	if err := f.SaveAs(filename); err != nil {
		return "", fmt.Errorf("failed to save excel file: %w", err)
	}

	return filename, nil
}
