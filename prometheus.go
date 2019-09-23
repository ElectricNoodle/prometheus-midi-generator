package main 

import (
	"fmt"
)

type prometheusScraper struct {
	serverAddress string
}

func newPrometheusScraper(serverAddress string, controlChannel chan <- string, outputChannel <- chan int) *prometheusScraper {
	
	fmt.Printf("Server: %s\n", serverAddress)
	
	prometheusScraper := prometheusScraper {serverAddress}

	return &prometheusScraper
}

func (collector *prometheusScraper) Test() {

	fmt.Printf("TEST\n")
}