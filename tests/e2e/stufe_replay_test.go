package e2e_test

import (
	"fmt"
	"testing"
)

// aufnahmeReplay sind die 8 Requests der Stufe: `$metadata` (der Request der Bereitschaft), beide Seiten,
// die Entität, `$value` und dreimal `$count`.
func aufnahmeReplay() []anfrage {
	return []anfrage{{pfad: "/odata/Products"}, {pfad: "/odata/Products?$skiptoken=1"}, {pfad: "/odata/Products(1)"},
		{pfad: "/odata/Products(1)/Name/$value"}, {pfad: "/odata/Products/$count"}, {pfad: "/odata/Products/$count"},
		{pfad: "/odata/Products/$count"}}
}

// TestE2EReplay ist die Stufe von LH-FA-04 (SPEC-037 Nr. 7): der Replay bei gestopptem Referenzdienst gleich
// der Aufnahme, ohne dass der Referenzdienst je kontaktiert wird.
func TestE2EReplay(t *testing.T) {
	binaryPfad(t)
	ref := neueReferenz(t)
	l := neuerLauf(t, ref)
	l.nimmAuf()
	for _, a := range aufnahmeReplay() {
		l.frage(a)
	}
	l.beendeAufnahme()
	l.importiere()
	t.Run("Übergang", func(t *testing.T) { uebergangFaelle(t, l) })
	vor := ref.zaehler()
	rp := l.spieleAb()
	keinKontakt := func(name string) {
		t.Helper()
		if got := ref.zaehler(); got != vor {
			t.Errorf("%s: Zähler (Verbindungen, Requests) %v, vor dem Replay %v", name, got, vor)
		}
	}
	keinKontakt("nach dem Start")
	t.Run("Aufnahme bei laufendem Server", func(t *testing.T) { durchgang(t, l, rp, "a") })
	keinKontakt("Durchgang A")
	t.Run("Fehlerfälle", func(t *testing.T) { fehlerfaelle(t, l, rp, keinKontakt) })
	ref.stoppe()
	t.Run("Aufnahme bei gestopptem Server", func(t *testing.T) { durchgang(t, l, rp, "b") })
	keinKontakt("Durchgang B")
	t.Run("Sequenz", func(t *testing.T) {
		for _, s := range []string{"s1", "s2"} {
			for i, soll := range zaehlerFolge() {
				w := l.replay(rp, anfrage{pfad: "/odata/Products/$count"}, s)
				if w.status != 200 || string(w.body) != soll {
					t.Errorf("Session %s, $count Nr. %d: %d %q, erwartet %q", s, i+1, w.status, w.body, soll)
				}
			}
		}
	})
	keinKontakt("Sequenz")
}

// durchgang spielt jeden aufgenommenen Request in der Reihenfolge der Aufnahme ab.
func durchgang(t *testing.T, l *lauf, rp *prozess, session string) {
	if len(l.aufnahme) != 8 {
		t.Fatalf("%d aufgezeichnete Requests, erwartet 8", len(l.aufnahme))
	}
	for i, p := range l.aufnahme {
		pruefeGleich(t, fmt.Sprintf("Request %d (%s)", i, p.a.pfad), p.w, l.replay(rp, p.a, session))
	}
}

// fehlerfaelle: kein Treffer ist 404 ODR-E5001, jeder der drei Steuerheader fehlend, leer oder doppelt ist
// 400 ODR-E5003, der vierte $count ist 410 ODR-E5002; der Referenzdienst bleibt unberührt (SPEC-135 Nr. 5).
func fehlerfaelle(t *testing.T, l *lauf, rp *prozess, keinKontakt func(string)) {
	pruefeFehler(t, "nie aufgezeichnet", l.replay(rp, anfrage{pfad: "/odata/Products(99)"}, "f"), 404, "ODR-E5001")
	keinKontakt("kein Treffer")
	for _, h := range []string{"X-Recorder-Workspace", "X-Recorder-Recording-Session", "X-Recorder-Replay-Session"} {
		for _, form := range []string{"fehlend", "leer", "doppelt"} {
			hd := l.replayHeader("f")
			switch form {
			case "fehlend":
				delete(hd, h)
			case "leer":
				hd[h] = []string{""}
			case "doppelt":
				hd[h] = append(hd[h], hd[h][0])
			}
			w := holen(t, rp.adresse, anfrage{pfad: "/odata/Products(1)", header: hd})
			pruefeFehler(t, h+" "+form, w, 400, "ODR-E5003")
			keinKontakt(h + " " + form)
		}
	}
	for range zaehlerFolge() {
		l.replay(rp, anfrage{pfad: "/odata/Products/$count"}, "erschoepft")
	}
	pruefeFehler(t, "vierter $count", l.replay(rp, anfrage{pfad: "/odata/Products/$count"}, "erschoepft"), 410, "ODR-E5002")
	keinKontakt("vierter $count")
}

// uebergangFaelle: unbekannte Session Exit 1 mit ODR-E7003, fehlendes und leeres Argument Exit 2; der
// gültige Aufruf lief im Aufbau und lieferte `zeige` zufolge alle Austausche.
func uebergangFaelle(t *testing.T, l *lauf) {
	out, _, code := uebergang(t, "zeige", l.dir, l.session)
	if code != 0 || out != fmt.Sprintf("status=complete exchanges=%d\n", len(l.aufnahme)) {
		t.Errorf("zeige: Exit %d, stdout %q, erwartet status=complete exchanges=%d", code, out, len(l.aufnahme))
	}
	_, errOut, code := uebergang(t, "import", l.dir, "gibt-es-nicht")
	if code != 1 || !contains(errOut, "ODR-E7003") {
		t.Errorf("unbekannte Session: Exit %d, stderr %q, erwartet 1 mit ODR-E7003", code, errOut)
	}
	for name, args := range map[string][]string{"fehlend": {"import", l.dir}, "leer": {"import", l.dir, ""}} {
		if _, errOut, code := uebergang(t, args...); code != 2 || !contains(errOut, "Aufruf:") {
			t.Errorf("Argument %s: Exit %d, stderr %q, erwartet 2 mit Aufruf:", name, code, errOut)
		}
	}
}
