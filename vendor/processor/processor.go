package processor

import (
	"container/list"
	"logging"
	"math"
	"midioutput"
	"time"

	"github.com/elliotchance/orderedmap"
)

var log *logging.Logger

/*noteIndexes Fixed values used to transpose scales into different keys */
var noteIndexes = map[string]int{
	"C":  0,
	"C#": 1,
	"D":  2,
	"D#": 3,
	"E":  4,
	"F":  5,
	"F#": 6,
	"G":  7,
	"G#": 8,
	"A":  9,
	"A#": 10,
	"B":  11,
}

/* 3 octaves of Chromatic scale which allows for generation of any scale type with any root note */
var notes = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B", "C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B", "C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B", "C"}

/*Scale Defines the format of a scale config */
type Scale struct {
	Name      string `yaml:"name"`
	Intervals []int  `yaml:"intervals,flow"`
}

/*Config Defines the format of the process */
type Config struct {
	Scales []Scale `yaml:"scales"`
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
	SetKey          MessageType = 0
	SetMode         MessageType = 1
	SetVelocityMode MessageType = 2
	SetBPM          MessageType = 3
)

/*ControlMessage Used for sending control messages to processor.*/
type ControlMessage struct {
	Type        MessageType
	ValueNum    int
	ValueString string
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
	Control             chan ControlMessage
	input               chan float64
	Output              chan midioutput.MIDIMessage
	BPM                 float64
	TickInc             time.Duration
	tick                float64
	scales              *orderedmap.OrderedMap
	activeScale         scaleMap
	rootNoteOffset      int
	velocitySensingMode velocityMode
	previousValues      *list.List
	maxVariance         float64
	events              []event
}

/*NewProcessor returns a new instance of the processor stack and starts the control/generation threads. */
func NewProcessor(logIn *logging.Logger, processorConfig Config, inputChannel chan float64) *ProcInfo {

	log = logIn
	processor := ProcInfo{make(chan ControlMessage, 6), inputChannel, make(chan midioutput.MIDIMessage, 6), defaultBPM,
		defaultTick, 0, orderedmap.NewOrderedMap(), scaleMap{},
		0, singleNoteVariance, list.New(), 0, make([]event, maxEvents)}

	processor.parseScales(processorConfig.Scales)
	processor.generateNotesOfScale(noteIndexes["C"])
	processor.setScale("Chromatic")

	go processor.controlThread()
	go processor.generationThread()

	return &processor

}

func (processor *ProcInfo) setScale(name string) {

	scale, exists := processor.scales.Get(name)

	if exists {

		processor.activeScale = scale.(scaleMap)

		log.Printf("Using %s scale in the key of %s.\n", processor.activeScale.name, notes[processor.rootNoteOffset])
		log.Printf("Notes: %v\n", processor.activeScale.notes)

	} else {
		log.Printf("Scale not found (%s).", name)
	}
}

/*parseScales Processes and stores the scales from the configuration file and generates note offset values for them. */
func (processor *ProcInfo) parseScales(scaleList []Scale) {

	for _, scale := range scaleList {

		var scaleMapping scaleMap

		scaleMapping.name = scale.Name
		scaleMapping.intervals = scale.Intervals
		scaleMapping.offsets = processor.getNoteOffsets(scaleMapping.intervals)

		processor.scales.Set(scale.Name, scaleMapping)
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
func (processor *ProcInfo) getNotes(rootOffset int, offsets []int) []string {

	retNotes := make([]string, len(offsets))

	for i, offset := range offsets {
		retNotes[i] = notes[((int(rootOffset) + offset) % len(notes))]
	}

	return retNotes
}

/*initScaleTypes Initializes all scale types and offsets for a specific root note for later use. */
func (processor *ProcInfo) generateNotesOfScale(rootNoteIndex int) {

	if rootNoteIndex <= len(noteIndexes) {

		for _, key := range processor.scales.Keys() {

			scale, exists := processor.scales.Get(key)

			if exists {

				castedScale := scale.(scaleMap)

				castedScale.notes = processor.getNotes(rootNoteIndex, castedScale.offsets)
				processor.scales.Set(key, castedScale)

			}
		}

		/* We need to store the root note offset so we can add it to the activeScale offset on when sending a note otherwise everything would be in C. */
		processor.rootNoteOffset = int(rootNoteIndex)

	} else {
		log.Printf("Invalid note index (%d) doing nothing. \n", rootNoteIndex)
	}
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
func (processor *ProcInfo) getMajorTriad(note int) {
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
func (processor *ProcInfo) getMinorTriad(note int) {

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
		message := <-processor.Control

		switch message.Type {

		case SetKey:

			processor.generateNotesOfScale(message.ValueNum)
			processor.setScale(processor.activeScale.name)

		case SetMode:
			processor.setScale(message.ValueString)

		case SetBPM:

		case SetVelocityMode:

		}
	}
}

/*GetKeyNames Returns an array of key names for the front end. */
func (processor *ProcInfo) GetKeyNames() []string {
	return notes[:12]
}

/*GetModeNames Returns an array of mode names for the front end. */
func (processor *ProcInfo) GetModeNames() []string {

	names := make([]string, processor.scales.Len())
	i := 0

	for _, key := range processor.scales.Keys() {
		names[i] = key.(string)
		i++
	}

	return names
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
	/*  */
	noteVal := int(value) % len(processor.activeScale.notes)
	event := event{note, ready, 4, processor.activeScale.offsets[noteVal], 3, processor.getVelocity(value)}

	processor.insertEvent(event)
	log.Printf("Note: %s Value: %f Index: %d Offset: %d\n", processor.activeScale.notes[noteVal], value, noteVal, processor.activeScale.offsets[noteVal])

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

					log.Printf("Send stop %d Oct: %d \n", processor.rootNoteOffset+e.value, e.octave)

					processor.events[i].state = stop
					processor.Output <- midioutput.MIDIMessage{Channel: midioutput.Channel1, Type: midioutput.NoteOff, Note: processor.rootNoteOffset + processor.events[i].value, Octave: processor.events[i].octave, Velocity: 50}

				}
			}
		}
	}
	for i, e := range processor.events {

		if (event{}) != e {
			if e.state == ready {

				log.Printf("Send start %d Oct: %d Vel: %d\n", processor.rootNoteOffset+e.value, e.octave, e.velocity)

				processor.events[i].state = active
				processor.Output <- midioutput.MIDIMessage{Channel: midioutput.Channel1, Type: midioutput.NoteOn, Note: processor.rootNoteOffset + processor.events[i].value, Octave: processor.events[i].octave, Velocity: e.velocity}
				break
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

	processor.tick += float64(processor.TickInc)
	processor.tick = math.Mod(processor.tick, (60/processor.BPM)*1000)
	time.Sleep(processor.TickInc * time.Millisecond)

}
