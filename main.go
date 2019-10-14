package main

import (
	"fmt"
	"time"
	//"fyne.io/fyne/widget"
	//"fyne.io/fyne/app"
)

func main() {

	//app := app.New()

	prometheusChannel := make(chan PrometheusControlMessage, 3)
	outputChannel := make(chan float64, 3)

	prometheus := newPrometheusScraper("http://192.168.150.187:9090/api/v1/query_range", Live, prometheusChannel, outputChannel)

	fmt.Printf("%s\n", prometheus.Target)

	queryInfo := QueryInfo{"stddev_over_time(pf_current_entries_total{instance=~\"sovapn[1|2]:9116\"}[12h])", 1568722200, 1569327600, 600}

	messageStart := PrometheusControlMessage{StartOutput, Live, queryInfo, 0}
	messageStop := PrometheusControlMessage{StopOutput, 0, QueryInfo{}, 0}

	prometheusChannel <- messageStart

	prometheusChannel <- messageStop

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
