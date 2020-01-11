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
	//rate(node_network_transmit_bytes_total{instance=~\"nos-analytics:9100\",device=\"ens18\"}[10m])
	//pf_current_entries_total{instance=~\"sovapn1:9116\"}
	//max(pf_states{instance=~'sovapn[1|2]:9100', protocol=~'tcp', state=~'ESTABLISHED:ESTABLISHED', type='fwstates', operator='jerseyt'})  + max(pf_states{instance=~'sovapn[1|2]:9100', protocol=~'tcp', state=~'ESTABLISHED:ESTABLISHED', type='nat', operator='jerseyt'})
	queryInfo := prometheus.QueryInfo{Query: "max(pf_states{instance=~'sovapn[1|2]:9100', protocol=~'tcp', state=~'ESTABLISHED:ESTABLISHED', type='fwstates', operator='jerseyt'})  + max(pf_states{instance=~'sovapn[1|2]:9100', protocol=~'tcp', state=~'ESTABLISHED:ESTABLISHED', type='nat', operator='jerseyt'})", Start: 1576281600, End: 1576886400, Step: 600}

	messageStart := prometheus.ControlMessage{Type: prometheus.StartOutput, OutputType: prometheus.Live, QueryInfo: queryInfo, Value: 0}
	//messageStop := prometheus.ControlMessage{prometheus.StopOutput, 0, prometheus.QueryInfo{}, 0}

	prometheusControlChannel <- messageStart
	//prometheusControlChannel <- messageStop

	initializeGUI()

}
func initializeBackend() {

	prometheusControlChannel = make(chan prometheus.ControlMessage, 6)
	prometheusOutputChannel = make(chan float64, 600)

	processorControlChannel = make(chan processor.ControlMessage, 6)
	processorOutputChannel = make(chan midioutput.MidiMessage, 6)

	midiControlChannel = make(chan midioutput.MidiControlMessage, 6)

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

	platform, err := platforms.NewGLFW(io, platforms.GLFWClientAPIOpenGL2)

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}

	defer platform.Dispose()

	renderer, err := renderers.NewOpenGL2(io)

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}

	defer renderer.Dispose()

	gui.Run(platform, renderer)
}
