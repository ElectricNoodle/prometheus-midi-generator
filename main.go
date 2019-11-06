package main

import (
	"fmt"
	"prometheus-midi-generator/midioutput"
	"prometheus-midi-generator/processor"
	"prometheus-midi-generator/prometheus"
	"time"
	//"fyne.io/fyne/widget"
	//"fyne.io/fyne/app"
)

func main() {

	//app := app.New()

	prometheusControlChannel := make(chan prometheus.ControlMessage, 3)
	prometheusOutputChannel := make(chan float64, 3)

	processorControlChannel := make(chan processor.ControlMessage, 3)
	processorOutputChannel := make(chan midioutput.MidiMessage, 3)

	midiControlChannel := make(chan midioutput.MidiControlMessage, 3)

	prometheusScraper := prometheus.NewScraper("http://192.168.150.187:9090/api/v1/query_range", prometheus.Live, prometheusControlChannel, prometheusOutputChannel)
	prometheusProcessor := processor.NewProcessor(processorControlChannel, prometheusOutputChannel, processorOutputChannel)
	midiOutput := midioutput.NewMidi(midiControlChannel, processorOutputChannel)

	fmt.Printf("%s\n", prometheusScraper.Target)
	fmt.Printf("%f\n", prometheusProcessor.BPM)
	fmt.Printf("%d\n", midiOutput.Port)
	//0:Array[1572469200,9216.632296877477]

	queryInfo := prometheus.QueryInfo{"stddev_over_time(pf_current_entries_total{instance=~\"sovapn1:9116\"}[12h])", 1573075602, 1573075902, 600}

	messageStart := prometheus.ControlMessage{prometheus.StartOutput, prometheus.Live, queryInfo, 0}
	//messageStop := prometheus.ControlMessage{prometheus.StopOutput, 0, prometheus.QueryInfo{}, 0}

	prometheusControlChannel <- messageStart
	//prometheusControlChannel <- messageStop

	for {
		time.Sleep(2000 * time.Millisecond)
	}

	/*
		w := app.NewWindow("Hello")
		w.SetContent(widget.NewVBox(
			widget.NewLabel("Hello Fyne!"),
			widget.NewButton("Quit", func() {
				app.Quit()
			}),
		))

		w.ShowAndRun()
	*/
}
