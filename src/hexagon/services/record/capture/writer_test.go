package capture_test

import (
	"context"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
)

func TestRedaktorAnbindung(t *testing.T) {
	const pw = `{"password":"S3CR3T-7f","a":1}`
	t.Run("Pfade und ContentType", func(t *testing.T) {
		u := neu(t)
		e := u.lauf(quelle(pw), anfrage())
		if got := vollstaendig(t, u, e, model.RedactionApplied); got != `{"password":"[REDACTED]","a":1}` {
			t.Errorf("Blob = %q", got)
		}
	})
	t.Run("ohne Pfade", func(t *testing.T) {
		u := neu(t, ohnePfade())
		e := u.lauf(quelle(pw), anfrage())
		if got := vollstaendig(t, u, e, model.RedactionNone); got != pw {
			t.Errorf("Blob = %q", got)
		}
	})
	t.Run("XML mit Metadata", func(t *testing.T) {
		u := neu(t)
		e := u.lauf(quelle(`<edmx/>`), xmlMetadata())
		if got := vollstaendig(t, u, e, model.RedactionNone); got != `<edmx/>` {
			t.Errorf("Blob = %q", got)
		}
	})
	t.Run("XML ohne Metadata", func(t *testing.T) {
		ein := xmlMetadata()
		ein.Metadata = false
		verworfen(t, neu(t).lauf(quelle(`<edmx/>`), ein), capture.Response, capture.Verstoss, capture.GrundTyp)
	})
	t.Run("ContentType fehlt", func(t *testing.T) {
		verworfen(t, neu(t).lauf(quelle(`{}`), capture.Eingang{Richtung: capture.Response}), capture.Response, capture.Verstoss, capture.GrundTyp)
	})
	t.Run("MaxBytes entpackt", func(t *testing.T) {
		for _, f := range []struct {
			laenge int
			grund  capture.Grund
		}{{200, ""}, {201, capture.GrundZuGross}} {
			u := neu(t, mitMax(200), mitVerh(1000000))
			e := u.lauf(quelle(string(gz(jsonString(f.laenge)))), mitEncoding(antwort(), "gzip"))
			if f.grund == "" {
				vollstaendig(t, u, e, model.RedactionNone)
			} else {
				verworfen(t, e, capture.Response, capture.Verstoss, f.grund)
			}
		}
	})
}

func TestWriterErstBeimErstenByte(t *testing.T) {
	ohneAusgabe := []struct {
		name  string
		ein   capture.Eingang
		roh   []byte
		grund capture.Grund
	}{
		{"leer, text/plain", textPlain(), nil, ""},
		{"leer, JSON", antwort(), nil, ""},
		{"leeres gzip-Glied", mitEncoding(antwort(), "gzip"), gz(""), ""},
		{"typ", textPlain(), []byte("hi"), capture.GrundTyp},
	}
	for _, f := range ohneAusgabe {
		t.Run(f.name, func(t *testing.T) {
			u := neu(t)
			e := u.lauf(quelleBytes(f.roh), f.ein)
			if f.grund == "" {
				ohneBody(t, e)
			} else {
				verworfen(t, e, capture.Response, capture.Verstoss, f.grund)
			}
			if st := u.blobs.stand(); st.begins != 0 {
				t.Errorf("Begin %d-mal gerufen", st.begins)
			}
		})
	}
	t.Run("Begin mit MaxBytes und dem Kontext der Aufnahme", func(t *testing.T) {
		u := neu(t, mitMax(1234))
		u.blobs.haltCommit = true
		a := u.tee(quelle(`{}`), antwort())
		leseInGoroutine(a)
		warteAufAnkunft(t, u.blobs.imCommit)
		if st := u.blobs.stand(); st.begins != 1 || st.commits != 1 {
			t.Errorf("Begin %d, Commit %d; erwartet 1 und 1", st.begins, st.commits)
		}
		if len(u.blobs.grenzen) != 1 || u.blobs.grenzen[0] != 1234 {
			t.Errorf("grenze = %v, erwartet [1234]", u.blobs.grenzen)
		}
		ctx := u.blobs.ctxs[0]
		if ctx.Err() != nil {
			t.Errorf("der Kontext von Begin ist vor Abbrechen beendet")
		}
		a.Abbrechen()
		if ctx.Err() == nil {
			t.Errorf("Abbrechen beendet den Kontext von Begin nicht")
		}
		close(u.blobs.freiCommit)
		vollstaendig(t, u, abwarten(t, a), model.RedactionNone)
	})
}

func TestGrundTabelle(t *testing.T) {
	for _, f := range captureFaelle() {
		if f.ausgang == capture.Vollstaendig {
			continue
		}
		t.Run(f.name, func(t *testing.T) {
			_, e := fuehreAus(t, f)
			r := f.ein.Richtung
			verworfen(t, e, r, f.ausgang, f.grund)
			if f.ursache != nil && !errors.Is(e.Fehler, f.ursache) {
				t.Errorf("errors.Is(%v, %v) ist falsch", e.Fehler, f.ursache)
			}
			if f.ausgang == capture.Verstoss && f.grund != capture.GrundAusgabeZuGross && f.ursache == model.ErrUnsicher && !errors.Is(e.Fehler, model.ErrUnsicher) {
				t.Errorf("model.ErrUnsicher fehlt am Fehler")
			}
		})
	}
}

func TestAbortNieCommit(t *testing.T) {
	for _, f := range captureFaelle() {
		if f.ausgang == capture.Vollstaendig {
			continue
		}
		t.Run(f.name, func(t *testing.T) {
			u, _ := fuehreAus(t, f)
			st := u.blobs.stand()
			if st.commits != f.commits {
				t.Errorf("Commit %d-mal gerufen, erwartet %d", st.commits, f.commits)
			}
			if f.writer != beliebig && st.writer != f.writer {
				t.Errorf("Writer %d, erwartet %d", st.writer, f.writer)
			}
			if st.aborts != st.writer {
				t.Errorf("Abort %d-mal bei %d Writern", st.aborts, st.writer)
			}
			if got := u.blobs.inhalte(); len(got) != 0 {
				t.Errorf("Blobs committet: %q", got)
			}
		})
	}
}

func TestCommitLaenge(t *testing.T) {
	for _, delta := range []int64{1, -1} {
		t.Run(fmt.Sprintf("Länge %+d", delta), func(t *testing.T) {
			u := neu(t)
			u.blobs.laengeDelta = delta
			e := u.lauf(quelle(gutBody), antwort())
			verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundAblage)
			if st := u.blobs.stand(); st.commits != 1 || st.aborts != 0 {
				t.Errorf("Writer: %+v; erwartet Commit 1 und kein Abort nach gelungenem Commit", st)
			}
		})
	}
	u := neu(t)
	vollstaendig(t, u, u.lauf(quelle(gutBody), antwort()), model.RedactionNone)
}

func TestAusgabeZuGross(t *testing.T) {
	t.Run("Ausgabe 25 bei MaxBytes 15", func(t *testing.T) {
		u := neu(t, mitMax(15))
		e := u.lauf(quelle(`{"password":""}`), antwort())
		verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundAusgabeZuGross)
		if !errors.Is(e.Fehler, outbound.ErrZuGross) || !errors.Is(e.Fehler, model.ErrUnsicher) {
			t.Errorf("Ursache %v erreicht ErrZuGross oder ErrUnsicher nicht", e.Fehler)
		}
	})
	t.Run("Ausgabe gleich der Eingabe", func(t *testing.T) {
		u := neu(t, mitMax(15))
		vollstaendig(t, u, u.lauf(quelle(jsonString(15)), antwort()), model.RedactionNone)
	})
	t.Run("Ausgabe ein Byte über MaxBytes", func(t *testing.T) {
		u := neu(t, mitMax(24))
		e := u.lauf(quelle(`{"password":"123456789"}`), antwort())
		verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundAusgabeZuGross)
		if !errors.Is(e.Fehler, outbound.ErrZuGross) {
			t.Errorf("Ursache %v erreicht ErrZuGross nicht", e.Fehler)
		}
	})
	t.Run("Ausgabe genau MaxBytes", func(t *testing.T) {
		u := neu(t, mitMax(25))
		e := u.lauf(quelle(`{"password":"123456789"}`), antwort())
		if got := vollstaendig(t, u, e, model.RedactionApplied); got != `{"password":"[REDACTED]"}` {
			t.Errorf("Blob = %q", got)
		}
	})
}

func TestWriterFehler(t *testing.T) {
	grosser := `{"a":"` + wiederholt("x", 200<<10) + `"}`
	faelle := []struct {
		name   string
		blobs  func(*fakeBlobs)
		abort  int
		commit int
	}{
		{"Begin", func(b *fakeBlobs) { b.beginFehler = errDisk }, 0, 0},
		{"erster Write", func(b *fakeBlobs) { b.writeFehler[1] = errDisk }, 1, 0},
		{"zweiter Write", func(b *fakeBlobs) { b.writeFehler[2] = errDisk }, 1, 0},
		{"dritter Write", func(b *fakeBlobs) { b.writeFehler[3] = errDisk }, 1, 0},
		{"Commit", func(b *fakeBlobs) { b.commitFehler = errDisk }, 1, 1},
		{"Write und Abort", func(b *fakeBlobs) { b.writeFehler[2] = errDisk; b.abortFehler = errDisk2 }, 1, 0},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			u := neu(t, mitMax(1<<20))
			f.blobs(u.blobs)
			e := u.lauf(quelle(grosser), antwort())
			verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundAblage)
			if !errors.Is(e.Fehler, errDisk) || errors.Is(e.Fehler, errDisk2) {
				t.Errorf("Ursache %v: errDisk muss erreichbar, errDisk2 darf es nicht sein", e.Fehler)
			}
			if st := u.blobs.stand(); st.aborts != f.abort || st.commits != f.commit {
				t.Errorf("Writer: %+v; erwartet Abort %d, Commit %d", st, f.abort, f.commit)
			}
			z, err := capture.Abschluss(requestOK(t), e)
			if err != nil || z.Capture != model.CaptureFailed {
				t.Errorf("Abschluss = %+v, %v; erwartet failed", z, err)
			}
		})
	}
	t.Run("Kontextfehler des Writers ist abbruch", func(t *testing.T) {
		u := neu(t)
		u.blobs.warteAufCtx = true
		a := u.tee(quelle(gutBody), antwort())
		leseInGoroutine(a)
		warteAufAnkunft(t, u.blobs.imWrite)
		a.Abbrechen()
		e := abwarten(t, a)
		verworfen(t, e, capture.Response, capture.Abbruch, capture.GrundAbbruch)
		if !errors.Is(e.Fehler, context.Canceled) {
			t.Errorf("Ursache %v ist nicht context.Canceled", e.Fehler)
		}
	})
}

func TestKeinGeheimnis(t *testing.T) {
	pruefe := func(t *testing.T, u *umg, e capture.Ergebnis) {
		t.Helper()
		texte := []string{u.blobs.alleAngeboten(), fmt.Sprintf("%+v", e), strings.Join(u.blobs.inhalte(), "|")}
		if e.Fehler != nil {
			texte = append(texte, e.Fehler.Error())
		}
		for _, text := range texte {
			for _, probe := range []string{geheim, "S3C", "3T-7f"} {
				if strings.Contains(text, probe) {
					t.Errorf("%q kommt vor in %q", probe, text)
				}
			}
		}
	}
	t.Run("jede Zeile der Tabelle", func(t *testing.T) {
		for _, f := range captureFaelle() {
			t.Run(f.name, func(t *testing.T) {
				u, e := fuehreAus(t, f)
				pruefe(t, u, e)
				if f.redaction == model.RedactionApplied && !strings.Contains(u.blobs.alleAngeboten(), "[REDACTED]") {
					t.Errorf("der Writer sah den Marker nicht")
				}
			})
		}
	})
	body := `{"password":"` + geheim + `","a":1}`
	t.Run("Fehler von src nach jedem Byte", func(t *testing.T) {
		for k := 0; k <= len(body); k++ {
			u := neu(t)
			e := u.lauf(quelleMit(errQuelle, body[:k]), antwort())
			if k == len(body) {
				verworfen(t, e, capture.Response, capture.Abbruch, capture.GrundAbbruch)
			}
			pruefe(t, u, e)
		}
	})
	t.Run("Fehler des Writers am k-ten Write", func(t *testing.T) {
		grosser := `{"password":"` + geheim + `","a":"` + wiederholt("x", 100<<10) + `"}`
		for k := 1; k <= 3; k++ {
			u := neu(t)
			u.blobs.writeFehler[k] = errDisk
			e := u.lauf(quelle(grosser), antwort())
			verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundAblage)
			pruefe(t, u, e)
		}
	})
	t.Run("gzip, zlib und Kette", func(t *testing.T) {
		faelle := []struct {
			enc string
			roh []byte
		}{
			{"gzip", gz(body)},
			{"deflate", zl([]byte(body))},
			{"deflate, gzip", gz(string(zl([]byte(body))))},
		}
		for _, f := range faelle {
			u := neu(t)
			e := u.lauf(quelleBytes(f.roh), mitEncoding(antwort(), f.enc))
			if got := vollstaendig(t, u, e, model.RedactionApplied); got != `{"password":"[REDACTED]","a":1}` {
				t.Errorf("Blob = %q", got)
			}
			pruefe(t, u, e)
		}
	})
}

func TestKeineDateiUndKeinLog(t *testing.T) {
	verboten := []string{"os", "io/ioutil", "path/filepath", "syscall", "log", "log/slog"}
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	geprueft := 0
	for _, e := range eintraege {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		geprueft++
		f, err := parser.ParseFile(token.NewFileSet(), name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			if p := strings.Trim(imp.Path.Value, `"`); slices.Contains(verboten, p) {
				t.Errorf("%s importiert %s", name, p)
			}
		}
	}
	if geprueft < 6 {
		t.Errorf("nur %d Dateien geprüft", geprueft)
	}
}

func TestRedigiertUndKey(t *testing.T) {
	regeln := model.SchluesselRegeln{InvalidEncoding: model.KodierungLiteral, BodyMode: model.BodyBytes}
	schluessel := func(t *testing.T, body []byte) string {
		t.Helper()
		k, err := model.BildeSchluessel(model.SchluesselEingabe{
			Method: "POST", ServiceID: "svc", Path: "/x", Body: body, RequestRedaction: model.RedactionApplied,
		}, regeln)
		if err != nil {
			t.Fatal(err)
		}
		return k.MatchKey
	}
	var keys, rohKeys []string
	for _, geheim := range []string{"S3CR3T-A", "S3CR3T-B"} {
		u := neu(t)
		roh := `{"password":"` + geheim + `","a":1}`
		e := u.lauf(quelle(roh), anfrage())
		blob := vollstaendig(t, u, e, model.RedactionApplied)
		if string(e.Redigiert) != `{"password":"[REDACTED]","a":1}` || string(e.Redigiert) != blob {
			t.Errorf("Redigiert = %q, Blob = %q", e.Redigiert, blob)
		}
		keys = append(keys, schluessel(t, e.Redigiert))
		rohKeys = append(rohKeys, schluessel(t, []byte(roh)))
	}
	if keys[0] != keys[1] || rohKeys[0] == rohKeys[1] {
		t.Errorf("Keys aus Redigiert %v, aus den Rohbytes %v: erwartet gleich und verschieden", keys, rohKeys)
	}
	t.Run("nil in allen anderen Fällen", func(t *testing.T) {
		u := neu(t)
		if e := u.lauf(quelle(`{"a":1}`), antwort()); e.Redigiert != nil {
			t.Errorf("Redigiert bei Response: %q", e.Redigiert)
		}
		if e := neu(t).lauf(quelle(), anfrage()); e.Redigiert != nil || e.Body != model.BodyAbsent {
			t.Errorf("Redigiert bei absent: %+v", e)
		}
		if e := neu(t).lauf(quelle(`{"a":x`), anfrage()); e.Redigiert != nil || e.Body != model.BodyDiscarded {
			t.Errorf("Redigiert bei discarded: %+v", e)
		}
	})
	t.Run("Ausgabe bis genau MaxBytes", func(t *testing.T) {
		genau := `{"password":""` + wiederholt(" ", 5) + `}`
		u := neu(t, mitMax(30))
		e := u.lauf(quelle(genau), anfrage())
		vollstaendig(t, u, e, model.RedactionApplied)
		if len(e.Redigiert) != 30 {
			t.Errorf("len(Redigiert) = %d, erwartet 30", len(e.Redigiert))
		}
		ueber := `{"password":""` + wiederholt(" ", 6) + `}`
		e = neu(t, mitMax(30)).lauf(quelle(ueber), anfrage())
		verworfen(t, e, capture.Request, capture.Verstoss, capture.GrundAusgabeZuGross)
	})
}

func TestZerlegungOhneWirkung(t *testing.T) {
	for _, f := range captureFaelle() {
		if f.treibe != nil || len(f.roh) > 200 {
			continue
		}
		t.Run(f.name, func(t *testing.T) {
			_, basis := fuehreAus(t, f)
			uBasis := basis
			zerlegungen := map[string][][]byte{"einzelne Bytes": zerlege(f.roh, 1)}
			for i := 1; i < len(f.roh); i++ {
				zerlegungen[fmt.Sprintf("Zweiteilung bei %d", i)] = [][]byte{f.roh[:i], f.roh[i:]}
			}
			for name, teile := range zerlegungen {
				for _, leer := range []bool{false, true} {
					u := neu(t, f.opt...)
					if f.blobs != nil {
						f.blobs(u.blobs)
					}
					s := &skript{ende: f.ende}
					for _, p := range teile {
						if leer {
							s.schritte = append(s.schritte, schritt{})
						}
						s.schritte = append(s.schritte, schritt{daten: string(p)})
					}
					e := u.lauf(s, f.ein)
					if e.Ausgang != uBasis.Ausgang || e.Grund != uBasis.Grund || e.Body != uBasis.Body ||
						e.Redaction != uBasis.Redaction || e.Integrity != uBasis.Integrity || !gleicherBlob(u, e, uBasis, f) {
						t.Errorf("%s (leere Reads: %v): %+v, erwartet wie ein Read: %+v", name, leer, e, uBasis)
					}
				}
			}
		})
	}
}

// gleicherBlob vergleicht den Inhalt des Blobs, nicht die Referenz zweier Umgebungen.
func gleicherBlob(u *umg, e, basis capture.Ergebnis, f fall) bool {
	if e.Body != model.BodyPresent {
		return e.Blob == "" && basis.Blob == ""
	}
	inhalt, ok := u.blobs.inhalt(e.Blob)
	return ok && inhalt == f.blob
}

func zerlege(b []byte, n int) [][]byte {
	var out [][]byte
	for len(b) > 0 {
		k := min(n, len(b))
		out = append(out, b[:k])
		b = b[k:]
	}
	return out
}

var _ = io.EOF
