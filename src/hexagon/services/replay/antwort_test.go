package replay_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// koerper sind die vier Bodies von Z34, aus Bytewerten gebaut (BOM, 0xFF, CRLF, alle 256 Werte).
func koerper() map[string][]byte {
	alle := make([]byte, 0, 65536)
	for i := 0; i < 65536; i++ {
		alle = append(alle, byte(i))
	}
	return map[string][]byte{
		"ein Byte 0x00":  {0x00},
		"BOM und ASCII":  {0xEF, 0xBB, 0xBF, 'a', 'b'},
		"0xFF 0xFE CRLF": {0xFF, 0xFE, 0x0D, 0x0A},
		"65536 Bytes":    alle,
	}
}

// TestStatusUndBodyByteGleich hält Z34: Menge 24 (6 Status × 4 Bodies).
func TestStatusUndBodyByteGleich(t *testing.T) {
	n := 0
	for _, status := range []int{200, 201, 204, 304, 404, 500} {
		for name, body := range koerper() {
			n++
			u := neu(t)
			e := u.mitBody(u.austausch(1, u.keyA(), true), body)
			e.ResponseStatus = status
			u.lege(e)
			resp, err := u.replay(context.Background(), anf())
			if err != nil {
				t.Errorf("%d %s: %v", status, name, err)
				continue
			}
			if resp.Status != status {
				t.Errorf("%d %s: Status %d", status, name, resp.Status)
			}
			if got := lies(t, resp); !bytes.Equal(got, body) {
				t.Errorf("%d %s: Body weicht ab (%d statt %d Bytes)", status, name, len(got), len(body))
			}
		}
	}
	if n != 24 {
		t.Errorf("Menge: %d, erwartet 24", n)
	}
}

// TestBodyAbsent hält Z35: Menge 9 (3 Methoden × 3 Status); leerer Strom, nie nil, Open nicht gerufen.
func TestBodyAbsent(t *testing.T) {
	for _, m := range []string{"GET", "HEAD", "POST"} {
		for _, status := range []int{200, 204, 304} {
			u := neu(t)
			a := anf()
			a.Methode = m
			e := ohneBody(u.austausch(1, u.keyFuer(a, svcID, "/Products"), true))
			e.ResponseStatus = status
			u.lege(e)
			resp, err := u.replay(context.Background(), a)
			if err != nil {
				t.Errorf("%s %d: %v", m, status, err)
				continue
			}
			if resp.Status != status || resp.Body == nil || len(lies(t, resp)) != 0 || u.p.anzahl("Blobs.Open") != 0 {
				t.Errorf("%s %d: Status %d, Body %v, Open %d", m, status, resp.Status, resp.Body, u.p.anzahl("Blobs.Open"))
			}
		}
	}
}

// header sucht den Namen ohne Beachtung der Schreibweise.
func headerOhneSchreibweise(h map[string][]string, name string) ([]string, bool) {
	for n, w := range h {
		if strings.EqualFold(n, name) {
			return w, true
		}
	}
	return nil, false
}

// schreibweisen liefert den Namen kanonisch, klein und groß.
func schreibweisen(n string) []string { return []string{n, strings.ToLower(n), strings.ToUpper(n)} }

// TestHeaderPolicy hält Z36: Menge 135 (15 Namen × 3 Schreibweisen × 3 Konfigurationen); erwartet aus der Handtabelle.
func TestHeaderPolicy(t *testing.T) {
	type name struct {
		n     string
		wert  []string
		bleib func(konfig string) bool
	}
	immer := func(string) bool { return true }
	nie := func(string) bool { return false }
	var namen []name
	for _, f := range fluechtig() {
		namen = append(namen, name{f, []string{"v-" + f}, func(k string) bool { return k != "Default" }})
	}
	namen = append(namen,
		name{"Content-Length", []string{"12"}, nie}, name{"Content-Encoding", []string{"gzip"}, nie},
		name{"X-Recorder-Match-Redacted", []string{"false"}, nie}, name{"Location", []string{"/p?token=" + model.Marker}, immer},
		name{"ETag", []string{`"e1"`}, immer}, name{"Content-Type", []string{"text/plain"}, immer},
		name{"Set-Cookie", []string{model.Marker}, nie}, name{"X-Custom", []string{"c1", "c2"}, func(k string) bool { return k != "X-Custom" }})
	if len(namen) != 15 {
		t.Fatalf("%d Namen", len(namen))
	}
	konfigs := map[string][]string{"Default": fluechtig(), "none": nil, "X-Custom": {"X-Custom"}}
	n := 0
	for _, nm := range namen {
		for _, schreib := range schreibweisen(nm.n) {
			for kn, kv := range konfigs {
				n++
				u := neu(t)
				u.k.Regeln.VolatileHeaders = kv
				e := u.austausch(1, u.keyA(), true)
				e.ResponseHeaders = map[string][]string{schreib: nm.wert, "X-Seq": {"1"}}
				u.lege(e)
				resp, err := u.replay(context.Background(), anf())
				if err != nil {
					t.Errorf("%s %s %s: %v", nm.n, schreib, kn, err)
					continue
				}
				w, da := headerOhneSchreibweise(resp.Header, nm.n)
				switch {
				case nm.bleib(kn) && (!da || !reflect.DeepEqual(w, nm.wert)):
					t.Errorf("%s %s %s: Header fehlt oder verändert: %v", nm.n, schreib, kn, resp.Header)
				case !nm.bleib(kn) && da:
					t.Errorf("%s %s %s: Header ausgeliefert: %v", nm.n, schreib, kn, w)
				}
			}
		}
	}
	if n != 135 {
		t.Errorf("Menge: %d, erwartet 135", n)
	}
}

// TestWertbasiertRedigiert hält Z37: Menge 12 (2 Namen × 6 Wertlisten) plus Location mit Marker im Query-Wert.
func TestWertbasiertRedigiert(t *testing.T) {
	m := model.Marker
	listen := []struct {
		name   string
		werte  []string
		bleibt bool
	}{
		{"leer", []string{}, true}, {"ein leerer Wert", []string{""}, true}, {"Marker", []string{m}, false},
		{"zwei Marker", []string{m, m}, false}, {"Marker und x", []string{m, "x"}, true}, {"x mit Marker", []string{"x" + m}, true},
	}
	n := 0
	for _, nm := range []string{"Set-Cookie", "Location"} {
		for _, l := range listen {
			n++
			u := neu(t)
			e := u.austausch(1, u.keyA(), true)
			e.ResponseHeaders = map[string][]string{nm: l.werte, "X-Seq": {"1"}}
			u.lege(e)
			resp, err := u.replay(context.Background(), anf())
			if err != nil {
				t.Errorf("%s %s: %v", nm, l.name, err)
				continue
			}
			w, da := headerOhneSchreibweise(resp.Header, nm)
			if l.bleibt && (!da || !reflect.DeepEqual(w, l.werte)) || !l.bleibt && da {
				t.Errorf("%s %s: ausgeliefert %v (da %v)", nm, l.name, w, da)
			}
		}
	}
	if n != 12 {
		t.Errorf("Menge: %d, erwartet 12", n)
	}
	u := neu(t)
	e := u.austausch(1, u.keyA(), true)
	e.ResponseHeaders = map[string][]string{"Location": {"/p?token=" + m}, "X-Seq": {"1"}}
	u.lege(e)
	resp, err := u.replay(context.Background(), anf())
	if err != nil || !reflect.DeepEqual(resp.Header["Location"], []string{"/p?token=" + m}) {
		t.Errorf("Location mit Marker im Query-Wert: %v, %v", resp.Header, err)
	}
}

// TestHeaderIstKopie hält Z38: Menge 2; Änderung an der Map oder einer Wertliste ändert den Exchange nicht, und
// zwei Aufrufe liefern unabhängige Maps.
func TestHeaderIstKopie(t *testing.T) {
	for _, was := range []string{"Map", "Wertliste"} {
		u := neu(t)
		u.k.Strategie = model.StrategyFirst
		e := u.austausch(1, u.keyA(), true)
		e.ResponseHeaders = map[string][]string{"X-A": {"v1", "v2"}, "X-Seq": {"1"}}
		u.lege(e)
		r1, err := u.replay(context.Background(), anf())
		if err != nil {
			t.Fatal(err)
		}
		if was == "Map" {
			delete(r1.Header, "X-A")
			r1.Header["X-Neu"] = []string{"n"}
		} else {
			r1.Header["X-A"][0] = "geändert"
		}
		r2, err := u.replay(context.Background(), anf())
		if err != nil {
			t.Fatal(err)
		}
		want := map[string][]string{"X-A": {"v1", "v2"}, "X-Seq": {"1"}}
		if !reflect.DeepEqual(r2.Header, want) {
			t.Errorf("%s: zweiter Aufruf %v, erwartet %v", was, r2.Header, want)
		}
		if got := u.ex.liste[sesID][0].ResponseHeaders; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: Exchange verändert: %v", was, got)
		}
	}
}

// TestMatchRedactedHeader hält Z39: Menge 16 (KeyRedacted × 4 Strategien × Body present/absent) und das Feld des
// Requests bei falschem Feld des Exchange.
func TestMatchRedactedHeader(t *testing.T) {
	n := 0
	for _, redacted := range []bool{false, true} {
		for _, s := range []model.ReplayStrategy{model.StrategySequential, model.StrategyStrict, model.StrategyFirst, model.StrategyLast} {
			for _, absent := range []bool{false, true} {
				n++
				u := neu(t)
				u.k.Strategie = s
				e := u.austausch(1, u.keyA(), true)
				e.KeyRedacted = redacted
				if absent {
					e = ohneBody(e)
				}
				u.lege(e)
				resp, err := u.replay(context.Background(), anf())
				if err != nil {
					t.Errorf("%v %s %v: %v", redacted, s, absent, err)
					continue
				}
				w, da := resp.Header["X-Recorder-Match-Redacted"]
				if redacted && (!da || !reflect.DeepEqual(w, []string{"true"})) || !redacted && da {
					t.Errorf("%v %s %v: Header %v", redacted, s, absent, w)
				}
			}
		}
	}
	if n != 16 {
		t.Errorf("Menge: %d, erwartet 16", n)
	}
	u := neu(t)
	u.k.Regeln.RedactQueryParams = []string{"token"}
	a := anf()
	a.RawQuery = "token=geheim"
	u.lege(u.austausch(1, u.keyFuer(a, svcID, "/Products"), true))
	resp, err := u.replay(context.Background(), a)
	if err != nil {
		t.Fatal(err)
	}
	if _, da := resp.Header["X-Recorder-Match-Redacted"]; da {
		t.Error("Feld des Requests setzt den Header")
	}
	if len(u.tel.Aufrufe()) != 0 {
		t.Error("Feld des Requests schreibt einen Eintrag")
	}
}

// TestProtokoll hält Z40: genau ein Eintrag WARN mit festem Text und sieben Attributen, nach der Bestätigung,
// nie bei einem Fehler, ohne Wert des Requests (Menge 9 plus Fehler bei der Bestätigung).
func TestProtokoll(t *testing.T) {
	canarys := []string{"Pfad", "Query", "Request-Header", "Response-Header", "Body"}
	type fall struct {
		name      string
		redact    bool
		setze     func(u *umg, a *inbound.Anfrage, e *model.Exchange)
		eintraege int
		fehler    bool
	}
	faelle := []fall{
		{"wahr und erfolgreich", true, func(*umg, *inbound.Anfrage, *model.Exchange) {}, 1, false},
		{"wahr mit Auswahlfehler", true, func(_ *umg, _ *inbound.Anfrage, e *model.Exchange) { e.Capture = model.CaptureFailed }, 0, true},
		{"wahr mit Put-Fehler", true, func(u *umg, _ *inbound.Anfrage, _ *model.Exchange) {
			u.p.setzeFehler("tx Cursors.Put", errors.New("put"))
		}, 0, true},
		{"wahr mit Fehler bei der Bestätigung", true, func(u *umg, _ *inbound.Anfrage, _ *model.Exchange) { u.st.commit = errors.New("commit") }, 0, true},
		{"falsch", false, func(*umg, *inbound.Anfrage, *model.Exchange) {}, 0, false},
	}
	for _, c := range canarys {
		c := c
		faelle = append(faelle, fall{"Canary in " + c, true, func(u *umg, a *inbound.Anfrage, e *model.Exchange) {
			switch c {
			case "Pfad":
				a.Pfad += "/" + canary
			case "Query":
				a.RawQuery = "q=" + canary
			case "Request-Header":
				a.Header["X-Canary"] = []string{canary}
			case "Response-Header":
				e.ResponseHeaders["X-Canary"] = []string{canary}
			case "Body":
				*e = u.mitBody(*e, []byte(canary))
			}
		}, 1, false})
	}
	for _, f := range faelle {
		u := neu(t)
		a := anf()
		e := u.austausch(1, "sha256:platzhalter", true)
		f.setze(u, &a, &e)
		e.MatchKey = u.keyFuer(a, svcID, strings.TrimPrefix(a.Pfad, root))
		e.KeyRedacted = f.redact
		u.lege(e)
		resp, err := u.replay(context.Background(), a)
		if f.fehler != (err != nil) {
			t.Errorf("%s: Fehler %v", f.name, err)
			continue
		}
		if err == nil {
			_ = resp.Body.Close()
		}
		logs := u.tel.Aufrufe()
		if len(logs) != f.eintraege {
			t.Errorf("%s: %d Einträge, erwartet %d", f.name, len(logs), f.eintraege)
			continue
		}
		if len(logs) == 0 {
			continue
		}
		l := logs[0]
		want := []support.Attr{support.Str("mode", "replay"), support.Str("ereignis", "key_redacted"), support.Str("service_id", svcID),
			support.Str("session_id", sesID), support.Str("workspace_id", wsID), support.Str("exchange_id", e.ID), support.Str("correlation_id", "cor-a")}
		if l.Op != supporttest.OpLog || l.Level != slog.LevelWarn || l.Msg != "replay: Treffer mit redigiertem Match-Key" || !reflect.DeepEqual(l.Attrs, want) {
			t.Errorf("%s: Eintrag %+v", f.name, l)
		}
		for _, at := range l.Attrs {
			if strings.Contains(at.Text, canary) {
				t.Errorf("%s: Attribut %s nennt den Canary", f.name, at.Key)
			}
		}
		if strings.Contains(l.Msg, canary) {
			t.Errorf("%s: Text nennt den Canary", f.name)
		}
	}
}

// TestReplayfaehigkeit hält Z41: Menge 576 (4 Strategien × 3 × 3 × 4 × 2 × 2 Body-Zustände); geliefert werden 48,
// alle anderen sind der einzige Kandidat und ODR-E5008. Das Orakel ist die Matrix von SPEC-045 und hängt weder von
// der Strategie noch vom Zustand des Bodys ab.
func TestReplayfaehigkeit(t *testing.T) {
	n, geliefert := 0, 0
	for _, strat := range []model.ReplayStrategy{model.StrategySequential, model.StrategyStrict, model.StrategyFirst, model.StrategyLast} {
		for _, capt := range []model.CaptureStatus{model.CaptureComplete, model.CaptureIncomplete, model.CaptureFailed} {
			for _, integ := range []model.BodyIntegrity{model.IntegrityVerified, model.IntegrityAbsent, model.IntegrityInvalid} {
				for _, red := range []model.RedactionStatus{model.RedactionNone, model.RedactionApplied, model.RedactionUnsafe, ""} {
					for _, flag := range []bool{false, true} {
						for _, body := range []string{"present", "absent"} {
							n++
							u := neu(t)
							u.k.Strategie = strat
							u.k.RedigierteBodies = flag
							e := u.austausch(1, u.keyA(), true)
							if body == "absent" {
								e = ohneBody(e)
							}
							e.Capture, e.Integrity, e.ResponseRedact = capt, integ, red
							u.lege(e)
							erlaubt := capt == model.CaptureComplete && (integ == model.IntegrityVerified || integ == model.IntegrityAbsent) &&
								(red == model.RedactionNone || red == model.RedactionApplied && flag)
							resp, err := u.replay(context.Background(), anf())
							was := fmt.Sprintf("%s %s %s %q %v %s", strat, capt, integ, red, flag, body)
							if erlaubt {
								geliefert++
								if err != nil {
									t.Errorf("%s: %v", was, err)
								} else {
									_ = resp.Body.Close()
								}
								continue
							}
							erwarteCode(t, was, err, model.CodeE5008)
							if u.p.anzahl("tx Cursors.Put") != 0 {
								t.Errorf("%s: Put", was)
							}
						}
					}
				}
			}
		}
	}
	if n != 576 || geliefert != 48 {
		t.Errorf("Menge: %d Elemente, %d geliefert; erwartet 576 und 48", n, geliefert)
	}
}

// TestStromSchliessen hält Z43: Menge 3; der Strom wird genau einmal geschlossen, bei Put-Fehler und bei Fehler der
// Bestätigung vom Dienst, bei Erfolg vom Aufrufer.
func TestStromSchliessen(t *testing.T) {
	for _, pfad := range []string{"Put-Fehler", "Fehler beim Bestätigen", "Erfolg"} {
		u := neu(t)
		erfolg(u)
		switch pfad {
		case "Put-Fehler":
			u.p.setzeFehler("tx Cursors.Put", errors.New("put"))
		case "Fehler beim Bestätigen":
			u.st.commit = errors.New("commit")
		}
		resp, err := u.replay(context.Background(), anf())
		st := u.bl.strome()
		if len(st) != 1 {
			t.Fatalf("%s: %d Ströme", pfad, len(st))
		}
		if pfad == "Erfolg" {
			if err != nil || st[0].geschlossen() != 0 {
				t.Errorf("%s: Fehler %v, Close durch den Dienst: %d", pfad, err, st[0].geschlossen())
			}
			if err == nil {
				_ = resp.Body.Close()
			}
		}
		if st[0].geschlossen() != 1 {
			t.Errorf("%s: Strom %d-mal geschlossen, erwartet 1", pfad, st[0].geschlossen())
		}
	}
}

// abbruchTel ist eine Telemetrie, die beim Eintrag den ctx beendet: der Eintrag entsteht nach der Bestätigung.
type abbruchTel struct {
	*supporttest.FakeTelemetry
	cancel context.CancelFunc
}

func (a abbruchTel) Log(ctx context.Context, level slog.Level, msg string, attrs ...support.Attr) {
	a.FakeTelemetry.Log(ctx, level, msg, attrs...)
	a.cancel()
}

// TestAbbruchNachDerAuswahl hält Z44: Menge 3; der Cursor steht nach der Rückkehr fest (position 1), gleich ob der
// Body nie gelesen, halb gelesen oder der ctx nach der Bestätigung beendet wird.
func TestAbbruchNachDerAuswahl(t *testing.T) {
	for _, fall := range []string{"nie gelesen und geschlossen", "halb gelesen und geschlossen", "ctx nach der Bestätigung beendet"} {
		u := neu(t)
		e := u.mitBody(u.austausch(1, u.keyA(), true), bytes.Repeat([]byte("x"), 100))
		ctx, cancel := context.WithCancel(context.Background())
		if fall == "ctx nach der Bestätigung beendet" {
			e.KeyRedacted = true
			u.k.Telemetrie = abbruchTel{u.tel, cancel}
		}
		u.lege(e)
		resp, err := u.replay(ctx, anf())
		if err != nil {
			t.Errorf("%s: %v", fall, err)
			cancel()
			continue
		}
		if c, da := u.cursorStand(); !da || c.Position != 1 {
			t.Errorf("%s: Cursor vor Lesen und Close %+v", fall, c)
		}
		if fall == "halb gelesen und geschlossen" {
			_, _ = resp.Body.Read(make([]byte, 50))
		}
		_ = resp.Body.Close()
		cancel()
		if c, da := u.cursorStand(); !da || c.Position != 1 || !reflect.DeepEqual(c.Consumed, []int64{1}) {
			t.Errorf("%s: Cursor nach Close und Abbruch %+v", fall, c)
		}
	}
}
