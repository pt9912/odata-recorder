package capture_test

import (
	"errors"
	"io"
	"math"
	"strconv"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
)

const gutBody = `{"a":1}`

func TestKette(t *testing.T) {
	gutFaelle := []struct {
		name string
		enc  []string
		roh  []byte
	}{
		{"kein Header", nil, []byte(gutBody)},
		{"gzip", []string{"gzip"}, gz(gutBody)},
		{"GZIP", []string{"GZIP"}, gz(gutBody)},
		{"Randleerraum", []string{" \tgzip \t"}, gz(gutBody)},
		{"deflate", []string{"deflate"}, zl([]byte(gutBody))},
		{"identity", []string{"identity"}, []byte(gutBody)},
		{"identity, gzip", []string{"identity, gzip"}, gz(gutBody)},
		{"gzip, identity", []string{"gzip, identity"}, gz(gutBody)},
		{"deflate, gzip in einer Zeile", []string{"deflate, gzip"}, gz(string(zl([]byte(gutBody))))},
		{"deflate, gzip in zwei Zeilen", []string{"deflate", "gzip"}, gz(string(zl([]byte(gutBody))))},
	}
	for _, f := range gutFaelle {
		t.Run(f.name, func(t *testing.T) {
			u := neu(t)
			e := u.lauf(quelleBytes(f.roh), mitEncoding(antwort(), f.enc...))
			if got := vollstaendig(t, u, e, model.RedactionNone); got != gutBody {
				t.Errorf("Blob = %q", got)
			}
		})
	}
	schlecht := []struct {
		name string
		enc  []string
	}{
		{"br", []string{"br"}},
		{"zstd", []string{"zstd"}},
		{"compress", []string{"compress"}},
		{"x-gzip", []string{"x-gzip"}},
		{"gzip;q=1", []string{"gzip;q=1"}},
		{"leerer Wert", []string{""}},
		{"nur Leerraum", []string{" \t "}},
		{"gzip,", []string{"gzip,"}},
		{",gzip", []string{",gzip"}},
		{"gzip und leere Zeile", []string{"gzip", ""}},
		{"identity,identity,gzip", []string{"identity,identity,gzip"}},
		{"gzip,gzip,gzip", []string{"gzip,gzip,gzip"}},
		{"drei Glieder in drei Zeilen", []string{"gzip", "gzip", "gzip"}},
		{"gzip, br", []string{"gzip, br"}},
	}
	for _, f := range schlecht {
		t.Run(f.name, func(t *testing.T) {
			u := neu(t)
			src := quelle(gutBody)
			a := u.tee(src, mitEncoding(antwort(), f.enc...))
			got, err := io.ReadAll(a)
			if err != nil || string(got) != gutBody {
				t.Errorf("Read reicht nicht durch: %q, %v", got, err)
			}
			verworfen(t, abwarten(t, a), capture.Response, capture.Verstoss, capture.GrundEncoding)
			if st := u.blobs.stand(); st.begins != 0 || st.aborts != 0 {
				t.Errorf("Writer berührt: %+v", st)
			}
		})
	}
	t.Run("br mit leerem Body", func(t *testing.T) {
		u := neu(t)
		e := u.lauf(quelle(), mitEncoding(antwort(), "br"))
		verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundEncoding)
	})
}

func TestKetteReihenfolge(t *testing.T) {
	roh := gz(string(zl([]byte(gutBody))))
	u := neu(t)
	e := u.lauf(quelleBytes(roh), mitEncoding(antwort(), "deflate, gzip"))
	if got := vollstaendig(t, u, e, model.RedactionNone); got != gutBody {
		t.Errorf("Blob = %q", got)
	}
	e = neu(t).lauf(quelleBytes(roh), mitEncoding(antwort(), "gzip, deflate"))
	verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundDekodierung)
}

func TestEntpackt(t *testing.T) {
	gutFaelle := []struct {
		name string
		enc  string
		roh  []byte
	}{
		{"gzip", "gzip", gz(gutBody)},
		{"zlib", "deflate", zl([]byte(gutBody))},
		{"zwei gzip-Glieder", "gzip", append(gz(`{"a":`), gz(`1}`)...)},
		{"identity", "identity", []byte(gutBody)},
	}
	for _, f := range gutFaelle {
		t.Run(f.name, func(t *testing.T) {
			u := neu(t)
			e := u.lauf(quelleBytes(f.roh), mitEncoding(antwort(), f.enc))
			if got := vollstaendig(t, u, e, model.RedactionNone); got != gutBody {
				t.Errorf("Blob = %q", got)
			}
		})
	}
	t.Run("raw-deflate", func(t *testing.T) {
		e := neu(t).lauf(quelleBytes(rohDeflate(gutBody)), mitEncoding(antwort(), "deflate"))
		verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundDekodierung)
	})
	leer := []struct {
		name string
		enc  string
		roh  []byte
	}{
		{"leeres gzip-Glied", "gzip", gz("")},
		{"leere Rohbytes mit gzip", "gzip", nil},
		{"leere Rohbytes mit deflate", "deflate", nil},
		{"leere Rohbytes mit identity", "identity", nil},
	}
	for _, f := range leer {
		t.Run(f.name, func(t *testing.T) {
			u := neu(t)
			ohneBody(t, u.lauf(quelleBytes(f.roh), mitEncoding(antwort(), f.enc)))
			if st := u.blobs.stand(); st.begins != 0 {
				t.Errorf("Begin %d-mal gerufen", st.begins)
			}
		})
	}
}

// aendere liefert eine Kopie von b mit dem Byte an Stelle i verändert.
func aendere(b []byte, i int) []byte {
	c := append([]byte(nil), b...)
	if i < 0 {
		i += len(c)
	}
	c[i] ^= 0xFF
	return c
}

func TestDekodierfehler(t *testing.T) {
	koerper := `{"a":"` + wiederholt("q", 50) + `"}`
	g, z := gz(koerper), zl([]byte(koerper))
	faelle := []struct {
		name string
		enc  string
		roh  []byte
	}{
		{"gzip mit verändertem CRC", "gzip", aendere(g, -8)},
		{"gzip mit verändertem Längenfeld", "gzip", aendere(g, -1)},
		{"gzip mit Magic 1f 8c", "gzip", func() []byte { b := append([]byte(nil), g...); b[1] = 0x8c; return b }()},
		{"gzip nach Byte 5 abgeschnitten", "gzip", g[:5]},
		{"gzip nach Byte 15 abgeschnitten", "gzip", g[:15]},
		{"gzip um ein Byte gekürzt", "gzip", g[:len(g)-1]},
		{"zlib mit falscher Adler-Summe", "deflate", aendere(z, -1)},
		{"zlib abgeschnitten", "deflate", z[:len(z)-1]},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			u := neu(t)
			e := u.lauf(quelleBytes(f.roh), mitEncoding(antwort(), f.enc))
			verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundDekodierung)
			if len(u.blobs.inhalte()) != 0 {
				t.Errorf("ein Blob entstand")
			}
		})
	}
}

func TestAbbruchImGzip(t *testing.T) {
	koerper := `{"a":"` + wiederholt("q", 50) + `"}`
	g := gz(koerper)
	for _, k := range []int{5, 15, len(g) - 1} {
		teil := string(g[:k])
		t.Run("Fehler von src nach "+strconv.Itoa(k), func(t *testing.T) {
			u := neu(t)
			e := u.lauf(quelleMit(errQuelle, teil), mitEncoding(antwort(), "gzip"))
			verworfen(t, e, capture.Response, capture.Abbruch, capture.GrundAbbruch)
			if !errors.Is(e.Fehler, errQuelle) {
				t.Errorf("Ursache erreicht errQuelle nicht: %v", e.Fehler)
			}
		})
		t.Run("io.EOF nach "+strconv.Itoa(k), func(t *testing.T) {
			u := neu(t)
			e := u.lauf(quelle(teil), mitEncoding(antwort(), "gzip"))
			verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundDekodierung)
		})
	}
}

func TestNachlauf(t *testing.T) {
	faelle := []struct {
		name  string
		enc   string
		roh   []byte
		grund capture.Grund
	}{
		{"zlib-Strom und x", "deflate", append(zl([]byte(gutBody)), 'x'), capture.GrundDekodierung},
		{"gzip-Strom und 0x00", "gzip", append(gz(gutBody), 0x00), capture.GrundDekodierung},
		{"zlib-Strom ohne Nachlauf", "deflate", zl([]byte(gutBody)), ""},
		{"gzip-Strom ohne Nachlauf", "gzip", gz(gutBody), ""},
		{"x hinter dem zlib-Strom in gzip", "deflate, gzip", gz(string(append(zl([]byte(gutBody)), 'x'))), capture.GrundDekodierung},
		{"x hinter dem Strom beider Glieder", "deflate, gzip", append(gz(string(zl([]byte(gutBody)))), 'x'), capture.GrundDekodierung},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			u := neu(t)
			e := u.lauf(quelleBytes(f.roh), mitEncoding(antwort(), f.enc))
			if f.grund == "" {
				vollstaendig(t, u, e, model.RedactionNone)
				return
			}
			verworfen(t, e, capture.Response, capture.Verstoss, f.grund)
		})
	}
}

func TestVerhaeltnisFunktion(t *testing.T) {
	faelle := []struct {
		name                  string
		roh, entpackt, grenze int64
		ok                    bool
	}{
		{"genau max × roh", 10, 200, 20, true},
		{"ein Byte mehr", 10, 201, 20, false},
		{"ein Byte weniger", 10, 199, 20, true},
		{"roh 1 genau", 1, 20, 20, true},
		{"roh 1 ein Byte mehr", 1, 21, 20, false},
		{"max 1 genau", 1, 1, 1, true},
		{"max 1 ein Byte mehr", 1, 2, 1, false},
		{"max × roh läuft über", math.MaxInt64 / 2, math.MaxInt64, 20, true},
		{"MaxInt64 bei roh 1", 1, math.MaxInt64, 20, false},
		{"nichts entpackt", 5, 0, 20, true},
		{"nichts roh, nichts entpackt", 0, 0, 20, true},
		{"nichts roh, etwas entpackt", 0, 1, 20, false},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			if got := capture.Verhaeltnis(f.roh, f.entpackt, f.grenze); got != f.ok {
				t.Errorf("Verhaeltnis(%d, %d, %d) = %v, erwartet %v", f.roh, f.entpackt, f.grenze, got, f.ok)
			}
		})
	}
}

func TestBombe(t *testing.T) {
	koerper := jsonString(1 << 20)
	gzBombe, zlBombe := gz(koerper), zl([]byte(koerper))
	faelle := []struct {
		name string
		enc  string
		roh  []byte
	}{
		{"gzip", "gzip", gzBombe},
		{"zlib", "deflate", zlBombe},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			u := neu(t, mitMax(8<<20), mitVerh(20))
			e := u.lauf(quelleBytes(f.roh), mitEncoding(antwort(), f.enc))
			verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundBombe)
			u = neu(t, mitMax(8<<20), mitVerh(100000))
			e = u.lauf(quelleBytes(f.roh), mitEncoding(antwort(), f.enc))
			if got := vollstaendig(t, u, e, model.RedactionNone); got != koerper {
				t.Errorf("Blob weicht ab (Länge %d)", len(got))
			}
		})
	}
	t.Run("Nachlauf und Verhältnis zugleich", func(t *testing.T) {
		e := neu(t, mitMax(8<<20), mitVerh(20)).lauf(quelleBytes(append(zlBombe, 'x')), mitEncoding(antwort(), "deflate"))
		verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundDekodierung)
	})
	t.Run("ohne Kompression ist das Verhältnis 1", func(t *testing.T) {
		u := neu(t, mitVerh(1))
		vollstaendig(t, u, u.lauf(quelle(gutBody), antwort()), model.RedactionNone)
	})
}

func TestEntpacktGrenze(t *testing.T) {
	const grenze = 1000
	faelle := []struct {
		name   string
		laenge int
		grund  capture.Grund
	}{
		{"genau MaxBytes entpackt", grenze, ""},
		{"MaxBytes + 1 entpackt", grenze + 1, capture.GrundZuGross},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			u := neu(t, mitMax(grenze), mitVerh(1000000))
			e := u.lauf(quelle(string(gz(jsonString(f.laenge)))), mitEncoding(antwort(), "gzip"))
			if f.grund == "" {
				vollstaendig(t, u, e, model.RedactionNone)
				return
			}
			verworfen(t, e, capture.Response, capture.Verstoss, f.grund)
		})
	}
	t.Run("gzip-Bombe über MaxBytes", func(t *testing.T) {
		const grenze = 1 << 20
		u := neu(t, mitMax(grenze), mitVerh(20))
		a := u.tee(quelleBytes(gz(jsonString(32<<20))), mitEncoding(antwort(), "gzip"))
		_, _ = io.Copy(io.Discard, a)
		verworfen(t, abwarten(t, a), capture.Response, capture.Verstoss, capture.GrundZuGross)
		if n := a.Entpackt(); n > grenze+1 {
			t.Errorf("der Redaktor las %d Bytes, erlaubt sind höchstens MaxBytes + 1 = %d", n, grenze+1)
		}
	})
}

// TestLeereRohbytesJeKette hält fest, dass ein Body ohne Rohbyte bei jeder bekannten Kette leer
// ist und nur eine unbekannte Kette oder ein Byte ohne gültigen Kopf ein Verstoß ist (SPEC-119 Nr. 6, 8).
func TestLeereRohbytesJeKette(t *testing.T) {
	for _, enc := range []string{"gzip", "deflate", "gzip,deflate", "identity"} {
		t.Run(enc, func(t *testing.T) {
			u := neu(t)
			e := u.lauf(quelle(), mitEncoding(antwort(), enc))
			ohneBody(t, e)
			if st := u.blobs.stand(); st.begins != 0 {
				t.Errorf("Begin %d-mal gerufen, erwartet 0", st.begins)
			}
		})
	}
	t.Run("br", func(t *testing.T) {
		e := neu(t).lauf(quelle(), mitEncoding(antwort(), "br"))
		verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundEncoding)
	})
	t.Run("ein Byte 0x00 bei gzip", func(t *testing.T) {
		e := neu(t).lauf(quelle("\x00"), mitEncoding(antwort(), "gzip"))
		verworfen(t, e, capture.Response, capture.Verstoss, capture.GrundDekodierung)
	})
}
