package prometheus

import (
	"encoding/json"
	"logging"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-collections/go-datastructures/queue"
)

var log *logging.Logger

type metric struct {
	Instance string
	Job      string
}

type point struct {
	Timestamp int64
	Value     float64
}

type timeSeries struct {
	m      metric
	Values []point `json:"values"`
}

type prometheusData struct {
	ResultType string       `json:"type"`
	Result     []timeSeries `json:"result"`
}

type apiResponse struct {
	Status string         `json:"status"`
	Data   prometheusData `json:"data"`
}

var httpClient = &http.Client{
	Timeout: time.Second * 3,
}

/*Scraper Holds all relevant variables for scraping Promthetheus.*/
type Scraper struct {
	Target     string
	Output     chan float64
	Control    chan ControlMessage
	mode       OutputType
	data       *queue.RingBuffer
	pollRate   int
	outputRate int
	isActive   bool
}

const defaultRingSize = 10000

const defaultPollRate = 600
const defaulttOutputRate = 600

/*MessageType The type of Control Message being sent. */
type MessageType int

/* Message Types for Control Messages */
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

/*UnmarshalJSON Parses timestamp/value from byte array */
func (tp *point) UnmarshalJSON(data []byte) error {

	var v []interface{}

	if err := json.Unmarshal(data, &v); err != nil {
		log.Printf("Error while decoding Point %v\n", err)
		return err
	}

	tp.Timestamp = int64(v[0].(float64))
	tp.Value, _ = strconv.ParseFloat(v[1].(string), 64)

	return nil
}

/*NewScraper Initializes a new instance of the scraper struct and starts the control thread. */
func NewScraper(logIn *logging.Logger, server string, mode OutputType) *Scraper {

	log = logIn
	queryEndpoint := "http://" + server + "/api/v1/query_range"
	scraper := Scraper{queryEndpoint, make(chan float64, 600), make(chan ControlMessage, 6), mode, queue.NewRingBuffer(defaultRingSize), defaultPollRate, defaulttOutputRate, true}

	go scraper.prometheusControlThread()

	return &scraper
}

/* This function listens for any incoming messages and handles them accordingly */
func (collector *Scraper) prometheusControlThread() {
	for {

		message := <-collector.Control

		switch message.Type {

		case StartOutput:

			log.Printf("Starting output thread.. Playback Type: %d\n", message.OutputType)
			log.Printf("Query: %s Start: %f Stop: %f Step: %d \n", message.QueryInfo.Query, message.QueryInfo.Start, message.QueryInfo.End, message.QueryInfo.Step)

			collector.isActive = true
			collector.queryPrometheus(message.OutputType, message.QueryInfo.Query, message.QueryInfo.Start, message.QueryInfo.End, message.QueryInfo.Step)

		case ChangePollRate:

			log.Printf("Changing PollRate to (%d) \n", message.Value)
			collector.pollRate = message.Value

		case ChangeOutputRate:

			log.Printf("Changing OutputRate to (%d) \n", message.Value)
			collector.outputRate = message.Value

		case StopOutput:
			log.Printf("Stopping polling/output of new data.\n")
			collector.isActive = false

		default:
			log.Printf("Unknown MessageType: (%d \n", message.Type)
		}
	}
}

/*  Stores the initial time series data, starts the output thread, and also the live playback query thread if required. */
func (collector *Scraper) queryPrometheus(mode OutputType, query string, start float64, end float64, step int) {

	data := collector.getTimeSeriesData(query, start, end, step)
	collector.populateRingBuffer(data)

	if mode == Live {
		log.Println("Running in live mode")
		go collector.queryThread(query, step)
	}

	go collector.outputThread()

}

/* Gets the next item from the RingBuffer and emits it on the output channel. Then sleeps for a configurable duration.
   Also disposes of current RingBuffer and assigns a new one on exit. This stops the Ringbuffer filling up and freezing
   the thread on multiple restarts. */
func (collector *Scraper) outputThread() {
	for {
		if collector.isActive {
			item, err := collector.data.Get()

			if err != nil {
				log.Printf("Error: %s", err)
			}

			collector.Output <- item.(float64)
			time.Sleep(time.Duration(collector.outputRate) * time.Millisecond)

		} else {

			collector.data.Dispose()
			collector.data = queue.NewRingBuffer(defaultRingSize)

			return
		}
	}
}

/* Queries for latest TimeSeries data, and sleeps for configurable duration. */
func (collector *Scraper) queryThread(query string, step int) {
	for {
		if collector.isActive {
			now := float64(time.Now().Unix())

			data := collector.getTimeSeriesData(query, now, now, step)
			collector.populateRingBuffer(data)

			time.Sleep(time.Duration(collector.pollRate) * time.Millisecond)
		} else {
			log.Println("Exiting query thread.")
			return
		}
	}
}

func (collector *Scraper) populateRingBuffer(data []point) {
	for _, point := range data {
		//log.Printf("PromValue: %f\n", point.Value)
		collector.data.Put(point.Value)
	}
}

/* Returns an array of points which represent the timeseries data for the specified query.
   NOTE: Doesn't handle more than one set of time series (Result[0]), Will expand to handle it later.
*/
func (collector *Scraper) getTimeSeriesData(query string, start float64, end float64, step int) []point {

	request, err := http.NewRequest("GET", collector.Target, nil)

	if err != nil {
		log.Printf("%s\n", err)
		return []point{}
	}

	q := request.URL.Query()

	q.Add("query", query)
	q.Add("start", strconv.FormatFloat(start, 'f', 6, 64))
	q.Add("end", strconv.FormatFloat(end, 'f', 6, 64))
	q.Add("step", strconv.Itoa(step))

	request.URL.RawQuery = q.Encode()

	result, err := httpClient.Do(request)

	if err != nil {
		log.Printf("Error: %s\n", err)
		return []point{}
	}

	defer result.Body.Close()

	var apiResponse apiResponse

	e := json.NewDecoder(result.Body).Decode(&apiResponse)

	if e != nil {
		log.Printf("Error: %s\n", e)
		return []point{}
	}
	/* Need to check that return value is valid before returning. */
	return apiResponse.Data.Result[0].Values
}
