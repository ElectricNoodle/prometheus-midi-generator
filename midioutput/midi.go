package midioutput

import (
	"fmt"
	//"github.com/rakyll/portmidi"
)

type OctaveOffset int

const (
	Octave0 OctaveOffset = 12
	Octave1 OctaveOffset = 24
	Octave2 OctaveOffset = 36
	Octave3 OctaveOffset = 48
	Octave4 OctaveOffset = 60
	Octave5 OctaveOffset = 72
	Octave6 OctaveOffset = 84
	Octave7 OctaveOffset = 96
	Octave8 OctaveOffset = 108
)

var octaveOffsets = []OctaveOffset{Octave0, Octave1, Octave2, Octave3, Octave4, Octave5, Octave6, Octave7, Octave8}

type MessageType int

/* Midi consts for message types */
const (
	Channel1  MessageType = 0x01
	Channel2  MessageType = 0x02
	Channel3  MessageType = 0x03
	Channel4  MessageType = 0x04
	Channel5  MessageType = 0x05
	Channel6  MessageType = 0x06
	Channel7  MessageType = 0x07
	Channel8  MessageType = 0x08
	Channel9  MessageType = 0x09
	Channel10 MessageType = 0x10
	Channel11 MessageType = 0x11
	Channel12 MessageType = 0x12
	Channel13 MessageType = 0x13
	Channel14 MessageType = 0x14
	Channel15 MessageType = 0x15

	NoteOn  MessageType = 0x90
	NoteOff MessageType = 0x80

	ControlChange MessageType = 2
)

type MidiMessage struct {
	Channel  MessageType
	Type     MessageType
	Note     int
	Octave   int
	Velocity int
}
type MidiControlMessage struct {
	Value int
}

type midiinfo struct {
	control <-chan MidiControlMessage
	input   <-chan MidiMessage
	Port    int
	//MIDIOutputStream *portmidi.Stream
}

func NewMidi(controlChannel <-chan MidiControlMessage, inputChannel <-chan MidiMessage) *midiinfo {

	midiProcessor := midiinfo{controlChannel, inputChannel, 2}
	//portmidi.Initialize()
	//out, err := portmidi.NewOutputStream(2, 1024, 0)

	//if err != nil {
	//	log.Fatal(err)
	//}

	//midiProcessor.MIDIOutputStream = out
	go midiProcessor.midiEmitThread()
	return &midiProcessor
}

func (midiEmitter *midiinfo) midiEmitThread() {
	for {
		message := <-midiEmitter.input

		fmt.Printf("MidiMessage: %v \n", message)
		fmt.Printf("MessageType + Channel: 0x%x\n", message.Type+message.Channel)
		fmt.Printf("Note: %d Octave: %d MIDINoteValue %d\n", message.Note, message.Octave, (int(octaveOffsets[message.Octave]) + (message.Note + 1)))
		/*midiEmitter.MIDIOutputStream.WriteShort(0x91, 60, 100)
		midiEmitter.MIDIOutputStream.WriteShort(0x91, 64, 100)
		midiEmitter.MIDIOutputStream.WriteShort(0x91, 67, 100)

		// notes will be sustained for 2 seconds
		time.Sleep(2 * time.Second)

		// note off events
		midiEmitter.MIDIOutputStream.WriteShort(0x81, 60, 100)
		midiEmitter.MIDIOutputStream.WriteShort(0x81, 64, 100)
		midiEmitter.MIDIOutputStream.WriteShort(0x81, 67, 100)
		*/
	}
}
