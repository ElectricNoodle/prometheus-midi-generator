package main

import (
	"fmt"
	"prometheus-midi-generator/prometheus"
	"time"
	//"fyne.io/fyne/widget"
	//"fyne.io/fyne/app"
)

func main() {

	//app := app.New()

	prometheusControlChannel := make(chan prometheus.ControlMessage, 3)
	prometheusOutputChannel := make(chan float64, 3)

	prometheusScraper := prometheus.NewPrometheusScraper("http://192.168.150.187:9090/api/v1/query_range", prometheus.Live, prometheusControlChannel, prometheusOutputChannel)

	fmt.Printf("%s\n", prometheusScraper.Target)

	queryInfo := prometheus.QueryInfo{"stddev_over_time(pf_current_entries_total{instance=~\"sovapn[1|2]:9116\"}[12h])", 1568722200, 1569327600, 600}

	messageStart := prometheus.ControlMessage{prometheus.StartOutput, prometheus.Live, queryInfo, 0}
	messageStop := prometheus.ControlMessage{prometheus.StopOutput, 0, prometheus.QueryInfo{}, 0}

	prometheusControlChannel <- messageStart
	prometheusControlChannel <- messageStop

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
