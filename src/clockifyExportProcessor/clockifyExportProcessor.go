package clockifyExportProcessor

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"errors"

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

	// // Modify the data using loops
	// for i, row := range rows {
	// 	for j, value := range row {
	// 		// Modify the value (e.g., append a prefix)
	// 		row[j] = "Modified: " + value
	// 	}
	// 	rows[i] = row
	// }

	// // Print the modified data
	// for _, row := range rows {
	// 	fmt.Println(row)
	// }

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

	fmt.Println("1")

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

	fmt.Println("2")

	// Create a new slice to store modified rows, excluding the header row
	modifiedRows := make([][]string, len(rows)-1)

	// Drop the columns from each row
	for i, row := range rows[1:] {
		modifiedRow := make([]string, 0, len(row)-len(columnIndices))
		for j, cell := range row {
			// Only keep cells that are not in columnIndices
			if !contains(columnIndices, j) {
				modifiedRow = append(modifiedRow, cell)
			}
		}
		modifiedRows[i] = modifiedRow
	}

	fmt.Println("3")

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

	// Write the modified data to the output CSV file
	writer := csv.NewWriter(newFile)
	err = writer.WriteAll(modifiedRows)
	if err != nil {
		return fmt.Errorf("error writing output CSV: %w", err)
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
