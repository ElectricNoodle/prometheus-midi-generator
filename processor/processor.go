package processor

import (
	"fmt"
	"math"
	"prometheus-midi-generator/midioutput"
	"time"
)

var notes = []string{"CLow", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B", "CHigh"}

var chromaticOffsets = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
var ionianOffsets = []int{0, 1, 4, 5, 7, 9, 11, 12}
var dorianOffsets = []int{0, 2, 3, 5, 7, 9, 10, 12}
var phrygianOffsets = []int{0, 1, 3, 5, 7, 8, 10, 12}
var mixolydianOffsets = []int{0, 2, 4, 5, 7, 9, 10, 12}
var aeolianOffsets = []int{0, 2, 3, 5, 7, 8, 10, 12}
var locrianOffsets = []int{0, 1, 3, 5, 6, 8, 10, 12}

type eventType int

/* Used fpr defining event types in MidiMessages */
const (
	Note      eventType = 0
	Parameter eventType = 1
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

type scaleTheory struct {
	Chromatic  []string
	Ionian     []string
	Dorian     []string
	Phrygian   []string
	Mixolydian []string
	Aeolian    []string
	Locrian    []string
}

const maxEvents = 10
const defaultBPM = 60
const defaultTick = 250

/*Processor Holds input/output info and generation parameters.*/
type Processor struct {
	control     <-chan ControlMessage
	input       <-chan float64
	output      chan<- midioutput.MidiMessage
	BPM         float64
	TickInc     time.Duration
	tick        float64
	scaleTypes  scaleTheory
	activeScale []string
	events      []event
}

/*NewProcessor returns a new instance of the processor stack and starts the control/generation threads. */
func NewProcessor(controlChannel <-chan ControlMessage, inputChannel <-chan float64, outputChannel chan<- midioutput.MidiMessage) *Processor {

	processor := Processor{controlChannel, inputChannel, outputChannel, defaultBPM, defaultTick, 0, scaleTheory{}, []string{}, []event{}}

	processor.initScaleTypes()
	processor.setActiveScale(processor.scaleTypes.Chromatic)

	processor.events = make([]event, maxEvents)

	go processor.controlThread()
	go processor.generationThread()

	return &processor
}

func (processor *Processor) setActiveScale(scale []string) {

	fmt.Printf("Active Scale: %#v\n", scale)
	processor.activeScale = scale

}

func (processor *Processor) getNotes(offsets []int) []string {

	retNotes := make([]string, len(offsets))

	for i, offset := range offsets {
		retNotes[i] = notes[offset]
	}

	return retNotes
}

func (processor *Processor) initScaleTypes() {

	processor.scaleTypes.Chromatic = make([]string, len(notes))
	processor.scaleTypes.Chromatic = notes

	processor.scaleTypes.Ionian = processor.getNotes(ionianOffsets)
	processor.scaleTypes.Dorian = processor.getNotes(dorianOffsets)
	processor.scaleTypes.Phrygian = processor.getNotes(phrygianOffsets)
	processor.scaleTypes.Mixolydian = processor.getNotes(mixolydianOffsets)
	processor.scaleTypes.Aeolian = processor.getNotes(aeolianOffsets)
	processor.scaleTypes.Locrian = processor.getNotes(locrianOffsets)

	fmt.Printf("Chromatic: %v+ \n", processor.scaleTypes.Chromatic)
	fmt.Printf("Ionian: %v+ \n", processor.scaleTypes.Ionian)
	fmt.Printf("Dorian: %v+ \n", processor.scaleTypes.Dorian)
	fmt.Printf("Phrygian: %v+ \n", processor.scaleTypes.Phrygian)
	fmt.Printf("Mixolydian: %v+ \n", processor.scaleTypes.Mixolydian)
	fmt.Printf("Aeolian: %v+ \n", processor.scaleTypes.Aeolian)
	fmt.Printf("Locrian: %v+ \n", processor.scaleTypes.Locrian)

}

/* This function listens for any incoming messages and handles them accordingly */
func (processor *Processor) controlThread() {

	for {
		message := <-processor.control
		fmt.Printf("TEST %f\n", message.Value)

	}
}

func (processor *Processor) generationThread() {

	processor.tick = 0

	for {

		select {

		case message := <-processor.input:

			processor.processMessage(message)

		default:

			/* Means we're on the beat. */
			if processor.tick == 0 {
				processor.handleEvents()
				fmt.Println("Boop")
			}
			processor.incrementTick()
		}

	}
}

func (processor *Processor) processMessage(value float64) {
	fmt.Printf("Processor Metric Value: %f Binary: %b\n", value, math.Float64bits(value))
	noteIndex := int(value) % len(processor.activeScale)
	event := event{Note, ready, 2, noteIndex, 4}
	processor.insertEvent(event)

}

func (processor *Processor) handleEvents() {

	for i, e := range processor.events {

		if (event{}) != e {

			if e.state == ready {

				fmt.Printf("Send start %s Oct: %d \n", processor.activeScale[e.value], e.octave)

				processor.events[i].state = active
				processor.output <- midioutput.MidiMessage{midioutput.Channel1, midioutput.NoteOn, processor.events[i].value, processor.events[i].octave, 50}

			} else if e.state == active {

				processor.events[i].duration--

				if e.duration == 1 {

					fmt.Printf("Send stop %s Oct: %d \n", processor.activeScale[e.value], e.octave)

					processor.events[i].state = stop
					processor.output <- midioutput.MidiMessage{midioutput.Channel1, midioutput.NoteOff, processor.events[i].value, processor.events[i].octave, 50}

				}

			} else if e.state == stop {
				processor.events[i] = event{}
			}
		}
	}
}

func (processor *Processor) insertEvent(eventIn event) {
	for i, e := range processor.events {
		if (event{}) == e {
			processor.events[i] = eventIn
			break
		}
	}

}

func (processor *Processor) incrementTick() {

	processor.tick += float64(processor.TickInc)
	processor.tick = math.Mod(processor.tick, (60/processor.BPM)*1000)
	time.Sleep(processor.TickInc * time.Millisecond)

}
