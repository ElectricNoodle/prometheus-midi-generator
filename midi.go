package main

type MidiMessage struct {
}
type MidiControlMessage struct {
	Value int
}

type midi struct {
	control <-chan MidiControlMessage
	input   <-chan MidiMessage
}

func newMidi(controlChannel <-chan MidiControlMessage, inputChannel <-chan MidiMessage) *midi {

	processor := midi{controlChannel, inputChannel}

	return &processor
}
