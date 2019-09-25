package main 

import (
	"fmt"
	"strconv"
	//"io/ioutil"
	"net/http"
	"encoding/json"
)

type Metric struct {
	instance string
	job string
}

type Value struct {
	data []string
}

type TimeSeries struct {
	metric Metric
	values []Value
}

type PrometheusData struct {
	resultType string
    results []TimeSeries
}

type APIResponse  struct {
	status string
	data PrometheusData
}

type prometheusScraper struct {
	queryEndpoint string
}



func newPrometheusScraper(queryEndpoint string, controlChannel chan <- string, outputChannel <- chan int) *prometheusScraper {
	
	fmt.Printf("Server: %s\n", queryEndpoint)
	
	prometheusScraper := prometheusScraper {queryEndpoint:queryEndpoint }

	

	return &prometheusScraper
}

func (collector *prometheusScraper) queryPrometheus(promQuery string, start int, end int, step int) {
	//stddev_over_time(pf_current_entries_total%7Binstance%3D~"sovapn%5B1%7C2%5D%3A9116"%7D%5B12h%5D)%20&start=1568722200&end=1569327600&step=600
	

	//promData := new(PrometheusData)

	request, err := http.NewRequest("GET", collector.queryEndpoint, nil)
    
    if err != nil {
        fmt.Printf("%s\n", err)
    }

    q := request.URL.Query()
    q.Add("query", promQuery)
    q.Add("start", strconv.Itoa(start))
    q.Add("end", strconv.Itoa(end))
    q.Add("step", strconv.Itoa(step))

    request.URL.RawQuery = q.Encode()

    fmt.Printf("URL      %+v\n", request.URL)
    fmt.Printf("RawQuery %+v\n", request.URL.RawQuery)
    fmt.Printf("Query    %+v\n", request.URL.Query())

    result,err := http.DefaultClient.Do(request)
    defer result.Body.Close()

  //  body, _ := ioutil.ReadAll(result.Body)

  //  fmt.Println("%s\n %v\n", string(body), result.Body)

    var apiResponse APIResponse

    e := json.NewDecoder(result.Body).Decode(&apiResponse)

    if e != nil {
        fmt.Printf("%s\n", e)
    }
    fmt.Println("%v\n", apiResponse)

}

func (collector *prometheusScraper) getJson(url string, target interface{}) error {
    r, err := http.Get(url)
    if err != nil {
        return err
    }
    defer r.Body.Close()
    fmt.Printf("%s", r.Body)
    return json.NewDecoder(r.Body).Decode(target)
}


//http://192.168.150.187:9090/api/v1/query_range?query=rate(pf_current_entries_total%7Binstance%3D~%22sovapn%5B1%7C2%5D%3A9116%22%7D%5B2m%5D)&start=1568833800&end=1569439200&step=600