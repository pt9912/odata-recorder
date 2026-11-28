package model_test

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func batchPruefRequest(methode, ziel, id string) model.BatchRequest {
	return model.BatchRequest{ContentID: id, Methode: methode, Ziel: ziel}
}

func batchPruefChangeset(requests ...model.BatchRequest) model.BatchTeil {
	return model.BatchTeil{Changeset: true, Requests: requests}
}

func batchPruefEinzel(requests ...model.BatchRequest) model.BatchTeil {
	return model.BatchTeil{Requests: requests}
}

func batchPruefBatch(teile ...model.BatchTeil) model.Batch {
	return model.Batch{Teile: teile}
}

// batchPruefSoll vergleicht den ganzen Fehlerwert; die Ursache "" verlangt das untypisierte nil. Klasse, Code und Sentinel folgen
// SPEC-182 Nr. 4.
func batchPruefSoll(t *testing.T, name string, b model.Batch, ursache model.BatchUrsache, ort string) {
	t.Helper()
	err := model.PruefeBatch(b)
	if ursache == "" {
		if err != nil {
			t.Errorf("%s: Fehler %#v, erwartet nil", name, err)
		}
		return
	}
	f, ok := err.(*model.BatchFehler)
	if !ok || f == nil || *f != (model.BatchFehler{Ursache: ursache, Ort: ort}) {
		t.Errorf("%s: Fehler %v, erwartet %s %q", name, err, ursache, ort)
		return
	}
	code, sentinel := model.CodeE6007, model.ErrBatchUngueltig
	if ursache == "methode_nicht_unterstuetzt" {
		code, sentinel = model.CodeE6001, model.ErrBatchNichtUnterstuetzt
	}
	if f.Code() != code || !errors.Is(err, sentinel) {
		t.Errorf("%s: Code %v, erwartet %v", name, f.Code(), code)
	}
}

// batchPruefFuellung sind n Teile ohne Verstoß, abwechselnd Einzelrequest und Changeset.
func batchPruefFuellung(n int) []model.BatchTeil {
	teile := []model.BatchTeil{}
	for i := range n {
		if i%2 == 0 {
			teile = append(teile, batchPruefEinzel(batchPruefRequest("GET", "a", "x")))
			continue
		}
		teile = append(teile, batchPruefChangeset(batchPruefRequest("POST", "a", "x"), batchPruefRequest("PATCH", "$x", "y")))
	}
	return teile
}

// batchPruefReihe sind n Requests ohne Verstoß mit den Content-IDs r1 bis rn.
func batchPruefReihe(n int) []model.BatchRequest {
	requests := []model.BatchRequest{}
	for i := range n {
		requests = append(requests, batchPruefRequest("POST", "a", fmt.Sprintf("r%d", i+1)))
	}
	return requests
}

// batchPruefMethodenSoll ist SPEC-184 Nr. 1 im Wortlaut.
func batchPruefMethodenSoll(methode string, imChangeset bool) model.BatchUrsache {
	switch {
	case methode == "POST", methode == "PATCH", methode == "PUT", methode == "DELETE", methode == "GET" && !imChangeset:
		return ""
	case methode == "GET":
		return "methode"
	}
	return "methode_nicht_unterstuetzt"
}

// TestPruefeBatchMethode hält Zeile 37: Methoden × Ebene × Stelle. Menge 11 Methoden × (12 Teilstellen außerhalb + 12 Teilstellen ×
// 12 Requeststellen innerhalb); davor und dahinter stehen Teile und Requests ohne Verstoß.
func TestPruefeBatchMethode(t *testing.T) {
	methoden := []string{"GET", "POST", "PATCH", "PUT", "DELETE", "HEAD", "OPTIONS", "MERGE", "PROPFIND", "get", ""}
	n := 0
	for _, m := range methoden {
		probe := batchPruefRequest(m, "a", "")
		for teil := 1; teil <= 12; teil++ {
			teile := append(batchPruefFuellung(teil-1), batchPruefEinzel(probe))
			teile = append(teile, batchPruefFuellung(2)...)
			batchPruefSoll(t, m, batchPruefBatch(teile...), batchPruefMethodenSoll(m, false), fmt.Sprintf("%d", teil))
			n++
			for stelle := 1; stelle <= 12; stelle++ {
				requests := append(batchPruefReihe(stelle-1), probe, batchPruefRequest("PUT", "a", "z"))
				teile = append(batchPruefFuellung(teil-1), batchPruefChangeset(requests...))
				teile = append(teile, batchPruefFuellung(2)...)
				batchPruefSoll(t, m, batchPruefBatch(teile...), batchPruefMethodenSoll(m, true), fmt.Sprintf("%d/%d", teil, stelle))
				n++
			}
		}
		// Ein von Hand gebauter Teil ohne Changeset mit mehreren Requests: jeder wird nach Changeset beurteilt.
		zwei := batchPruefBatch(batchPruefEinzel(batchPruefRequest("GET", "a", ""), probe))
		batchPruefSoll(t, "zwei ohne Changeset "+m, zwei, batchPruefMethodenSoll(m, false), "1")
	}
	if n != 11*(12+12*12) {
		t.Errorf("Menge %d", n)
	}
}

// batchPruefNachbarn ruft f für jede Schreibweise, jede Einfügung und Ersetzung eines Bytes und jede Auslassung eines Zeichens
// von wort.
func batchPruefNachbarn(wort string, f func(string)) {
	for maske := range 1 << len(wort) {
		w := []byte(wort)
		for i := range w {
			if maske&(1<<i) != 0 {
				w[i] ^= 0x20
			}
		}
		f(string(w))
	}
	for i := 0; i <= len(wort); i++ {
		for b := range 256 {
			c := string([]byte{byte(b)})
			f(wort[:i] + c + wort[i:])
			if i < len(wort) {
				f(wort[:i] + c + wort[i+1:])
			}
		}
		if i < len(wort) {
			f(wort[:i] + wort[i+1:])
		}
	}
}

// TestPruefeBatchMethodeNachbarn hält die Zusage bytegleich aus Zeile 37 an den Nachbarn der fünf Methoden auf beiden Ebenen. Menge
// je Wort der Länge l: 2^l + 256 × (l+1) + 256 × l + l, zusammen 12181.
func TestPruefeBatchMethodeNachbarn(t *testing.T) {
	n := 0
	for _, wort := range []string{"GET", "POST", "PATCH", "PUT", "DELETE"} {
		batchPruefNachbarn(wort, func(m string) {
			r := batchPruefRequest(m, "a", "")
			batchPruefSoll(t, fmt.Sprintf("%q außen", m), batchPruefBatch(batchPruefEinzel(r)), batchPruefMethodenSoll(m, false), "1")
			batchPruefSoll(t, fmt.Sprintf("%q innen", m), batchPruefBatch(batchPruefChangeset(r)), batchPruefMethodenSoll(m, true), "1/1")
			n++
		})
	}
	if n != 12181 {
		t.Errorf("Menge %d", n)
	}
}

// batchPruefKennungen sind die 72 Content-IDs der Länge 1 und 2 aus acht Zeichen, darunter zwei Schreibweisen und führende Nullen.
func batchPruefKennungen() []string {
	zeichen := []string{"a", "A", "0", "1", ".", "~", "-", "_"}
	alle := append([]string{}, zeichen...)
	for _, x := range zeichen {
		for _, y := range zeichen {
			alle = append(alle, x+y)
		}
	}
	return alle
}

// TestPruefeBatchContentID hält Zeile 38 am Changeset: alle Paare von Stellen in Changesets der Länge 2 bis 12 (286), jedes Paar aus
// den 72 Kennungen bytegleich (5184) und die Länge 128.
func TestPruefeBatchContentID(t *testing.T) {
	n := 0
	for laenge := 2; laenge <= 12; laenge++ {
		for i := range laenge {
			for j := i + 1; j < laenge; j++ {
				requests := batchPruefReihe(laenge)
				requests[j].ContentID = requests[i].ContentID
				ort := fmt.Sprintf("1/%d", j+1)
				batchPruefSoll(t, "Paar", batchPruefBatch(batchPruefChangeset(requests...)), "content_id_doppelt", ort)
				// Ein drittes Vorkommen ändert den Ort nicht: gemeldet wird das zweite.
				requests = append(requests, requests[i])
				batchPruefSoll(t, "dreimal", batchPruefBatch(batchPruefChangeset(requests...)), "content_id_doppelt", ort)
				n++
			}
		}
	}
	lang := strings.Repeat("a", 127)
	paare := [][2]string{{lang + "a", lang + "a"}, {lang + "a", lang + "b"}}
	for _, x := range batchPruefKennungen() {
		for _, y := range batchPruefKennungen() {
			paare = append(paare, [2]string{x, y})
		}
	}
	for _, p := range paare {
		soll := model.BatchUrsache("")
		if p[0] == p[1] {
			soll = "content_id_doppelt"
		}
		b := batchPruefBatch(batchPruefChangeset(batchPruefRequest("POST", "a", p[0]), batchPruefRequest("POST", "a", p[1])))
		batchPruefSoll(t, fmt.Sprintf("%q %q", p[0], p[1]), b, soll, "1/2")
	}
	if n != 286 || len(paare) != 72*72+2 {
		t.Errorf("Menge %d und %d", n, len(paare))
	}
}

// TestPruefeBatchContentIDEbene hält Zeile 38 über die Teile: gleiche Werte in zwei Teilen stoßen nicht an (2 Arten × 2 Arten ×
// 3 Abstände), Requests ohne Content-ID auch nicht, und der Ort trägt die Nummer des Teils.
func TestPruefeBatchContentIDEbene(t *testing.T) {
	r := batchPruefRequest("POST", "a", "1")
	arten := []model.BatchTeil{batchPruefEinzel(r), batchPruefChangeset(r)}
	n := 0
	for _, erste := range arten {
		for _, zweite := range arten {
			for abstand := range 3 {
				teile := append([]model.BatchTeil{erste}, batchPruefFuellung(abstand)...)
				batchPruefSoll(t, "zwei Teile", batchPruefBatch(append(teile, zweite)...), "", "")
				n++
			}
		}
	}
	for laenge := 2; laenge <= 5; laenge++ {
		leer := []model.BatchRequest{}
		for range laenge {
			leer = append(leer, batchPruefRequest("POST", "a", ""))
		}
		batchPruefSoll(t, "ohne Content-ID", batchPruefBatch(batchPruefChangeset(leer...)), "", "")
		batchPruefSoll(t, "ohne Content-ID außen", batchPruefBatch(batchPruefEinzel(leer...)), "", "")
	}
	batchPruefSoll(t, "gleich ohne Changeset", batchPruefBatch(batchPruefEinzel(r, r)), "", "")
	for teil := 1; teil <= 12; teil++ {
		teile := append(batchPruefFuellung(teil-1), batchPruefChangeset(r, r))
		batchPruefSoll(t, "Teilnummer", batchPruefBatch(teile...), "content_id_doppelt", fmt.Sprintf("%d/2", teil))
	}
	if n != 12 {
		t.Errorf("Menge %d", n)
	}
}

// batchPruefMethodenChangeset sind die im Changeset erlaubten Methoden (SPEC-184 Nr. 1).
func batchPruefMethodenChangeset() []string { return []string{"POST", "PATCH", "PUT", "DELETE"} }

// batchPruefGestalten sind die Requests ohne Content-ID aus den 4 Methoden im Changeset und 5 Gestalten: nackt, mit Header, mit Body,
// mit beidem und mit Verweis auf q (20).
func batchPruefGestalten() []model.BatchRequest {
	header, body := []model.BatchKopf{{Name: "n", Wert: "w"}}, []byte{'b'}
	alle := []model.BatchRequest{}
	for _, m := range batchPruefMethodenChangeset() {
		alle = append(alle,
			model.BatchRequest{Methode: m, Ziel: "a"},
			model.BatchRequest{Methode: m, Ziel: "a", Header: header},
			model.BatchRequest{Methode: m, Ziel: "a", Body: body},
			model.BatchRequest{Methode: m, Ziel: "a", Header: header, Body: body},
			model.BatchRequest{Methode: m, Ziel: "$q"},
		)
	}
	return alle
}

// TestPruefeBatchContentIDMethode hält die Zeilen 38 und 40 an Methode und Gestalt: 20 Gestalten des Trägers der Content-ID p ×
// 20 Gestalten des zweiten bzw. verweisenden Requests × 3 Abstände. Je Fall ist das zweite Vorkommen doppelt, der Verweis löst auf den
// früheren Träger auf, nicht auf den späteren und nie außerhalb eines Changeset; seine Fortsetzung wechselt mit dem Abstand.
func TestPruefeBatchContentIDMethode(t *testing.T) {
	q := batchPruefRequest("POST", "a", "q")
	n := 0
	for _, traeger := range batchPruefGestalten() {
		traeger.ContentID = "p"
		for _, anderer := range batchPruefGestalten() {
			for abstand := range 3 {
				bau := func(erster, letzter model.BatchRequest) model.Batch {
					requests := append([]model.BatchRequest{q, erster}, batchPruefReihe(abstand)...)
					return batchPruefBatch(batchPruefChangeset(append(requests, letzter)...))
				}
				name := fmt.Sprintf("%s %q, %s %q, Abstand %d", traeger.Methode, traeger.Ziel, anderer.Methode, anderer.Ziel, abstand)
				doppelt, verweis := anderer, anderer
				doppelt.ContentID = "p"
				verweis.Ziel = "$p" + []string{"", "/x", "?a"}[abstand]
				batchPruefSoll(t, name, bau(traeger, doppelt), "content_id_doppelt", fmt.Sprintf("1/%d", abstand+3))
				batchPruefSoll(t, name, bau(traeger, verweis), "", "")
				batchPruefSoll(t, name, bau(verweis, traeger), "referenz", "1/2")
				batchPruefSoll(t, name, batchPruefBatch(batchPruefEinzel(verweis)), "referenz", "1")
				if t.Failed() {
					return
				}
				n++
			}
		}
	}
	if n != 20*20*3 {
		t.Errorf("Menge %d", n)
	}
}

// batchPruefWeite ist die Zahl der Teile eines Batch und der Requests eines Changeset, bis zu der jede Stelle gefahren wird; der Leser
// lässt zusammen 1000 Teile zu (SPEC-183 Nr. 7).
const batchPruefWeite = 1000

// batchPruefPaare zählt die verschiedenen Paare, an denen ein Fänger geprüft hat, und davon die innerhalb von batchPruefWeite.
type batchPruefPaare map[[2]int]bool

func (p batchPruefPaare) imLeser() (n int) {
	for paar := range p {
		if paar[0]+paar[1] <= batchPruefWeite {
			n++
		}
	}
	return n
}

// TestPruefeBatchWeiteTeile hält die Zeilen 37, 38, 40 und 41 an der Teilnummer 1 bis 1000, ringsum Einzelrequests oder Changesets mit
// der Content-ID p: Verstoß eines Einzelrequests (Methode, Verweis) und eines Changeset an Request 1 und 2, ein auflösbarer Verweis.
// Dazu die Kante des Lesers: ein Changeset an Teil n mit dem Verstoß an seinem letzten Request 1000 − n.
func TestPruefeBatchWeiteTeile(t *testing.T) {
	traeger := batchPruefRequest("POST", "a", "p")
	einzel := batchPruefEinzel(batchPruefRequest("GET", "a", "p"))
	proben := []struct {
		teil    model.BatchTeil
		ursache model.BatchUrsache
		request int
	}{
		{batchPruefEinzel(batchPruefRequest("MERGE", "a", "")), "methode_nicht_unterstuetzt", 0},
		{batchPruefChangeset(batchPruefRequest("GET", "a", "")), "methode", 1},
		{batchPruefChangeset(traeger, traeger), "content_id_doppelt", 2},
		{batchPruefChangeset(traeger, batchPruefRequest("PUT", "$fehlt", "")), "referenz", 2},
		{batchPruefChangeset(traeger, batchPruefRequest("PUT", "$p", "")), "", 2},
		{batchPruefEinzel(batchPruefRequest("GET", "$p/x", "")), "referenz", 0},
		{batchPruefChangeset(batchPruefRequest("MERGE", "a", "")), "methode_nicht_unterstuetzt", 1},
		{batchPruefChangeset(batchPruefRequest("PUT", "$p", "")), "referenz", 1},
	}
	paare := batchPruefPaare{}
	for _, ringsum := range []model.BatchTeil{einzel, batchPruefChangeset(traeger)} {
		teile := slices.Repeat([]model.BatchTeil{ringsum}, batchPruefWeite)
		batchPruefSoll(t, "ohne Verstoß", batchPruefBatch(teile...), "", "")
		for n := 1; n <= batchPruefWeite; n++ {
			for _, p := range proben {
				ort := strconv.Itoa(n)
				if p.teil.Changeset {
					ort += "/" + strconv.Itoa(p.request)
					paare[[2]int{n, p.request}] = true
				}
				if p.ursache == "" {
					ort = ""
				}
				teile[n-1] = p.teil
				batchPruefSoll(t, fmt.Sprintf("Teil %d", n), batchPruefBatch(teile...), p.ursache, ort)
			}
			teile[n-1] = ringsum
			if t.Failed() {
				return
			}
		}
	}
	for n := 1; n < batchPruefWeite; n++ {
		requests := batchPruefReihe(batchPruefWeite - n)
		requests[len(requests)-1].Methode = "GET"
		teile := append(slices.Repeat([]model.BatchTeil{einzel}, n-1), batchPruefChangeset(requests...))
		batchPruefSoll(t, "Kante", batchPruefBatch(teile...), "methode", fmt.Sprintf("%d/%d", n, len(requests)))
		paare[[2]int{n, len(requests)}] = true
		if t.Failed() {
			return
		}
	}
	// Teil n mit Request 1 und 2 und die Kante, ohne die Paare in beidem.
	if n := paare.imLeser(); n != 999+998+997 {
		t.Errorf("Menge %d", n)
	}
}

// TestPruefeBatchWeiteRequests hält die Zeilen 37, 38, 40 und 41 an jedem Request 1 bis 1000 eines Changeset, dessen Methoden reihum
// wechseln: GET und MERGE, Verweis auf sich selbst und auf nichts, und gegen den ersten und den vorigen Request das zweite Vorkommen,
// der auflösbare Verweis und dessen Gegenrichtung.
func TestPruefeBatchWeiteRequests(t *testing.T) {
	requests, methoden := batchPruefReihe(batchPruefWeite), batchPruefMethodenChangeset()
	for i := range requests {
		requests[i].Methode = methoden[i%len(methoden)]
	}
	b := batchPruefBatch(batchPruefChangeset(requests...))
	batchPruefSoll(t, "ohne Verstoß", b, "", "")
	paare := map[[2]int]bool{}
	for n := 1; n <= batchPruefWeite; n++ {
		gut := requests[n-1]
		ort := fmt.Sprintf("1/%d", n)
		probe := func(r model.BatchRequest, ursache model.BatchUrsache) {
			t.Helper()
			requests[n-1] = r
			soll := ort
			if ursache == "" {
				soll = ""
			}
			batchPruefSoll(t, fmt.Sprintf("Request %d %q %q", n, r.ContentID, r.Ziel), b, ursache, soll)
			requests[n-1] = gut
		}
		methode, selbst := gut, gut
		methode.Methode = "GET"
		selbst.Ziel = "$" + gut.ContentID
		probe(methode, "methode")
		methode.Methode = "MERGE"
		probe(methode, "methode_nicht_unterstuetzt")
		probe(selbst, "referenz")
		probe(model.BatchRequest{Methode: "PUT", Ziel: "$fehlt"}, "referenz")
		for _, von := range []int{1, n - 1} {
			if von < 1 || von >= n {
				continue
			}
			doppelt, verweis := gut, gut
			doppelt.ContentID = requests[von-1].ContentID
			verweis.Ziel = "$" + requests[von-1].ContentID
			probe(doppelt, "content_id_doppelt")
			probe(verweis, "")
			requests[von-1].Ziel = "$" + gut.ContentID
			batchPruefSoll(t, fmt.Sprintf("von %d auf %d", von, n), b, "referenz", "1/"+strconv.Itoa(von))
			requests[von-1].Ziel = "a"
			paare[[2]int{von, n}] = true
		}
		if t.Failed() {
			return
		}
	}
	// (1, n) und (n − 1, n) für n bis 1000; (1, 2) ist beides.
	if len(paare) != 999+998 {
		t.Errorf("Menge %d", len(paare))
	}
}

// batchPruefVerweisSoll ist SPEC-184 Nr. 3 als regulärer Ausdruck über das ganze Ziel, ohne Code des Modells.
func batchPruefVerweisSoll() func(string) string {
	form := regexp.MustCompile(`(?s)^\$([A-Za-z0-9\-._~]{1,128})(?:[/?].*)?$`)
	reserviert := map[string]bool{"metadata": true, "batch": true, "entity": true, "all": true, "crossjoin": true}
	return func(ziel string) string {
		treffer := form.FindStringSubmatch(ziel)
		if treffer == nil || reserviert[treffer[1]] {
			return ""
		}
		return treffer[1]
	}
}

func batchPruefVerweis(t *testing.T, ziel, soll string) {
	t.Helper()
	r := batchPruefRequest("MERGE", ziel, "1")
	if ist := r.Verweis(); ist != soll {
		t.Errorf("Verweis(%q) = %q, erwartet %q", ziel, ist, soll)
	}
}

// TestBatchVerweis hält Zeile 39 an den benannten Formen, an jedem Byte am Anfang, in der Mitte und am Ende der <id> (256 × 3) und an
// den Längen 0 bis 130 mit vier Fortsetzungen.
func TestBatchVerweis(t *testing.T) {
	for ziel, soll := range map[string]string{
		"$1": "1", "$1/x": "1", "$1?a": "1", "$1/x?a": "1", "$1?a/x": "1", "$1/a/b": "1", "$1?a?b": "1", "$": "", "$/x": "", "$?a": "",
		"/$1": "", "%241": "", "$1(2)": "", "": "", "1": "", "$$1": "", " $1": "", "a$1": "", "$%31": "", "$a%2F": "", "$1#x": "",
	} {
		batchPruefVerweis(t, ziel, soll)
	}
	n := 0
	for b := range 256 {
		c := string([]byte{byte(b)})
		inID := strings.IndexByte(contentIDAlphabet, byte(b)) >= 0
		ende := c == "/" || c == "?"
		for _, f := range []struct{ id, davor string }{{c + "ab", ""}, {"a" + c + "b", "a"}, {"ab" + c, "ab"}} {
			soll := ""
			if inID {
				soll = f.id
			} else if ende {
				soll = f.davor
			}
			batchPruefVerweis(t, "$"+f.id, soll)
			n++
		}
		batchPruefVerweis(t, "$1/"+c, "1")
		batchPruefVerweis(t, "$1?"+c, "1")
	}
	for laenge := range 131 {
		id := strings.Repeat("a", laenge)
		soll := id
		if laenge > 128 {
			soll = ""
		}
		for _, rest := range []string{"", "/x", "?a", "/" + strings.Repeat("y", 300)} {
			batchPruefVerweis(t, "$"+id+rest, soll)
			n++
		}
	}
	if n != 256*3+131*4 {
		t.Errorf("Menge %d", n)
	}
}

// TestBatchVerweisWoerter hält Zeile 39 an den fünf reservierten Wörtern: jede Schreibweise und jeder Nachbar mit den
// Fortsetzungen "", "/x" und "?a"; reserviert ist allein das bytegleiche Wort, verglichen mit der abgeschnittenen <id>.
func TestBatchVerweisWoerter(t *testing.T) {
	orakel := batchPruefVerweisSoll()
	n := 0
	for _, wort := range []string{"metadata", "batch", "entity", "all", "crossjoin"} {
		batchPruefNachbarn(wort, func(w string) {
			for _, rest := range []string{"", "/x", "?a"} {
				batchPruefVerweis(t, "$"+w+rest, orakel("$"+w+rest))
			}
			n++
		})
		batchPruefVerweis(t, "$"+wort, "")
		batchPruefVerweis(t, "$"+wort+"/x", "")
		batchPruefVerweis(t, "$"+wort+"?a", "")
		batchPruefVerweis(t, "$"+strings.ToUpper(wort), strings.ToUpper(wort))
		batchPruefVerweis(t, "$"+wort+"1", wort+"1")
	}
	// je Wort der Länge l: 2^l + 256 × (l+1) + 256 × l + l mit l = 8, 5, 6, 3, 9
	if n != 256+32+64+8+512+256*(31+5)+256*31+31 {
		t.Errorf("Menge %d", n)
	}
}

// TestBatchVerweisZiele hält Zeile 39 an jedem Ziel bis zur Länge 5 über acht Zeichen der Struktur (37449 Ziele).
func TestBatchVerweisZiele(t *testing.T) {
	orakel := batchPruefVerweisSoll()
	zeichen := []string{"$", "/", "?", "a", "1", "%", "(", " "}
	ziele := []string{""}
	for start, laenge := 0, 0; laenge < 5; laenge++ {
		ende := len(ziele)
		for _, z := range ziele[start:ende] {
			for _, c := range zeichen {
				ziele = append(ziele, z+c)
			}
		}
		start = ende
	}
	for _, z := range ziele {
		batchPruefVerweis(t, z, orakel(z))
	}
	if len(ziele) != 1+8+64+512+4096+32768 {
		t.Errorf("Menge %d", len(ziele))
	}
}

// TestPruefeBatchReferenz hält Zeile 40 im Changeset: Länge 1 bis 12 × Stelle des Verweises × Stelle des Ziels × drei Fortsetzungen
// (650 × 3); auflösbar ist allein ein früherer Request. Dazu jedes Paar der 72 Kennungen und die Länge 128 bytegleich.
func TestPruefeBatchReferenz(t *testing.T) {
	n := 0
	for laenge := 1; laenge <= 12; laenge++ {
		for von := range laenge {
			for nach := range laenge {
				for _, rest := range []string{"", "/x", "?a"} {
					requests := batchPruefReihe(laenge)
					requests[von].Ziel = "$" + requests[nach].ContentID + rest
					soll := model.BatchUrsache("referenz")
					if nach < von {
						soll = ""
					}
					batchPruefSoll(t, requests[von].Ziel, batchPruefBatch(batchPruefChangeset(requests...)), soll, fmt.Sprintf("1/%d", von+1))
					n++
				}
			}
		}
	}
	paare := 0
	for _, x := range batchPruefKennungen() {
		for _, y := range batchPruefKennungen() {
			soll := model.BatchUrsache("referenz")
			if x == y {
				soll = ""
			}
			b := batchPruefBatch(batchPruefChangeset(batchPruefRequest("POST", "a", x), batchPruefRequest("PATCH", "$"+y, "")))
			batchPruefSoll(t, fmt.Sprintf("%q %q", x, y), b, soll, "1/2")
			paare++
		}
	}
	if n != 650*3 || paare != 72*72 {
		t.Errorf("Menge %d und %d", n, paare)
	}
	lang := strings.Repeat("a", 127)
	traeger := batchPruefRequest("POST", "a", lang+"a")
	batchPruefSoll(t, "Länge 128", batchPruefBatch(batchPruefChangeset(traeger, batchPruefRequest("PUT", "$"+lang+"a", ""))), "", "")
	anders := batchPruefChangeset(traeger, batchPruefRequest("PUT", "$"+lang+"b", ""))
	batchPruefSoll(t, "Länge 128 anders", batchPruefBatch(anders), "referenz", "1/2")
}

// TestPruefeBatchReferenzEbene hält Zeile 40 über die Teile: ein Träger der Content-ID in einem anderen Teil löst nie auf (2 Arten
// des Trägers × davor und dahinter × Verweis im Changeset und außerhalb), außerhalb eines Changeset ist jeder Verweis ein Verstoß.
func TestPruefeBatchReferenzEbene(t *testing.T) {
	traeger := batchPruefRequest("POST", "a", "1")
	verweis := batchPruefRequest("PATCH", "$1", "")
	n := 0
	for _, anderer := range []model.BatchTeil{batchPruefEinzel(traeger), batchPruefChangeset(traeger)} {
		for _, stelle := range []struct {
			teil model.BatchTeil
			ort  string
		}{{batchPruefEinzel(verweis), ""}, {batchPruefChangeset(verweis), "/1"}} {
			batchPruefSoll(t, "davor", batchPruefBatch(anderer, stelle.teil), "referenz", "2"+stelle.ort)
			batchPruefSoll(t, "dahinter", batchPruefBatch(stelle.teil, anderer), "referenz", "1"+stelle.ort)
			n += 2
		}
	}
	batchPruefSoll(t, "ohne Träger", batchPruefBatch(batchPruefChangeset(verweis)), "referenz", "1/1")
	batchPruefSoll(t, "zwei ohne Changeset", batchPruefBatch(batchPruefEinzel(traeger, verweis)), "referenz", "1")
	kette := batchPruefChangeset(traeger, verweis, verweis, batchPruefRequest("PUT", "$1", "2"), batchPruefRequest("DELETE", "$2", ""))
	batchPruefSoll(t, "Kette", batchPruefBatch(kette), "", "")
	for _, wort := range []string{"metadata", "batch", "entity", "all", "crossjoin"} {
		r := batchPruefRequest("GET", "$"+wort, "")
		batchPruefSoll(t, wort, batchPruefBatch(batchPruefEinzel(r)), "", "")
		post := batchPruefRequest("POST", "$"+wort, wort)
		batchPruefSoll(t, wort+" im Changeset", batchPruefBatch(batchPruefChangeset(post)), "", "")
	}
	if n != 8 {
		t.Errorf("Menge %d", n)
	}
}

// batchPruefVerstoss ist ein Request, der im Changeset hinter dem Träger der Content-ID p die genannten Verstöße trägt.
func batchPruefVerstoss(methode string, doppelt, verweis bool) model.BatchRequest {
	r := batchPruefRequest(methode, "a", "")
	if doppelt {
		r.ContentID = "p"
	}
	if verweis {
		r.Ziel = "$fehlt"
	}
	return r
}

type batchPruefArt struct {
	r       model.BatchRequest
	ursache model.BatchUrsache
}

func batchPruefArten() []batchPruefArt {
	return []batchPruefArt{
		{batchPruefVerstoss("GET", false, false), "methode"},
		{batchPruefVerstoss("MERGE", false, false), "methode_nicht_unterstuetzt"},
		{batchPruefVerstoss("PUT", true, false), "content_id_doppelt"},
		{batchPruefVerstoss("PUT", false, true), "referenz"},
	}
}

// TestPruefeBatchRang hält Zeile 41 am selben Request (3 Methoden × doppelt × Verweis, damit jedes Paar der drei Prüfungen und alle
// drei) und an zwei Requests (4 Arten × 4 Arten × 4 Lagen): der erste Teil, darin der erste Request entscheidet.
func TestPruefeBatchRang(t *testing.T) {
	p := batchPruefRequest("POST", "a", "p")
	gut := batchPruefRequest("POST", "a", "")
	n := 0
	for _, methode := range []string{"PUT", "GET", "MERGE"} {
		for _, doppelt := range []bool{false, true} {
			for _, verweis := range []bool{false, true} {
				soll := batchPruefMethodenSoll(methode, true)
				if soll == "" && doppelt {
					soll = "content_id_doppelt"
				}
				if soll == "" && verweis {
					soll = "referenz"
				}
				r := batchPruefVerstoss(methode, doppelt, verweis)
				batchPruefSoll(t, methode, batchPruefBatch(batchPruefChangeset(p, r)), soll, "1/2")
				n++
			}
		}
	}
	for _, a := range batchPruefArten() {
		for _, b := range batchPruefArten() {
			name := string(a.ursache) + " vor " + string(b.ursache)
			for _, lage := range []struct {
				batch model.Batch
				ort   string
			}{
				{batchPruefBatch(batchPruefChangeset(p, a.r, b.r)), "1/2"},
				{batchPruefBatch(batchPruefChangeset(p, a.r, gut, b.r)), "1/2"},
				{batchPruefBatch(batchPruefChangeset(p, a.r), batchPruefChangeset(p, b.r)), "1/2"},
				{batchPruefBatch(batchPruefChangeset(p, gut, gut, a.r), batchPruefChangeset(p, b.r)), "1/4"},
			} {
				batchPruefSoll(t, name, lage.batch, a.ursache, lage.ort)
				n++
			}
		}
	}
	if n != 3*2*2+4*4*4 {
		t.Errorf("Menge %d", n)
	}
}

// TestPruefeBatchRangEinzel hält Zeile 41 zwischen einem Einzelrequest mit Verstoß (2) und einem Changeset mit Verstoß (4) in beiden
// Folgen.
func TestPruefeBatchRangEinzel(t *testing.T) {
	p := batchPruefRequest("POST", "a", "p")
	einzeln := []batchPruefArt{
		{batchPruefRequest("MERGE", "a", ""), "methode_nicht_unterstuetzt"},
		{batchPruefRequest("GET", "$p", ""), "referenz"},
	}
	n := 0
	for _, e := range einzeln {
		for _, c := range batchPruefArten() {
			cs := batchPruefChangeset(p, c.r)
			batchPruefSoll(t, "Einzel zuerst", batchPruefBatch(batchPruefEinzel(e.r), cs), e.ursache, "1")
			batchPruefSoll(t, "Changeset zuerst", batchPruefBatch(cs, batchPruefEinzel(e.r)), c.ursache, "1/2")
			n += 2
		}
	}
	if n != 2*4*2 {
		t.Errorf("Menge %d", n)
	}
}

// batchPruefUnsortiert baut bei jedem Aufruf denselben Batch neu; Teile und Requests stehen in keiner sortierten Folge.
func batchPruefUnsortiert(letzte string) model.Batch {
	r := func(methode, ziel, id string) model.BatchRequest {
		return model.BatchRequest{
			ContentID: id, Methode: methode, Ziel: ziel,
			Header: []model.BatchKopf{{Name: "b", Wert: " 2 "}, {Name: "A", Wert: "1"}}, Body: []byte{'z', 'a'},
		}
	}
	return batchPruefBatch(
		batchPruefChangeset(r("PUT", "z", "9"), r("POST", "$9/x", "Z"), r("DELETE", "$Z", "1")),
		batchPruefEinzel(r("GET", "b", "7")),
		batchPruefChangeset(r("PATCH", "y", "2"), r("DELETE", letzte, "")),
	)
}

// TestPruefeBatchRein hält Zeile 42: PruefeBatch und Verweis ändern den Wert nicht, mit und ohne Verstoß, und der zweite Aufruf
// liefert dasselbe.
func TestPruefeBatchRein(t *testing.T) {
	orakel := batchPruefVerweisSoll()
	for _, f := range []struct {
		letzte  string
		ursache model.BatchUrsache
	}{{"$2", ""}, {"$9", "referenz"}} {
		b := batchPruefUnsortiert(f.letzte)
		for range 2 {
			batchPruefSoll(t, f.letzte, b, f.ursache, "3/2")
		}
		for _, teil := range b.Teile {
			for _, r := range teil.Requests {
				for range 2 {
					if ist := r.Verweis(); ist != orakel(r.Ziel) {
						t.Errorf("Verweis(%q) = %q", r.Ziel, ist)
					}
				}
			}
		}
		if !reflect.DeepEqual(b, batchPruefUnsortiert(f.letzte)) {
			t.Errorf("%s: Batch geändert: %#v", f.letzte, b)
		}
	}
}

// batchPruefLeereFormen sind die Batches ohne Request, jeder bei jedem Aufruf neu gebaut.
func batchPruefLeereFormen() []model.Batch {
	return []model.Batch{
		{},
		{Teile: []model.BatchTeil{}},
		batchPruefBatch(model.BatchTeil{Changeset: true}),
		batchPruefBatch(model.BatchTeil{Changeset: true, Requests: []model.BatchRequest{}}),
		batchPruefBatch(model.BatchTeil{}),
		batchPruefBatch(model.BatchTeil{Requests: []model.BatchRequest{}}, model.BatchTeil{Changeset: true}),
	}
}

// TestPruefeBatchLeer hält Zeile 42 an den leeren Formen (6); ein leerer Teil zählt für den Ort mit.
func TestPruefeBatchLeer(t *testing.T) {
	formen := batchPruefLeereFormen()
	for i, b := range formen {
		batchPruefSoll(t, fmt.Sprintf("Form %d", i), b, "", "")
		if !reflect.DeepEqual(b, batchPruefLeereFormen()[i]) {
			t.Errorf("Form %d geändert", i)
		}
	}
	leer := model.BatchTeil{Changeset: true}
	b := batchPruefBatch(leer, model.BatchTeil{}, batchPruefEinzel(batchPruefRequest("MERGE", "a", "")))
	batchPruefSoll(t, "hinter leeren Teilen", b, "methode_nicht_unterstuetzt", "3")
	if len(formen) != 6 {
		t.Errorf("Menge %d", len(formen))
	}
}

// batchPruefBeispiel ist das Ergebnis des Beispiels aus SPEC-183 Nr. 8, von Hand gebaut: das Modell ruft keinen Leser.
func batchPruefBeispiel() model.Batch {
	json := []model.BatchKopf{{Name: "Content-Type", Wert: "application/json"}}
	return batchPruefBatch(
		batchPruefEinzel(model.BatchRequest{
			Methode: "GET", Ziel: "Products(1)", Header: []model.BatchKopf{{Name: "Host", Wert: "recorder.test"}},
		}),
		batchPruefChangeset(
			model.BatchRequest{ContentID: "1", Methode: "POST", Ziel: "Products", Header: json, Body: []byte(`{"Id":9}`)},
			model.BatchRequest{ContentID: "2", Methode: "PATCH", Ziel: "$1", Header: json, Body: []byte(`{"Name":"Nail"}`)},
		),
	)
}

// TestPruefeBatchContract ist der Validator-Contract-Test (LH-FA-16.a, Zeile 43): das Beispiel, die vier Zeilen der Prüfung in
// seiner Tabelle und der leere Batch.
func TestPruefeBatchContract(t *testing.T) {
	b := batchPruefBeispiel()
	batchPruefSoll(t, "Beispiel", b, "", "")
	if ist := b.Teile[1].Requests[1].Verweis(); ist != "1" {
		t.Errorf("Verweis in 2/2: %q", ist)
	}
	zeilen := []struct {
		aenderung func(*model.BatchRequest)
		ursache   model.BatchUrsache
	}{
		{func(r *model.BatchRequest) { r.Methode = "GET" }, "methode"},
		{func(r *model.BatchRequest) { r.Methode = "MERGE" }, "methode_nicht_unterstuetzt"},
		{func(r *model.BatchRequest) { r.ContentID = "1" }, "content_id_doppelt"},
		{func(r *model.BatchRequest) { r.Ziel = "$3" }, "referenz"},
	}
	for _, z := range zeilen {
		geaendert := batchPruefBeispiel()
		z.aenderung(&geaendert.Teile[1].Requests[1])
		batchPruefSoll(t, string(z.ursache), geaendert, z.ursache, "2/2")
	}
	batchPruefSoll(t, "leerer Batch", model.Batch{Teile: []model.BatchTeil{}}, "", "")
	if len(zeilen) != 4 {
		t.Errorf("Menge %d", len(zeilen))
	}
}
