package printPDF

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"timesheet/internal/email"
	"unicode"

	"github.com/jung-kurt/gofpdf"
)

// stripANSI removes ANSI escape sequences, replaces box-drawing characters, and handles emojis
func stripANSI(str string) string {
	// Remove ANSI escape sequences
	re := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\[[0-9;]*[a-zA-Z]`)
	str = re.ReplaceAllString(str, "")

	// Replace box-drawing characters with ASCII equivalents
	replacements := map[rune]string{
		'┌': "+", // top-left corner (U+250C)
		'┐': "+", // top-right corner (U+2510)
		'└': "+", // bottom-left corner (U+2514)
		'┘': "+", // bottom-right corner (U+2518)
		'─': "-", // horizontal line (U+2500)
		'│': "|", // vertical line (U+2502)
		'├': "+", // left T (U+251C)
		'┤': "+", // right T (U+2524)
		'┬': "+", // top T (U+252C)
		'┴': "+", // bottom T (U+2534)
		'┼': "+", // cross (U+253C)
		// Emoji replacements
		'💤': "  ", // person emoji
	}

	// Remove or replace other non-ASCII characters
	var result strings.Builder
	for _, r := range str {
		if r < 128 { // ASCII characters
			result.WriteRune(r)
		} else if replacement, ok := replacements[r]; ok {
			result.WriteString(replacement)
		} else if unicode.IsPrint(r) {
			// For other printable Unicode characters, try to keep them
			// but if they don't render well, you might want to replace or skip them
			result.WriteRune(r)
		}
		// Skip non-printable, non-ASCII characters that aren't in the replacements map
	}

	return result.String()
}

// TimesheetToPDF converts a timesheet view to a PDF file
func TimesheetToPDF(viewContent string, sendAsEmail bool) (string, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Courier", "", 10) // Monospaced font works better for tabular data

	// Clean the view content
	viewContent = stripANSI(viewContent)
	lines := strings.Split(viewContent, "\n")

	// Remove the last line (if there are any lines)
	if len(lines) > 0 {
		lines = lines[:len(lines)-1]
	}

	// Set starting position
	y := 10.0
	lineHeight := 5.0

	// Add each line to the PDF
	for _, line := range lines {
		// Special formatting for the total line
		if strings.HasPrefix(line, "Total:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				pdf.Text(10, y, "Total:")
				pdf.Text(124, y, strings.TrimSpace(parts[1])) // Position the numbers at x=50
			} else {
				pdf.Text(10, y, line)
			}
		} else {
			pdf.Text(10, y, line)
		}
		y += lineHeight
	}

	// Save the PDF with a more descriptive filename
	filename := fmt.Sprintf("timesheet_%s.pdf", time.Now().Format("2006-01-02"))
	err := pdf.OutputFileAndClose(filename)
	if err != nil {
		return "", err
	}

	if sendAsEmail {
		email.EmailAttachment(filename)
	}

	return filename, nil
}
