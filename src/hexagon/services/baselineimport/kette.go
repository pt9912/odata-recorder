package baselineimport

import (
	"slices"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Enden einer Kette (SPEC-127 Nr. 12 b, 13); endeBelegt steht bei der Coverage.
const (
	endeSeiteNichtLesbar    = "seite_nicht_lesbar"
	endeKeyRedigiert        = "key_redacted"
	endeLinkRedigiert       = "link_redigiert"
	endeLinkNichtAufloesbar = "link_nicht_aufloesbar"
	endeSchleife            = "schleife"
	endeSeiteFehlt          = "seite_fehlt"
	endeAnzahlAbweichend    = "anzahl_abweichend"
)

// kettenlauf trägt, was der Kettenlauf über die Session wissen muss: die Exchanges nach sequence_no,
// die gelesenen Antworten nach Kennung, den Dienst (Upstream) und die Regeln des Match-Keys.
type kettenlauf struct {
	exchanges []model.Exchange
	gelesen   map[string]*antwort
	dienst    Dienst
	regeln    model.SchluesselRegeln
}

// laufeKetten führt für jede ungefilterte Abfrage des Sets eine Kette (Nr. 12 f), in der Folge der
// sequence_no ihrer ersten Seite.
func (k kettenlauf) laufeKetten(set string, antworten []*antwort) []kette {
	var out []kette
	for _, a := range antworten {
		if a.ungefiltert {
			out = append(out, k.laufe(set, a))
		}
	}
	return out
}

// laufe geht von der ersten Seite Seite für Seite in der Reihenfolge der Prüfungen (1) bis (7)
// von Nr. 12 b bis zum Ende der Kette.
func (k kettenlauf) laufe(set string, start *antwort) kette {
	ke := kette{start: start, schluessel: map[string]struct{}{}}
	var keys []string
	a := start
	for ke.ende == "" {
		switch {
		case a == nil || a.spur != spurCollection || a.set != set:
			ke.ende = endeSeiteNichtLesbar
		case a.ex.KeyRedacted:
			ke.ende = endeKeyRedigiert
		default:
			ke.seiten = append(ke.seiten, seitenBeleg{elemente: a.elemente, exchangeID: a.ex.ID, seq: a.ex.SequenceNo})
			for s := range a.schluessel {
				ke.schluessel[s] = struct{}{}
			}
			keys = append(keys, a.ex.MatchKey)
			a, ke.ende = k.weiter(a, keys)
		}
	}
	if start.seite.Anzahl.Vorhanden {
		ke.count = &anzahlBeleg{deklariert: start.seite.Anzahl.Wert, beobachtet: int64(len(ke.schluessel))}
	}
	if ke.ende == endeBelegt && ke.count != nil && ke.count.deklariert != ke.count.beobachtet {
		ke.ende = endeAnzahlAbweichend
	}
	return ke
}

// weiter prüft den Link der Seite a (3) bis (7): es liefert die nächste Seite oder das Ende; die
// nächste Seite ist nil, wenn der Exchange kein gelesener Exchange der Spur Collection ist.
func (k kettenlauf) weiter(a *antwort, keys []string) (*antwort, string) {
	switch l := a.seite.FolgeLink; {
	case l.Redigiert:
		return nil, endeLinkRedigiert
	case !l.Vorhanden:
		return nil, endeBelegt
	}
	key, ok := schluesselDesLinks(k.dienst, k.regeln, a)
	if !ok {
		return nil, endeLinkNichtAufloesbar
	}
	if slices.Contains(keys, key) {
		return nil, endeSchleife
	}
	i := slices.IndexFunc(k.exchanges, func(e model.Exchange) bool { return e.MatchKey == key && e.SequenceNo > a.ex.SequenceNo })
	if i < 0 {
		return nil, endeSeiteFehlt
	}
	return k.gelesen[k.exchanges[i].ID], ""
}
