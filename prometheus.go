package main 

import (
    "fmt"
 //   "time"
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
    Target string
    output chan <- int
    control <- chan ControlMessage
    data []Point
    mode OutputType
}

type MessageType int
const(
    StartOutput MessageType = 0
    StopOutput  MessageType = 1

)

type OutputType int
const(
    Playback    OutputType = 0
    Live        OutputType = 1
    Init        OutputType = -1
)

type QueryInfo struct {
    Query string
    Start float64
    End   float64
    Step  int
}
type ControlMessage struct {
    Type MessageType
    OutputType OutputType
    QueryInfo QueryInfo 
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

func newPrometheusScraper(queryEndpoint string, mode string, controlChannel <- chan ControlMessage, outputChannel chan <- int) *prometheusScraper {
	
	fmt.Printf("Server: %s\n", queryEndpoint)
	
	prometheusScraper := prometheusScraper {queryEndpoint,outputChannel, controlChannel, []Point{},Init}

    go prometheusScraper.controlThread()
	

	return &prometheusScraper
}

/* This function listens for any incoming messages and handles them accordingly */
func (collector *prometheusScraper) controlThread() {
    for {

        message := <-collector.control

        switch message.Type {
        
        case StartOutput:
           
            fmt.Printf("Starting output thread.. Playback Type: %i\n", message.OutputType)
            fmt.Printf("Query: %s Start: %i Stop: %i Step: %i \n", message.QueryInfo.Query, message.QueryInfo.Start, message.QueryInfo.End, message.QueryInfo.Step)
           
            collector.queryPrometheus(message.OutputType, message.QueryInfo.Query, message.QueryInfo.Start, message.QueryInfo.End, message.QueryInfo.Step)


        case StopOutput:
            fmt.Printf("Stopping output thread..\n")
        default:
            fmt.Printf("Unknown MessageType: %i \n", message.Type )
        } 
    }
}

/*  Stores the initial time series data, starts the output thread, and also the live playback query thread if required. */
func (collector *prometheusScraper) queryPrometheus(mode OutputType, query string, start float64, end float64, step int) bool {

    collector.data = collector.getTimeSeriesData(query, start, end, step)
    go collector.outputThread()

    if mode == Live {
        go collector.queryThread(query, step)
    }

    return true
}

func (collector *prometheusScraper) outputThread() {
    /* Query data structure using mutex in timed loop based on step and emit message to output channel. */
}

func (collector *prometheusScraper) queryThread(query string, step int) {
    /* Populate data structure using mutex in timed loop based on step. Need to make sure the query poll rate is a division of step. */
}
 
/* Returns an array of points which represent the timeseries data for the specified query.
   NOTE: Doesn't handle more than one set of time series (Result[0] ret), will expand to handle it later.
*/
func (collector *prometheusScraper) getTimeSeriesData(query string, start float64, end float64, step int) []Point {
	request, err := http.NewRequest("GET", collector.Target, nil)
    
    if err != nil {
        fmt.Printf("%s\n", err)
        return []Point{}
    }

    
    q := request.URL.Query()

    q.Add("query", query)
    q.Add("start", strconv.FormatFloat(start, 'f', 6, 64))
    q.Add("end",   strconv.FormatFloat(end, 'f', 6, 64))
    q.Add("step",  strconv.Itoa(step))

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
        return []Point{}
    }

    fmt.Printf("%+v\n", apiResponse.Status)
    //fmt.Printf("%+v\n", apiResponse.Data.ResultType)
    fmt.Printf("%+v\n", apiResponse.Data.Result)

    return apiResponse.Data.Result[0].Values
}
