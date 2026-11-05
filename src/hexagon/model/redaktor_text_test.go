package model_test

import (
	"bytes"
	"errors"
	"io"
	"net/url"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Tests dieser Datei fangen die Zusagen Z1 bis Z11 und Z9 von SPEC-118 Nr. 3 und Nr. 11
// (Typ Text). Bytes ab 0x80 stehen als \x-Folgen im Quelltext, nie als Zeichen.

const (
	zeichenE  = "\xC3\xA9"
	zeichenEu = "\xE2\x82\xAC"
	zeichenSm = "\xF0\x9F\x98\x80"
)

// zielEin baut den Eingang: art ist "" (kein Ziel), "zaehlung" oder "wert" (Name `Name`, kein Muster).
func zielEin(art string, ct ...string) model.BodyEingang {
	e := model.BodyEingang{ContentType: ct}
	switch art {
	case "zaehlung":
		e.Zaehlung = true
	case "wert":
		e.Wert, e.Name = true, "Name"
	}
	return e
}

// wertEin ist der Eingang eines `$value` unter name.
func wertEin(name string, ct ...string) model.BodyEingang {
	return model.BodyEingang{ContentType: ct, Wert: true, Name: name}
}

func istUnsicher(g ergebnis, grund model.UnsicherGrund) bool {
	return g.status == model.RedactionUnsafe && grundVon(g.err) == grund
}

// Z1: Text ist nur mit Ziel bekannt.
func TestTextNurMitZiel(t *testing.T) {
	typen := []string{"text/plain", "text/csv", "text/html", "TEXT/Plain", " text/plain ", "\ttext/xml\t", "text/json", "text/plain; charset=utf-8"}
	n := 0
	for _, art := range []string{"", "zaehlung", "wert"} {
		for _, ct := range typen {
			n++
			g := redigiere(t, einmal("12"), zielEin(art, ct), bodyRegeln(t))
			if art == "" {
				if !istUnsicher(g, model.GrundTyp) || g.aus != "" {
					t.Errorf("%q ohne Ziel: %+v", ct, g)
				}
				continue
			}
			if g.err != nil || g.status != model.RedactionNone || g.aus != "12" {
				t.Errorf("%q mit Ziel %s: %+v", ct, art, g)
			}
		}
	}
	if n != 24 {
		t.Errorf("%d Fälle, erwartet 24", n)
	}
}

// Z2: Randformen des Typs bei gesetztem Ziel.
func TestTextTypRandformen(t *testing.T) {
	unbekannt := [][]string{
		{"text/"}, {"text"}, {"xtext/plain"}, {"application/text"}, {"text/;x"}, {" text/ "}, {"texts/plain"},
		nil, {}, {""}, {"text/plain", "text/plain"}, {"text/plain", ""}, {"", "text/plain"},
	}
	n := 0
	for _, art := range []string{"zaehlung", "wert"} {
		for _, ct := range unbekannt {
			n++
			g := redigiere(t, einmal("12"), zielEin(art, ct...), bodyRegeln(t))
			if !istUnsicher(g, model.GrundTyp) || g.aus != "" {
				t.Errorf("%q mit Ziel %s: %+v", ct, art, g)
			}
		}
		// Der Untertyp wird nicht auf Bytes geprüft.
		for _, ct := range []string{"text/ x", "text/\x01", "text/a b;c", "text/" + zeichenE} {
			n++
			g := redigiere(t, einmal("1"), zielEin(art, ct), bodyRegeln(t))
			if g.err != nil || g.aus != "1" {
				t.Errorf("%q mit Ziel %s: %+v", ct, art, g)
			}
		}
	}
	if n != 34 {
		t.Errorf("%d Fälle, erwartet 34", n)
	}
}

// Z3: JSON bleibt JSON bei gesetztem Ziel.
func TestTextJSONBleibtJSON(t *testing.T) {
	for _, art := range []string{"zaehlung", "wert"} {
		g := redigiere(t, einmal("3"), zielEin(art, "application/json"), bodyRegeln(t))
		if g.err != nil || g.status != model.RedactionNone || g.aus != "3" {
			t.Errorf("%s: 3 als JSON: %+v", art, g)
		}
		g = redigiere(t, einmal("abc"), zielEin(art, "application/json"), bodyRegeln(t))
		if !istUnsicher(g, model.GrundSyntax) {
			t.Errorf("%s: abc als JSON: %+v", art, g)
		}
		g = redigiere(t, einmal(`{"password":"s3cr3t"}`), zielEin(art, "application/vnd.x+json"), bodyRegeln(t))
		if g.err != nil || g.status != model.RedactionApplied || g.aus != `{"password":"[REDACTED]"}` {
			t.Errorf("%s: JSON-Redaktion: %+v", art, g)
		}
	}
}

// zeichensatzFall nennt einen Wert des Headers und ob er erlaubt ist.
type zeichensatzFall struct {
	ct      string
	erlaubt bool
}

// Z4: Zeichensatz, jede Form, für beide Ziele.
func TestTextZeichensatz(t *testing.T) {
	faelle := []zeichensatzFall{
		{"text/plain", true}, {"text/plain; charset=utf-8", true}, {"text/plain; charset=UTF-8", true},
		{`text/plain; charset="utf-8"`, true}, {`text/plain; charset="UTF-8"`, true}, {"text/plain; charset=us-ascii", true},
		{"text/plain; charset=US-ASCII", true}, {`text/plain; charset="us-ascii"`, true}, {"text/plain;charset=utf-8", true},
		{"text/plain; CHARSET=utf-8", true}, {"text/plain; charset = utf-8", true}, {"text/plain;\tcharset=utf-8\t", true},
		{"text/plain; foo=bar; charset=utf-8", true}, {"text/plain; foo", true}, {"text/plain; charsetx=latin1", true},
		{"text/plain; charset=utf-8; charset=utf-8", true}, {"text/plain; charset=utf-8; charset=US-ASCII", true},
		{"text/plain;", true}, {"text/plain; ;", true},
		{"text/plain; charset=iso-8859-1", false}, {"text/plain; charset=latin1", false}, {"text/plain; charset=utf-16", false},
		{"text/plain; charset=utf8", false}, {"text/plain; charset=", false}, {"text/plain; charset", false},
		{`text/plain; charset=""`, false}, {`text/plain; charset="utf-8`, false}, {`text/plain; charset=utf-8"`, false},
		{`text/plain; charset="`, false}, {"text/plain; charset=utf-8x", false}, {"text/plain; charset=xutf-8", false},
		{`text/plain; charset=" utf-8"`, false}, {`text/plain; charset="utf-8 "`, false}, {`text/plain; charset="utf-8"x`, false},
		{"text/plain; charset=utf-8 x", false}, {"text/plain; charset=utf-8; charset=latin1", false},
		{"text/plain; charset=latin1; charset=utf-8", false}, {"text/plain; charset=utf-8; charset=", false},
		{"text/plain; charset=utf-8; charset", false}, {"text/plain; foo=bar; charset=latin1", false},
		{"text/csv; charset=latin1", false}, {"text/plain; charset=utf-8,us-ascii", false},
		{"text/plain; CHARSET=latin1", false}, {"text/plain; charset = latin1", false}, {"text/plain; \tcharset\t=\tlatin1", false},
		{"text/plain; charset=LATIN1", false},
	}
	n := 0
	for _, art := range []string{"zaehlung", "wert"} {
		for _, f := range faelle {
			n++
			g := redigiere(t, einmal("12"), zielEin(art, f.ct), bodyRegeln(t))
			if f.erlaubt && (g.err != nil || g.aus != "12" || g.status != model.RedactionNone) {
				t.Errorf("%q mit Ziel %s: %+v", f.ct, art, g)
			}
			if !f.erlaubt && (!istUnsicher(g, model.GrundCharset) || g.aus != "") {
				t.Errorf("%q mit Ziel %s: %+v", f.ct, art, g)
			}
		}
	}
	if n != 2*len(faelle) || len(faelle) != 46 {
		t.Errorf("%d Fälle", n)
	}
	// Der Zeichensatz gehört zum Typ Text: bei JSON ohne Wirkung.
	g := redigiere(t, einmal(`{"a":1}`), zielEin("zaehlung", "application/json; charset=latin1"), bodyRegeln(t))
	if g.err != nil {
		t.Errorf("JSON mit charset: %v", g.err)
	}
}

// Z4/Z11: Rangfolge der Verstöße beim Zeichensatz.
func TestTextZeichensatzRangfolge(t *testing.T) {
	for _, art := range []string{"zaehlung", "wert"} {
		g := redigiere(t, einmal("1a"), zielEin(art, "text/plain; charset=iso-8859-1"), bodyRegeln(t))
		if !istUnsicher(g, model.GrundCharset) || g.aus != "" {
			t.Errorf("%s: charset und Nicht-Ziffer: %+v", art, g)
		}
		regeln := model.BodyRegeln{Pfade: standardPfade(t), MaxBytes: 2}
		g = redigiere(t, einmal("12345"), zielEin(art, "text/plain; charset=latin1"), regeln)
		if !istUnsicher(g, model.GrundCharset) || g.aus != "" {
			t.Errorf("%s: charset und zu_gross: %+v", art, g)
		}
		g = redigiere(t, einmal("12"), zielEin(art, "text/plain; charset=latin1"), bodyRegeln(t))
		if !istUnsicher(g, model.GrundCharset) {
			t.Errorf("%s: charset: %+v", art, g)
		}
	}
	// typ vor charset: ohne Ziel ist Text unbekannt.
	g := redigiere(t, einmal("12"), zielEin("", "text/plain; charset=latin1"), bodyRegeln(t))
	if !istUnsicher(g, model.GrundTyp) {
		t.Errorf("typ und charset: %+v", g)
	}
	// Ein Typ, der nicht Text ist, hat keinen Zeichensatz: `application/octet-stream` bleibt typ.
	g = redigiere(t, einmal("12"), zielEin("zaehlung", "application/octet-stream; charset=latin1"), bodyRegeln(t))
	if !istUnsicher(g, model.GrundTyp) {
		t.Errorf("octet-stream: %+v", g)
	}
}

// Z5: Zählung, jedes Byte an erster, mittlerer und letzter Stelle.
func TestTextZaehlungBytes(t *testing.T) {
	zahl := 0
	for b := 0; b < 256; b++ {
		ziffer := b >= '0' && b <= '9'
		c := string([]byte{byte(b)})
		for _, f := range []struct{ body, vor string }{{c + "12", ""}, {"1" + c + "2", "1"}, {"12" + c, "12"}} {
			zahl++
			g := redigiere(t, einmal(f.body), zielEin("zaehlung", "text/plain"), bodyRegeln(t))
			switch {
			case ziffer && (g.err != nil || g.aus != f.body || g.status != model.RedactionNone):
				t.Errorf("Byte %#x in %q: %+v", b, f.body, g)
			case !ziffer && (!istUnsicher(g, model.GrundSyntax) || g.aus != f.vor):
				t.Errorf("Byte %#x in %q: %+v", b, f.body, g)
			}
		}
	}
	if zahl != 768 {
		t.Errorf("%d Fälle", zahl)
	}
	for _, body := range []string{" ", "\n", "12\n", "12\r\n", " 12", "12 ", "\xEF\xBB\xBF12", "-1", "+1", "1.5", "1e3", "0x1", "1,5", "\xD9\xA1", "\xEF\xBC\x91"} {
		g := redigiere(t, einmal(body), zielEin("zaehlung", "text/plain"), bodyRegeln(t))
		if !istUnsicher(g, model.GrundSyntax) {
			t.Errorf("%q: %+v", body, g)
		}
	}
	for _, body := range []string{"0", "00", "007", "9", "1234567890", strings.Repeat("9", 100000)} {
		for _, z := range zerlegungen(body[:min(len(body), 12)]) {
			g := redigiere(t, z.leser(), zielEin("zaehlung", "text/plain"), bodyRegeln(t))
			if g.err != nil || g.aus != body[:min(len(body), 12)] || g.status != model.RedactionNone {
				t.Errorf("%q, %s: %+v", body, z.name, g)
			}
		}
		g := redigiere(t, einmal(body), zielEin("zaehlung", "text/plain"), bodyRegeln(t))
		if g.err != nil || g.aus != body {
			t.Errorf("%d Bytes: %v", len(body), g.err)
		}
	}
}

// Z6: Redaktion des Werts nach dem Namen.
func TestTextWertRedaktion(t *testing.T) {
	treffer := []string{"password", "PASSWORD", "PassWord", "passwd", "secret", "client_secret", "access_token", "refresh_token", "api_key", "API_KEY"}
	inhalte := []string{"S3CR3T-7f", "a", "line1\nline2", " ", zeichenE + zeichenEu + zeichenSm, "[REDACTED] ", " [REDACTED]", "[REDACTED]x", "[REDACTE", "[REDACTED", "[redacted]", "\"x\""}
	n := 0
	for _, name := range treffer {
		for _, in := range inhalte {
			n++
			g := redigiere(t, einmal(in), wertEin(name, "text/plain"), bodyRegeln(t))
			if g.err != nil || g.aus != model.Marker || g.status != model.RedactionApplied {
				t.Errorf("Name %s, Inhalt %q: %+v", name, in, g)
			}
		}
		g := redigiere(t, einmal(model.Marker), wertEin(name, "text/plain"), bodyRegeln(t))
		if g.err != nil || g.aus != model.Marker || g.status != model.RedactionNone {
			t.Errorf("Name %s, Marker schon da: %+v", name, g)
		}
	}
	if n != 120 {
		t.Errorf("%d Fälle", n)
	}
	// Kein Treffer: byte-gleich, none. Der Marker als Inhalt eines nicht treffenden Namens bleibt er selbst.
	for _, name := range []string{"Name", "passwor", "passwords", "xpassword", "pass word", "", "$password", "password "} {
		if name == "" {
			continue
		}
		for _, in := range append(slices.Clone(inhalte), model.Marker) {
			g := redigiere(t, einmal(in), wertEin(name, "text/plain"), bodyRegeln(t))
			if g.err != nil || g.aus != in || g.status != model.RedactionNone {
				t.Errorf("Name %q, Inhalt %q: %+v", name, in, g)
			}
		}
	}
	// Leerer Inhalt: none und nichts vor dem Typ.
	g := redigiere(t, einmal(""), wertEin("password", "text/plain"), bodyRegeln(t))
	if g.err != nil || g.aus != "" || g.status != model.RedactionNone {
		t.Errorf("leer: %+v", g)
	}
	// Die Funktion auf ihre Ausgabe angewendet liefert dieselben Bytes und none.
	for _, name := range []string{"password", "Name"} {
		erst := redigiere(t, einmal("S3CR3T-7f"), wertEin(name, "text/plain"), bodyRegeln(t))
		zweit := redigiere(t, einmal(erst.aus), wertEin(name, "text/plain"), bodyRegeln(t))
		if zweit.err != nil || zweit.aus != erst.aus || zweit.status != model.RedactionNone {
			t.Errorf("Wiederholung %s: %+v nach %+v", name, zweit, erst)
		}
	}
	// Ohne Muster (nil, leer) trifft nichts.
	for _, pfade := range [][]model.JSONPfad{nil, {}} {
		g := redigiere(t, einmal("S3CR3T-7f"), wertEin("password", "text/plain"), model.BodyRegeln{Pfade: pfade, MaxBytes: 100})
		if g.err != nil || g.aus != "S3CR3T-7f" || g.status != model.RedactionNone {
			t.Errorf("ohne Muster: %+v", g)
		}
	}
	// Auch ein eigenes Muster trifft den Namen.
	g = redigiere(t, einmal("v"), wertEin("Token", "text/plain"), model.BodyRegeln{Pfade: pfadeAus(t, "$..token"), MaxBytes: 100})
	if g.aus != model.Marker || g.status != model.RedactionApplied {
		t.Errorf("eigenes Muster: %+v", g)
	}
}

// aufzeichnerWriter hält jeden Write fest.
type aufzeichnerWriter struct{ writes [][]byte }

func (w *aufzeichnerWriter) Write(p []byte) (int, error) {
	w.writes = append(w.writes, slices.Clone(p))
	return len(p), nil
}

func (w *aufzeichnerWriter) alles() []byte { return slices.Concat(w.writes...) }

// Z7: ins dst gelangt nie ein Byte des Geheimnisses.
func TestTextKeinGeheimnisImDst(t *testing.T) {
	const geheim = "s3cr3t-7f9"
	hat := func(dst []byte) bool {
		for i := 0; i < len(geheim); i++ {
			if bytes.IndexByte(dst, geheim[i]) >= 0 {
				return true
			}
		}
		return false
	}
	rest := func(r io.Reader, regeln model.BodyRegeln, grund model.UnsicherGrund, name string) {
		t.Helper()
		w := &aufzeichnerWriter{}
		st, err := model.RedigiereBody(w, r, wertEin("password", "text/plain"), regeln)
		if st != model.RedactionUnsafe || grundVon(err) != grund {
			t.Errorf("%s: Status %q, Fehler %v", name, st, err)
		}
		if hat(w.alles()) || len(w.alles()) != 0 {
			t.Errorf("%s: dst trägt %q", name, w.alles())
		}
	}
	regeln := bodyRegeln(t)
	rest(einmal(geheim+"\xFF"), regeln, model.GrundUTF8, "ungültiges UTF-8")
	rest(einmal(geheim+"\xE2\x82"), regeln, model.GrundUTF8, "halbe Folge am Ende")
	rest(einmal(geheim), model.BodyRegeln{Pfade: regeln.Pfade, MaxBytes: 4}, model.GrundZuGross, "zu_gross")
	rest(&abbruchReader{daten: geheim, k: len(geheim), err: errTest}, regeln, model.GrundEingabe, "Abbruch am Ende")
	rest(&abbruchReader{daten: geheim, k: 3, err: errTest, zusammen: true}, regeln, model.GrundEingabe, "Abbruch mit Daten")
	rest(teile(false, false, einzelneBytes(geheim+"\xFF")...), regeln, model.GrundUTF8, "einzelne Bytes")
	// Bei Erfolg steht genau der Marker, und zwar in einem Write.
	w := &aufzeichnerWriter{}
	st, err := model.RedigiereBody(w, teile(false, false, einzelneBytes(geheim)...), wertEin("password", "text/plain"), regeln)
	if err != nil || st != model.RedactionApplied || string(w.alles()) != model.Marker || len(w.writes) != 1 || hat(w.alles()) {
		t.Errorf("Erfolg: %v, %q", err, w.alles())
	}
	// Ohne Treffer steht das Präfix vor dem Verstoßbyte.
	g := redigiere(t, einmal("ab\xFFcd"), wertEin("Name", "text/plain"), regeln)
	if !istUnsicher(g, model.GrundUTF8) || g.aus != "ab" {
		t.Errorf("ohne Treffer: %+v", g)
	}
}

// Z8: UTF-8 beim Wert, an jeder Teilungsstelle und über jedes Bytepaar.
func TestTextWertUTF8(t *testing.T) {
	zahl := 0
	for _, zeichen := range []string{zeichenE, zeichenEu, zeichenSm, "a", "\x00", "\x7F"} {
		body := "a" + zeichen + "b"
		for _, name := range []string{"password", "Name"} {
			erwartet := body
			if name == "password" {
				erwartet = model.Marker
			}
			for _, z := range zerlegungen(body) {
				zahl++
				g := redigiere(t, z.leser(), wertEin(name, "text/plain"), bodyRegeln(t))
				if g.err != nil || g.aus != erwartet {
					t.Errorf("%q als %s, %s: %+v", body, name, z.name, g)
				}
			}
		}
	}
	if zahl != 84 {
		t.Errorf("Zählprüfung: %d", zahl)
	}
	// Ungültige Folgen, mit dem erwarteten Präfix ohne Treffer.
	ungueltig := []struct{ in, vor string }{
		{"\xC0\x80", ""}, {"\xC1\xBF", ""}, {"\xE0\x80\x80", "\xE0"}, {"\xED\xA0\x80", "\xED"}, {"\xF4\x90\x80\x80", "\xF4"},
		{"\xF0\x80\x80\x80", "\xF0"}, {"\xF5\x80\x80\x80", ""}, {"\xFF", ""}, {"\x80", ""}, {"\xE2\x28\xA1", "\xE2"},
		{"\xC3" + "a", "\xC3"}, {"\xE2\x82", "\xE2\x82"}, {"\xF0\x9F\x98", "\xF0\x9F\x98"}, {"\xBF", ""},
	}
	for _, u := range ungueltig {
		body := "ab" + u.in
		for _, name := range []string{"password", "Name"} {
			vor := "ab" + u.vor
			if name == "password" {
				vor = ""
			}
			for _, z := range zerlegungen(body) {
				g := redigiere(t, z.leser(), wertEin(name, "text/plain"), bodyRegeln(t))
				if !istUnsicher(g, model.GrundUTF8) || g.aus != vor {
					t.Fatalf("%q als %s, %s: %+v, erwartet Präfix %q", body, name, z.name, g, vor)
				}
			}
		}
	}
}

// Z8: Generator über jedes Bytepaar (und die langen Anfänge); das Orakel ist utf8.Valid.
func TestTextWertUTF8Generator(t *testing.T) {
	zahl := 0
	pruefe := func(mitte string) {
		body := "a" + mitte + "z"
		soll := utf8.ValidString(body)
		for _, name := range []string{"password", "Name"} {
			zahl++
			g := redigiere(t, einmal(body), wertEin(name, "text/plain"), bodyRegeln(t))
			if soll != (g.err == nil) || (!soll && grundVon(g.err) != model.GrundUTF8) {
				t.Fatalf("%q als %s: %+v, utf8.Valid %v", body, name, g, soll)
			}
		}
	}
	for b1 := 0; b1 < 256; b1++ {
		for b2 := 0; b2 < 256; b2++ {
			pruefe(string([]byte{byte(b1), byte(b2)}))
		}
	}
	for _, b1 := range []byte{0xE0, 0xE1, 0xEC, 0xED, 0xEE, 0xEF, 0xF0, 0xF1, 0xF3, 0xF4, 0xF5} {
		for b2 := 0; b2 < 256; b2++ {
			for _, b3 := range []byte{0x80, 0x9F, 0xA0, 0xBF, 0xC0, 0x41} {
				pruefe(string([]byte{b1, byte(b2), b3}))
				pruefe(string([]byte{b1, byte(b2), b3, 0x80}))
				pruefe(string([]byte{b1, byte(b2), b3, 0xBF}))
			}
		}
	}
	if zahl != 2*(65536+11*256*6*3) {
		t.Errorf("%d Fälle", zahl)
	}
}

// Z10: Kombinationen der Ziele.
func TestTextZielKombinationen(t *testing.T) {
	zahl := 0
	for bits := 0; bits < 8; bits++ {
		for _, name := range []string{"", "n"} {
			e := model.BodyEingang{ContentType: []string{"text/plain"}, Metadata: bits&1 != 0, Zaehlung: bits&2 != 0, Wert: bits&4 != 0, Name: name}
			gueltig := (bits == 0 && name == "") || (bits == 1 && name == "") || (bits == 2 && name == "") || (bits == 4 && name == "n")
			zahl++
			r := einmal("12")
			w := &zaehlWriter{}
			_, err := model.RedigiereBody(w, r, e, bodyRegeln(t))
			if gueltig == errors.Is(err, model.ErrUngueltig) {
				t.Errorf("Metadata %v, Zaehlung %v, Wert %v, Name %q: %v", e.Metadata, e.Zaehlung, e.Wert, name, err)
			}
			if !gueltig && (r.reads != 0 || w.writes != 0 || errors.Is(err, model.ErrUnsicher)) {
				t.Errorf("Metadata %v, Zaehlung %v, Wert %v, Name %q: %d Reads, %d Writes, %v", e.Metadata, e.Zaehlung, e.Wert, name, r.reads, w.writes, err)
			}
			if !e.Gueltig() != !gueltig {
				t.Errorf("Gueltig() weicht ab: %+v", e)
			}
		}
	}
	if zahl != 16 {
		t.Errorf("%d Kombinationen", zahl)
	}
}

// Z11: die Rangfolge der Nummern gilt für Text.
func TestTextRangfolge(t *testing.T) {
	for _, art := range []string{"zaehlung", "wert"} {
		ein := func(ct string) model.BodyEingang { return zielEin(art, ct) }
		regeln := func(grenze int64) model.BodyRegeln {
			return model.BodyRegeln{Pfade: standardPfade(t), MaxBytes: grenze}
		}
		// leer: none vor dem Typ und vor dem Zeichensatz.
		for _, ct := range []string{"text/plain", "application/octet-stream", "text/plain; charset=latin1", ""} {
			g := redigiere(t, einmal(""), ein(ct), regeln(4))
			if g.err != nil || g.status != model.RedactionNone || g.aus != "" {
				t.Errorf("%s: leer mit %q: %+v", art, ct, g)
			}
		}
		// MaxBytes: genau die Grenze ist zulässig, ein Byte mehr nicht.
		g := redigiere(t, einmal("1234"), ein("text/plain"), regeln(4))
		if g.err != nil || g.aus != "1234" {
			t.Errorf("%s: genau MaxBytes: %+v", art, g)
		}
		g = redigiere(t, einmal("12345"), ein("text/plain"), regeln(4))
		if !istUnsicher(g, model.GrundZuGross) || g.aus != "1234" {
			t.Errorf("%s: MaxBytes+1: %+v", art, g)
		}
		g = redigiere(t, teile(false, false, einzelneBytes("12345")...), ein("text/plain"), regeln(4))
		if !istUnsicher(g, model.GrundZuGross) || g.aus != "1234" {
			t.Errorf("%s: MaxBytes+1 einzeln: %+v", art, g)
		}
		// Fehler von src und dst.
		for k := 0; k <= 2; k++ {
			g = redigiere(t, &abbruchReader{daten: "12", k: k, err: errTest}, ein("text/plain"), regeln(4))
			if !istUnsicher(g, model.GrundEingabe) || !errors.Is(g.err, errTest) {
				t.Errorf("%s: Abbruch nach %d: %+v", art, k, g)
			}
		}
		w := &scheiterWriter{bei: 1, err: errTest}
		st, err := model.RedigiereBody(w, einmal("12"), ein("text/plain"), regeln(4))
		if st != model.RedactionUnsafe || grundVon(err) != model.GrundAusgabe || !errors.Is(err, errTest) {
			t.Errorf("%s: Fehler von dst: %q, %v", art, st, err)
		}
	}
	// utf8 und zu_gross beim Wert: der erste Verstoß im Strom entscheidet.
	for _, name := range []string{"Name", "password"} {
		g := redigiere(t, einmal("abc\xFFd"), wertEin(name, "text/plain"), model.BodyRegeln{Pfade: standardPfade(t), MaxBytes: 4})
		if !istUnsicher(g, model.GrundUTF8) {
			t.Errorf("%s: ungültiges Byte auf MaxBytes: %+v", name, g)
		}
		g = redigiere(t, einmal("abc\xFFd"), wertEin(name, "text/plain"), model.BodyRegeln{Pfade: standardPfade(t), MaxBytes: 3})
		if !istUnsicher(g, model.GrundZuGross) {
			t.Errorf("%s: ungültiges Byte auf MaxBytes+1: %+v", name, g)
		}
	}
}

// Z9: TextZiel, eine Zeile je Pfadform.
func TestTextZiel(t *testing.T) {
	faelle := []struct {
		pfad     string
		zaehlung bool
		wert     bool
		name     string
	}{
		{"/Products/$count", true, false, ""}, {"//$count", true, false, ""}, {"/Products(1)/Orders/$count", true, false, ""},
		{"/a/$count", true, false, ""}, {"x/$count", true, false, ""}, {"/Products/$count/$count", true, false, ""},
		{"/$count", false, false, ""}, {"$count", false, false, ""}, {"/%24count", false, false, ""}, {"/P/%24count", false, false, ""},
		{"/P/$Count", false, false, ""}, {"/P/$count/", false, false, ""}, {"/P/$counts", false, false, ""}, {"/P$count", false, false, ""},
		{"/Products/Name/$value", false, true, "Name"}, {"/Products(1)/Name/$value", false, true, "Name"},
		{"/Products(1)/$value", false, false, ""}, {"/$value", false, false, ""}, {"//$value", false, false, ""},
		{"/P//$value", false, false, ""}, {"$value", false, false, ""}, {"x/$value", false, false, ""}, {"/P/$value/$value", false, false, ""},
		{"/P/a(b/$value", false, false, ""}, {"/P/a)b/$value", false, false, ""}, {"/P/a'b/$value", false, false, ""},
		{"/P/(/$value", false, false, ""}, {"/P/)/$value", false, false, ""}, {"/P/'/$value", false, false, ""},
		{"/P/$x/$value", false, false, ""}, {"/P/%24x/$value", false, true, "$x"}, {"/P/a$x/$value", false, true, "a$x"},
		{"/P/%28/$value", false, true, "("}, {"/P/%27/$value", false, true, "'"}, {"/P('a/password')/$value", false, false, ""},
		{"/P/pass%77ord/$value", false, true, "password"}, {"/P/%zz/$value", false, true, "%zz"}, {"/P/a%2/$value", false, true, "a%2"},
		{"/P/%41%zz/$value", false, true, "%41%zz"}, {"/P/%2541/$value", false, true, "%41"}, {"/P/a%20b/$value", false, true, "a b"},
		{"/P/a+b/$value", false, true, "a+b"}, {"/P/" + zeichenE + "/$value", false, true, zeichenE}, {"/P/%C3%A9/$value", false, true, zeichenE},
		{"/P/%FF/$value", false, true, "\xFF"}, {"/P/Name/$value ", false, false, ""}, {"/P/Name/$values", false, false, ""},
		{"/P/Name/$VALUE", false, false, ""}, {"/P/Name/%24value", false, false, ""}, {"", false, false, ""}, {"/", false, false, ""},
		{"/P/Name", false, false, ""},
	}
	for _, f := range faelle {
		z, w, n := model.TextZiel(f.pfad)
		if z != f.zaehlung || w != f.wert || n != f.name {
			t.Errorf("TextZiel(%q) = %v, %v, %q; erwartet %v, %v, %q", f.pfad, z, w, n, f.zaehlung, f.wert, f.name)
		}
	}
	if len(faelle) != 52 {
		t.Errorf("%d Fälle", len(faelle))
	}
}

// Z9: Generator über Pfade aus Segmenten; das Orakel arbeitet auf den Segmenten, nicht auf dem Suffix.
func TestTextZielGenerator(t *testing.T) {
	segmente := []string{"", "a", "(", ")", "'", "$x", "%28", "%zz", "$count", "$value", "%24count", "P(1)", "%41", "a$"}
	var pfade [][]string
	var baue func(rest int, aktuell []string)
	baue = func(rest int, aktuell []string) {
		if len(aktuell) > 0 {
			pfade = append(pfade, slices.Clone(aktuell))
		}
		if rest == 0 {
			return
		}
		for _, s := range segmente {
			baue(rest-1, append(aktuell, s))
		}
	}
	baue(4, nil)
	if len(pfade) != 14+14*14+14*14*14+14*14*14*14 {
		t.Fatalf("%d Pfade", len(pfade))
	}
	for _, segs := range pfade {
		pfad := "/" + strings.Join(segs, "/")
		letztes := segs[len(segs)-1]
		var zaehlung, wert bool
		var name string
		if letztes == "$count" {
			// Ein Gegenstand davor ist mindestens das Zeichen `/` vor dem letzten Segment.
			zaehlung = len(pfad) > len("/$count")
		}
		if letztes == "$value" && len(segs) >= 2 {
			roh := segs[len(segs)-2]
			if roh != "" && !strings.ContainsAny(roh, "()'") && roh[0] != '$' {
				wert, name = true, roh
				if d, err := url.PathUnescape(roh); err == nil {
					name = d
				}
			}
		}
		z, w, n := model.TextZiel(pfad)
		if z != zaehlung || w != wert || n != name || (z && w) {
			t.Fatalf("TextZiel(%q) = %v, %v, %q; erwartet %v, %v, %q", pfad, z, w, n, zaehlung, wert, name)
		}
	}
}

// Z9: eine Länge ohne Obergrenze und jeder Text sind definiert.
func TestTextZielRein(t *testing.T) {
	lang := "/P/" + strings.Repeat("a", 100000) + "/$value"
	if z, w, n := model.TextZiel(lang); z || !w || len(n) != 100000 {
		t.Errorf("langer Name: %v %v %d", z, w, len(n))
	}
	for b := 0; b < 256; b++ {
		s := "/P/" + string([]byte{byte(b)}) + "/$value"
		z, w, n := model.TextZiel(s)
		soll := b != '(' && b != ')' && b != '\'' && b != '$' && b != '/'
		if z || w != soll {
			t.Errorf("Byte %#x im Namen: %v %v %q", b, z, w, n)
		}
	}
}
