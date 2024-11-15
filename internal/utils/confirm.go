package utils

import (
	"fmt"
	"log"
	"slices"
	"stamus-ctl/internal/logging"
)

func AskForConfirmation(question string) bool {
	//Ask
	var response string
	fmt.Print(question)
	_, err := fmt.Scan(&response)
	if err != nil {
		log.Fatal(err)
	}
	//Check
	okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	nokayResponses := []string{"n", "N", "no", "No", "NO"}
	if slices.Contains(okayResponses, response) {
		return true
	} else if slices.Contains(nokayResponses, response) {
		logging.Sugar.Info("You did not confirm. Exiting.")
		return false
	} else {
		fmt.Print("Please type yes or no :")
		return AskForConfirmation("")
	}
}
