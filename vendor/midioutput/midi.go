package midioutput

import (
	"fmt"
	"log"

	"github.com/rakyll/portmidi"
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
	Channel1  MessageType = 0x00
	Channel2  MessageType = 0x01
	Channel3  MessageType = 0x02
	Channel4  MessageType = 0x03
	Channel5  MessageType = 0x04
	Channel6  MessageType = 0x05
	Channel7  MessageType = 0x06
	Channel8  MessageType = 0x07
	Channel9  MessageType = 0x08
	Channel10 MessageType = 0x09
	Channel11 MessageType = 0x10
	Channel12 MessageType = 0x11
	Channel13 MessageType = 0x12
	Channel14 MessageType = 0x13
	Channel15 MessageType = 0x14
	Channel16 MessageType = 0x15

	NoteOn  MessageType = 0x90
	NoteOff MessageType = 0x80

	//ControlChange MessageType = 2
)

/*MidiMessage Hold all of the information required to build a MIDI message, recieved from processor.go*/
type MidiMessage struct {
	Channel  MessageType
	Type     MessageType
	Note     int
	Octave   int
	Velocity int64
}

/*MidiControlMessage Used to store information on control messages recieved. */
type MidiControlMessage struct {
	Value int
}

/*MidiInfo Holds relevant info needed to recieve input/emit midi messages. */
type MidiInfo struct {
	control          <-chan MidiControlMessage
	input            <-chan MidiMessage
	port             int
	midiOutputStream *portmidi.Stream
}

/*NewMidi Returns a new instance of midi struct, and inits midi connection. */
func NewMidi(controlChannel <-chan MidiControlMessage, inputChannel <-chan MidiMessage) *MidiInfo {

	midiProcessor := MidiInfo{controlChannel, inputChannel, 2, nil}
	portmidi.Initialize()
	count := portmidi.CountDevices()
	fmt.Printf("Count: %d\n", count)
	out, err := portmidi.NewOutputStream(2, 1024, 0)

	if err != nil {
		log.Fatal(err)

	}

	midiProcessor.midiOutputStream = out
	go midiProcessor.midiEmitThread()
	return &midiProcessor
}

func (midiEmitter *MidiInfo) midiEmitThread() {
	for {
		message := <-midiEmitter.input
		fmt.Printf("Type: 0x%x MiDINote: Not+Oct:%d Note:%d\n", int64(message.Type+message.Channel), int64(int(octaveOffsets[message.Octave])+message.Note), int(message.Note))
		midiEmitter.midiOutputStream.WriteShort(int64(message.Type+message.Channel), int64(int(octaveOffsets[message.Octave])+message.Note), message.Velocity)
	}
}
