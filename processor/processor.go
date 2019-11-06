package processor

import (
	"fmt"
	"math"
	"prometheus-midi-generator/midioutput"
	"time"
)

var notes = [13]string{"CLow", "C#", "D", "D#", "E", "G", "F#", "G", "G#", "A", "A#", "B", "CHigh"}

var ionianOffsets = [8]int{0, 1, 4, 5, 7, 9, 11, 12}
var dorianOffsets = [8]int{0, 2, 3, 5, 7, 9, 10, 12}
var phrygianOffsets = [8]int{0, 1, 3, 5, 7, 8, 10, 12}
var mixolydianOffsets = [8]int{0, 2, 4, 5, 7, 9, 10, 12}
var aeolianOffsets = [8]int{0, 2, 3, 5, 7, 8, 10, 12}
var locrianOffsets = [8]int{0, 1, 3, 5, 6, 8, 10, 12}

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
	ChromaticScale [13]string
	Ionian         [8]string
	Dorian         [8]string
	Phrygian       [8]string
	Mixolydian     [8]string
	Aeolian        [8]string
	Locrian        [8]string
}

const defaultBPM = 30
const defaultTick = 250

type processor struct {
	control <-chan ControlMessage
	input   <-chan float64
	output  chan<- midioutput.MidiMessage
	BPM     float64
	Tick    time.Duration
}

/*NewProcessor returns a new instance of the processor stack and starts the control/generation threads. */
func NewProcessor(controlChannel <-chan ControlMessage, inputChannel <-chan float64, outputChannel chan<- midioutput.MidiMessage) *processor {

	processor := processor{controlChannel, inputChannel, outputChannel, defaultBPM, defaultTick}

	go processor.controlThread()
	go processor.generationThread()

	return &processor
}

/* This function listens for any incoming messages and handles them accordingly */
func (collector *processor) controlThread() {
	for {
		message := <-collector.control
		fmt.Printf("TEST %f\n", message.Value)

	}
}

func (collector *processor) generationThread() {
	var tick float64
	tick = 0
	for {

		select {
		case message := <-collector.input:
			fmt.Printf("ProcessorValue: %f \n", message)
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

func (collector *processor) tick() {

}
