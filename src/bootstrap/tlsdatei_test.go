package bootstrap_test

import (
	"bytes"
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/tlstest"
)

// Die Namen der zwei Verwendungen von ladeTLS (SPEC-291 Nr. 2 Kopf, Nr. 7 c): der OData-Listener und der Admin-Listener.
const (
	listenCert = "LISTEN_TLS_CERT"
	listenKey  = "LISTEN_TLS_KEY"
	adminCert  = "ADMIN_TLS_CERT"
	adminKey   = "ADMIN_TLS_KEY"
)

// Die fünf festen Gründe und der Rang 1 aus der Tabelle von SPEC-291 Nr. 2; das Orakel nennt sie ohne den Code des Bausteins.
const (
	gNichtLesbar = "Datei nicht lesbar"
	gKeinZert    = "Datei enthält kein PEM-Zertifikat"
	gKeinSchl    = "Datei enthält keinen PEM-Schlüssel"
	gAusserhalb  = "Zertifikat außerhalb seiner Gültigkeit"
	gPasstNicht  = "Schlüssel ungültig oder passt nicht zum Zertifikat"
)

func gErforderlich(anderer string) string { return "erforderlich, wenn " + anderer + " gesetzt ist" }

// tlsNamen sind die Schlüsselnamen eines Aufrufers von ladeTLS.
type tlsNamen struct{ cert, key string }

func (n tlsNamen) sechsGruende() []string {
	return []string{gErforderlich(n.cert), gErforderlich(n.key), gNichtLesbar, gKeinZert, gKeinSchl, gAusserhalb, gPasstNicht}
}

func tv(schluessel, grund string) bootstrap.TLSVerstoss {
	return bootstrap.TLSVerstoss{Schluessel: schluessel, Grund: grund}
}

// tlsFall ist ein Element einer Menge: zwei Pfade, die Zeit und die erwartete Meldung (leer: gültig; kein: kein TLS).
type tlsFall struct {
	name      string
	n         tlsNamen
	cert, key string
	jetzt     time.Time
	kein      bool
	erwartet  []bootstrap.TLSVerstoss
}

// werkstatt hält die Zertifikate und das Verzeichnis eines Tests; das Verzeichnis trägt einen Namen, der in keiner Meldung stehen darf.
type werkstatt struct {
	t                      *testing.T
	dir                    string
	zaehler                int
	jetzt                  time.Time
	wurzel                 *tlstest.Aussteller
	gut, abgelaufen, fremd *tlstest.Blatt
	rsa                    *tlstest.Blatt
}

func neueWerkstatt(t *testing.T) *werkstatt {
	t.Helper()
	w := &werkstatt{t: t, dir: filepath.Join(t.TempDir(), "geheim-pfad-4711"), jetzt: time.Now(), wurzel: tlstest.NeueWurzel(t)}
	// Subject geheim-cn-4711 steht in keiner Meldung; abgelaufen gilt eine Stunde bis vor zwei Stunden, fremd ist ein anderes Blatt.
	w.gut = w.wurzel.Blatt(t, tlstest.Optionen{Name: "geheim-cn-4711"})
	w.abgelaufen = w.wurzel.Blatt(t, tlstest.Optionen{Name: "geheim-cn-4711", Von: w.jetzt.Add(-3 * time.Hour), Bis: w.jetzt.Add(-2 * time.Hour)})
	w.fremd = w.wurzel.Blatt(t, tlstest.Optionen{Name: "geheim-cn-4711"})
	return w
}

func (w *werkstatt) datei(inhalt []byte) string { return w.dateiModus(inhalt, 0o600) }

func (w *werkstatt) dateiModus(inhalt []byte, modus os.FileMode) string {
	w.zaehler++
	return tlstest.Datei(w.t, w.dir, strconv.Itoa(w.zaehler)+".pem", inhalt, modus)
}

// fehlt liefert einen Pfad, unter dem keine Datei liegt.
func (w *werkstatt) fehlt() string { return filepath.Join(w.dir, "fehlt.pem") }

func (w *werkstatt) schluessel(b *tlstest.Blatt) []byte { return b.SchluesselPEM(w.t, tlstest.PKCS8) }

// rsaBlatt liefert ein RSA-Blatt, das erst beim ersten Aufruf entsteht.
func (w *werkstatt) rsaBlatt() *tlstest.Blatt {
	if w.rsa == nil {
		w.rsa = w.wurzel.Blatt(w.t, tlstest.Optionen{Art: tlstest.RSA, Name: "geheim-cn-4711"})
	}
	return w.rsa
}

func (w *werkstatt) fall(n tlsNamen, name, cert, key string, erwartet ...bootstrap.TLSVerstoss) tlsFall {
	return tlsFall{name: name, n: n, cert: cert, key: key, jetzt: w.jetzt, erwartet: erwartet}
}

func tlsLade(f tlsFall) (*tls.Config, []bootstrap.TLSVerstoss) {
	return bootstrap.LadeTLS(f.n.cert, f.n.key, f.cert, f.key, f.jetzt)
}

// pruefe führt jedes Element der Menge und vergleicht mit dem Orakel; die Zahl der Elemente ist die der Quelle.
func (w *werkstatt) pruefe(menge string, zahl int, faelle []tlsFall) {
	t := w.t
	t.Helper()
	if len(faelle) != zahl {
		t.Fatalf("Menge %s hat %d Elemente, erwartet %d", menge, len(faelle), zahl)
	}
	for _, f := range faelle {
		cfg, vs := tlsLade(f)
		switch {
		case f.kein:
			if cfg != nil || len(vs) != 0 {
				t.Errorf("%s %q: %+v, %v; erwartet kein TLS (nil, nil)", menge, f.name, cfg != nil, vs)
			}
		case len(f.erwartet) == 0:
			if cfg == nil || len(vs) != 0 {
				t.Errorf("%s %q: Konfiguration %v, Verstöße %v; erwartet gültig", menge, f.name, cfg != nil, vs)
			}
		default:
			if cfg != nil || !slices.Equal(vs, f.erwartet) {
				t.Errorf("%s %q: Konfiguration %v, Verstöße %v; erwartet %v", menge, f.name, cfg != nil, vs, f.erwartet)
			}
		}
	}
}

// ladeGueltig lädt Zertifikatsdatei und Schlüssel mit den Namen des OData-Listeners und verlangt kein Verstoß.
func (w *werkstatt) ladeGueltig(cert, key []byte) *tls.Config {
	t := w.t
	t.Helper()
	cfg, vs := bootstrap.LadeTLS(listenCert, listenKey, w.datei(cert), w.datei(key), w.jetzt)
	if cfg == nil || len(vs) != 0 {
		t.Fatalf("gültiges Paar: Konfiguration %v, Verstöße %v", cfg != nil, vs)
	}
	return cfg
}

// beschreibe nennt die Eigenschaften einer Konfiguration, die zwei Aufrufe gleich haben müssen.
func beschreibe(cfg *tls.Config) string {
	if cfg == nil {
		return "nil"
	}
	return fmt.Sprintf("%x|%d|%q", cfg.Certificates[0].Certificate, cfg.MinVersion, cfg.NextProtos)
}

// mengeE: genau einer der beiden Pfade leer, die gesetzte Datei gültig oder fehlend (SPEC-291 Nr. 2, Rang 1).
func (w *werkstatt) mengeE(n tlsNamen) []tlsFall {
	gutCert, gutKey, fehlt := w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut)), w.fehlt()
	return []tlsFall{
		w.fall(n, "Zertifikat leer, Schlüssel gültig", "", gutKey, tv(n.cert, gErforderlich(n.key))),
		w.fall(n, "Zertifikat leer, Schlüssel fehlt", "", fehlt, tv(n.cert, gErforderlich(n.key)), tv(n.key, gNichtLesbar)),
		w.fall(n, "Schlüssel leer, Zertifikat gültig", gutCert, "", tv(n.key, gErforderlich(n.cert))),
		w.fall(n, "Schlüssel leer, Zertifikat fehlt", fehlt, "", tv(n.cert, gNichtLesbar), tv(n.key, gErforderlich(n.cert))),
	}
}

// mengeL: Rang 2, drei Formen je Schlüssel: Datei fehlt, Pfad ist ein Verzeichnis, Pfad trägt ein NUL.
func (w *werkstatt) mengeL(n tlsNamen) []tlsFall {
	gutCert, gutKey := w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut))
	var faelle []tlsFall
	for _, f := range []struct{ name, pfad string }{{"Datei fehlt", w.fehlt()}, {"Verzeichnis", w.dir}, {"NUL im Pfad", w.dir + "/nul\x00datei.pem"}} {
		faelle = append(faelle,
			w.fall(n, "Zertifikat: "+f.name, f.pfad, gutKey, tv(n.cert, gNichtLesbar)),
			w.fall(n, "Schlüssel: "+f.name, gutCert, f.pfad, tv(n.key, gNichtLesbar)))
	}
	return faelle
}

// mengeZ: Rang 3 des Zertifikats, sieben Formen eines Inhalts ohne lesbares erstes Blatt.
func (w *werkstatt) mengeZ(n tlsNamen) []tlsFall {
	gutKey := w.datei(w.schluessel(w.gut))
	zert := w.gut.ZertifikatPEM()
	fremd := tlstest.Block("CERTIFICATE", []byte("kein Zertifikat"))
	var faelle []tlsFall
	for _, f := range []struct {
		name   string
		inhalt []byte
	}{
		{"leere Datei", nil},
		{"nur Text", []byte("das ist kein PEM\n")},
		{"nur ein Schlüsselblock", w.schluessel(w.gut)},
		{"Block CERTIFICATE mit fremden Bytes", fremd},
		{"Block CERTIFICATE ohne END", zert[:bytes.Index(zert, []byte("-----END"))]},
		{"nur ein Block X509 CRL", tlstest.Block("X509 CRL", []byte{1, 2, 3})},
		{"erster Block ungültig, zweiter gültig", slices.Concat(fremd, zert)},
	} {
		faelle = append(faelle, w.fall(n, f.name, w.datei(f.inhalt), gutKey, tv(n.cert, gKeinZert)))
	}
	return faelle
}

// mengeS: Rang 3 des Schlüssels, vier Formen eines Inhalts ohne Block `PRIVATE KEY` oder `… PRIVATE KEY`.
func (w *werkstatt) mengeS(n tlsNamen) []tlsFall {
	gutCert := w.datei(w.gut.ZertifikatPEM())
	schl := w.schluessel(w.gut)
	var faelle []tlsFall
	for _, f := range []struct {
		name   string
		inhalt []byte
	}{
		{"leere Datei", nil},
		{"nur Text", []byte("das ist kein PEM\n")},
		{"nur ein Zertifikatblock", w.gut.ZertifikatPEM()},
		{"Block PRIVATE KEY ohne END", schl[:bytes.Index(schl, []byte("-----END"))]},
	} {
		faelle = append(faelle, w.fall(n, f.name, gutCert, w.datei(f.inhalt), tv(n.key, gKeinSchl)))
	}
	return faelle
}

// mengeP: Rang 4 des Schlüssels, sechs Formen, die nicht zum ersten Zertifikat der Datei passen.
func (w *werkstatt) mengeP(n tlsNamen) []tlsFall {
	gutCert := w.datei(w.gut.ZertifikatPEM())
	zw := w.wurzel.Zwischen(w.t, tlstest.Optionen{})
	unter := zw.Blatt(w.t, tlstest.Optionen{})
	mitZwischen := slices.Concat(tlstest.Block("CERTIFICATE", zw.Zertifikat.Raw), unter.ZertifikatPEM())
	var faelle []tlsFall
	for _, f := range []struct {
		name         string
		cert, inhalt []byte
	}{
		{"Schlüssel eines anderen Blatts gleicher Art", nil, w.schluessel(w.fremd)},
		{"Schlüssel anderer Art", nil, w.schluessel(w.rsaBlatt())},
		{"Block PRIVATE KEY mit fremden Bytes", nil, tlstest.Block("PRIVATE KEY", []byte("kein Schlüssel"))},
		{"Block ENCRYPTED PRIVATE KEY", nil, tlstest.Block("ENCRYPTED PRIVATE KEY", []byte("verschlüsselt"))},
		{"Block EC PRIVATE KEY mit fremden Bytes", nil, tlstest.Block("EC PRIVATE KEY", []byte("kein Schlüssel"))},
		{"Zwischenzertifikat vor dem Blatt", mitZwischen, w.schluessel(unter)},
	} {
		cert := gutCert
		if f.cert != nil {
			cert = w.datei(f.cert)
		}
		faelle = append(faelle, w.fall(n, f.name, cert, w.datei(f.inhalt), tv(n.key, gPasstNicht)))
	}
	return faelle
}

// mengeG: die Gültigkeit des Blatts zur Zeit jetzt (SPEC-291 Nr. 3): beide Grenzen, je eine Nanosekunde daneben, die Mitte, und ein
// abgelaufenes Zwischenzertifikat hinter einem gültigen Blatt.
func (w *werkstatt) mengeG(n tlsNamen) []tlsFall {
	g := w.wurzel.Blatt(w.t, tlstest.Optionen{})
	von, bis := g.Zertifikat.NotBefore, g.Zertifikat.NotAfter
	cert, key := w.datei(g.ZertifikatPEM()), w.datei(w.schluessel(g))
	zw := w.wurzel.Zwischen(w.t, tlstest.Optionen{Von: w.jetzt.Add(-3 * time.Hour), Bis: w.jetzt.Add(-2 * time.Hour)})
	unter := zw.Blatt(w.t, tlstest.Optionen{})
	aussen := tv(n.cert, gAusserhalb)
	zeiten := []struct {
		name  string
		jetzt time.Time
		aus   []bootstrap.TLSVerstoss
	}{
		{"NotBefore minus 1 ns", von.Add(-time.Nanosecond), []bootstrap.TLSVerstoss{aussen}},
		{"NotBefore", von, nil},
		{"Mitte", von.Add(bis.Sub(von) / 2), nil},
		{"NotAfter", bis, nil},
		{"NotAfter plus 1 ns", bis.Add(time.Nanosecond), []bootstrap.TLSVerstoss{aussen}},
	}
	var faelle []tlsFall
	for _, z := range zeiten {
		f := w.fall(n, z.name, cert, key, z.aus...)
		f.jetzt = z.jetzt
		faelle = append(faelle, f)
	}
	return append(faelle, w.fall(n, "gültiges Blatt, abgelaufenes Zwischenzertifikat", w.datei(unter.KettePEM()), w.datei(w.schluessel(unter))))
}

// Die Zustände einer der beiden Dateien in der Menge Q; zBesonders ist beim Zertifikat der Zustand außerhalb der Gültigkeit und beim
// Schlüssel der Zustand passt nicht zum Zertifikat.
const (
	zGueltig = iota
	zLeer
	zNichtLesbar
	zKeinPEM
	zBesonders
	zAnzahl
)

// paar bildet das Element der Menge Q aus dem Zustand c des Zertifikats und k des Schlüssels; ist der Schlüssel gültig, gehört er
// zum Zertifikat des Paares.
func (w *werkstatt) paar(n tlsNamen, c, k int) tlsFall {
	blatt := w.gut
	if c == zBesonders {
		blatt = w.abgelaufen
	}
	var cert, key string
	switch c {
	case zGueltig, zBesonders:
		cert = w.datei(blatt.ZertifikatPEM())
	case zNichtLesbar:
		cert = w.fehlt()
	case zKeinPEM:
		cert = w.datei([]byte("kein PEM\n"))
	}
	switch k {
	case zGueltig:
		key = w.datei(w.schluessel(blatt))
	case zNichtLesbar:
		key = w.fehlt()
	case zKeinPEM:
		key = w.datei([]byte("kein PEM\n"))
	case zBesonders:
		key = w.datei(w.schluessel(w.fremd))
	}
	f := w.fall(n, fmt.Sprintf("Zertifikat %d, Schlüssel %d", c, k), cert, key, erwartetPaar(n, c, k)...)
	f.kein = c == zLeer && k == zLeer
	return f
}

// erwartetPaar ist das Orakel der Menge Q nach der Tabelle von SPEC-291 Nr. 2: je Schlüssel ein Rang, der Rang 4 des Schlüssels nur,
// wenn beide den Rang 3 bestanden haben.
func erwartetPaar(n tlsNamen, c, k int) []bootstrap.TLSVerstoss {
	if c == zLeer && k == zLeer {
		return nil
	}
	var aus []bootstrap.TLSVerstoss
	switch c {
	case zLeer:
		aus = append(aus, tv(n.cert, gErforderlich(n.key)))
	case zNichtLesbar:
		aus = append(aus, tv(n.cert, gNichtLesbar))
	case zKeinPEM:
		aus = append(aus, tv(n.cert, gKeinZert))
	case zBesonders:
		aus = append(aus, tv(n.cert, gAusserhalb))
	}
	switch k {
	case zLeer:
		aus = append(aus, tv(n.key, gErforderlich(n.cert)))
	case zNichtLesbar:
		aus = append(aus, tv(n.key, gNichtLesbar))
	case zKeinPEM:
		aus = append(aus, tv(n.key, gKeinSchl))
	case zBesonders:
		if c == zGueltig || c == zBesonders {
			aus = append(aus, tv(n.key, gPasstNicht))
		}
	}
	return aus
}

func (w *werkstatt) mengeQ(n tlsNamen) []tlsFall {
	var faelle []tlsFall
	for c := range zAnzahl {
		for k := range zAnzahl {
			faelle = append(faelle, w.paar(n, c, k))
		}
	}
	return faelle
}

// mengeF: die zugesagten Formen der Dateien (SPEC-291 Nr. 4), aus Bytewerten gebaut (AGENTS.md §3.18); Schlüsseldatei mit 0666 (Nr. 10).
func (w *werkstatt) mengeF(n tlsNamen) []tlsFall {
	zw := w.wurzel.Zwischen(w.t, tlstest.Optionen{})
	kette := zw.Blatt(w.t, tlstest.Optionen{})
	cert, key := w.gut.ZertifikatPEM(), w.schluessel(w.gut)
	paar := func(name string, f func([]byte) []byte) tlsFall {
		return w.fall(n, name, w.datei(f(cert)), w.datei(f(key)))
	}
	beide := w.datei(slices.Concat(cert, key))
	return []tlsFall{
		paar("Zeilenende LF", func(b []byte) []byte { return b }),
		paar("Zeilenende CRLF", func(b []byte) []byte { return bytes.ReplaceAll(b, []byte("\n"), []byte("\r\n")) }),
		paar("Text vor den Blöcken", func(b []byte) []byte { return slices.Concat([]byte("Fremdtext vor dem Block\n"), b) }),
		paar("Text nach den Blöcken", func(b []byte) []byte { return slices.Concat(b, []byte("Fremdtext nach dem Block\n")) }),
		w.fall(n, "Kette Blatt und Zwischenzertifikat", w.datei(kette.KettePEM()), w.datei(w.schluessel(kette))),
		w.fall(n, "Zertifikat und Schlüssel in einer Datei", beide, beide),
		paar("kein Zeilenende am Ende", func(b []byte) []byte { return bytes.TrimSuffix(b, []byte("\n")) }),
		w.fall(n, "Schlüsseldatei 0666", w.datei(cert), w.dateiModus(key, 0o666)),
	}
}

// TestLadeTLSKeinTLS (A01): beide Pfade leer heißt kein TLS; im Arbeitsverzeichnis liegen gültige Dateien der Namen cert.pem und key.pem,
// die nichts liest.
func TestLadeTLSKeinTLS(t *testing.T) {
	w := neueWerkstatt(t)
	arbeit := filepath.Join(t.TempDir(), "arbeit")
	tlstest.Datei(t, arbeit, "cert.pem", w.gut.ZertifikatPEM(), 0o600)
	tlstest.Datei(t, arbeit, "key.pem", w.schluessel(w.gut), 0o600)
	t.Chdir(arbeit)
	cfg, vs := bootstrap.LadeTLS(listenCert, listenKey, "", "", w.jetzt)
	if cfg != nil || vs != nil {
		t.Fatalf("Konfiguration %v, Verstöße %v; erwartet (nil, nil)", cfg != nil, vs)
	}
}

// TestLadeTLSEinerLeer (A02): genau ein Pfad leer; die gesetzte Datei wird weiter geprüft (Menge E, 4 Elemente).
func TestLadeTLSEinerLeer(t *testing.T) {
	w := neueWerkstatt(t)
	w.pruefe("E", 4, w.mengeE(tlsNamen{listenCert, listenKey}))
}

// TestLadeTLSNichtLesbar (A03): Rang 2 (Menge L, 6 Elemente).
func TestLadeTLSNichtLesbar(t *testing.T) {
	w := neueWerkstatt(t)
	w.pruefe("L", 6, w.mengeL(tlsNamen{listenCert, listenKey}))
}

// TestLadeTLSKeinPEMZertifikat (A04): Rang 3 des Zertifikats (Menge Z, 7 Elemente).
func TestLadeTLSKeinPEMZertifikat(t *testing.T) {
	w := neueWerkstatt(t)
	w.pruefe("Z", 7, w.mengeZ(tlsNamen{listenCert, listenKey}))
}

// TestLadeTLSKeinPEMSchluessel (A05): Rang 3 des Schlüssels (Menge S, 4 Elemente).
func TestLadeTLSKeinPEMSchluessel(t *testing.T) {
	w := neueWerkstatt(t)
	w.pruefe("S", 4, w.mengeS(tlsNamen{listenCert, listenKey}))
}

// TestLadeTLSSchluesselPasstNicht (A06): Rang 4 des Schlüssels (Menge P, 6 Elemente).
func TestLadeTLSSchluesselPasstNicht(t *testing.T) {
	w := neueWerkstatt(t)
	w.pruefe("P", 6, w.mengeP(tlsNamen{listenCert, listenKey}))
}

// TestLadeTLSGueltigkeit (A07): Rang 4 des Zertifikats (Menge G, 6 Elemente); dazu jede Zeit aus Abstand und Seite beider Grenzen, geprüft
// gegen NotBefore ≤ jetzt ≤ NotAfter in Nanosekunden (Menge G+, 16 Elemente: 2 Grenzen × 7 Abstände, die Zeit null, das Jahr 9999).
func TestLadeTLSGueltigkeit(t *testing.T) {
	w := neueWerkstatt(t)
	n := tlsNamen{listenCert, listenKey}
	w.pruefe("G", 6, w.mengeG(n))
	g := w.wurzel.Blatt(t, tlstest.Optionen{})
	von, bis := g.Zertifikat.NotBefore, g.Zertifikat.NotAfter
	cert, key := w.datei(g.ZertifikatPEM()), w.datei(w.schluessel(g))
	abstaende := []time.Duration{-48 * time.Hour, -time.Second, -time.Nanosecond, 0, time.Nanosecond, time.Second, 48 * time.Hour}
	var faelle []tlsFall
	for _, grenze := range []time.Time{von, bis} {
		for _, d := range abstaende {
			f := w.fall(n, fmt.Sprintf("%v%+d", grenze.Equal(von), d), cert, key)
			f.jetzt = grenze.Add(d)
			if f.jetzt.UnixNano() < von.UnixNano() || f.jetzt.UnixNano() > bis.UnixNano() {
				f.erwartet = []bootstrap.TLSVerstoss{tv(n.cert, gAusserhalb)}
			}
			faelle = append(faelle, f)
		}
	}
	null, ende := w.fall(n, "Zeit null", cert, key, tv(n.cert, gAusserhalb)), w.fall(n, "Jahr 9999", cert, key, tv(n.cert, gAusserhalb))
	null.jetzt, ende.jetzt = time.Time{}, time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)
	w.pruefe("G+", 16, append(faelle, null, ende))
}

// TestLadeTLSPaare (A08): je Schlüssel ein Grund, die Schlüssel melden unabhängig, Rang n nur nach Rang n-1, der Rang 4 des Schlüssels nur nach
// Rang 3 beider (Menge Q, 25 Paare).
func TestLadeTLSPaare(t *testing.T) {
	w := neueWerkstatt(t)
	w.pruefe("Q", 25, w.mengeQ(tlsNamen{listenCert, listenKey}))
}

// TestLadeTLSFormen (A09): die zugesagten Formen der Dateien sind gültig (Menge F, 8 Elemente); dazu die Prüfgrenze jeder Länge: jeder
// Anfang einer Datei ist gültig genau dann, wenn die END-Zeile vollständig ist (Menge T, je Länge der Zertifikats- und der Schlüsseldatei).
func TestLadeTLSFormen(t *testing.T) {
	w := neueWerkstatt(t)
	n := tlsNamen{listenCert, listenKey}
	w.pruefe("F", 8, w.mengeF(n))
	cert, key := w.gut.ZertifikatPEM(), w.schluessel(w.gut)
	certPfad, keyPfad := w.datei(cert), w.datei(key)
	zahl := 0
	for _, seite := range []struct {
		name  string
		voll  []byte
		pfad  string
		fehlt bootstrap.TLSVerstoss
	}{{"Zertifikat", cert, certPfad, tv(n.cert, gKeinZert)}, {"Schlüssel", key, keyPfad, tv(n.key, gKeinSchl)}} {
		for laenge := 0; laenge <= len(seite.voll); laenge++ {
			if err := os.WriteFile(seite.pfad, seite.voll[:laenge], 0o600); err != nil {
				t.Fatal(err)
			}
			zahl++
			f := tlsFall{name: fmt.Sprintf("%s, Länge %d", seite.name, laenge), n: n, cert: certPfad, key: keyPfad, jetzt: w.jetzt}
			if laenge < len(seite.voll)-1 {
				f.erwartet = []bootstrap.TLSVerstoss{seite.fehlt}
			}
			w.pruefe("T", 1, []tlsFall{f})
		}
		if err := os.WriteFile(seite.pfad, seite.voll, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if zahl != len(cert)+1+len(key)+1 {
		t.Fatalf("Menge T hat %d Elemente, erwartet %d", zahl, len(cert)+len(key)+2)
	}
}

// TestLadeTLSFremdtext: jeder Fremdtext an jeder Stelle der Dateien ändert nichts (Menge X, 5 Texte × 5 Stellen): vor, zwischen und nach der
// Kette der Zertifikatsdatei, vor und nach dem Block der Schlüsseldatei.
func TestLadeTLSFremdtext(t *testing.T) {
	w := neueWerkstatt(t)
	n := tlsNamen{listenCert, listenKey}
	zw := w.wurzel.Zwischen(t, tlstest.Optionen{})
	b := zw.Blatt(t, tlstest.Optionen{})
	blatt, unter := b.ZertifikatPEM(), tlstest.Block("CERTIFICATE", b.Kette[0])
	key := w.schluessel(b)
	var faelle []tlsFall
	for _, text := range []string{"Fremdtext\n", "\n\n\n", "\t \n", "Zeile ohne Ende, dann Zeilenende\r\n", "# Kommentar: -----BEGIN fehlt\n"} {
		tx := []byte(text)
		stellen := []struct {
			name      string
			cert, key []byte
		}{
			{"vor der Kette", slices.Concat(tx, blatt, unter), key},
			{"zwischen Blatt und Zwischenzertifikat", slices.Concat(blatt, tx, unter), key},
			{"nach der Kette", slices.Concat(blatt, unter, tx), key},
			{"vor dem Schlüssel", slices.Concat(blatt, unter), slices.Concat(tx, key)},
			{"nach dem Schlüssel", slices.Concat(blatt, unter), slices.Concat(key, tx)},
		}
		for _, s := range stellen {
			faelle = append(faelle, w.fall(n, fmt.Sprintf("%q %s", text, s.name), w.datei(s.cert), w.datei(s.key)))
		}
	}
	w.pruefe("X", 25, faelle)
}

// TestLadeTLSBlocktypen: der Typ des Blocks entscheidet, was Zertifikat und Schlüssel sind (SPEC-291 Nr. 4): `CERTIFICATE` allein ist ein
// Zertifikat, `PRIVATE KEY` und jeder Typ mit der Endung ` PRIVATE KEY` ein Schlüssel; jeder andere Typ trägt gültige Bytes und zählt nicht
// (Menge B, 10 Typen für den Schlüssel und 6 für das Zertifikat).
func TestLadeTLSBlocktypen(t *testing.T) {
	w := neueWerkstatt(t)
	n := tlsNamen{listenCert, listenKey}
	schlDER, _ := pem.Decode(w.schluessel(w.gut))
	gutCert, gutKey := w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut))
	var faelle []tlsFall
	for typ, gilt := range map[string]bool{
		"PRIVATE KEY": true, "RSA PRIVATE KEY": true, "EC PRIVATE KEY": true, "OPENSSH PRIVATE KEY": true,
		"PUBLIC KEY": false, "XPRIVATE KEY": false, "PRIVATE KEYS": false, "private key": false, "PRIVATE  KEY": false, "CERTIFICATE": false,
	} {
		f := w.fall(n, "Schlüssel "+typ, gutCert, w.datei(tlstest.Block(typ, schlDER.Bytes)))
		if !gilt {
			f.erwartet = []bootstrap.TLSVerstoss{tv(n.key, gKeinSchl)}
		}
		faelle = append(faelle, f)
	}
	for typ, gilt := range map[string]bool{
		"CERTIFICATE": true, "X509 CERTIFICATE": false, "TRUSTED CERTIFICATE": false, "CERTIFICATE REQUEST": false, "certificate": false, "PUBLIC KEY": false,
	} {
		f := w.fall(n, "Zertifikat "+typ, w.datei(tlstest.Block(typ, w.gut.Zertifikat.Raw)), gutKey)
		if !gilt {
			f.erwartet = []bootstrap.TLSVerstoss{tv(n.cert, gKeinZert)}
		}
		faelle = append(faelle, f)
	}
	w.pruefe("B", 16, faelle)
}

// TestLadeTLSPEMHeader: PEM-Header eines Blocks bleiben ohne Wirkung auf das Zertifikat (SPEC-291 Nr. 4); ein Schlüssel im alten Format
// mit den Headern `Proc-Type` und `DEK-Info` und verschlüsselten Bytes ist ungültig, denn es gibt keine Passphrase (Menge H, 2 Elemente).
func TestLadeTLSPEMHeader(t *testing.T) {
	w := neueWerkstatt(t)
	n := tlsNamen{listenCert, listenKey}
	kopf := map[string]string{"Proc-Type": "4,ENCRYPTED", "DEK-Info": "AES-128-CBC,00112233445566778899AABBCCDDEEFF"}
	mitKopf := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Headers: kopf, Bytes: w.gut.Zertifikat.Raw})
	verschluesselt := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Headers: kopf, Bytes: []byte("chiffre")})
	w.pruefe("H", 2, []tlsFall{
		w.fall(n, "Zertifikat mit Header", w.datei(mitKopf), w.datei(w.schluessel(w.gut))),
		w.fall(n, "verschlüsselter Schlüssel mit Header", w.datei(w.gut.ZertifikatPEM()), w.datei(verschluesselt), tv(n.key, gPasstNicht)),
	})
}

// TestLadeTLSPfadWieUebergeben: ladeTLS kürzt keinen Pfad (SPEC-291 Nr. 2, Leerraum entfernt die Konfiguration): ein Pfad mit Leerraum am
// Rand oder allein aus Leerraum ist ein Pfad, der nicht lesbar ist, nicht ein leerer.
func TestLadeTLSPfadWieUebergeben(t *testing.T) {
	w := neueWerkstatt(t)
	n := tlsNamen{listenCert, listenKey}
	cert, key := w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut))
	w.pruefe("Pfad", 4, []tlsFall{
		w.fall(n, "Zertifikat mit Leerraum hinter dem Pfad", cert+" ", key, tv(n.cert, gNichtLesbar)),
		w.fall(n, "Zertifikat allein Leerraum", " ", key, tv(n.cert, gNichtLesbar)),
		w.fall(n, "Schlüssel mit Tabulator hinter dem Pfad", cert, key+"\t", tv(n.key, gNichtLesbar)),
		w.fall(n, "Schlüssel allein Leerraum", cert, "\n", tv(n.key, gNichtLesbar)),
	})
}

// TestLadeTLSArten: das Paar passt genau dann, wenn der Schlüssel zum Blatt gehört, in jeder Paarung der Arten (Menge R, 3 Arten für das
// Blatt × 3 Arten für den fremden Schlüssel, dazu je Art das eigene Paar; 12 Elemente).
func TestLadeTLSArten(t *testing.T) {
	w := neueWerkstatt(t)
	n := tlsNamen{listenCert, listenKey}
	arten := []tlstest.Art{tlstest.ECDSA, tlstest.RSA, tlstest.Ed25519}
	blaetter := make([]*tlstest.Blatt, len(arten))
	fremde := make([]*tlstest.Blatt, len(arten))
	for i, a := range arten {
		blaetter[i] = w.wurzel.Blatt(t, tlstest.Optionen{Art: a})
		fremde[i] = w.wurzel.Blatt(t, tlstest.Optionen{Art: a})
	}
	var faelle []tlsFall
	for i, b := range blaetter {
		cert := w.datei(b.ZertifikatPEM())
		faelle = append(faelle, w.fall(n, fmt.Sprintf("Art %d eigenes Paar", i), cert, w.datei(w.schluessel(b))))
		for j, fr := range fremde {
			faelle = append(faelle, w.fall(n, fmt.Sprintf("Blatt Art %d, Schlüssel Art %d", i, j), cert, w.datei(w.schluessel(fr)), tv(n.key, gPasstNicht)))
		}
	}
	w.pruefe("R", 12, faelle)
}

// TestLadeTLSSchluesselformen (A10): die fünf Schlüsselformen sind gültig und tragen einen Handshake (Menge K, 5 Elemente).
func TestLadeTLSSchluesselformen(t *testing.T) {
	w := neueWerkstatt(t)
	formen := []struct {
		name string
		art  tlstest.Art
		form tlstest.Form
	}{
		{"RSA als RSA PRIVATE KEY", tlstest.RSA, tlstest.PKCS1},
		{"RSA als PRIVATE KEY", tlstest.RSA, tlstest.PKCS8},
		{"ECDSA als EC PRIVATE KEY", tlstest.ECDSA, tlstest.SEC1},
		{"ECDSA als PRIVATE KEY", tlstest.ECDSA, tlstest.PKCS8},
		{"Ed25519 als PRIVATE KEY", tlstest.Ed25519, tlstest.PKCS8},
	}
	if len(formen) != 5 {
		t.Fatalf("Menge K hat %d Elemente, erwartet 5", len(formen))
	}
	for _, f := range formen {
		b := w.wurzel.Blatt(t, tlstest.Optionen{Art: f.art})
		cfg := w.ladeGueltig(b.ZertifikatPEM(), b.SchluesselPEM(t, f.form))
		if r := tlstest.Handshake(t, cfg, w.wurzel.Klient("localhost")); r.ServerFehler != nil || r.ClientFehler != nil {
			t.Errorf("%s: Handshake: Server %v, Client %v", f.name, r.ServerFehler, r.ClientFehler)
		}
	}
}

// TestLadeTLSVersion (A11): Mindestversion TLS 1.2, Höchstversion und Suiten der Bibliothek, keine Client-Zertifikate, kein Nachladen.
func TestLadeTLSVersion(t *testing.T) {
	w := neueWerkstatt(t)
	b := w.wurzel.Blatt(t, tlstest.Optionen{})
	cfg := w.ladeGueltig(b.ZertifikatPEM(), w.schluessel(b))
	if cfg.MinVersion != tls.VersionTLS12 || cfg.MaxVersion != 0 || cfg.CipherSuites != nil || cfg.ClientAuth != tls.NoClientCert ||
		cfg.ClientCAs != nil || cfg.GetCertificate != nil || cfg.GetConfigForClient != nil || len(cfg.Certificates) != 1 {
		t.Errorf("Felder: MinVersion %x, MaxVersion %x, Suiten %v, ClientAuth %v, ClientCAs %v, GetCertificate %v, GetConfigForClient %v, Paare %d",
			cfg.MinVersion, cfg.MaxVersion, cfg.CipherSuites, cfg.ClientAuth, cfg.ClientCAs != nil, cfg.GetCertificate != nil,
			cfg.GetConfigForClient != nil, len(cfg.Certificates))
	}
	for _, c := range []struct {
		name      string
		min, max  uint16
		scheitert bool
		version   uint16
	}{
		{"Client bis TLS 1.1", tls.VersionTLS10, tls.VersionTLS11, true, 0},
		{"Client bis TLS 1.2", 0, tls.VersionTLS12, false, tls.VersionTLS12},
		{"Client ohne Grenze, ohne Zertifikat", 0, 0, false, tls.VersionTLS13},
	} {
		client := w.wurzel.Klient("localhost")
		client.MinVersion, client.MaxVersion = c.min, c.max
		r := tlstest.Handshake(t, cfg, client)
		if (r.ServerFehler != nil) != c.scheitert || (r.ClientFehler != nil) != c.scheitert || r.Version != c.version {
			t.Errorf("%s: Server %v, Client %v, Version %x; erwartet scheitert=%v, Version %x", c.name, r.ServerFehler, r.ClientFehler, r.Version, c.scheitert, c.version)
		}
	}
}

// TestLadeTLSALPN (A12): ALPN nennt allein `http/1.1`.
func TestLadeTLSALPN(t *testing.T) {
	w := neueWerkstatt(t)
	b := w.wurzel.Blatt(t, tlstest.Optionen{})
	cfg := w.ladeGueltig(b.ZertifikatPEM(), w.schluessel(b))
	if !slices.Equal(cfg.NextProtos, []string{"http/1.1"}) {
		t.Errorf("NextProtos %q, erwartet [http/1.1]", cfg.NextProtos)
	}
	for _, c := range []struct {
		name      string
		protos    []string
		scheitert bool
		erwartet  string
	}{
		{"h2 und http/1.1", []string{"h2", "http/1.1"}, false, "http/1.1"},
		{"ohne ALPN", nil, false, ""},
		{"allein h2", []string{"h2"}, true, ""},
	} {
		client := w.wurzel.Klient("localhost")
		client.NextProtos = c.protos
		r := tlstest.Handshake(t, cfg, client)
		if (r.ServerFehler != nil) != c.scheitert || (r.ClientFehler != nil) != c.scheitert || r.Protokoll != c.erwartet {
			t.Errorf("%s: Server %v, Client %v, Protokoll %q; erwartet scheitert=%v, Protokoll %q", c.name, r.ServerFehler, r.ClientFehler, r.Protokoll, c.scheitert, c.erwartet)
		}
	}
}

// TestLadeTLSKette (A13): die Kette der Datei wird gesendet; ein Client mit nur der Wurzel vertraut Blatt und Zwischenzertifikat. Die
// Gegenprobe ist dieselbe Prüfung mit der Datei des Blatts allein: der Client lehnt ab.
func TestLadeTLSKette(t *testing.T) {
	w := neueWerkstatt(t)
	zw := w.wurzel.Zwischen(t, tlstest.Optionen{})
	b := zw.Blatt(t, tlstest.Optionen{})
	cfg := w.ladeGueltig(b.KettePEM(), w.schluessel(b))
	if len(cfg.Certificates) != 1 || len(cfg.Certificates[0].Certificate) != 2 {
		t.Fatalf("Paare %d, Zertifikate der Kette %d; erwartet 1 und 2", len(cfg.Certificates), len(cfg.Certificates[0].Certificate))
	}
	if r := tlstest.Handshake(t, cfg, w.wurzel.Klient("localhost")); r.ServerFehler != nil || r.ClientFehler != nil {
		t.Errorf("Kette: Server %v, Client %v", r.ServerFehler, r.ClientFehler)
	}
	allein := w.ladeGueltig(b.ZertifikatPEM(), w.schluessel(b))
	if r := tlstest.Handshake(t, allein, w.wurzel.Klient("localhost")); r.ClientFehler == nil {
		t.Errorf("Blatt allein: der Client nahm ein Blatt ohne Zwischenzertifikat an")
	}
}

// TestLadeTLSKeinGeheimnis (A14): jede Meldung der Mengen E, L, Z, S, P und G (33 Elemente) ist einer der sechs Texte und nennt weder Pfad noch
// Subject noch Schlüsselbytes noch Text der Bibliothek; der Quelltext trägt genau die fünf festen Literale (go/parser).
func TestLadeTLSKeinGeheimnis(t *testing.T) {
	w := neueWerkstatt(t)
	n := tlsNamen{listenCert, listenKey}
	var faelle []tlsFall
	for _, m := range [][]tlsFall{w.mengeE(n), w.mengeL(n), w.mengeZ(n), w.mengeS(n), w.mengeP(n), w.mengeG(n)} {
		faelle = append(faelle, m...)
	}
	if len(faelle) != 33 {
		t.Fatalf("%d Elemente, erwartet 33 (4+6+7+4+6+6)", len(faelle))
	}
	if len(n.sechsGruende()) != 7 {
		t.Fatalf("die Tabelle von Nr. 2 hat sechs Gründe, der Rang 1 trägt zwei Namen: %d Texte", len(n.sechsGruende()))
	}
	schl := string(w.schluessel(w.gut))
	koerper := strings.ReplaceAll(strings.TrimSuffix(strings.SplitN(schl, "\n", 2)[1], "-----END PRIVATE KEY-----\n"), "\n", "")
	verboten := []string{"geheim-pfad-4711", "geheim-cn-4711", filepath.Dir(w.dir), ".pem", koerper[:24], "tls:", "x509:", "/"}
	gesehen, soll := 0, 0
	for _, f := range faelle {
		_, vs := tlsLade(f)
		soll += len(f.erwartet)
		for _, got := range vs {
			gesehen++
			if !slices.Contains(n.sechsGruende(), got.Grund) || (got.Schluessel != n.cert && got.Schluessel != n.key) {
				t.Errorf("%q: Meldung %q ist keiner der sechs Texte", f.name, got)
			}
			for _, s := range verboten {
				if strings.Contains(got.Grund, s) || strings.Contains(got.Schluessel, s) {
					t.Errorf("%q: Meldung %q nennt %q", f.name, got, s)
				}
			}
		}
	}
	if gesehen != soll || soll != 31 {
		t.Errorf("%d Meldungen gesehen, %d erwartet (die 33 Elemente tragen 31 Zeilen)", gesehen, soll)
	}
	pruefeLiterale(t)
}

// pruefeLiterale zählt die Zeichenketten in tlsdatei.go: außer dem leeren Wert, den Teilen des Rangs 1, den Typnamen der Blöcke und
// `http/1.1` stehen dort genau die fünf festen Gründe, jeder einmal.
func pruefeLiterale(t *testing.T) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "tlsdatei.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	erlaubt := []string{"", "erforderlich, wenn ", " gesetzt ist", "CERTIFICATE", "PRIVATE KEY", " PRIVATE KEY", "http/1.1"}
	var uebrig []string
	ast.Inspect(f, func(node ast.Node) bool {
		if _, importiert := node.(*ast.ImportSpec); importiert {
			return false
		}
		if lit, ok := node.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			s, err := strconv.Unquote(lit.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Contains(erlaubt, s) {
				uebrig = append(uebrig, s)
			}
		}
		return true
	})
	fuenf := []string{gNichtLesbar, gKeinZert, gKeinSchl, gAusserhalb, gPasstNicht}
	slices.Sort(uebrig)
	slices.Sort(fuenf)
	if !slices.Equal(uebrig, fuenf) {
		t.Errorf("Literale in tlsdatei.go außer den erlaubten: %q, erwartet die fünf Gründe %q", uebrig, fuenf)
	}
}

// momentaufnahme hält Inhalt und Änderungszeit der Dateien und die Namen im Verzeichnis fest.
func (w *werkstatt) momentaufnahme(pfade ...string) string {
	var sb strings.Builder
	for _, p := range pfade {
		inhalt, err := os.ReadFile(p)
		info, err2 := os.Stat(p)
		if err != nil || err2 != nil {
			fmt.Fprintf(&sb, "%s: fehlt\n", filepath.Base(p))
			continue
		}
		fmt.Fprintf(&sb, "%s: %x %d %v\n", filepath.Base(p), inhalt, info.ModTime().UnixNano(), info.Mode())
	}
	eintraege, err := os.ReadDir(w.dir)
	if err != nil {
		w.t.Fatal(err)
	}
	for _, e := range eintraege {
		sb.WriteString(e.Name() + "\n")
	}
	return sb.String()
}

// TestLadeTLSZustandslos (A15): ladeTLS ändert keine Datei und legt keine an, und dieselben Dateien und Zeit geben dasselbe Ergebnis; die Fälle
// laufen nacheinander im selben Prozess, ein Erfolg zuletzt (Menge Y, 4 Elemente).
func TestLadeTLSZustandslos(t *testing.T) {
	w := neueWerkstatt(t)
	n := tlsNamen{listenCert, listenKey}
	gutCert, gutKey := w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut))
	altCert, altKey := w.datei(w.abgelaufen.ZertifikatPEM()), w.datei(w.schluessel(w.abgelaufen))
	text := w.datei([]byte("kein PEM\n"))
	faelle := []tlsFall{
		w.fall(n, "Rang 2", w.fehlt(), gutKey, tv(n.cert, gNichtLesbar)),
		w.fall(n, "Rang 3", text, gutKey, tv(n.cert, gKeinZert)),
		w.fall(n, "Rang 4", altCert, altKey, tv(n.cert, gAusserhalb)),
		w.fall(n, "Erfolg", gutCert, gutKey),
	}
	vor := time.Now().Add(-2 * time.Hour)
	for _, p := range []string{gutCert, gutKey, altCert, altKey, text} {
		if err := os.Chtimes(p, vor, vor); err != nil {
			t.Fatal(err)
		}
	}
	vorher := w.momentaufnahme(gutCert, gutKey, altCert, altKey, text)
	w.pruefe("Y", 4, faelle)
	for _, f := range faelle {
		c1, v1 := tlsLade(f)
		c2, v2 := tlsLade(f)
		if beschreibe(c1) != beschreibe(c2) || !slices.Equal(v1, v2) {
			t.Errorf("%s: zwei Aufrufe, zwei Ergebnisse", f.name)
		}
	}
	if nachher := w.momentaufnahme(gutCert, gutKey, altCert, altKey, text); nachher != vorher {
		t.Errorf("Dateien oder Verzeichnis geändert:\n%s\nvorher:\n%s", nachher, vorher)
	}
}

// TestLadeTLSGleichzeitig (A16): 16 Goroutinen laden dieselben Dateien, eine Hälfte ein gültiges Paar, die andere ein ungültiges; alle Ergebnisse
// einer Hälfte sind gleich. Fänger: ein Zwischenspeicher ohne Sperre meldet `make test-race` je Prozess als Datenrennen; ein Zwischenspeicher unter
// einer Sperre liefert einer Hälfte das Ergebnis der anderen und färbt die Gleichheit rot; ein nur gelesener, geteilter Zustand ist Gegenstand von
// `TestLadeTLSTeiltNichts`.
func TestLadeTLSGleichzeitig(t *testing.T) {
	w := neueWerkstatt(t)
	n := tlsNamen{listenCert, listenKey}
	gut := w.fall(n, "gültig", w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut)))
	schlecht := w.fall(n, "ungültig", w.datei(w.abgelaufen.ZertifikatPEM()), w.datei(w.schluessel(w.abgelaufen)), tv(n.cert, gAusserhalb))
	const zahl = 16
	ergebnisse := make([]string, zahl)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range zahl {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f := gut
			if i%2 == 1 {
				f = schlecht
			}
			<-start
			cfg, vs := tlsLade(f)
			ergebnisse[i] = fmt.Sprintf("%s|%v", beschreibe(cfg), vs)
		}()
	}
	close(start)
	wg.Wait()
	for i, e := range ergebnisse {
		if e != ergebnisse[i%2] {
			t.Errorf("Goroutine %d: %q, erwartet %q", i, e, ergebnisse[i%2])
		}
	}
	if !strings.HasSuffix(ergebnisse[1], "|[{LISTEN_TLS_CERT "+gAusserhalb+"}]") || strings.HasPrefix(ergebnisse[0], "nil") {
		t.Errorf("Ergebnisse %q und %q", ergebnisse[0], ergebnisse[1])
	}
}

// TestLadeTLSZweiteVerwendung (A17): der Baustein meldet unter den Namen des Aufrufers, nie unter einem festen (Menge N, 2 Namenspaare) für die
// Mengen E und L und das Paar „Gültigkeit und passt nicht“ aus A08 (2 × 11 Elemente).
func TestLadeTLSZweiteVerwendung(t *testing.T) {
	w := neueWerkstatt(t)
	zahl := 0
	paare := []tlsNamen{{listenCert, listenKey}, {adminCert, adminKey}}
	if len(paare) != 2 {
		t.Fatalf("Menge N hat %d Elemente, erwartet 2", len(paare))
	}
	for _, n := range paare {
		faelle := slices.Concat(w.mengeE(n), w.mengeL(n), []tlsFall{w.paar(n, zBesonders, zBesonders)})
		if want := []bootstrap.TLSVerstoss{tv(n.cert, gAusserhalb), tv(n.key, gPasstNicht)}; !slices.Equal(faelle[10].erwartet, want) {
			t.Fatalf("Paar Gültigkeit und passt nicht: Orakel %v, erwartet %v", faelle[10].erwartet, want)
		}
		w.pruefe("N", 11, faelle)
		zahl += len(faelle)
	}
	if zahl != 22 {
		t.Fatalf("%d Elemente, erwartet 2 × 11", zahl)
	}
	// Der Rang 1 trägt den Namen des anderen Schlüssels, den der Aufrufer übergibt, und nur diesen.
	_, vs := bootstrap.LadeTLS("A_CERT", "B_KEY", "", w.datei(w.schluessel(w.gut)), w.jetzt)
	if want := []bootstrap.TLSVerstoss{tv("A_CERT", gErforderlich("B_KEY"))}; !reflect.DeepEqual(vs, want) {
		t.Errorf("fremde Namen: %v, erwartet %v", vs, want)
	}
}

// TestLadeTLSSortiert: die Meldung ist nach Schlüssel (Bytefolge) sortiert, auch wenn der Aufrufer den Schlüssel vor das Zertifikat sortiert
// (SPEC-110 Nr. 12, `SPEC-291` Nr. 2).
func TestLadeTLSSortiert(t *testing.T) {
	w := neueWerkstatt(t)
	_, vs := bootstrap.LadeTLS("Z_CERT", "A_KEY", w.fehlt(), w.fehlt(), w.jetzt)
	if want := []bootstrap.TLSVerstoss{tv("A_KEY", gNichtLesbar), tv("Z_CERT", gNichtLesbar)}; !slices.Equal(vs, want) {
		t.Errorf("Meldung %v, erwartet %v", vs, want)
	}
}

// TestLadeTLSSortiertJeNamenpaar: die Reihenfolge ist die Bytefolge der Namen für jedes Paar verschiedener Namen aus einer Menge, die Groß- und
// Kleinschreibung, Ziffer, Unterstrich und Präfix mischt (Menge O, 8 Namen, 8 × 7 Paare); das Orakel ist die Ordnung von `slices.Sort`.
func TestLadeTLSSortiertJeNamenpaar(t *testing.T) {
	w := neueWerkstatt(t)
	namen := []string{"0", "A", "AB", "Z", "_", "a", "ab", "z"}
	zahl := 0
	for _, c := range namen {
		for _, k := range namen {
			if c == k {
				continue
			}
			zahl++
			soll := []string{c, k}
			slices.Sort(soll)
			_, vs := bootstrap.LadeTLS(c, k, w.fehlt(), w.fehlt(), w.jetzt)
			if want := []bootstrap.TLSVerstoss{tv(soll[0], gNichtLesbar), tv(soll[1], gNichtLesbar)}; !slices.Equal(vs, want) {
				t.Errorf("Namen %q und %q: Meldung %v, erwartet %v", c, k, vs, want)
			}
		}
	}
	if zahl != len(namen)*(len(namen)-1) {
		t.Fatalf("Menge O hat %d Paare, erwartet %d", zahl, len(namen)*(len(namen)-1))
	}
}

// TestLadeTLSBloeckeDavor: Blöcke anderen Typs überspringt das Lesen an jeder Stelle (SPEC-291 Nr. 4): der gesuchte Block steht hinter 0 bis 10
// fremden Blöcken, in der Zertifikatsdatei, in der Schlüsseldatei und in einer Datei für beide (Schlüssel vor Zertifikat, Zertifikat vor
// Schlüssel, Kette vor Schlüssel), je Art des fremden Blocks (ein Schlüssel, ein Zertifikat, `EC PARAMETERS`, `X509 CRL`, ein unbekannter Typ,
// `PUBLIC KEY`) und gemischt. Menge D: das Produkt der Achsen Dateiform × Art × Zahl, die der Test zählt; das Orakel ist „gültig“.
func TestLadeTLSBloeckeDavor(t *testing.T) {
	w := neueWerkstatt(t)
	n := tlsNamen{listenCert, listenKey}
	zw := w.wurzel.Zwischen(t, tlstest.Optionen{})
	b := zw.Blatt(t, tlstest.Optionen{})
	blatt, zwischen, schl := b.ZertifikatPEM(), tlstest.Block("CERTIFICATE", b.Kette[0]), b.SchluesselPEM(t, tlstest.SEC1)
	fremd := map[string][]byte{
		"PRIVATE KEY":   w.fremd.SchluesselPEM(t, tlstest.PKCS8),
		"CERTIFICATE":   w.fremd.ZertifikatPEM(),
		"EC PARAMETERS": tlstest.Block("EC PARAMETERS", []byte{0x06, 0x08, 0x2A, 0x86, 0x48, 0xCE, 0x3D, 0x03, 0x01, 0x07}),
		"X509 CRL":      tlstest.Block("X509 CRL", []byte{1, 2, 3}),
		"FOO BAR":       tlstest.Block("FOO BAR", []byte("unbekannt")),
		"PUBLIC KEY":    tlstest.Block("PUBLIC KEY", []byte("oeffentlich")),
	}
	// Eine Dateiform bildet aus dem fremden Vorspann die Inhalte der Zertifikats- und der Schlüsseldatei; gleiche Inhalte heißen: eine Datei.
	formen := []struct {
		name  string
		arten []string
		bau   func(vor []byte) (cert, key []byte)
	}{
		{"Zertifikatsdatei", []string{"PRIVATE KEY", "EC PARAMETERS", "X509 CRL", "FOO BAR"}, func(v []byte) ([]byte, []byte) { return slices.Concat(v, blatt), schl }},
		{"Schlüsseldatei", []string{"CERTIFICATE", "EC PARAMETERS", "X509 CRL", "FOO BAR"}, func(v []byte) ([]byte, []byte) { return blatt, slices.Concat(v, schl) }},
		{"Schlüssel vor Zertifikat", []string{"EC PARAMETERS", "X509 CRL", "FOO BAR", "PUBLIC KEY"}, func(v []byte) ([]byte, []byte) {
			d := slices.Concat(v, schl, blatt)
			return d, d
		}},
		{"Zertifikat vor Schlüssel", []string{"EC PARAMETERS", "X509 CRL", "FOO BAR", "PUBLIC KEY"}, func(v []byte) ([]byte, []byte) {
			d := slices.Concat(v, blatt, schl)
			return d, d
		}},
		{"Kette vor Schlüssel", []string{"EC PARAMETERS", "X509 CRL", "FOO BAR", "PUBLIC KEY"}, func(v []byte) ([]byte, []byte) {
			d := slices.Concat(v, blatt, zwischen, schl)
			return d, d
		}},
	}
	const hoechstens = 10
	var faelle []tlsFall
	soll := 0
	for _, f := range formen {
		soll += (len(f.arten) + 1) * (hoechstens + 1)
		for _, art := range append(slices.Clone(f.arten), "gemischt") {
			for zahl := 0; zahl <= hoechstens; zahl++ {
				var vor []byte
				for i := range zahl {
					a := art
					if art == "gemischt" {
						a = f.arten[i%len(f.arten)]
					}
					vor = append(vor, fremd[a]...)
				}
				cert, key := f.bau(vor)
				certPfad := w.datei(cert)
				keyPfad := certPfad
				if !bytes.Equal(cert, key) {
					keyPfad = w.datei(key)
				}
				faelle = append(faelle, w.fall(n, fmt.Sprintf("%s, %s, %d davor", f.name, art, zahl), certPfad, keyPfad))
			}
		}
	}
	w.pruefe("D", soll, faelle)
}

// TestLadeTLSRechte: ladeTLS prüft die Rechte nie und ändert weder Rechte noch Inhalt noch Änderungszeit einer Datei (SPEC-291 Nr. 10, Nr. 12):
// jedes Rechtemuster der Menge (14 Muster mit Leserecht des Eigentümers) für die Zertifikats- und für die Schlüsseldatei, bei einem gültigen Paar
// und bei einem Zertifikat außerhalb der Gültigkeit (Menge M, 14 × 2 × 2 Elemente); das Ergebnis ist vom Muster unabhängig.
func TestLadeTLSRechte(t *testing.T) {
	w := neueWerkstatt(t)
	n := tlsNamen{listenCert, listenKey}
	modi := []os.FileMode{0o400, 0o440, 0o444, 0o600, 0o640, 0o644, 0o660, 0o664, 0o666, 0o700, 0o750, 0o755, 0o770, 0o777}
	alt := time.Now().Add(-2 * time.Hour)
	zahl := 0
	for _, abgelaufen := range []bool{false, true} {
		blatt := w.gut
		var erwartet []bootstrap.TLSVerstoss
		if abgelaufen {
			blatt, erwartet = w.abgelaufen, []bootstrap.TLSVerstoss{tv(n.cert, gAusserhalb)}
		}
		for _, datei := range []string{"Zertifikat", "Schlüssel"} {
			for _, modus := range modi {
				zahl++
				mc, mk := os.FileMode(0o600), os.FileMode(0o600)
				if datei == "Zertifikat" {
					mc = modus
				} else {
					mk = modus
				}
				f := w.fall(n, fmt.Sprintf("%s %v abgelaufen=%v", datei, modus, abgelaufen),
					w.dateiModus(blatt.ZertifikatPEM(), mc), w.dateiModus(w.schluessel(blatt), mk), erwartet...)
				for _, p := range []string{f.cert, f.key} {
					if err := os.Chtimes(p, alt, alt); err != nil {
						t.Fatal(err)
					}
				}
				vorher := w.momentaufnahme(f.cert, f.key)
				w.pruefe("M", 1, []tlsFall{f})
				if nachher := w.momentaufnahme(f.cert, f.key); nachher != vorher {
					t.Errorf("%s: Dateien geändert:\n%s\nvorher:\n%s", f.name, nachher, vorher)
				}
			}
		}
	}
	if zahl != len(modi)*2*2 {
		t.Fatalf("Menge M hat %d Elemente, erwartet %d", zahl, len(modi)*2*2)
	}
}

// TestLadeTLSFelder: das gelieferte `*tls.Config` setzt allein `Certificates`, `MinVersion` und `NextProtos`; jedes andere Feld des Typs, auch ein
// künftiges, ist der Nullwert (SPEC-291 Nr. 5, Nr. 11). Menge C: jedes Feld des Typs `tls.Config`, per Reflexion gezählt.
func TestLadeTLSFelder(t *testing.T) {
	w := neueWerkstatt(t)
	cfg := w.ladeGueltig(w.gut.ZertifikatPEM(), w.schluessel(w.gut))
	typ := reflect.TypeFor[tls.Config]()
	werte := reflect.ValueOf(cfg).Elem()
	zugesagt := map[string]bool{"Certificates": true, "MinVersion": true, "NextProtos": true}
	gesehen := 0
	for i := range typ.NumField() {
		gesehen++
		name := typ.Field(i).Name
		if zugesagt[name] == werte.Field(i).IsZero() {
			t.Errorf("Feld %s: Nullwert %v, zugesagt %v", name, werte.Field(i).IsZero(), zugesagt[name])
		}
	}
	for name := range zugesagt {
		if _, da := typ.FieldByName(name); !da {
			t.Errorf("Feld %s gibt es nicht im Typ", name)
		}
	}
	if gesehen != typ.NumField() || gesehen < len(zugesagt) {
		t.Fatalf("Menge C hat %d Felder, der Typ %d", gesehen, typ.NumField())
	}
}

// TestLadeTLSTeiltNichts: zwei Aufrufe teilen keine Liste und keinen Wert (SPEC-291 Nr. 12): ändert ein Aufrufer das gelieferte Config (Eintrag
// von `NextProtos`, Paar, Bytes des Zertifikats), bleiben das Ergebnis eines früheren und eines späteren Aufrufs gleich dem ersten.
func TestLadeTLSTeiltNichts(t *testing.T) {
	w := neueWerkstatt(t)
	n := tlsNamen{listenCert, listenKey}
	f := w.fall(n, "gültig", w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut)))
	frueher, _ := tlsLade(f)
	soll := beschreibe(frueher)
	geaendert, _ := tlsLade(f)
	geaendert.NextProtos[0] = "veraendert"
	geaendert.Certificates[0].Certificate[0][0] ^= 0xff
	geaendert.Certificates[0] = tls.Certificate{}
	spaeter, _ := tlsLade(f)
	for name, cfg := range map[string]*tls.Config{"früherer Aufruf": frueher, "späterer Aufruf": spaeter} {
		if got := beschreibe(cfg); got != soll {
			t.Errorf("%s: %s, erwartet %s", name, got, soll)
		}
	}
}
