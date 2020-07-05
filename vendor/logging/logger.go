package logging

import (
	"fmt"
	"log"
	"os"
)

/*Logger Stores  */
type Logger struct {
	Channel chan string
}

/*NewLogger Makes a new logger instance and initializes channel. */
func NewLogger() *Logger {

	return &Logger{make(chan string, 600)}
}

/*Printf Prints log message in the correct format and tries to send it on logging channel for GUI.*/
func (logger *Logger) Printf(format string, a ...interface{}) {

	str := fmt.Sprintf(format, a...)
	log.Printf("%s", str)
	logger.send(str)
}

/*Println Prints the log message in the correct format and tries to send it on logging channel for GUI.*/
func (logger *Logger) Println(i ...interface{}) {

	str := fmt.Sprintln(i...)
	log.Printf("%s\n", str)
	logger.send(str)
}

/*Fatalf Prints the log message sends to channel, then exits. */
func (logger *Logger) Fatalf(format string, a ...interface{}) {

	str := fmt.Sprintf(format, a...)
	log.Printf("%s", str)
	logger.send(str)
	os.Exit(1)

}

/*Fatal Prints the log message sends to channel, then exits. */
func (logger *Logger) Fatal(a ...interface{}) {

	str := fmt.Sprint(a...)
	log.Printf("%s", str)
	logger.send(str)
	os.Exit(1)

}

func (logger *Logger) send(str string) {

	select {
	case logger.Channel <- str:

	default:
		log.Println("Error sending to channel.")
	}

}
