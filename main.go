package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/HasinduLanka/console"
)

type Hashes struct {
	Hashes   map[string]string
	Excludes map[string]struct{}
}

type Validation struct {
	AllValid bool
	Invalid  map[string]string
	Valid    map[string]string
}

var Excludes map[string]struct{}

func main() {

	console.GlobalReader = &console.Reader{}
	console.GlobalWriter = &console.Writer{}

	mode := 'v'
	filename := "hashes.json"
	dir := "."
	validationFile := "validation.json"

	const usage = `
	Usage :
		Create hash file:
			hasher -h <filename> <directory>
	
		Validate hash file:
			hasher -v <filename> <directory> <validationFile>

`

	if len(os.Args) == 1 {
		console.Print(usage)
		console.Print("No arguments provided. Defaulting to 'hasher -v hashes.json . validation.json'")

	} else if len(os.Args) == 2 {

		console.Print(usage)
		console.Print("One argument provided. Defaulting to 'hasher " + os.Args[1] + " hashes.json . validation.json'")

		// check if the switch is valid
		if os.Args[1] == "-h" {
			mode = 'h'
		} else if os.Args[1] == "-v" {
			mode = 'v'
		} else {
			console.Print(usage)
		}

	} else if len(os.Args) > 3 {

		// check if the switch is valid
		if os.Args[1] == "-h" {
			mode = 'h'
		} else if os.Args[1] == "-v" {
			mode = 'v'
		} else {
			console.Print(usage)
		}

		filename = os.Args[2]
		dir = os.Args[3]

		if len(os.Args) > 4 {
			validationFile = os.Args[4]

		}

	} else {
		console.Print(usage)
		return
	}

	if mode == 'v' {
		// check if the file exists
		if _, err := os.Stat(filename); err != nil {
			// if the file doesn't exist, print an error
			log.Fatal(err)
		}
	}

	// check if the directory exists
	if _, err := os.Stat(dir); err != nil {
		// if the directory doesn't exist, print an error
		log.Fatal("Directory does not exist ", err)
	}

	// Exclude these files
	Excludes = make(map[string]struct{})

	for _, exclude := range []string{
		filename,
		validationFile,
	} {
		Excludes[exclude] = struct{}{}
	}

	if mode == 'h' {
		CreateHashes(filename, dir)
	} else if mode == 'v' {
		CreateValidation(filename, dir, validationFile)
	}
}

func CreateHashes(filename string, dir string) *Hashes {
	console.Print("\nCreating hashes...")

	// List all files in the directory
	files := make(chan string)
	go ListFilesRecursive(dir, files)

	// Create a map to store the hashes
	hashes := make(map[string]string, len(files))

	// Iterate through the files
	for file := range files {

		filePath := file

		// Skip the excluded files
		if _, ok := Excludes[filePath]; ok {
			continue
		}

		// Get the hash of the file
		hash := HashFile(filePath)

		// Add the hash to the map
		hashes[filePath] = hash

		console.Print(hash + " : " + filePath)
	}

	// Create a new Hashes struct
	h := &Hashes{
		Hashes:   hashes,
		Excludes: Excludes,
	}

	// Write the hashes to the file
	console.Print("\nWriting hashes to file...")
	console.NewWriterToFile(filename).Log(h)

	return h
}

func HashFile(filename string) string {

	hash := sha1.New()
	fileReader := console.NewReaderFromFile(filename)

	if _, err := io.Copy(hash, fileReader.ScannerFile); err != nil {
		log.Fatal("File read error : ", filename, err)
	}

	sum := hash.Sum(nil)
	return fmt.Sprintf("%x", sum)

}

func ValidateHashes(h *Hashes, dir string) *Validation {

	var validation = &Validation{
		AllValid: true,
		Invalid:  make(map[string]string),
		Valid:    make(map[string]string),
	}

	console.Print("\nValidating hashes...")

	// List all files in the directory
	files := make(chan string)
	go ListFilesRecursive(dir, files)

	// Iterate through the files
	for file := range files {

		filePath := file

		// Skip the excluded files
		if _, ok := h.Excludes[filePath]; ok {
			continue
		}

		// Get the hash of the file
		hash := HashFile(filePath)

		// Check if the hash matches the stored hash
		if hash != h.Hashes[filePath] {
			validation.AllValid = false
			validation.Invalid[filePath] = hash

			console.Print("\nHash mismatch for " + filePath + ":")
			console.Print("\t" + hash + " : " + h.Hashes[filePath])
		} else {
			validation.Valid[filePath] = hash

			// console.Print("\nHash match for " + filePath + ":")
			// console.Print("\t" + hash + " : " + h.Hashes[filePath])
		}
	}

	if validation.AllValid {
		console.Print("\nAll hashes match!")
	} else {
		console.Print("\nSome hashes do not match!")
	}

	return validation
}

func ValidateHashesFromFile(filename string, dir string) *Validation {

	// Read the hashes from the file
	fileBytes, fileErr := os.ReadFile(filename)

	if fileErr != nil {
		log.Fatal(fileErr)
	}

	var h Hashes
	JErr := json.Unmarshal(fileBytes, &h)

	if JErr != nil {
		log.Fatal(JErr)
	}

	// Validate the hashes
	return ValidateHashes(&h, dir)
}

func CreateValidation(filename string, dir string, validationFile string) {

	// Create the validation file
	validation := ValidateHashesFromFile(filename, dir)

	// Write the validation file
	console.Print("\nWriting validation to file...")
	console.NewWriterToFile(validationFile).Log(validation)
}

func ListFilesRecursive(dir string, ch chan string) {
	filepath.Walk(dir, func(fpath string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			ch <- filepath.ToSlash(path.Clean(fpath))
		}
		return nil
	})
	close(ch)
}
