package bootstrap_test

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/csdl"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodec"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
)

// abEtagMocks baut den Prüfling wie abOeffne, aber mit Dienst.Etag wahr: lax und strikt über derselben Datei und Ablage von u.
func abEtagMocks(t *testing.T, u *abUmgebung) (lax, strikt *mock.Mocker) {
	t.Helper()
	bau := func(strikt bool) *mock.Mocker {
		m, err := mock.New(mock.Konfig{
			Dienste: []mock.Dienst{{ID: "svc", Root: "/odata", Etag: true}}, Metadaten: u.db.Metadata(), Workspaces: u.db.Workspaces(), Entities: u.db.Entities(),
			Coverage: u.db.Coverage(), Blobs: u.store, MetadatenCodec: csdl.New(), URLCodec: urlparser.New(urlparser.WithFilter(urlparser.NewFilterLeser())),
			Schreiber: jsoncodec.NewSchreiber(), Eingabe: jsoncodec.NewEingabe(), Transaktion: u.db, PublicBaseURL: "http://recorder.test",
			StrictCoverage: strikt, MaxBytes: 8388608,
		})
		abMust(t, err)
		return m
	}
	return bau(false), bau(true)
}

// abEtagMitglied ist ein Mitglied `@odata.etag` im Body, wie der Schreiber es hinter `@odata.id` schreibt: `,"@odata.etag":"W/\"<n>\""`; das Muster
// entsteht aus Bytewerten (AGENTS.md §3.18).
var abEtagMitglied = regexp.MustCompile(regexp.QuoteMeta(`,"@odata.etag":"W/`+"\x5c"+`"`) + `[0-9]+` + regexp.QuoteMeta("\x5c"+`""`))

// abEntityPfad sagt, ob der Pfad eine Entity nennt (`/Products(1)`), nicht eine Collection, `$count` oder `$metadata`.
var abEntityPfad = regexp.MustCompile(`^/[A-Za-z]+\([^)]*\)$`)

// abOhneEtag entfernt die Kopfzeile ETag und die Mitglieder `@odata.etag` aus einer Antwort und liefert, wie viele Mitglieder es waren.
func abOhneEtag(ist abIst) (abIst, int) {
	if ist.F != nil || ist.Err != nil {
		return ist, 0
	}
	h := map[string][]string{}
	for k, v := range ist.Header {
		if http.CanonicalHeaderKey(k) != "Etag" {
			h[k] = v
		}
	}
	n := len(abEtagMitglied.FindAllString(ist.Body, -1))
	ist.Header, ist.Body = h, abEtagMitglied.ReplaceAllString(ist.Body, "")
	return ist, n
}

// abEtagErwartet sagt, ob der Schritt die Kopfzeile ETag trägt (SPEC-281 Nr. 1): GET und HEAD einer Entity mit 200, POST mit 201, PATCH und PUT mit 204.
func abEtagErwartet(a abAnfrage, status int) bool {
	switch {
	case status == 200 && (a.Methode == "GET" || a.Methode == "HEAD"):
		return abEntityPfad.MatchString(a.Pfad)
	case status == 201:
		return a.Methode == "POST"
	case status == 204:
		return a.Methode == "PATCH" || a.Methode == "PUT"
	}
	return false
}

// abEtagFixtureLaufen führt die Schritte einer Fixture mit Dienst.Etag wahr aus: vom ETag befreit gleicht jede Antwort der Fixture (Header-Namen,
// Werte, Body byte-gleich), die Kopfzeile ETag steht genau bei den Arten von SPEC-281 Nr. 1, und jedes Mitglied `@odata.id` des Bodys trägt
// ein Mitglied `@odata.etag`.
func abEtagFixtureLaufen(t *testing.T, f abFixture) int {
	t.Helper()
	u := abFrisch(t)
	lax, strikt := abEtagMocks(t, u)
	n := 0
	for _, s := range f.Schritte {
		m := lax
		if s.Strikt {
			m = strikt
		}
		ist := abSende(context.Background(), m, "ws-1", s.Anf)
		n++
		mit := ist
		ohne, mitglieder := abOhneEtag(ist)
		for _, d := range abVergleiche(ohne, s.Soll) {
			t.Errorf("%s %s %s %s: %s", f.ID, s.Name, s.Anf.Methode, s.Anf.Pfad, d)
		}
		if ist.F != nil || ist.Err != nil {
			if len(mit.Header) != 0 {
				t.Errorf("%s %s: Header neben einem Fehler", f.ID, s.Name)
			}
			continue
		}
		kopf := 0
		for k := range mit.Header {
			if http.CanonicalHeaderKey(k) == "Etag" {
				kopf++
			}
		}
		if want := abEtagErwartet(s.Anf, ist.Status); (kopf == 1) != want || kopf > 1 {
			t.Errorf("%s %s %s %s: %d Kopfzeilen ETag, erwartet %v", f.ID, s.Name, s.Anf.Methode, s.Anf.Pfad, kopf, want)
		}
		if ids := strings.Count(ohne.Body, `"@odata.id":"`); mitglieder != ids {
			t.Errorf("%s %s: %d Mitglieder @odata.etag zu %d Mitgliedern @odata.id", f.ID, s.Name, mitglieder, ids)
		}
		// Header und Mitglied tragen denselben Text; die Stufe none schreibt das Mitglied nie (SPEC-169 Nr. 5).
		if h := mit.Header["Etag"]; len(h) == 1 && mit.Body != "" && !abNone(s.Anf) && !strings.Contains(mit.Body, `"@odata.etag":`+jsonText(h[0])) {
			t.Errorf("%s %s: Header %v, Body %s", f.ID, s.Name, h, mit.Body)
		}
	}
	if w, err := u.db.Workspaces().Get(context.Background(), "ws-1"); err != nil || w.Revision != f.Revision {
		t.Errorf("%s: Revision %d (%v), erwartet %d", f.ID, w.Revision, err, f.Revision)
	}
	return n
}

// jsonText ist der Etag als JSON-String (Anführungszeichen mit Backslash), aus Bytewerten.
func jsonText(etag string) string {
	return `"` + strings.ReplaceAll(etag, `"`, "\x5c"+`"`) + `"`
}

// abNone sagt, ob der Schritt die Stufe none verlangt.
func abNone(a abAnfrage) bool {
	for _, v := range a.Header["Accept"] {
		if strings.Contains(v, "odata.metadata=none") {
			return true
		}
	}
	return false
}

// TestAbnahmeMatrixMitEtag (E23): alle 36 IDs mit 92 Schritten gegen den Mock mit Etag an; vom ETag befreit gleichen die Antworten den Fixtures. Die Zählung ist
// die der Abnahme.
func TestAbnahmeMatrixMitEtag(t *testing.T) {
	gelaufen := map[string]int{}
	fixtures := abFixtures()
	schritte := 0
	for _, f := range fixtures {
		t.Run(f.ID, func(t *testing.T) { gelaufen[f.ID] = abEtagFixtureLaufen(t, f) })
	}
	for _, f := range fixtures {
		schritte += gelaufen[f.ID]
	}
	if len(fixtures) != 36 || schritte != 92 {
		t.Errorf("%d IDs und %d Schritte, erwartet 36 und 92", len(fixtures), schritte)
	}
	for _, d := range abPruefeZaehlung(fixtures, gelaufen) {
		t.Error(d)
	}
}
