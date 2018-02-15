package hmcpcm

import (
	"github.com/Sirupsen/logrus"
	"net/http"
)

var (
	logDir string
)

// SetLogDir xx
func SetLogDir(dir string) {
	logDir = dir
}

// Logger interface
type Logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
}

// Session is the HTTP session struct
type Session struct {
	client   *http.Client
	User     string
	Password string
	url      string
	Debug    bool
	samples  int
	slog     *logrus.Logger // The sytem Log
	dlog     Logger         // The debug log
}
