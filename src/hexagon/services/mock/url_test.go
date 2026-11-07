package mock_test

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// urlKlassen ist das Orakel von Z08: der Code je Ursache, von Hand gesetzt (SPEC-160 Nr. 4).
func urlKlassen() map[model.URLUrsache]model.Code {
	return map[model.URLUrsache]model.Code{
		model.URLZuLang:                  model.CodeE6005,
		model.URLZuViele:                 model.CodeE6005,
		model.URLKodierung:               model.CodeE6005,
		model.URLSegment:                 model.CodeE6005,
		model.URLPfadForm:                model.CodeE6005,
		model.URLKey:                     model.CodeE6005,
		model.URLLiteral:                 model.CodeE6005,
		model.URLOptionName:              model.CodeE6005,
		model.URLOptionDoppelt:           model.CodeE6005,
		model.URLWert:                    model.CodeE6005,
		model.URLServiceDokument:         model.CodeE6001,
		model.URLPfadNichtUnterstuetzt:   model.CodeE6001,
		model.URLLiteralTyp:              model.CodeE6001,
		model.URLOptionNichtUnterstuetzt: model.CodeE6001,
		model.URLFilter:                  model.CodeE6001,
		model.URLWertNichtUnterstuetzt:   model.CodeE6001,
	}
}

// TestURLUrsachen hält Z08: über alle sechzehn Ursachen des Modells ist die Klasse `ungueltig` ODR-E6005 und die Klasse
// `nicht_unterstuetzt` ODR-E6001; die Ursache bleibt über Unwrap erreichbar, der Ort steht in keinem Text, und kein Workspace
// wird gelesen. Der Pfad des Service-Roots erreicht den Leser als `/` und endet mit ODR-E6001.
func TestURLUrsachen(t *testing.T) {
	alle, klassen := model.AlleURLUrsachen(), urlKlassen()
	if len(alle) != 16 || len(klassen) != 16 {
		t.Fatalf("%d Ursachen im Modell, %d im Orakel; Z08 nennt 16", len(alle), len(klassen))
	}
	for _, ursache := range alle {
		code, da := klassen[ursache]
		if !da {
			t.Errorf("Ursache %s fehlt im Orakel", ursache)
			continue
		}
		for _, pfad := range []string{dienstRoot + "/Set", dienstRoot, dienstRoot + "/"} {
			u := neueUmg(t)
			injiziert := &model.URLFehler{Ursache: ursache, Ort: "ort-" + kanarienwert}
			var gesehen string
			u.c.url.f = func(p, _ string) (model.ODataURL, error) { gesehen = p; return model.ODataURL{}, injiziert }
			_, err := u.rufe(anfrage("GET", pfad))
			f := erwarteFehler(t, string(ursache)+" "+pfad, err, code)
			if f == nil {
				continue
			}
			var uf *model.URLFehler
			if !errors.As(f, &uf) || uf != injiziert {
				t.Errorf("%s: Ursache %v nicht über Unwrap erreichbar", ursache, f.Ursache)
			}
			if strings.Contains(f.Error(), kanarienwert) {
				t.Errorf("%s: Text nennt den Ort", ursache)
			}
			if ops := u.p.ops(); !reflect.DeepEqual(ops, []string{"URL.ParseURL"}) {
				t.Errorf("%s: Aufrufe %v", ursache, ops)
			}
			if pfad != dienstRoot+"/Set" && gesehen != "/" {
				t.Errorf("%s: Pfad %q erreicht den Leser als %q, erwartet /", ursache, pfad, gesehen)
			}
		}
	}
}

// TestKompatUndRest hält Z09: jede der sechs Ursachen eines KompatFehler ist ODR-E6006, ErrUngueltig ohne URLFehler und
// jeder andere Fehler des Lesers ODR-E9001; eine Art der Ressource außerhalb der vier ist ODR-E9001.
func TestKompatUndRest(t *testing.T) {
	alle := model.AlleKompatUrsachen()
	if len(alle) != 6 {
		t.Fatalf("%d Ursachen, Z09 nennt 6", len(alle))
	}
	for _, ursache := range alle {
		u := neueUmg(t)
		injiziert := &model.KompatFehler{Ursache: ursache, Ort: "ort-" + kanarienwert}
		u.c.url.f = func(string, string) (model.ODataURL, error) { return model.ODataURL{}, injiziert }
		_, err := u.rufe(metadatenAnfrage())
		f := erwarteFehler(t, string(ursache), err, model.CodeE6006)
		var kf *model.KompatFehler
		if f != nil && (!errors.As(f, &kf) || kf != injiziert || strings.Contains(f.Error(), kanarienwert)) {
			t.Errorf("%s: Ursache %v", ursache, f.Ursache)
		}
	}
	rest := []struct {
		name string
		err  error
	}{
		{"ErrUngueltig", model.ErrUngueltig},
		{"ErrUngueltig mit Text", errors.Join(model.ErrUngueltig, errors.New("form "+kanarienwert))},
		{"anderer Fehler", errors.New("roh " + kanarienwert)},
	}
	for _, r := range rest {
		u := neueUmg(t)
		u.c.url.f = func(string, string) (model.ODataURL, error) { return model.ODataURL{}, r.err }
		_, err := u.rufe(metadatenAnfrage())
		f := erwarteFehler(t, r.name, err, model.CodeE9001)
		if f != nil && !errors.Is(f, r.err) {
			t.Errorf("%s: Ursache nicht erreichbar", r.name)
		}
	}
	gueltig := []model.RessourcenArt{model.RessourceMetadaten, model.RessourceMenge, model.RessourceEntity, model.RessourceAnzahl}
	for _, art := range []model.RessourcenArt{0, 5, 6, 99, -1} {
		if slices.Contains(gueltig, art) {
			t.Fatalf("Art %d ist gültig", art)
		}
		u := neueUmg(t)
		u.c.url.f = func(string, string) (model.ODataURL, error) {
			return model.ODataURL{Ressource: model.Ressource{Art: art}}, nil
		}
		_, err := u.rufe(metadatenAnfrage())
		if f := erwarteFehler(t, "Art", err, model.CodeE9001); f != nil && f.Ursache != nil {
			t.Errorf("Art %d: Ursache %v", art, f.Ursache)
		}
		if ops := u.p.ops(); !reflect.DeepEqual(ops, []string{"URL.ParseURL"}) {
			t.Errorf("Art %d: Aufrufe %v", art, ops)
		}
	}
}
