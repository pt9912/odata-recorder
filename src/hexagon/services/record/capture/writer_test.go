package capture_test

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

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
	t.Run("mehrere Werte von Content-Type", func(t *testing.T) {
		for _, werte := range [][]string{{"application/json", "text/plain"}, {"text/plain", "application/json"}} {
			ein := capture.Eingang{Richtung: capture.Response, ContentType: werte}
			verworfen(t, neu(t).lauf(quelle(`{}`), ein), capture.Response, capture.Verstoss, capture.GrundTyp)
		}
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
	// ErrZuGross aus Begin oder Commit ist wie aus Write ausgabe_zu_gross, auch bei beendetem
	// Kontext (SPEC-119 Nr. 9, 10); ein Writer, den Begin nicht lieferte, bekommt kein Abort.
	for _, f := range []struct {
		aufruf  string
		beendet bool
	}{{"Begin", false}, {"Begin", true}, {"Commit", false}, {"Commit", true}} {
		t.Run(fmt.Sprintf("aus %s, Kontext beendet: %v", f.aufruf, f.beendet), func(t *testing.T) {
			u := neu(t)
			var im <-chan struct{}
			var frei chan struct{}
			if f.aufruf == "Begin" {
				u.blobs.beginFehler, u.blobs.haltBegin = outbound.ErrZuGross, f.beendet
				im, frei = u.blobs.imBegin, u.blobs.freiBegin
			} else {
				u.blobs.commitFehler, u.blobs.haltCommit = outbound.ErrZuGross, f.beendet
				im, frei = u.blobs.imCommit, u.blobs.freiCommit
			}
			a := u.tee(quelle(gutBody), antwort())
			leseInGoroutine(a)
			if f.beendet {
				warteAufAnkunft(t, im)
				a.Abbrechen()
				close(frei)
			}
			e := abwarten(t, a)
			verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundAusgabeZuGross)
			if !errors.Is(e.Fehler, outbound.ErrZuGross) {
				t.Errorf("Ursache %v erreicht outbound.ErrZuGross nicht", e.Fehler)
			}
			want := stand{begins: 1, writer: 1, commits: 1, aborts: 1}
			if f.aufruf == "Begin" {
				want = stand{begins: 1}
			}
			if st := u.blobs.stand(); st != want {
				t.Errorf("Writer: %+v; erwartet %+v", st, want)
			}
		})
	}
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
	// Ein Kontextfehler des Writers bei lebendem Kontext der Aufnahme ist ein Fehler des Writers
	// wie jeder andere (SPEC-119 Nr. 9, 10).
	for _, cerr := range []error{context.Canceled, context.DeadlineExceeded} {
		for _, aufruf := range []string{"Begin", "Write", "Commit"} {
			t.Run(fmt.Sprintf("%v aus %s bei lebendem Kontext", cerr, aufruf), func(t *testing.T) {
				u := neu(t)
				switch aufruf {
				case "Begin":
					u.blobs.beginFehler = cerr
				case "Write":
					u.blobs.writeFehler[1] = cerr
				default:
					u.blobs.commitFehler = cerr
				}
				e := u.lauf(quelle(gutBody), antwort())
				verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundAblage)
				if !errors.Is(e.Fehler, cerr) {
					t.Errorf("Ursache %v erreicht %v nicht", e.Fehler, cerr)
				}
			})
		}
	}
	// Ist der Kontext beendet, zählt der Kontextfehler des Writers als die Ursache, die ihn beendete.
	for _, aufruf := range []string{"Begin", "Write", "Commit"} {
		for _, frist := range []bool{false, true} {
			t.Run(fmt.Sprintf("Kontextfehler aus %s bei beendetem Kontext, Frist: %v", aufruf, frist), func(t *testing.T) {
				var opt []func(*capture.Konfig)
				if frist {
					opt = append(opt, mitFrist(30*time.Millisecond))
				}
				u := neu(t, opt...)
				im := u.blobs.imWrite
				switch aufruf {
				case "Begin":
					u.blobs.warteBegin, im = true, u.blobs.imBegin
				case "Write":
					u.blobs.warteAufCtx = true
				default:
					u.blobs.warteCommit, im = true, u.blobs.imCommit
				}
				a := u.tee(quelle(gutBody), antwort())
				leseInGoroutine(a)
				if !frist {
					warteAufAnkunft(t, im)
					a.Abbrechen()
				}
				e := abwarten(t, a)
				if frist {
					verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundFrist)
					if !errors.Is(e.Fehler, context.DeadlineExceeded) {
						t.Errorf("Ursache %v ist nicht context.DeadlineExceeded", e.Fehler)
					}
				} else {
					verworfen(t, e, capture.Response, capture.Abbruch, capture.GrundAbbruch)
					if !errors.Is(e.Fehler, context.Canceled) {
						t.Errorf("Ursache %v ist nicht context.Canceled", e.Fehler)
					}
				}
				if st := u.blobs.stand(); st.commits != 0 && aufruf != "Commit" || st.aborts != st.writer {
					t.Errorf("Writer: %+v", st)
				}
			})
		}
	}
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

// erlaubteImporte liefert die Importe der Nicht-Test-Dateien des Pakets; ein weiterer ist ein Befund
// von TestKeineDateiUndKeinLog und braucht hier eine Begründung.
func erlaubteImporte() []string {
	return []string{
		"bufio", "bytes", "compress/gzip", "compress/zlib", "context", "errors", "fmt", "io", "slices", "strings", "sync", "time",
		"github.com/pt9912/odata-recorder/src/hexagon/model",
		"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound",
	}
}

// schreibwege nennt, was eine Quelldatei gegen "kein Rohbyte in einer Datei, kein Log" verstößt:
// ein Import außerhalb von erlaubteImporte(), ein Aufruf von print, println oder panic und ein
// Aufruf von fmt.Print* oder fmt.Fprint*. src nil liest die Datei name.
func schreibwege(t *testing.T, name string, src any) []string {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), name, src, 0)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, imp := range f.Imports {
		if p := strings.Trim(imp.Path.Value, `"`); !slices.Contains(erlaubteImporte(), p) {
			out = append(out, "Import "+p)
		}
	}
	ast.Inspect(f, func(n ast.Node) bool {
		c, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fn := c.Fun.(type) {
		case *ast.Ident:
			if slices.Contains([]string{"print", "println", "panic"}, fn.Name) {
				out = append(out, "Aufruf "+fn.Name)
			}
		case *ast.SelectorExpr:
			if x, ok := fn.X.(*ast.Ident); ok && x.Name == "fmt" &&
				(strings.HasPrefix(fn.Sel.Name, "Print") || strings.HasPrefix(fn.Sel.Name, "Fprint")) {
				out = append(out, "Aufruf fmt."+fn.Sel.Name)
			}
		}
		return true
	})
	return out
}

// TestKeineDateiUndKeinLog hält fest, dass die Nicht-Test-Dateien des Pakets keinen Import
// außerhalb der Liste haben und weder print, println, panic noch fmt.Print* oder fmt.Fprint*
// aufrufen. Grenze: der Test sieht Importe und diese Aufrufe, keine Schreibwege über Ports
// (dafür steht TestKeinGeheimnis) und keine Umwege über einen Funktionswert.
func TestKeineDateiUndKeinLog(t *testing.T) {
	t.Run("Gegenprobe: die Prüfung erkennt jeden Weg", func(t *testing.T) {
		faelle := []struct{ name, src, want string }{
			{"os", "package p\nimport \"os\"\n", "Import os"},
			{"net/http", "package p\nimport \"net/http\"\n", "Import net/http"},
			{"log", "package p\nimport \"log/slog\"\n", "Import log/slog"},
			{"println", "package p\nfunc f(b []byte) { println(string(b)) }\n", "Aufruf println"},
			{"print", "package p\nfunc f(b []byte) { print(string(b)) }\n", "Aufruf print"},
			{"panic", "package p\nfunc f(b []byte) { panic(string(b)) }\n", "Aufruf panic"},
			{"fmt.Println", "package p\nimport \"fmt\"\nfunc f(b []byte) { fmt.Println(b) }\n", "Aufruf fmt.Println"},
			{"fmt.Fprintf", "package p\nimport \"fmt\"\nfunc f(w fmt.Stringer) { fmt.Fprintf(nil, \"%s\", w) }\n", "Aufruf fmt.Fprintf"},
		}
		for _, f := range faelle {
			if got := schreibwege(t, f.name+".go", f.src); !slices.Equal(got, []string{f.want}) {
				t.Errorf("%s: %v, erwartet [%s]", f.name, got, f.want)
			}
		}
		if got := schreibwege(t, "gut.go", "package p\nimport \"fmt\"\nfunc f(b []byte) error { return fmt.Errorf(\"%d\", len(b)) }\n"); len(got) != 0 {
			t.Errorf("fmt.Errorf ist erlaubt: %v", got)
		}
	})
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
		for _, v := range schreibwege(t, name, nil) {
			t.Errorf("%s: %s", name, v)
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
	t.Run("Request über 32 KiB hat mehrere Writes", func(t *testing.T) {
		u := neu(t)
		roh := `{"a":"` + wiederholt("x", 100<<10) + `","password":"` + geheim + `"}`
		want := `{"a":"` + wiederholt("x", 100<<10) + `","password":"[REDACTED]"}`
		e := u.lauf(quelle(roh), anfrage())
		blob := vollstaendig(t, u, e, model.RedactionApplied)
		if n := u.blobs.writes(); n < 3 {
			t.Fatalf("%d Writes; der Fall braucht mehrere", n)
		}
		if blob != want || string(e.Redigiert) != want {
			t.Errorf("Blob und Redigiert sind nicht gleich der ganzen Ausgabe (Längen %d und %d, erwartet %d)", len(blob), len(e.Redigiert), len(want))
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
