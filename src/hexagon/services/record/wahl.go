package record

import (
	"context"
	"errors"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// wahl ist das Ergebnis der Prüfungen aus SPEC-121 Nr. 2b bis 2d: der Dienst, der OData-Pfad,
// die redigierte Darstellung ohne Body und die offenen Sessions, unter denen die Annahme wählt.
type wahl struct {
	dienst     Dienst
	pfad       string
	vorab      model.Schluessel
	kandidaten []string
}

func (r *Recorder) pruefe(ctx context.Context, a inbound.Anfrage) (wahl, error) {
	d, pfad, ok := dienstFuer(r.k.Dienste, a.Pfad)
	if !ok {
		return wahl{}, mitCode(model.CodeE5005, nil)
	}
	vorab, err := model.BildeSchluessel(model.SchluesselEingabe{
		Method: a.Methode, ServiceID: d.ID, Path: pfad, RawQuery: a.RawQuery, Header: a.Header,
		RequestRedaction: model.RedactionNone,
	}, r.k.Regeln)
	switch {
	case errors.Is(err, model.ErrKodierung):
		return wahl{}, mitCode(model.CodeE5006, err)
	case err != nil:
		return wahl{}, mitCode(model.CodeE9001, err)
	}
	kandidaten, err := r.leseSession(ctx, d, a.Header)
	if err != nil {
		return wahl{}, err
	}
	return wahl{dienst: d, pfad: pfad, vorab: vorab, kandidaten: kandidaten}, nil
}

// dienstFuer wählt den Dienst, dessen Root gleich pfad ist oder ihm an einer Segmentgrenze
// vorausgeht, byteweise, den mit dem längsten Root, und liefert den OData-Pfad (SPEC-121 Nr. 2b).
func dienstFuer(dienste []Dienst, pfad string) (Dienst, string, bool) {
	var (
		gewaehlt Dienst
		rest     string
		gefunden bool
	)
	for _, d := range dienste {
		r, ok := unterRoot(d.Root, pfad)
		if ok && (!gefunden || len(d.Root) > len(gewaehlt.Root)) {
			gewaehlt, rest, gefunden = d, r, true
		}
	}
	if !gefunden {
		return Dienst{}, "", false
	}
	if rest == "" {
		rest = "/"
	}
	return gewaehlt, rest, true
}

func unterRoot(root, pfad string) (string, bool) {
	switch {
	case root == "/":
		return pfad, strings.HasPrefix(pfad, "/")
	case pfad == root:
		return "", true
	case strings.HasPrefix(pfad, root+"/"):
		return pfad[len(root):], true
	}
	return "", false
}

// leseSession liefert die offenen Sessions, unter denen die Annahme wählt (SPEC-121 Nr. 2d): mit
// dem Header die eine benannte, sonst die offenen des Service. Ein Kontextfehler kehrt unverändert
// zurück, jeder andere Fehler der Ablage als Code mit dem Gegenstand Session.
func (r *Recorder) leseSession(ctx context.Context, d Dienst, h map[string][]string) ([]string, error) {
	werte := werteVon(h, "x-recorder-recording-session")
	if len(werte) == 0 {
		liste, err := r.k.Sessions.List(ctx, d.ID)
		if err != nil {
			return nil, sessionFehler(err)
		}
		var offen []string
		for _, s := range liste {
			if s.Status == model.SessionOpen {
				offen = append(offen, s.ID)
			}
		}
		return offen, nil
	}
	if len(werte) > 1 || werte[0] == "" {
		return nil, mitCode(model.CodeE5003, nil)
	}
	s, err := r.k.Sessions.Get(ctx, werte[0])
	switch {
	case err != nil:
		return nil, sessionFehler(err)
	case s.ServiceID != d.ID:
		return nil, mitCode(model.CodeE5004, nil)
	case s.Status != model.SessionOpen:
		return nil, mitCode(model.CodeE5009, nil)
	}
	return []string{werte[0]}, nil
}

func sessionFehler(err error) error {
	c, ok := outbound.Abbilden(err, outbound.GegenstandSession)
	if !ok {
		return err
	}
	return mitCode(c, err)
}

// nimmAn prüft unter der Sperre den Drain und nimmt den Exchange an (SPEC-121 Nr. 2d, 2e): von
// den Kandidaten zählen die, die nicht im Drain sind; keine ist ODR-E5009, mehrere ODR-E5003.
func (r *Recorder) nimmAn(kandidaten []string, l *lauf) (string, *sessionZustand, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var frei []string
	for _, id := range kandidaten {
		if z := r.sessions[id]; z == nil || !z.drain {
			frei = append(frei, id)
		}
	}
	switch {
	case len(frei) == 0:
		return "", nil, mitCode(model.CodeE5009, nil)
	case len(frei) > 1:
		return "", nil, mitCode(model.CodeE5003, nil)
	}
	sz := r.zustand(frei[0])
	if len(sz.laufend) == 0 {
		sz.leer = make(chan struct{})
	}
	sz.laufend[l] = struct{}{}
	return frei[0], sz, nil
}
