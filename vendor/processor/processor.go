package processor

import (
	"container/list"
	"fmt"
	"math"
	"midioutput"
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

/*Scale Defines the format of a scale config */
type Scale struct {
	Name      string `yaml:"name"`
	Intervals []int  `yaml:"intervals,flow"`
}

/*Config Defines the format of the process */
type Config struct {
	DefaultKey   string  `yaml:"default_key"`
	DefaultScale string  `yaml:"default_scale"`
	Scales       []Scale `yaml:"scales"`
}

type eventType int

const (
	note      eventType = 0
	parameter eventType = 1
)

/*eventState Defines all of the states a sequencer event can be in */
type eventState int

const (
	ready  eventState = 0
	active eventState = 1
	stop   eventState = 2
)

/*event Stores information needed to send different types of MIDI Message. */
type event struct {
	eventType eventType
	state     eventState
	duration  int
	value     int
	octave    int
	velocity  int64
}

/*MessageType Defines the different types of Control Message.*/
type MessageType int

/* Used to nicely assign values to message types */
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

/*scaleMap Used for storing all useful information of a scale. */
type scaleMap struct {
	name      string
	notes     []string
	offsets   []int
	intervals []int
}

type velocityMode int

const (
	fixed              velocityMode = 0
	singleNoteVariance velocityMode = 1
)

const maxVelocity = 127
const defaultVelocity = 0

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
	scales              map[string]scaleMap
	activeScale         scaleMap
	rootNoteOffset      int
	velocitySensingMode velocityMode
	previousValues      *list.List
	maxVariance         float64
	events              []event
}

/*NewProcessor returns a new instance of the processor stack and starts the control/generation threads. */
func NewProcessor(processorConfig Config, controlChannel <-chan ControlMessage, inputChannel <-chan float64, outputChannel chan<- midioutput.MidiMessage) *ProcInfo {

	processor := ProcInfo{controlChannel, inputChannel, outputChannel, defaultBPM, defaultTick, 0, make(map[string]scaleMap), scaleMap{}, 0, 0, list.New(), 0, []event{}}

	processor.parseScales(processorConfig.Scales)
	processor.generateNotesOfScale(A)

	processor.activeScale = processor.scales[processorConfig.DefaultScale]
	processor.velocitySensingMode = singleNoteVariance //fixed

	fmt.Printf("Active Scale: %v+\n", processor.activeScale)
	processor.events = make([]event, maxEvents)

	go processor.controlThread()
	go processor.generationThread()

	return &processor

}

/*parseScales Processes and stores the scales from the configuration file and generates note offset values for them. */
func (processor *ProcInfo) parseScales(scaleList []Scale) {

	for _, scale := range scaleList {

		var scaleMapping scaleMap

		scaleMapping.name = scale.Name
		scaleMapping.intervals = scale.Intervals
		scaleMapping.offsets = processor.getNoteOffsets(scaleMapping.intervals)

		processor.scales[scale.Name] = scaleMapping
	}
}

/*getNoteOffsets Generates the note offset values for the specified array of intervals */
func (processor *ProcInfo) getNoteOffsets(intervals []int) []int {

	offsets := make([]int, len(intervals)+1)
	offsets[0] = 0

	for i, interval := range intervals {
		offsets[i+1] = offsets[i] + interval
	}

	return offsets
}

/*getNotes Given a root note and an array of offsets into the chromatic scale, this function returns an array of scale notes. */
func (processor *ProcInfo) getNotes(rootOffset rootNote, offsets []int) []string {

	retNotes := make([]string, len(offsets))

	for i, offset := range offsets {
		retNotes[i] = notes[((int(rootOffset) + offset) % len(notes))]
	}

	return retNotes
}

/*initScaleTypes Initializes all scale types and offsets for a specific root note for later use. */
func (processor *ProcInfo) generateNotesOfScale(rootNoteIndex rootNote) {

	for key, scale := range processor.scales {

		scale.notes = processor.getNotes(rootNoteIndex, scale.offsets)
		processor.scales[key] = scale

		fmt.Printf("Scale Name: %s\n Intervals:\t %v\n Offsets:\t %v\n Notes:\t\t %v \n", scale.name, scale.intervals, scale.offsets, scale.notes)
	}

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
  singleNoteVariance	The max variance between the current value and the last is tracked over time
						and used to calculate what percentage the most recent variance is of that.
						This is then added to the defaultVelocity.
						NOTE: Should maybe try changing it so it calcs that as a percentage of change
						of the total velocity range to see how it sounds.
*/
func (processor *ProcInfo) getVelocity(value float64) int64 {
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

			if currentVariance > processor.maxVariance {
				processor.maxVariance = currentVariance
				return maxVelocity
			}

			velocity := (defaultVelocity + int64((currentVariance/processor.maxVariance)*100))

			if velocity > maxVelocity {
				return maxVelocity
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
		fmt.Printf("Received Message: %v\n", message.Value)

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

/*addToPreviousValues  */
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

				fmt.Printf("Send start %d Oct: %d Vel: %d\n", processor.rootNoteOffset+e.value, e.octave, e.velocity)

				processor.events[i].state = active
				processor.output <- midioutput.MidiMessage{Channel: midioutput.Channel1, Type: midioutput.NoteOn, Note: processor.rootNoteOffset + processor.events[i].value, Octave: processor.events[i].octave, Velocity: e.velocity}

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
