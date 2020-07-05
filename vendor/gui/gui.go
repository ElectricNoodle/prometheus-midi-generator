package gui

import (
	"fmt"
	"log"
	"midioutput"
	"processor"
	"prometheus"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
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

var midiDevicesPos int32

var prometheusPollRatePos int32
var prometheusPollRate = 4000

var metric = "max(pf_states{instance=~'sovapn[1|2]:9116', protocol=~'tcp', state=~'ESTABLISHED:ESTABLISHED', type='fwstates', operator='jerseyt'})  + max(pf_states{instance=~'sovapn[1|2]:9100', protocol=~'tcp', state=~'ESTABLISHED:ESTABLISHED', type='nat', operator='jerseyt'})"

var prometheusModePos int32
var prometheusMode = prometheus.Live
var prometheusModes = []string{"Live", "Playback"}

var prometheusStartDate = "2020-06-01 00:00"
var prometheusEndDate = "2020-06-30 23:59"

var processorKeysPos int32

var processorModePos int32
var processorMode = "Chromatic"
var processorModes = []string{"Chromatic", "Ionian", "Dorian", "Phrygian", "Lydian", "Mixolydian", "Aeolian", "Locrian"}

var processorGenerationTypePos int32
var processorGenerationType = "Chromatic"
var processorGenerationTypes = []string{"Modulus(Ch1)", "ModulusPlus(Ch1)", "ModulusChords(Ch1)", "ModulusPlusChords(Ch1)", "Binary Arp(Ch1)", "Modulus(Ch1) + BinaryArp(Ch2)", "ModulusPlus(Ch1) + BinaryArp(Ch2)"}

/*Run Main GUI Loop that handles rendering of interface and at some point fractals... */
func Run(p Platform, r Renderer, scraper *prometheus.Scraper, procInfo *processor.ProcInfo, midiEmitter *midioutput.MIDIEmitter) {
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

			renderMIDIOptions(midiEmitter)

			imgui.Text("\t\t")
			imgui.Separator()

			renderPrometheusOptions(scraper)

			imgui.Text("\t")
			imgui.Separator()

			renderProcessorOptions(procInfo)

			renderStartStopButtons(scraper)

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
		//renderFractal(p.DisplaySize(), p.FramebufferSize())

		r.Render(p.DisplaySize(), p.FramebufferSize(), imgui.RenderedDrawData())
		p.PostRender()

		// sleep to avoid 100% CPU usage for this demo
		<-time.After(time.Millisecond * 25)
	}
}

func renderFractal(displaySize [2]float32, framebufferSize [2]float32) {

	//displayWidth, displayHeight := displaySize[0], displaySize[1]
	fbWidth, fbHeight := framebufferSize[0], framebufferSize[1]
	if (fbWidth <= 0) || (fbHeight <= 0) {
		return
	}

	// Backup GL state
	var lastActiveTexture int32
	gl.GetIntegerv(gl.ACTIVE_TEXTURE, &lastActiveTexture)
	gl.ActiveTexture(gl.TEXTURE0)
	var lastProgram int32
	gl.GetIntegerv(gl.CURRENT_PROGRAM, &lastProgram)
	var lastTexture int32
	gl.GetIntegerv(gl.TEXTURE_BINDING_2D, &lastTexture)
	var lastSampler int32
	gl.GetIntegerv(gl.SAMPLER_BINDING, &lastSampler)
	var lastArrayBuffer int32
	gl.GetIntegerv(gl.ARRAY_BUFFER_BINDING, &lastArrayBuffer)
	var lastElementArrayBuffer int32
	gl.GetIntegerv(gl.ELEMENT_ARRAY_BUFFER_BINDING, &lastElementArrayBuffer)
	var lastVertexArray int32
	gl.GetIntegerv(gl.VERTEX_ARRAY_BINDING, &lastVertexArray)
	var lastPolygonMode [2]int32
	gl.GetIntegerv(gl.POLYGON_MODE, &lastPolygonMode[0])
	var lastViewport [4]int32
	gl.GetIntegerv(gl.VIEWPORT, &lastViewport[0])
	var lastScissorBox [4]int32
	gl.GetIntegerv(gl.SCISSOR_BOX, &lastScissorBox[0])
	var lastBlendSrcRgb int32
	gl.GetIntegerv(gl.BLEND_SRC_RGB, &lastBlendSrcRgb)
	var lastBlendDstRgb int32
	gl.GetIntegerv(gl.BLEND_DST_RGB, &lastBlendDstRgb)
	var lastBlendSrcAlpha int32
	gl.GetIntegerv(gl.BLEND_SRC_ALPHA, &lastBlendSrcAlpha)
	var lastBlendDstAlpha int32
	gl.GetIntegerv(gl.BLEND_DST_ALPHA, &lastBlendDstAlpha)
	var lastBlendEquationRgb int32
	gl.GetIntegerv(gl.BLEND_EQUATION_RGB, &lastBlendEquationRgb)
	var lastBlendEquationAlpha int32
	gl.GetIntegerv(gl.BLEND_EQUATION_ALPHA, &lastBlendEquationAlpha)
	lastEnableBlend := gl.IsEnabled(gl.BLEND)
	lastEnableCullFace := gl.IsEnabled(gl.CULL_FACE)
	lastEnableDepthTest := gl.IsEnabled(gl.DEPTH_TEST)
	lastEnableScissorTest := gl.IsEnabled(gl.SCISSOR_TEST)

	var vertices = []float64{
		0.0, 0.5, 0.0,
		0.5, -0.5, 0.0,
		-0.5, -0.5, 0.0,
	}

	var VBO uint32
	var VAO uint32

	//var test [20]int
	gl.GenVertexArrays(1, &VAO)
	gl.BindVertexArray(VAO)

	gl.Viewport(0, 0, int32(fbWidth), int32(fbHeight))

	gl.GenBuffers(1, &VBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices), gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
	gl.DrawArrays(gl.TRIANGLES, 0, 3)
	gl.DisableVertexAttribArray(0)

	// Restore modified GL state
	gl.UseProgram(uint32(lastProgram))
	gl.BindTexture(gl.TEXTURE_2D, uint32(lastTexture))
	gl.BindSampler(0, uint32(lastSampler))
	gl.ActiveTexture(uint32(lastActiveTexture))
	gl.BindVertexArray(uint32(lastVertexArray))
	gl.BindBuffer(gl.ARRAY_BUFFER, uint32(lastArrayBuffer))
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, uint32(lastElementArrayBuffer))
	gl.BlendEquationSeparate(uint32(lastBlendEquationRgb), uint32(lastBlendEquationAlpha))
	gl.BlendFuncSeparate(uint32(lastBlendSrcRgb), uint32(lastBlendDstRgb), uint32(lastBlendSrcAlpha), uint32(lastBlendDstAlpha))
	if lastEnableBlend {
		gl.Enable(gl.BLEND)
	} else {
		gl.Disable(gl.BLEND)
	}
	if lastEnableCullFace {
		gl.Enable(gl.CULL_FACE)
	} else {
		gl.Disable(gl.CULL_FACE)
	}
	if lastEnableDepthTest {
		gl.Enable(gl.DEPTH_TEST)
	} else {
		gl.Disable(gl.DEPTH_TEST)
	}
	if lastEnableScissorTest {
		gl.Enable(gl.SCISSOR_TEST)
	} else {
		gl.Disable(gl.SCISSOR_TEST)
	}
	gl.PolygonMode(gl.FRONT_AND_BACK, uint32(lastPolygonMode[0]))
	gl.Viewport(lastViewport[0], lastViewport[1], lastViewport[2], lastViewport[3])
	gl.Scissor(lastScissorBox[0], lastScissorBox[1], lastScissorBox[2], lastScissorBox[3])

}

func renderMIDIOptions(midiEmitter *midioutput.MIDIEmitter) {

	imgui.Text("MIDI Configuration:")
	imgui.Text("\t")
	imgui.Text("Select Device: ")

	if imgui.ListBoxV("", &midiDevicesPos, midiEmitter.GetDeviceNames(), 2) {

		midiEmitter.Control <- midioutput.MIDIControlMessage{Type: midioutput.SetDevice, Value: int(midiDevicesPos)}

	}

	imgui.Text("\t")

}
func parseDateString(dateString string) float64 {

	layout := "2006-01-02 15:04"
	t, err := time.Parse(layout, dateString)

	if err != nil {
		log.Println(err)
		return 0
	}

	fmt.Println(t)

	return float64(t.Unix())
}
func renderPrometheusOptions(scraper *prometheus.Scraper) {

	imgui.Text("Prometheus Configuration:")
	imgui.Text("\t")

	imgui.Text("Metric:    ")
	imgui.InputText("", &metric)

	imgui.Text("\t")

	if imgui.ListBoxV(" ", &prometheusModePos, prometheusModes, 2) {
		if prometheusModes[prometheusModePos] == "Live" {
			prometheusMode = prometheus.Live
		}
		if prometheusModes[prometheusModePos] == "Playback" {
			prometheusMode = prometheus.Playback
		}
		log.Printf("Mode changed to %s\n", prometheusModes[prometheusModePos])
	}

	if prometheusMode == prometheus.Playback {

		imgui.Text("\t")
		imgui.Text("Start Time: ")
		imgui.InputText(" ", &prometheusStartDate)

		imgui.Text("\t")
		imgui.Text("End Time:   ")
		imgui.InputText("  ", &prometheusEndDate)

	}

}

/*renderProcessorOptions displays all the configurable options for sound generation. */
func renderProcessorOptions(procInfo *processor.ProcInfo) {

	imgui.Text("Processor Musical Options:")
	imgui.Text("\t")
	imgui.Text("Key:")

	if imgui.ListBoxV("  ", &processorKeysPos, procInfo.GetKeyNames(), 5) {

		message := processor.ControlMessage{Type: processor.SetKey, ValueNum: int(processorKeysPos), ValueString: ""}
		procInfo.Control <- message

	}

	imgui.Text("\t")
	imgui.Text("Mode:")

	if imgui.ListBoxV("   ", &processorModePos, procInfo.GetModeNames(), 5) {

		message := processor.ControlMessage{Type: processor.SetMode, ValueNum: 0, ValueString: processorModes[processorModePos]}
		procInfo.Control <- message

	}
	imgui.Text("\t")
	/*
		imgui.Text("Type of Generation:")
		if imgui.ListBoxV("    ", &processorGenerationTypePos, processorGenerationTypes, -1) {
			processorGenerationType = processorGenerationTypes[processorModePos]
		}
	*/
}

func renderStartStopButtons(scraper *prometheus.Scraper) {

	imgui.Text("\t")

	if imgui.Button("Start") {

		queryInfo := prometheus.QueryInfo{Query: metric, Start: parseDateString(prometheusStartDate), End: parseDateString(prometheusEndDate), Step: 600}
		message := prometheus.ControlMessage{Type: prometheus.StartOutput, OutputType: prometheusMode, QueryInfo: queryInfo, Value: 0}
		scraper.Control <- message

	}

	imgui.SameLine()
	imgui.Text(" ")
	imgui.SameLine()

	if imgui.Button("Stop") {

		messageStop := prometheus.ControlMessage{Type: prometheus.StopOutput, OutputType: prometheus.Playback, QueryInfo: prometheus.QueryInfo{}, Value: 0}
		scraper.Control <- messageStop

	}
}
