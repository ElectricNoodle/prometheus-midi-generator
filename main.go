package main

import (
	"fyne.io/fyne/widget"
	"fyne.io/fyne/app"
)

func main() {


	app := app.New()

	prometheusChannel := make(chan string)
	outputChannel := make(chan int)
	
	prometheus := newPrometheusScraper("192.168.150.187:9090", prometheusChannel, outputChannel)

	prometheus.Test()

	w := app.NewWindow("Hello")
	w.SetContent(widget.NewVBox(
		widget.NewLabel("Hello Fyne!"),
		widget.NewButton("Quit", func() {
			app.Quit()
		}),
	))

	w.ShowAndRun()
}