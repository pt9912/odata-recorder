package telemetry_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/adapters/driven/telemetry"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// FakeSenke ist die Senke der Tests (ein io.Writer): sie hält jedes Argument von Write als Kopie und
// als Verweis, zählt jeden Write, hält den Höchstwert gleichzeitiger Write, liefert je Aufruf eine
// Rückgabe (n, err) und blockiert auf Wunsch (meldet den Eintritt, wartet auf die Freigabe).
type FakeSenke struct {
	mu                    sync.Mutex
	kopien, verweise      [][]byte
	aufrufe               int
	gleichzeitig, hoechst int
	antwort               func(p []byte) (int, error)
	eintritt              chan struct{}
	freigabe              chan struct{}
	frei                  sync.Once
}

// Write erfüllt io.Writer; ohne antwort nimmt es alle Bytes an.
func (s *FakeSenke) Write(p []byte) (int, error) {
	s.mu.Lock()
	s.aufrufe++
	s.gleichzeitig++
	s.hoechst = max(s.hoechst, s.gleichzeitig)
	s.kopien = append(s.kopien, bytes.Clone(p))
	s.verweise = append(s.verweise, p)
	antwort, eintritt, freigabe := s.antwort, s.eintritt, s.freigabe
	s.mu.Unlock()
	runtime.Gosched() // weitet das Fenster zweier gleichzeitiger Write; es verdeckt keines
	if eintritt != nil {
		eintritt <- struct{}{}
	}
	if freigabe != nil {
		<-freigabe
	}
	n, err := len(p), error(nil)
	if antwort != nil {
		n, err = antwort(p)
	}
	s.mu.Lock()
	s.gleichzeitig--
	s.mu.Unlock()
	return n, err
}

// blockieren lässt jedes folgende Write den Eintritt melden und auf freigeben warten.
func (s *FakeSenke) blockieren() {
	s.eintritt = make(chan struct{}, 4096)
	s.freigabe = make(chan struct{})
}

func (s *FakeSenke) freigeben() { s.frei.Do(func() { close(s.freigabe) }) }

// warteEintritt wartet, bis ein Write eintritt; false, wenn d verstreicht.
func (s *FakeSenke) warteEintritt(d time.Duration) bool {
	select {
	case <-s.eintritt:
		return true
	case <-time.After(d):
		return false
	}
}

func (s *FakeSenke) zeilen() [][]byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.kopien)
}

func (s *FakeSenke) anzahl() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.aufrufe
}

func (s *FakeSenke) hoechstwert() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hoechst
}

// bytesAlle sind die Bytes aller Write aneinander.
func (s *FakeSenke) bytesAlle() []byte { return bytes.Join(s.zeilen(), nil) }

// zaehlUhr liefert t und zählt die Aufrufe von Now.
type zaehlUhr struct {
	t time.Time
	n atomic.Int64
}

func (u *zaehlUhr) Now() time.Time {
	u.n.Add(1)
	return u.t
}

func festeUhr() *zaehlUhr {
	return &zaehlUhr{t: time.Date(2026, 10, 30, 12, 34, 56, 123456789, time.UTC)}
}

const festerStempel = "2026-10-30T12:34:56.123456789Z"

// RueckrufUhr ruft beim ersten Now den Rückruf auf und liefert t.
type RueckrufUhr struct {
	t        time.Time
	gerufen  atomic.Bool
	rueckruf func()
}

// Now ruft den Rückruf höchstens einmal, auch bei einem Aufruf aus dem Rückruf selbst.
func (u *RueckrufUhr) Now() time.Time {
	if u.gerufen.CompareAndSwap(false, true) && u.rueckruf != nil {
		u.rueckruf()
	}
	return u.t
}

func neuMit(t *testing.T, w io.Writer, uhr support.Clock, mindest slog.Level) *telemetry.Adapter {
	t.Helper()
	a, err := telemetry.New(telemetry.Config{Writer: w, Clock: uhr, MinLevel: mindest})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return a
}

// binnen führt f in einer Goroutine aus und meldet, ob sie innerhalb d zurückkehrt; eine Panik
// färbt den Test mit dem Namen wo.
func binnen(t *testing.T, d time.Duration, wo string, f func()) bool {
	t.Helper()
	fertig := make(chan struct{})
	go func() {
		defer close(fertig)
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("%s: Panik %v", wo, r)
			}
		}()
		f()
	}()
	select {
	case <-fertig:
		return true
	case <-time.After(d):
		return false
	}
}

// gelesen ist eine Zeile, gelesen von json.Decoder: Schlüssel in der Folge der Datei, Werte als Text,
// Zahlen als ihr Text im Strom.
type gelesen struct {
	schluessel []string
	text       map[string]string
	zahl       map[string]bool
}

// leseZeile liest eine Zeile mit eigenem Leser; ein Fehler nennt, was an der Form bricht.
func leseZeile(zeile []byte) (gelesen, error) {
	g := gelesen{text: map[string]string{}, zahl: map[string]bool{}}
	if !bytes.HasSuffix(zeile, []byte("\n")) || bytes.Count(zeile, []byte("\n")) != 1 {
		return g, errors.New("kein einzelnes Zeilenende am Schluss")
	}
	d := json.NewDecoder(bytes.NewReader(zeile))
	d.UseNumber()
	if tok, err := d.Token(); err != nil || tok != json.Delim('{') {
		return g, fmt.Errorf("kein Objekt: %v %v", tok, err)
	}
	for d.More() {
		k, err := d.Token()
		if err != nil {
			return g, err
		}
		key, ok := k.(string)
		if !ok {
			return g, fmt.Errorf("Schlüssel %v", k)
		}
		v, err := d.Token()
		if err != nil {
			return g, err
		}
		g.schluessel = append(g.schluessel, key)
		switch w := v.(type) {
		case string:
			g.text[key] = w
		case json.Number:
			g.text[key], g.zahl[key] = w.String(), true
		default:
			return g, fmt.Errorf("Wert von %s: %v", key, v)
		}
	}
	if tok, err := d.Token(); err != nil || tok != json.Delim('}') {
		return g, fmt.Errorf("Objekt nicht geschlossen: %v %v", tok, err)
	}
	if _, err := d.Token(); !errors.Is(err, io.EOF) {
		return g, fmt.Errorf("Rest hinter dem Objekt: %v", err)
	}
	return g, nil
}

func lies(t *testing.T, zeile []byte) gelesen {
	t.Helper()
	g, err := leseZeile(zeile)
	if err != nil {
		t.Fatalf("Zeile %q: %v", zeile, err)
	}
	return g
}

// erwarteterText ist der Text eines Werts in der Zeile: Text, Ganzzahl oder drei Nachkommastellen.
func erwarteterText(a support.Attr) string {
	switch a.Art {
	case support.ArtInt:
		return strconv.FormatInt(a.Ganz, 10)
	case support.ArtFloat:
		return strconv.FormatFloat(a.Gleit, 'f', 3, 64)
	}
	return a.Text
}

// orakelFolge sind die 13 Felder der Menge Log in der Reihenfolge von SPEC-231 Nr. 1.
func orakelFolge() []string {
	return []string{
		"mode", "service_id", "session_id", "exchange_id", "workspace_id", "correlation_id",
		"error_code", "duration_ms", "ereignis", "richtung", "grund", "blob", "befunde",
	}
}

// gueltigFuer liefert einen gültigen Wert für key; für einen Schlüssel ohne Zeile in der Tabelle
// einen Text, den die Textformen annehmen.
func gueltigFuer(key string) support.Attr {
	for _, z := range katalog() {
		if z.key == key {
			return gueltigerWert(z)
		}
	}
	return support.Str(key, "record")
}

// andererWert ist ein zweiter gültiger Wert für den Schlüssel von z, verschieden von gueltigerWert.
func andererWert(z zeile) support.Attr {
	switch z.form {
	case "kennung":
		return support.Str(z.key, "ses-0002")
	case "fehlercode":
		return support.Str(z.key, "ODR-E2101")
	case "dauer":
		return support.Millis(z.key, 2500*time.Microsecond)
	case "blob":
		return support.Str(z.key, "sha256:"+strings.Repeat("b", 64))
	case "zahl":
		return support.Int(z.key, 7)
	}
	return support.Str(z.key, "replay")
}

// logSchluessel liefert die Orakel-Liste der Felder der Zeile und prüft ihre Zahl gegen die Quelle:
// aus den Kandidaten (die Liste, die elf Pflicht-Attribute der Spans, zwölf fremde Schlüssel) nimmt
// der Filter mit einem gültigen Wert genau die Liste an.
func logSchluessel(t *testing.T) []string {
	t.Helper()
	folge := orakelFolge()
	kandidaten := slices.Clone(folge)
	for _, s := range support.Spans() {
		kandidaten = append(kandidaten, s.Pflicht...)
	}
	kandidaten = append(kandidaten, "", "x", "timestamp", "level", "msg", "Mode", "mode ", "a b", "ODR-E2001", "senke", "wert", "fehlercode")
	angenommen := 0
	for _, k := range kandidaten {
		if _, v := telemetry.Filtern("", []support.Attr{gueltigFuer(k)}); len(v) == 0 {
			angenommen++
			if !slices.Contains(folge, k) {
				t.Errorf("der Filter nimmt %q an, die Orakel-Liste nennt ihn nicht", k)
			}
		} else if slices.Contains(folge, k) {
			t.Errorf("der Filter weist %q ab, die Orakel-Liste nennt ihn", k)
		}
	}
	if angenommen != len(folge) || len(kandidaten) != len(folge)+11+12 {
		t.Fatalf("%d angenommen, %d Kandidaten; erwartet %d und %d", angenommen, len(kandidaten), len(folge), len(folge)+11+12)
	}
	return folge
}

// modellMsg ist das Modell von SPEC-231 Nr. 3: je ungültigem Byte ein U+FFFD, dann der längste
// Anfang bis 256 Bytes an einer Zeichengrenze.
func modellMsg(roh string) string {
	var voll []byte
	for i := 0; i < len(roh); {
		r, n := utf8.DecodeRuneInString(roh[i:])
		voll = utf8.AppendRune(voll, r)
		i += n
	}
	for len(voll) > 256 {
		_, n := utf8.DecodeLastRune(voll)
		voll = voll[:len(voll)-n]
	}
	return string(voll)
}

// modellStufe ist die Tabelle von SPEC-231 Nr. 2: DEBUG unter 0, INFO 0 bis 3, WARN 4 bis 7, ERROR ab 8.
func modellStufe(l int) string {
	switch {
	case l < 0:
		return "DEBUG"
	case l < 4:
		return "INFO"
	case l < 8:
		return "WARN"
	}
	return "ERROR"
}
