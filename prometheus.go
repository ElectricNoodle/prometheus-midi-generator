package main 

import (
	"fmt"
	"net/url"
)

type prometheusScraper struct {
	serverAddress string
}

func newPrometheusScraper(serverAddress string, controlChannel chan <- string, outputChannel <- chan int) *prometheusScraper {
	
	fmt.Printf("Server: %s\n", serverAddress)
	
	prometheusScraper := prometheusScraper {serverAddress}

	return &prometheusScraper
}

func (collector *prometheusScraper) queryProme

func (collector *prometheusScraper) Test() {

	fmt.Printf("TEST\n")
}


//url:"api/datasources/proxy/1/api/v1/query_range?query=stddev_over_time(pf_current_entries_total%7Binstance%3D~%22sovapn%5B1%7C2%5D%3A9116%22%7D%5B12h%5D)%20&start=1568722200&end=1569327600&step=600
//192.168.150.187:9090/api/datasources/proxy/1/api/v1/query_range?query=stddev_over_time(pf_current_entries_total%7Binstance%3D~%22sovapn%5B1%7C2%5D%3A9116%22%7D%5B12h%5D)%20&start=1568722200&end=1569327600&step=600
