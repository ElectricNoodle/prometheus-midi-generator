package main

type ProcessorControlMessage struct {
	Value int
}

type processor struct {
	control <-chan ProcessorControlMessage
	input   <-chan float64
	output  chan<- MidiMessage
}

func newProcessor(controlChannel <-chan ProcessorControlMessage, inputChannel <-chan float64, outputChannel chan<- MidiMessage) *processor {

	processor := processor{controlChannel, inputChannel, outputChannel}

	return &processor
}
