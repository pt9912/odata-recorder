package session_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/session"
)

// ungueltig prüft, dass err ErrUngueltig ohne Code ist, der Text die id nicht enthält und weder
// ein Port noch Now noch SessionID gerufen wurde.
func ungueltig(t *testing.T, u *welt, name, id string, err error) {
	t.Helper()
	if !errors.Is(err, model.ErrUngueltig) || !ohneCode(err) {
		t.Errorf("%s: %v", name, err)
	}
	if err != nil && id != "" && strings.Contains(err.Error(), id) {
		t.Errorf("%s: der Text enthält die id: %q", name, err)
	}
	if n := u.st.gesamt(); n != 0 {
		t.Errorf("%s: %d Aufrufe von Ports, Now oder SessionID", name, n)
	}
}

// TestCreateArgumente: leerer serviceID und ungültige id sind ErrUngueltig ohne Code und ohne den
// Wert im Text; kein Port, keine Kennung, keine Zeit (V4).
func TestCreateArgumente(t *testing.T) {
	ids := map[string]string{
		"65 Bytes":       strings.Repeat("a", 65),
		"Großbuchstabe":  "xAy",
		"Unterstrich":    "m_7f",
		"Schrägstrich":   "x/y",
		"Byte 0x80":      "x" + string([]byte{0x80}) + "y",
		"Punkt":          "x.y",
		"Leerzeichen":    "x y",
		"ungültig UTF-8": "x" + string([]byte{0xff}) + "y",
	}
	for name, id := range ids {
		u := neu(t)
		_, err := u.create(context.Background(), dienstA, id)
		ungueltig(t, u, name, id, err)
		if name == "Unterstrich" && strings.Contains(err.Error(), "m_7f") {
			t.Errorf("der Text enthält den Marker: %q", err)
		}
		if id := u.folgeID(); id != sesID(1) {
			t.Errorf("%s: nächste Kennung %s", name, id)
		}
	}
	for _, id := range []string{"", "x1", "xAy"} {
		u := neu(t)
		_, err := u.create(context.Background(), "", id)
		ungueltig(t, u, "serviceID leer, id "+id, id, err)
	}
}

// TestCreate: Kennung (nur bei leerer id) und Zeit je genau einmal vor Sessions.Create und auch
// wenn es scheitert; genau ein Create mit Status open und der bezogenen Zeit; keine Wartung (V5).
func TestCreate(t *testing.T) {
	vorCreate := func(t *testing.T, u *welt, id string, bezogen ...string) {
		t.Helper()
		ev := u.p.liste()
		if len(ev) != len(bezogen)+1 {
			t.Fatalf("Ereignisse %v", ev)
		}
		kopf := slices.Clone(ev[:len(bezogen)])
		slices.Sort(kopf)
		if !slices.Equal(kopf, bezogen) || ev[len(bezogen)] != "Sessions.Create("+id+")" {
			t.Errorf("Ereignisse %v, erwartet %v, dann Create(%s)", ev, bezogen, id)
		}
	}
	// (a) leere id: Kennung und Zeit je einmal.
	u := neu(t)
	u.sessions.fuege("alt", dienstA, t0().Add(-time.Hour), model.SessionOpen)
	s, err := u.create(context.Background(), dienstA, "")
	soll := model.Session{ID: sesID(1), ServiceID: dienstA, CreatedAt: t0(), Status: model.SessionOpen}
	if err != nil || s != soll {
		t.Errorf("(a) %+v, %v; erwartet %+v", s, err, soll)
	}
	if got, ok := u.sessions.hat(sesID(1)); !ok || got != soll {
		t.Errorf("(a) abgelegt %+v", got)
	}
	if n, m := u.st.anzahl("Now"), u.st.anzahl("SessionID"); n != 1 || m != 1 {
		t.Errorf("(a) Now %d, SessionID %d", n, m)
	}
	vorCreate(t, u, sesID(1), "Now", "SessionID")
	if alt, _ := u.sessions.hat("alt"); alt.Status != model.SessionOpen || u.sessions.anzahl() != 2 {
		t.Errorf("(a) die vorhandene Session wurde geändert: %+v", alt)
	}
	if n := u.st.anzahl("CleanupBlobs"); n != 0 {
		t.Errorf("(a) Wartung gerufen: %d", n)
	}
	if id := u.folgeID(); id != sesID(2) {
		t.Errorf("(a) nächste Kennung %s", id)
	}

	// (b) gelieferte id: keine Kennung, Zeit einmal.
	u = neu(t)
	s, err = u.create(context.Background(), dienstA, "x1")
	soll = model.Session{ID: "x1", ServiceID: dienstA, CreatedAt: t0(), Status: model.SessionOpen}
	if err != nil || s != soll {
		t.Errorf("(b) %+v, %v; erwartet %+v", s, err, soll)
	}
	if n, m := u.st.anzahl("Now"), u.st.anzahl("SessionID"); n != 1 || m != 0 {
		t.Errorf("(b) Now %d, SessionID %d", n, m)
	}
	vorCreate(t, u, "x1", "Now")
	if id := u.folgeID(); id != sesID(1) {
		t.Errorf("(b) nächste Kennung %s", id)
	}

	// (c) Sessions.Create scheitert: Kennung und Zeit sind trotzdem bezogen, vor Create.
	u = neu(t)
	u.st.scheitere("Sessions.Create", 1, errDisk)
	if _, err = u.create(context.Background(), dienstA, ""); err == nil {
		t.Fatal("(c) kein Fehler")
	}
	if n, m := u.st.anzahl("Now"), u.st.anzahl("SessionID"); n != 1 || m != 1 {
		t.Errorf("(c) Now %d, SessionID %d", n, m)
	}
	vorCreate(t, u, sesID(1), "Now", "SessionID")
}

// bytesAlsText hält die Form der Fixtures aus Bytewerten, ohne Escape-Folgen im Quelltext.
func bytesAlsText(b ...byte) string { return string(b) }

// formGueltig sagt, ob b zu 0-9, a-z oder - gehört (unabhängig vom Code, Zeichen für Zeichen).
func formGueltig(b byte) bool {
	return (b >= 0x30 && b <= 0x39) || (b >= 0x61 && b <= 0x7a) || b == 0x2d
}

// TestIDForm: 1 bis 64 Bytes aus 0-9, a-z, -; leer vergibt; jedes der 256 Bytewerte einzeln und an
// Anfang, Mitte und Ende (V6).
func TestIDForm(t *testing.T) {
	anlegen := func(id string) (*welt, error) {
		u := neu(t)
		_, err := u.create(context.Background(), dienstA, id)
		return u, err
	}
	// Längen
	u, err := anlegen("")
	if got, ok := u.sessions.hat(sesID(1)); err != nil || !ok || got.ID != sesID(1) {
		t.Errorf("Länge 0 vergibt: %v", err)
	}
	for _, n := range []int{1, 64} {
		if _, err := anlegen(strings.Repeat("a", n)); err != nil {
			t.Errorf("Länge %d: %v", n, err)
		}
	}
	for _, n := range []int{65, 100} {
		if _, err := anlegen(strings.Repeat("a", n)); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("Länge %d: %v", n, err)
		}
	}
	// Alle 256 Bytewerte als Zeichen, an Anfang, Mitte und Ende.
	for i := 0; i < 256; i++ {
		b := byte(i)
		formen := map[string]string{
			"einzeln": bytesAlsText(b), "Anfang": bytesAlsText(b, 'a', 'c'),
			"Mitte": bytesAlsText('a', b, 'c'), "Ende": bytesAlsText('a', 'c', b),
		}
		for ort, id := range formen {
			u, err := anlegen(id)
			if formGueltig(b) {
				if _, ok := u.sessions.hat(id); err != nil || !ok {
					t.Errorf("Byte 0x%02x %s: %v", b, ort, err)
				}
			} else if !errors.Is(err, model.ErrUngueltig) || u.sessions.anzahl() != 0 {
				t.Errorf("Byte 0x%02x %s: %v", b, ort, err)
			}
		}
	}
	// Benannte Zeichen in einer sonst gültigen Folge.
	benannt := map[string]string{
		"Leerzeichen": bytesAlsText(0x20), "Tab": bytesAlsText(0x09), "Zeilenumbruch": bytesAlsText(0x0a),
		"NUL": bytesAlsText(0x00), "DEL": bytesAlsText(0x7f), "ä": string([]rune{0xe4}), "ungültiges UTF-8": bytesAlsText(0xff),
	}
	for name, z := range benannt {
		if _, err := anlegen("a" + z + "c"); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%s: %v", name, err)
		}
	}
	// Die Server-Form ist gültig und trifft eine vergebene Kennung als Doppelung.
	frei := "ses-0123456789abcdef0123456789abcdef"
	if u, err := anlegen(frei); err != nil || len(frei) != 36 {
		t.Errorf("Server-Form: %v", err)
	} else if _, ok := u.sessions.hat(frei); !ok {
		t.Error("Server-Form nicht angelegt")
	}
	u = neu(t)
	if _, err := u.create(context.Background(), dienstA, ""); err != nil {
		t.Fatal(err)
	}
	_, err = u.create(context.Background(), dienstA, sesID(1))
	if c, st := codeVon(t, err); c != "ODR-E7010" || st != 409 {
		t.Errorf("Server-Kennung vergeben: %v", err)
	}
}

// TestCreateFehler: Doppelung ODR-E7010, unbekannter Service ODR-E5005, der Rest nach SPEC-109
// Nr. 5 mit dem Gegenstand Session; ein zweites Create ändert nichts (V7).
func TestCreateFehler(t *testing.T) {
	faelle := []struct {
		name    string
		ursache error
		code    string
		status  int
	}{
		{"Doppelung", verpackt(outbound.ErrBereitsVorhanden), "ODR-E7010", 409},
		{"Service unbekannt", verpackt(outbound.ErrDienstUnbekannt), "ODR-E5005", 404},
		{"beschädigt", verpackt(outbound.ErrBeschaedigt), "ODR-E3001", 500},
		{"Platte", errDisk, "ODR-E9001", 500},
		{"Port weist ab", verpackt(model.ErrUngueltig), "ODR-E9001", 500},
	}
	for _, f := range faelle {
		u := neu(t)
		u.st.scheitere("Sessions.Create", 1, f.ursache)
		s, err := u.create(context.Background(), dienstA, "")
		c, st := codeVon(t, err)
		if c != f.code || st != f.status || !errors.Is(err, f.ursache) || s != (model.Session{}) {
			t.Errorf("%s: %s %d, %+v, %v", f.name, c, st, s, err)
		}
	}
	// Eine Zeit außerhalb der Jahre 1 bis 9999 weist der Port mit ErrUngueltig ab: Code, nicht ohne.
	u := neu(t, func(k *session.Konfig) { k.Uhr = festeZeit(time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)) })
	_, err := u.create(context.Background(), dienstA, "x1")
	if c, st := codeVon(t, err); c != "ODR-E9001" || st != 500 || !errors.Is(err, model.ErrUngueltig) {
		t.Errorf("Jahr 10000: %s %d %v", c, st, err)
	}
	// Wiederholung: das zweite Create mit derselben id ändert die vorhandene Session nicht.
	u = neu(t)
	erste, err := u.create(context.Background(), dienstA, "x1")
	if err != nil {
		t.Fatal(err)
	}
	u.uhr.Advance(time.Hour)
	if err := u.sessions.SetStatus(context.Background(), "x1", model.SessionComplete); err != nil {
		t.Fatal(err)
	}
	_, err = u.create(context.Background(), dienstB, "x1")
	if c, _ := codeVon(t, err); c != "ODR-E7010" {
		t.Errorf("zweites Create: %v", err)
	}
	erste.Status = model.SessionComplete
	if got, _ := u.sessions.hat("x1"); got != erste {
		t.Errorf("die Session wurde geändert: %+v, erwartet %+v", got, erste)
	}
}

// lauf führt op gegen eine frische Umgebung und liefert Ergebnis, Fehler als Text, Ereignisse und Zähler.
type lauf struct {
	s    model.Session
	fehl string
	ev   []string
	z    map[string]int
}

func fuehre(t *testing.T, vorbereite func(*welt), op func(*welt) (model.Session, error)) lauf {
	t.Helper()
	u := neu(t)
	if vorbereite != nil {
		vorbereite(u)
	}
	var s model.Session
	var err error
	if fange(t, "Operation", func() { s, err = op(u) }) {
		err = errPanik
	}
	pruefeFehlerWert(t, err)
	fehl := "nil"
	if err != nil {
		fehl = err.Error()
		if !ohneCode(err) {
			c, st := codeVon(t, err)
			fehl = fmt.Sprintf("%s %d", c, st)
		}
	}
	z := map[string]int{}
	for _, name := range []string{"Now", "SessionID", "Sessions.Create", "CleanupBlobs"} {
		z[name] = u.st.anzahl(name)
	}
	return lauf{s: s, fehl: fehl, ev: u.p.liste(), z: z}
}

// ctxVon liefert den ctx des Falls: nil, wenn ohne gesetzt ist.
func ctxVon(ohne bool) context.Context {
	var keinCtx context.Context
	if ohne {
		return keinCtx
	}
	return context.Background()
}

// TestOpenInitialSession: OpenInitialSession ist Create mit leerer id: dieselbe Session, dieselben
// Ereignisse und Zähler, dieselben Fehler (V8).
func TestOpenInitialSession(t *testing.T) {
	faelle := map[string]struct {
		vorbereite func(*welt)
		dienst     string
		ohneCtx    bool
	}{
		"Gutfall":           {nil, dienstA, false},
		"serviceID leer":    {nil, "", false},
		"Service unbekannt": {func(u *welt) { u.st.scheitere("Sessions.Create", 1, verpackt(outbound.ErrDienstUnbekannt)) }, dienstA, false},
		"Kennung vergeben":  {func(u *welt) { u.sessions.fuege(sesID(1), dienstB, t0(), model.SessionComplete) }, dienstA, false},
		"Platte":            {func(u *welt) { u.st.scheitere("Sessions.Create", 1, errDisk) }, dienstA, false},
		"ctx nil":           {nil, dienstA, true},
	}
	for name, f := range faelle {
		oeffne := fuehre(t, f.vorbereite, func(u *welt) (model.Session, error) {
			return u.v.OpenInitialSession(ctxVon(f.ohneCtx), f.dienst)
		})
		create := fuehre(t, f.vorbereite, func(u *welt) (model.Session, error) {
			return u.v.Create(ctxVon(f.ohneCtx), f.dienst, "")
		})
		if !reflect.DeepEqual(oeffne, create) {
			t.Errorf("%s: OpenInitialSession %+v, Create %+v", name, oeffne, create)
		}
	}
	// Die Werte selbst, damit Gleichheit zweier gleich falscher Läufe nicht genügt.
	g := fuehre(t, nil, func(u *welt) (model.Session, error) { return u.v.OpenInitialSession(context.Background(), dienstA) })
	if soll := (model.Session{ID: sesID(1), ServiceID: dienstA, CreatedAt: t0(), Status: model.SessionOpen}); g.s != soll || g.fehl != "nil" {
		t.Errorf("Gutfall: %+v %s", g.s, g.fehl)
	}
	for name, soll := range map[string]string{"Service unbekannt": "ODR-E5005 404", "Kennung vergeben": "ODR-E7010 409"} {
		f := faelle[name]
		l := fuehre(t, f.vorbereite, func(u *welt) (model.Session, error) { return u.v.OpenInitialSession(context.Background(), f.dienst) })
		if l.fehl != soll {
			t.Errorf("%s: %s", name, l.fehl)
		}
	}
}
