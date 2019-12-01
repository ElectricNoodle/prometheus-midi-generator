package gui

import (
	"time"

	"github.com/inkyblackness/imgui-go"
)

// Platform covers mouse/keyboard/gamepad inputs, cursor shape, timing, windowing.
type Platform interface {
	// ShouldStop is regularly called as the abort condition for the program loop.
	ShouldStop() bool
	// ProcessEvents is called once per render loop to dispatch any pending events.
	ProcessEvents()
	// DisplaySize returns the dimension of the display.
	DisplaySize() [2]float32
	// FramebufferSize returns the dimension of the framebuffer.
	FramebufferSize() [2]float32
	// NewFrame marks the begin of a render pass. It must update the imgui IO state according to user input (mouse, keyboard, ...)
	NewFrame()
	// PostRender marks the completion of one render pass. Typically this causes the display buffer to be swapped.
	PostRender()
	// ClipboardText returns the current text of the clipboard, if available.
	ClipboardText() (string, error)
	// SetClipboardText sets the text as the current text of the clipboard.
	SetClipboardText(text string)
}

type clipboard struct {
	platform Platform
}

func (board clipboard) Text() (string, error) {
	return board.platform.ClipboardText()
}

func (board clipboard) SetText(text string) {
	board.platform.SetClipboardText(text)
}

// Renderer covers rendering imgui draw data.
type Renderer interface {
	// PreRender causes the display buffer to be prepared for new output.
	PreRender(clearColor [4]float32)
	// Render draws the provided imgui draw data.
	Render(displaySize [2]float32, framebufferSize [2]float32, drawData imgui.DrawData)
}

var prometheusPollRatePos int32
var prometheusPollRate = 4000

var metric = "pf_current_entries_total{instance=~'sovapn1:9116'}"

var prometheusPollRates = []int{4000, 5000, 6000, 7000, 8000}
var prometheusPollRatesString = []string{"4000", "5000", "6000", "7000", "8000"}

var prometheusModePos int32
var prometheusMode = "Live"
var prometheusModes = []string{"Live", "Playback"}

var prometheusStartDate = "2019-11-25 12:00"
var prometheusEndDate = "2019-11-30 12:00"

var processorKeysPos int32
var processorKey = "C"
var processorKeys = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

var processorModePos int32
var processorMode = "Chromatic"
var processorModes = []string{"Chromatic", "Ionian", "Dorian", "Phrygian", "Lydian", "Mixolydian", "Aeolian", "Locrian"}

var processorGenerationTypePos int32
var processorGenerationType = "Chromatic"
var processorGenerationTypes = []string{"Modulus(Ch1)", "ModulusPlus(Ch1)", "ModulusChords(Ch1)", "ModulusPlusChords(Ch1)", "Binary Arp(Ch1)", "Modulus(Ch1) + BinaryArp(Ch2)", "ModulusPlus(Ch1) + BinaryArp(Ch2)"}

// Run implements the main program loop of the demo. It returns when the platform signals to stop.
// This demo application shows some basic features of ImGui, as well as exposing the standard demo window.
func Run(p Platform, r Renderer) {
	imgui.CurrentIO().SetClipboard(clipboard{platform: p})

	showDemoWindow := false
	clearColor := [4]float32{0.0, 0.0, 0.0, 1.0}
	//f := float32(0.0)
	//counter := 0
	showAnotherWindow := false

	for !p.ShouldStop() {
		p.ProcessEvents()

		// Signal start of a new frame
		p.NewFrame()
		imgui.NewFrame()

		// 1. Show the big demo window (Most of the sample code is in ImGui::ShowDemoWindow()!
		// You can browse its code to learn more about Dear ImGui!).
		if showDemoWindow {
			imgui.ShowDemoWindow(&showDemoWindow)
		}

		{
			imgui.Begin("Prometheus Fractal/MIDI Generator")                                     // Create a window called "Hello, world!" and append into it.
			imgui.Text("A visual/musical generation/exploration tool using Prometheus metrics.") // Display some text

			imgui.Text("\t\t")
			imgui.Separator()

			renderPrometheusOptions()

			imgui.Text("\t")
			imgui.Separator()

			renderProcessorOptions()

			//imgui.Checkbox("Demo Window", &showDemoWindow) // Edit bools storing our window open/close state
			//imgui.Checkbox("Another Window", &showAnotherWindow)
			//	imgui
			//imgui.SliderFloat("float", &f, 0.0, 1.0) // Edit one float using a slider from 0.0f to 1.0f
			// TODO add example of ColorEdit3 for clearColor

			//if imgui.Button("Button") { // Buttons return true when clicked (most widgets return true when edited/activated)
			//		counter++
			//	}
			//	imgui.SameLine()
			//imgui.Text(fmt.Sprintf("counter = %d", counter))

			// TODO add text of FPS based on IO.Framerate()

			imgui.End()
		}

		// 3. Show another simple window.
		if showAnotherWindow {
			// Pass a pointer to our bool variable (the window will have a closing button that will clear the bool when clicked)
			imgui.BeginV("Another window", &showAnotherWindow, 0)

			imgui.Text("Hello from another window!")
			if imgui.Button("Close Me") {
				showAnotherWindow = false
			}
			imgui.End()
		}

		// Rendering
		imgui.Render() // This call only creates the draw data list. Actual rendering to framebuffer is done below.

		r.PreRender(clearColor)
		// A this point, the application could perform its own rendering...
		// app.RenderScene()

		r.Render(p.DisplaySize(), p.FramebufferSize(), imgui.RenderedDrawData())
		p.PostRender()

		// sleep to avoid 100% CPU usage for this demo
		<-time.After(time.Millisecond * 25)
	}
}

func renderPrometheusOptions() {

	imgui.Text("Prometheus Configuration:")

	imgui.Text("\t")

	imgui.Text("Metric:    ")
	//imgui.SameLine()
	imgui.InputText("", &metric)

	imgui.Text("\t")

	imgui.Text("Poll Rate (ms): ")
	//	imgui.SameLine()
	//imgui.SliderInt("", &prometheusPollRate, 1000, 10000) // Edit one float using a slider from 0.0f to 1.0f
	if imgui.ListBoxV(" ", &prometheusPollRatePos, prometheusPollRatesString, 1) {
		prometheusPollRate = prometheusPollRates[prometheusPollRatePos]
	}

	imgui.Text("\t")
	if imgui.ListBoxV("\t", &prometheusModePos, prometheusModes, 2) {
		prometheusMode = prometheusModes[prometheusModePos]
	}

	if prometheusMode == "Playback" {

		imgui.Text("\t")
		imgui.Text("Start Time: ")
		imgui.InputText("", &prometheusStartDate)

		imgui.Text("\t")
		imgui.Text("End Time:   ")
		imgui.InputText("", &prometheusEndDate)

	}

	imgui.Text("\t")
	if imgui.Button("Start") {

	}
	imgui.SameLine()
	imgui.Text(" ")
	imgui.SameLine()
	if imgui.Button("Stop") {

	}
}

func renderProcessorOptions() {
	imgui.Text("Processor Musical Options:")
	imgui.Text("\t")

	imgui.Text("Key:")
	if imgui.ListBoxV("  ", &processorKeysPos, processorKeys, 1) {
		processorKey = processorKeys[processorKeysPos]
	}

	imgui.Text("\t")

	imgui.Text("Mode:")
	if imgui.ListBoxV("   ", &processorModePos, processorModes, 1) {
		processorMode = processorModes[processorModePos]
	}

	imgui.Text("\t")

	imgui.Text("Type of Generation:")
	if imgui.ListBoxV("    ", &processorGenerationTypePos, processorGenerationTypes, -1) {
		processorGenerationType = processorGenerationTypes[processorModePos]
	}
}
