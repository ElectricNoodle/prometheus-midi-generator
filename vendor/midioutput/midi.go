package midioutput

import (
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

/*MIDIControlMessage Used to store information on control messages recieved. */
type MIDIControlMessage struct {
	Type  MessageType
	Value int
}

type midiDevice struct {
	id   int
	info *portmidi.DeviceInfo
}

/*MIDIEmitter Holds relevant info needed to recieve input/emit midi messages. */
type MIDIEmitter struct {
	Control        chan MIDIControlMessage
	input          <-chan MIDIMessage
	MIDIDevices    []midiDevice
	configuredPort int
	deviceCount    int
	midiOutput     *portmidi.Stream
}

var maxDevices = 10

/*NewMidi Returns a new instance of midi struct, and inits midi connection. */
func NewMidi(controlChannel chan MIDIControlMessage, inputChannel <-chan MIDIMessage) *MIDIEmitter {

	midiEmitter := MIDIEmitter{controlChannel, inputChannel, []midiDevice{}, 0, 0, nil}

	midiEmitter.initMIDI()

	go midiEmitter.controlThread()
	go midiEmitter.emitThread()
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

			midiEmitter.MIDIDevices = append(midiEmitter.MIDIDevices, midiDevice{id: i, info: device})
			midiEmitter.deviceCount++
		}
	}

	if midiEmitter.deviceCount < 1 {
		log.Println("Error no MIDI devices available")
		return
	}

	midiEmitter.setDevice(0)
}

/*GetDeviceNames returns an array of midi device names. */
func (midiEmitter *MIDIEmitter) GetDeviceNames() []string {

	names := make([]string, midiEmitter.deviceCount)

	if midiEmitter.deviceCount < 1 {
		return []string{"No devices found."}
	}

	for i := 0; i < midiEmitter.deviceCount; i++ {
		names[i] = midiEmitter.MIDIDevices[i].info.Name
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

func (midiEmitter *MIDIEmitter) setDevice(id int) {

	if midiEmitter.midiOutput != nil {

		err := midiEmitter.midiOutput.Close()

		if err != nil {
			log.Printf("Error: %s", err)
			return
		}

	}

	if midiEmitter.MIDIDevices[id] != (midiDevice{}) {

		if midiEmitter.MIDIDevices[id].info != nil {
			out, err := portmidi.NewOutputStream(portmidi.DeviceID(midiEmitter.MIDIDevices[id].id), 1024, 0)

			if err != nil {
				log.Printf("Error: %s", err)
				return
			}

			midiEmitter.midiOutput = out
			log.Printf("Switched to %s.\n", midiEmitter.MIDIDevices[id].info.Name)
		}
	}

}

func (midiEmitter *MIDIEmitter) emitThread() {
	for {
		message := <-midiEmitter.input
		//fmt.Printf("Type: 0x%x MiDINote: Not+Oct:%d Note:%d\n", int64(message.Type+message.Channel), int64(int(octaveOffsets[message.Octave])+message.Note), int(message.Note))
		if midiEmitter.midiOutput != nil {
			midiEmitter.midiOutput.WriteShort(int64(message.Type+message.Channel), int64(int(octaveOffsets[message.Octave])+message.Note), message.Velocity)
		} else {
			log.Println("No MIDI Device configured.")
		}
	}
}
