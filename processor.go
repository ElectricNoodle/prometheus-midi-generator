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

	go processor.processorControlThread()

	return &processor
}

/* This function listens for any incoming messages and handles them accordingly */
func (collector *processor) processorControlThread() {
	for {

		//message := <-collector.control

	}
}
