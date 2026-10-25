// Package baselineimport ist der Importdienst: er überführt die Exchanges einer beendeten Session
// in eine Baseline und setzt ImportUseCase um (SPEC-127). Er importiert Domain, Ports und die
// Standardbibliothek, keinen Adapter und keinen anderen Dienst, startet keine Goroutine und hält
// nach New keinen Zustand.
package baselineimport

import (
	"context"
	"errors"
	"io"
	"math"
	"net/url"
	"slices"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

var _ inbound.ImportUseCase = (*Importer)(nil)

// Dienst ist ein Service der Konfiguration: Kennung und Upstream (SPEC-127 Nr. 1).
type Dienst struct {
	ID          string
	UpstreamURL string
}

// Konfig ist die Konfiguration des Dienstes (SPEC-127 Nr. 1). Dienste dient der Prüfung von
// service_id und der Auflösung der Links, Regeln allein der Bildung des Match-Keys.
type Konfig struct {
	Dienste        []Dienst
	Sessions       outbound.SessionRepository
	Exchanges      outbound.RecordingRepository
	Blobs          outbound.BlobRepository
	Metadaten      outbound.MetadataRepository
	Snapshots      outbound.SnapshotRepository
	Imports        outbound.ImportRepository
	MetadatenCodec outbound.MetadataCodec
	DatenCodec     outbound.ODataCodec
	Regeln         model.SchluesselRegeln
	MaxCaptureBody int64
	Uhr            support.Clock
	IDs            support.IDGenerator
}

// Importer setzt ImportUseCase um. Er hält allein seine Konfiguration.
type Importer struct {
	k Konfig
}

// New prüft die Konfiguration (SPEC-127 Nr. 1).
func New(k Konfig) (*Importer, error) {
	if err := pruefeKonfig(k); err != nil {
		return nil, err
	}
	dienste, err := normiereDienste(k.Dienste)
	if err != nil {
		return nil, err
	}
	k.Dienste = dienste
	return &Importer{k: k}, nil
}

func repositoryFehlt(k Konfig) bool {
	return k.Sessions == nil || k.Exchanges == nil || k.Blobs == nil || k.Metadaten == nil || k.Snapshots == nil || k.Imports == nil
}

func pruefeKonfig(k Konfig) error {
	switch {
	case repositoryFehlt(k):
		return ungueltig("Repository fehlt")
	case k.MetadatenCodec == nil || k.DatenCodec == nil || k.Uhr == nil || k.IDs == nil:
		return ungueltig("Codec, Uhr oder Generator fehlt")
	case k.MaxCaptureBody < 1:
		return ungueltig("MaxCaptureBody ist kleiner als 1")
	case k.Regeln.InvalidEncoding != model.KodierungLiteral && k.Regeln.InvalidEncoding != model.KodierungReject:
		return ungueltig("MATCH_INVALID_ENCODING ist unbekannt")
	case k.Regeln.BodyMode != model.BodyBytes && k.Regeln.BodyMode != model.BodyJSONCanonical:
		return ungueltig("MATCH_BODY_MODE ist unbekannt")
	}
	return nil
}

// normiereDienste prüft die Dienste und liefert eine Kopie, deren Upstream ohne abschließenden
// `/` des Pfades steht.
func normiereDienste(in []Dienst) ([]Dienst, error) {
	if len(in) == 0 {
		return nil, ungueltig("kein Dienst")
	}
	out := make([]Dienst, 0, len(in))
	for _, d := range in {
		if d.ID == "" || slices.ContainsFunc(out, func(v Dienst) bool { return v.ID == d.ID }) {
			return nil, ungueltig("Dienst ohne Kennung oder mit doppelter Kennung")
		}
		u, err := url.Parse(d.UpstreamURL)
		if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return nil, ungueltig("UpstreamURL ist keine absolute http- oder https-URL")
		}
		u.Path = strings.TrimSuffix(u.Path, "/")
		u.RawPath = strings.TrimSuffix(u.RawPath, "/")
		out = append(out, Dienst{ID: d.ID, UpstreamURL: u.String()})
	}
	return out, nil
}

func ctxNil() error { return ungueltig("Kontext fehlt") }

// Import führt die Schritte von SPEC-127 Nr. 4 in ihrer Reihenfolge aus.
func (i *Importer) Import(ctx context.Context, a inbound.ImportAnfrage) (inbound.ImportErgebnis, error) {
	if ctx == nil {
		return inbound.ImportErgebnis{}, ctxNil()
	}
	if err := pruefeAnfrage(a); err != nil {
		return inbound.ImportErgebnis{}, err
	}
	if !slices.ContainsFunc(i.k.Dienste, func(d Dienst) bool { return d.ID == a.ServiceID }) {
		return inbound.ImportErgebnis{}, ohneCode(model.CodeE5005, outbound.ErrDienstUnbekannt)
	}
	hash := anfrageHash(a)
	if res, fertig, err := i.wiederhole(ctx, a, hash); fertig || err != nil {
		return res, err
	}
	exs, err := i.ladeSession(ctx, a)
	if err != nil {
		return inbound.ImportErgebnis{}, err
	}
	ex, err := waehleDokument(exs, a.MetadataHash)
	if err != nil {
		return inbound.ImportErgebnis{}, err
	}
	doc, err := i.liesDokument(ctx, ex)
	if err != nil {
		return inbound.ImportErgebnis{}, err
	}
	r, err := i.verarbeite(ctx, a, exs, doc)
	if err != nil {
		return inbound.ImportErgebnis{}, err
	}
	return i.veroeffentliche(ctx, hash, r)
}

// Get liest einen Import-Eintrag (SPEC-127 Nr. 16).
func (i *Importer) Get(ctx context.Context, importID string) (model.Import, error) {
	if ctx == nil {
		return model.Import{}, ctxNil()
	}
	imp, err := i.k.Imports.Get(ctx, importID)
	if err != nil {
		return model.Import{}, mitCode(err, outbound.GegenstandImport)
	}
	return imp, nil
}

// wiederhole erkennt eine Wiederholung (Nr. 5): fertig heißt, der Aufruf ist beendet.
func (i *Importer) wiederhole(ctx context.Context, a inbound.ImportAnfrage, hash string) (inbound.ImportErgebnis, bool, error) {
	imp, err := i.k.Imports.Get(ctx, a.ImportID)
	if err != nil {
		if errors.Is(err, outbound.ErrNichtGefunden) && !kontextFehler(err) {
			return inbound.ImportErgebnis{}, false, nil
		}
		return inbound.ImportErgebnis{}, true, mitCode(err, outbound.GegenstandImport)
	}
	if imp.RequestHash != hash {
		return inbound.ImportErgebnis{}, true, mitCode(outbound.ErrBereitsVorhanden, outbound.GegenstandImport)
	}
	snap, err := i.k.Snapshots.Get(ctx, imp.BaselineSnapshotID)
	if err != nil {
		return inbound.ImportErgebnis{}, true, sonst(err)
	}
	if err := i.setzeMetadaten(ctx, a.ServiceID, snap.MetadataHash); err != nil {
		return inbound.ImportErgebnis{}, true, err
	}
	return inbound.ImportErgebnis{Import: imp}, true, nil
}

func (i *Importer) setzeMetadaten(ctx context.Context, serviceID, ref string) error {
	if err := i.k.Metadaten.SetMetadata(ctx, serviceID, ref, ref); err != nil {
		return mitCode(err, outbound.GegenstandDienst)
	}
	return nil
}

// ladeSession liest die Session und ihre Exchanges nach sequence_no geordnet (Nr. 6).
func (i *Importer) ladeSession(ctx context.Context, a inbound.ImportAnfrage) ([]model.Exchange, error) {
	id := a.SourceSessionIDs[0]
	s, err := i.k.Sessions.Get(ctx, id)
	if err != nil {
		return nil, mitCode(err, outbound.GegenstandSession)
	}
	if s.ServiceID != a.ServiceID {
		return nil, mitCode(outbound.ErrFremderService, outbound.GegenstandSession)
	}
	if s.Status == model.SessionOpen {
		return nil, ohneCode(model.CodeE7009, nil)
	}
	exs, err := i.k.Exchanges.ListBySession(ctx, id)
	if err != nil {
		return nil, mitCode(err, outbound.GegenstandSession)
	}
	exs = slices.Clone(exs)
	slices.SortStableFunc(exs, func(x, y model.Exchange) int { return x.SequenceNo - y.SequenceNo })
	return exs, nil
}

// verarbeite wählt die Antworten aus, liest und führt sie zusammen und belegt die Coverage
// (Nr. 8 bis 12); bis hierher ist nichts geschrieben.
func (i *Importer) verarbeite(ctx context.Context, a inbound.ImportAnfrage, exs []model.Exchange, doc dokument) (ergebnis, error) {
	r := ergebnis{
		importID: a.ImportID, serviceID: a.ServiceID, sessionID: a.SourceSessionIDs[0], metaRef: doc.ref,
		gesamt: len(exs), uebersprungen: map[string]int{}, metadatenFunde: doc.bericht.Funde,
	}
	r.quellen = append(r.quellen, quelle{art: "metadata", exchangeID: doc.ex.ID, seq: doc.ex.SequenceNo})
	z := neuZusammen()
	var gelesen []*antwort
	for _, e := range exs {
		u := klassifiziere(e, &doc.modell)
		if e.ID == doc.ex.ID {
			u = urteil{spur: spurMetadaten}
		} else if u.spur == spurMetadaten && u.grund == "" {
			u.grund = grundMetadatenDoppelt
		}
		if u.grund != "" {
			r.uebersprungen[u.grund]++
			continue
		}
		r.gelesen++
		if u.spur == spurMetadaten {
			continue
		}
		an, err := i.liesAntwort(ctx, e, u, doc.modell)
		if err != nil {
			return ergebnis{}, err
		}
		for _, ent := range an.seite.Entities {
			if err := z.beobachte(e.ID, u.set, ent); err != nil {
				return ergebnis{}, sonst(err)
			}
		}
		gelesen = append(gelesen, an)
		art := "entity"
		if u.spur == spurCollection {
			art = "collection"
		}
		r.quellen = append(r.quellen, quelle{elemente: an.elemente, set: u.set, exchangeID: e.ID, art: art, seq: e.SequenceNo})
	}
	sortiereQuellen(r.quellen)
	r.entities = z.fertige(doc.modell, r.sessionID)
	r.coverage = coverageEintraege(doc.modell, gelesen, z)
	r.beobachtungen = z.beobachtungen
	r.maskiert = z.maskierte()
	r.nichtUnterst = z.nichtUnterstuetzt()
	return r, nil
}

// liesAntwort öffnet den Body, ruft den Leser der Spur und schließt den Strom (Nr. 9).
func (i *Importer) liesAntwort(ctx context.Context, e model.Exchange, u urteil, m model.ServiceModel) (*antwort, error) {
	an := &antwort{ex: e, spur: u.spur, set: u.set, ungefiltert: u.ungefiltert}
	err := i.oeffne(ctx, e.ResponseBlob, func(src io.Reader) error {
		if u.spur == spurCollection {
			s, err := i.k.DatenCodec.ParseCollection(ctx, src, m, u.set, i.k.MaxCaptureBody)
			an.seite = s
			return err
		}
		ent, err := i.k.DatenCodec.ParseEntity(ctx, src, m, u.set, i.k.MaxCaptureBody)
		an.seite = model.EntitySeite{Entities: []model.Entity{ent}}
		return err
	})
	if err != nil {
		return nil, err
	}
	an.elemente = len(an.seite.Entities)
	an.schluessel = map[string]struct{}{}
	for _, ent := range an.seite.Entities {
		an.schluessel[ent.Schluessel] = struct{}{}
	}
	return an, nil
}

// veroeffentliche bezieht Kennung und Zeit, schreibt den Blob des Manifests, ruft Publish und
// setzt die Metadaten (Nr. 4 h bis l, Nr. 15).
func (i *Importer) veroeffentliche(ctx context.Context, hash string, r ergebnis) (inbound.ImportErgebnis, error) {
	r.snapshotID = i.k.IDs.SnapshotID()
	r.zeit = i.k.Uhr.Now()
	ref, err := i.schreibeBlob(ctx, r.manifest())
	if err != nil {
		return inbound.ImportErgebnis{}, err
	}
	pub := model.Publication{
		Import: model.Import{
			ID: r.importID, RequestHash: hash, BaselineSnapshotID: r.snapshotID, ServiceID: r.serviceID,
			CreatedAt: r.zeit, ReportJSON: r.antwortObjekt(),
		},
		Snapshot: model.BaselineSnapshot{
			ID: r.snapshotID, ServiceID: r.serviceID, CreatedAt: r.zeit, StateBlob: ref, StateHash: ref, MetadataHash: r.metaRef,
		},
		SourceSessionIDs: []string{r.sessionID},
		Coverage:         r.coverage,
	}
	imp, neu, err := i.k.Imports.Publish(ctx, pub)
	if err != nil {
		g := outbound.GegenstandImport
		if errors.Is(err, outbound.ErrNichtGefunden) {
			g = outbound.GegenstandSession
		}
		return inbound.ImportErgebnis{}, mitCode(err, g)
	}
	if err := i.setzeMetadaten(ctx, r.serviceID, r.metaRef); err != nil {
		return inbound.ImportErgebnis{}, err
	}
	return inbound.ImportErgebnis{Import: imp, Neu: neu}, nil
}

// schreibeBlob legt die Bytes als Blob ab (Nr. 15 a); scheitert Write oder Commit, ruft es Abort
// genau einmal.
func (i *Importer) schreibeBlob(ctx context.Context, daten []byte) (string, error) {
	w, err := i.k.Blobs.Begin(ctx, math.MaxInt64)
	if err != nil {
		return "", sonst(err)
	}
	ref := ""
	for len(daten) > 0 && err == nil {
		var n int
		n, err = w.Write(daten)
		switch {
		case err != nil:
		case n <= 0 || n > len(daten):
			err = errors.New("Write meldet eine Länge außerhalb von 1 bis len(p)")
		default:
			daten = daten[n:]
		}
	}
	if err == nil {
		ref, _, err = w.Commit()
	}
	if err != nil {
		_ = w.Abort()
		return "", sonst(err)
	}
	return ref, nil
}
