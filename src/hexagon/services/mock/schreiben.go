package mock

import (
	"bytes"
	"context"
	"io"
	"slices"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Die vier Methoden, die der Dienst ausführt (SPEC-192); jede andere Schreibweise und jede andere Methode ist keine Mutation.
const (
	methodePost   = "POST"
	methodePatch  = "PATCH"
	methodePut    = "PUT"
	methodeDelete = "DELETE"
)

// istMutation sagt, ob die Methode eine der vier Mutationen ist (SPEC-192 Nr. 2); die Schreibweise zählt.
func istMutation(methode string) bool {
	return methode == methodePost || methode == methodePatch || methode == methodePut || methode == methodeDelete
}

// mutation trägt, was die Schritte nach dem Lesen des Modells brauchen: nur Werte (SPEC-195 Nr. 12). schluessel ist der
// Schlüssel des Eintrags (SPEC-193 Nr. 1), eingabe der gelesene Body (leer bei DELETE).
type mutation struct {
	lauf
	methode    string
	schluessel string
	eingabe    model.EntityEingabe
}

// schreibe führt POST, PATCH, PUT und DELETE aus (SPEC-192 Nr. 2): Methode und Ressource, Workspace und Service, Modell und Set,
// Schlüssel der URL, Body, dann die Transaktion. Bis zur Transaktion wird nichts geschrieben.
func (m *Mocker) schreibe(ctx context.Context, a inbound.Anfrage, d Dienst, wsID string, u model.ODataURL) (inbound.Antwort, error) {
	if err := pruefeMethodeUndRessource(a.Methode, d, u); err != nil {
		return inbound.Antwort{}, err
	}
	w, err := m.holeWorkspace(ctx, d, wsID)
	if err != nil {
		return inbound.Antwort{}, err
	}
	s, err := m.holeService(ctx, d)
	if err != nil {
		return inbound.Antwort{}, err
	}
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
	mu := mutation{
		lauf:    lauf{d: d, w: w, u: u, modell: modell, set: set.Name, typ: typ, accept: werteVon(a.Header, "accept")},
		methode: a.Methode,
	}
	if a.Methode != methodePost {
		schluessel, ok := bildeSchluessel(typ, u.Ressource.Schluessel)
		if !ok {
			return inbound.Antwort{}, mitCode(model.CodeE6005, nil)
		}
		mu.schluessel = schluessel
	}
	if a.Methode != methodeDelete {
		if mu, err = m.liesBody(ctx, a, mu); err != nil {
			return inbound.Antwort{}, err
		}
	}
	return m.transaktion(ctx, mu)
}

// pruefeMethodeUndRessource ist Schritt (e) (SPEC-192 Nr. 3): ein Dienst, dessen Satz nur liest, die Paarung, ein $filter und ein
// Satz ohne Leser für Client-Bodies; kein Port wird gerufen.
func pruefeMethodeUndRessource(methode string, d Dienst, u model.ODataURL) error {
	if d.Codecs.NurLesend {
		return mitCode(model.CodeE6006, &model.KompatFehler{Ursache: model.KompatMethode})
	}
	art := u.Ressource.Art
	if (methode == methodePost) != (art == model.RessourceMenge) || methode != methodePost && art != model.RessourceEntity {
		return mitCode(model.CodeE6002, nil)
	}
	if u.Optionen.Filter != nil {
		return mitCode(model.CodeE6002, nil)
	}
	if d.Codecs.Eingabe == nil && methode != methodeDelete {
		return mitCode(model.CodeE6002, nil)
	}
	return nil
}

// liesBody ruft den Leser genau einmal (SPEC-192 Nr. 5) und prüft danach den Schlüssel gegen die URL und den reservierten Text
// (SPEC-193 Nr. 1, 3 b); der Body wird weder vorher gelesen noch geschlossen.
func (m *Mocker) liesBody(ctx context.Context, a inbound.Anfrage, mu mutation) (mutation, error) {
	var src io.Reader = bytes.NewReader(nil)
	if a.Body != nil {
		src = a.Body
	}
	e, err := mu.d.Codecs.Eingabe.LeseEntity(ctx, src, werteVon(a.Header, "content-type"), mu.modell, mu.set, model.EingabeArt(a.Methode), m.k.MaxBytes)
	if err != nil {
		return mu, abbilden(err, outbound.GegenstandSonstiges)
	}
	if reserviert(e) {
		return mu, mitCode(model.CodeE6007, nil)
	}
	if a.Methode == methodePost {
		if e.Schluessel == "" {
			return mu, mitCode(model.CodeE9001, nil)
		}
		mu.schluessel = e.Schluessel
	} else if widerspricht(mu.typ, mu.u.Ressource.Schluessel, e) {
		return mu, mitCode(model.CodeE6007, nil)
	}
	mu.eingabe = e
	return mu, nil
}

// reserviert sagt, ob eine String-Property des Bodys den Text `[REDACTED]` trägt (SPEC-193 Nr. 3 b).
func reserviert(e model.EntityEingabe) bool {
	return slices.ContainsFunc(e.Properties, func(p model.PropertyWert) bool {
		return p.Wert.Art == model.ArtString && p.Wert.Text == model.Marker
	})
}

// widerspricht sagt, ob eine Schlüssel-Property des Bodys einen anderen Text hat als der Teil der URL (SPEC-193 Nr. 1): der Text
// ist die Literalform von `Entity.Schluessel`, byte-gleich verglichen; eine Schlüssel-Property, die der Body nicht nennt,
// widerspricht nicht.
func widerspricht(typ model.EntityType, teile []model.SchluesselTeil, e model.EntityEingabe) bool {
	for _, t := range teile {
		name := t.Name
		if name == "" && len(typ.Keys) == 1 {
			name = typ.Keys[0]
		}
		i := slices.IndexFunc(e.Properties, func(p model.PropertyWert) bool { return p.Name == name })
		if i >= 0 && literalForm(typ, name, e.Properties[i].Wert) != t.Literal.Text {
			return true
		}
	}
	return false
}

// literalForm ist das Literal der URL zu einem Wert: bei `Edm.String` der Text in einfachen Anführungszeichen mit verdoppelten
// einfachen Anführungszeichen, sonst der Text (SPEC-126 Nr. 11).
func literalForm(typ model.EntityType, name string, w model.Wert) string {
	i := slices.IndexFunc(typ.Properties, func(p model.Property) bool { return p.Name == name })
	if i >= 0 && typ.Properties[i].EdmType == "Edm.String" && w.Art == model.ArtString {
		return "'" + strings.ReplaceAll(w.Text, "'", "''") + "'"
	}
	return w.Text
}

// transaktion führt (k) und (l) in einer Workspace-Transaktion aus (SPEC-192 Nr. 6, 7): ein Fehler von f kehrt unverändert zurück,
// jeder andere Fehler der Transaktion ist `Sonstiges`; bei einem Fehler gibt es keine Antwort.
func (m *Mocker) transaktion(ctx context.Context, mu mutation) (inbound.Antwort, error) {
	var antwort inbound.Antwort
	err := m.k.Transaktion.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
		var err error
		antwort, err = m.imWorkspace(ctx, s, mu)
		return err
	})
	if err == nil {
		return antwort, nil
	}
	if _, ok := err.(*inbound.Fehler); !ok && !kontextFehler(err) {
		err = abbilden(err, outbound.GegenstandSonstiges)
	}
	return inbound.Antwort{}, err
}

// imWorkspace ist f: Workspace lesen, Eintrag lesen, entscheiden, Put, Advance, Antwort (SPEC-192 Nr. 6).
func (m *Mocker) imWorkspace(ctx context.Context, s outbound.WorkspaceStores, mu mutation) (inbound.Antwort, error) {
	w, err := s.Workspaces.Get(ctx, mu.w.ID)
	if err != nil {
		return inbound.Antwort{}, abbilden(err, outbound.GegenstandWorkspace)
	}
	es, err := s.Entities.Get(ctx, mu.w.ID, mu.set, mu.schluessel)
	fehlend := nichtGefunden(err)
	if err != nil && !fehlend {
		return inbound.Antwort{}, abbilden(err, outbound.GegenstandWorkspace)
	}
	if fehlend {
		es = model.EntityState{}
	}
	neu, err := m.entscheide(ctx, mu, es, fehlend)
	if err != nil {
		return inbound.Antwort{}, err
	}
	if err := s.Entities.Put(ctx, neu); err != nil {
		return inbound.Antwort{}, abbilden(err, outbound.GegenstandWorkspace)
	}
	if _, err := s.Workspaces.Advance(ctx, mu.w.ID, w.Revision); err != nil {
		return inbound.Antwort{}, abbilden(err, outbound.GegenstandWorkspace)
	}
	return m.antwort(ctx, mu, neu)
}

// entscheide wendet die Tabelle von SPEC-193 Nr. 2 an und liefert den Eintrag, den `Put` schreibt. Die Coverage liest sie nur
// bei einem fehlenden Eintrag von PATCH, PUT und DELETE.
func (m *Mocker) entscheide(ctx context.Context, mu mutation, es model.EntityState, fehlend bool) (model.EntityState, error) {
	if mu.methode == methodePost {
		if !fehlend && !es.Deleted {
			return model.EntityState{}, mitCode(model.CodeE6009, nil)
		}
		return legeAn(mu, es.Version+1)
	}
	if !fehlend && es.Deleted {
		return model.EntityState{}, mitCode(model.CodeE6008, nil)
	}
	if fehlend {
		status, err := m.coverage(ctx, mu.lauf)
		switch {
		case err != nil:
			return model.EntityState{}, err
		case status == model.CoverageComplete:
			return model.EntityState{}, mitCode(model.CodeE6008, nil)
		}
		return model.EntityState{}, mitCode(model.CodeE6003, nil)
	}
	es.Version++
	switch mu.methode {
	case methodeDelete:
		es.Deleted = true
		return es, nil
	case methodePatch:
		if es.Completeness != model.EntityComplete {
			return model.EntityState{}, mitCode(model.CodeE6003, nil)
		}
	}
	mitglieder, typ, ok := liesEintrag(es.Payload, mu.typ.Name)
	switch {
	case !ok:
		return model.EntityState{}, mitCode(model.CodeE9001, nil)
	case typ != mu.eingabe.Typ:
		return model.EntityState{}, mitCode(model.CodeE6007, nil)
	}
	if mu.methode == methodePut {
		return legeAn(mu, es.Version)
	}
	return aendere(mu, es, mitglieder)
}

// antwort bildet die Antwort in der Transaktion (SPEC-194): `201` mit Entity und Location bei POST, sonst `204` mit dem
// Kennzeichen des Satzes; ein leeres Kennzeichen oder ein leerer Content-Type ist ODR-E9001. Bei einem Dienst mit Etag
// trägt die Antwort außer bei DELETE den ETag des Zählers von neu, des Eintrags, den die Transaktion mit Put schrieb
// (SPEC-281 Nr. 2); sie liest den Eintrag nicht erneut.
func (m *Mocker) antwort(ctx context.Context, mu mutation, neu model.EntityState) (inbound.Antwort, error) {
	sch := mu.d.Codecs.Schreiber
	var buf bytes.Buffer
	if mu.methode != methodePost {
		dar, err := sch.SchreibeAnzahl(ctx, io.Discard, 0)
		if err != nil {
			return inbound.Antwort{}, abbilden(err, outbound.GegenstandSonstiges)
		}
		if darstellungLeer(dar) {
			return inbound.Antwort{}, mitCode(model.CodeE9001, nil)
		}
		return inbound.Antwort{Status: 204, Header: kopf204(mu, dar, neu), Body: io.NopCloser(bytes.NewReader(nil))}, nil
	}
	e, ok := deute(mu.modell, mu.typ, neu)
	if !ok {
		return inbound.Antwort{}, mitCode(model.CodeE9001, nil)
	}
	e.Etag = etagVon(mu.d, neu)
	wurzel := m.wurzel(mu.d)
	dar, err := sch.SchreibeEntity(ctx, &buf, mu.modell, model.EntityAntwort{Wurzel: wurzel, Set: mu.set, Entity: e, Accept: mu.accept})
	if err != nil {
		return inbound.Antwort{}, abbilden(err, outbound.GegenstandSonstiges)
	}
	if darstellungLeer(dar) {
		return inbound.Antwort{}, mitCode(model.CodeE9001, nil)
	}
	ort, err := sch.Adresse(mu.modell, wurzel, mu.set, neu.EntityKey)
	if err != nil {
		return inbound.Antwort{}, abbilden(err, outbound.GegenstandSonstiges)
	}
	h := map[string][]string{"Content-Type": {dar.ContentType}, dar.Kennzeichen.Name: {dar.Kennzeichen.Wert}, "Location": {ort}}
	if e.Etag != "" {
		h[headerEtag] = []string{e.Etag}
	}
	return inbound.Antwort{Status: 201, Header: h, Body: io.NopCloser(bytes.NewReader(buf.Bytes()))}, nil
}

// kopf204 ist der Header einer 204: das Kennzeichen der Darstellung und, bei PATCH und PUT eines Dienstes mit Etag, der ETag von neu
// (SPEC-281 Nr. 2); DELETE trägt keinen.
func kopf204(mu mutation, dar model.Darstellung, neu model.EntityState) map[string][]string {
	h := map[string][]string{dar.Kennzeichen.Name: {dar.Kennzeichen.Wert}}
	if etag := etagVon(mu.d, neu); etag != "" && mu.methode != methodeDelete {
		h[headerEtag] = []string{etag}
	}
	return h
}

// darstellungLeer sagt, ob die Darstellung des Schreibers unbrauchbar ist (SPEC-198 Nr. 8).
func darstellungLeer(dar model.Darstellung) bool {
	return dar.ContentType == "" || dar.Kennzeichen.Name == "" || dar.Kennzeichen.Wert == ""
}
