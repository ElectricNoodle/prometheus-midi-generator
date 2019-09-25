package main

import (
	"fyne.io/fyne/widget"
	"fyne.io/fyne/app"
)

func main() {


	app := app.New()

	prometheusChannel := make(chan string)
	outputChannel := make(chan int)
	
	prometheus := newPrometheusScraper("http://192.168.150.187:9090/api/v1/query_range", prometheusChannel, outputChannel)
	prometheus.queryPrometheus("stddev_over_time(pf_current_entries_total{instance=~\"sovapn[1|2]:9116\"}[12h])", 1568722200, 1569327600, 600);

	w := app.NewWindow("Hello")
	w.SetContent(widget.NewVBox(
		widget.NewLabel("Hello Fyne!"),
		widget.NewButton("Quit", func() {
			app.Quit()
		}),
	))

	w.ShowAndRun()
}