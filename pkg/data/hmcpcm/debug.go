package hmcpcm

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type writer struct {
	io.Writer
	timeFormat string
}

func (w writer) Write(b []byte) (n int, err error) {
	return w.Writer.Write(append([]byte(time.Now().Format(w.timeFormat)), b...))
}

// GetDebugLogger returns a logger handler for HMC API debug data
func GetDebugLogger(filename string) (*log.Logger, error) {
	name := filepath.Join(logDir, "hmcapi_debug_"+strings.Replace(filename, ".", "-", -1)+".log")
	l, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err == nil {
		return log.New(&writer{l, "2006-01-02 15:04:05.00000"}, " [HMCAPI-DEBUG] ", 0), nil
	}
	return nil, err
}

/*/PENDING Close debug file

func CloseDebugLogger(l *log.Logger) {
	if file, ok := l.Out.(*os.File); ok {
		file.Sync()
		file.Close()
	} else if handler, ok := l.Out.(io.Closer); ok {
		handler.Close()
	}
}*/
