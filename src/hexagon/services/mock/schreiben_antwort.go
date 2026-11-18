package mock

import (
	"bytes"
	"context"
	"io"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Kopfzeilen der Antwort einer Mutation mit Wahl von `Prefer` (SPEC-286 Nr. 1) und ihre beiden Werte.
const (
	headerLocation = "Location"
	headerEntityID = "OData-EntityId"
	headerApplied  = "Preference-Applied"
	appliedMinimal = "return=minimal"
	appliedEntity  = "return=representation"
)

// antwort bildet die Antwort in der Transaktion (SPEC-194, SPEC-286): die Wahl aus `Prefer` liest sie hier und nur bei POST, PATCH und
// PUT (SPEC-287 Nr. 1, 3). `201` oder `200` mit Entity, sonst `204`; bei einem Dienst mit Etag trägt die Antwort außer bei DELETE den
// ETag des Zählers von neu, des Eintrags, den die Transaktion mit Put schrieb (SPEC-281 Nr. 2); sie liest den Eintrag nicht erneut
// (SPEC-288 Nr. 1). Ein leeres Kennzeichen oder ein leerer Content-Type ist ODR-E9001.
func (m *Mocker) antwort(ctx context.Context, mu mutation, neu model.EntityState) (inbound.Antwort, error) {
	wahl := model.RueckgabeKeine
	if mu.methode != methodeDelete {
		wahl = model.LiesRueckgabe(werteVon(mu.kopf, "prefer"))
	}
	if mu.methode == methodePost && wahl != model.RueckgabeMinimal || mu.methode != methodePost && wahl == model.RueckgabeEntity {
		return m.antwortMitEntity(ctx, mu, neu, wahl)
	}
	return m.antwortOhneBody(ctx, mu, neu, wahl)
}

// antwortMitEntity ist `201` bei POST und `200` bei PATCH und PUT mit der Entity aus neu (SPEC-286 Nr. 1, SPEC-288 Nr. 2, 3): ein
// Aufruf von SchreibeEntity, bei POST einer von Adresse für `Location`.
func (m *Mocker) antwortMitEntity(ctx context.Context, mu mutation, neu model.EntityState, wahl model.Rueckgabe) (inbound.Antwort, error) {
	sch := mu.d.Codecs.Schreiber
	e, ok := deute(mu.modell, mu.typ, neu)
	if !ok {
		return inbound.Antwort{}, mitCode(model.CodeE9001, nil)
	}
	e.Etag = etagVon(mu.d, neu)
	wurzel := m.wurzel(mu.d)
	var buf bytes.Buffer
	dar, err := sch.SchreibeEntity(ctx, &buf, mu.modell, model.EntityAntwort{Wurzel: wurzel, Set: mu.set, Entity: e, Accept: mu.accept})
	if err != nil {
		return inbound.Antwort{}, abbilden(err, outbound.GegenstandSonstiges)
	}
	if darstellungLeer(dar) {
		return inbound.Antwort{}, mitCode(model.CodeE9001, nil)
	}
	h := map[string][]string{"Content-Type": {dar.ContentType}, dar.Kennzeichen.Name: {dar.Kennzeichen.Wert}}
	status := 200
	if mu.methode == methodePost {
		status = 201
		ort, err := m.adresse(mu, neu)
		if err != nil {
			return inbound.Antwort{}, err
		}
		h[headerLocation] = []string{ort}
	}
	if e.Etag != "" {
		h[headerEtag] = []string{e.Etag}
	}
	if wahl == model.RueckgabeEntity {
		h[headerApplied] = []string{appliedEntity}
	}
	return inbound.Antwort{Status: status, Header: h, Body: io.NopCloser(bytes.NewReader(buf.Bytes()))}, nil
}

// antwortOhneBody ist `204` mit dem Kennzeichen der Darstellung (SPEC-194 Nr. 3, SPEC-288 Nr. 3: ein Aufruf von SchreibeAnzahl mit
// verworfenem Puffer); mit der Wahl `Minimal` kommen `OData-EntityId` und bei POST `Location` aus einem Aufruf von Adresse und
// `Preference-Applied` hinzu (SPEC-286 Nr. 1).
func (m *Mocker) antwortOhneBody(ctx context.Context, mu mutation, neu model.EntityState, wahl model.Rueckgabe) (inbound.Antwort, error) {
	sch := mu.d.Codecs.Schreiber
	dar, err := sch.SchreibeAnzahl(ctx, io.Discard, 0)
	if err != nil {
		return inbound.Antwort{}, abbilden(err, outbound.GegenstandSonstiges)
	}
	if darstellungLeer(dar) {
		return inbound.Antwort{}, mitCode(model.CodeE9001, nil)
	}
	h := kopf204(mu, dar, neu)
	if wahl == model.RueckgabeMinimal {
		ort, err := m.adresse(mu, neu)
		if err != nil {
			return inbound.Antwort{}, err
		}
		h[headerEntityID] = []string{ort}
		if mu.methode == methodePost {
			h[headerLocation] = []string{ort}
		}
		h[headerApplied] = []string{appliedMinimal}
	}
	return inbound.Antwort{Status: 204, Header: h, Body: io.NopCloser(bytes.NewReader(nil))}, nil
}

// adresse ist die Aufrufstelle S25 von Adresse für `Location` und `OData-EntityId`: die Adresse des geschriebenen Eintrags aus der Wurzel des
// Dienstes, nie aus einem Header der Anfrage (SPEC-286 Nr. 1).
func (m *Mocker) adresse(mu mutation, neu model.EntityState) (string, error) {
	ort, err := mu.d.Codecs.Schreiber.Adresse(mu.modell, m.wurzel(mu.d), mu.set, neu.EntityKey)
	if err != nil {
		return "", abbilden(err, outbound.GegenstandSonstiges)
	}
	return ort, nil
}

// kopf204 ist der Header einer 204 ohne Wahl: das Kennzeichen der Darstellung und, außer bei DELETE eines Dienstes mit Etag, der ETag
// von neu (SPEC-281 Nr. 2).
func kopf204(mu mutation, dar model.Darstellung, neu model.EntityState) map[string][]string {
	h := map[string][]string{dar.Kennzeichen.Name: {dar.Kennzeichen.Wert}}
	if etag := etagVon(mu.d, neu); etag != "" && mu.methode != methodeDelete {
		h[headerEtag] = []string{etag}
	}
	return h
}
