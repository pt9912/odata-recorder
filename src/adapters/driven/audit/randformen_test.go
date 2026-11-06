package audit_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/audit"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

func schreibeDatei(t *testing.T, pfad, inhalt string) {
	t.Helper()
	if err := os.WriteFile(pfad, []byte(inhalt), 0o600); err != nil {
		t.Fatal(err)
	}
}

// V2: SPEC-235 Nr. 4 f: ist das letzte Byte der Datei nicht LF, beginnt die neue Zeile mit LF.
// Menge: 256 Bytewerte als letztes Byte × Dateilänge 1 und 2 Bytes = 512 Fälle; nur das Byte
// LF ergibt kein vorangestelltes LF.
func TestLetztesByteDerDatei(t *testing.T) {
	verz := schnellVerz(t)
	zeile := `{"timestamp":"2026-11-06T10:00:00.000000000Z","correlation_id":"` + corID(1) + `","art":"workspace_geloescht","workspace_id":"w-workspace_id"}` + "\n"
	faelle := 0
	for laenge := 1; laenge <= 2; laenge++ {
		for b := range 256 {
			vorher := strings.Repeat("x", laenge-1) + string([]byte{byte(b)})
			a := neuAn(t, filepath.Join(verz, fmt.Sprintf("b-%d-%d.jsonl", laenge, b)), festeZeit)
			schreibeDatei(t, a.pfad, vorher)
			if err := a.d.Append(gueltig(support.ArtWorkspaceGeloescht)); err != nil {
				t.Fatalf("Länge %d, Byte %#02x: %v", laenge, b, err)
			}
			trenner := "\n"
			if b == '\n' {
				trenner = ""
			}
			if got, want := a.inhalt(t), vorher+trenner+zeile; got != want {
				t.Errorf("Länge %d, letztes Byte %#02x: %q, erwartet %q", laenge, b, got, want)
			}
			faelle++
		}
	}
	if faelle != 512 {
		t.Errorf("%d Fälle, erwartet 512", faelle)
	}
}

// V3: SPEC-234 Nr. 1: der Fehlertext von Append nennt keinen Wert des Eintrags. Menge der
// Fehlerpfade: Form je Feld (5), Art, Code, verbotenes Feld, fehlendes Pflichtfeld, Zeit, Öffnen,
// Schreiben = 12; die Werte des Eintrags (und 4242 und 12345) sind die Marker.
func TestFehlertextOhneWert(t *testing.T) {
	type fall struct {
		name string
		e    support.Eintrag
		a    *aufbau
	}
	mit := func(art support.AuditArt, f func(*support.Eintrag)) support.Eintrag {
		e := gueltig(art)
		f(&e)
		return e
	}
	var faelle []fall
	feld := func(name string, art support.AuditArt, f func(*support.Eintrag)) {
		faelle = append(faelle, fall{"Form " + name, mit(art, f), neu(t, festeZeit)})
	}
	feld("service_id", support.ArtSessionGeloescht, func(e *support.Eintrag) { e.ServiceID = "CANARY-svc!" })
	feld("session_id", support.ArtSessionGeloescht, func(e *support.Eintrag) { e.SessionID = "CANARY-ses!" })
	feld("workspace_id", support.ArtWorkspaceGeloescht, func(e *support.Eintrag) { e.WorkspaceID = "CANARY-wsp!" })
	feld("import_id", support.ArtImportFehlgeschlagen, func(e *support.Eintrag) { e.ImportID = "CANARY-imp!" })
	feld("KorrelationsID", support.ArtSessionGeloescht, func(e *support.Eintrag) { e.KorrelationsID = "CANARY-kor!" })
	faelle = append(faelle,
		fall{"Art", mit(support.ArtSessionGeloescht, func(e *support.Eintrag) { e.Art = "CANARY-art" }), neu(t, festeZeit)},
		fall{"Code", mit(support.ArtImportFehlgeschlagen, func(e *support.Eintrag) { e.ErrorCode = model.Code(4242) }), neu(t, festeZeit)},
		fall{"verbotenes Feld", mit(support.ArtWorkspaceGeloescht, func(e *support.Eintrag) { e.ServiceID = "CANARY-verboten" }), neu(t, festeZeit)},
		fall{"fehlendes Pflichtfeld", mit(support.ArtSessionGeloescht, func(e *support.Eintrag) { e.SessionID, e.ServiceID = "", "CANARY-fehlt" }), neu(t, festeZeit)},
		fall{"Zeit", mit(support.ArtSessionGeloescht, func(e *support.Eintrag) { e.ServiceID = "CANARY-zeit" }),
			neu(t, func(int) time.Time { return time.Date(12345, 1, 1, 0, 0, 0, 0, time.UTC) })},
		fall{"Öffnen", mit(support.ArtSessionGeloescht, func(e *support.Eintrag) { e.ServiceID = "CANARY-oeffnen" }),
			neuAn(t, filepath.Join(t.TempDir(), "fehlt", audit.DateiName), festeZeit)},
		fall{"Schreiben", mit(support.ArtSessionGeloescht, func(e *support.Eintrag) { e.ServiceID = "CANARY-schreiben" }),
			neuAn(t, "/dev/full", festeZeit)},
	)
	if len(faelle) != 12 {
		t.Fatalf("%d Fälle, erwartet 12", len(faelle))
	}
	for _, f := range faelle {
		err := f.a.d.Append(f.e)
		if err == nil {
			t.Errorf("%s: Append gelang", f.name)
			continue
		}
		marker := []string{string(f.e.Art), f.e.ServiceID, f.e.SessionID, f.e.WorkspaceID, f.e.ImportID, f.e.KorrelationsID, "4242", "12345", "CANARY"}
		for _, m := range marker {
			if m != "" && strings.Contains(err.Error(), m) {
				t.Errorf("%s: der Fehlertext nennt %q: %v", f.name, m, err)
			}
		}
	}
}
