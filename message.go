package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"log"
	"os"
	"sync"
	"time"
)

const OutputLogFileName = "output.log" // File name for storing output logs

type MessageManager struct {
	mu   sync.Mutex // Mutex for sent message synchronization
	file *os.File
}

func NewMessageManager() *MessageManager {
	m := &MessageManager{}
	// init initializes the output log file and handles potential errors.
	var err error
	if m.file, err = os.OpenFile(OutputLogFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		panic(err)
	}
	return m
}

// Send sends a notification and logs the message.
func (m *MessageManager) Send(format string, a ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	log.Printf(format+"\n", a...)
	fyne.CurrentApp().SendNotification(&fyne.Notification{
		Title:   "Dots-And-Boxes",
		Content: fmt.Sprintf(format, a...),
	})
	if _, err := OutputLogFile.WriteString(time.Now().Format(time.DateTime) + " " + fmt.Sprintf(format, a...) + "\n"); err != nil {
		log.Println(err)
		return
	}
}
