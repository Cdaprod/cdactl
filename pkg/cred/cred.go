package cred

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func HandleCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: cdactl cred <store|retrieve> [service] [username]")
		return
	}

	switch args[0] {
	case "store":
		if len(args) < 3 {
			fmt.Println("Usage: cdactl cred store <service> <username>")
			return
		}
		storeCredentials(args[1], args[2])
	case "retrieve":
		retrieveCredentials()
	default:
		fmt.Println("Invalid cred command. Use: store or retrieve")
	}
}

func storeCredentials(service, username string) {
	fmt.Print("Enter password: ")
	password, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fmt.Println("Error reading password:", err)
		return
	}
	password = strings.TrimSpace(password)

	credFile := os.Getenv("HOME") + "/.credentials"
	file, err := os.OpenFile(credFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println("Error opening credentials file:", err)
		return
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "service=%s\nusername=%s\npassword=%s\n\n", service, username, password)
	if err != nil {
		fmt.Println("Error storing credentials:", err)
	} else {
		fmt.Printf("Credentials for %s stored successfully.\n", service)
	}
}

func retrieveCredentials() {
	credFile := os.Getenv("HOME") + "/.credentials"
	file, err := os.Open(credFile)
	if err != nil {
		fmt.Println("Error opening credentials file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading credentials:", err)
	}
}

// Handler returns a placeholder message for credential operations.
//
// Example:
//
//	msg, err := cred.Handler()
//	if err != nil {
//	        fmt.Println(err)
//	}
//	fmt.Println(msg)
func Handler() (string, error) {
	return "cred command not supported in TUI", nil
}
