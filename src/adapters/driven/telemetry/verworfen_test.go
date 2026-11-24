package telemetry_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// verworfeneTexte sind die vier Auslöser der Fallback-Policy aus dem Wortlaut von SPEC-230 Nr. 11, in
// dieser Reihenfolge; verworfeneOrakel liest sie als Codes und prüft ihre Zahl gegen die Quelle.
func verworfeneTexte() []string {
	return []string{"ODR-E5001", "ODR-E6001", "ODR-E6002", "ODR-E6003"}
}

func verworfeneOrakel(t *testing.T) []model.Code {
	t.Helper()
	var codes []model.Code
	for _, text := range verworfeneTexte() {
		c, ok := model.ParseCode(text)
		if !ok {
			t.Fatalf("%s ist kein Code der Tabelle", text)
		}
		codes = append(codes, c)
	}
	if len(codes) != 4 || !slices.Equal(codes, support.VerworfeneCodes()) {
		t.Fatalf("Orakel %v, support.VerworfeneCodes() %v; erwartet 4 gleiche Codes", codes, support.VerworfeneCodes())
	}
	return codes
}

// verworfenGleich prüft, dass m genau die Schlüssel von schluessel trägt und dass jeder Zähler den
// Wert in soll hat (fehlt er dort, 0).
func verworfenGleich(t *testing.T, wo string, m map[model.Code]uint64, schluessel []model.Code, soll map[model.Code]uint64) {
	t.Helper()
	if m == nil || len(m) != len(schluessel) {
		t.Errorf("%s: %d Einträge (nil: %t), erwartet %d", wo, len(m), m == nil, len(schluessel))
	}
	for _, c := range schluessel {
		if n, ok := m[c]; !ok || n != soll[c] {
			t.Errorf("%s: %s zählt %d (vorhanden %t), erwartet %d", wo, c, n, ok, soll[c])
		}
	}
}

// aussenCodes sind die Codes der Tabelle von SPEC-109 (die Texte aus alleCodes) außerhalb von menge.
func aussenCodes(t *testing.T, menge []model.Code) []model.Code {
	t.Helper()
	var aussen []model.Code
	for _, text := range alleCodes() {
		c, ok := model.ParseCode(text)
		if !ok {
			t.Fatalf("%s ist kein Code der Tabelle", text)
		}
		if !slices.Contains(menge, c) {
			aussen = append(aussen, c)
		}
	}
	return aussen
}

// P4: CountDiscarded(c) erhöht den Zähler von c aus der Menge der vier Codes um 1; jeder andere Code
// erhöht keinen Zähler und zählt fehlercode; Errors bleibt unverändert. Menge M-Codes: die 4 Codes der
// Menge (je 1, dann 2), die 37 Codes der Tabelle von SPEC-109 außerhalb und die Randwerte 0, -1 und 42
// gleich 44 Eingaben. Orakel: Literale (die Texte der Tabelle, `Verworfen` und `Dropped[fehlercode]`
// als erwartete Zahlen), nicht der Code des Adapters. Zählprüfung: 41 gleich 4 + 37.
func TestCountDiscardedJeCode(t *testing.T) {
	menge := verworfeneOrakel(t)
	aussen := aussenCodes(t, menge)
	randwerte := []model.Code{0, -1, 42}
	for _, c := range randwerte {
		if c.Gueltig() {
			t.Fatalf("der Randwert %d ist ein Code der Tabelle", int(c))
		}
	}
	if len(alleCodes()) != 41 || len(model.AlleCodes()) != 41 || len(aussen) != 37 || len(menge) != 4 || len(aussen)+len(menge) != 41 ||
		len(aussen)+len(menge)+len(randwerte) != 44 {
		t.Fatalf("Zählprüfung: %d Codes im Orakel, %d in der Quelle, %d außerhalb, %d in der Menge; erwartet 41, 41, 37 und 4", len(alleCodes()), len(model.AlleCodes()), len(aussen), len(menge))
	}
	fehlercode := func(s support.Snapshot) uint64 { return s.Dropped[support.VerwurfFehlercode] }
	for i, c := range menge {
		a, _ := neu(t)
		for wiederholung := uint64(1); wiederholung <= 2; wiederholung++ {
			a.CountDiscarded(c)
			s := a.Snapshot()
			wo := fmt.Sprintf("%s, Aufruf %d", verworfeneTexte()[i], wiederholung)
			verworfenGleich(t, wo, s.Verworfen, menge, map[model.Code]uint64{c: wiederholung})
			pruefeVektor(t, wo, s, nil)
			if n := summe(s.Errors); n != 0 || len(s.Errors) != 41 {
				t.Errorf("%s: Errors zählt %d bei %d Einträgen, erwartet 0 bei 41", wo, n, len(s.Errors))
			}
		}
	}
	for _, c := range append(slices.Clone(aussen), randwerte...) {
		a, _ := neu(t)
		for wiederholung := uint64(1); wiederholung <= 2; wiederholung++ {
			a.CountDiscarded(c)
			s := a.Snapshot()
			wo := fmt.Sprintf("Code %d (%s), Aufruf %d", int(c), c, wiederholung)
			verworfenGleich(t, wo, s.Verworfen, menge, nil)
			pruefeVektor(t, wo, s, map[support.Verwurf]uint64{support.VerwurfFehlercode: wiederholung})
			if n := summe(s.Errors); n != 0 || len(s.Errors) != 41 || fehlercode(s) != wiederholung {
				t.Errorf("%s: Errors zählt %d bei %d Einträgen, fehlercode %d; erwartet 0 bei 41 und %d", wo, n, len(s.Errors), fehlercode(s), wiederholung)
			}
		}
	}
}

// P4 (Uhr): CountDiscarded liest die Uhr nicht (SPEC-230 Nr. 11). Menge: die 4 Codes der Menge und 3
// Codes außerhalb, je ein Aufruf; Orakel: 0 Ablesungen der Uhr, auch nach einem Snapshot.
func TestCountDiscardedOhneUhr(t *testing.T) {
	uhr := festeUhr()
	a := neuMit(t, io.Discard, uhr, slog.LevelInfo)
	codes := append(slices.Clone(verworfeneOrakel(t)), 0, -1, 42)
	for _, c := range codes {
		a.CountDiscarded(c)
	}
	a.Snapshot()
	if n := uhr.n.Load(); n != 0 || len(codes) != 7 {
		t.Errorf("%d Ablesungen der Uhr bei %d Aufrufen, erwartet 0 bei 7", n, len(codes))
	}
}

// P5: Snapshot().Verworfen hat je Code der Menge einen Eintrag, auch mit 0, ist nie nil und eine
// tiefe Kopie in beide Richtungen. Zählprüfung: 4 Einträge.
func TestVerworfenStand(t *testing.T) {
	menge := verworfeneOrakel(t)
	a, _ := neu(t)
	leer := a.Snapshot()
	verworfenGleich(t, "frischer Adapter", leer.Verworfen, menge, nil)
	a.CountDiscarded(menge[2])
	s := a.Snapshot()
	verworfenGleich(t, "nach einer Zählung", s.Verworfen, menge, map[model.Code]uint64{menge[2]: 1})
	// Änderung der Kopie wirkt nicht zurück (jeder Eintrag, auch ein neuer Schlüssel und ein gelöschter)
	for _, c := range menge {
		s.Verworfen[c] = 99
	}
	s.Verworfen[model.CodeE2001] = 99
	delete(s.Verworfen, menge[0])
	verworfenGleich(t, "nach der Änderung der Kopie", a.Snapshot().Verworfen, menge, map[model.Code]uint64{menge[2]: 1})
	// ein späterer Aufruf ändert den früheren Stand nicht
	frueher := a.Snapshot()
	a.CountDiscarded(menge[0])
	verworfenGleich(t, "früherer Stand nach einem Aufruf", frueher.Verworfen, menge, map[model.Code]uint64{menge[2]: 1})
	verworfenGleich(t, "späterer Stand", a.Snapshot().Verworfen, menge, map[model.Code]uint64{menge[0]: 1, menge[2]: 1})
	if len(menge) != 4 {
		t.Errorf("%d Codes, erwartet 4", len(menge))
	}
}

// P7: die beiden Reihen sind getrennt: CountError ändert Verworfen nie (die 41 Codes der Tabelle), und
// CountDiscarded ändert Errors nie (die 4 Codes der Menge). Ein Code der Menge in beiden Reihen zählt
// in beiden je einmal. Zählprüfung: 41 und 4.
func TestReihenGetrennt(t *testing.T) {
	menge := verworfeneOrakel(t)
	codes := model.AlleCodes()
	if len(codes) != 41 {
		t.Fatalf("%d Codes, erwartet 41", len(codes))
	}
	for _, c := range codes {
		a, _ := neu(t)
		a.CountError(c)
		s := a.Snapshot()
		verworfenGleich(t, "CountError "+c.String(), s.Verworfen, menge, nil)
		if s.Errors[c] != 1 || summe(s.Errors) != 1 {
			t.Errorf("CountError %s: Errors[c] %d, Summe %d, erwartet 1 und 1", c, s.Errors[c], summe(s.Errors))
		}
	}
	for _, c := range menge {
		a, _ := neu(t)
		a.CountDiscarded(c)
		if s := a.Snapshot(); summe(s.Errors) != 0 || len(s.Errors) != 41 || s.Verworfen[c] != 1 {
			t.Errorf("CountDiscarded %s: Errors Summe %d bei %d Einträgen, Verworfen %d; erwartet 0 bei 41 und 1", c, summe(s.Errors), len(s.Errors), s.Verworfen[c])
		}
		a.CountError(c)
		if s := a.Snapshot(); s.Errors[c] != 1 || s.Verworfen[c] != 1 || summe(s.Errors) != 1 || summe(s.Verworfen) != 1 {
			t.Errorf("%s in beiden Reihen: Errors %d, Verworfen %d; erwartet je 1", c, s.Errors[c], s.Verworfen[c])
		}
	}
}

// hybridPaare sind die sechs Paare Art × Ergebnis des Spans odata.hybrid.fallback aus dem Wortlaut von
// SPEC-321 Nr. 4 (zwei Arten, drei Ergebnisse).
func hybridPaare() [][2]string {
	var paare [][2]string
	for _, art := range []string{"replay", "upstream"} {
		for _, ergebnis := range []string{"antwort", "fehler", "abgebrochen"} {
			paare = append(paare, [2]string{art, ergebnis})
		}
	}
	return paare
}

// P2: der Adapter nimmt odata.hybrid.fallback mit beiden Attributen ohne Verwurf an, zeichnet
// Verteilung und Eintrag auf und schreibt die Zeile in der Reihenfolge der Tabelle. Menge: die 6 Paare
// Art × Ergebnis × 2 Wege (Art in StartSpan und Ergebnis in Set; beide in StartSpan). Orakel: die Zeile
// als Literal, `Durations` mit 6 Einträgen zu je 12 Eimern, Dropped alle 0.
func TestHybridSpanZeile(t *testing.T) {
	start := time.Date(2026, 11, 7, 10, 0, 0, 0, time.UTC)
	const vorlage = `{"timestamp":"2026-11-07T10:00:00.007000000Z","level":"INFO","msg":"odata.hybrid.fallback","ereignis":"span","span_id":"spn-0000000000000001","duration_ms":7.000,"recorder.fallback_type":"%s","recorder.fallback_result":"%s"}` + "\n"
	laeufe := 0
	for _, p := range hybridPaare() {
		for _, weg := range []string{"Set", "StartSpan"} {
			laeufe++
			senke := &FakeSenke{}
			a := neuMitMax(t, senke, &folgeUhr{zeiten: []time.Time{start, start.Add(7 * time.Millisecond)}}, slog.LevelInfo, 0)
			wo := p[0] + " " + p[1] + " " + weg
			vor := a.Snapshot()
			if len(vor.Durations) != 6 || len(vor.Durations["odata.hybrid.fallback"].Counts) != 12 {
				t.Fatalf("%s: vor dem Beginn %d Verteilungen, hybrid mit %d Eimern; erwartet 6 und 12", wo, len(vor.Durations), len(vor.Durations["odata.hybrid.fallback"].Counts))
			}
			var sp support.Span
			if weg == "Set" {
				_, sp = a.StartSpan(context.Background(), "odata.hybrid.fallback", support.Str("recorder.fallback_type", p[0]))
				sp.Set(support.Str("recorder.fallback_result", p[1]))
			} else {
				_, sp = a.StartSpan(context.Background(), "odata.hybrid.fallback",
					support.Str("recorder.fallback_type", p[0]), support.Str("recorder.fallback_result", p[1]))
			}
			sp.End()
			s := a.Snapshot()
			pruefeVektor(t, wo, s, nil)
			want := []support.Attr{support.Str("recorder.fallback_type", p[0]), support.Str("recorder.fallback_result", p[1])}
			if len(s.Spans) != 1 || s.Spans[0].Name != "odata.hybrid.fallback" || !attrsGleich(s.Spans[0].Attrs, want) || s.Spans[0].ParentID != "" {
				t.Errorf("%s: Eintrag %+v, erwartet %v", wo, s.Spans, want)
			}
			v := s.Durations["odata.hybrid.fallback"]
			counts := make([]uint64, 12)
			counts[3] = 1 // 7 ms liegen im Eimer bis 10 ms
			if len(s.Durations) != 6 || v.Count != 1 || v.Sum != 7*time.Millisecond || !slices.Equal(v.Counts, counts) {
				t.Errorf("%s: %d Verteilungen, hybrid %+v; erwartet 6, Count 1, Sum 7ms und Eimer 3", wo, len(s.Durations), v)
			}
			if z := senke.zeilen(); len(z) != 1 || string(z[0]) != fmt.Sprintf(vorlage, p[0], p[1]) {
				t.Errorf("%s: Zeilen %q, erwartet %q", wo, z, fmt.Sprintf(vorlage, p[0], p[1]))
			}
		}
	}
	if len(hybridPaare()) != 6 || laeufe != 12 {
		t.Errorf("%d Paare, %d Läufe; erwartet 6 und 12", len(hybridPaare()), laeufe)
	}
}

// P3 (1): die zwei Schlüssel haben die Form Token: ein Verstoß zählt wert (am Ende pflicht, das
// Attribut entfällt), ein gültiger Wert nicht. Menge M-Schlüssel: 2 Schlüssel × 3 Verstöße (leer,
// Großbuchstabe, Byte ab 0x80) = 6, dazu die Ränder der Form Token (33 Zeichen, erstes Zeichen Ziffer,
// Unterstrich, Punkt, Bindestrich, Leerraum: 6 × 2 = 12) und die gültigen Werte (die 5 der Stelle, 1 und
// 32 Zeichen, Ziffern und Unterstrich im Wort: 2 × 9 = 18); je Element 2 Wege (StartSpan, Set). Paar:
// beide Werte ungültig zählt wert 2 und am Ende pflicht 2, der Span wird aufgezeichnet.
func TestHybridSchluesselForm(t *testing.T) {
	keys := []string{"recorder.fallback_type", "recorder.fallback_result"}
	andere := map[string]string{"recorder.fallback_type": "replay", "recorder.fallback_result": "antwort"}
	basis := []string{"", "Replay", string([]byte{0x80})}
	rand := []string{strings.Repeat("a", 33), "1replay", "_replay", "re.play", "re-play", "re play"}
	gut := []string{"replay", "upstream", "antwort", "fehler", "abgebrochen", "a", strings.Repeat("a", 32), "a1", "a_b"}
	// anderer ist der zweite Schlüssel des Spans, der mit gültigem Wert mitläuft
	anderer := func(key string) string {
		if key == keys[0] {
			return keys[1]
		}
		return keys[0]
	}
	setzen := func(key, wert, weg string) support.Snapshot {
		u := neuUmgebung(t, slog.LevelInfo, 0)
		var sp support.Span
		if weg == "StartSpan" {
			_, sp = u.a.StartSpan(context.Background(), "odata.hybrid.fallback", support.Str(key, wert), support.Str(anderer(key), andere[anderer(key)]))
		} else {
			_, sp = u.a.StartSpan(context.Background(), "odata.hybrid.fallback")
			sp.Set(support.Str(key, wert), support.Str(anderer(key), andere[anderer(key)]))
		}
		sp.End()
		return u.a.Snapshot()
	}
	zaehler := map[string]int{}
	pruefe := func(art string, werte []string, gueltig bool) {
		for _, key := range keys {
			for _, wert := range werte {
				for _, weg := range []string{"StartSpan", "Set"} {
					zaehler[art]++
					wo := fmt.Sprintf("%s=%q %s", key, wert, weg)
					s := setzen(key, wert, weg)
					if gueltig {
						pruefeVektor(t, wo, s, nil)
						if len(s.Spans) != 1 || len(s.Spans[0].Attrs) != 2 {
							t.Errorf("%s: Eintrag %+v, erwartet beide Attribute", wo, s.Spans)
						}
						continue
					}
					pruefeVektor(t, wo, s, map[support.Verwurf]uint64{support.VerwurfWert: 1, support.VerwurfPflicht: 1})
					if len(s.Spans) != 1 || !attrsGleich(s.Spans[0].Attrs, []support.Attr{support.Str(anderer(key), andere[anderer(key)])}) {
						t.Errorf("%s: Eintrag %+v, erwartet allein %s", wo, s.Spans, anderer(key))
					}
				}
			}
		}
	}
	pruefe("basis", basis, false)
	pruefe("rand", rand, false)
	pruefe("gut", gut, true)
	if zaehler["basis"] != 2*3*2 || zaehler["rand"] != 2*6*2 || zaehler["gut"] != 2*9*2 || len(basis) != 3 || len(rand) != 6 || len(gut) != 9 {
		t.Errorf("Zählprüfung: %v; erwartet 12, 24 und 36", zaehler)
	}
	// Paar von Verstößen am selben Span
	for _, weg := range []string{"StartSpan", "Set"} {
		u := neuUmgebung(t, slog.LevelInfo, 0)
		var sp support.Span
		schlecht := []support.Attr{support.Str(keys[0], "Replay"), support.Str(keys[1], "")}
		if weg == "StartSpan" {
			_, sp = u.a.StartSpan(context.Background(), "odata.hybrid.fallback", schlecht...)
		} else {
			_, sp = u.a.StartSpan(context.Background(), "odata.hybrid.fallback")
			sp.Set(schlecht...)
		}
		sp.End()
		s := u.a.Snapshot()
		pruefeVektor(t, "Paar "+weg, s, map[support.Verwurf]uint64{support.VerwurfWert: 2, support.VerwurfPflicht: 2})
		if len(s.Spans) != 1 || s.Spans[0].Attrs == nil || len(s.Spans[0].Attrs) != 0 {
			t.Errorf("Paar %s: Eintrag %+v, erwartet aufgezeichnet ohne Attribute (nicht nil)", weg, s.Spans)
		}
	}
}

// P3 (2): jeder Schlüssel ist allein am eigenen Span zulässig. Menge: 6 Spans × 13 Pflicht-Attribute
// gleich 78 Paare, 13 zulässig und 65 fremd (schluessel, das Attribut entfällt); das Orakel ist die
// Tabelle in §5 als Literal (spanTabelle). Zählprüfung: 6, 13 und 78.
func TestHybridSchluesselMatrix(t *testing.T) {
	tab := spanTabelle(t)
	keys := alleSpanKeys(t)
	paare, zulaessig, fremd := 0, 0, 0
	for _, z := range tab {
		for _, k := range keys {
			paare++
			u := neuUmgebung(t, slog.LevelInfo, 0)
			_, sp := u.a.StartSpan(context.Background(), support.SpanName(z.name), gueltigSpan(k))
			sp.End()
			s := u.a.Snapshot()
			erlaubt := slices.Contains(z.pflicht, k)
			if erlaubt {
				zulaessig++
				if s.Dropped[support.VerwurfSchluessel] != 0 || len(s.Spans) != 1 || !attrsGleich(s.Spans[0].Attrs, []support.Attr{gueltigSpan(k)}) {
					t.Errorf("%s an %s: schluessel %d, Eintrag %+v; erwartet zulässig", k, z.name, s.Dropped[support.VerwurfSchluessel], s.Spans)
				}
				continue
			}
			fremd++
			if s.Dropped[support.VerwurfSchluessel] != 1 || len(s.Spans) != 1 || len(s.Spans[0].Attrs) != 0 {
				t.Errorf("%s an %s: schluessel %d, Eintrag %+v; erwartet fremd (1, ohne Attribut)", k, z.name, s.Dropped[support.VerwurfSchluessel], s.Spans)
			}
		}
	}
	if len(tab) != 6 || len(keys) != 13 || paare != 78 || zulaessig != 13 || fremd != 65 {
		t.Errorf("Zählprüfung: %d Spans, %d Schlüssel, %d Paare, %d zulässig, %d fremd; erwartet 6, 13, 78, 13 und 65", len(tab), len(keys), paare, zulaessig, fremd)
	}
}

// P8: CountDiscarded verliert keinen Zuwachs unter gleichzeitigem Aufruf, in jedem Zweig und gemischt
// (SPEC-232 Nr. 5 a, SPEC-230 Nr. 11). Menge M-Zuwachs: 16 Goroutinen × 250 Runden × (4 Codes der
// Menge, je gefolgt von 10 der 40 Codes außerhalb: die 37 der Tabelle und 0, -1, 42, und einem Aufruf
// CountError(0)), Summe je Code der Menge 4000, fehlercode 176000 (160000 aus CountDiscarded und 16000
// aus CountError; Orakel: das Produkt der Konstanten, nicht die Zahl der Aufrufe im Code),
// mit 2 Lesern, die Snapshot rufen, solange die Schreiber laufen: kein Zähler fällt, keiner übersteigt
// seine Summe, und jeder Stand liest die Codes der Menge als einen Zeitpunkt (jeder Schreiber zählt sie
// in der Folge der Menge: Nachbarn liegen um höchstens 16 auseinander, der spätere nie darüber). Der
// Detektor von make test-race meldet das Fehlen der Sperre an der Zählung beider Zweige und am Lesen mit
// einer Schleife; die Summe zeigt den Verlust eines Zuwachses nicht in jedem Lauf (die Raten stehen im
// Plan).
func TestCountDiscardedGleichzeitig(t *testing.T) {
	const goroutinen, runden, leser, jeCode, ungueltig = 16, 250, 2, 4000, 176000
	menge := verworfeneOrakel(t)
	aussen := append(aussenCodes(t, menge), 0, -1, 42)
	if len(aussen) != 40 || goroutinen*runden != jeCode || goroutinen*runden*(len(aussen)+4) != ungueltig {
		t.Fatalf("%d Codes außerhalb, %d und %d Zählungen; erwartet 40, %d und %d", len(aussen), goroutinen*runden, goroutinen*runden*(len(aussen)+4), jeCode, ungueltig)
	}
	a, _ := neu(t)
	if !binnen(t, 60*time.Second, "TestCountDiscardedGleichzeitig", func() {
		start := make(chan struct{})
		var schreibend, lesend sync.WaitGroup
		var fertig atomic.Bool
		var fehler, gelesen atomic.Int64
		for range goroutinen {
			schreibend.Add(1)
			go func() {
				defer schreibend.Done()
				<-start
				for r := range runden {
					gib := r%8 == 0 // die Abgabe lässt Schreiber und Leser auch bei einem Prozessor ineinander laufen
					for i, c := range menge {
						a.CountDiscarded(c)
						if gib {
							runtime.Gosched()
						}
						for k, u := range aussen[10*i : 10*i+10] {
							a.CountDiscarded(u)
							if gib && k == 4 {
								runtime.Gosched()
							}
						}
						a.CountError(0) // zählt fehlercode unter mu: ein zweiter Schreiber derselben Karte für den Detektor
					}
				}
			}()
		}
		for range leser {
			lesend.Add(1)
			go func() {
				defer lesend.Done()
				<-start
				zuletzt, zuletztF := map[model.Code]uint64{}, uint64(0)
				for {
					s := a.Snapshot()
					var kette []uint64
					for _, c := range menge {
						n, ok := s.Verworfen[c]
						if !ok || n < zuletzt[c] || n > jeCode {
							fehler.Add(1)
						}
						zuletzt[c], kette = n, append(kette, n)
					}
					if f := s.Dropped[support.VerwurfFehlercode]; f < zuletztF || f > ungueltig {
						fehler.Add(1)
					} else {
						zuletztF = f
					}
					for i := 0; i+1 < len(kette); i++ {
						if kette[i+1] > kette[i] || kette[i] > kette[i+1]+goroutinen {
							fehler.Add(1) // ein Stand liest zwei Nachbarn zu verschiedenen Zeiten
						}
					}
					gelesen.Add(1)
					if fertig.Load() {
						return
					}
					runtime.Gosched()
				}
			}()
		}
		close(start)
		schreibend.Wait()
		fertig.Store(true)
		lesend.Wait()
		if n := fehler.Load(); n != 0 {
			t.Errorf("%d Lesungen mit fallendem Zähler, fehlendem Eintrag oder Zähler über der Summe", n)
		}
		if gelesen.Load() < leser {
			t.Errorf("%d Lesungen, erwartet mindestens %d", gelesen.Load(), leser)
		}
	}) {
		t.Fatal("TestCountDiscardedGleichzeitig endet nicht innerhalb von 60 s")
	}
	s := a.Snapshot()
	pruefeVektor(t, "gleichzeitig", s, map[support.Verwurf]uint64{support.VerwurfFehlercode: ungueltig})
	verworfenGleich(t, "gleichzeitig", s.Verworfen, menge, map[model.Code]uint64{menge[0]: jeCode, menge[1]: jeCode, menge[2]: jeCode, menge[3]: jeCode})
	if n := summe(s.Errors); n != 0 || len(menge) != 4 {
		t.Errorf("Errors zählt %d bei %d Codes; erwartet 0 und 4", n, len(menge))
	}
}

// P9: CountDiscarded nimmt die Sperre der Zeile nicht. Haltepunkt am Writer: A hängt in Write (die
// Senke meldet den Eintritt), danach ruft B CountDiscarded für jeden der 4 Codes der Menge und für
// einen Code außerhalb (0); jeder Aufruf kehrt innerhalb 2 s zurück, und die Zähler stehen im Stand,
// während A noch hängt. Nach der Freigabe kehrt A zurück, eine Zeile. Das Muster ist das von
// TestBlockierteSenke für CountError; A hängt im Write, bevor B läuft, so dass die Mutation in jedem
// Lauf rot wird.
func TestCountDiscardedBlockierteSenke(t *testing.T) {
	menge := verworfeneOrakel(t)
	senke := &FakeSenke{}
	senke.blockieren()
	t.Cleanup(senke.freigeben)
	a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
	fertigA := logInGoroutine(a, "a1")
	if !senke.warteEintritt(5 * time.Second) {
		t.Fatal("A erreicht Write nicht")
	}
	codes := append(slices.Clone(menge), 0)
	for _, c := range codes {
		if !binnen(t, 2*time.Second, fmt.Sprintf("CountDiscarded(%d)", int(c)), func() { a.CountDiscarded(c) }) {
			t.Fatalf("CountDiscarded(%d) wartet auf die blockierte Senke", int(c))
		}
	}
	s := a.Snapshot()
	verworfenGleich(t, "während A hängt", s.Verworfen, menge, map[model.Code]uint64{menge[0]: 1, menge[1]: 1, menge[2]: 1, menge[3]: 1})
	pruefeVektor(t, "während A hängt", s, map[support.Verwurf]uint64{support.VerwurfFehlercode: 1})
	if zurueck(fertigA, 100*time.Millisecond) {
		t.Fatal("A kehrt vor der Freigabe zurück")
	}
	senke.freigeben()
	if !zurueck(fertigA, 2*time.Second) || len(senke.zeilen()) != 1 {
		t.Errorf("A nach der Freigabe: %d Zeilen, erwartet 1", len(senke.zeilen()))
	}
	if len(codes) != 5 {
		t.Errorf("%d Codes geprüft, erwartet 5", len(codes))
	}
}
