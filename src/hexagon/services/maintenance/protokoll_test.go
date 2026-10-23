package maintenance_test

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// keinMarker prüft, dass der Marker in keinem Text und keinem Attribut eines Eintrags steht.
func keinMarker(t *testing.T, name string, eintraege []eintrag) {
	t.Helper()
	for _, e := range eintraege {
		alle := e.text
		for k, v := range e.attrs {
			alle += " " + k + "=" + v
		}
		if strings.Contains(alle, marker) {
			t.Errorf("%s: der Marker steht im Eintrag %q", name, alle)
		}
	}
}

// TestProtokoll: je Ereignis genau die Attribute aus SPEC-123 Nr. 11 und kein Text eines Fehlers
// oder Befunds; CleanupBlobs schreibt nichts (W21).
func TestProtokoll(t *testing.T) {
	u := neu(t)
	u.speicher.befunde = []string{"Fremdschlüssel " + marker, "Index " + marker}
	_, err := u.startup(context.Background())
	soll := []eintrag{{stufe: slog.LevelError, text: "wartung: Integritätsprüfung mit Befund", attrs: map[string]string{
		"ereignis": "integritaet", "error_code": "ODR-E3001", "befunde": "2"}}}
	if err == nil || !reflect.DeepEqual(u.log.alle(), soll) {
		t.Errorf("integritaet: %v %+v", err, u.log.alle())
	}
	keinMarker(t, "integritaet", u.log.alle())

	v := neu(t, mitMaxAge(30*tag))
	v.dienst(dienstA)
	v.sessions.fuege("q1", dienstA, t0().Add(-40*tag), model.SessionComplete)
	v.speicher.fest = []string{r("d"), r("f")}
	v.blobs.lege(r("f"))
	v.st.scheitere("Delete", 1, fmt.Errorf("%s: %w", marker, outbound.ErrReferenziert))
	v.st.scheitere("Verify", 1, fmt.Errorf("%s: %w", marker, outbound.ErrBlobFehlt))
	v.st.scheitere("Verify", 2, fmt.Errorf("%s: %w", marker, outbound.ErrBlobBeschaedigt))
	if _, err := v.startup(context.Background()); err != nil {
		t.Fatal(err)
	}
	soll = []eintrag{
		{stufe: slog.LevelWarn, text: "wartung: Session als Quelle einer Baseline nicht gelöscht", attrs: map[string]string{
			"ereignis": "aufbewahrung_uebersprungen", "error_code": "ODR-E7004", "service_id": dienstA, "session_id": "q1"}},
		{stufe: slog.LevelError, text: "wartung: referenzierte Blob-Datei fehlt", attrs: map[string]string{
			"ereignis": "blob_fehlt", "error_code": "ODR-E3001", "blob": r("d")}},
		{stufe: slog.LevelError, text: "wartung: referenzierte Blob-Datei beschädigt", attrs: map[string]string{
			"ereignis": "blob_beschaedigt", "error_code": "ODR-E3001", "blob": r("f")}},
	}
	if got := v.log.alle(); !reflect.DeepEqual(got, soll) {
		t.Errorf("Ereignisse: %+v", got)
	}
	keinMarker(t, "Ereignisse", v.log.alle())

	// Ein Fehler eines Ports, dessen Text den Marker trägt, schreibt keinen Eintrag.
	for _, op := range []string{"Check", "FindOpen", "SetStatus", "Dienste.List", "Sessions.List", "Delete", "ReferencedBlobs",
		"Verify", "Blobs.List", "Remove"} {
		w := neu(t, mitMaxAge(30*tag))
		grosserBestand(w)
		w.st.scheitere(op, 1, errDisk)
		if _, err := w.startup(context.Background()); err == nil {
			t.Errorf("%s: kein Fehler", op)
		}
		if got := w.log.alle(); len(got) != 0 {
			t.Errorf("%s: Einträge %+v", op, got)
		}
	}

	// CleanupBlobs schreibt nichts, mit und ohne Fehler und bei Zurückstellen.
	for _, op := range []string{"", "FindOpen", "Blobs.List", "Remove"} {
		w := neu(t)
		bestandBereinigung(w)
		if op != "" {
			w.st.scheitere(op, 1, errDisk)
		}
		_, _ = w.cleanup(context.Background())
		if got := w.log.alle(); len(got) != 0 {
			t.Errorf("CleanupBlobs (%q): Einträge %+v", op, got)
		}
	}
	z := neu(t)
	z.sessions.fuege("s1", dienstA, t0(), model.SessionOpen)
	_, _ = z.cleanup(context.Background())
	if got := z.log.alle(); len(got) != 0 {
		t.Errorf("CleanupBlobs zurückgestellt: Einträge %+v", got)
	}
}
