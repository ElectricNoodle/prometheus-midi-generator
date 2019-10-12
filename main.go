package main

import (
	"fmt"
	"time"
	//"fyne.io/fyne/widget"
	//"fyne.io/fyne/app"
)

func main() {

	//app := app.New()

	prometheusChannel := make(chan ControlMessage, 3)
	outputChannel := make(chan int, 3)

	prometheus := newPrometheusScraper("http://192.168.150.187:9090/api/v1/query_range", Live, prometheusChannel, outputChannel)

	fmt.Printf("%s\n", prometheus.Target)

	queryInfo := QueryInfo{"stddev_over_time(pf_current_entries_total{instance=~\"sovapn[1|2]:9116\"}[12h])", 1568722200, 1569327600, 600}
	messageStart := ControlMessage{StartOutput, Live, queryInfo}
	messageStop := ControlMessage{StopOutput, 0, QueryInfo{}}

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
