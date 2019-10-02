package main 

import (
	"fmt"
	"strconv"
	"net/http"
	"encoding/json"
)

type Metric struct {
	Instance string
	Job string
}

type Point struct {
    Timestamp int64
    Value float64
}

type TimeSeries struct {
	Metric Metric
	Values []Point
}

type PrometheusData struct {
	ResultType string
    Result []TimeSeries
}

type APIResponse  struct {
	Status string 
	Data PrometheusData
}

type prometheusScraper struct {
	queryEndpoint string
}

 func (tp *Point) UnmarshalJSON(data []byte) error {

    var v []interface{}
    
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Printf("Error while decoding Point %v\n", err)
		return err
    }
    
	tp.Timestamp = int64(v[0].(float64))
	tp.Value, _ = strconv.ParseFloat(v[1].(string), 64)

	return nil
} 

func newPrometheusScraper(queryEndpoint string, controlChannel chan <- string, outputChannel <- chan int) *prometheusScraper {
	
	fmt.Printf("Server: %s\n", queryEndpoint)
	
	prometheusScraper := prometheusScraper {queryEndpoint:queryEndpoint }

	

	return &prometheusScraper
}

func (collector *prometheusScraper) queryPrometheus(promQuery string, start int, end int, step int) {

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

    var apiResponse APIResponse

    e := json.NewDecoder(result.Body).Decode(&apiResponse)

    if e != nil {
        fmt.Printf("%s\n", e)
    }

    fmt.Printf("%+v\n", apiResponse.Status)
    fmt.Printf("%+v\n", apiResponse.Data.ResultType)
    fmt.Printf("%+v\n", apiResponse.Data.Result)
    
}
