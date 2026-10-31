package replay

import (
	"bytes"
	"context"
	"io"
	"log/slog"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// leererStrom ist der Body einer Antwort ohne Blob (SPEC-135 Nr. 12): nie nil, ohne dass ein Blob geöffnet wird.
func leererStrom() io.ReadCloser {
	return io.NopCloser(bytes.NewReader(nil))
}

// antwortHeader bildet die Header der Auslieferung als Kopie (SPEC-135 Nr. 13): ohne Header, deren Werte
// alle der Marker sind, ohne Content-Length, Content-Encoding, die flüchtigen Header der Regeln und
// ein gespeichertes X-Recorder-Match-Redacted; bei KeyRedacted mit X-Recorder-Match-Redacted: true.
func antwortHeader(e model.Exchange, fluechtig []string) map[string][]string {
	h := model.OhneRedigierte(e.ResponseHeaders)
	weg := map[string]bool{"content-length": true, "content-encoding": true, "x-recorder-match-redacted": true}
	for _, n := range fluechtig {
		weg[asciiKlein(n)] = true
	}
	for n := range h {
		if weg[asciiKlein(n)] {
			delete(h, n)
		}
	}
	if e.KeyRedacted {
		h["X-Recorder-Match-Redacted"] = []string{"true"}
	}
	return h
}

// protokolliere schreibt den einen Eintrag je Antwort auf einen Exchange mit KeyRedacted (SPEC-135 Nr. 16):
// fester Text, sieben Attribute, kein Wert des Requests außer der KorrelationsID.
func protokolliere(ctx context.Context, t support.Telemetry, e model.Exchange, dienst, session, workspace, korrelation string) {
	if !e.KeyRedacted {
		return
	}
	t.Log(ctx, slog.LevelWarn, "replay: Treffer mit redigiertem Match-Key",
		support.Str("mode", "replay"), support.Str("ereignis", "key_redacted"), support.Str("service_id", dienst),
		support.Str("session_id", session), support.Str("workspace_id", workspace),
		support.Str("exchange_id", e.ID), support.Str("correlation_id", korrelation))
}
