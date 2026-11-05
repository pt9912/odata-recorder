package urlparserv2_test

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparserv2"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestGeneratorURLs (Zeile 48): der Generator bildet URLs aus einer Pfadform (40) und einer Optionsfolge (0 bis 4 aus 30 Formen,
// mit gewollten Wiederholungen); der Leser klassifiziert jede URL genauso wie das Orakel des Tests (orakel_test.go, ohne Import aus
// dem Produktcode) und liefert, wo sie gültig ist, das Ergebnis, das die Formen ergeben. Menge: 2000 Läufe, Saat 1 bis 2000;
// Zählprüfung: jede der 15 Ursachen mindestens 20 Mal, Erfolge mindestens 300.
func TestGeneratorURLs(t *testing.T) {
	pGut, pSchlecht := pfadFormen()
	qGut, qSchlecht := optionsFormen()
	if len(pGut)+len(pSchlecht) != 40 || len(qGut)+len(qSchlecht) != 30 {
		t.Fatalf("%d Pfadformen und %d Optionsformen, erwartet 40 und 30", len(pGut)+len(pSchlecht), len(qGut)+len(qSchlecht))
	}
	var qKompat []qForm
	for _, f := range qSchlecht {
		if f.v != nil && !f.v.klasseU {
			qKompat = append(qKompat, f)
		}
	}
	p := urlparserv2.New()
	zaehler := map[string]int{}
	for saat := uint64(1); saat <= 2000; saat++ {
		r := rand.New(rand.NewPCG(saat, saat*7919+13))
		wahlP := func() pForm {
			if r.IntN(100) < 60 {
				return pGut[r.IntN(len(pGut))]
			}
			return pSchlecht[r.IntN(len(pSchlecht))]
		}
		pf := wahlP()
		kompatLauf := r.IntN(100) < 15 // ein Lauf aus Kompatibilitätsoptionen bei gültigem Pfad
		if kompatLauf {
			pf = pGut[r.IntN(len(pGut))]
		}
		var qs []qForm
		var texte []string
		for range r.IntN(5) {
			var f qForm
			switch {
			case kompatLauf:
				f = qKompat[r.IntN(len(qKompat))]
			case len(qs) > 0 && r.IntN(100) < 8:
				f = qs[r.IntN(len(qs))]
			case r.IntN(100) < 60:
				f = qGut[r.IntN(len(qGut))]
			default:
				f = qSchlecht[r.IntN(len(qSchlecht))]
			}
			qs = append(qs, f)
			texte = append(texte, f.text)
		}
		q := strings.Join(texte, "&")
		soll, etikett := orakel(pf, qs)
		zaehler[etikett]++
		t.Run("", func(t *testing.T) {
			t.Helper()
			t.Logf("Saat %d: ParseURL(%.60q, %.60q)", saat, pf.text, q)
			pruefe(t, p, pf.text, q, soll)
		})
	}
	ursachen := 0
	for _, u := range model.AlleURLUrsachen() {
		if !u.NichtUnterstuetzt() {
			ursachen++
			if zaehler["u:"+string(u)] < 20 {
				t.Errorf("Ursache %s nur %d-mal getroffen, erwartet mindestens 20", u, zaehler["u:"+string(u)])
			}
		}
	}
	for _, k := range model.AlleKompatUrsachen() {
		if k != model.KompatMethode {
			ursachen++
			if zaehler["k:"+string(k)] < 20 {
				t.Errorf("Ursache %s nur %d-mal getroffen, erwartet mindestens 20", k, zaehler["k:"+string(k)])
			}
		}
	}
	if ursachen != 15 || zaehler["ok"] < 300 {
		t.Errorf("%d Ursachen, %d Erfolge; erwartet 15 und mindestens 300", ursachen, zaehler["ok"])
	}
	t.Logf("Verteilung %v", zaehler)
}
