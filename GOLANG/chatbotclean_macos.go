package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/sqweek/dialog"
	"github.com/xuri/excelize/v2"
)

func main() {
	start := time.Now()

	fmt.Println("PLEASE SELECT THE FILE")

	// Open the Excel file
	excelFileName, err := dialog.File().Filter("Excel files", "xlsx").Title("Select an Excel file").Load()
	if err != nil {
		fmt.Println("\nError selecting Excel file:", err)
		return
	}

	fmt.Println("\nSelected File:", excelFileName)
	fmt.Println("\nWELCOME TO ARTHNIRMITI")
	fmt.Println("\nSTARTED THE DATA CLEANING PROCESS")
	f, err := excelize.OpenFile(excelFileName)
	if err != nil {
		fmt.Println("Error opening Excel file:", err)
		return
	}

	// Get the sheet named "Data"
	sheetName := "Data"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		fmt.Println("\nError getting rows from sheet:", err)
		return
	}
	currentTime := time.Now()
	// Create the CSV file
	formattedTime := currentTime.Format("0502150106")
	csvFileName := fmt.Sprintf("/Users/padmchowdhary/Desktop/BOTDATA_%s.csv", formattedTime)

	csvFile, err := os.Create(csvFileName)
	if err != nil {
		fmt.Println("Error creating CSV file:", err)
		return
	}
	defer csvFile.Close()

	// Create a CSV writer
	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Initialize a map to store the data
	dataMap := make(map[string][]string)

	// Define chunk size (number of rows to process at a time)
	chunkSize := 10000 // Adjust this based on memory constraints

	// Process rows in chunks
	for i := 0; i < len(rows); i += chunkSize {
		end := i + chunkSize
		if end > len(rows) {
			end = len(rows)
		}
		processChunk(rows[i:end], dataMap)
	}

	writeMapToCSV(dataMap, writer)

	// Process and write map data to CSV
	//processedDict := process_details(dataMap)
	//writeProcessedDictToCSV(processedDict, csvFileName)

	// Print execution time
	duration := time.Since(start)
	fmt.Printf("Processing time: %v seconds\n", duration.Seconds())
	fmt.Printf("FILE IS STORED ON THE DESKTOP NAMED BOTDATA_%s.csv", formattedTime)

	desktopPath, err1 := getDesktopPath()
	if err1 != nil {
		fmt.Println("ERROR MOVING THE FILE TO BACKUP FOLDER")
	}
	dateFolder := filepath.Join(desktopPath, formattedTime)
	err = os.MkdirAll(dateFolder, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating date folder:", err)
		return
	}

	// Move the selected Excel file and the processed CSV file to the new folder
	moveFile := func(sourcePath, destDir string) error {
		destPath := filepath.Join(destDir, filepath.Base(sourcePath))
		return os.Rename(sourcePath, destPath)
	}

	err = moveFile(excelFileName, dateFolder)
	if err != nil {
		fmt.Println("Error moving Excel file:", err)
		return
	}

	err = moveFile(csvFileName, dateFolder)
	if err != nil {
		fmt.Println("Error moving CSV file:", err)
		return
	}

	fmt.Println("Files have been moved to:", dateFolder)

	fmt.Println("\n \nTHANK YOU , HAVE A NICE DAY!")
}

// processChunk processes a chunk of rows and updates the dataMap
func processChunk(rows [][]string, dataMap map[string][]string) {

	for _, row := range rows {

		// Get the key from column G (index 6)
		a1 := CleanPhoneNumber(row[6])
		a2 := CleanPhoneNumber(row[7])
		var key string
		if a1 == "7977845332" {
			key = a2
		} else {
			key = a1
		}

		// Get the value from column U (index 20)
		value := row[20]

		if key != "" {
			if _, exists := dataMap[key]; !exists {
				dataMap[key] = []string{value}
			} else {
				dataMap[key] = append(dataMap[key], value)
			}
		}
	}
}

// processemail_Swayam processes email data and updates the processedDict
func process_details(dataMap map[string][]string) map[string]map[string]string {
	processedDict := make(map[string]map[string]string)
	nameRegex := `(?i)^[A-Z][a-z]*(?:\.[a-z]+)*(?: (?:[a-z]+(?:\.[a-z]+)*)?){0,3}$`
	emailRegex := `(?i)^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`
	reName, errName := regexp.Compile(nameRegex)
	if errName != nil {
		fmt.Println("Error compiling name regex:", errName)
		return processedDict
	}

	reEmail, errEmail := regexp.Compile(emailRegex)
	if errEmail != nil {
		fmt.Println("Error compiling email regex:", errEmail)
		return processedDict
	}

	for key, values := range dataMap {
		if _, exists := processedDict[key]; !exists {
			processedDict[key] = make(map[string]string)
		}

		var indexofemail int // Temporary variable to store the index of email
		var indexofname int  // Temporary variable to store the index of name
		indexFound1 := false // Flag to check if the email index was found
		indexFound2 := false // Flag to check if the name index was found

		// Initialize default values
		panStatus := "NULL"
		bankStatus := "NULL"
		certiStatus := "NO"
		pan_bank_quest := "NO"

		// Check for PAN status
		for i, value := range values {
			value = strings.TrimSpace(value) // Clean up whitespace

			if strings.Contains(value, "Y_PAN_mar") || strings.Contains(value, "Y_PAN_hin") || strings.Contains(value, "Y_PAN_eng") {
				panStatus = "YES"
				pan_bank_quest = "YES"
				break
			}
			i = i + 1
		}

		if panStatus != "YES" {
			for i, value := range values {
				value = strings.TrimSpace(value)

				if strings.Contains(value, "pa_eng") || strings.Contains(value, "pa_mar") || strings.Contains(value, "pa_hin") {
					panStatus = "APPLIED"
					pan_bank_quest = "YES"
					break
				}
				i = i + 1
			}

		}

		if panStatus != "YES" && panStatus != "APPLIED" {
			for i, value := range values {
				value = strings.TrimSpace(value)

				if strings.Contains(value, "N_PAN_eng") || strings.Contains(value, "N_PAN_mar") || strings.Contains(value, "N_PAN_hin") {
					panStatus = "NO"
					pan_bank_quest = "YES"
					break
				} else if strings.Contains(value, "MEDIA_TEMPLATE - abhi_pan_bank_ask") || strings.Contains(value, "abhi_pan_bank_ask") {
					pan_bank_quest = "YES"

				}

				i = i + 1

			}
		}

		// Check for bank status
		for _, value := range values {
			value = strings.TrimSpace(value)

			if strings.Contains(value, "yihmo_mar") || strings.Contains(value, "yihmo_eng") || strings.Contains(value, "yihmo_hin") {
				bankStatus = "YES I DO"
			} else if strings.Contains(value, "bac_eng") || strings.Contains(value, "bac_mar") || strings.Contains(value, "bac_hin") || strings.Contains(value, "bac2_eng") || strings.Contains(value, "bac2_mar") || strings.Contains(value, "bac2_hin") {
				bankStatus = "CREATEMY"
			}

			if strings.Contains(value, "DOCUMENT-") || strings.Contains(value, "DOCUMENT -") || strings.Contains(value, "DOCUMENT") {
				certiStatus = "YES"
			}
		}

		// Check for NAME and EMAIL
		for j, value := range values {
			value = strings.TrimSpace(value)

			if strings.Contains(value, "(Please enter your full name)") || strings.Contains(value, "Let's Start With Your Name!â†²(Please Enter Your Full Name)") || strings.Contains(value, "Let's Start With Your Name!") || strings.Contains(value, "Let's Start With Your Name!â†²") {
				if j+1 < len(values) && reName.MatchString(values[j+1]) && isIrrelevantName(strings.TrimSpace(values[j+1])) == false {
					indexFound2 = true
					indexofname = j
					processedDict[key]["NAME"] = values[j+1]

				}
			} else if j+1 < len(values) && reName.MatchString(value) && isIrrelevantName(strings.TrimSpace(values[j+1])) == false {
				// If regex matches, use the value
				processedDict[key]["NAME"] = value

			}

			if strings.Contains(value, "Please enter your email id") {
				indexofemail = j
				indexFound1 = true
			} else if reEmail.MatchString(value) {
				// If regex matches, use the value
				processedDict[key]["Email"] = value
				processedDict[key]["VALID EMAIL"] = "YES"
				for _, value := range values {
					value = strings.TrimSpace(value)
					if strings.Contains(value, "you've unlocked exclusive access to Swayam Plus!") {
						processedDict[key]["SWAYAM ACCESS"] = "YES"

					}
				}
				//indexFound1 = true
			}

		}

		// Update PAN and BANK statuses
		processedDict[key]["PAN"] = panStatus
		processedDict[key]["BANK"] = bankStatus
		processedDict[key]["CertiStatus"] = certiStatus
		processedDict[key]["PAN_BANK_QUEST"] = pan_bank_quest

		if indexFound2 {
			newIndex1 := indexofname + 1
			if newIndex1 < len(values) {
				name := values[newIndex1]
				processedDict[key]["NAME"] = name
			} else {
				processedDict[key]["NAME"] = "No Name Found"
			}
			processedDict[key]["Index1"] = strconv.Itoa(indexofname)
		}

		if indexFound1 {
			newIndex := indexofemail + 1
			if newIndex < len(values) {
				email := values[newIndex]
				processedDict[key]["Email"] = email

				if check_valid_email(email) {
					processedDict[key]["VALID EMAIL"] = "YES"
					processedDict[key]["SWAYAM ACCESS"] = "NO"

					for _, value := range values {
						value = strings.TrimSpace(value)
						if strings.Contains(value, "you've unlocked exclusive access to Swayam Plus!") {
							processedDict[key]["SWAYAM ACCESS"] = "YES"
							break
						}
					}
				} else {
					processedDict[key]["VALID EMAIL"] = "NO"
				}
				processedDict[key]["Index"] = strconv.Itoa(indexofemail)
			} else {
				processedDict[key]["Email"] = "No email found"
			}
		}
	}

	return processedDict
}

func writeProcessedDictToCSV(processedDict map[string]map[string]string, csvFileName string) {
	// Open the existing CSV file to append the new data
	csvFile, err := os.OpenFile(csvFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening CSV file:", err)
		return
	}
	defer csvFile.Close()

	// Create a CSV writer
	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Check if the file is new or empty and write the header
	if fileInfo, err := os.Stat(csvFileName); err != nil || fileInfo.Size() == 0 {
		header := []string{"Key", "Name", "QUESTION ASKED", "PAN", "BANK", "CERTIFICATE", "Email", "VALID EMAIL", "SWAYAM ACCESS"}
		if err := writer.Write(header); err != nil {
			fmt.Println("Error writing header to CSV:", err)
			return
		}
	}

	// Write the processed data from processedDict to the CSV file
	for key, value := range processedDict {
		csvRow := []string{string(key), value["NAME"], value["PAN_BANK_QUEST"], value["PAN"], value["BANK"], value["CertiStatus"], value["Email"], value["VALID EMAIL"], value["SWAYAM ACCESS"]}
		if err := writer.Write(csvRow); err != nil {
			fmt.Println("Error writing row to CSV:", err)
		}
	}

	fmt.Println("Processed data written to the CSV file:", csvFileName)
}

func check_valid_email(email_str string) bool {
	const emailPattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	re := regexp.MustCompile(emailPattern)

	return re.MatchString(email_str)
}

func writeMapToCSV(dataMap map[string][]string, writer *csv.Writer) {
	for key, values := range dataMap {
		csvRow := append([]string{key}, values...)
		if err := writer.Write(csvRow); err != nil {
			fmt.Println("Error writing to CSV:", err)
		}
	}
}

// CleanPhoneNumber ensures the phone number is exactly 10 digits long.
func CleanPhoneNumber(phoneNumber string) string {
	phoneNumber = strings.TrimSpace(phoneNumber) // Remove any leading or trailing spaces

	for len(phoneNumber) > 10 {
		phoneNumber = phoneNumber[1:] // Remove the first character
	}

	return phoneNumber
}
func getDesktopPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var desktopDir string
	switch runtime.GOOS {
	case "darwin": // macOS
		desktopDir = filepath.Join(homeDir, "Desktop")
	case "windows":
		desktopDir = filepath.Join(homeDir, "Desktop")
	case "linux":
		desktopDir = filepath.Join(homeDir, "Desktop")
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return desktopDir, nil
}

func isIrrelevantName(name string) bool {
	irrelevantNames := []string{
		"INTERACTIVE", "hi", "Hi", "Thanks", "Ok", "ok", "Yes", "No", "certificate", "Certificate",
	}

	for _, irrelevant := range irrelevantNames {
		if strings.Contains(name, irrelevant) {
			return true
		}
	}
	return false
}
