package logger

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type Logger struct {
	LogPath         string
	LogLineLock     *sync.Mutex
	LogLines        []string
	PrintToTerminal bool
}

func NewLogger(logPath string) *Logger {
	return &Logger{
		LogPath:     logPath,
		LogLineLock: &sync.Mutex{},
		LogLines:    make([]string, 0),
	}
}

// is goroutine that with a interval writes the log lines to a file
// and clears the log lines
func (l *Logger) StartLogger() {
	msg := fmt.Sprintf("\nLogger started, %s", time.Now().Format(time.RFC3339))
	l.LogLines = append(l.LogLines, msg)
	fmt.Println(msg)
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.LogLineLock.Lock()
			if len(l.LogLines) > 0 {
				length := 0
				length = len(l.LogLines)
				file, err := os.OpenFile(l.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Println("Error opening log file:", err)
					l.LogLineLock.Unlock()
					continue
				}
				defer file.Close()

				for _, line := range l.LogLines {
					if _, err := file.WriteString(line + "\n"); err != nil {
						fmt.Println("Error writing to log file:", err)
					}
				}
				l.LogLines = make([]string, 0)
				fmt.Println("Log lines written to file:", length)
			}
			l.LogLineLock.Unlock()
		}
	}

}

func (l *Logger) SetToPrintToTerminal() {
	l.PrintToTerminal = true
}

func (l *Logger) SetToNotPrintToTerminal() {
	l.PrintToTerminal = false
}

func (l *Logger) Log(args ...interface{}) {
	l.LogLineLock.Lock()
	defer l.LogLineLock.Unlock()
	msg := fmt.Sprintln(args...)
	l.LogLines = append(l.LogLines, "\t"+msg)

	if l.PrintToTerminal {
		fmt.Println(msg)
	}
}

func (l *Logger) Logf(format string, args ...interface{}) {
	l.LogLineLock.Lock()
	defer l.LogLineLock.Unlock()
	msg := fmt.Sprintf(format, args...)
	l.LogLines = append(l.LogLines, "\t"+msg)
	fmt.Println(msg)
}

func (l *Logger) LogError(msg string) {
	l.Log(msg)
}

func (l *Logger) Close() {
	l.Log("Logger closed")
	l.LogLineLock.Lock()
	defer l.LogLineLock.Unlock()
	file, err := os.OpenFile(l.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer file.Close()

	for _, line := range l.LogLines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			fmt.Println("Error writing to log file:", err)
		}
	}
	l.LogLines = make([]string, 0)
}
