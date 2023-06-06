package clockifyExportProcessor

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"errors"

	"strconv"
	"time"

	"github.com/atotto/clipboard"
)

type Command struct {
	fs *flag.FlagSet
}

func Cmd() *Command {
	gc := &Command{
		fs: flag.NewFlagSet("clockifyExportProcessor", flag.ContinueOnError),
	}

	return gc
}

func (g *Command) Name() string {
	return g.fs.Name()
}

func (g *Command) Init(args []string) error {
	return g.fs.Parse(args)
}

func (g *Command) Run() error {
	// Get the clipboardContent of the clipboard
	clipboardContent, clipboardErr := readClipboard()
	if clipboardErr != nil {
		return nil
	}

	// Print the content of the clipboard
	fmt.Println("Got the following content from clipboard:", clipboardContent)

	// Check if the file exists
	fileErr := checkIfFileExists(clipboardContent)
	if fileErr != nil {
		return nil
	}

	// Check if the file is a CSV file
	csvCheckError := isCsvFile(clipboardContent)
	if csvCheckError != nil {
		return nil
	}

	// File exists
	fmt.Println("File exists and seems to be a valid CSV file:", clipboardContent)

	// Process the CSV file
	csvProcessErr := processCSVFile(clipboardContent)
	if csvProcessErr != nil {
		return nil
	}

	// sub-script end
	return nil
}

func readClipboard() (string, error) {
	content, err := clipboard.ReadAll()
	if err != nil {
		fmt.Println("Error reading clipboard:", err)
		return "", err
	}
	return content, nil
}

func checkIfFileExists(filePath string) error {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("File does not exist:", filePath)
		} else {
			fmt.Println("Error checking file existence:", err)
		}
		return err
	}
	return nil
}

func isCsvFile(filePath string) error {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".csv" { // has no correct file extension
		fmt.Println("File is not a CSV file:", filePath)
		return errors.New("File is not a CSV file:" + filePath)
	}
	return nil
}

func processCSVFile(filePath string) error {
	// Open the CSV file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err) //TODO use this also in other places instead of `return nil`?
	}
	defer file.Close()

	// Read the CSV data
	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading CSV: %w", err)
	}

	// Identify the columns to be dropped
	columnsToDrop := []string{
		"Client",
		"Task",
		"User",
		"Group",
		"Email",
		"Tags",
		"Start Time",
		"End Date",
		"End Time",
		"Duration (h)",
		"Billable Rate (EUR)",
		"Billable Amount (EUR)",
	}

	// Determine the column indices to drop
	columnIndices := make([]int, 0, len(columnsToDrop))
	headerRow := rows[0]
	for i, columnName := range headerRow {
		for _, columnToDrop := range columnsToDrop {
			if columnName == columnToDrop {
				columnIndices = append(columnIndices, i)
				break
			}
		}
	}

	// Create a map to store the combined rows
	combinedRows := make(map[string]float64)

	// Process the rows and combine rows with the same Project, Description, Start Date, and Billable
	for _, row := range rows[1:] {
		// save some columns without modifications
		project := row[0]
		description := row[2]
		billable := row[8]

		// parse duration as decimal
		durationDecimal, err := strconv.ParseFloat(row[14], 64)
		if err != nil {
			return fmt.Errorf("error parsing duration decimal: %w", err)
		}

		// parse start date as date in format `yyyy-mm-dd`
		startDateParsed, err := time.Parse("02/01/2006", row[9])
		if err != nil {
			return fmt.Errorf("error parsing date: %w", err)
		}
		startDate := startDateParsed.Format("2006-01-02")

		// save
		key := fmt.Sprintf("%s|%s|%s|%s", project, description, billable, startDate)
		combinedRows[key] += durationDecimal
	}

	// Generate the output file path
	outputFilePath := generateOutputFilePath(filePath)

	// Open the output CSV file
	newFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer newFile.Close()

	// Write the header row to the output CSV file
	err = writeHeader(newFile, headerRow, columnIndices)
	if err != nil {
		return fmt.Errorf("error writing output CSV header: %w", err)
	}

	// Write the combined rows to the output CSV file
	writer := csv.NewWriter(newFile)
	for key, durationDecimal := range combinedRows {
		combinedRow := strings.Split(key, "|")
		combinedRow = append(combinedRow, strconv.FormatFloat(durationDecimal, 'f', 2, 64))

		// Calculate the time string representation from the Duration (decimal)
		durationHour := int(durationDecimal)
		durationMinute := int((durationDecimal - float64(durationHour)) * 60)
		if durationMinute >= 30 {
			durationHour++ // Round up to the nearest hour if more than or equal to 30 minutes
		}
		durationTimeString := fmt.Sprintf("%02d:%02d", durationHour, durationMinute)

		combinedRow = append(combinedRow, durationTimeString)

		err = writer.Write(combinedRow)
		if err != nil {
			return fmt.Errorf("error writing combined row to output CSV: %w", err)
		}
	}

	writer.Flush()

	return nil
}

func generateOutputFilePath(inputFilePath string) string {
	fileName := strings.TrimSuffix(filepath.Base(inputFilePath), filepath.Ext(inputFilePath))
	outputFilePath := fileName + "_modified.csv"
	return filepath.Join(filepath.Dir(inputFilePath), outputFilePath)
}

func writeHeader(file *os.File, headerRow []string, columnIndices []int) error {
	headerRow = append(headerRow, "Duration (h short)")
	writer := csv.NewWriter(file)

	modifiedHeaderRow := make([]string, 0, len(headerRow)-len(columnIndices))
	for i, columnName := range headerRow {
		if !contains(columnIndices, i) {
			modifiedHeaderRow = append(modifiedHeaderRow, columnName)
		}
	}

	err := writer.Write(modifiedHeaderRow)
	if err != nil {
		return fmt.Errorf("error writing output CSV header: %w", err)
	}

	writer.Flush()

	return nil
}

func contains(slice []int, value int) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
