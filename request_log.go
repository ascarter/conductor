package conductor

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type responseLogger struct {
	w      http.ResponseWriter
	status int
	size   int
}

func (r *responseLogger) Header() http.Header {
	return r.w.Header()
}

func (r *responseLogger) Write(b []byte) (int, error) {
	if r.status == 0 {
		// Status will be StatusOK if WriteHeader not called yet
		r.status = http.StatusOK
	}
	size, err := r.w.Write(b)
	r.size += size
	return size, err
}

func (r *responseLogger) WriteHeader(s int) {
	r.w.WriteHeader(s)
	r.status = s
}

func (r *responseLogger) Status() int {
	return r.status
}

func (r *responseLogger) Size() int {
	return r.size
}

// A LoggerOutput provides an output target for a log.
type LoggerOutput interface {
	Printf(string, ...interface{})
}

// RequestLogHandler logs request start and end. The start log line includes information
// on the request. The end log line includes result of request and time elapsed.
func RequestLogHandler(h http.Handler, logger LoggerOutput) http.Handler {
	if logger == nil {
		// Emulate standard logger
		logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseLogger{w: w}

		raddr := r.Header.Get("X-Forwarded-For")
		if raddr == "" {
			raddr = r.RemoteAddr
		}

		rid := r.Header.Get("X-Request-ID")
		if rid != "" {
			rid = fmt.Sprintf("[%s] ", rid)
		}

		logger.Printf("%sStarted %s %s for %s", rid, r.Method, r.URL.Path, raddr)
		h.ServeHTTP(rw, r)
		logger.Printf("%sCompleted %v %s in %v", rid, rw.Status(), http.StatusText(rw.Status()), time.Since(start))
	})
}

// RequestLogComponent returns a RequestLogHandler as a Component
func RequestLogComponent(logger LoggerOutput) Component {
	return ComponentFunc(func(h http.Handler) http.Handler {
		return RequestLogHandler(h, logger)
	})
}

// DefaultRequestLogComponent is RequestLogComponent that logs to stderr
var DefaultRequestLogComponent = RequestLogComponent(nil)

// DefaultRequestLogHandler is an http.Handler that logs to stderr
var DefaultRequestLogHandler = func(h http.Handler) http.Handler {
	return RequestLogHandler(h, nil)
}
