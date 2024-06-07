package hogger

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	humanize "github.com/dustin/go-humanize"
)

// Styles
var (
	timeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "240"})

	uriStyle = timeStyle.Copy()

	methodStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "62", Dark: "62"})

	http200Style = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "35", Dark: "48"})

	http300Style = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "208", Dark: "192"})

	http400Style = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "39", Dark: "86"})

	http500Style = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "203", Dark: "204"})

	subtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "250"})

	addressStyle = subtleStyle.Copy()
)

type logWriter struct {
	http.ResponseWriter
	code, bytes int
}

func (r *logWriter) Write(p []byte) (int, error) {
	written, err := r.ResponseWriter.Write(p)
	return written, err
}

func (r *logWriter) WriteHeader(code int) {
	r.code = code
	r.ResponseWriter.WriteHeader(code)
}

// Hijack support for WebSocket
func (r *logWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("WebServer does not support hijacking")
	}
	return hj.Hijack()
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addr := r.RemoteAddr
		if colon := strings.LastIndex(addr, ":"); colon != -1 {
			addr = addr[:colon]
		}

		arrow := subtleStyle.Render("<-")
		method := methodStyle.Render(r.Method)
		uri := uriStyle.Render(r.RequestURI)
		address := addressStyle.Render(addr)

		// Log request
		log.Printf("%s %s %s %s", arrow, method, uri, address)

		writer := &logWriter{
			ResponseWriter: w,
			code:           http.StatusOK,
		}

		arrow = subtleStyle.Render("->")
		startTime := time.Now()

		if r == nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			writer.code = http.StatusInternalServerError
		} else {
			next.ServeHTTP(writer, r)
		}

		elapsedTime := time.Now().Sub(startTime)

		var statusStyle lipgloss.Style

		if writer.code < 300 {
			statusStyle = http200Style
		} else if writer.code < 400 {
			statusStyle = http300Style
		} else if writer.code < 500 {
			statusStyle = http400Style
		} else {
			statusStyle = http500Style
		}

		status := statusStyle.Render(fmt.Sprintf("%d %s", writer.code, http.StatusText(writer.code)))

		formattedBytes := strings.Replace(
			humanize.Bytes(uint64(writer.bytes)),
			" ", "", 1)

		bytes := subtleStyle.Render(formattedBytes)
		time := timeStyle.Render(fmt.Sprintf("%s", elapsedTime))

		log.Printf("%s %s %s %v", arrow, status, bytes, time)
	})
}
