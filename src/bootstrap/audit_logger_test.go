package bootstrap_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// puffer ist ein Schreiber, den mehrere Goroutinen nutzen dürfen.
type puffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (p *puffer) Write(x []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.b.Write(x)
}

func (p *puffer) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.b.String()
}

// TestAuditLoggerDesProzesses (SPEC-139 Nr. 3, SPEC-235 Nr. 5): der Adapter protokolliert einen gescheiterten
// Append an dem einen Logger des Prozesses, den Build erhielt, in beiden Modi.
func TestAuditLoggerDesProzesses(t *testing.T) {
	for _, modus := range []string{modusRecord, modusReplay} {
		p := &puffer{}
		log := slog.New(slog.NewTextHandler(p, nil))
		h := neuHaken()
		baueMitLog(context.Background(), t, konfigMitWerten(t.TempDir(), modus), h, log)
		if err := h.wartung.Audit.Append(support.Eintrag{Art: "keine-art"}); err == nil {
			t.Fatalf("%s: Append eines ungültigen Eintrags ohne Fehler", modus)
		}
		if !strings.Contains(p.String(), "audit_fehlgeschlagen") {
			t.Errorf("%s: der Logger des Prozesses sah den Fehler nicht: %q", modus, p.String())
		}
	}
}
