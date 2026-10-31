package urlparser_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

type ursachenFall struct {
	u                model.URLUrsache
	pfad, query, ort string
}

// ursachenFaelle ist je Ursache eine URL, die genau sie auslöst, mit ihrem Ort. In den Vorlagen stehen die Marker
// {N} (Name), {W} (Wert, Literal) und {S} (Segment), die TestFehlerOhneEingabe belegt.
func ursachenFaelle() []ursachenFall {
	return []ursachenFall{
		{model.URLZuLang, "/{S}" + wiederhole("a", 2100), "", "pfad"},
		{model.URLZuViele, "/{S}" + wiederhole("/a", 17), "", "pfad"},
		{model.URLKodierung, "/{S}%G1", "", "pfad/1"},
		{model.URLSegment, "/{S}-", "", "pfad/1"},
		{model.URLPfadForm, "/$metadata/{S}", "", "pfad/2"},
		{model.URLKey, "/S({N}=1,2)", "", "pfad/1"},
		{model.URLLiteral, "/S({N}={W}x)", "", "pfad/1"},
		{model.URLOptionName, "/S", "{N}=1", "query/1"},
		{model.URLOptionDoppelt, "/S", "$top=1&$top={W}", "query/2"},
		{model.URLWert, "/S", "$top={W}", "query/1"},
		{model.URLServiceDokument, "/", "", "pfad"},
		{model.URLPfadNichtUnterstuetzt, "/$batch/{S}", "", "pfad/1"},
		{model.URLLiteralTyp, "/S({N}=X'{W}')", "", "pfad/1"},
		{model.URLOptionNichtUnterstuetzt, "/S", "$expand={W}", "query/1"},
		{model.URLFilter, "/S", "$filter={W}", "query/1"},
		{model.URLWertNichtUnterstuetzt, "/S", "$select={N}/{W}", "query/1"},
	}
}

func markiert(s string) string {
	s = strings.ReplaceAll(s, "{N}", "ZnN\xc3\xa4")
	s = strings.ReplaceAll(s, "{W}", "ZwW\xc3\xa4")
	return strings.ReplaceAll(s, "{S}", "ZsS\xc3\xa4")
}

// TestFehlerwert (Zeile 4): jede der 16 Ursachen ist ein *URLFehler mit dem Code ihrer Klasse; errors.Is trifft genau die
// Sentinel der Klasse; der Ort hat die Form pfad, pfad/<n>, query, query/<n>; das Ergebnis ist der Nullwert.
func TestFehlerwert(t *testing.T) {
	alle := ursachenFaelle()
	if len(alle) != 16 {
		t.Fatalf("%d Ursachen", len(alle))
	}
	gesehen := map[model.URLUrsache]bool{}
	for _, f := range alle {
		gesehen[f.u] = true
		pruefeFall(t, parserNeu(), fall{pfad: markiert(f.pfad), query: markiert(f.query), u: f.u, ort: f.ort})
		_, err := parse(markiert(f.pfad), markiert(f.query))
		var uf *model.URLFehler
		if !errors.As(err, &uf) {
			t.Fatalf("%s: %v", f.u, err)
		}
		soll := model.CodeE6005
		if klasse501(f.u) {
			soll = model.CodeE6001
		}
		if uf.Code() != soll || uf.Ursache.NichtUnterstuetzt() != klasse501(f.u) {
			t.Errorf("%s: Code %v", f.u, uf.Code())
		}
		msg := "ODR-E6005: OData-URL ungültig: " + string(f.u)
		if klasse501(f.u) {
			msg = "ODR-E6001: OData-URL nicht unterstützt: " + string(f.u)
		}
		if uf.Error() != msg+" "+`"`+f.ort+`"` {
			t.Errorf("%s: Error() = %q", f.u, uf.Error())
		}
	}
	if len(gesehen) != 16 || len(model.AlleURLUrsachen()) != 16 {
		t.Errorf("%d verschiedene Ursachen", len(gesehen))
	}
	for _, u := range model.AlleURLUrsachen() {
		if !gesehen[u] || !u.Gueltig() || u.NichtUnterstuetzt() != klasse501(u) {
			t.Errorf("Ursache %s", u)
		}
	}
	// Ort mit zweistelliger Nummer ohne führende Null.
	pruefeFall(t, parserNeu(), fall{pfad: "/a/b/c/d/e/f/g/h/i/%G1", u: kodi, ort: "pfad/10"})
	pruefeFall(t, parserNeu(), fall{pfad: "/S", query: wiederhole("x=1&", 9) + "$top=-1", u: wertU, ort: "query/10"})
}

// TestFehlerOhneEingabe (Zeile 5): kein Teil der Eingabe steht in Error() oder Ort. Menge: 18 Zellen (Ursache × Marker für
// Name, Wert und Segment), in denen der Marker in der Eingabe steht (von 48 Zellen aus 16 Ursachen × 3 Marker; in den übrigen
// 30 trägt die Eingabe diesen Marker nicht), dazu der Nicht-ASCII-Teil der Marker in den 16 Eingaben; für alle anderen Teile der
// Eingabe gilt die feste Form von Ort und Text über den Korpus (pruefeFehlerform).
func TestFehlerOhneEingabe(t *testing.T) {
	n, mitMarker := 0, 0
	for _, f := range ursachenFaelle() {
		eingabe := markiert(f.pfad) + markiert(f.query)
		_, err := parse(markiert(f.pfad), markiert(f.query))
		var uf *model.URLFehler
		if !errors.As(err, &uf) {
			t.Fatalf("%s: %v", f.u, err)
		}
		for _, marker := range []string{"ZnN", "ZwW", "ZsS"} {
			if strings.Contains(eingabe, marker) {
				mitMarker++
				if strings.Contains(uf.Error(), marker) || strings.Contains(uf.Ort, marker) {
					t.Errorf("%s: Marker %q im Fehler %q", f.u, marker, uf.Error())
				}
			}
		}
		if strings.Contains(uf.Error(), "\xc3\xa4") || strings.Contains(uf.Ort, "\xc3\xa4") {
			t.Errorf("%s: Nicht-ASCII-Marker im Fehler %q", f.u, uf.Error())
		}
		n++
	}
	if n != 16 || mitMarker != 18 {
		t.Fatalf("%d Ursachen, %d Zellen mit Marker, erwartet 16 und 18", n, mitMarker)
	}
	// Auch kein anderer Teil der Eingabe: jeder Fehler des Korpus hat die feste Form von Ort und Text.
	fehler := 0
	for _, f := range korpus() {
		if _, err := parse(f[0], f[1]); err != nil {
			pruefeFehlerform(t, err)
			fehler++
		}
	}
	if fehler < 25 {
		t.Fatalf("nur %d Fehler im Korpus", fehler)
	}
}
