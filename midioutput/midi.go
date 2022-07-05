package midioutput

import (
	"github.com/ElectricNoodle/prometheus-midi-generator/logging"
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/portmididrv" // autoregisters driver
)

var log *logging.Logger

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

/*MIDIValue Used to define consts for different MIDI events.*/
type MIDIValue int

/* Midi consts for message types */
const (
	Channel1  MIDIValue = 0x00
	Channel2  MIDIValue = 0x01
	Channel3  MIDIValue = 0x02
	Channel4  MIDIValue = 0x03
	Channel5  MIDIValue = 0x04
	Channel6  MIDIValue = 0x05
	Channel7  MIDIValue = 0x06
	Channel8  MIDIValue = 0x07
	Channel9  MIDIValue = 0x08
	Channel10 MIDIValue = 0x09
	Channel11 MIDIValue = 0x10
	Channel12 MIDIValue = 0x11
	Channel13 MIDIValue = 0x12
	Channel14 MIDIValue = 0x13
	Channel15 MIDIValue = 0x14
	Channel16 MIDIValue = 0x15

	NoteOn  MIDIValue = 0x90
	NoteOff MIDIValue = 0x80
)

/*MIDIMessage Hold all of the information required to build a MIDI message, recieved from processor.go*/
type MIDIMessage struct {
	Channel  MIDIValue
	Type     MIDIValue
	Note     int
	Octave   int
	Velocity int64
}

/*MessageType Defines type of control message.*/
type MessageType int

/*const values for message types */
const (
	SetDevice MessageType = 0
)

/*ControlMessage Used to store information on control messages recieved. */
type ControlMessage struct {
	Type  MessageType
	Value string
}

/*MIDIEmitter Holds relevant info needed to recieve input/emit midi messages. */
type MIDIEmitter struct {
	Control            chan ControlMessage
	input              <-chan MIDIMessage
	selectedMIDIDevice string
	deviceCount        int
	midiOutput         int
}

var maxDevices = 10

/*NewMidi Returns a new instance of midi struct, and inits midi connection. */
func NewMidi(logIn *logging.Logger, inputChannel <-chan MIDIMessage) *MIDIEmitter {

	log = logIn
	midiEmitter := MIDIEmitter{make(chan ControlMessage, 6), inputChannel, "USB MIDI", 0, -1}

	go midiEmitter.controlThread()
	go midiEmitter.emitThread()

	return &midiEmitter
}

/*GetDeviceNames returns an array of midi device names. */
func (midiEmitter *MIDIEmitter) GetDeviceNames() []string {

	names := midi.OutPorts()
	numDevices := len(names)

	if numDevices < 1 {
		return []string{"No devices found."}
	}

	return names
}

func (midiEmitter *MIDIEmitter) controlThread() {
	for {

		message := <-midiEmitter.Control

		switch message.Type {

		case SetDevice:
			midiEmitter.setDevice(message.Value)

		}
	}
}

func (midiEmitter *MIDIEmitter) setDevice(name string) {

	midiEmitter.selectedMIDIDevice = name
	log.Printf("Midi Device set to %v\n", name)
}

func (midiEmitter *MIDIEmitter) emitThread() {

	out := midi.FindOutPort("USB Midi")
	sendMessage, err := midi.SendTo(out)

	if err != nil {
		log.Printf("Failed to send midi message. (%v)\n", err)
	}

	for {

		message := <-midiEmitter.input

		if sendMessage != nil {

			var midiMessage midi.Message

			if message.Type == NoteOn {
				//	log.Printf("Type: 0x%x MiDINote: Not+Oct:%d Note:%d\n", int64(message.Type+message.Channel), int64(int(octaveOffsets[message.Octave])+message.Note), int(message.Note))
				midiMessage = midi.NoteOn(1, uint8(int(octaveOffsets[message.Octave])+message.Note), uint8(message.Velocity))

			} else if message.Type == NoteOff {
				//	log.Printf("Type: 0x%x MiDINote: Not+Oct:%d Note:%d\n", int64(message.Type+message.Channel), int64(int(octaveOffsets[message.Octave])+message.Note), int(message.Note))
				midiMessage = midi.NoteOff(1, uint8(int(octaveOffsets[message.Octave])+message.Note))
			}

			err := sendMessage(midiMessage)

			if err != nil {
				log.Printf("Failed to send midi message. (%v)\n", err)
			}

		} else {
			log.Println("No MIDI Device configured.")
		}
		/*miiEmitter.midiOutput.WriteShort(0x80, 60, 100)

		if midiEmitter.midiOutput != nil {

			midiEmitter.midiOutput.WriteShort(int64(message.Type+message.Channel), int64(int(octaveOffsets[message.Octave])+message.Note), message.Velocity)

		} else {
		}
		*/
	}
}
