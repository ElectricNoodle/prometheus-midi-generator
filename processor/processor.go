package processor

import (
	"fmt"
	"prometheus-midi-generator/midi"
)

var notes = [13]string{"CLow", "C#", "D", "D#", "E", "G", "F#", "G", "G#", "A", "A#", "B", "CHigh"}

var ionianOffsets = [8]int{0, 1, 4, 5, 7, 9, 11, 12}
var dorianOffsets = [8]int{0, 2, 3, 5, 7, 9, 10, 12}
var phrygianOffsets = [8]int{0, 1, 3, 5, 7, 8, 10, 12}
var mixolydianOffsets = [8]int{0, 2, 4, 5, 7, 9, 10, 12}
var aeolianOffsets = [8]int{0, 2, 3, 5, 7, 8, 10, 12}
var locrianOffsets = [8]int{0, 1, 3, 5, 6, 8, 10, 12}

type MessageType int

const (
	StartOutput      MessageType = 0
	StopOutput       MessageType = 1
	ChangePollRate   MessageType = 2
	ChangeOutputRate MessageType = 3
)

type ControlMessage struct {
	Type  MessageType
	Value float64
}

type ScaleTheory struct {
	ChromaticScale [13]string
	Ionian         [8]string
	Dorian         [8]string
	Phrygian       [8]string
	Mixolydian     [8]string
	Aeolian        [8]string
	Locrian        [8]string
}

const DEFAULT_BPM = 60

type processor struct {
	control <-chan ControlMessage
	input   <-chan float64
	output  chan<- midi.MidiMessage
	BPM     float64
}

func NewProcessor(controlChannel <-chan ControlMessage, inputChannel <-chan float64, outputChannel chan<- midi.MidiMessage) *processor {

	processor := processor{controlChannel, inputChannel, outputChannel, DEFAULT_BPM}

	go processor.processorControlThread()

	return &processor
}

/* This function listens for any incoming messages and handles them accordingly */
func (collector *processor) processorControlThread() {
	for {
		message := <-collector.control
		fmt.Printf("TEST %f\n", message.Value)

	}
}

func (collector *processor) processorGenerationThread() {
	for {
		message := <-collector.input
		fmt.Printf("Value: %f \n", message)
	}
}
