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

const (
	note      eventType = 0
	parameter eventType = 1
)

type eventState int

const (
	ready   eventState = 0
	active  eventState = 1
	stopped eventState = 2
)

type event struct {
	eventType eventType
	state     eventState
	duration  int
	value     string
}

var events []event

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
	ChromaticScale []string
	Ionian         []string
	Dorian         []string
	Phrygian       []string
	Mixolydian     []string
	Aeolian        []string
	Locrian        []string
}

const defaultBPM = 30
const defaultTick = 250

/*Processor Holds input/output info and generation parameters.*/
type Processor struct {
	control     <-chan ControlMessage
	input       <-chan float64
	output      chan<- midioutput.MidiMessage
	BPM         float64
	Tick        time.Duration
	scaleTypes  scaleTheory
	activeScale []string
}

/*NewProcessor returns a new instance of the processor stack and starts the control/generation threads. */
func NewProcessor(controlChannel <-chan ControlMessage, inputChannel <-chan float64, outputChannel chan<- midioutput.MidiMessage) *Processor {

	processor := Processor{controlChannel, inputChannel, outputChannel, defaultBPM, defaultTick, scaleTheory{}, []string{}}

	processor.initScaleTypes()
	processor.setActiveScale(processor.scaleTypes.ChromaticScale)

	go processor.controlThread()
	go processor.generationThread()

	return &processor
}

func (collector *Processor) setActiveScale(scale []string) {
	collector.activeScale = scale
}

func (collector *Processor) getNotes(offsets []int) []string {
	retNotes := make([]string, len(offsets))

	for i, offset := range offsets {
		retNotes[i] = notes[offset]
	}

	return retNotes
}

func (collector *Processor) initScaleTypes() {

	collector.scaleTypes.ChromaticScale = make([]string, len(notes))
	collector.scaleTypes.ChromaticScale = notes

	collector.scaleTypes.Ionian = collector.getNotes(ionianOffsets)
	collector.scaleTypes.Dorian = collector.getNotes(dorianOffsets)
	collector.scaleTypes.Phrygian = collector.getNotes(phrygianOffsets)
	collector.scaleTypes.Mixolydian = collector.getNotes(mixolydianOffsets)
	collector.scaleTypes.Aeolian = collector.getNotes(aeolianOffsets)
	collector.scaleTypes.Locrian = collector.getNotes(locrianOffsets)

	fmt.Printf("Chromatic: %v+ \n", collector.scaleTypes.ChromaticScale)
	fmt.Printf("Ionian: %v+ \n", collector.scaleTypes.Ionian)
	fmt.Printf("Dorian: %v+ \n", collector.scaleTypes.Dorian)
	fmt.Printf("Phrygian: %v+ \n", collector.scaleTypes.Phrygian)
	fmt.Printf("Mixolydian: %v+ \n", collector.scaleTypes.Mixolydian)
	fmt.Printf("Aeolian: %v+ \n", collector.scaleTypes.Aeolian)
	fmt.Printf("Locrian: %v+ \n", collector.scaleTypes.Locrian)

}

/* This function listens for any incoming messages and handles them accordingly */
func (collector *Processor) controlThread() {
	for {
		message := <-collector.control
		fmt.Printf("TEST %f\n", message.Value)

	}
}

func (collector *Processor) generationThread() {
	var tick float64
	tick = 0
	for {

		select {
		case message := <-collector.input:
			fmt.Printf("ProcessorValue: %f \n", message)
			collector.pushEvent(message)
			//collector.output <- message
		default:
			//fmt.Println("no message received")

			//tick = tick % ((collector.BPM / 60) * 1000)
			//fmt.Printf("Tick..%f\n", tick)
			//time.Duration(value) * time.Millisecond

			if tick == 0 {
				fmt.Println("BEEP")
			}

			tick += float64(collector.Tick)
			tick = math.Mod(tick, (60/collector.BPM)*1000)
			time.Sleep(collector.Tick * time.Millisecond)
		}

	}
}

func (collector *Processor) pushEvent(value float64) {

	note := collector.activeScale[int(value)%len(collector.activeScale)]
	fmt.Printf("Note: %s \n", note)

}

func (collector *Processor) tick() {

}
