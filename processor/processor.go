package processor

import (
	"fmt"
	"math"
	"prometheus-midi-generator/midioutput"
	"time"
)

type rootNote int

/* Const indexes for note values at positions so code is more readable. NOTE: Removed CLow/CHigh concept, nnot sure if it's needed here. */
const (
	C      rootNote = 0
	CSharp rootNote = 1
	D      rootNote = 2
	DSharp rootNote = 3
	E      rootNote = 4
	F      rootNote = 5
	FSharp rootNote = 6
	G      rootNote = 7
	GSharp rootNote = 8
	A      rootNote = 9
	ASharp rootNote = 10
	B      rootNote = 11
)

/* 3 octaves of Chromatic scale which allows for generation of any scale type with any root note */
var notes = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B", "C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B", "C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B", "C"}

var chromaticOffsets = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

/* C D E F G A B C*/
var ionianOffsets = []int{0, 2, 4, 5, 7, 9, 11, 12}
var dorianOffsets = []int{0, 2, 3, 5, 7, 9, 10, 12}
var phrygianOffsets = []int{0, 1, 3, 5, 7, 8, 10, 12}
var lydianOffsets = []int{0, 2, 4, 6, 7, 9, 11, 12}
var mixolydianOffsets = []int{0, 2, 4, 5, 7, 9, 10, 12}
var aeolianOffsets = []int{0, 2, 3, 5, 7, 8, 10, 12}
var locrianOffsets = []int{0, 1, 3, 5, 6, 8, 10, 12}

type eventType int

const (
	note      eventType = 0
	parameter eventType = 1
)

type eventState int

const (
	ready  eventState = 0
	active eventState = 1
	stop   eventState = 2
)

type event struct {
	eventType eventType
	state     eventState
	duration  int
	value     int
	octave    int
}

/*MessageType Defines the different type of Control Message.*/
type MessageType int

/* Values for MessageType */
const (
	StartOutput      MessageType = 0
	StopOutput       MessageType = 1
	ChangePollRate   MessageType = 2
	ChangeOutputRate MessageType = 3
)

/*ControlMessage Used for sending control messages to processor.*/
type ControlMessage struct {
	Type  MessageType
	Value float64
}
type scaleMap struct {
	notes   []string
	offsets []int
}
type scaleTypes struct {
	Chromatic  scaleMap
	Ionian     scaleMap
	Dorian     scaleMap
	Phrygian   scaleMap
	Lydian     scaleMap
	Mixolydian scaleMap
	Aeolian    scaleMap
	Locrian    scaleMap
}

const maxEvents = 10
const defaultBPM = 60
const defaultTick = 250

/*ProcInfo Holds input/output info and generation parameters.*/
type ProcInfo struct {
	control        <-chan ControlMessage
	input          <-chan float64
	output         chan<- midioutput.MidiMessage
	BPM            float64
	TickInc        time.Duration
	tick           float64
	scales         scaleTypes
	activeScale    scaleMap
	rootNoteOffset int
	events         []event
}

/*NewProcessor returns a new instance of the processor stack and starts the control/generation threads. */
func NewProcessor(controlChannel <-chan ControlMessage, inputChannel <-chan float64, outputChannel chan<- midioutput.MidiMessage) *ProcInfo {

	processor := ProcInfo{controlChannel, inputChannel, outputChannel, defaultBPM, defaultTick, 0, scaleTypes{}, scaleMap{}, 0, []event{}}
	// TODO: BUG: Need to think about how to store original index value of note into activeScale along with the string of note.
	// Otherwise note choice will always be wrong since it uses the index along with an octave offset to generate midi note.
	processor.initScaleTypes(F)
	processor.activeScale = processor.scales.Phrygian

	fmt.Printf("ActiveScale: %v+\n", processor.activeScale)
	processor.events = make([]event, maxEvents)

	go processor.controlThread()
	go processor.generationThread()

	return &processor
}
func (processor *ProcInfo) setActiveScale(scale scaleTypes) {

}
func (processor *ProcInfo) getNotes(rootOffset rootNote, offsets []int) []string {
	fmt.Printf("RootOffset: %d \n", rootOffset)
	retNotes := make([]string, len(offsets))

	for i, offset := range offsets {
		retNotes[i] = notes[((int(rootOffset) + offset) % len(notes))]
	}

	return retNotes
}

func (processor *ProcInfo) initScaleTypes(rootNoteIndex rootNote) {

	processor.scales.Chromatic.notes = make([]string, len(notes))
	processor.scales.Chromatic.notes = notes
	processor.scales.Chromatic.offsets = chromaticOffsets

	processor.scales.Ionian.notes = processor.getNotes(rootNoteIndex, ionianOffsets)
	processor.scales.Ionian.offsets = ionianOffsets

	processor.scales.Dorian.notes = processor.getNotes(rootNoteIndex, dorianOffsets)
	processor.scales.Dorian.offsets = dorianOffsets

	processor.scales.Phrygian.notes = processor.getNotes(rootNoteIndex, phrygianOffsets)
	processor.scales.Phrygian.offsets = phrygianOffsets

	processor.scales.Mixolydian.notes = processor.getNotes(rootNoteIndex, mixolydianOffsets)
	processor.scales.Mixolydian.offsets = mixolydianOffsets

	processor.scales.Lydian.notes = processor.getNotes(rootNoteIndex, lydianOffsets)
	processor.scales.Lydian.offsets = lydianOffsets

	processor.scales.Aeolian.notes = processor.getNotes(rootNoteIndex, aeolianOffsets)
	processor.scales.Aeolian.offsets = aeolianOffsets

	processor.scales.Locrian.notes = processor.getNotes(rootNoteIndex, locrianOffsets)
	processor.scales.Locrian.offsets = locrianOffsets

	processor.rootNoteOffset = int(rootNoteIndex)

}

/*
	note_index = allnotes.index(note)
	maj_third_i = note_index + 4
	major_third = allnotes[maj_third_i]
	min_fifth_i = maj_third_i + 3
	minor_fifth = allnotes[min_fifth_i]
	majortriad = (note,major_third,minor_fifth)
	return majortriad
*/
func (processor *ProcInfo) getMajorTriad(note rootNote) {
	//index := int(note)
	//majorThirdIndex := index + 4
	//	minorFifthIndex := majorThirdIndex + 3

}

/*
	note_index = allnotes.index(note)
	min_third_i = note_index + 3
	minor_third = allnotes[min_third_i]
	maj_fifth_i = min_third_i + 4
	major_fifth = allnotes[maj_fifth_i]
	minortriad = (note,minor_third,major_fifth)
	return minortriad
*/
func (processor *ProcInfo) getMinorTriad(note rootNote) {

}

/* This function listens for any incoming messages and handles them accordingly */
func (processor *ProcInfo) controlThread() {

	for {
		message := <-processor.control
		fmt.Printf("TEST %f\n", message.Value)

	}
}

func (processor *ProcInfo) generationThread() {

	processor.tick = 0

	for {

		select {

		case message := <-processor.input:

			processor.processMessage(message)

		default:

			if processor.tick == 0 {

				processor.handleEvents()

			}

			processor.incrementTick()

		}

	}
}

func (processor *ProcInfo) processMessage(value float64) {

	noteVal := int(value) % len(processor.activeScale.notes)
	event := event{note, ready, 4, processor.activeScale.offsets[noteVal], 3}
	processor.insertEvent(event)
	fmt.Printf("Note: %s Value: %f Index: %d Offset: %d\n", processor.activeScale.notes[noteVal], value, noteVal, processor.activeScale.offsets[noteVal])

}

/*
   handleEvents is used to trigger different kinds of events,
   If an event is in state ready, then it means it's new and we need to send a NoteOn message to the midi channel.
   If an event is in state active, then we decrement the duration by 1 and Send a NoteOff message if duration == 1.(Not 0, as this makes the No of beats more readable)
   if an event is in state stop, we deallocate the event entry so it can be used again.
*/
func (processor *ProcInfo) handleEvents() {

	for i, e := range processor.events {

		if (event{}) != e {

			if e.state == ready {

				fmt.Printf("Send start %d Oct: %d \n", processor.rootNoteOffset+e.value, e.octave)

				processor.events[i].state = active

				processor.output <- midioutput.MidiMessage{midioutput.Channel1, midioutput.NoteOn, processor.rootNoteOffset + processor.events[i].value, processor.events[i].octave, 80}

			} else if e.state == active {

				processor.events[i].duration--

				if e.duration == 1 {

					fmt.Printf("Send stop %d Oct: %d \n", processor.rootNoteOffset+e.value, e.octave)

					processor.events[i].state = stop

					processor.output <- midioutput.MidiMessage{midioutput.Channel1, midioutput.NoteOff, processor.rootNoteOffset + processor.events[i].value, processor.events[i].octave, 50}

				}

			} else if e.state == stop {
				processor.events[i] = event{}
			}
		}
	}
}

func (processor *ProcInfo) insertEvent(eventIn event) {
	for i, e := range processor.events {
		if (event{}) == e {
			processor.events[i] = eventIn
			break
		}
	}

}

func (processor *ProcInfo) incrementTick() {

	fmt.Printf("Tick: %f \n", processor.tick)
	processor.tick += float64(processor.TickInc)
	processor.tick = math.Mod(processor.tick, (60/processor.BPM)*1000)
	time.Sleep(processor.TickInc * time.Millisecond)

}
