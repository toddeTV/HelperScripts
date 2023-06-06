package clockifyExportProcessor

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"errors"

	"math"
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
		return fmt.Errorf("error opening input CSV file: %w", err) //TODO use this also in other places instead of `return nil`?
	}
	defer file.Close()

	// Create the output CSV file
	newFilePath := strings.TrimSuffix(filePath, ".csv") + "_modified.csv"
	newFile, err := os.Create(newFilePath)
	if err != nil {
		return fmt.Errorf("error creating output CSV file: %w", err)
	}
	defer newFile.Close()

	// Read the input CSV file
	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading input CSV file: %w", err)
	}

	// Find the indices of the columns to drop
	columnIndices := []int{
		findColumnIndex(rows[0], "Client"),
		findColumnIndex(rows[0], "Task"),
		findColumnIndex(rows[0], "User"),
		findColumnIndex(rows[0], "Group"),
		findColumnIndex(rows[0], "Email"),
		findColumnIndex(rows[0], "Tags"),
		findColumnIndex(rows[0], "Start Time"),
		findColumnIndex(rows[0], "End Date"),
		findColumnIndex(rows[0], "End Time"),
		findColumnIndex(rows[0], "Duration (h)"),
		findColumnIndex(rows[0], "Billable Rate (EUR)"),
		findColumnIndex(rows[0], "Billable Amount (EUR)"),
	}

	// Drop the columns from each row
	for i := range rows {
		row := rows[i]
		for j := len(columnIndices) - 1; j >= 0; j-- {
			columnIndex := columnIndices[j]
			if columnIndex >= 0 && columnIndex < len(row) {
				row = append(row[:columnIndex], row[columnIndex+1:]...)
			}
		}
		rows[i] = row
	}

	// Combine rows with the same key and sum up the durations
	combinedRows := make(map[string]CombinedRow)
	for _, row := range rows[1:] {
		project := row[0]
		description := row[1]
		billable := row[2]                                   // Adjusted column index
		startDate, _ := time.Parse("02/01/2006", row[3])     // Adjusted column index
		durationDecimal, _ := strconv.ParseFloat(row[4], 64) // Adjusted column index

		key := strings.Join([]string{project, description, billable, startDate.Format("2006-01-02")}, "|")

		if existingRow, ok := combinedRows[key]; ok {
			existingRow.DurationDecimal += durationDecimal
			combinedRows[key] = existingRow
		} else {
			combinedRows[key] = CombinedRow{
				Project:         project,
				Description:     description,
				Billable:        billable,
				StartDate:       startDate,
				DurationDecimal: durationDecimal,
			}
		}
	}

	// Sort the combined rows by Start Date
	sortedKeys := make([]string, 0, len(combinedRows))
	for key := range combinedRows {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Slice(sortedKeys, func(i, j int) bool {
		startDateI := combinedRows[sortedKeys[i]].StartDate
		startDateJ := combinedRows[sortedKeys[j]].StartDate
		return startDateI.Before(startDateJ)
	})

	// Create the CSV writer for the output file
	writer := csv.NewWriter(newFile)
	defer writer.Flush()

	// Write the header row to the output CSV file
	headerRow := []string{"Project", "Description", "Billable", "Start Date", "Duration (decimal)", "Duration (h short)"}
	err = writer.Write(headerRow)
	if err != nil {
		return fmt.Errorf("error writing header row to output CSV: %w", err)
	}

	// Write the combined rows to the output CSV file
	for _, key := range sortedKeys {
		combinedRow := combinedRows[key]
		durationDecimal := combinedRow.DurationDecimal

		// Calculate the time string representation from the Duration (decimal)
		durationHour := int(math.Floor(durationDecimal))
		durationMinute := int(math.Floor((durationDecimal - float64(durationHour)) * 60))
		durationTimeString := fmt.Sprintf("%02d:%02d", durationHour, durationMinute)

		row := []string{
			combinedRow.Project,
			combinedRow.Description,
			combinedRow.Billable,
			combinedRow.StartDate.Format("2006-01-02"),
			strconv.FormatFloat(durationDecimal, 'f', 2, 64),
			durationTimeString,
		}

		err = writer.Write(row)
		if err != nil {
			return fmt.Errorf("error writing combined row to output CSV: %w", err)
		}
	}

	return nil
}

// Helper function to find the index of a column in a row
func findColumnIndex(row []string, columnName string) int {
	for i, col := range row {
		if col == columnName {
			return i
		}
	}
	return -1
}

// Struct to store combined rows
type CombinedRow struct {
	Project         string
	Description     string
	Billable        string
	StartDate       time.Time
	DurationDecimal float64
}
