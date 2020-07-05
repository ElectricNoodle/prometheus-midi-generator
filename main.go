package main

import (
	"fmt"
	"gui"
	"gui/platforms"
	"gui/renderers"
	"io/ioutil"
	"logging"
	"midioutput"
	"os"
	"processor"
	"prometheus"

	"github.com/inkyblackness/imgui-go"
	"gopkg.in/yaml.v2"
)

type config struct {
	PrometheusServer string           `yaml:"prometheus_server"`
	ProcessorConfig  processor.Config `yaml:"processor_config"`
}

var configuration *config

var log *logging.Logger
var prometheusScraper *prometheus.Scraper
var metricProcessor *processor.ProcInfo
var midiEmitter *midioutput.MIDIEmitter

var prometheusControlChannel chan prometheus.ControlMessage
var prometheusOutputChannel chan float64

var processorControlChannel chan processor.ControlMessage
var processorOutputChannel chan midioutput.MIDIMessage

var midiControlChannel chan midioutput.MIDIControlMessage

var guiLogChannel chan string

func main() {

	configuration = loadConfig("config/config.yml")

	initializeBackend()
	initializeGUI()

}

func loadConfig(path string) *config {

	var c *config

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	//fmt.Printf("%v\n", yamlFile)
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	if c.PrometheusServer == "" {
		log.Fatal("Configuration file invalid: No Prometheus server is defined.\n")
	}

	if len(c.ProcessorConfig.Scales) < 1 {
		log.Fatal("Processor configuration doesn't contain any scale definitions.\n")
	}

	for _, scale := range c.ProcessorConfig.Scales {
		if scale.Name == "" {
			log.Fatal("Configuration file invalid: Scale defined without name.\n")
		}
		if scale.Intervals == nil || len(scale.Intervals) < 1 {
			log.Fatalf("Configuration file invalid: %s scale defined without any intervals.\n", scale.Name)
		}
	}

	return c
}

func initializeBackend() {

	prometheusControlChannel = make(chan prometheus.ControlMessage, 6)
	prometheusOutputChannel = make(chan float64, 600)

	processorControlChannel = make(chan processor.ControlMessage, 6)
	processorOutputChannel = make(chan midioutput.MIDIMessage, 6)

	midiControlChannel = make(chan midioutput.MIDIControlMessage, 6)

	log = logging.NewLogger()

	prometheusScraper = prometheus.NewScraper(log, configuration.PrometheusServer, prometheus.Playback, prometheusControlChannel, prometheusOutputChannel)
	metricProcessor = processor.NewProcessor(log, configuration.ProcessorConfig, processorControlChannel, prometheusOutputChannel, processorOutputChannel)
	midiEmitter = midioutput.NewMidi(log, midiControlChannel, processorOutputChannel)

	fmt.Printf("%s\n", prometheusScraper.Target)
	fmt.Printf("%f\n", metricProcessor.BPM)
	fmt.Printf("%v+\n", midiEmitter)

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

	gui.Run(platform, renderer, log, prometheusScraper, metricProcessor, midiEmitter)
}
