package main

import (
	"fmt"
	"os"
	"prometheus-midi-generator/gui"
	"prometheus-midi-generator/gui/platforms"
	"prometheus-midi-generator/gui/renderers"
	"prometheus-midi-generator/midioutput"
	"prometheus-midi-generator/processor"
	"prometheus-midi-generator/prometheus"

	"github.com/inkyblackness/imgui-go"
)

var prometheusScraper *prometheus.Scraper
var metricProcessor *processor.ProcInfo
var midiOutput *midioutput.MidiInfo

var prometheusControlChannel chan prometheus.ControlMessage
var prometheusOutputChannel chan float64

var processorControlChannel chan processor.ControlMessage
var processorOutputChannel chan midioutput.MidiMessage

var midiControlChannel chan midioutput.MidiControlMessage

func main() {

	initializeBackend()

	/* Test messages to set Query Info and Start playback. */
	//queryInfo := prometheus.QueryInfo{"stddev_over_time(pf_current_entries_total{instance=~\"sovapn1:9116\"}[12h])", 1573075902, 1573075902, 600}
	queryInfo := prometheus.QueryInfo{Query: "pf_current_entries_total{instance=~\"sovapn1:9116\"}", Start: 1573035902, End: 1573075902, Step: 600}

	messageStart := prometheus.ControlMessage{Type: prometheus.StartOutput, OutputType: prometheus.Live, QueryInfo: queryInfo, Value: 0}
	//messageStop := prometheus.ControlMessage{prometheus.StopOutput, 0, prometheus.QueryInfo{}, 0}

	prometheusControlChannel <- messageStart
	//prometheusControlChannel <- messageStop

	initializeGUI()

}
func initializeBackend() {

	prometheusControlChannel = make(chan prometheus.ControlMessage, 3)
	prometheusOutputChannel = make(chan float64, 3)

	processorControlChannel = make(chan processor.ControlMessage, 3)
	processorOutputChannel = make(chan midioutput.MidiMessage, 3)

	midiControlChannel = make(chan midioutput.MidiControlMessage, 3)

	prometheusScraper = prometheus.NewScraper("http://192.168.150.187:9090/api/v1/query_range", prometheus.Playback, prometheusControlChannel, prometheusOutputChannel)
	metricProcessor = processor.NewProcessor(processorControlChannel, prometheusOutputChannel, processorOutputChannel)
	midiOutput = midioutput.NewMidi(midiControlChannel, processorOutputChannel)

	fmt.Printf("%s\n", prometheusScraper.Target)
	fmt.Printf("%f\n", metricProcessor.BPM)
	fmt.Printf("%v+\n", midiOutput)

}

func initializeGUI() {
	context := imgui.CreateContext(nil)
	defer context.Destroy()
	io := imgui.CurrentIO()

	platform, err := platforms.NewGLFW(io, platforms.GLFWClientAPIOpenGL3)

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}

	defer platform.Dispose()

	renderer, err := renderers.NewOpenGL3(io)

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}

	defer renderer.Dispose()

	gui.Run(platform, renderer)
}
