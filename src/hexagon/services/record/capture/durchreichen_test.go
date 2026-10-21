package capture_test

import (
	"bytes"
	"errors"
	"io"
	"runtime"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
)

// jsonString ist ein gültiger JSON-Body der Länge n (n ab 8): ein Objekt mit einem String.
func jsonString(n int) string {
	return `{"a":"` + wiederholt("x", n-8) + `"}`
}

func TestDurchreichen(t *testing.T) {
	folge := []schritt{
		{daten: "a"},
		{daten: "bcdefgh"},
		{daten: wiederholt("i", 4096)},
		{daten: "jkl"},
		{daten: "m", err: io.EOF},
		{daten: "n", err: errQuelle},
		{daten: ""},
	}
	zustaende := []struct {
		name string
		vor  *schritt // Schritt, der die Aufnahme in den Zustand bringt; nil: laufend
		opt  []func(*capture.Konfig)
		tut  func(a *capture.Aufnahme)
	}{
		{name: "laufend"},
		{name: "nach Verstoß", vor: &schritt{daten: "x"}},
		{name: "nach Überlauf", vor: &schritt{daten: wiederholt("y", 100)}, opt: []func(*capture.Konfig){mitQueue(10)}},
		{name: "nach Abbrechen", vor: &schritt{daten: "{"}, tut: func(a *capture.Aufnahme) { a.Abbrechen() }},
		{name: "nach dem Ergebnis", vor: &schritt{daten: "{}", err: io.EOF}},
	}
	for _, z := range zustaende {
		t.Run(z.name, func(t *testing.T) {
			u := neu(t, z.opt...)
			s := &skript{}
			if z.vor != nil {
				s.schritte = append(s.schritte, *z.vor)
			}
			s.schritte = append(s.schritte, folge...)
			var frei func()
			if z.vor == nil {
				frei = u.pipe().AnhaltenAmStart()
			}
			a := u.tee(s, antwort())
			p := make([]byte, 8192)
			if z.vor != nil {
				if _, err := a.Read(p); err != nil && !errors.Is(err, io.EOF) {
					t.Fatalf("Vorlauf: %v", err)
				}
				if z.tut != nil {
					z.tut(a)
				}
				_ = abwarten(t, a)
			}
			for i, w := range folge {
				vorher := s.anzahlReads()
				n, err := a.Read(p)
				if n != len(w.daten) || string(p[:n]) != w.daten || err != w.err {
					t.Errorf("Read %d = %d, %q, %v; erwartet %d, %q, %v", i, n, p[:min(n, 20)], err, len(w.daten), w.daten[:min(len(w.daten), 20)], w.err)
				}
				if d := s.anzahlReads() - vorher; d != 1 {
					t.Errorf("Read %d rief src.Read %d-mal", i, d)
				}
			}
			if frei != nil {
				a.Abbrechen()
				frei()
				_ = abwarten(t, a)
			}
		})
	}
}

func TestReadWartetNichtAufSchreiber(t *testing.T) {
	u := neu(t, mitQueue(64<<10), mitMax(4<<20))
	u.blobs.haltWrite = 1
	body := jsonString(1 << 20)
	src := zerlegt([]byte(body), 4096)
	a := u.tee(src, antwort())
	p := make([]byte, 4096)
	if _, err := a.Read(p); err != nil {
		t.Fatalf("erster Read: %v", err)
	}
	select {
	case <-u.blobs.imWrite:
	case <-time.After(zeitgrenze):
		t.Fatal("der Writer kam nicht in Write")
	}
	fertig := make(chan error, 1)
	go func() {
		_, err := io.Copy(io.Discard, a)
		fertig <- err
	}()
	select {
	case err := <-fertig:
		if err != nil {
			t.Fatalf("Read: %v", err)
		}
	case <-time.After(zeitgrenze):
		t.Fatal("Read wartet auf den blockierten Schreiber")
	}
	close(u.blobs.freiWrite)
	verworfen(t, abwarten(t, a), capture.Response, capture.Verstoss, capture.GrundUeberlauf)
}

func TestReadBehaeltPNicht(t *testing.T) {
	u := neu(t)
	frei := u.pipe().AnhaltenAmStart()
	body := `{"k":"abcdefghij"}`
	a := u.tee(quelle(body), antwort())
	p := make([]byte, 5)
	var gelesen []byte
	for {
		n, err := a.Read(p)
		gelesen = append(gelesen, p[:n]...)
		for i := range p {
			p[i] = 0xFF
		}
		if err != nil {
			break
		}
	}
	if string(gelesen) != body {
		t.Fatalf("Read lieferte %q", gelesen)
	}
	frei()
	e := abwarten(t, a)
	if got := vollstaendig(t, u, e, model.RedactionNone); got != body {
		t.Errorf("Blob = %q, erwartet %q", got, body)
	}
}

func TestQueueGrenze(t *testing.T) {
	faelle := []struct {
		name   string
		body   string
		reads  []int // Längen der Reads; 0 ist ein leerer Read
		ueber  bool
		warten int64 // Ausstehend, solange die Goroutine angehalten ist
	}{
		{"60 und 40", jsonString(100), []int{60, 40}, false, 100},
		{"60 und 41", jsonString(101), []int{60, 41}, true, 0},
		{"ein Read mit 100", jsonString(100), []int{100}, false, 100},
		{"ein Read mit 101", jsonString(101), []int{101}, true, 0},
		{"leere Reads dazwischen", jsonString(100), []int{60, 0, 0, 40}, false, 100},
		{"leere Reads und 101", jsonString(101), []int{60, 0, 41}, true, 0},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			u := neu(t, mitQueue(100))
			frei := u.pipe().AnhaltenAmStart()
			s := &skript{}
			rest := f.body
			for _, n := range f.reads {
				s.schritte = append(s.schritte, schritt{daten: rest[:n]})
				rest = rest[n:]
			}
			a := u.tee(s, antwort())
			p := make([]byte, 200)
			for range f.reads {
				if _, err := a.Read(p); err != nil {
					t.Fatalf("Read: %v", err)
				}
			}
			if got := a.Ausstehend(); got != f.warten {
				t.Errorf("Ausstehend = %d, erwartet %d", got, f.warten)
			}
			_, _ = a.Read(p)
			frei()
			e := abwarten(t, a)
			if f.ueber {
				verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundUeberlauf)
			} else if got := vollstaendig(t, u, e, model.RedactionNone); got != f.body {
				t.Errorf("Blob = %q", got)
			}
		})
	}
}

// wartetBis wartet, bis f wahr ist, höchstens bis zur Zeitgrenze.
func wartetBis(t *testing.T, was string, f func() bool) {
	t.Helper()
	ende := time.Now().Add(zeitgrenze)
	for !f() {
		if time.Now().After(ende) {
			t.Fatalf("Zeitgrenze beim Warten auf %s", was)
		}
		runtime.Gosched()
		time.Sleep(50 * time.Microsecond)
	}
}

func TestQueueHoechststand(t *testing.T) {
	u := neu(t, mitQueue(5000))
	u.blobs.haltWrite = 1
	teile := []string{`{"a":"` + wiederholt("x", 994)}
	for range 39 {
		teile = append(teile, wiederholt("x", 1000))
	}
	a := u.tee(quelle(teile...), antwort())
	p := make([]byte, 1000)
	if _, err := a.Read(p); err != nil {
		t.Fatal(err)
	}
	select {
	case <-u.blobs.imWrite:
	case <-time.After(zeitgrenze):
		t.Fatal("der Writer kam nicht in Write")
	}
	for range 39 {
		if _, err := a.Read(p); err != nil {
			t.Fatal(err)
		}
	}
	if h := a.Hoechststand(); h != 5000 {
		t.Errorf("Hoechststand = %d, erwartet 5000 (5 Reads zu 1000 passen, der 6. läuft über)", h)
	}
	if n := a.Ausstehend(); n != 0 {
		t.Errorf("Ausstehend nach dem Überlauf = %d, erwartet 0", n)
	}
	close(u.blobs.freiWrite)
	verworfen(t, abwarten(t, a), capture.Response, capture.Verstoss, capture.GrundUeberlauf)
	if n := a.Ausstehend(); n != 0 {
		t.Errorf("Ausstehend nach dem Ergebnis = %d", n)
	}
}

func TestKeinUeberlaufBeiSchnellemKonsument(t *testing.T) {
	u := neu(t, mitQueue(4096), mitMax(4<<20))
	body := jsonString(1 << 20)
	a := u.tee(zerlegt([]byte(body), 1000), antwort())
	p := make([]byte, 1000)
	for {
		n, err := a.Read(p)
		wartetBis(t, "Ausstehend == 0", func() bool { return a.Ausstehend() == 0 })
		if err != nil || n == 0 {
			break
		}
	}
	e := abwarten(t, a)
	if got := vollstaendig(t, u, e, model.RedactionNone); got != body {
		t.Errorf("Blob weicht vom Body ab (Länge %d, erwartet %d)", len(got), len(body))
	}
	if h := a.Hoechststand(); h > 1000 {
		t.Errorf("Hoechststand = %d, erwartet höchstens 1000", h)
	}
}

// unendlich liefert erst ein Präfix und danach beliebig viele Bytes, ohne zu allozieren.
type unendlich struct{ erst string }

func (s *unendlich) Read(p []byte) (int, error) {
	if s.erst != "" {
		n := copy(p, s.erst)
		s.erst = s.erst[n:]
		return n, nil
	}
	return len(p), nil
}

func TestNachEndeKeineKopie(t *testing.T) {
	zustaende := []struct {
		name string
		erst string
		opt  []func(*capture.Konfig)
		aus  capture.Grund
	}{
		{"nach Verstoß", "x", nil, capture.GrundSyntax},
		{"nach Überlauf", wiederholt("y", 100), []func(*capture.Konfig){mitQueue(10)}, capture.GrundUeberlauf},
	}
	for _, z := range zustaende {
		t.Run(z.name, func(t *testing.T) {
			u := neu(t, z.opt...)
			a := u.tee(&unendlich{erst: z.erst}, antwort())
			p := make([]byte, 1<<20)
			if _, err := a.Read(p); err != nil {
				t.Fatal(err)
			}
			e := abwarten(t, a)
			verworfen(t, e, capture.Response, capture.Verstoss, z.aus)
			h := a.Hoechststand()
			for range 4 {
				_, _ = a.Read(p)
			}
			if a.Hoechststand() != h || a.Ausstehend() != 0 {
				t.Errorf("Read kopiert nach dem Ergebnis: Hoechststand %d -> %d, Ausstehend %d", h, a.Hoechststand(), a.Ausstehend())
			}
			if n := testing.AllocsPerRun(50, func() { _, _ = a.Read(p) }); n != 0 {
				t.Errorf("Read allozierte %v-mal je Aufruf nach dem Ergebnis", n)
			}
		})
	}
}

func TestSpaetereReadsNurDurch(t *testing.T) {
	u := neu(t)
	s := &skript{schritte: []schritt{{daten: "{}", err: io.EOF}, {daten: "x"}}}
	a := u.tee(s, antwort())
	p := make([]byte, 16)
	if n, err := a.Read(p); n != 2 || err != io.EOF {
		t.Fatalf("erster Read = %d, %v", n, err)
	}
	e := abwarten(t, a)
	if got := vollstaendig(t, u, e, model.RedactionNone); got != "{}" {
		t.Errorf("Blob = %q", got)
	}
	if n, err := a.Read(p); n != 1 || p[0] != 'x' || err != nil {
		t.Errorf("späterer Read = %d, %v", n, err)
	}
	if n := a.Ausstehend(); n != 0 {
		t.Errorf("Ausstehend = %d", n)
	}
	if e2 := a.Ergebnis(); !bytes.Equal([]byte(e2.Blob), []byte(e.Blob)) || e2.Ausgang != capture.Vollstaendig {
		t.Errorf("Ergebnis änderte sich: %+v", e2)
	}
}

func TestNachEndeVorDemErgebnis(t *testing.T) {
	faelle := []struct {
		name  string
		erst  schritt
		blob  string
		grund capture.Grund
	}{
		{"nach io.EOF", schritt{daten: "{}", err: io.EOF}, "{}", ""},
		{"nach einem Fehler von src", schritt{daten: `{"a":1}`, err: errQuelle}, "", capture.GrundAbbruch},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			u := neu(t)
			frei := u.pipe().AnhaltenAmStart()
			s := &skript{schritte: []schritt{f.erst, {daten: "x"}, {daten: "yz"}}}
			a := u.tee(s, antwort())
			p := make([]byte, 16)
			for range 3 {
				_, _ = a.Read(p)
			}
			if n, want := a.Ausstehend(), int64(len(f.erst.daten)); n != want {
				t.Errorf("Ausstehend = %d, erwartet %d: Reads nach dem Ende legen nichts ein", n, want)
			}
			frei()
			e := abwarten(t, a)
			if f.grund != "" {
				verworfen(t, e, capture.Response, capture.Abbruch, f.grund)
			} else if got := vollstaendig(t, u, e, model.RedactionNone); got != f.blob {
				t.Errorf("Blob = %q", got)
			}
		})
	}
}

func TestRohgroesse(t *testing.T) {
	leereGlieder := string(gz(""))
	faelle := []struct {
		name  string
		max   int64
		ein   capture.Eingang
		roh   string
		grund capture.Grund // leer: Vollstaendig
	}{
		{"genau MaxBytes", 20, antwort(), jsonString(20), ""},
		{"ein Byte mehr", 20, antwort(), jsonString(20) + "\n", capture.GrundZuGross},
		{"gzip genau MaxBytes roh", int64(len(gz("{}"))), mitEncoding(antwort(), "gzip"), string(gz("{}")), ""},
		{"gzip ein Byte über MaxBytes roh", int64(len(gz("{}")) - 1), mitEncoding(antwort(), "gzip"), string(gz("{}")), capture.GrundZuGross},
		{"1000 leere gzip-Glieder", 500, mitEncoding(antwort(), "gzip"), wiederholt(leereGlieder, 1000), capture.GrundZuGross},
		{"Verstoß vor der Rohgröße", 20, antwort(), "[1,x" + wiederholt(" ", 96), capture.GrundSyntax},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			u := neu(t, mitMax(f.max))
			e := u.lauf(quelle(f.roh), f.ein)
			if f.grund == "" {
				vollstaendig(t, u, e, model.RedactionNone)
				return
			}
			verworfen(t, e, capture.Response, capture.Verstoss, f.grund)
		})
	}
}

// TestUeberlaufGegenRohgroesse hält fest, dass der Überlauf die eingelegten Bytes misst, also
// die auf MaxBytes gekürzten, und dass ein Read ohne eingelegtes Byte nie überläuft (SPEC-119 Nr. 4, 5).
func TestUeberlaufGegenRohgroesse(t *testing.T) {
	faelle := []struct {
		name  string
		queue int64
		reads []int
		grund capture.Grund
	}{
		{"ein Read, eingelegt 10 gleich Queue", 10, []int{12}, capture.GrundZuGross},
		{"ein Read, eingelegt 10 über Queue", 9, []int{12}, capture.GrundUeberlauf},
		{"zwei Reads, Summe 10 gleich Queue", 10, []int{6, 8}, capture.GrundZuGross},
		{"zwei Reads, Summe 10 über Queue", 9, []int{6, 8}, capture.GrundUeberlauf},
		{"zweiter Read ohne eingelegtes Byte", 10, []int{10, 5}, capture.GrundZuGross},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			u := neu(t, mitMax(10), mitQueue(f.queue))
			frei := u.pipe().AnhaltenAmStart()
			gesamt := 0
			for _, n := range f.reads {
				gesamt += n
			}
			rest := "[" + wiederholt(" ", gesamt-1)
			s := &skript{}
			for _, n := range f.reads {
				s.schritte = append(s.schritte, schritt{daten: rest[:n]})
				rest = rest[n:]
			}
			a := u.tee(s, antwort())
			p := make([]byte, 64)
			for range f.reads {
				if _, err := a.Read(p); err != nil {
					t.Fatalf("Read: %v", err)
				}
			}
			_, _ = a.Read(p)
			frei()
			verworfen(t, abwarten(t, a), capture.Response, capture.Verstoss, f.grund)
		})
	}
}
