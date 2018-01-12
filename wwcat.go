/* WarpWallet brute forcer - https://keybase.io/warp
   August 2017 - Nick Earwood - nearwood.net
   
   TODO: Discard passwords with 3 or more consecutive characters
   
*/

package main

import "database/sql"
import _ "github.com/go-sql-driver/mysql"

import "fmt"
import "os"
import "math/rand"
import "time"
import "warpwallet/warpwallet"
import "os/signal"
import "syscall"
import "runtime"
import "flag"
import "strings"

type kpTuple struct {
	pw string
	pub string
	priv string
}

func main() {
	quit := false

	signals := make(chan os.Signal, 2)
	defer close(signals)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	
  go func () {
		sig := <-signals
		fmt.Printf("Interrupt received...\n")
        switch sig {
        case os.Interrupt:
            quit = true
        case syscall.SIGTERM:
            quit = true
		}
	}()

	rand.Seed(time.Now().UnixNano())

	pNumCPU := flag.Int("parallel", runtime.NumCPU(), "an int")
	flag.Parse()
	fmt.Printf("Using %d logical CPUs\n", *pNumCPU)	
	
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting hostname: %v\n", err)
		hostname = "unknown"
	}
	
	dbName := "warpwallet"
	dbUser := ""
	dbPass := ""
	dbHost := ""
	fmt.Printf("Connecting to DB... ")
	db, err := sql.Open("mysql", dbUser + ":" + dbPass + "@tcp(" + dbHost + ")/" + dbName + "?tls=false")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to DB: %v\n", err)
	} else {
		fmt.Printf("Connected.\n")
	}
	
	defer db.Close()
	
	//Test query to actually try to connect
	fmt.Printf("Testing DB connection... ")
	err = db.Ping()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running test query on db: %v\n", err)
		return
	} else {
		fmt.Printf("OK\n")
	}

	fmt.Printf("Initializing... ")
	generateCount := uint32(0)
	currentCount := uint32(0)
  maxRate := float32(0)

	passwords := make(chan string, *pNumCPU)
	defer close(passwords)
	for i := 0; i < *pNumCPU; i++ {
		go getPassword(passwords)
	}
	
	jobs := make(chan kpTuple, *pNumCPU)
	defer close(jobs)
	storeJobs := make(chan bool, *pNumCPU)
	defer close(storeJobs)

	fmt.Printf("OK\n")
	timeStart := time.Now()

	for !quit {
		select {
			case password := <-passwords:
				//fmt.Printf("Password generated: %s\n", password)
				go getKeypair(password, jobs)
				//go getPassword(passwords)

			case result := <-jobs:
				//fmt.Printf("Job complete: %s\n", result.pw)
				
				generateCount++
				currentCount++
        timeTaken := float32(time.Now().Sub(timeStart) / time.Millisecond) / 1000
				if timeTaken >= 5 {
					rate := float32(currentCount) / timeTaken
          if (rate > maxRate) { maxRate = rate }
					fmt.Printf("%d keypairs, %1.1f seconds, %3.2f/%3.2f kp/s - %d keypairs total\n", currentCount, timeTaken, rate, maxRate, generateCount)
          currentCount = 0
          timeStart = time.Now()
				}

				go getPassword(passwords)
				go storeResult(db, result, hostname, storeJobs)

			//case <-storeJobs:
		}
	}

	fmt.Printf("Cleaning up...\n")
}

func storeResult(db *sql.DB, result kpTuple, hostname string, c chan bool) {
	_, err := db.Exec("INSERT INTO history (text, privateKey, publicAddress, hostname) values (?, ?, ?, ?)", result.pw, result.priv, result.pub, hostname)
	if (err != nil) {
		fmt.Fprintf(os.Stderr, "Error saving history: %v\n", err)
		c <- false
	} else {
		c <- true
	}
}

func getPassword(c chan string) {
	c <- generatePassword(8)
}

func getKeypair(pw string, c chan kpTuple) {
	const salt = "a@b.c"
	//fmt.Printf("Generating key pair for pw: %s...\n", pw)
	priv, pub := warpwallet.Generate(pw, salt)
	c <- kpTuple{pw, pub, priv}
}

const charPool = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generatePassword(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charPool[rand.Intn(len(charPool))]
	}

	//validate to ensure password meets 'at least one of each' assumption
	pw := string(b)
	if (strings.ContainsAny(pw, "abcdefghijklmnopqrstuvwxyz") &&
		strings.ContainsAny(pw, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") &&
		strings.ContainsAny(pw, "0123456789")) {
		return pw
	} else {
		//fmt.Printf("Discarding password: %s\n", pw)
		return generatePassword(n)
	}
}
