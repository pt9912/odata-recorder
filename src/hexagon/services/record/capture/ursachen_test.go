package capture_test

import (
	"context"
	"errors"
	"io"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
)

// blockiertSrc liefert daten, hält danach den nächsten Read an und meldet seine Ankunft; der
// Read kehrt zurück, sobald frei geschlossen ist.
func blockiertSrc(t *testing.T, daten string) (s *skript, ankunft chan struct{}, freigeben func()) {
	t.Helper()
	return blockiertSrcMit(t, daten, "", errQuelle)
}

// blockiertSrcMit liefert nach der Freigabe danach und err.
func blockiertSrcMit(t *testing.T, daten, danach string, err error) (s *skript, ankunft chan struct{}, freigeben func()) {
	t.Helper()
	ankunft = make(chan struct{}, 1)
	frei := make(chan struct{})
	var einmal sync.Once
	freigeben = func() { einmal.Do(func() { close(frei) }) }
	t.Cleanup(freigeben)
	s = quelle(daten)
	s.haltVor(frei, ankunft, danach, err)
	return s, ankunft, freigeben
}

func warteAufAnkunft(t *testing.T, c <-chan struct{}) {
	t.Helper()
	select {
	case <-c:
	case <-time.After(zeitgrenze):
		t.Fatal("src wurde nicht erreicht")
	}
}

// leseInGoroutine liest a in einer Goroutine leer und meldet das Ende.
func leseInGoroutine(a *capture.Aufnahme) <-chan struct{} {
	ende := make(chan struct{})
	go func() {
		defer close(ende)
		_, _ = io.Copy(io.Discard, a)
	}()
	return ende
}

func TestAbbruchUrsachen(t *testing.T) {
	const praefix = `{"a":`
	// start bringt die Aufnahme in den Zustand, in dem die Ursache eintritt, und liefert die
	// Ursache, die errors.Is erreichen muss.
	faelle := []struct {
		name  string
		ursa  error
		start func(t *testing.T, u *umg) (*capture.Aufnahme, func())
	}{
		{"Fehler von src", errQuelle, func(t *testing.T, u *umg) (*capture.Aufnahme, func()) {
			a := u.tee(quelleMit(errQuelle, praefix), antwort())
			_, _ = io.Copy(io.Discard, a)
			return a, func() {}
		}},
		{"Abbrechen bei blockiertem src", context.Canceled, func(t *testing.T, u *umg) (*capture.Aufnahme, func()) {
			s, ankunft, _ := blockiertSrc(t, praefix)
			a := u.tee(s, antwort())
			leseInGoroutine(a)
			warteAufAnkunft(t, ankunft)
			return a, a.Abbrechen
		}},
		{"cancel des ctx", context.Canceled, func(t *testing.T, u *umg) (*capture.Aufnahme, func()) {
			s, ankunft, _ := blockiertSrc(t, praefix)
			ctx, cancel := context.WithCancel(context.Background())
			a, err := u.pipe().Tee(ctx, s, antwort())
			if err != nil {
				t.Fatal(err)
			}
			leseInGoroutine(a)
			warteAufAnkunft(t, ankunft)
			return a, cancel
		}},
		{"Deadline des Aufrufers", context.DeadlineExceeded, func(t *testing.T, u *umg) (*capture.Aufnahme, func()) {
			s, ankunft, _ := blockiertSrc(t, praefix)
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			t.Cleanup(cancel)
			a, err := u.pipe().Tee(ctx, s, antwort())
			if err != nil {
				t.Fatal(err)
			}
			leseInGoroutine(a)
			warteAufAnkunft(t, ankunft)
			return a, func() {}
		}},
		{"Abbrechen vor dem ersten Byte", context.Canceled, func(t *testing.T, u *umg) (*capture.Aufnahme, func()) {
			a := u.tee(quelle(), antwort())
			return a, a.Abbrechen
		}},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			u := neu(t)
			a, ausloeser := f.start(t, u)
			ausloeser()
			e := abwarten(t, a)
			verworfen(t, e, capture.Response, capture.Abbruch, capture.GrundAbbruch)
			if !errors.Is(e.Fehler, f.ursa) {
				t.Errorf("errors.Is(%v, %v) ist falsch", e.Fehler, f.ursa)
			}
			st := u.blobs.stand()
			if st.commits != 0 || st.aborts != st.writer || len(u.blobs.inhalte()) != 0 {
				t.Errorf("Writer: %+v, Blobs %v; erwartet Commit 0 und Abort je Writer", st, u.blobs.inhalte())
			}
		})
	}
}

func TestFehlerNachVollstaendigemBody(t *testing.T) {
	e := neu(t).lauf(quelleMit(errQuelle, gutBody), antwort())
	verworfen(t, e, capture.Response, capture.Abbruch, capture.GrundAbbruch)
	u := neu(t)
	vollstaendig(t, u, u.lauf(quelle(gutBody), antwort()), model.RedactionNone)
}

func TestVerstossVorFehler(t *testing.T) {
	verworfen(t, neu(t).lauf(quelleMit(errQuelle, `{"a":x`), antwort()), capture.Response, capture.Verstoss, capture.GrundSyntax)
	verworfen(t, neu(t).lauf(quelleMit(errQuelle, `{"a":1,`), antwort()), capture.Response, capture.Abbruch, capture.GrundAbbruch)
	// Daten und Fehler kommen im selben Read: die Daten gehen vor (SPEC-119 Nr. 9).
	zusammen := func(daten string) *skript {
		return &skript{schritte: []schritt{{daten: daten, err: errQuelle}}}
	}
	verworfen(t, neu(t).lauf(zusammen(`{"a":x`), antwort()), capture.Response, capture.Verstoss, capture.GrundSyntax)
	verworfen(t, neu(t).lauf(zusammen(`{"a":1,`), antwort()), capture.Response, capture.Abbruch, capture.GrundAbbruch)
}

func TestUeberlaufVorVerstoss(t *testing.T) {
	u := neu(t, mitQueue(50))
	frei := u.pipe().AnhaltenAmStart()
	a := u.tee(quelle(`{"a":x`, wiederholt(" ", 100)), antwort())
	_, _ = io.Copy(io.Discard, a)
	frei()
	verworfen(t, abwarten(t, a), capture.Response, capture.Verstoss, capture.GrundUeberlauf)
}

func TestFrist(t *testing.T) {
	faelle := []struct {
		name  string
		start func(t *testing.T, u *umg) *capture.Aufnahme
	}{
		{"src blockiert ohne Byte", func(t *testing.T, u *umg) *capture.Aufnahme {
			s, ankunft, _ := blockiertSrc(t, "")
			a := u.tee(s, antwort())
			leseInGoroutine(a)
			warteAufAnkunft(t, ankunft)
			return a
		}},
		{"src blockiert nach einem Präfix", func(t *testing.T, u *umg) *capture.Aufnahme {
			s, ankunft, _ := blockiertSrc(t, `{"a":"`)
			a := u.tee(s, antwort())
			leseInGoroutine(a)
			warteAufAnkunft(t, ankunft)
			return a
		}},
		{"Writer wartet auf den Kontext", func(t *testing.T, u *umg) *capture.Aufnahme {
			u.blobs.warteAufCtx = true
			a := u.tee(quelle(gutBody), antwort())
			leseInGoroutine(a)
			return a
		}},
		{"Begin wartet auf den Kontext", func(t *testing.T, u *umg) *capture.Aufnahme {
			u.blobs.warteBegin = true
			a := u.tee(quelle(gutBody), antwort())
			leseInGoroutine(a)
			return a
		}},
		{"Commit wartet auf den Kontext", func(t *testing.T, u *umg) *capture.Aufnahme {
			u.blobs.warteCommit = true
			a := u.tee(quelle(gutBody), antwort())
			leseInGoroutine(a)
			return a
		}},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			u := neu(t, mitFrist(30*time.Millisecond))
			a := f.start(t, u)
			e := abwarten(t, a)
			verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundFrist)
			if !errors.Is(e.Fehler, context.DeadlineExceeded) {
				t.Errorf("Ursache %v erreicht context.DeadlineExceeded nicht", e.Fehler)
			}
			z, err := capture.Abschluss(requestOK(t), e)
			if err != nil || z.Capture != model.CaptureFailed {
				t.Errorf("Abschluss = %+v, %v; erwartet failed", z, err)
			}
			st := u.blobs.stand()
			if len(u.blobs.inhalte()) != 0 || st.aborts != st.writer {
				t.Errorf("Writer: %+v, Blobs %v", st, u.blobs.inhalte())
			}
		})
	}
}

// requestOK ist das Ergebnis eines gelungenen Request-Bodys.
func requestOK(t *testing.T) capture.Ergebnis {
	t.Helper()
	return neu(t).lauf(quelle(gutBody), anfrage())
}

func TestErsteUrsacheGewinnt(t *testing.T) {
	t.Run("Abbrechen bei langer Frist", func(t *testing.T) {
		u := neu(t, mitFrist(time.Hour))
		s, ankunft, _ := blockiertSrc(t, "")
		a := u.tee(s, antwort())
		leseInGoroutine(a)
		warteAufAnkunft(t, ankunft)
		a.Abbrechen()
		verworfen(t, abwarten(t, a), capture.Response, capture.Abbruch, capture.GrundAbbruch)
	})
	t.Run("spätere Ursachen nach der Frist", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		u := neu(t, mitFrist(20*time.Millisecond))
		s, ankunft, _ := blockiertSrc(t, "")
		a, err := u.pipe().Tee(ctx, s, antwort())
		if err != nil {
			t.Fatal(err)
		}
		leseInGoroutine(a)
		warteAufAnkunft(t, ankunft)
		erst := abwarten(t, a)
		verworfen(t, erst, capture.Response, capture.Verstoss, capture.GrundFrist)
		a.Abbrechen()
		cancel()
		if nach := a.Ergebnis(); !reflect.DeepEqual(erst, nach) {
			t.Errorf("Ergebnis änderte sich: %+v -> %+v", erst, nach)
		}
	})
	// Die Frist ist so lang, dass sie in keinem Lauf vor Abbrechen abläuft; der Test wartet auf keine Zeit.
	t.Run("Abbrechen bei angehaltener Goroutine", func(t *testing.T) {
		u := neu(t, mitFrist(time.Hour))
		frei := u.pipe().AnhaltenAmStart()
		a := u.tee(quelle(), antwort())
		a.Abbrechen()
		frei()
		verworfen(t, abwarten(t, a), capture.Response, capture.Abbruch, capture.GrundAbbruch)
	})
	// Der Writer meldet das Ende des Kontexts auf ctxEnde und hält an: Abbrechen und cancel()
	// kommen nach der Frist und vor der Auswertung, ihre Ordnung bestimmt ein Kanal. Die Frist
	// ist 250 ms; der Test setzt voraus, dass die Goroutine in dieser Zeit den Writer erreicht.
	t.Run("Frist vor Abbrechen und Abbruch von ctx", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		u := neu(t, mitFrist(250*time.Millisecond))
		u.blobs.warteAufCtx, u.blobs.nachCtx = true, make(chan struct{})
		a, err := u.pipe().Tee(ctx, quelle(gutBody), antwort())
		if err != nil {
			t.Fatal(err)
		}
		leseInGoroutine(a)
		warteAufAnkunft(t, u.blobs.ctxEnde)
		a.Abbrechen()
		cancel()
		close(u.blobs.nachCtx)
		e := abwarten(t, a)
		verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundFrist)
		if !errors.Is(e.Fehler, context.DeadlineExceeded) {
			t.Errorf("Ursache %v ist nicht context.DeadlineExceeded", e.Fehler)
		}
	})
}

// TestFehlerBleibtNachUrsache hält fest, dass ein Fehler des Redaktors, der Nr. 7 und des
// Writers bleibt, was er ist, wenn danach eine Ursache außerhalb des Stroms eintritt (SPEC-119
// Nr. 9): der Writer hält an, nachdem der Fehler entschieden ist, die Ursache kommt, dann läuft er weiter.
func TestFehlerBleibtNachUrsache(t *testing.T) {
	ursachen := []struct {
		name string
		los  func(a *capture.Aufnahme, cancel func())
	}{
		{"Abbrechen", func(a *capture.Aufnahme, _ func()) { a.Abbrechen() }},
		{"cancel des ctx", func(_ *capture.Aufnahme, cancel func()) { cancel() }},
	}
	faelle := []struct {
		name    string
		enc     string
		roh     []byte
		aufbau  func(b *fakeBlobs)
		commit  bool
		grund   capture.Grund
		ursache error
	}{
		{name: "Verstoß des Redaktors", roh: []byte(`{"a":1,"b":x`), aufbau: func(b *fakeBlobs) { b.haltWrite = 1 },
			grund: capture.GrundSyntax, ursache: model.ErrUnsicher},
		{name: "Nachlauf", enc: "deflate", roh: append(zl([]byte(gutBody)), 'x'), aufbau: func(b *fakeBlobs) { b.haltWrite = 1 },
			grund: capture.GrundDekodierung},
		{name: "Fehler des Writers im Write", roh: []byte(gutBody), grund: capture.GrundAblage, ursache: errDisk,
			aufbau: func(b *fakeBlobs) { b.haltWrite, b.fehlerNachHalt, b.writeFehler[1] = 1, true, errDisk }},
		{name: "Fehler des Writers im Commit", roh: []byte(gutBody), commit: true, grund: capture.GrundAblage, ursache: errDisk,
			aufbau: func(b *fakeBlobs) { b.haltCommit, b.commitFehler = true, errDisk }},
	}
	for _, f := range faelle {
		for _, ur := range ursachen {
			t.Run(f.name+" und "+ur.name, func(t *testing.T) {
				u := neu(t)
				f.aufbau(u.blobs)
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				ein := antwort()
				if f.enc != "" {
					ein = mitEncoding(ein, f.enc)
				}
				a, err := u.pipe().Tee(ctx, quelleBytes(f.roh), ein)
				if err != nil {
					t.Fatal(err)
				}
				leseInGoroutine(a)
				im, frei := u.blobs.imWrite, u.blobs.freiWrite
				if f.commit {
					im, frei = u.blobs.imCommit, u.blobs.freiCommit
				}
				warteAufAnkunft(t, im)
				ur.los(a, cancel)
				close(frei)
				e := abwarten(t, a)
				verworfen(t, e, capture.Response, capture.Verstoss, f.grund)
				if f.ursache != nil && !errors.Is(e.Fehler, f.ursache) {
					t.Errorf("Ursache %v erreicht %v nicht", e.Fehler, f.ursache)
				}
			})
		}
	}
}

// aufnahmeGoroutinen zählt die Goroutinen, die Tee gestartet hat und die noch laufen.
func aufnahmeGoroutinen() int {
	buf := make([]byte, 1<<20)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, 2*len(buf))
	}
	return strings.Count(string(buf), "created by github.com/pt9912/odata-recorder/src/hexagon/services/record/capture.(*Pipeline).Tee")
}

// TestGoroutineEndetImmer hält fest, dass die Goroutine jeder Aufnahme endet und dass der Kontext,
// den der Writer bekam, mit dem Ergebnis beendet ist (SPEC-119 Nr. 14), für jede Zeile der Tabelle.
func TestGoroutineEndetImmer(t *testing.T) {
	t.Run("Gegenprobe: die Zählung sieht eine laufende Aufnahme", func(t *testing.T) {
		u := neu(t)
		vorher := aufnahmeGoroutinen()
		frei := u.pipe().AnhaltenAmStart()
		a := u.tee(quelle(), antwort())
		if n := aufnahmeGoroutinen(); n != vorher+1 {
			t.Errorf("Goroutinen der Aufnahmen %d -> %d, erwartet +1", vorher, n)
		}
		a.Abbrechen()
		frei()
		_ = abwarten(t, a)
		wartetBis(t, "die Goroutine der Aufnahme endet", func() bool { return aufnahmeGoroutinen() <= vorher })
	})
	gesamt := t
	for _, f := range captureFaelle() {
		t.Run(f.name, func(t *testing.T) {
			if gesamt.Failed() {
				t.Skip("ein früherer Fall ist rot; die Wartezeit je Fall ist die Zeitgrenze")
			}
			vorher := aufnahmeGoroutinen()
			u, _ := fuehreAus(t, f)
			wartetBis(t, "die Goroutine der Aufnahme endet", func() bool { return aufnahmeGoroutinen() <= vorher })
			for i, ctx := range u.blobs.kontexte() {
				if ctx.Err() == nil {
					t.Errorf("der Kontext von Begin %d ist nach dem Ergebnis nicht beendet", i+1)
				}
			}
		})
	}
}

func TestErgebnisStabil(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	u := neu(t)
	a, err := u.pipe().Tee(ctx, quelle(`{"password":"x","a":1}`), anfrage())
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, a)
	erst := abwarten(t, a)
	if erst.Redigiert == nil {
		t.Fatalf("Redigiert fehlt: %+v", erst)
	}
	var wg sync.WaitGroup
	ergebnisse := make([]capture.Ergebnis, 3)
	for i := range ergebnisse {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ergebnisse[i] = a.Ergebnis()
		}()
	}
	wg.Wait()
	cancel()
	a.Abbrechen()
	ergebnisse = append(ergebnisse, a.Ergebnis(), a.Ergebnis())
	for i, e := range ergebnisse {
		if !reflect.DeepEqual(e, erst) {
			t.Errorf("Ergebnis %d weicht ab: %+v", i, e)
		}
	}
}

func TestErgebnisWartet(t *testing.T) {
	t.Run("wartet auf das Ende von src", func(t *testing.T) {
		u := neu(t)
		warte, ankunft := make(chan struct{}), make(chan struct{}, 1)
		s := quelle(gutBody).haltVor(warte, ankunft, "", io.EOF)
		a := u.tee(s, antwort())
		leseInGoroutine(a)
		warteAufAnkunft(t, ankunft)
		got := make(chan capture.Ergebnis, 1)
		go func() { got <- a.Ergebnis() }()
		select {
		case <-got:
			t.Fatal("Ergebnis() kehrte vor dem Ende von src zurück")
		case <-time.After(50 * time.Millisecond):
		}
		close(warte)
		select {
		case e := <-got:
			if e.Ausgang != capture.Vollstaendig {
				t.Errorf("Ergebnis = %+v", e)
			}
		case <-time.After(zeitgrenze):
			t.Fatal("Ergebnis() kehrt nach dem Ende von src nicht zurück")
		}
	})
	t.Run("ohne Read und ohne Ende kehrt es mit der Frist zurück", func(t *testing.T) {
		u := neu(t, mitFrist(30*time.Millisecond))
		a := u.tee(quelle(gutBody), antwort())
		verworfen(t, abwarten(t, a), capture.Response, capture.Verstoss, capture.GrundFrist)
		st := u.blobs.stand()
		if st.aborts+st.commits != st.writer {
			t.Errorf("nicht alle Writer abgeschlossen: %+v", st)
		}
	})
}

func TestCommitUndAbbruch(t *testing.T) {
	t.Run("gelungener Commit bleibt", func(t *testing.T) {
		u := neu(t)
		u.blobs.haltCommit = true
		ctx, cancel := context.WithCancel(context.Background())
		a, err := u.pipe().Tee(ctx, quelle(gutBody), antwort())
		if err != nil {
			t.Fatal(err)
		}
		leseInGoroutine(a)
		warteAufAnkunft(t, u.blobs.imCommit)
		cancel()
		a.Abbrechen()
		close(u.blobs.freiCommit)
		e := abwarten(t, a)
		if got := vollstaendig(t, u, e, model.RedactionNone); got != gutBody {
			t.Errorf("Blob = %q", got)
		}
		if st := u.blobs.stand(); st.aborts != 0 {
			t.Errorf("Abort nach gelungenem Commit: %+v", st)
		}
	})
	t.Run("Ursache vor Commit verhindert ihn", func(t *testing.T) {
		u := neu(t)
		frei := u.pipe().AnhaltenAmStart()
		u.blobs.haltWrite = 1
		a := u.tee(quelle(`{}`), antwort())
		_, _ = io.Copy(io.Discard, a)
		frei()
		warteAufAnkunft(t, u.blobs.imWrite)
		a.Abbrechen()
		close(u.blobs.freiWrite)
		verworfen(t, abwarten(t, a), capture.Response, capture.Abbruch, capture.GrundAbbruch)
		if st := u.blobs.stand(); st.commits != 0 || st.aborts != 1 {
			t.Errorf("Writer: %+v; erwartet Commit 0 und Abort 1", st)
		}
	})
}

func TestAbbrechenBlockiertNicht(t *testing.T) {
	u := neu(t)
	s, ankunft, frei := blockiertSrcMit(t, `{"a":`, "1}", nil)
	a := u.tee(s, antwort())
	p := make([]byte, 16)
	if n, err := a.Read(p); n != 5 || err != nil {
		t.Fatalf("erster Read = %d, %v", n, err)
	}
	gelesen := make(chan string, 1)
	go func() {
		n, err := a.Read(p)
		gelesen <- string(p[:n]) + "|" + errString(err)
	}()
	warteAufAnkunft(t, ankunft)
	fertig := make(chan struct{})
	go func() {
		a.Abbrechen()
		e := a.Ergebnis()
		if e.Ausgang != capture.Abbruch {
			t.Errorf("Ergebnis = %+v", e)
		}
		close(fertig)
	}()
	select {
	case <-fertig:
	case <-time.After(zeitgrenze):
		t.Fatal("Abbrechen oder Ergebnis wartet auf src")
	}
	frei()
	select {
	case got := <-gelesen:
		if got != "1}|" {
			t.Errorf("Read nach Abbrechen lieferte %q", got)
		}
	case <-time.After(zeitgrenze):
		t.Fatal("Read kehrt nicht zurück")
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.Abbrechen()
			_ = a.Ergebnis()
		}()
	}
	a.Abbrechen()
	a.Abbrechen()
	wg.Wait()
	verworfen(t, a.Ergebnis(), capture.Response, capture.Abbruch, capture.GrundAbbruch)
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func TestWriterEinerGoroutine(t *testing.T) {
	t.Run("Abbrechen bei angehaltenem Write", func(t *testing.T) {
		u := neu(t)
		u.blobs.haltWrite = 1
		a := u.tee(quelle(`{"a":`), antwort())
		if _, err := a.Read(make([]byte, 16)); err != nil {
			t.Fatal(err)
		}
		warteAufAnkunft(t, u.blobs.imWrite)
		a.Abbrechen()
		if v := u.blobs.verstoesse(); len(v) != 0 {
			t.Errorf("Abbrechen berührte den Writer: %v", v)
		}
		close(u.blobs.freiWrite)
		verworfen(t, abwarten(t, a), capture.Response, capture.Abbruch, capture.GrundAbbruch)
	})
	t.Run("cancel bei angehaltenem Commit", func(t *testing.T) {
		u := neu(t)
		u.blobs.haltCommit = true
		ctx, cancel := context.WithCancel(context.Background())
		a, err := u.pipe().Tee(ctx, quelle(gutBody), antwort())
		if err != nil {
			t.Fatal(err)
		}
		leseInGoroutine(a)
		warteAufAnkunft(t, u.blobs.imCommit)
		cancel()
		if v := u.blobs.verstoesse(); len(v) != 0 {
			t.Errorf("cancel berührte den Writer: %v", v)
		}
		close(u.blobs.freiCommit)
		vollstaendig(t, u, abwarten(t, a), model.RedactionNone)
	})
}

func TestAufnahmenUnabhaengig(t *testing.T) {
	t.Run("Überlauf der einen", aufnahmenUeberlauf)
	t.Run("beide halten Bytes zugleich", aufnahmenHaltenZugleich)
	t.Run("zwei Pipelines mit denselben Blobs und Pfaden", zweiPipelines)
	t.Run("derselbe ctx, Abbrechen trifft nur die eine", derselbeCtx)
}

func aufnahmenUeberlauf(t *testing.T) {
	u := neu(t, mitQueue(100))
	frei := u.pipe().AnhaltenAmStart()
	a := u.tee(quelle(wiederholt("x", 200)), antwort())
	b := u.tee(quelle(jsonString(90)), antwort())
	pa, pb := make([]byte, 300), make([]byte, 300)
	_, _ = a.Read(pa)
	_, _ = b.Read(pb)
	if a.Ausstehend() != 0 || b.Ausstehend() != 90 {
		t.Errorf("Ausstehend A %d, B %d; erwartet 0 und 90", a.Ausstehend(), b.Ausstehend())
	}
	_, _ = b.Read(pb)
	frei()
	verworfen(t, abwarten(t, a), capture.Response, capture.Verstoss, capture.GrundUeberlauf)
	if got := vollstaendig(t, u, abwarten(t, b), model.RedactionNone); got != jsonString(90) {
		t.Errorf("Blob B = %q", got)
	}
	c := u.lauf(quelle(gutBody), antwort())
	vollstaendig(t, u, c, model.RedactionNone)
}

// aufnahmenHaltenZugleich: jede Aufnahme zählt ihr Wasser allein (SPEC-119 Nr. 14); zusammen
// liegen 180 Bytes in den Queues, QueueBytes ist 100.
func aufnahmenHaltenZugleich(t *testing.T) {
	u := neu(t, mitQueue(100))
	frei := u.pipe().AnhaltenAmStart()
	a := u.tee(quelle(jsonString(90)), antwort())
	b := u.tee(quelle(jsonString(90)), antwort())
	_, _ = io.Copy(io.Discard, a)
	_, _ = io.Copy(io.Discard, b)
	if a.Ausstehend() != 90 || b.Ausstehend() != 90 {
		t.Errorf("Ausstehend A %d, B %d; erwartet 90 und 90", a.Ausstehend(), b.Ausstehend())
	}
	frei()
	for _, x := range []*capture.Aufnahme{a, b} {
		if got := vollstaendig(t, u, abwarten(t, x), model.RedactionNone); got != jsonString(90) {
			t.Errorf("Blob = %q", got)
		}
	}
}

// zweiPipelines: zwei Pipelines mit denselben Blobs und denselben Pfaden teilen nichts; derselbe
// Body ergibt dasselbe Ergebnis, und die Queues beider halten zugleich je 90 Bytes bei QueueBytes 100.
func zweiPipelines(t *testing.T) {
	u := neu(t, mitQueue(100))
	var ps [2]*capture.Pipeline
	var frei [2]func()
	for i := range ps {
		p, err := capture.New(u.k)
		if err != nil {
			t.Fatal(err)
		}
		ps[i], frei[i] = p, p.AnhaltenAmStart()
	}
	body := `{"password":"` + geheim + `","a":"` + wiederholt("y", 50) + `"}`
	var as [2]*capture.Aufnahme
	for i, p := range ps {
		a, err := p.Tee(context.Background(), quelle(body), anfrage())
		if err != nil {
			t.Fatal(err)
		}
		as[i] = a
		_, _ = io.Copy(io.Discard, a)
	}
	if as[0].Ausstehend() != int64(len(body)) || as[1].Ausstehend() != int64(len(body)) {
		t.Errorf("Ausstehend %d und %d; erwartet je %d", as[0].Ausstehend(), as[1].Ausstehend(), len(body))
	}
	frei[0]()
	frei[1]()
	e0, e1 := abwarten(t, as[0]), abwarten(t, as[1])
	want := `{"password":"[REDACTED]","a":"` + wiederholt("y", 50) + `"}`
	if got := vollstaendig(t, u, e0, model.RedactionApplied); got != want {
		t.Errorf("Blob der ersten = %q", got)
	}
	if !reflect.DeepEqual(e0, e1) {
		t.Errorf("gleicher Body, verschiedene Ergebnisse: %+v und %+v", e0, e1)
	}
}

// derselbeCtx: Request und Response eines Austauschs teilen den ctx des Aufrufers und sonst nichts.
func derselbeCtx(t *testing.T) {
	u := neu(t)
	ctx := context.Background()
	frei := u.pipe().AnhaltenAmStart()
	req, err := u.pipe().Tee(ctx, quelle(gutBody), anfrage())
	if err != nil {
		t.Fatal(err)
	}
	resp, err := u.pipe().Tee(ctx, quelle(gutBody), antwort())
	if err != nil {
		t.Fatal(err)
	}
	req.Abbrechen()
	_, _ = io.Copy(io.Discard, resp)
	frei()
	verworfen(t, abwarten(t, req), capture.Request, capture.Abbruch, capture.GrundAbbruch)
	vollstaendig(t, u, abwarten(t, resp), model.RedactionNone)
}
