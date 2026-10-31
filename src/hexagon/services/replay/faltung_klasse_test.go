package replay_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Klasse dieser Datei ist die ASCII-Faltung (SPEC-135, Kopf): Header-Namen und der Wert identity vergleicht der
// Dienst nach ASCII, ohne Beachtung der Schreibweise. Die Menge ist jedes Byte 0x00 bis 0xFF an jeder Stelle des
// Namens, auf beiden Seiten des Vergleichs; das Orakel steht hier von Hand, nicht in asciiKlein, und ist die Aussage
// des Vertrags: zwei Bytes sind gleich, wenn sie gleich sind oder beide Buchstaben, die sich allein in der Schreibweise
// unterscheiden.

func buchstabe(c byte) bool { return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' }

// gleichNachASCII ist das Orakel der Faltung von Hand: ein Buchstabe und sein Partner in der anderen Schreibweise
// unterscheiden sich allein im Bit 0x20.
func gleichNachASCII(a, b byte) bool {
	return a == b || buchstabe(a) && buchstabe(b) && a^b == 0x20
}

// TestFaltungOrakel hält das Orakel selbst: Menge 65 536 Paare, gleich genau bei 256 + 2·26 = 308 Paaren.
func TestFaltungOrakel(t *testing.T) {
	gleich := 0
	for a := 0; a < 256; a++ {
		for b := 0; b < 256; b++ {
			if gleichNachASCII(byte(a), byte(b)) {
				gleich++
			}
		}
	}
	if gleich != 256+2*26 {
		t.Errorf("%d gleiche Paare, erwartet %d", gleich, 256+2*26)
	}
}

// namePos setzt b an die Stelle pos eines Namens.
func namePos(name string, pos int, b byte) string {
	r := []byte(name)
	r[pos] = b
	return string(r)
}

// TestHeaderNamenAntwortUndKonfig hält Z36 für die Faltung der Namen in der Antwort: Menge 196 608 = Stelle {Anfang,
// Mitte, Ende eines Namens aus drei Bytes} × Byte der Konfiguration (VOLATILE_HEADERS) {256} × Byte der Antwort {256}.
// Der Header der Antwort entfällt genau dann, wenn beide Bytes gleich nach ASCII sind (Orakel von Hand).
func TestHeaderNamenAntwortUndKonfig(t *testing.T) {
	n, abweichungen := 0, 0
	for stelle := 0; stelle < 3; stelle++ {
		for b1 := 0; b1 < 256; b1++ {
			u := neu(t)
			u.k.Strategie = model.StrategyFirst
			u.k.Regeln.VolatileHeaders = []string{namePos("q-q", stelle, byte(b1))}
			r := u.replayer()
			basis := u.austausch(1, u.keyA(), true)
			for b2 := 0; b2 < 256; b2++ {
				n++
				name := namePos("q-q", stelle, byte(b2))
				e := basis
				e.ResponseHeaders = map[string][]string{name: {"v"}}
				u.lege(e)
				resp, err := r.Replay(context.Background(), anf())
				if err != nil {
					t.Fatalf("Stelle %d, Konfig %#x, Antwort %#x: %v", stelle, b1, b2, err)
				}
				_ = resp.Body.Close()
				if _, bleibt := resp.Header[name]; bleibt == gleichNachASCII(byte(b1), byte(b2)) {
					abweichungen++
					if abweichungen <= 5 {
						t.Errorf("Stelle %d, Konfig %#x, Antwort %#x: Header bleibt %v", stelle, b1, b2, bleibt)
					}
				}
			}
		}
	}
	meldeAbweichungen(t, abweichungen)
	if n != 3*256*256 {
		t.Errorf("Menge: %d, erwartet %d", n, 3*256*256)
	}
}

func meldeAbweichungen(t *testing.T, n int) {
	t.Helper()
	if n > 5 {
		t.Errorf("%d Abweichungen insgesamt", n)
	}
}

// TestHeaderNamenDerAnfrage hält Z10, Z17 und Z18 für die Faltung der Namen im Request: Menge 3840 = Name {die drei
// Steuerheader, Content-Type, Content-Encoding} × Stelle {Anfang, Mitte, Ende} × Byte {256}. Der Name mit dem Byte an
// der Stelle ist der Name des Vertrags genau dann, wenn das Byte dem Zeichen gleich nach ASCII ist (Orakel von Hand);
// sonst fehlt der Header: ein Steuerheader ergibt E5003, ein fehlendes Content-Type (typ) E5001, ein fehlendes
// Content-Encoding hebt die Ablehnung von gzip auf (Erfolg). Erreichbar sind allein die Buchstaben der fünf Namen; die
// übrigen trägt die Menge der Antwort.
func TestHeaderNamenDerAnfrage(t *testing.T) {
	type name struct {
		text     string
		mitBody  bool
		wert     []string
		gleich   model.Code // 0 ist Erfolg
		ungleich model.Code
	}
	namen := []name{
		{"X-Recorder-Workspace", false, []string{wsID}, 0, model.CodeE5003},
		{"X-Recorder-Recording-Session", false, []string{sesID}, 0, model.CodeE5003},
		{"X-Recorder-Replay-Session", false, []string{repID}, 0, model.CodeE5003},
		{"Content-Type", true, []string{"application/json"}, 0, model.CodeE5001},
		{"Content-Encoding", true, []string{"gzip"}, model.CodeE5001, 0},
	}
	body := []byte(`{"a":1}`)
	n, abweichungen := 0, 0
	for _, nm := range namen {
		for _, pos := range []int{0, len(nm.text) / 2, len(nm.text) - 1} {
			for b := 0; b < 256; b++ {
				n++
				u := neu(t)
				a := anf()
				key := u.keyA()
				if nm.mitBody {
					a = mitBodyAnfrage(body, "application/json")
					rb, rr := redigiert(t, body, nil)
					key = u.keyMitBody(a, rb, rr)
				}
				delete(a.Header, nm.text)
				a.Header[namePos(nm.text, pos, byte(b))] = nm.wert
				u.lege(u.austausch(1, key, true))
				resp, err := u.replay(context.Background(), a)
				want := nm.ungleich
				if gleichNachASCII(byte(b), nm.text[pos]) {
					want = nm.gleich
				}
				if err == nil {
					_ = resp.Body.Close()
				}
				if code, _, da := fehlerCode(err); (want == 0) != (err == nil) || err != nil && (!da || code != want) {
					abweichungen++
					if abweichungen <= 5 {
						t.Errorf("%s Stelle %d Byte %#x: Fehler %v, erwartet Code %v", nm.text, pos, b, err, want)
					}
				}
			}
		}
	}
	meldeAbweichungen(t, abweichungen)
	if n != 5*3*256 {
		t.Errorf("Menge: %d, erwartet %d", n, 5*3*256)
	}
}

// TestIdentityJedesByte hält Z18 für die Faltung des Werts identity: Menge 2048 = Stelle {8} × Byte {256}. Der Wert
// mit dem Byte an der Stelle ist identity genau dann, wenn das Byte dem Zeichen gleich nach ASCII ist (Orakel von
// Hand); sonst ist es E5001 ohne ein gelesenes Byte.
func TestIdentityJedesByte(t *testing.T) {
	const identity = "identity"
	body := []byte(`{"a":1}`)
	n, abweichungen := 0, 0
	for pos := 0; pos < len(identity); pos++ {
		for b := 0; b < 256; b++ {
			n++
			u := neu(t)
			rd := &zaehlLeser{daten: bytes.Clone(body)}
			a := mitBodyAnfrage(body, "application/json")
			a.Body = rd
			a.Header["Content-Encoding"] = []string{namePos(identity, pos, byte(b))}
			rb, rr := redigiert(t, body, nil)
			u.lege(u.austausch(1, u.keyMitBody(a, rb, rr), true))
			resp, err := u.replay(context.Background(), a)
			code, _, da := fehlerCode(err)
			if gleichNachASCII(byte(b), identity[pos]) {
				if err != nil {
					abweichungen++
					if abweichungen <= 5 {
						t.Errorf("Stelle %d Byte %#x: %v, erwartet identity", pos, b, err)
					}
					continue
				}
				_ = resp.Body.Close()
			} else if !da || code != model.CodeE5001 || rd.gelesen != 0 {
				abweichungen++
				if abweichungen <= 5 {
					t.Errorf("Stelle %d Byte %#x: %v, %d Bytes gelesen; erwartet E5001 ohne Lesen", pos, b, err, rd.gelesen)
				}
			}
		}
	}
	meldeAbweichungen(t, abweichungen)
	if n != 8*256 {
		t.Errorf("Menge: %d, erwartet %d", n, 8*256)
	}
}
