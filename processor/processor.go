package processor

import (
	"container/list"
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

/*Scales to add:
2 - 1 - 3 - 1 - 1 - 3 - 1 - 2 - 1 - 2 (Long Version)
Arabic:
2 - 2 - 1 - 1 - 2 - 2 - 2
Augmented:
3 - 1 - 3 - 1 - 3 - 1
Balinese:
 1 - 2 - 4 - 1 - 4
Byzantine:
 1 - 3 - 1 - 2 - 1 - 3 - 1
Chinese:
 4 - 2 - 1 - 4 - 1
Dimnished:
2 - 1 - 2 - 1 - 2 - 1 - 2 - 1
DominantDiminished:
1 - 2 - 1 - 2 - 1 - 2 - 1 - 2
Egyptian:
 2 - 3 - 2 - 3 - 2
Eight Tone Spanish:
1 - 2 - 1 - 1 - 1- 2 - 2 - 2
Geez (Ethiopian):
2 - 1 - 2 - 2 - 1 - 2 - 2
Hindu:
2 - 2 - 1 - 2 - 1 - 2 - 2
Hirajoshi:
1 - 4 - 1 - 4 - 2
Hungarian Gypsy:
2 - 1 - 3 - 1 - 1 - 3 - 1
Hungarian Major:
3 - 1 - 2 - 1 - 2 - 1 - 2
Japanese (in sen):
1 - 4 - 2 - 3 - 2
Lydian Dominant:
2 - 2 - 2 - 1 - 2 - 1 - 2
Neopolitan Minor:
1 - 2 - 2 - 2 - 1 - 3 - 1
Neopolitan Major:
1 - 2 - 2 - 2 - 2 - 2 - 1
Octatonic:
1- 2 - 1 - 2 - 1 - 2 - 1 - 2
Oriental:
1 - 3 - 1 - 1 - 3 - 1 - 2
Romanian Minor:
2 - 1 - 3 - 1 - 2 - 1 - 2
Spanish Gypsy:
1 - 3 - 1 - 2 - 1 - 2 - 2
Super Locrian:
1 - 2 - 1 - 2 - 2 - 2 - 2
*/
/* 3 octaves of Chromatic scale which allows for generation of any scale type with any root note */
var notes = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B", "C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B", "C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B", "C"}

//                    0    1     2    3     4    5    6     7    8     9    10    11   12
var chromaticOffsets = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

var ionianOffsets = []int{0, 2, 4, 5, 7, 9, 11, 12}  // Major
var locrianOffsets = []int{0, 1, 3, 5, 6, 8, 10, 12} // Minor
var dorianOffsets = []int{0, 2, 3, 5, 7, 9, 10, 12}
var phrygianOffsets = []int{0, 1, 3, 5, 7, 8, 10, 12}
var lydianOffsets = []int{0, 2, 4, 6, 7, 9, 11, 12}
var mixolydianOffsets = []int{0, 2, 4, 5, 7, 9, 10, 12}
var aeolianOffsets = []int{0, 2, 3, 5, 7, 8, 10, 12}
var japaneseYoOffsets = []int{0, 2, 5, 7, 9, 12}
var harmonicMinorOffsets = []int{0, 2, 3, 5, 7, 8, 11, 12}
var wholeToneOffsets = []int{0, 2, 4, 6, 8, 10, 12}
var algerianOffsets = []int{0, 2, 3, 5, 6, 7, 8, 11, 12}

//2 - 1 - 3 - 1 - 1 - 3 - 1 - 2 - 1 - 2
var algerianLongOffsets = []int{0, 2, 3, 6, 7, 8, 11, 12, 14, 15, 17}

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
	velocity  int
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
	Chromatic     scaleMap
	Ionian        scaleMap
	Dorian        scaleMap
	Phrygian      scaleMap
	Lydian        scaleMap
	Mixolydian    scaleMap
	Aeolian       scaleMap
	Locrian       scaleMap
	HarmonicMinor scaleMap
	JapaneseYo    scaleMap
	WholeTone     scaleMap
	Algerian      scaleMap
	AlgerianLong  scaleMap
}

type VelocityMode int

const (
	fixed              VelocityMode = 0
	singleNoteVariance VelocityMode = 1
)

const defaultVelocity = 80

const maxEvents = 15
const defaultBPM = 120
const defaultTick = 250

const maxPreviousValues = 20

/*ProcInfo Holds input/output info and generation parameters.*/
type ProcInfo struct {
	control             <-chan ControlMessage
	input               <-chan float64
	output              chan<- midioutput.MidiMessage
	BPM                 float64
	TickInc             time.Duration
	tick                float64
	scales              scaleTypes
	activeScale         scaleMap
	rootNoteOffset      int
	velocitySensingMode VelocityMode
	previousValues      *list.List
	largestVariance     float64
	events              []event
}

/*NewProcessor returns a new instance of the processor stack and starts the control/generation threads. */
func NewProcessor(controlChannel <-chan ControlMessage, inputChannel <-chan float64, outputChannel chan<- midioutput.MidiMessage) *ProcInfo {

	processor := ProcInfo{controlChannel, inputChannel, outputChannel, defaultBPM, defaultTick, 0, scaleTypes{}, scaleMap{}, 0, 0, list.New(), 0, []event{}}

	processor.initScaleTypes(A)
	processor.activeScale = processor.scales.Algerian
	processor.velocitySensingMode = fixed

	fmt.Printf("ActiveScale: %v+\n", processor.activeScale)
	processor.events = make([]event, maxEvents)

	go processor.controlThread()
	go processor.generationThread()

	return &processor
}

/*getNotes Given a root note and an array of offsets into the chromatic scale, this function returns an array of scale notes. */
func (processor *ProcInfo) getNotes(rootOffset rootNote, offsets []int) []string {
	fmt.Printf("RootOffset: %d \n", rootOffset)
	retNotes := make([]string, len(offsets))

	for i, offset := range offsets {
		retNotes[i] = notes[((int(rootOffset) + offset) % len(notes))]
	}

	return retNotes
}

/*initScaleTypes Initializes all scale types and offsets for a specific root note for later use. */
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

	processor.scales.JapaneseYo.notes = processor.getNotes(rootNoteIndex, japaneseYoOffsets)
	processor.scales.JapaneseYo.offsets = japaneseYoOffsets

	processor.scales.HarmonicMinor.notes = processor.getNotes(rootNoteIndex, harmonicMinorOffsets)
	processor.scales.HarmonicMinor.offsets = harmonicMinorOffsets

	processor.scales.WholeTone.notes = processor.getNotes(rootNoteIndex, wholeToneOffsets)
	processor.scales.WholeTone.offsets = wholeToneOffsets

	processor.scales.Algerian.notes = processor.getNotes(rootNoteIndex, algerianOffsets)
	processor.scales.Algerian.offsets = algerianOffsets

	processor.scales.AlgerianLong.notes = processor.getNotes(rootNoteIndex, algerianLongOffsets)
	processor.scales.AlgerianLong.offsets = algerianLongOffsets
	/* We need to store the root note offset so we can add it to the activeScale offset on when sending a note otherwise everything would be in C. */
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

/*getVelocity implements the logic for different types of velocity sensing based on input metrics:
  fixed					The input metrics have no effect on velocity and the default is used.
  singleNoteVariance	The largest variance seen so far is used to calculate the current variance as
						a percentage which is then used to control velocity.
*/
func (processor *ProcInfo) getVelocity(noteVal float64) int {
	switch processor.velocitySensingMode {
	case fixed:
		return defaultVelocity
	case singleNoteVariance:

		if processor.previousValues.Front() != nil && processor.previousValues.Len() > 1 {

			i := 0
			values := make([]float64, 2)

			for e := processor.previousValues.Front(); e != nil; e = e.Next() {
				if i < 2 {
					values[i] = e.Value.(float64)
				} else {
					break
				}
				i++
			}

			currentVariance := math.Abs(values[0] - values[1])

			if currentVariance > processor.largestVariance {
				processor.largestVariance = currentVariance
				return 100
			}

			velocity := (defaultVelocity + int((currentVariance/processor.largestVariance)*100))

			if velocity > 100 {
				return 100
			}

			return velocity

		}
		return defaultVelocity
	default:
		return 0
	}
}

/*controlThread listens for any incoming messages and handles them accordingly, updating parameters etc. */
func (processor *ProcInfo) controlThread() {

	for {
		message := <-processor.control
		fmt.Printf("TEST %f\n", message.Value)

	}
}

/*generationThread Handles event processing and timing of note emission acting like a sequencer for notes.*/
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

func (processor *ProcInfo) addToPreviousValues(value float64) {

	if processor.previousValues.Len() >= maxPreviousValues {
		processor.previousValues.Remove(processor.previousValues.Back())
	}
	processor.previousValues.PushFront(value)
}

/*processMessage Handles mapping metric value into note value. Also pushes event into sequencer. */
func (processor *ProcInfo) processMessage(value float64) {

	noteVal := int(value) % len(processor.activeScale.notes)
	event := event{note, ready, 4, processor.activeScale.offsets[noteVal], 3, processor.getVelocity(value)}

	processor.insertEvent(event)
	fmt.Printf("Note: %s Value: %f Index: %d Offset: %d\n", processor.activeScale.notes[noteVal], value, noteVal, processor.activeScale.offsets[noteVal])

	processor.addToPreviousValues(value)
}

/*
   handleEvents is used to trigger different kinds of events:

   If an event is in state ready, then it means it's new and we need to send a NoteOn message to the midi channel.
   If an event is in state active, then we decrement the duration by 1 and Send a NoteOff message if duration == 1.(Not 0, as this makes the No of beats more readable)
   if an event is in state stop, we deallocate the event entry so it can be used again.

   It needs two loops, otherwise if a previous note is the same as one being activated it might trigger a note off for the old note after the new note with the same value has been fired.
*/
func (processor *ProcInfo) handleEvents() {
	for i, e := range processor.events {
		if (event{}) != e {
			if e.state == active {
				processor.events[i].duration--

				if e.duration == 1 {

					fmt.Printf("Send stop %d Oct: %d \n", processor.rootNoteOffset+e.value, e.octave)

					processor.events[i].state = stop
					processor.output <- midioutput.MidiMessage{Channel: midioutput.Channel1, Type: midioutput.NoteOff, Note: processor.rootNoteOffset + processor.events[i].value, Octave: processor.events[i].octave, Velocity: 50}

				}
			}
		}
	}
	for i, e := range processor.events {

		if (event{}) != e {
			if e.state == ready {

				fmt.Printf("Send start %d Oct: %d \n", processor.rootNoteOffset+e.value, e.octave)

				processor.events[i].state = active
				processor.output <- midioutput.MidiMessage{Channel: midioutput.Channel1, Type: midioutput.NoteOn, Note: processor.rootNoteOffset + processor.events[i].value, Octave: processor.events[i].octave, Velocity: 80}

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

	//fmt.Printf("Tick: %f \n", processor.tick)
	processor.tick += float64(processor.TickInc)
	processor.tick = math.Mod(processor.tick, (60/processor.BPM)*1000)
	time.Sleep(processor.TickInc * time.Millisecond)

}
