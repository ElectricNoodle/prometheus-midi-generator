package midi

import (
	"bytes"

	. "github.com/gomidi/midi/midimessage/channel"
	"github.com/gomidi/midi/midimessage/realtime"
	"github.com/gomidi/midi/midiwriter"
)

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

	midi := midi{controlChannel, inputChannel}
	var bf bytes.Buffer

	wr := midiwriter.New(&bf)

	wr.Write(Channel2.Pitchbend(5000))
	wr.Write(Channel2.NoteOn(65, 90))
	wr.Write(realtime.Reset)
	wr.Write(Channel2.NoteOff(65))

	return &midi
}
