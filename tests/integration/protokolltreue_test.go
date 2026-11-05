package integration_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/bootstrap"
)

// Der Vergleichstest der Protokolltreue (LH-QA-07, LH-QA-07.a): derselbe Request direkt an den Referenzdienst
// und über den Recorder im Modus record; verglichen wird die entpackte Darstellung der Antworten. Der Vergleich
// selbst trägt Mutationen: ein Veränderer zwischen Client und Recorder bricht je eine Zusage, und der Vergleich
// muss genau diese Abweichung nennen.

// fall ist ein Request an den Referenzdienst samt Erwartung an die direkte Antwort: eine Antwortart des Dienstes.
type fall struct {
	name, methode, pfad, accept string
	// ae sind die Zeilen von Accept-Encoding; nil heißt: der Header fehlt.
	ae []string
	// roh nennt den Fall mit einem Coding, das der Client nicht nannte (LH-QA-07.a Nr. 4).
	roh    bool
	status int
	typ    string
	// kodierung ist der Wert von Content-Encoding der direkten Antwort, "" ohne den Header.
	kodierung string
	// inhalt ist der Body nach dem Entpacken, "" wenn er nicht festgelegt ist; enthaelt ein Teilstring davon.
	inhalt, enthaelt string
	// leer nennt eine Antwort ohne Body; relevant sind die Namen der relevanten Header, sortiert und klein.
	leer     bool
	relevant []string
}

func relevantMit(extra ...string) []string {
	return slices.Sorted(slices.Values(slices.Concat([]string{"odata-version", "x-dienst", "x-echo-accept-encoding"}, extra)))
}

// faelle ist die Menge der Antwortarten des Referenzdienstes: $metadata, Collection, Einzel-Entity, $count und
// $value als text/plain, Fehlerstatus, Antwort mit Content-Encoding, Antwort ohne Body, Weiterleitung, Header
// mit mehreren Zeilen und einem leeren Wert, ein Coding, das der Client nicht nannte.
func faelle() []fall {
	odataJSON := "application/json;odata.metadata=minimal"
	ct := "content-type"
	return []fall{
		{name: "metadata", pfad: "/odata/$metadata", accept: acceptXML, status: 200, typ: "application/xml", inhalt: csdlText, relevant: relevantMit(ct)},
		{name: "sammlung", pfad: "/odata/Products", accept: acceptJSON, status: 200, typ: odataJSON, enthaelt: `"Name":"Bolt"`, relevant: relevantMit(ct)},
		{name: "entity", pfad: "/odata/Products(1)", accept: acceptJSON, status: 200, typ: odataJSON, enthaelt: `"Id":1`, relevant: relevantMit(ct, "etag")},
		{name: "zaehler", pfad: "/odata/Products/$count", accept: acceptText, status: 200, typ: "text/plain", inhalt: "7", relevant: relevantMit(ct)},
		{name: "wert", pfad: "/odata/Products(1)/Name/$value", accept: acceptText, status: 200, typ: "text/plain", inhalt: "Anvil", relevant: relevantMit(ct)},
		{name: "fehler", pfad: "/odata/Products(99)", accept: acceptJSON, status: 404, typ: "application/json", enthaelt: `"code":"404"`, relevant: relevantMit(ct)},
		{name: "gzip", pfad: "/odata/Products", accept: acceptJSON, ae: []string{"gzip"}, status: 200, typ: odataJSON, kodierung: "gzip", enthaelt: `"Name":"Bolt"`, relevant: relevantMit(ct)},
		{name: "deflate", pfad: "/odata/Products", accept: acceptJSON, ae: []string{"deflate"}, status: 200, typ: odataJSON, kodierung: "deflate", enthaelt: `"Name":"Bolt"`, relevant: relevantMit(ct)},
		{name: "kette", pfad: "/odata/Products", accept: acceptJSON, ae: []string{"gzip, deflate"}, status: 200, typ: odataJSON, kodierung: "deflate, gzip", enthaelt: `"Name":"Bolt"`, relevant: relevantMit(ct)},
		{name: "identity", pfad: "/odata/Products", accept: acceptJSON, ae: []string{"identity"}, status: 200, typ: odataJSON, enthaelt: `"Name":"Bolt"`, relevant: relevantMit(ct)},
		{name: "head", methode: http.MethodHead, pfad: "/odata/Products(1)", accept: acceptJSON, status: 200, typ: odataJSON, leer: true, relevant: relevantMit(ct, "etag")},
		{name: "ohne-inhalt", pfad: "/odata/Leer", accept: acceptJSON, status: 204, leer: true, relevant: relevantMit()},
		{name: "unveraendert", pfad: "/odata/Unveraendert", accept: acceptJSON, status: 304, leer: true, relevant: relevantMit("etag")},
		{name: "weiterleitung", pfad: "/odata/Weiter", accept: acceptJSON, status: 302, leer: true, relevant: relevantMit("location")},
		{name: "header", pfad: "/odata/Header", accept: acceptJSON, status: 200, typ: odataJSON, enthaelt: `"value"`, relevant: relevantMit(ct, "x-mehrfach", "x-leer")},
		{name: "unbekanntes-coding", pfad: "/odata/Brotli", accept: acceptJSON, ae: []string{"gzip"}, roh: true, status: 200, typ: odataJSON, kodierung: "br", relevant: relevantMit(ct, "content-encoding")},
	}
}

// vclient liest Antworten roh: kein Entpacken des Transports, keine Weiterleitung (LH-QA-07.a Nr. 5 und 6).
type vclient struct {
	t     *testing.T
	c     *http.Client
	basis string
}

func neuerVClient(t *testing.T, basis string) *vclient {
	t.Helper()
	tr := &http.Transport{DisableKeepAlives: true, DisableCompression: true}
	t.Cleanup(tr.CloseIdleConnections)
	return &vclient{t: t, basis: basis, c: &http.Client{
		Timeout: 20 * time.Second, Transport: tr,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

// holeFehler sendet den Fall; Accept-Encoding geht nur mit, wenn der Fall ihn nennt.
func (c *vclient) holeFehler(f fall) (rohantwort, error) {
	m := f.methode
	if m == "" {
		m = http.MethodGet
	}
	req, err := http.NewRequest(m, c.basis+f.pfad, nil)
	if err != nil {
		return rohantwort{}, err
	}
	req.Header.Set("Accept", f.accept)
	if f.ae != nil {
		req.Header["Accept-Encoding"] = f.ae
	}
	resp, err := c.c.Do(req)
	if err != nil {
		return rohantwort{}, fmt.Errorf("%s: %w", f.name, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return rohantwort{}, fmt.Errorf("%s: Body: %w", f.name, err)
	}
	return rohantwort{status: resp.StatusCode, phrase: strings.TrimPrefix(resp.Status, strconv.Itoa(resp.StatusCode)+" "),
		header: resp.Header, body: body}, nil
}

func (c *vclient) hole(f fall) rohantwort {
	c.t.Helper()
	a, err := c.holeFehler(f)
	if err != nil {
		c.t.Fatal(err)
	}
	return a
}

// veraenderung ändert eine Antwort des Recorders vor dem Client: Byte, Header oder Status.
type veraenderung func(a *rohantwort) error

// veraenderer ist ein Handler zwischen Client und Recorder, der die Antwort des Recorders verändert, ohne den
// Recorder zu ändern; der Recorder bleibt der echte.
type veraenderer struct {
	srv  *httptest.Server
	c    *http.Client
	ziel string
	v    veraenderung
	mu   sync.Mutex
	err  error
}

func neuerVeraenderer(ziel string, v veraenderung) *veraenderer {
	m := &veraenderer{ziel: ziel, v: v}
	tr := &http.Transport{DisableKeepAlives: true, DisableCompression: true}
	m.c = &http.Client{Timeout: 20 * time.Second, Transport: tr,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	m.srv = httptest.NewServer(m)
	return m
}

func (m *veraenderer) melde(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err == nil {
		m.err = err
	}
}

func istHopByHop(name string) bool {
	return slices.Contains(hopByHopNamen(), http.CanonicalHeaderKey(name)) || asciiKlein(name) == "te"
}

func (m *veraenderer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	out, err := http.NewRequest(req.Method, m.ziel+req.URL.RequestURI(), nil)
	if err != nil {
		m.melde(err)
		return
	}
	for n, v := range req.Header {
		if !istHopByHop(n) {
			out.Header[n] = v
		}
	}
	resp, err := m.c.Do(out)
	if err != nil {
		m.melde(err)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		m.melde(err)
		return
	}
	a := rohantwort{status: resp.StatusCode, header: resp.Header, body: body}
	if err := m.v(&a); err != nil {
		m.melde(err)
		return
	}
	for n, v := range a.header {
		if !istHopByHop(n) && asciiKlein(n) != "content-length" {
			w.Header()[n] = v
		}
	}
	w.WriteHeader(a.status)
	_, _ = w.Write(a.body)
}

// vgumgebung ist der Aufbau eines Tests: der Referenzdienst, der Recorder im Modus record und je ein Client.
type vgumgebung struct {
	t             *testing.T
	dienst        *vergleichsdienst
	rec           *bootstrap.Handle
	k             bootstrap.Konfiguration
	faelle        []fall
	direkt, proxy *vclient
}

func neueVgUmgebung(t *testing.T) *vgumgebung {
	t.Helper()
	d := neuerVergleichsdienst(t)
	k := konfig(t, "record", t.TempDir(), d.srv.URL+"/odata")
	rec := hochgefahren(context.Background(), t, k)
	return &vgumgebung{t: t, dienst: d, rec: rec, k: k, faelle: faelle(),
		direkt: neuerVClient(t, d.srv.URL), proxy: neuerVClient(t, "http://"+rec.Adresse())}
}

// fall liefert den Fall mit dem Namen.
func (u *vgumgebung) fall(name string) fall {
	u.t.Helper()
	for _, f := range u.faelle {
		if f.name == name {
			return f
		}
	}
	u.t.Fatalf("kein Fall %q", name)
	return fall{}
}

func (u *vgumgebung) regeln(f fall) vergleichsregeln {
	return vergleichsregeln{volatile: u.k.VolatileHeaders, roh: f.roh}
}

// viaVeraenderer holt den Fall über einen Veraenderer vor dem Recorder.
func (u *vgumgebung) viaVeraenderer(f fall, v veraenderung) rohantwort {
	u.t.Helper()
	m := neuerVeraenderer("http://"+u.rec.Adresse(), v)
	defer m.srv.Close()
	a := neuerVClient(u.t, m.srv.URL).hole(f)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		u.t.Fatalf("%s: Veränderung: %v", f.name, m.err)
	}
	return a
}

// abweichungen nennt die Unterschiede der direkten Antwort zu der gegebenen; ein Fehler des Tests bricht ab.
func (u *vgumgebung) abweichungen(f fall, proxy rohantwort) []abweichung {
	u.t.Helper()
	ab, err := vergleiche(u.regeln(f), u.direkt.hole(f), proxy)
	if err != nil {
		u.t.Fatalf("%s: %v", f.name, err)
	}
	return ab
}

// mitVeraenderung sagt, ob der Vergleich genau die erwartete Abweichung nennt, wenn der Veraenderer sie einbaut.
func (u *vgumgebung) mitVeraenderung(f fall, v veraenderung) []abweichung {
	u.t.Helper()
	return u.abweichungen(f, u.viaVeraenderer(f, v))
}

func (u *vgumgebung) erwarte(f fall, was string, v veraenderung, soll ...abweichung) {
	u.t.Helper()
	if got := u.mitVeraenderung(f, v); !slices.Equal(got, soll) {
		u.t.Errorf("%s, %s: Abweichungen %v, erwartet %v", f.name, was, got, soll)
	}
}

// --- Bausteine der Veränderung

func headerAendern(klein string, f func([]string) []string) veraenderung {
	return func(a *rohantwort) error {
		for n, v := range a.header {
			if asciiKlein(n) == klein {
				a.header[n] = f(slices.Clone(v))
				return nil
			}
		}
		return fmt.Errorf("Header %s fehlt in der Antwort", klein)
	}
}

func headerEntfernen(klein string) veraenderung {
	return func(a *rohantwort) error {
		for n := range a.header {
			if asciiKlein(n) == klein {
				delete(a.header, n)
				return nil
			}
		}
		return fmt.Errorf("Header %s fehlt in der Antwort", klein)
	}
}

func headerHinzufuegen(name, wert string) veraenderung {
	return func(a *rohantwort) error {
		a.header[name] = []string{wert}
		return nil
	}
}

func headerSetzen(klein string, werte ...string) veraenderung {
	return headerAendern(klein, func([]string) []string { return werte })
}

func headerWertAendern(klein string) veraenderung {
	return headerAendern(klein, func(v []string) []string {
		for i := range v {
			v[i] += "~"
		}
		return v
	})
}

func andererStatus(s int) int {
	switch s {
	case 200:
		return 201
	case 404:
		return 500
	case 302:
		return 303
	}
	return 200
}

func statusWechsel(a *rohantwort) error {
	a.status = andererStatus(a.status)
	return nil
}

// stelle ist die Position im Inhalt: das erste, ein mittleres oder das letzte Byte.
func stelle(pos string, n int) int {
	switch pos {
	case "erstes":
		return 0
	case "mitte":
		return n / 2
	}
	return n - 1
}

func stellen() []string { return []string{"erstes", "mitte", "letztes"} }

// inhaltByte ändert ein Byte des entpackten Inhalts und kodiert mit derselben Kette neu.
func inhaltByte(pos string) veraenderung {
	return umkodieren(nil, pos)
}

// setzeCoding ersetzt Body und Content-Encoding durch den Inhalt in der Kette neu; eine leere Kette ist ohne Coding.
func setzeCoding(a *rohantwort, inhalt []byte, neu []string) {
	a.body = kodiere(neu, inhalt)
	for n := range a.header {
		if asciiKlein(n) == "content-encoding" {
			delete(a.header, n)
		}
	}
	if len(neu) > 0 {
		a.header["Content-Encoding"] = []string{strings.Join(neu, ", ")}
	}
}

// umkodieren entpackt die Antwort, ändert bei pos ein Byte und kodiert mit neu; ein nil neu lässt die Kette der
// Antwort, ein leeres neu ist ohne Coding.
func umkodieren(neu []string, pos string) veraenderung {
	return func(a *rohantwort) error {
		alt, err := kette(a.header)
		if err != nil {
			return err
		}
		inhalt, err := entpacke(a.header, a.body)
		if err != nil {
			return err
		}
		if len(inhalt) == 0 {
			return errors.New("kein Inhalt zu verändern")
		}
		inhalt[stelle(pos, len(inhalt))] ^= 0x01
		if neu == nil {
			neu = alt
		}
		setzeCoding(a, inhalt, neu)
		return nil
	}
}

// umkodierenOhneAenderung wechselt das Coding und lässt den Inhalt; eine leere Kette ist ohne Coding.
func umkodierenOhneAenderung(neu []string) veraenderung {
	return func(a *rohantwort) error {
		inhalt, err := entpacke(a.header, a.body)
		if err != nil {
			return err
		}
		setzeCoding(a, inhalt, neu)
		return nil
	}
}

// rohByte ändert ein Byte der Rohbytes des Bodys.
func rohByte(pos string) veraenderung {
	return func(a *rohantwort) error {
		if len(a.body) == 0 {
			return errors.New("kein Body zu verändern")
		}
		a.body[stelle(pos, len(a.body))] ^= 0x01
		return nil
	}
}

func bodySetzen(neu string) veraenderung {
	return func(a *rohantwort) error {
		a.body = []byte(neu)
		return nil
	}
}

// mitInhalt: Fälle mit einem Body nach dem Entpacken, den der Vergleich vergleicht.
func mitInhalt(fs []fall) []fall {
	var out []fall
	for _, f := range fs {
		if !f.leer && !f.roh {
			out = append(out, f)
		}
	}
	return out
}

// --- Prüfung der direkten Antworten

// TestFaelleDirektWieErwartet fängt den Referenzdienst gegen die Annahmen der Fälle: Status, Typ, Coding, Inhalt,
// die Namen der relevanten Header und das Echo des Accept-Encoding je Antwortart; ohne sie wäre ein grüner
// Vergleich kein Beleg für die Antwortart (Z16, Z18, LH-QA-07.a Nr. 4 bis 6).
func TestFaelleDirektWieErwartet(t *testing.T) {
	u := neueVgUmgebung(t)
	if len(u.faelle) != 16 {
		t.Fatalf("%d Fälle, erwartet 16", len(u.faelle))
	}
	for _, f := range u.faelle {
		for seite, c := range map[string]*vclient{"direkt": u.direkt, "über den Proxy": u.proxy} {
			a := c.hole(f)
			soll := "fehlt"
			if f.ae != nil {
				soll = strings.Join(f.ae, "|")
			}
			if a.status != f.status || strings.Join(a.header["Content-Type"], "|") != f.typ ||
				strings.Join(a.header["Content-Encoding"], ", ") != f.kodierung || strings.Join(a.header["X-Echo-Accept-Encoding"], "|") != soll {
				t.Errorf("%s, %s: Status %d, Typ %q, Coding %q, Echo %q; erwartet %d, %q, %q, %q", f.name, seite, a.status,
					a.header["Content-Type"], a.header["Content-Encoding"], a.header["X-Echo-Accept-Encoding"], f.status, f.typ, f.kodierung, soll)
			}
			if got := u.regeln(f).relevanteNamen(a); !slices.Equal(got, f.relevant) {
				t.Errorf("%s, %s: relevante Header %v, erwartet %v", f.name, seite, got, f.relevant)
			}
			pruefeInhalt(t, f, seite, a)
		}
	}
}

func pruefeInhalt(t *testing.T, f fall, seite string, a rohantwort) {
	t.Helper()
	switch {
	case f.leer:
		if len(a.body) != 0 {
			t.Errorf("%s, %s: Body %q, erwartet keiner", f.name, seite, a.body)
		}
		return
	case f.roh:
		if !bytes.Equal(a.body, brotliBytes()) {
			t.Errorf("%s, %s: Rohbytes %x, erwartet %x", f.name, seite, a.body, brotliBytes())
		}
		return
	case f.kodierung != "" && !bytes.HasPrefix(a.body, []byte{0x1f, 0x8b}) && strings.HasSuffix(f.kodierung, "gzip"):
		t.Errorf("%s, %s: Body beginnt nicht mit den Bytes 1f 8b: %x", f.name, seite, a.body)
	}
	inhalt, err := entpacke(a.header, a.body)
	if err != nil {
		t.Errorf("%s, %s: %v", f.name, seite, err)
	}
	if (f.inhalt != "" && string(inhalt) != f.inhalt) || !bytes.Contains(inhalt, []byte(f.enthaelt)) || len(inhalt) == 0 {
		t.Errorf("%s, %s: Inhalt %q, erwartet %q beziehungsweise mit %q", f.name, seite, inhalt, f.inhalt, f.enthaelt)
	}
}

// --- Der Vergleich über den Recorder

// TestProtokolltreue ist der Vergleichstest: je Antwortart derselbe Request direkt und über den Recorder, ohne
// Abweichung (LH-QA-07, LH-FA-01); der Fall mit einem Coding, das der Client nicht nannte, vergleicht Rohbytes.
func TestProtokolltreue(t *testing.T) {
	u := neueVgUmgebung(t)
	for _, f := range u.faelle {
		if ab := u.abweichungen(f, u.proxy.hole(f)); len(ab) != 0 {
			t.Errorf("%s: Abweichungen %v", f.name, ab)
		}
	}
}

// TestSelbstvergleich: der Referenzdienst antwortet auf denselben Request gleich, direkt gegen direkt (Z17,
// LH-QA-07.a Nr. 6); der Client folgt keiner Weiterleitung.
func TestSelbstvergleich(t *testing.T) {
	u := neueVgUmgebung(t)
	for _, f := range u.faelle {
		if ab := u.abweichungen(f, u.direkt.hole(f)); len(ab) != 0 {
			t.Errorf("%s: zwei direkte Antworten weichen ab: %v", f.name, ab)
		}
	}
	for seite, c := range map[string]*vclient{"direkt": u.direkt, "über den Proxy": u.proxy} {
		w := c.hole(fall{name: "weiterleitung", pfad: "/odata/Weiter", accept: acceptJSON})
		if w.status != 302 || strings.Join(w.header["Location"], "") != u.dienst.srv.URL+"/odata/Products" || len(w.body) != 0 {
			t.Errorf("%s: Weiterleitung als %d %q mit Body %q", seite, w.status, w.header["Location"], w.body)
		}
	}
}

// TestStatusAbweichung (Z1): ein anderer Status bei jeder Antwortart ist die Abweichung Status, und nur sie.
func TestStatusAbweichung(t *testing.T) {
	u := neueVgUmgebung(t)
	for _, f := range u.faelle {
		u.erwarte(f, "Status", statusWechsel, abweichung{"Status", ""})
	}
}

// TestStatusNichtPhrase (Z1): die Reason-Phrase zählt nicht.
func TestStatusNichtPhrase(t *testing.T) {
	a := rohantwort{status: 200, phrase: "OK"}
	for _, b := range []rohantwort{{status: 200, phrase: "Fine"}, {status: 200, phrase: ""}} {
		if ab, err := vergleiche(vergleichsregeln{}, a, b); err != nil || len(ab) != 0 {
			t.Errorf("Phrase %q gegen %q: %v, %v", a.phrase, b.phrase, ab, err)
		}
	}
	if ab, err := vergleiche(vergleichsregeln{}, a, rohantwort{status: 201, phrase: "OK"}); err != nil || !slices.Equal(ab, []abweichung{{"Status", ""}}) {
		t.Errorf("anderer Status bei gleicher Phrase: %v, %v", ab, err)
	}
}

// TestBodyAbweichung (Z2): ein geändertes erstes, mittleres und letztes Byte des entpackten Bodys ist die
// Abweichung Body, bei jeder Antwortart mit Body; die Länge bleibt gleich.
func TestBodyAbweichung(t *testing.T) {
	u := neueVgUmgebung(t)
	n := 0
	for _, f := range mitInhalt(u.faelle) {
		for _, pos := range stellen() {
			u.erwarte(f, "Byte "+pos, inhaltByte(pos), abweichung{"Body", ""})
			n++
		}
	}
	if n != 33 {
		t.Errorf("%d Fälle des Bodys, erwartet 33 (11 Antwortarten mit Body, drei Stellen)", n)
	}
}

// TestHeaderWertAbweichung (Z3): der Wert jedes relevanten Headers jedes Falls, einzeln geändert, ist die
// Abweichung dieses Headers.
func TestHeaderWertAbweichung(t *testing.T) {
	u := neueVgUmgebung(t)
	n := 0
	for _, f := range u.faelle {
		for _, name := range f.relevant {
			u.erwarte(f, "Wert von "+name, headerWertAendern(name), abweichung{"Header", name})
			n++
		}
	}
	if n != 68 {
		t.Errorf("%d relevante Header über die Fälle, erwartet 68", n)
	}
}

// TestHeaderNurEineSeite (Z4): ein relevanter Name nur auf einer Seite ist eine Abweichung, in beiden Richtungen:
// jeder relevante Name jedes Falls fehlt über den Proxy, und ein Name kommt über den Proxy hinzu.
func TestHeaderNurEineSeite(t *testing.T) {
	u := neueVgUmgebung(t)
	for _, f := range u.faelle {
		for _, name := range f.relevant {
			u.erwarte(f, "ohne "+name, headerEntfernen(name), abweichung{"Header", name})
		}
		u.erwarte(f, "mit x-zusatz", headerHinzufuegen("X-Zusatz", "1"), abweichung{"Header", "x-zusatz"})
	}
}

// TestLeererWert (Z5): ein leerer Wert ist ein Wert; leer gleich leer, leer ungleich fehlend, leer ungleich x.
func TestLeererWert(t *testing.T) {
	u := neueVgUmgebung(t)
	f := u.fall("header")
	if d := u.direkt.hole(f); !slices.Equal(d.header["X-Leer"], []string{""}) {
		t.Fatalf("der Dienst sendet X-Leer als %q, erwartet eine leere Zeile", d.header["X-Leer"])
	}
	if ab := u.abweichungen(f, u.proxy.hole(f)); len(ab) != 0 {
		t.Errorf("leer gegen leer: %v", ab)
	}
	u.erwarte(f, "leer wird fehlend", headerEntfernen("x-leer"), abweichung{"Header", "x-leer"})
	u.erwarte(f, "leer wird x", headerSetzen("x-leer", "x"), abweichung{"Header", "x-leer"})
}

// TestFeldzeilen (Z6): zwei Zeilen desselben Namens zählen in ihrer Reihenfolge, und zwei Zeilen sind nicht eine
// Zeile mit Komma.
func TestFeldzeilen(t *testing.T) {
	u := neueVgUmgebung(t)
	f := u.fall("header")
	if d := u.direkt.hole(f); !slices.Equal(d.header["X-Mehrfach"], []string{"a", "b"}) {
		t.Fatalf("der Dienst sendet X-Mehrfach als %q, erwartet zwei Zeilen a und b", d.header["X-Mehrfach"])
	}
	vertauscht := func(v []string) []string { slices.Reverse(v); return v }
	zusammen := func(v []string) []string { return []string{strings.Join(v, ", ")} }
	u.erwarte(f, "Zeilen vertauscht", headerAendern("x-mehrfach", vertauscht), abweichung{"Header", "x-mehrfach"})
	u.erwarte(f, "Zeilen zu einer", headerAendern("x-mehrfach", zusammen), abweichung{"Header", "x-mehrfach"})
	u.erwarte(f, "dritte Zeile", headerAendern("x-mehrfach", func(v []string) []string { return append(v, "c") }), abweichung{"Header", "x-mehrfach"})
}

// --- Der Vergleich an handgebauten Antworten

func namenSchreibweisen(n string) []string {
	return []string{n, asciiKlein(n), strings.ToUpper(n)}
}

// TestSchreibweiseDesNamens (Z7): ein Name in anderer Schreibweise ist derselbe Name; net/http kanonisiert auf dem
// Draht, darum stehen die Header handgebaut in Maps. Der Wert zählt mit seiner Schreibweise.
func TestSchreibweiseDesNamens(t *testing.T) {
	n := 0
	for name, werte := range map[string][]string{"X-Dienst": {"ref"}, "Content-Type": {"text/plain"}, "ETag": {`W/"e1"`}, "X-Mehrfach": {"a", "b"}} {
		for _, links := range namenSchreibweisen(name) {
			for _, rechts := range namenSchreibweisen(name) {
				n++
				d := rohantwort{status: 200, header: map[string][]string{links: werte}}
				p := rohantwort{status: 200, header: map[string][]string{rechts: werte}}
				if ab, err := vergleiche(vergleichsregeln{}, d, p); err != nil || len(ab) != 0 {
					t.Errorf("%s als %s und %s: %v, %v", name, links, rechts, ab, err)
				}
				anders := rohantwort{status: 200, header: map[string][]string{rechts: append(slices.Clone(werte), "x")}}
				if ab, err := vergleiche(vergleichsregeln{}, d, anders); err != nil || !slices.Equal(ab, []abweichung{{"Header", asciiKlein(name)}}) {
					t.Errorf("%s als %s und %s mit anderem Wert: %v, %v", name, links, rechts, ab, err)
				}
			}
		}
	}
	if n != 36 {
		t.Errorf("%d Paare von Schreibweisen, erwartet 36", n)
	}
	d := rohantwort{status: 200, header: map[string][]string{"X-Dienst": {"Ref"}}}
	p := rohantwort{status: 200, header: map[string][]string{"x-dienst": {"ref"}}}
	if ab, _ := vergleiche(vergleichsregeln{}, d, p); !slices.Equal(ab, []abweichung{{"Header", "x-dienst"}}) {
		t.Errorf("Wert in anderer Schreibweise: %v", ab)
	}
}

// volatileOrakel sind die sieben Namen von VOLATILE_HEADERS mit ihrem Default (SPEC-063), ausgeschrieben.
func volatileOrakel() []string {
	return []string{"Date", "Server", "Via", "Age", "X-Request-Id", "X-Correlation-Id", "Traceparent"}
}

// TestVolatileHeaders (Z8): jeder Name von VOLATILE_HEADERS, wie der Recorder ihn wirksam führt, ist ausgenommen,
// auch in anderer Schreibweise und bei verschiedenem Wert je Seite; sieben Namen, drei Schreibweisen je Seite.
func TestVolatileHeaders(t *testing.T) {
	volatile := konfig(t, "record", t.TempDir(), "http://127.0.0.1:1/odata").VolatileHeaders
	n := 0
	for _, name := range volatileOrakel() {
		for _, links := range namenSchreibweisen(name) {
			for _, rechts := range namenSchreibweisen(name) {
				n++
				d := rohantwort{status: 200, header: map[string][]string{links: {"1"}}}
				p := rohantwort{status: 200, header: map[string][]string{rechts: {"2"}}}
				if ab, err := vergleiche(vergleichsregeln{volatile: volatile}, d, p); err != nil || len(ab) != 0 {
					t.Errorf("%s als %s und %s: %v, %v", name, links, rechts, ab, err)
				}
				fehlt := rohantwort{status: 200, header: map[string][]string{}}
				if ab, err := vergleiche(vergleichsregeln{volatile: volatile}, d, fehlt); err != nil || len(ab) != 0 {
					t.Errorf("%s als %s nur auf einer Seite: %v, %v", name, links, ab, err)
				}
			}
		}
	}
	if n != 63 {
		t.Errorf("%d Paare, erwartet 63 (sieben Namen, drei Schreibweisen je Seite)", n)
	}
}

// TestHopByHopUndConnection (Z9): jeder Name der festen Liste ist ausgenommen, und jeder Name, den Connection einer der
// beiden Antworten nennt, in jeder Form des Werts, von der direkten, der Proxy-Seite und von beiden.
func TestHopByHopUndConnection(t *testing.T) {
	liste := []string{"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "TE",
		"Trailer", "Transfer-Encoding", "Upgrade"}
	n := 0
	for _, name := range liste {
		for _, schreibweise := range namenSchreibweisen(name) {
			n++
			d := rohantwort{status: 200, header: map[string][]string{schreibweise: {"a1"}}}
			p := rohantwort{status: 200, header: map[string][]string{schreibweise: {"b2"}}}
			if ab, err := vergleiche(vergleichsregeln{}, d, p); err != nil || len(ab) != 0 {
				t.Errorf("%s als %s: %v, %v", name, schreibweise, ab, err)
			}
		}
	}
	if n != 27 {
		t.Errorf("%d Schreibweisen der Liste, erwartet 27", n)
	}
	formen := [][]string{{"x-foo"}, {"  X-FOO\t"}, {"close, x-foo"}, {",x-foo,,"}, {"close", "x-foo"}}
	n = 0
	for _, wert := range formen {
		for _, traeger := range []string{"direkt", "proxy", "beide"} {
			n++
			d := rohantwort{status: 200, header: map[string][]string{"X-Foo": {"1"}, "X-Baz": {"s"}}}
			p := rohantwort{status: 200, header: map[string][]string{"X-Foo": {"2"}, "X-Baz": {"s"}}}
			if traeger != "proxy" {
				d.header["Connection"] = wert
			}
			if traeger != "direkt" {
				p.header["Connection"] = wert
			}
			if ab, err := vergleiche(vergleichsregeln{}, d, p); err != nil || len(ab) != 0 {
				t.Errorf("Connection %q auf %s: %v, %v", wert, traeger, ab, err)
			}
			p.header["X-Baz"] = []string{"t"}
			if ab, _ := vergleiche(vergleichsregeln{}, d, p); !slices.Equal(ab, []abweichung{{"Header", "x-baz"}}) {
				t.Errorf("Connection %q auf %s: ein Name außerhalb von Connection zählt weiter: %v", wert, traeger, ab)
			}
		}
	}
	if n != 15 {
		t.Errorf("%d Formen von Connection, erwartet 15", n)
	}
	// Die Vereinigung beider Seiten: jede nennt einen anderen Namen, beide Namen fallen aus dem Vergleich.
	d := rohantwort{status: 200, header: map[string][]string{"Connection": {"x-foo"}, "X-Foo": {"1"}, "X-Bar": {"1"}}}
	p := rohantwort{status: 200, header: map[string][]string{"Connection": {"x-bar"}, "X-Foo": {"2"}, "X-Bar": {"2"}}}
	if ab, err := vergleiche(vergleichsregeln{}, d, p); err != nil || len(ab) != 0 {
		t.Errorf("Connection nennt auf jeder Seite einen anderen Namen: %v, %v", ab, err)
	}
}

// mitKette baut eine Antwort mit dem Inhalt in der genannten Kette; eine Kette aus einem Glied "-" ist ohne Header.
func mitKette(inhalt string, glieder ...string) rohantwort {
	a := rohantwort{status: 200, header: map[string][]string{}, body: kodiere(glieder, []byte(inhalt))}
	if len(glieder) > 0 {
		a.header["Content-Encoding"] = []string{strings.Join(glieder, ", ")}
	}
	return a
}

// TestCodingUndLaenge (Z10): Content-Encoding und Content-Length sind ausgenommen; ein anderes Coding mit gleichem
// entpacktem Inhalt ist keine Abweichung, auch nicht über den Recorder.
func TestCodingUndLaenge(t *testing.T) {
	gz, de, ohne := mitKette("Anvil Bolt", "gzip"), mitKette("Anvil Bolt", "deflate"), mitKette("Anvil Bolt")
	for _, paar := range [][2]rohantwort{{gz, de}, {gz, ohne}, {ohne, de}} {
		if ab, err := vergleiche(vergleichsregeln{}, paar[0], paar[1]); err != nil || len(ab) != 0 {
			t.Errorf("Coding %q gegen %q: %v, %v", paar[0].header["Content-Encoding"], paar[1].header["Content-Encoding"], ab, err)
		}
	}
	l, r := mitKette("Anvil Bolt"), mitKette("Anvil Bolt")
	l.header["Content-Length"], r.header["Content-Length"] = []string{"10"}, []string{"11"}
	if ab, err := vergleiche(vergleichsregeln{}, l, r); err != nil || len(ab) != 0 {
		t.Errorf("Content-Length 10 gegen 11: %v, %v", ab, err)
	}
	u := neueVgUmgebung(t)
	n := 0
	for _, f := range mitInhalt(u.faelle) {
		for _, neu := range [][]string{nil, {"gzip"}, {"deflate"}, {"deflate", "gzip"}} {
			n++
			if ab := u.mitVeraenderung(f, umkodierenOhneAenderung(neu)); len(ab) != 0 {
				t.Errorf("%s: anderes Coding %v: %v", f.name, neu, ab)
			}
		}
	}
	if n != 44 {
		t.Errorf("%d Fälle des Codings, erwartet 44", n)
	}
}

// TestKetten (Z11): jede Kette der Menge {identity, gzip, deflate, deflate und gzip in einer Zeile, in zwei Zeilen}
// entpackt zum selben Inhalt wie jede andere; ein leerer Rohbody ist bei jeder bekannten Kette leer.
func TestKetten(t *testing.T) {
	zeilen := mitKette("Anvil Bolt", "deflate", "gzip")
	zeilen.header["Content-Encoding"] = []string{"deflate", "gzip"}
	ketten := []rohantwort{mitKette("Anvil Bolt", "identity"), mitKette("Anvil Bolt", "gzip"), mitKette("Anvil Bolt", "deflate"),
		mitKette("Anvil Bolt", "deflate", "gzip"), zeilen}
	n := 0
	for i, l := range ketten {
		for j, r := range ketten {
			n++
			if ab, err := vergleiche(vergleichsregeln{}, l, r); err != nil || len(ab) != 0 {
				t.Errorf("Kette %d gegen Kette %d: %v, %v", i, j, ab, err)
			}
		}
		anders := mitKette("Anvil Bolu", "gzip")
		if ab, err := vergleiche(vergleichsregeln{}, l, anders); err != nil || !slices.Equal(ab, []abweichung{{"Body", ""}}) {
			t.Errorf("Kette %d gegen anderen Inhalt: %v, %v", i, ab, err)
		}
	}
	leer := rohantwort{status: 200, header: map[string][]string{"Content-Encoding": {"gzip"}}}
	gegen := []rohantwort{
		{status: 200, header: map[string][]string{"Content-Encoding": {"identity"}}}, {status: 200, header: map[string][]string{}}, leer,
		{status: 200, header: map[string][]string{"Content-Encoding": {"deflate"}}}, {status: 200, header: map[string][]string{"Content-Encoding": {"deflate, gzip"}}},
	}
	for i, r := range gegen {
		n++
		if ab, err := vergleiche(vergleichsregeln{}, leer, r); err != nil || len(ab) != 0 {
			t.Errorf("leerer Rohbody mit gzip gegen leeren Rohbody %d: %v, %v", i, ab, err)
		}
	}
	if n != 30 {
		t.Errorf("%d Paare, erwartet 30 (25 Ketten, 5 leere Rohbodies)", n)
	}
}

// TestUmkodierenMitAenderung (Z12): ein anderes Coding und ein geändertes Byte zugleich: das Entpacken verdeckt den
// Unterschied nicht; bei jeder Antwort mit Coding über den Recorder.
func TestUmkodierenMitAenderung(t *testing.T) {
	u := neueVgUmgebung(t)
	n := 0
	for _, f := range u.faelle {
		if f.kodierung == "" || f.roh {
			continue
		}
		neu := map[string][]string{"gzip": {"deflate"}, "deflate": {"gzip"}, "deflate, gzip": {}}[f.kodierung]
		for _, pos := range stellen() {
			n++
			u.erwarte(f, "anderes Coding und Byte "+pos, umkodieren(neu, pos), abweichung{"Body", ""})
		}
	}
	if n != 9 {
		t.Errorf("%d Fälle, erwartet 9 (drei Antworten mit Coding, drei Stellen)", n)
	}
}

// TestOhneBody (Z13): eine Antwort ohne Body (HEAD, 204, 304) ist auf beiden Seiten ohne Body, ihre Header gelten.
func TestOhneBody(t *testing.T) {
	u := neueVgUmgebung(t)
	n := 0
	for _, f := range u.faelle {
		if f.name != "head" && f.status != 204 && f.status != 304 {
			continue
		}
		n++
		if d, p := u.direkt.hole(f), u.proxy.hole(f); len(d.body) != 0 || len(p.body) != 0 {
			t.Errorf("%s: Body direkt %q, über den Proxy %q", f.name, d.body, p.body)
		}
		switch f.name {
		case "head":
			u.erwarte(f, "Header", headerWertAendern("etag"), abweichung{"Header", "etag"})
		case "unveraendert":
			u.erwarte(f, "Status", statusWechsel, abweichung{"Status", ""})
			u.erwarte(f, "Header", headerWertAendern("etag"), abweichung{"Header", "etag"})
		case "ohne-inhalt":
			u.erwarte(f, "Body", func(a *rohantwort) error { a.status, a.body = 200, []byte("x"); return nil }, abweichung{"Status", ""}, abweichung{"Header", "content-type"}, abweichung{"Body", ""})
			u.erwarte(f, "Header", headerHinzufuegen("X-Zusatz", "1"), abweichung{"Header", "x-zusatz"})
		}
	}
	if n != 3 {
		t.Errorf("%d Arten ohne Body, erwartet 3", n)
	}
	// Ein Body nach einem 204 oder 304 trägt der Server nicht aus; der Vergleich sieht ihn an der handgebauten Antwort,
	// auf jeder Seite.
	for _, status := range []int{204, 304} {
		leer := rohantwort{status: status, header: map[string][]string{}}
		mit := rohantwort{status: status, header: map[string][]string{}, body: []byte("x")}
		for seite, paar := range map[string][2]rohantwort{"direkt": {mit, leer}, "proxy": {leer, mit}} {
			if ab, err := vergleiche(vergleichsregeln{}, paar[0], paar[1]); err != nil || !slices.Equal(ab, []abweichung{{"Body", ""}}) {
				t.Errorf("%d mit Body auf der Seite %s: %v, %v", status, seite, ab, err)
			}
		}
	}
}

type kaputt struct {
	name string
	ce   []string
	body []byte
}

// kaputteKetten ist die Menge der Antworten, die der Test nicht entpackt (LH-QA-07.a Nr. 3): ein unbekanntes Glied,
// mehr als zwei Glieder, ein leeres Glied, ein Strom, den der Dekoder ablehnt.
func kaputteKetten() []kaputt {
	gz := kodiere([]string{"gzip"}, []byte("Anvil Bolt"))
	ok := []byte("x")
	var out []kaputt
	for _, g := range []string{"br", "zstd", "compress", "x-gzip", "gzipx"} {
		out = append(out, kaputt{"unbekannt " + g, []string{g}, ok})
	}
	for _, g := range []string{"gzip,", ",gzip", "", " ", "gzip, ,deflate"} {
		out = append(out, kaputt{"leeres Glied " + strconv.Quote(g), []string{g}, ok})
	}
	for _, g := range []string{"gzip, gzip, gzip", "identity, gzip, deflate", "identity,identity,identity"} {
		out = append(out, kaputt{"drei Glieder " + g, []string{g}, ok})
	}
	out = append(out, kaputt{"vier Glieder in zwei Zeilen", []string{"gzip, deflate", "gzip, deflate"}, ok})
	out = append(out,
		kaputt{"gzip ohne Strom", []string{"gzip"}, []byte("kein gzip")},
		kaputt{"gzip abgeschnitten", []string{"gzip"}, gz[:len(gz)-6]},
		kaputt{"deflate ohne zlib-Kopf", []string{"deflate"}, []byte{0x01, 0x00, 0x00, 0xff, 0xff}},
		kaputt{"deflate mit gzip-Bytes", []string{"deflate"}, gz},
		kaputt{"gzip mit zlib-Bytes", []string{"gzip"}, kodiere([]string{"deflate"}, []byte("Anvil Bolt"))})
	return out
}

// TestNichtEntpackbar (Z14): eine Kette, die der Test nicht entpackt, und ein abgelehnter Strom machen den Fall rot
// und werden nicht übersprungen, auf jeder Seite; der Status ist dabei gleich.
func TestNichtEntpackbar(t *testing.T) {
	n := 0
	for _, k := range kaputteKetten() {
		schlecht := rohantwort{status: 200, header: map[string][]string{"Content-Encoding": k.ce}, body: k.body}
		gut := rohantwort{status: 200, header: map[string][]string{}, body: []byte("x")}
		for seite, paar := range map[string][2]rohantwort{"direkt": {schlecht, gut}, "proxy": {gut, schlecht}} {
			n++
			if ab, err := vergleiche(vergleichsregeln{}, paar[0], paar[1]); err == nil {
				t.Errorf("%s auf der Seite %s: kein Fehler des Tests, Abweichungen %v", k.name, seite, ab)
			}
		}
	}
	if n != 38 {
		t.Errorf("%d Fälle, erwartet 38 (19 Ketten und Ströme, zwei Seiten)", n)
	}
}

// TestNichtAngefragtesCoding (Z15): br auf Accept-Encoding gzip: Rohbytes und Content-Encoding sind byte-gleich über
// den Recorder, ein geändertes Byte und ein anderer Wert von Content-Encoding sind Abweichungen.
func TestNichtAngefragtesCoding(t *testing.T) {
	u := neueVgUmgebung(t)
	f := u.fall("unbekanntes-coding")
	if ab := u.abweichungen(f, u.proxy.hole(f)); len(ab) != 0 {
		t.Errorf("Rohbytes über den Proxy: %v", ab)
	}
	for _, pos := range stellen() {
		u.erwarte(f, "Byte "+pos, rohByte(pos), abweichung{"Body", ""})
	}
	u.erwarte(f, "Wert von Content-Encoding", headerSetzen("content-encoding", "brx"), abweichung{"Header", "content-encoding"})
	if _, err := vergleiche(vergleichsregeln{}, u.direkt.hole(f), u.proxy.hole(f)); err == nil {
		t.Error("ohne den Modus roh entpackt der Vergleich br")
	}
}

// TestAcceptEncodingErreichtDenDienst (Z16): der Accept-Encoding des Clients erreicht den Dienst wie gesetzt;
// fehlend bleibt fehlend, der Transport entpackt nicht.
func TestAcceptEncodingErreichtDenDienst(t *testing.T) {
	u := neueVgUmgebung(t)
	n := 0
	for _, f := range u.faelle {
		if f.name != "zaehler" && f.name != "gzip" {
			continue
		}
		n++
		soll := "fehlt"
		if f.ae != nil {
			soll = f.ae[0]
		}
		for seite, c := range map[string]*vclient{"direkt": u.direkt, "über den Proxy": u.proxy} {
			a := c.hole(f)
			if got := strings.Join(a.header["X-Echo-Accept-Encoding"], "|"); got != soll {
				t.Errorf("%s, %s: der Dienst empfing Accept-Encoding %q, erwartet %q", f.name, seite, got, soll)
			}
			if f.ae == nil && (len(a.header["Content-Encoding"]) != 0 || string(a.body) != "7") {
				t.Errorf("%s, %s: ohne Accept-Encoding Coding %q und Body %q", f.name, seite, a.header["Content-Encoding"], a.body)
			}
			if f.ae != nil && (!slices.Equal(a.header["Content-Encoding"], []string{"gzip"}) || !bytes.HasPrefix(a.body, []byte{0x1f, 0x8b})) {
				t.Errorf("%s, %s: Coding %q, Body beginnt mit %x", f.name, seite, a.header["Content-Encoding"], a.body[:min(2, len(a.body))])
			}
		}
	}
	if n != 2 {
		t.Errorf("%d Fälle, erwartet 2", n)
	}
}

// TestTextAntworten (Z18): $count und $value sind text/plain mit gleichem Typ und Body über den Recorder; ein anderer
// Typ und ein anderer Wert sind Abweichungen.
func TestTextAntworten(t *testing.T) {
	u := neueVgUmgebung(t)
	n := 0
	for _, f := range u.faelle {
		if f.name != "zaehler" && f.name != "wert" {
			continue
		}
		n++
		p := u.proxy.hole(f)
		if !slices.Equal(p.header["Content-Type"], []string{"text/plain"}) || string(p.body) != f.inhalt {
			t.Errorf("%s über den Proxy: %q mit %q", f.name, p.header["Content-Type"], p.body)
		}
		u.erwarte(f, "Typ", headerSetzen("content-type", "application/json"), abweichung{"Header", "content-type"})
		u.erwarte(f, "Wert", bodySetzen(f.inhalt+"8"), abweichung{"Body", ""})
	}
	if n != 2 {
		t.Errorf("%d Textantworten, erwartet 2", n)
	}
}

// TestAcceptEncodingNurBekannte (Z19): jedes Glied von Accept-Encoding, das der Client in einem Fall nennt, ist gzip,
// deflate oder identity; der Dienst gibt den empfangenen Wert zurück.
func TestAcceptEncodingNurBekannte(t *testing.T) {
	u := neueVgUmgebung(t)
	for _, f := range u.faelle {
		echo := strings.Join(u.direkt.hole(f).header["X-Echo-Accept-Encoding"], "|")
		if echo == "fehlt" {
			continue
		}
		for _, wert := range strings.Split(echo, "|") {
			for _, g := range teileKomma(wert) {
				if !slices.Contains([]string{"gzip", "deflate", "identity"}, asciiKlein(strings.SplitN(g, ";", 2)[0])) {
					t.Errorf("%s: der Client nennt das Coding %q", f.name, g)
				}
			}
		}
	}
}

// TestParallel (Z21): Client, Referenzdienst und Recorder laufen in einem Prozess; mehrere Clients fragen zugleich,
// jeder Fall gleich über den Recorder. Der Detektor von make test-race meldet ein Datenrennen des Referenzdienstes.
func TestParallel(t *testing.T) {
	u := neueVgUmgebung(t)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			<-start
			for _, f := range u.faelle {
				d, err := u.direkt.holeFehler(f)
				if err != nil {
					t.Error(err)
					continue
				}
				p, err := u.proxy.holeFehler(f)
				if err != nil {
					t.Error(err)
					continue
				}
				if ab, err := vergleiche(u.regeln(f), d, p); err != nil || len(ab) != 0 {
					t.Errorf("%s: %v, %v", f.name, ab, err)
				}
			}
		})
	}
	close(start)
	wg.Wait()
}
