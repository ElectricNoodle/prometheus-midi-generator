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

/*MIDIMessage Hold all of the information required to build a MIDI message, recieved from processor.go*/
type MIDIMessage struct {
	Channel  MessageType
	Type     MessageType
	Note     int
	Octave   int
	Velocity int64
}

/*MIDIControlMessage Used to store information on control messages recieved. */
type MIDIControlMessage struct {
	Value int
}

/*MIDIEmitter Holds relevant info needed to recieve input/emit midi messages. */
type MIDIEmitter struct {
	control        <-chan MIDIControlMessage
	input          <-chan MIDIMessage
	DeviceInfo     []*portmidi.DeviceInfo
	configuredPort int
	deviceCount    int
	midiOutput     *portmidi.Stream
}

var maxDevices = 10

/*NewMidi Returns a new instance of midi struct, and inits midi connection. */
func NewMidi(controlChannel <-chan MIDIControlMessage, inputChannel <-chan MIDIMessage) *MIDIEmitter {

	midiEmitter := MIDIEmitter{controlChannel, inputChannel, []*portmidi.DeviceInfo{}, 0, 0, nil}

	midiEmitter.initMIDI()

	portmidi.Initialize()
	count := portmidi.CountDevices()

	fmt.Printf("Count: %d\n", count)
	fmt.Printf("Default ID: %v\n", portmidi.Info(0))

	out, err := portmidi.NewOutputStream(2, 1024, 0)

	if err != nil {
		log.Fatal(err)

	}

	midiEmitter.midiOutput = out

	go midiEmitter.midiEmitThread()
	return &midiEmitter
}

/*initMIDI Initializes the portmidi library and stores device info. */
func (midiEmitter *MIDIEmitter) initMIDI() {

	err := portmidi.Initialize()

	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < portmidi.CountDevices(); i++ {

		device := portmidi.Info(portmidi.DeviceID(i))

		if device.IsOutputAvailable == true && device.IsOpened == false {

			midiEmitter.DeviceInfo = append(midiEmitter.DeviceInfo, device)
			midiEmitter.deviceCount++
		}
	}

	if midiEmitter.deviceCount < 1 {
		log.Println("Error no MIDI devices available")
	}
}

/*GetDeviceNames returns an array of midi device names. */
func (midiEmitter *MIDIEmitter) GetDeviceNames() []string {

	names := make([]string, midiEmitter.deviceCount)

	if midiEmitter.deviceCount < 1 {
		return []string{"No devices found."}
	}

	for i := 0; i < midiEmitter.deviceCount; i++ {
		names[i] = midiEmitter.DeviceInfo[i].Name
	}

	return names
}

func (midiEmitter *MIDIEmitter) midiEmitThread() {
	for {
		message := <-midiEmitter.input
		fmt.Printf("Type: 0x%x MiDINote: Not+Oct:%d Note:%d\n", int64(message.Type+message.Channel), int64(int(octaveOffsets[message.Octave])+message.Note), int(message.Note))
		midiEmitter.midiOutput.WriteShort(int64(message.Type+message.Channel), int64(int(octaveOffsets[message.Octave])+message.Note), message.Velocity)
	}
}
