package bootstrap_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/bootstrap"
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

// TestAuditLoggerDesProzesses (SPEC-139 Nr. 13 d, SPEC-235 Nr. 5): der Adapter protokolliert einen gescheiterten Append als Zeile in der Senke
// des Adapters des Prozesses; der Logger, den Build erhielt, bleibt leer, in beiden Modi.
func TestAuditLoggerDesProzesses(t *testing.T) {
	for _, modus := range []string{modusRecord, modusReplay} {
		p, senke := &puffer{}, &puffer{}
		h := neuHaken()
		hd, err := bootstrap.BuildMitSenke(context.Background(), konfigMitWerten(t.TempDir(), modus), slog.New(slog.NewTextHandler(p, nil)), senke, h.bauteile())
		if err != nil {
			t.Fatalf("%s: Build: %v", modus, err)
		}
		t.Cleanup(func() { _ = hd.Herunterfahren(context.Background()) })
		if err := h.wartung.Audit.Append(support.Eintrag{Art: "keine-art"}); err == nil {
			t.Fatalf("%s: Append eines ungültigen Eintrags ohne Fehler", modus)
		}
		if !strings.Contains(senke.String(), `"ereignis":"audit_fehlgeschlagen"`) || p.String() != "" {
			t.Errorf("%s: Senke %q, Logger von Build %q; erwartet die Zeile in der Senke und einen leeren Logger", modus, senke.String(), p.String())
		}
	}
}
