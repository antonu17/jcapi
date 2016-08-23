package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/TheJumpCloud/jcapi"
)

// URLBase is the production api endpoint.
const URLBase string = "https://console.jumpcloud.com/api"

func main() {
	var apiKey string
	var commandID string
	var url string
	var outfile string

	flag.StringVar(&apiKey, "key", "", "Your JumpCloud Administrator API Key")
	flag.StringVar(&commandID, "commandid", "", "The id of the command to run")
	flag.StringVar(&outfile, "out", "", "File path for CSV output")
	flag.StringVar(&url, "url", URLBase, "Alternative Jumpcloud API URL (optional)")
	flag.Parse()

	if apiKey == "" {
		log.Fatalln("API key must be provided.")
	}

	if commandID == "" {
		log.Fatalln("Command id must be provided")
	}

	api := jcapi.NewJCAPI(apiKey, url)
	results, err := api.GetCommandResultsBySavedCommandID(commandID)
	if err != nil {
		log.Fatalln(err)
	}

	var output *os.File
	if outfile != "" {
		path, err := filepath.Abs(outfile)
		if err != nil {
			log.Fatalln("Entered an incorrect file path for CSV output")
		}
		output, err = getFileWriter(path)
		if err != nil {
			log.Fatalln("Problem with the outfile: %s", err.Error)
		}
	} else {
		output = os.Stdout
	}

	defer output.Close()

	if err := writeResultsToCSV(results, output); err != nil {
		log.Fatalln("Error writing to csv: %s", err.Error())
	}
}

func getFileWriter(absPath string) (*os.File, error) {
	if _, err := os.Stat(absPath); !os.IsNotExist(err) {
		return nil, fmt.Errorf("Output already exists")
	}

	return os.Create(absPath)
}

func writeResultsToCSV(results []jcapi.JCCommandResult, writer io.Writer) error {
	w := csv.NewWriter(writer)

	if err := w.Write([]string{"SYSTEM ID", "USERNAME", "JUMPCLOUD USERNAME", "COMMAND REQUEST TIME"}); err != nil {
		return err
	}
	w.Flush()

	for _, result := range results {
		lines := strings.Split(result.Response.Data.Output, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				if err := w.Write([]string{result.System, line, "", result.RequestTime}); err != nil {
					return err
				}

			}
		}
		w.Flush()
	}
	return nil
}
