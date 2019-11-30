package processor

import (
	"fmt"
	"math"
	"prometheus-midi-generator/midioutput"
	"time"
)

type rootNote int

/* Const indexes for note values at positions so code is more readable. NOTE: Removed CLow/CHigh concept, nnot sure if it's needed here. */
const (
	C      rootNote = 0
	CSharp rootNote = 1
	D      rootNote = 2
	DSharp rootNote = 3
	E      rootNote = 4
	F      rootNote = 5
	FSharp rootNote = 6
	G      rootNote = 7
	GSharp rootNote = 8
	A      rootNote = 9
	ASharp rootNote = 10
	B      rootNote = 11
)

/* 3 octaves of Chromatic scale which allows for generation of any scale type with any root note */
var notes = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B", "C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B", "C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B", "C"}

var chromaticOffsets = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
var ionianOffsets = []int{0, 2, 4, 5, 7, 9, 11, 12}
var dorianOffsets = []int{0, 2, 3, 5, 7, 9, 10, 12}
var phrygianOffsets = []int{0, 1, 3, 5, 7, 8, 10, 12}
var lydianOffsets = []int{0, 2, 4, 6, 7, 9, 11, 12}
var mixolydianOffsets = []int{0, 2, 4, 5, 7, 9, 10, 12}
var aeolianOffsets = []int{0, 2, 3, 5, 7, 8, 10, 12}
var locrianOffsets = []int{0, 1, 3, 5, 6, 8, 10, 12}

type eventType int

const (
	note      eventType = 0
	parameter eventType = 1
)

type eventState int

const (
	ready  eventState = 0
	active eventState = 1
	stop   eventState = 2
)

type event struct {
	eventType eventType
	state     eventState
	duration  int
	value     int
	octave    int
}

/*MessageType Defines the different type of Control Message.*/
type MessageType int

/* Values for MessageType */
const (
	StartOutput      MessageType = 0
	StopOutput       MessageType = 1
	ChangePollRate   MessageType = 2
	ChangeOutputRate MessageType = 3
)

/*ControlMessage Used for sending control messages to processor.*/
type ControlMessage struct {
	Type  MessageType
	Value float64
}

type scaleTypes struct {
	Chromatic  []string
	Ionian     []string
	Dorian     []string
	Phrygian   []string
	Lydian     []string
	Mixolydian []string
	Aeolian    []string
	Locrian    []string
}

const maxEvents = 10
const defaultBPM = 60
const defaultTick = 250

/*ProcInfo Holds input/output info and generation parameters.*/
type ProcInfo struct {
	control        <-chan ControlMessage
	input          <-chan float64
	output         chan<- midioutput.MidiMessage
	BPM            float64
	TickInc        time.Duration
	tick           float64
	scales         scaleTypes
	activeScale    []string
	rootNoteOffset int
	events         []event
}

/*NewProcessor returns a new instance of the processor stack and starts the control/generation threads. */
func NewProcessor(controlChannel <-chan ControlMessage, inputChannel <-chan float64, outputChannel chan<- midioutput.MidiMessage) *ProcInfo {

	processor := ProcInfo{controlChannel, inputChannel, outputChannel, defaultBPM, defaultTick, 0, scaleTypes{}, []string{}, 0, []event{}}

	processor.initScaleTypes(F)
	processor.activeScale = processor.scales.Ionian
	fmt.Printf("ActiveScale: %v+\n", processor.activeScale)
	processor.events = make([]event, maxEvents)

	go processor.controlThread()
	go processor.generationThread()

	return &processor
}

func (processor *ProcInfo) getNotes(rootOffset rootNote, offsets []int) []string {
	fmt.Printf("RootOffset: %d \n", rootOffset)
	retNotes := make([]string, len(offsets))

	for i, offset := range offsets {
		retNotes[i] = notes[((int(rootOffset) + offset) % len(notes))]
	}

	return retNotes
}

func (processor *ProcInfo) initScaleTypes(rootNoteIndex rootNote) {

	processor.scales.Chromatic = make([]string, len(notes))
	processor.scales.Chromatic = notes

	processor.scales.Ionian = processor.getNotes(rootNoteIndex, ionianOffsets)
	processor.scales.Dorian = processor.getNotes(rootNoteIndex, dorianOffsets)
	processor.scales.Phrygian = processor.getNotes(rootNoteIndex, phrygianOffsets)
	processor.scales.Mixolydian = processor.getNotes(rootNoteIndex, mixolydianOffsets)
	processor.scales.Lydian = processor.getNotes(rootNoteIndex, lydianOffsets)
	processor.scales.Aeolian = processor.getNotes(rootNoteIndex, aeolianOffsets)
	processor.scales.Locrian = processor.getNotes(rootNoteIndex, locrianOffsets)

	processor.rootNoteOffset = int(rootNoteIndex)

}

/* This function listens for any incoming messages and handles them accordingly */
func (processor *ProcInfo) controlThread() {

	for {
		message := <-processor.control
		fmt.Printf("TEST %f\n", message.Value)

	}
}

func (processor *ProcInfo) generationThread() {

	processor.tick = 0

	for {

		select {

		case message := <-processor.input:

			processor.processMessage(message)

		default:

			if processor.tick == 0 {

				processor.handleEvents()

			}

			processor.incrementTick()

		}

	}
}

func (processor *ProcInfo) processMessage(value float64) {

	noteVal := int(value) % len(processor.activeScale)
	event := event{note, ready, 4, noteVal, 3}
	processor.insertEvent(event)
	fmt.Printf("Note: %s Value: %f \n", processor.activeScale[int(value)%len(processor.activeScale)], value)

}

func (processor *ProcInfo) handleEvents() {

	for i, e := range processor.events {

		if (event{}) != e {

			if e.state == ready {

				fmt.Printf("Send start %d Oct: %d \n", processor.rootNoteOffset+e.value, e.octave)

				processor.events[i].state = active

				processor.output <- midioutput.MidiMessage{midioutput.Channel1, midioutput.NoteOn, processor.rootNoteOffset + processor.events[i].value, processor.events[i].octave, 80}

			} else if e.state == active {

				processor.events[i].duration--

				if e.duration == 1 {

					fmt.Printf("Send stop %d Oct: %d \n", processor.rootNoteOffset+e.value, e.octave)

					processor.events[i].state = stop

					processor.output <- midioutput.MidiMessage{midioutput.Channel1, midioutput.NoteOff, processor.rootNoteOffset + processor.events[i].value, processor.events[i].octave, 50}

				}

			} else if e.state == stop {
				processor.events[i] = event{}
			}
		}
	}
}

func (processor *ProcInfo) insertEvent(eventIn event) {
	for i, e := range processor.events {
		if (event{}) == e {
			processor.events[i] = eventIn
			break
		}
	}

}

func (processor *ProcInfo) incrementTick() {
	fmt.Printf("Tick: %f \n", processor.tick)
	processor.tick += float64(processor.TickInc)
	processor.tick = math.Mod(processor.tick, (60/processor.BPM)*1000)
	time.Sleep(processor.TickInc * time.Millisecond)

}
