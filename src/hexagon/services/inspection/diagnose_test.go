package inspection_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/inspection"
)

// Die Literale des Vertrags (SPEC-236 Nr. 1 und 7): das Orakel liest weder Modi noch Schlüssel aus dem
// Paket.
func modiV() []string { return []string{"record", "replay", "mock", "hybrid"} }

func erlaubtV() []string {
	return []string{
		"REPLAY_STRATEGY", "REPLAY_ASSIGNMENT", "REPLAY_REDACTED_BODIES", "MATCH_VARY_HEADERS",
		"MATCH_INVALID_ENCODING", "MATCH_BODY_MODE", "REDACT_HEADERS", "REDACT_QUERY_PARAMS",
		"REDACT_JSON_PATHS", "MAX_CAPTURE_BODY", "UPSTREAM_TIMEOUT", "UPSTREAM_TLS_VERIFY",
		"UPSTREAM_TLS_INSECURE_SKIP_VERIFY", "SHUTDOWN_TIMEOUT", "RETENTION_MAX_AGE",
		"SQLITE_BUSY_TIMEOUT", "VOLATILE_HEADERS", "LINK_REWRITE",
	}
}

// ausgeschlossenV sind die zehn Namen der Nr. 7, je ein Vertreter der Muster ADMIN_*, HYBRID_*,
// FALLBACK_*, LISTEN_TLS_*, PUBLIC_BASE_URL, STRICT_COVERAGE, ein unbekannter Name und der leere Name.
func ausgeschlossenV() []string {
	return []string{
		"MODE", "LISTEN", "DATA_DIR", "SERVICE_ID", "SERVICE_ROOT_PATH", "SERVICE_ODATA_VERSION",
		"UPSTREAM_URL", "SERVICES_FILE", "UPSTREAM_CA_FILE", "ADMIN_TOKEN_FILE",
		"ADMIN_TOKEN", "ADMIN_LISTEN", "HYBRID_UPSTREAM_TOKEN", "FALLBACK_URL", "LISTEN_TLS_CERT",
		"PUBLIC_BASE_URL", "STRICT_COVERAGE", "UNBEKANNTER_NAME", "",
	}
}

// TestErlaubtGleichWortlaut (D5, Zählprüfung der Menge M-Erlaubt): Erlaubt() ist als Menge und in der
// Länge die Liste des Vertrags, frei von Doppelten, und disjunkt zu den ausgeschlossenen Namen.
func TestErlaubtGleichWortlaut(t *testing.T) {
	got := inspection.Erlaubt()
	if len(got) != len(erlaubtV()) {
		t.Errorf("Erlaubt() hat %d Schlüssel, der Vertrag nennt %d", len(got), len(erlaubtV()))
	}
	for _, k := range erlaubtV() {
		if !slices.Contains(got, k) {
			t.Errorf("Erlaubt() lässt %q aus", k)
		}
	}
	for _, k := range got {
		if !slices.Contains(erlaubtV(), k) {
			t.Errorf("Erlaubt() führt %q, der Vertrag nicht", k)
		}
	}
	seen := map[string]bool{}
	for _, k := range erlaubtV() {
		if seen[k] {
			t.Errorf("die Literal-Liste führt %q doppelt", k)
		}
		seen[k] = true
	}
	for _, k := range ausgeschlossenV() {
		if seen[k] {
			t.Errorf("%q steht in der Liste der erlaubten und der ausgeschlossenen", k)
		}
	}
	got[0] = "GEAENDERT"
	if inspection.Erlaubt()[0] == "GEAENDERT" {
		t.Error("Erlaubt() liefert den Slice des Pakets statt eines frischen")
	}
}

// diagnose liest die Diagnose mit lebendem Kontext und verlangt Erfolg.
func diagnose(t *testing.T, u *umgebung) inbound.Diagnose {
	t.Helper()
	d, err := u.dienst(t).Diagnose(context.Background())
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}
	return d
}

// TestDiagnoseKontext (D1): ein beendeter ctx ist sein Fehler, unverändert und ohne Code, und Snapshot
// wird nicht gerufen; ein lebender liefert das Ergebnis, auch wenn er in Snapshot endet.
func TestDiagnoseKontext(t *testing.T) {
	abgelaufen, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
	defer cancel()
	faelle := []struct {
		name   string
		ctx    func() context.Context
		fehler error
	}{
		{"lebend", context.Background, nil},
		{"Canceled", ctxBeendet, context.Canceled},
		{"DeadlineExceeded", func() context.Context { return abgelaufen }, context.DeadlineExceeded},
	}
	for _, c := range faelle {
		u := neueUmgebung()
		d, err := u.dienst(t).Diagnose(c.ctx())
		wantSnapshot := 1
		if c.fehler != nil {
			wantSnapshot = 0
			var fe *inbound.Fehler
			if err != c.fehler || errors.As(err, &fe) || !reflect.DeepEqual(d, inbound.Diagnose{}) {
				t.Errorf("%s: Fehler %v, Ergebnis %+v", c.name, err, d)
			}
		} else if err != nil || len(d.Fehler) == 0 {
			t.Errorf("%s: Fehler %v", c.name, err)
		}
		if got := u.tel.Anzahl(supporttest.OpSnapshot); got != wantSnapshot {
			t.Errorf("%s: Snapshot %d, erwartet %d", c.name, got, wantSnapshot)
		}
	}

	// Der ctx endet im Aufruf von Snapshot: das Ergebnis bleibt, es gibt keine Prüfung danach.
	u := neueUmgebung()
	ctx, cancel2 := context.WithCancel(context.Background())
	u.k.Zaehler = &hakenTelemetrie{FakeTelemetry: u.tel, haken: func(*support.Snapshot) { cancel2() }}
	d, err := u.dienst(t).Diagnose(ctx)
	if err != nil || len(d.Fehler) != len(model.AlleCodes()) || ctx.Err() == nil || u.tel.Anzahl(supporttest.OpSnapshot) != 1 {
		t.Errorf("ctx endet in Snapshot: Fehler %v, %d Zähler, beendet: %v", err, len(d.Fehler), ctx.Err() != nil)
	}
}

// TestDiagnoseNurSnapshot (D2): der Dienst ruft Snapshot genau einmal je Aufruf, die Operationen von
// Telemetry nie; die Zahl der Snapshot-Aufrufe ist die der Aufrufe, die Snapshot erreichen.
func TestDiagnoseNurSnapshot(t *testing.T) {
	u := neueUmgebung()
	d := u.dienst(t)
	erreichen := 0
	for _, ctx := range []context.Context{context.Background(), ctxBeendet(), nilCtx(), context.Background()} {
		_, _ = d.Diagnose(ctx)
		if ctx != nil && ctx.Err() == nil {
			erreichen++
		}
	}
	aufrufe := u.tel.Aufrufe()
	for i, a := range aufrufe {
		if a.Op != supporttest.OpSnapshot {
			t.Errorf("Aufruf %d ist Operation %d, nicht Snapshot", i, a.Op)
		}
	}
	if len(aufrufe) != erreichen || u.tel.Anzahl(supporttest.OpSnapshot) != erreichen {
		t.Errorf("%d aufgezeichnete Aufrufe, %d Snapshot, %d Aufrufe erreichen Snapshot", len(aufrufe), u.tel.Anzahl(supporttest.OpSnapshot), erreichen)
	}
	summe := 0
	for op := supporttest.OpLog; op <= supporttest.OpSnapshot; op++ {
		n := u.tel.Anzahl(op)
		summe += n
		if op != supporttest.OpSnapshot && n != 0 {
			t.Errorf("Operation %d wurde %d-mal gerufen", op, n)
		}
	}
	if summe != len(aufrufe) {
		t.Errorf("Summe der Operationen %d, Aufzeichnung %d", summe, len(aufrufe))
	}
}

// stand erzeugt den Stand der Menge M-Codes für den Durchlauf s: der Eintrag des Codes i hat die Art
// (i + s) mod 5 und ist je Code ein eigener Wert (fehlt, 0, 1, 2 hoch 32 plus Index, größter uint64
// minus Index).
func standMCodes(s int) (map[model.Code]uint64, []uint64) {
	codes := model.AlleCodes()
	errs := map[model.Code]uint64{}
	soll := make([]uint64, len(codes))
	for i, c := range codes {
		switch (i + s) % 5 {
		case 0:
			soll[i] = 0
		case 1:
			errs[c], soll[i] = 0, 0
		case 2:
			errs[c], soll[i] = 1, 1
		case 3:
			errs[c], soll[i] = 1<<32+uint64(i), 1<<32+uint64(i)
		case 4:
			errs[c], soll[i] = math.MaxUint64-uint64(i), math.MaxUint64-uint64(i)
		}
	}
	return errs, soll
}

// TestDiagnoseFehlerZaehler (D3, Menge M-Codes): Fehler hat einen Eintrag je Code von model.AlleCodes()
// in dessen Reihenfolge, Anzahl gleich dem Zähler, ein Code ohne Eintrag hat 0.
func TestDiagnoseFehlerZaehler(t *testing.T) {
	codes := model.AlleCodes()
	laeufe, elemente := 0, 0
	for s := range 5 {
		laeufe++
		errs, soll := standMCodes(s)
		u := neueUmgebung()
		u.tel.SetzeStand(support.Snapshot{Errors: errs})
		d := diagnose(t, u)
		if len(d.Fehler) != len(codes) {
			t.Fatalf("Durchlauf %d: %d Einträge, %d Codes", s, len(d.Fehler), len(codes))
		}
		for i, c := range codes {
			elemente++
			if d.Fehler[i].Code != c || d.Fehler[i].Anzahl != soll[i] {
				t.Errorf("Durchlauf %d, Eintrag %d: Code %v Anzahl %d, erwartet %v %d", s, i, d.Fehler[i].Code, d.Fehler[i].Anzahl, c, soll[i])
			}
		}
	}
	if elemente != laeufe*len(codes) {
		t.Errorf("%d Elemente gezählt, die Achsen verlangen %d", elemente, laeufe*len(codes))
	}
}

// TestDiagnoseErrorsRand (D4): Errors nil, leer oder mit einem Code gültig; Einträge außerhalb der
// Tabelle werden verworfen.
func TestDiagnoseErrorsRand(t *testing.T) {
	codes := model.AlleCodes()
	groesster := slices.Max(codes)
	aussen := []model.Code{0, -1, groesster + 1, model.Code(math.MaxInt)}
	for _, c := range aussen {
		if slices.Contains(codes, c) {
			t.Fatalf("der Code %d liegt in der Tabelle", c)
		}
	}
	faelle := map[string]map[model.Code]uint64{
		"nil": nil, "leer": {}, "ein Code": {codes[len(codes)-1]: 7},
	}
	for name, errs := range faelle {
		u := neueUmgebung()
		u.tel.SetzeStand(support.Snapshot{Errors: errs})
		d := diagnose(t, u)
		if len(d.Fehler) != len(codes) {
			t.Errorf("%s: %d Einträge", name, len(d.Fehler))
		}
		for i, c := range codes {
			if d.Fehler[i].Code != c || d.Fehler[i].Anzahl != errs[c] {
				t.Errorf("%s: Eintrag %d: %v %d, erwartet %v %d", name, i, d.Fehler[i].Code, d.Fehler[i].Anzahl, c, errs[c])
			}
		}
	}
	errs := map[model.Code]uint64{}
	for i, c := range aussen {
		errs[c] = uint64(100 + i)
	}
	u := neueUmgebung()
	u.tel.SetzeStand(support.Snapshot{Errors: errs})
	d := diagnose(t, u)
	if len(d.Fehler) != len(codes) {
		t.Errorf("außerhalb: %d Einträge, %d Codes", len(d.Fehler), len(codes))
	}
	for _, f := range d.Fehler {
		if slices.Contains(aussen, f.Code) || f.Anzahl >= 100 {
			t.Errorf("Eintrag außerhalb der Tabelle durchgereicht: %+v", f)
		}
	}
}

// abwandlungen sind die fünf Abwandlungen eines erlaubten Schlüssels, die kein Treffer sind.
func abwandlungen(k string) []string {
	return []string{strings.ToLower(k), " " + k, k + " ", "X_" + k, k + "_X"}
}

// TestDiagnoseEinstellungen (D5, Menge M-Erlaubt: Schlüssel × Abwandlung × Modus): Einstellungen enthält
// genau die Einträge mit einem Schlüssel byte-gleich aus Erlaubt(), mit unverändertem Wert, in jedem
// Modus gleich.
func TestDiagnoseEinstellungen(t *testing.T) {
	faelle := 0
	for _, modus := range modiV() {
		for _, k := range erlaubtV() {
			for i, schluessel := range append([]string{k}, abwandlungen(k)...) {
				faelle++
				u := neueUmgebung()
				u.k.Modus = modus
				u.k.Einstellungen = map[string]string{schluessel: "wert-" + k}
				d := diagnose(t, u)
				want := map[string]string{}
				if i == 0 {
					want[k] = "wert-" + k
				}
				if !reflect.DeepEqual(d.Einstellungen, want) {
					t.Errorf("Modus %s, Schlüssel %q: %v, erwartet %v", modus, schluessel, d.Einstellungen, want)
				}
			}
		}
	}
	if want := len(modiV()) * len(erlaubtV()) * (1 + len(abwandlungen("X"))); faelle != want {
		t.Errorf("%d Fälle gelaufen, die Achsen verlangen %d", faelle, want)
	}

	// Alle Schlüssel und alle Abwandlungen zugleich, in jedem Modus gleich.
	voll := map[string]string{}
	soll := map[string]string{}
	for _, k := range erlaubtV() {
		voll[k], soll[k] = "wert-"+k, "wert-"+k
		for _, a := range abwandlungen(k) {
			voll[a] = "anderer-" + a
		}
	}
	var erste map[string]string
	for _, modus := range modiV() {
		u := neueUmgebung()
		u.k.Modus = modus
		u.k.Einstellungen = voll
		d := diagnose(t, u)
		if !reflect.DeepEqual(d.Einstellungen, soll) {
			t.Errorf("Modus %s: %d Einträge, erwartet %d", modus, len(d.Einstellungen), len(soll))
		}
		if erste != nil && !reflect.DeepEqual(erste, d.Einstellungen) {
			t.Errorf("Modus %s: Einstellungen weichen vom ersten Modus ab", modus)
		}
		erste = d.Einstellungen
	}

	// Ein erlaubter Schlüssel ohne Eintrag fehlt, einer mit leerem Wert bleibt mit leerem Wert.
	u := neueUmgebung()
	if d := diagnose(t, u); len(d.Einstellungen) != 0 || d.Einstellungen == nil {
		t.Errorf("ohne Einträge: %v", d.Einstellungen)
	}
	for _, k := range erlaubtV() {
		u := neueUmgebung()
		u.k.Einstellungen = map[string]string{k: ""}
		d := diagnose(t, u)
		if v, da := d.Einstellungen[k]; !da || v != "" || len(d.Einstellungen) != 1 {
			t.Errorf("%s mit leerem Wert: %v", k, d.Einstellungen)
		}
	}
}

// canary bildet den Wert des Schlüssels i aus Bytewerten (ein Zeichen jenseits von ASCII, dann Text).
func canary(i int) string {
	return string([]byte{0xc3, 0xa4}) + "Kanarie" + fmt.Sprint(i)
}

// zeichenketten sammelt jede Zeichenkette in v (Schlüssel und Werte von Abbildungen, Felder, Elemente).
func zeichenketten(v reflect.Value, out *[]string) {
	switch v.Kind() {
	case reflect.String:
		*out = append(*out, v.String())
	case reflect.Map:
		for _, k := range v.MapKeys() {
			zeichenketten(k, out)
			zeichenketten(v.MapIndex(k), out)
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			zeichenketten(v.Index(i), out)
		}
	case reflect.Struct:
		for i := range v.NumField() {
			zeichenketten(v.Field(i), out)
		}
	}
}

// diagnoseKodierungen sind die drei Kodierungen der Menge M-Ausgeschlossen.
func diagnoseKodierungen(t *testing.T, d inbound.Diagnose) map[string]string {
	t.Helper()
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var alle []string
	zeichenketten(reflect.ValueOf(d), &alle)
	return map[string]string{"json": string(b), "%+v": fmt.Sprintf("%+v", d), "Gang": strings.Join(alle, "\n")}
}

// TestDiagnoseKeinGeheimnis (D6, Menge M-Ausgeschlossen: Name × Kodierung): kein Wert unter einem
// ausgeschlossenen Schlüssel erscheint in einer Kodierung; derselbe Wert unter einem erlaubten
// Schlüssel erscheint in jeder.
func TestDiagnoseKeinGeheimnis(t *testing.T) {
	faelle := 0
	for i, name := range ausgeschlossenV() {
		u := neueUmgebung()
		u.k.Einstellungen = map[string]string{name: canary(i)}
		for kod, text := range diagnoseKodierungen(t, diagnose(t, u)) {
			faelle++
			if strings.Contains(text, canary(i)) {
				t.Errorf("Name %q, Kodierung %s: der Wert erscheint", name, kod)
			}
		}
	}
	if want := len(ausgeschlossenV()) * 3; faelle != want {
		t.Errorf("%d Fälle gelaufen, die Achsen verlangen %d", faelle, want)
	}
	for i, name := range erlaubtV() {
		u := neueUmgebung()
		u.k.Einstellungen = map[string]string{name: canary(i)}
		for kod, text := range diagnoseKodierungen(t, diagnose(t, u)) {
			if !strings.Contains(text, canary(i)) {
				t.Errorf("Gegenprobe, Schlüssel %q, Kodierung %s: der Wert erscheint nicht", name, kod)
			}
		}
	}
}

// TestDiagnoseModus (D7): Modus ist der Modus der Konfiguration, unverändert, auch wenn Einstellungen
// einen anderen unter MODE nennen.
func TestDiagnoseModus(t *testing.T) {
	for _, m := range modiV() {
		u := neueUmgebung()
		u.k.Modus = m
		u.k.Einstellungen = map[string]string{"MODE": "andere-" + m}
		if d := diagnose(t, u); d.Modus != m {
			t.Errorf("Modus %q: Diagnose nennt %q", m, d.Modus)
		}
	}
}

// TestDiagnoseKopie (D8): Einstellungen und Fehler sind nie nil; eine Änderung am Ergebnis wirkt weder
// auf den Dienst noch auf das nächste Ergebnis; jeder Aufruf liest den Stand neu.
func TestDiagnoseKopie(t *testing.T) {
	u := neueUmgebung()
	u.k.Einstellungen = nil
	u.tel.SetzeStand(support.Snapshot{})
	d := diagnose(t, u)
	if b, err := json.Marshal(d.Einstellungen); err != nil || string(b) != "{}" {
		t.Errorf("Einstellungen als JSON: %s, %v", b, err)
	}
	if b, err := json.Marshal(d.Fehler); d.Fehler == nil || err != nil || !strings.HasPrefix(string(b), "[") {
		t.Errorf("Fehler als JSON: %s, %v", b, err)
	}

	v := neueUmgebung()
	v.k.Einstellungen = map[string]string{"LINK_REWRITE": "off"}
	dienst := v.dienst(t)
	v.tel.SetzeStand(support.Snapshot{Errors: map[model.Code]uint64{model.AlleCodes()[0]: 5}})
	erst, _ := dienst.Diagnose(context.Background())
	erst.Einstellungen["LINK_REWRITE"] = "geaendert"
	erst.Einstellungen["REPLAY_STRATEGY"] = "neu"
	delete(erst.Einstellungen, "MATCH_BODY_MODE")
	erst.Fehler[0].Anzahl = 99
	zweit, err := dienst.Diagnose(context.Background())
	if err != nil || zweit.Einstellungen["LINK_REWRITE"] != "off" || len(zweit.Einstellungen) != 1 || zweit.Fehler[0].Anzahl != 5 {
		t.Errorf("nach der Änderung des ersten Ergebnisses: %+v, %v", zweit.Einstellungen, err)
	}
	v.tel.SetzeStand(support.Snapshot{Errors: map[model.Code]uint64{model.AlleCodes()[0]: 6}})
	dritt, _ := dienst.Diagnose(context.Background())
	if dritt.Fehler[0].Anzahl != 6 {
		t.Errorf("der Stand wurde nicht neu gelesen: %d", dritt.Fehler[0].Anzahl)
	}
}
