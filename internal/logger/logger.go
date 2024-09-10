// Copyright 2024 Aleksei Grigorev
// https://aleksvgrig.com, https://github.com/AlekseiGrigorev, aleksvgrig@gmail.com.
// Package define interfaces, structures and functions for working with logs
package logger

import (
	"fmt"
	"log"
	"slices"
	"time"
)

// Define log instance
type Log struct {
	logger          log.Logger // Log instance
	PrefixDelimiter string     // Prefix delimiter
	PrintToStdout   bool       // Print to stdout flag
}

// Returns Log instance
func (l *Log) Log() *log.Logger {
	return &l.logger
}

// Print to stdout
func (l *Log) printToStdoutInternal(params ...any) {
	fmt.Println(params...)
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

// Format header for print log to stdout
func (l *Log) formatHeader() string {
	flag := l.logger.Flags()
	if flag&(log.Ldate|log.Ltime|log.Lmicroseconds) != 0 {
		buf := []byte{}
		t := time.Now()
		if flag&log.LUTC != 0 {
			t = t.UTC()
		}
		if flag&log.Ldate != 0 {
			year, month, day := t.Date()
			itoa(&buf, year, 4)
			buf = append(buf, '/')
			itoa(&buf, int(month), 2)
			buf = append(buf, '/')
			itoa(&buf, day, 2)
			buf = append(buf, ' ')
		}
		if flag&(log.Ltime|log.Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(&buf, hour, 2)
			buf = append(buf, ':')
			itoa(&buf, min, 2)
			buf = append(buf, ':')
			itoa(&buf, sec, 2)
			if flag&log.Lmicroseconds != 0 {
				buf = append(buf, '.')
				itoa(&buf, t.Nanosecond()/1e3, 6)
			}
			buf = append(buf, ' ')
		}
		return string(buf)
	}
	return ""
}

// Print error message
func (log *Log) Error(params ...any) *Log {
	prefix := "ERROR:" + log.PrefixDelimiter
	log.Log().SetPrefix(prefix)
	log.Log().Println(params...)
	if log.PrintToStdout {
		params = slices.Insert[[]any, any](params, 0, prefix, log.formatHeader())
		log.printToStdoutInternal(params...)
	}
	return log
}

// Print info message
func (log *Log) Info(params ...any) *Log {
	prefix := "INFO:" + log.PrefixDelimiter
	log.Log().SetPrefix(prefix)
	log.Log().Println(params...)
	if log.PrintToStdout {
		params = slices.Insert[[]any, any](params, 0, prefix, log.formatHeader())
		log.printToStdoutInternal(params...)
	}
	return log
}

// Print debug message
func (log *Log) Debug(params ...any) *Log {
	prefix := "DEBUG:" + log.PrefixDelimiter
	log.Log().SetPrefix(prefix)
	log.Log().Println(params...)
	if log.PrintToStdout {
		params = slices.Insert[[]any, any](params, 0, prefix, log.formatHeader())
		log.printToStdoutInternal(params...)
	}
	return log
}
