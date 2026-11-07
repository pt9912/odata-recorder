package mock

import (
	"bytes"
	"context"
	"errors"
	"io"
	"slices"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// headerCoverage ist der Header der Coverage auf Collection und Anzahl (SPEC-197 Nr. 4).
const headerCoverage = "X-Recorder-Coverage"

// lauf trägt, was die Schritte nach dem Lesen des Modells brauchen: nur Werte (SPEC-195 Nr. 12).
type lauf struct {
	ctx    context.Context
	kopf   bool
	d      Dienst
	w      model.Workspace
	u      model.ODataURL
	modell model.ServiceModel
	set    string
	typ    model.EntityType
	accept []string
}

// lese beantwortet Collection, Entity und Anzahl (SPEC-196, SPEC-197, SPEC-198): das Modell, das Entity-Set, dann der
// Ablauf der Art. Das Modell lebt nur für den Aufruf.
func (m *Mocker) lese(ctx context.Context, a inbound.Anfrage, d Dienst, w model.Workspace, s model.Service, u model.ODataURL) (inbound.Antwort, error) {
	modell, err := m.ladeModell(ctx, d, s)
	if err != nil {
		return inbound.Antwort{}, err
	}
	set, ok := modell.EntitySet(u.Ressource.EntitySet)
	if !ok {
		return inbound.Antwort{}, mitCode(model.CodeE6008, nil)
	}
	typ, ok := modell.EntityType(set.EntityType)
	if !ok {
		return inbound.Antwort{}, mitCode(model.CodeE9001, nil)
	}
	l := lauf{ctx: ctx, kopf: a.Methode == "HEAD", d: d, w: w, u: u, modell: modell, set: set.Name, typ: typ, accept: werteVon(a.Header, "accept")}
	switch u.Ressource.Art {
	case model.RessourceMenge:
		return m.collection(l)
	case model.RessourceAnzahl:
		return m.anzahl(l)
	}
	return m.entity(l)
}

// ladeModell öffnet den Blob, liest das Modell und schließt den Strom genau einmal (SPEC-196 Nr. 2).
func (m *Mocker) ladeModell(ctx context.Context, d Dienst, s model.Service) (model.ServiceModel, error) {
	strom, err := m.k.Blobs.Open(ctx, s.MetadataBlob)
	if err != nil {
		return model.ServiceModel{}, abbilden(err, outbound.GegenstandSonstiges)
	}
	modell, _, err := d.Codecs.Metadaten.ParseMetadata(ctx, strom, m.k.MaxBytes)
	_ = strom.Close()
	if err != nil {
		return model.ServiceModel{}, abbilden(err, outbound.GegenstandSonstiges)
	}
	return modell, nil
}

// kontextFehler sagt, ob err einen Kontextfehler enthält (SPEC-109 Nr. 6).
func kontextFehler(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// nichtGefunden sagt, ob err die fehlende Zeile des Speichers ist; ein Kontextfehler daneben gewinnt.
func nichtGefunden(err error) bool {
	return errors.Is(err, outbound.ErrNichtGefunden) && !kontextFehler(err)
}

// coverage liest den Zustand des Sets an der Baseline des Workspace (SPEC-197 Nr. 1); kein Eintrag ist `unknown`.
func (m *Mocker) coverage(l lauf) (model.CoverageStatus, error) {
	c, err := m.k.Coverage.Get(l.ctx, l.w.BaselineSnapshotID, l.set)
	switch {
	case nichtGefunden(err):
		return model.CoverageUnknown, nil
	case err != nil:
		return "", abbilden(err, outbound.GegenstandSonstiges)
	case !slices.Contains(model.AlleCoverageStatus(), c.Status):
		return "", mitCode(model.CodeE9001, nil)
	}
	return c.Status, nil
}

// vorErgebnis ist der gemeinsame Anfang von Collection und Anzahl: Coverage vor der Ausführung (SPEC-197 Nr. 2),
// ListSet genau einmal, Deuten und die Pflichten 1, 3 e und 3 f (SPEC-196 Nr. 3, 4).
func (m *Mocker) vorErgebnis(l lauf) (model.CoverageStatus, []model.Entity, func(string) bool, error) {
	status, err := m.coverage(l)
	if err != nil {
		return "", nil, nil, err
	}
	if status != model.CoverageComplete && m.k.StrictCoverage {
		return "", nil, nil, mitCode(model.CodeE6003, nil)
	}
	liste, err := m.k.Entities.ListSet(l.ctx, l.w.ID, l.set)
	if err != nil {
		return "", nil, nil, abbilden(err, outbound.GegenstandWorkspace)
	}
	entities, err := deuteListe(l, liste)
	if err != nil {
		return "", nil, nil, err
	}
	return status, entities, maske(liste, status), nil
}

// collection beantwortet die Collection (SPEC-196 Nr. 5).
func (m *Mocker) collection(l lauf) (inbound.Antwort, error) {
	status, entities, beobachtet, err := m.vorErgebnis(l)
	if err != nil {
		return inbound.Antwort{}, err
	}
	o := l.u.Optionen
	opt := model.Optionen{Filter: o.Filter, OrderBy: o.OrderBy, Top: o.Top, Skip: o.Skip, Select: o.Select}
	erg, err := fuehreAus(l.typ, entities, opt, beobachtet)
	if err != nil {
		return inbound.Antwort{}, err
	}
	var buf bytes.Buffer
	dar, err := l.d.Codecs.Schreiber.SchreibeCollection(l.ctx, &buf, l.modell, model.CollectionAntwort{
		Wurzel:   m.wurzel(l.d),
		Set:      l.set,
		Entities: erg.Entities,
		Anzahl:   model.Anzahl{Vorhanden: o.Count.Gesetzt && o.Count.Wert, Wert: int64(erg.Anzahl)},
		Auswahl:  auswahl(o.Select),
		Accept:   l.accept,
	})
	if err != nil {
		return inbound.Antwort{}, abbilden(err, outbound.GegenstandSonstiges)
	}
	return liefere(l, dar, buf.Bytes(), string(status))
}

// anzahl beantwortet `/$count`: wie die Collection, der Interpreter erhält allein den Filter (SPEC-196 Nr. 6).
func (m *Mocker) anzahl(l lauf) (inbound.Antwort, error) {
	status, entities, beobachtet, err := m.vorErgebnis(l)
	if err != nil {
		return inbound.Antwort{}, err
	}
	erg, err := fuehreAus(l.typ, entities, model.Optionen{Filter: l.u.Optionen.Filter}, beobachtet)
	if err != nil {
		return inbound.Antwort{}, err
	}
	var buf bytes.Buffer
	dar, err := l.d.Codecs.Schreiber.SchreibeAnzahl(l.ctx, &buf, int64(erg.Anzahl))
	if err != nil {
		return inbound.Antwort{}, abbilden(err, outbound.GegenstandSonstiges)
	}
	return liefere(l, dar, buf.Bytes(), string(status))
}

// entity beantwortet den Einzel-Lookup (SPEC-196 Nr. 7).
func (m *Mocker) entity(l lauf) (inbound.Antwort, error) {
	schluessel, ok := bildeSchluessel(l.typ, l.u.Ressource.Schluessel)
	if !ok {
		return inbound.Antwort{}, mitCode(model.CodeE6005, nil)
	}
	if l.u.Optionen.Filter != nil {
		return inbound.Antwort{}, mitCode(model.CodeE6001, nil)
	}
	es, err := m.k.Entities.Get(l.ctx, l.w.ID, l.set, schluessel)
	switch {
	case nichtGefunden(err):
		return m.ohneTreffer(l)
	case err != nil:
		return inbound.Antwort{}, abbilden(err, outbound.GegenstandWorkspace)
	case es.Deleted:
		return inbound.Antwort{}, mitCode(model.CodeE6008, nil)
	}
	entities, err := deuteListe(l, []model.EntityState{es})
	if err != nil {
		return inbound.Antwort{}, err
	}
	sel := l.u.Optionen.Select
	erg, err := fuehreAus(l.typ, entities, model.Optionen{Select: sel}, nil)
	if err != nil {
		return inbound.Antwort{}, err
	}
	if len(erg.Entities) != 1 {
		return inbound.Antwort{}, mitCode(model.CodeE9001, nil)
	}
	if !vollstaendig(l.modell, erg.Entities[0], es, sel) {
		return inbound.Antwort{}, mitCode(model.CodeE6003, nil)
	}
	var buf bytes.Buffer
	dar, err := l.d.Codecs.Schreiber.SchreibeEntity(l.ctx, &buf, l.modell, model.EntityAntwort{
		Wurzel:  m.wurzel(l.d),
		Set:     l.set,
		Entity:  erg.Entities[0],
		Auswahl: auswahl(sel),
		Accept:  l.accept,
	})
	if err != nil {
		return inbound.Antwort{}, abbilden(err, outbound.GegenstandSonstiges)
	}
	return liefere(l, dar, buf.Bytes(), "")
}

// ohneTreffer entscheidet den Lookup ohne Treffer nach der Coverage (SPEC-197 Nr. 3), in beiden Modi gleich.
func (m *Mocker) ohneTreffer(l lauf) (inbound.Antwort, error) {
	status, err := m.coverage(l)
	switch {
	case err != nil:
		return inbound.Antwort{}, err
	case status == model.CoverageComplete:
		return inbound.Antwort{}, mitCode(model.CodeE6008, nil)
	}
	return inbound.Antwort{}, mitCode(model.CodeE6003, nil)
}

// vollstaendig prüft, ob jede angeforderte Property in KnownProperties des Eintrags steht (SPEC-196 Nr. 7 e).
func vollstaendig(modell model.ServiceModel, e model.Entity, es model.EntityState, sel *model.SelectOption) bool {
	var gefragt []string
	if sel == nil || sel.Alle {
		t, _ := modell.EntityType(e.Typ)
		for _, p := range t.Properties {
			if p.Unterstuetzt() {
				gefragt = append(gefragt, p.Name)
			}
		}
	} else {
		gefragt = sel.Namen
	}
	return !slices.ContainsFunc(gefragt, func(n string) bool { return !slices.Contains(es.KnownProperties, n) })
}

// fuehreAus ruft den Interpreter mit den Pflichten des Aufrufers (SPEC-196 Nr. 4, 8): die Eingabe ist auf Key-Properties,
// Schlüssel und `redigiert` in den Optionen geprüft; die Werte `redigiert` entfernt es aus dem Ergebnis.
func fuehreAus(typ model.EntityType, entities []model.Entity, opt model.Optionen, beobachtet func(string) bool) (model.Ergebnis, error) {
	if !redigiertFrei(entities, opt) {
		return model.Ergebnis{}, mitCode(model.CodeE9001, nil)
	}
	erg, err := model.FuehreAus(typ, entities, opt, beobachtet)
	if err != nil {
		return model.Ergebnis{}, abbilden(err, outbound.GegenstandSonstiges)
	}
	erg.Entities = ohneRedigiert(erg.Entities)
	return erg, nil
}

// liefere prüft die Darstellung des Schreibers (SPEC-198 Nr. 8) und bildet die Antwort; bei HEAD ist der Body leer.
func liefere(l lauf, dar model.Darstellung, body []byte, coverage string) (inbound.Antwort, error) {
	if dar.ContentType == "" || dar.Kennzeichen.Name == "" || dar.Kennzeichen.Wert == "" {
		return inbound.Antwort{}, mitCode(model.CodeE9001, nil)
	}
	h := map[string][]string{"Content-Type": {dar.ContentType}, dar.Kennzeichen.Name: {dar.Kennzeichen.Wert}}
	if coverage != "" {
		h[headerCoverage] = []string{coverage}
	}
	if l.kopf {
		body = nil
	}
	return inbound.Antwort{Status: 200, Header: h, Body: io.NopCloser(bytes.NewReader(body))}, nil
}

// wurzel ist PublicBaseURL und der Root ohne abschließenden `/`, nie ein Header der Anfrage (SPEC-198 Nr. 6).
func (m *Mocker) wurzel(d Dienst) string {
	return m.k.PublicBaseURL + strings.TrimSuffix(d.Root, "/")
}

// auswahl ist `Auswahl` des Schreibers: nil ohne `$select`, sonst `*` (wenn Alle) und die Namen der URL (SPEC-198 Nr. 3).
func auswahl(s *model.SelectOption) []string {
	if s == nil {
		return nil
	}
	var r []string
	if s.Alle {
		r = append(r, "*")
	}
	return append(r, s.Namen...)
}
