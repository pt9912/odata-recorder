package httpupstream

import (
	"net/http"
	"time"
)

// SetzeFrist setzt den Test-Haken der Frist: f startet den Timer der Frist und liefert seine
// Stop-Funktion.
func SetzeFrist(c *Client, f func(d time.Duration, ablauf func()) (stop func() bool)) {
	c.neueFrist = f
}

// SetzeTransport setzt den Test-Haken des Transports, etwa einen Fake.
func SetzeTransport(c *Client, rt http.RoundTripper) { c.transport = rt }

// Transport liefert den *http.Transport, den New setzt.
func Transport(c *Client) *http.Transport { return c.transport.(*http.Transport) }
