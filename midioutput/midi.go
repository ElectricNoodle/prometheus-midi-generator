package midioutput

import (
	"fmt"
	//"github.com/rakyll/portmidi"
)

type octaveOffset int

/* */
const (
	Octave0 octaveOffset = 12
	Octave1 octaveOffset = 24
	Octave2 octaveOffset = 36
	Octave3 octaveOffset = 48
	Octave4 octaveOffset = 60
	Octave5 octaveOffset = 72
	Octave6 octaveOffset = 84
	Octave7 octaveOffset = 96
	Octave8 octaveOffset = 108
)

var octaveOffsets = []octaveOffset{Octave0, Octave1, Octave2, Octave3, Octave4, Octave5, Octave6, Octave7, Octave8}

/*MessageType Used to define consts for different MIDI events.*/
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

/*MidiMessage Hold all of the information required to build a MIDI message, recieved from processor.go*/
type MidiMessage struct {
	Channel  MessageType
	Type     MessageType
	Note     int
	Octave   int
	Velocity int
}

/*MidiControlMessage Used to store information on control messages recieved. */
type MidiControlMessage struct {
	Value int
}

/*MidiInfo Holds relevant info needed to recieve input/emit midi messages. */
type MidiInfo struct {
	control <-chan MidiControlMessage
	input   <-chan MidiMessage
	Port    int
	//MIDIOutputStream *portmidi.Stream
}

/*NewMidi Returns a new instance of midi struct, and inits midi connection. */
func NewMidi(controlChannel <-chan MidiControlMessage, inputChannel <-chan MidiMessage) *MidiInfo {

	midiProcessor := MidiInfo{controlChannel, inputChannel, 2}
	//portmidi.Initialize()
	//out, err := portmidi.NewOutputStream(2, 1024, 0)

	//if err != nil {
	//	log.Fatal(err)
	//}

	//midiProcessor.MIDIOutputStream = out
	go midiProcessor.midiEmitThread()
	return &midiProcessor
}

func (midiEmitter *MidiInfo) midiEmitThread() {
	for {
		message := <-midiEmitter.input

		fmt.Printf("MidiMessage: %v \n", message)
		fmt.Printf("MessageType + Channel: 0x%x\n", message.Type+message.Channel)
		fmt.Printf("Note: %d Octave: %d MIDINoteValue %d\n", message.Note, message.Octave, (int(octaveOffsets[message.Octave]) + message.Note))
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
