package config

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)

// readInput reads a line of input from stdin
func readInput() string {
    reader := bufio.NewReader(os.Stdin)
    input, _ := reader.ReadString('\n')
    return strings.TrimSpace(input)
}

// SetupConfig creates a new configuration file
func SetupConfig() error {
    fmt.Print("Enter your name: ")
    name := readInput()

    fmt.Print("Enter your company name: ")
    companyName := readInput()

    fmt.Print("Enter your free speech (default: 'I confirm that the information provided is accurate'): ")
    freeSpeech := readInput()
    if freeSpeech == "" {
        freeSpeech = "I confirm that the information provided is accurate"
    }

    fmt.Print("Start API server? (y/n, default: y): ")
    startAPIServerStr := readInput()
    startAPIServer := startAPIServerStr != "n"

    fmt.Print("Enter API port (default: 8080): ")
    apiPortStr := readInput()
    apiPort, err := strconv.Atoi(apiPortStr)
    if err != nil || apiPort <= 0 {
        apiPort = 8080
    }

    fmt.Print("Enable development mode? (y/n, default: n): ")
    developmentModeStr := readInput()
    developmentMode := developmentModeStr == "y"

    fmt.Print("Enter document type to send (default: pdf): ")
    sendDocumentType := readInput()
    if sendDocumentType == "" {
        sendDocumentType = "pdf"
    }

    fmt.Print("Send to others? (y/n, default: n): ")
    sendToOthersStr := readInput()
    sendToOthers := sendToOthersStr == "y"

    fmt.Print("Enter recipient email: ")
    recipientEmail := readInput()

    fmt.Print("Enter sender email: ")
    senderEmail := readInput()

    fmt.Print("Enter reply-to email: ")
    replyToEmail := readInput()

    fmt.Print("Enter Resend API key: ")
    resendAPIKey := readInput()

    // Training hours setup
    fmt.Print("\nTraining Hours Setup:\n")
    fmt.Print("Enter yearly training hours target (default: 40): ")
    trainingHoursStr := readInput()
    trainingHours, err := strconv.Atoi(trainingHoursStr)
    if err != nil || trainingHours <= 0 {
        trainingHours = 40
    }

    // Vacation hours setup
    fmt.Print("\nVacation Hours Setup:\n")
    fmt.Print("Enter yearly vacation hours target (default: 180): ")
    vacationHoursStr := readInput()
    vacationHours, err := strconv.Atoi(vacationHoursStr)
    if err != nil || vacationHours <= 0 {
        vacationHours = 180
    }

    config := Config{
        Name:            name,
        CompanyName:     companyName,
        FreeSpeech:      freeSpeech,
        StartAPIServer:  startAPIServer,
        APIPort:         apiPort,
        DevelopmentMode: developmentMode,
        SendDocumentType: sendDocumentType,
        SendToOthers:    sendToOthers,
        RecipientEmail:  recipientEmail,
        SenderEmail:     senderEmail,
        ReplyToEmail:    replyToEmail,
        ResendAPIKey:    resendAPIKey,
        TrainingHours: TrainingHours{
            YearlyTarget: trainingHours,
            Category:     "Training",
        },
        VacationHours: VacationHours{
            YearlyTarget: vacationHours,
            Category:     "Vacation",
        },
    }

    if err := SaveConfig(config); err != nil {
        return fmt.Errorf("failed to save config: %w", err)
    }

    return nil
} 