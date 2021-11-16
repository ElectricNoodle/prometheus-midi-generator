2package main

import (
	"fmt"
	"fractals"
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

var log *logging.Logger

var configuration *config
var scraper *prometheus.Scraper
var metricProcessor *processor.ProcInfo
var midiEmitter *midioutput.MIDIEmitter
var fractalRenderer *fractals.FractalRenderer

func main() {

	configuration = loadConfig("config/config.yml")

	initializeBackend()
	initializeGUI()

}

func loadConfig(path string) *config {

	var conf *config

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	if conf.PrometheusServer == "" {
		log.Fatal("Configuration file invalid: No Prometheus server is defined.\n")
	}

	if len(conf.ProcessorConfig.Scales) < 1 {
		log.Fatal("Processor configuration doesn't contain any scale definitions.\n")
	}

	for _, scale := range conf.ProcessorConfig.Scales {
		if scale.Name == "" {
			log.Fatal("Configuration file invalid: Scale defined without name.\n")
		}
		if scale.Intervals == nil || len(scale.Intervals) < 1 {
			log.Fatalf("Configuration file invalid: %s scale defined without any intervals.\n", scale.Name)
		}
	}

	return conf
}

func initializeBackend() {

	log = logging.NewLogger()

	scraper = prometheus.NewScraper(log, configuration.PrometheusServer, prometheus.Playback)
	metricProcessor = processor.NewProcessor(log, configuration.ProcessorConfig, scraper.Output)
	midiEmitter = midioutput.NewMidi(log, metricProcessor.Output)
	fractalRenderer = fractals.NewFractalRenderer(log)

}

func initializeGUI() {

	context := imgui.CreateContext(nil)
	defer context.Destroy()
	io := imgui.CurrentIO()

	platform, err := platforms.NewGLFW(io, platforms.GLFWClientAPIOpenGL3)

	platform.AddKeyboardCallback(fractalRenderer.KeyCallback)

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

	gui.Run(platform, renderer, log, scraper, metricProcessor, midiEmitter, fractalRenderer)
}
