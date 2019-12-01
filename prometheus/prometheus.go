package prometheus

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-collections/go-datastructures/queue"
)

type metric struct {
	Instance string
	Job      string
}

type Point struct {
	Timestamp int64
	Value     float64
}

type TimeSeries struct {
	m      metric
	Values []Point `json:"values"`
}

type PrometheusData struct {
	ResultType string       `json:"type"`
	Result     []TimeSeries `json:"result"`
}

type APIResponse struct {
	Status string         `json:"status"`
	Data   PrometheusData `json:"data"`
}

/*Scraper Holds all relevant variables for scraping Promthetheus.*/
type Scraper struct {
	Target     string
	output     chan<- float64
	control    <-chan ControlMessage
	mode       OutputType
	data       *queue.RingBuffer
	pollRate   int
	outputRate int
}

const defaultRingSize = 10000

const defaultPollRate = 600
const defaulttOutputRate = 600

/*MessageType The type of Control Message being sent. */
type MessageType int

/* Message Types for Control Messages1 */
const (
	StartOutput      MessageType = 0
	StopOutput       MessageType = 1
	ChangePollRate   MessageType = 2
	ChangeOutputRate MessageType = 3
)

/*OutputType The type of output to use.*/
type OutputType int

/* Constants for Output Mode. */
const (
	Playback OutputType = 0
	Live     OutputType = 1
	Init     OutputType = -1
)

/*QueryInfo Information used to store information on query being used to scrape metric values.*/
type QueryInfo struct {
	Query string
	Start float64
	End   float64
	Step  int
}

/*ControlMessage Message used to change behaviour of Prometheus scraper.*/
type ControlMessage struct {
	Type       MessageType
	OutputType OutputType
	QueryInfo  QueryInfo
	Value      int
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

/*NewScraper Initializes a new instance of the scraper struct and starts the control thread. */
func NewScraper(queryEndpoint string, mode OutputType, controlChannel <-chan ControlMessage, outputChannel chan<- float64) *Scraper {

	scraper := Scraper{queryEndpoint, outputChannel, controlChannel, mode, queue.NewRingBuffer(defaultRingSize), defaultPollRate, defaulttOutputRate}
	go scraper.prometheusControlThread()

	return &scraper
}

/* This function listens for any incoming messages and handles them accordingly */
func (collector *Scraper) prometheusControlThread() {
	for {

		message := <-collector.control

		switch message.Type {

		case StartOutput:

			fmt.Printf("Starting output thread.. Playback Type: %d\n", message.OutputType)
			fmt.Printf("Query: %s Start: %f Stop: %f Step: %d \n", message.QueryInfo.Query, message.QueryInfo.Start, message.QueryInfo.End, message.QueryInfo.Step)

			collector.queryPrometheus(message.OutputType, message.QueryInfo.Query, message.QueryInfo.Start, message.QueryInfo.End, message.QueryInfo.Step)

		case ChangePollRate:

			fmt.Printf("Changing PollRate by (%d) \n", message.Value)
			collector.pollRate += message.Value

		case ChangeOutputRate:

			fmt.Printf("Changing OutputRate by (%d) \n", message.Value)
			collector.outputRate += message.Value

		case StopOutput:
			fmt.Printf("Stopping output thread..\n")
		default:
			fmt.Printf("Unknown MessageType: (%d \n", message.Type)
		}
	}
}

/*  Stores the initial time series data, starts the output thread, and also the live playback query thread if required. */
func (collector *Scraper) queryPrometheus(mode OutputType, query string, start float64, end float64, step int) bool {

	data := collector.getTimeSeriesData(query, start, end, step)
	collector.populateRingBuffer(data)

	if mode == Live {
		fmt.Println("In live mode")
		go collector.queryThread(query, step)
	}

	go collector.outputThread()

	return true
}

/* Gets the next item from the RingBuffer and emits it on the output channel. Then sleeps for a configurable duration. */
func (collector *Scraper) outputThread() {
	for {

		item, err := collector.data.Get()

		if err != nil {
			fmt.Printf("Error: %s", err)
		}

		collector.output <- item.(float64)

		time.Sleep(time.Duration(collector.outputRate) * time.Millisecond)
	}
}

/* Queries for latest TimeSeries data, and sleeps for configurable duration. */
func (collector *Scraper) queryThread(query string, step int) {
	for {

		now := float64(time.Now().Unix())
		//	fmt.Printf("Polling for data..\n")

		data := collector.getTimeSeriesData(query, now, now, step)
		collector.populateRingBuffer(data)

		time.Sleep(time.Duration(collector.pollRate) * time.Millisecond)

	}
}

func (collector *Scraper) populateRingBuffer(data []Point) {
	for _, point := range data {
		//fmt.Printf("PromValue: %f\n", point.Value)
		collector.data.Put(point.Value)
	}
}

/* Returns an array of points which represent the timeseries data for the specified query.
   NOTE: Doesn't handle more than one set of time series (Result[0]), Will expand to handle it later.
*/
func (collector *Scraper) getTimeSeriesData(query string, start float64, end float64, step int) []Point {
	request, err := http.NewRequest("GET", collector.Target, nil)

	if err != nil {
		fmt.Printf("%s\n", err)
		return []Point{}
	}

	q := request.URL.Query()

	q.Add("query", query)
	q.Add("start", strconv.FormatFloat(start, 'f', 6, 64))
	q.Add("end", strconv.FormatFloat(end, 'f', 6, 64))
	q.Add("step", strconv.Itoa(step))

	request.URL.RawQuery = q.Encode()

	//fmt.Printf("URL      %+v\n", request.URL)
	//fmt.Printf("RawQuery %+v\n", request.URL.RawQuery)
	//fmt.Printf("Query    %+v\n", request.URL.Query())

	result, err := http.DefaultClient.Do(request)

	if err != nil {
		fmt.Printf("Error1: %s\n", err)
		return []Point{}
	}

	defer result.Body.Close()

	var apiResponse APIResponse

	e := json.NewDecoder(result.Body).Decode(&apiResponse)

	if e != nil {
		fmt.Printf("Error2: %s\n", e)
		return []Point{}
	}
	/* Need to check that return value is valid before returning. */
	return apiResponse.Data.Result[0].Values
}
