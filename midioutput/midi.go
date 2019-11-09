package midioutput

import (
	"fmt"
	"log"
	"time"

	"github.com/rakyll/portmidi"
)

type MessageType int

const (
	NoteOn        MessageType = 0
	NoteOff       MessageType = 1
	ControlChange MessageType = 2
)

type MidiMessage struct {
	Channel  int
	Type     MessageType
	Note     string
	Octave   int
	Velocity int
}
type MidiControlMessage struct {
	Value int
}

type midiinfo struct {
	control          <-chan MidiControlMessage
	input            <-chan MidiMessage
	Port             int
	MIDIOutputStream *portmidi.Stream
}

func NewMidi(controlChannel <-chan MidiControlMessage, inputChannel <-chan MidiMessage) *midiinfo {

	midiProcessor := midiinfo{controlChannel, inputChannel, 2, nil}
	portmidi.Initialize()
	out, err := portmidi.NewOutputStream(2, 1024, 0)

	if err != nil {
		log.Fatal(err)
	}

	midiProcessor.MIDIOutputStream = out

	return &midiProcessor
}

func (midiEmitter *midiinfo) midiEmitThread() {
	for {
		message := <-midiEmitter.input

		fmt.Printf("MProcessorValue: %v \n", message)

		midiEmitter.MIDIOutputStream.WriteShort(0x91, 60, 100)
		midiEmitter.MIDIOutputStream.WriteShort(0x91, 64, 100)
		midiEmitter.MIDIOutputStream.WriteShort(0x91, 67, 100)

		// notes will be sustained for 2 seconds
		time.Sleep(2 * time.Second)

		// note off events
		midiEmitter.MIDIOutputStream.WriteShort(0x81, 60, 100)
		midiEmitter.MIDIOutputStream.WriteShort(0x81, 64, 100)
		midiEmitter.MIDIOutputStream.WriteShort(0x81, 67, 100)

	}
}
