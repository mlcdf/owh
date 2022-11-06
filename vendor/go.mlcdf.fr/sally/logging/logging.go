// logging provides leveled log methods for the default logger
package logging // import "go.mlcdf.fr/sally/logging"

import (
	"fmt"
	"log"
	"sync/atomic"
)

type Level int32 // sync/atomic int32

const (
	TRACE Level = iota * 10
	DEBUG
	INFO
	WARNING
	ERROR
)

var level Level

// SetLevel sets a min level for the default logger (log.std)
func SetLevel(l Level) {
	atomic.StoreInt32((*int32)(&level), int32(l))
}

// Debugf prints to the default logger
func Debugf(format string, v ...interface{}) {
	if level >= DEBUG {
		log.Output(3, fmt.Sprintf(format+"\n", v...))
	}
}

// Infof prints to the default logger
func Infof(format string, v ...interface{}) {
	if level >= INFO {
		log.Output(3, fmt.Sprintf(format+"\n", v...))
	}
}

// Warningf prints to the default logger
func Warningf(format string, v ...interface{}) {
	if level >= WARNING {
		log.Output(3, fmt.Sprintf(format+"\n", v...))
	}
}

// Errorf prints to the default logger
func Errorf(format string, v ...interface{}) {
	if level >= ERROR {
		log.Output(3, fmt.Sprintf(format+"\n", v...))
	}
}
