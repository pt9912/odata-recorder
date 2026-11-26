package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/bootstrap"
)

// Die Mengen und die Suche des Canary-Tests (SPEC-041 Nr. 4 und 6, LH-QA-04.a). Die Mengen sind die Default-Listen der Spezifikation und je ein konfigurierter
// Eintrag; sie stehen als Literale hier, nie aus dem Code des Lesers. Alle Namen der Dateien `sicherheit_*_test.go` tragen das Präfix `sicherheit`.

// Der konfigurierte Eintrag je Liste (REDACT_HEADERS, REDACT_QUERY_PARAMS, REDACT_JSON_PATHS) und das Präfix jedes Wertes.
const (
	sicherheitEigenHeader = "X-Canary-Eigen"
	sicherheitEigenQuery  = "canary_eigen"
	sicherheitEigenJSON   = "$..canary_eigen"
	sicherheitPraefix     = "CANARY-"
	sicherheitStellenZahl = 94
)

// sicherheitHeaderDefault sind die sieben Header der Liste von LH-FA-14.a, sicherheitQueryDefault die neun Namen von SPEC-084 und sicherheitJSONDefault die
// sieben Pfade von SPEC-085.
func sicherheitHeaderDefault() []string {
	return []string{"Authorization", "Proxy-Authorization", "Cookie", "Set-Cookie", "X-API-Key", "X-Auth-Token", "X-CSRF-Token"}
}

func sicherheitQueryDefault() []string {
	return []string{"access_token", "token", "api_key", "apikey", "key", "sig", "signature", "code", "client_secret"}
}

func sicherheitJSONDefault() []string {
	return []string{"$..password", "$..passwd", "$..secret", "$..client_secret", "$..access_token", "$..refresh_token", "$..api_key"}
}

// sicherheitListe liefert die Liste mit dem konfigurierten Eintrag (mit) oder ohne ihn.
func sicherheitListe(standard []string, eigen string, mit bool) []string {
	if mit {
		return append(slices.Clone(standard), eigen)
	}
	return slices.Clone(standard)
}

// sicherheitStelle ist ein Ort eines Canary-Wertes: art ist hq (Header der Anfrage), hr (Header der Antwort), qq (Query der Anfrage), ql (Query in
// `Location`), qc (Query in `Content-Location`), jq (Body der Anfrage) oder jr (Body der Antwort); tiefe gilt für JSON.
type sicherheitStelle struct {
	kennung, art, name string
	tiefe              int
}

// anfrage sagt, ob die Stelle auf der Seite der Anfrage liegt.
func (s sicherheitStelle) anfrage() bool { return s.art == "hq" || s.art == "qq" || s.art == "jq" }

// wert ist der Wert der Stelle: `CANARY-<Kennung>-q"+/x`, verschieden je Stelle und in jedem Lauf gleich.
func (s sicherheitStelle) wert() string { return sicherheitPraefix + s.kennung + `-q"+/x` }

// formen sind die drei Formen des Wertes: roh, Query-kodiert (url.QueryEscape) und JSON-maskiert (encoding/json, ohne die Anführungszeichen).
func (s sicherheitStelle) formen() [3]string {
	roh := s.wert()
	j, _ := json.Marshal(roh)
	return [3]string{roh, url.QueryEscape(roh), string(j[1 : len(j)-1])}
}

func sicherheitFormNamen() []string {
	return []string{"roh", "Query-kodiert", "JSON-maskiert", "Präfix"}
}

// sicherheitSlug bildet aus einem Namen einen Teil einer Kennung aus `A-Za-z0-9-`.
func sicherheitSlug(name string) string {
	return strings.ToLower(strings.NewReplacer("_", "-", "$..", "").Replace(name))
}

// sicherheitStellen erzeugt die 94 Stellen aus den Listen: 8 Header in Anfrage und Antwort, 10 Query-Namen in der Query der Anfrage und in den zwei Headern der
// Antwort, 8 JSON-Namen in Anfrage und Antwort in drei Tiefen (Wurzel, verschachteltes Objekt, Objekt in einem Array).
func sicherheitStellen() []sicherheitStelle {
	var aus []sicherheitStelle
	for _, n := range sicherheitListe(sicherheitHeaderDefault(), sicherheitEigenHeader, true) {
		for _, art := range []string{"hq", "hr"} {
			aus = append(aus, sicherheitStelle{kennung: art + "-" + sicherheitSlug(n), art: art, name: n})
		}
	}
	for _, n := range sicherheitListe(sicherheitQueryDefault(), sicherheitEigenQuery, true) {
		for _, art := range []string{"qq", "ql", "qc"} {
			aus = append(aus, sicherheitStelle{kennung: art + "-" + sicherheitSlug(n), art: art, name: n})
		}
	}
	for _, p := range sicherheitListe(sicherheitJSONDefault(), sicherheitEigenJSON, true) {
		for _, art := range []string{"jq", "jr"} {
			for tiefe := range 3 {
				aus = append(aus, sicherheitStelle{kennung: fmt.Sprintf("%s-%s-d%d", art, sicherheitSlug(p), tiefe), art: art, name: strings.TrimPrefix(p, "$.."), tiefe: tiefe})
			}
		}
	}
	return aus
}

// sicherheitAnfrageSeite filtert die Stellen der Seite der Anfrage (wahr) oder der Antwort (falsch).
func sicherheitAnfrageSeite(stellen []sicherheitStelle, anfrage bool) []sicherheitStelle {
	var aus []sicherheitStelle
	for _, s := range stellen {
		if s.anfrage() == anfrage {
			aus = append(aus, s)
		}
	}
	return aus
}

// sicherheitFund ist ein Fund der Suche: Artefakt, Form (eine von sicherheitFormNamen), Kennung der Stelle und die Bytes ab dem Fundort.
type sicherheitFund struct{ artefakt, form, kennung, kontext string }

func (f sicherheitFund) String() string {
	return fmt.Sprintf("Artefakt %s, Form %s, Stelle %s, ab dem Fundort %q", f.artefakt, f.form, f.kennung, f.kontext)
}

// sicherheitKlassiere teilt jedes Vorkommen des Präfixes in b auf: eine der drei Formen einer Stelle (die Form steht vor dem Präfix) oder das Präfix mit
// einem Rest, der keine Form einer Stelle ist (dann trägt der Fund die Kennung, die hinter dem Präfix steht, soweit sie eine Stelle nennt).
func sicherheitKlassiere(artefakt string, b []byte, stellen []sicherheitStelle) []sicherheitFund {
	var funde []sicherheitFund
	for off := 0; off < len(b); {
		i := bytes.Index(b[off:], []byte(sicherheitPraefix))
		if i < 0 {
			break
		}
		pos := off + i
		rest := b[pos:]
		f := sicherheitFund{artefakt: artefakt, form: sicherheitFormNamen()[3], kennung: "(keine Stelle)", kontext: string(rest[:min(48, len(rest))])}
	suche:
		for _, s := range stellen {
			for n, form := range s.formen() {
				if bytes.HasPrefix(rest, []byte(form)) {
					f.form, f.kennung = sicherheitFormNamen()[n], s.kennung
					break suche
				}
			}
			if bytes.HasPrefix(rest, []byte(sicherheitPraefix+s.kennung+"-q")) {
				f.kennung = s.kennung
			}
		}
		funde = append(funde, f)
		off = pos + len(sicherheitPraefix)
	}
	return funde
}

// sicherheitWurzel ist ein durchsuchtes Verzeichnis mit dem Namen, den ein Fund und die Liste der Dateien ihm voranstellen.
type sicherheitWurzel struct{ name, dir string }

// sicherheitErgebnis hält, was eine Suche über die Artefakte eines Falls zu einem Zeitpunkt sah: die Funde der Canary-Werte, die Funde der Kontrollwerte (Datei
// und Strom, `ort` mit dem Namen der Wurzel), die Namen der gelesenen Dateien und die Zahl der Suchzellen.
type sicherheitErgebnis struct {
	funde    []sicherheitFund
	kontroll []fund
	dateien  []string
	zellen   int
}

// sicherheitZaehleDateien zählt die Dateien unter dir mit os.ReadDir, ohne die Suche zu benutzen (der zweite Durchlauf von M-Artefakt).
func sicherheitZaehleDateien(t *testing.T, dir string) int {
	t.Helper()
	es, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Verzeichnis %s: %v", dir, err)
	}
	n := 0
	for _, e := range es {
		if e.IsDir() {
			n += sicherheitZaehleDateien(t, filepath.Join(dir, e.Name()))
		} else {
			n++
		}
	}
	return n
}

// sicherheitSuche sucht in jeder Datei unter den Wurzeln (rekursiv, auf Rohbytes, ungeachtet der Größe, mit `durchsuche`) und in den Strömen nach dem Präfix
// (jede der drei Formen beginnt mit ihm) und nach den Kontrollwerten und ordnet jeden Fund einer der drei Formen einer Stelle oder dem Präfix zu; die Zahl der
// gelesenen Dateien gleicht der eines zweiten Durchlaufs mit os.ReadDir. Die Suchzellen sind die verschiedenen Paare aus Kennung und Form der Menge; welche Form
// die Suche findet, unterscheidet allein die Zuordnung (`TestSicherheitFormen`).
func sicherheitSuche(t *testing.T, wurzeln []sicherheitWurzel, strom map[string][]byte, stellen []sicherheitStelle, kontrollwerte []string) sicherheitErgebnis {
	t.Helper()
	var e sicherheitErgebnis
	werte := append([]string{sicherheitPraefix}, kontrollwerte...)
	for _, w := range wurzeln {
		dateien, funde := durchsuche(t, w.dir, werte)
		if zweiter := sicherheitZaehleDateien(t, w.dir); zweiter != len(dateien) {
			t.Errorf("die Suche las %d Dateien unter %s, ein zweiter Durchlauf zählt %d", len(dateien), w.name, zweiter)
		}
		for _, d := range dateien {
			e.dateien = append(e.dateien, filepath.Join(w.name, d))
		}
		for _, f := range funde {
			ort := filepath.Join(w.name, f.ort)
			if f.wert != sicherheitPraefix {
				e.kontroll = append(e.kontroll, fund{ort, f.wert})
				continue
			}
			b, err := os.ReadFile(filepath.Join(w.dir, f.ort))
			if err != nil {
				t.Fatalf("Datei %s: %v", ort, err)
			}
			e.funde = append(e.funde, sicherheitKlassiere(ort, b, stellen)...)
		}
	}
	for name, b := range strom {
		for _, f := range inBytes(name, b, werte) {
			if f.wert == sicherheitPraefix {
				e.funde = append(e.funde, sicherheitKlassiere(name, b, stellen)...)
			} else {
				e.kontroll = append(e.kontroll, f)
			}
		}
	}
	zellen := map[string]bool{}
	for _, s := range stellen {
		for _, form := range sicherheitFormNamen() {
			zellen[s.kennung+"|"+form] = true
		}
	}
	e.zellen = len(zellen)
	return e
}

// TestSicherheitListen (TS09, M-Liste): die Literale der Mengen haben 7, 9, 7 und mit dem konfigurierten Eintrag 8, 10, 8 Elemente, zusammen 26, und 94 Stellen,
// davon 42 der Anfrage und 52 der Antwort; die wirksamen Listen der Konfiguration (ohne und mit REDACT_*; Header ohne Beachtung der Schreibweise) gleichen ihnen.
func TestSicherheitListen(t *testing.T) {
	h, q, j := sicherheitHeaderDefault(), sicherheitQueryDefault(), sicherheitJSONDefault()
	if len(h) != 7 || len(q) != 9 || len(j) != 7 {
		t.Fatalf("die Default-Listen haben %d, %d und %d Elemente, erwartet 7, 9 und 7", len(h), len(q), len(j))
	}
	stellen := sicherheitStellen()
	if n := len(h) + len(q) + len(j) + 3; n != 26 || len(stellen) != sicherheitStellenZahl ||
		len(sicherheitAnfrageSeite(stellen, true)) != 42 || len(sicherheitAnfrageSeite(stellen, false)) != 52 {
		t.Errorf("%d Elemente, %d Stellen (%d der Anfrage, %d der Antwort); erwartet 26, 94, 42 und 52", len(h)+len(q)+len(j)+3, len(stellen),
			len(sicherheitAnfrageSeite(stellen, true)), len(sicherheitAnfrageSeite(stellen, false)))
	}
	for _, mit := range []bool{false, true} {
		var extra []string
		if mit {
			extra = []string{"REDACT_HEADERS=" + sicherheitEigenHeader, "REDACT_QUERY_PARAMS=" + sicherheitEigenQuery, "REDACT_JSON_PATHS=" + sicherheitEigenJSON}
		}
		k, err := bootstrap.LadeKonfiguration(konfig{modus: "record", dir: t.TempDir(), upstream: "http://127.0.0.1:9", extra: extra}.env("127.0.0.1:8080"))
		if err != nil {
			t.Fatalf("Konfiguration (mit %v): %v", mit, err)
		}
		gelten := func(was string, ist, soll []string, norm func(string) string) {
			a, b := slices.Clone(ist), slices.Clone(soll)
			for i := range a {
				a[i] = norm(a[i])
			}
			for i := range b {
				b[i] = norm(b[i])
			}
			slices.Sort(a)
			slices.Sort(b)
			if !slices.Equal(a, b) {
				t.Errorf("%s (mit konfiguriertem Eintrag: %v): die Konfiguration trägt %v, erwartet %v", was, mit, a, b)
			}
		}
		gelten("REDACT_HEADERS", k.RedactHeaders, sicherheitListe(h, sicherheitEigenHeader, mit), strings.ToLower)
		gelten("REDACT_QUERY_PARAMS", k.RedactQueryParams, sicherheitListe(q, sicherheitEigenQuery, mit), func(s string) string { return s })
		gelten("REDACT_JSON_PATHS", k.RedactJSONPaths, sicherheitListe(j, sicherheitEigenJSON, mit), func(s string) string { return s })
	}
}

// sicherheitDatei schreibt eine Datei mit Inhalt in dir.
func sicherheitDatei(t *testing.T, dir, name string, inhalt []byte) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), inhalt, 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestSicherheitFormen (TS15, M-Form): je eine Datei mit nur der rohen, nur der Query-kodierten, nur der JSON-maskierten Form und nur dem Präfix mit fremdem Rest
// wird mit je einem Fund der richtigen Form gemeldet; die drei Formen jeder der 94 Stellen sind paarweise verschieden, und keine Kennung steht als Anfang vor der
// einer anderen.
func TestSicherheitFormen(t *testing.T) {
	stellen := sicherheitStellen()
	s := stellen[0]
	dir := t.TempDir()
	formen := s.formen()
	sicherheitDatei(t, dir, "roh.txt", []byte("vor "+formen[0]+" nach"))
	sicherheitDatei(t, dir, "query.txt", []byte("vor "+formen[1]+" nach"))
	sicherheitDatei(t, dir, "json.txt", []byte("vor "+formen[2]+" nach"))
	sicherheitDatei(t, dir, "praefix.txt", []byte("vor "+sicherheitPraefix+"fremd-1 nach"))
	e := sicherheitSuche(t, []sicherheitWurzel{{"w", dir}}, nil, stellen, nil)
	soll := map[string]string{"w/roh.txt": "roh", "w/query.txt": "Query-kodiert", "w/json.txt": "JSON-maskiert", "w/praefix.txt": "Präfix"}
	if len(e.funde) != len(soll) {
		t.Errorf("%d Funde, erwartet %d: %v", len(e.funde), len(soll), e.funde)
	}
	for _, f := range e.funde {
		if soll[f.artefakt] != f.form {
			t.Errorf("Fund %v, erwartet die Form %q", f, soll[f.artefakt])
		}
		if f.form != "Präfix" && f.kennung != s.kennung {
			t.Errorf("Fund %v, erwartet die Stelle %s", f, s.kennung)
		}
	}
	if e.zellen != len(stellen)*4 || len(e.dateien) != 4 {
		t.Errorf("%d Suchzellen und %d Dateien, erwartet %d und 4", e.zellen, len(e.dateien), len(stellen)*4)
	}
	paare := 0
	for i, st := range stellen {
		f := st.formen()
		for a := range 3 {
			for b := a + 1; b < 3; b++ {
				paare++
				if f[a] == f[b] {
					t.Errorf("Stelle %s: die Formen %d und %d sind gleich: %q", st.kennung, a, b, f[a])
				}
			}
		}
		for _, anderer := range stellen[i+1:] {
			if strings.HasPrefix(anderer.kennung+"-q", st.kennung+"-q") || strings.HasPrefix(st.kennung+"-q", anderer.kennung+"-q") {
				t.Errorf("die Kennungen %s und %s sind nicht eindeutig", st.kennung, anderer.kennung)
			}
		}
	}
	if paare != sicherheitStellenZahl*3 {
		t.Errorf("%d Paare von Formen, erwartet %d", paare, sicherheitStellenZahl*3)
	}
}
