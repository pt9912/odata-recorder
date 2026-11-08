// Package mock ist der Mock-Dienst: er setzt MockUseCase um (SPEC-195). Er importiert Domain und Ports, keinen
// Adapter, keinen anderen Dienst und keinen Port zu einem Upstream oder zum Replay.
package mock

import (
	"context"
	"slices"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

var _ inbound.MockUseCase = (*Mocker)(nil)

// Dienst ist ein Service der Konfiguration: Kennung, Root und der Satz der Codecs (SPEC-195 Nr. 1); Codecs gleich nil
// heißt, dass MetadatenCodec, URLCodec und Schreiber der Konfig gelten. Etag wahr liefert ETags aus dem Zähler `entity_version`
// der Entity (SPEC-280 Nr. 2, SPEC-281); der Nullwert liefert keine.
type Dienst struct {
	ID     string
	Root   string
	Codecs *outbound.CodecSatz
	Etag   bool
}

// Konfig ist die Konfiguration des Dienstes (SPEC-195 Nr. 1).
type Konfig struct {
	Dienste        []Dienst
	Metadaten      outbound.MetadataRepository
	Workspaces     outbound.WorkspaceRepository
	Entities       outbound.EntityRepository
	Coverage       outbound.CoverageRepository
	Blobs          outbound.BlobRepository
	MetadatenCodec outbound.MetadataCodec
	URLCodec       outbound.URLCodec
	Schreiber      outbound.ODataSchreiber
	PublicBaseURL  string
	StrictCoverage bool
	MaxBytes       int64
	// Transaktion führt jede Mutation in einer Workspace-Transaktion aus (SPEC-192 Nr. 1); nil ist ungültig.
	Transaktion outbound.WorkspaceTransactor
	// Eingabe ist der Leser der Client-Bodies für einen Dienst ohne Satz; nil ist erlaubt (SPEC-192 Nr. 1).
	Eingabe outbound.ODataEingabe
}

// Mocker setzt MockUseCase um. Er hält nach New nur seine Konfig, die er nicht ändert, keine Sperre und keine
// Goroutine (SPEC-195 Nr. 12).
type Mocker struct {
	k Konfig
}

// New prüft die Konfiguration (SPEC-195 Nr. 1) und ruft keinen Port; Form und Überlappung der Roots und die Form von
// PublicBaseURL prüft sie nicht. Ein Dienst ohne Satz erhält eine Kopie des Satzes der Konfig.
func New(k Konfig) (*Mocker, error) {
	if err := pruefeKonfig(k); err != nil {
		return nil, err
	}
	k.Dienste = slices.Clone(k.Dienste)
	for i, d := range k.Dienste {
		s := outbound.CodecSatz{Metadaten: k.MetadatenCodec, URL: k.URLCodec, Schreiber: k.Schreiber, Eingabe: k.Eingabe}
		if d.Codecs != nil {
			s = *d.Codecs
		}
		k.Dienste[i].Codecs = &s
	}
	return &Mocker{k: k}, nil
}

func pruefeKonfig(k Konfig) error {
	switch {
	case len(k.Dienste) == 0:
		return ungueltig("kein Dienst")
	case slices.ContainsFunc(k.Dienste, func(d Dienst) bool { return d.ID == "" || d.Root == "" }):
		return ungueltig("Dienst ohne Kennung oder Root")
	case portFehlt(k):
		return ungueltig("Port fehlt")
	case slices.ContainsFunc(k.Dienste, func(d Dienst) bool {
		return d.Codecs != nil && (d.Codecs.Metadaten == nil || d.Codecs.URL == nil || d.Codecs.Schreiber == nil)
	}):
		return ungueltig("Satz der Codecs unvollständig")
	case k.PublicBaseURL == "":
		return ungueltig("PublicBaseURL ist leer")
	case k.MaxBytes < 1:
		return ungueltig("MaxBytes ist kleiner als 1")
	}
	return nil
}

// portFehlt sagt, ob eines der neun Felder mit einem Port nil ist (SPEC-195 Nr. 1, SPEC-192 Nr. 1); Eingabe darf nil sein.
func portFehlt(k Konfig) bool {
	return k.Metadaten == nil || k.Workspaces == nil || k.Entities == nil || k.Coverage == nil || k.Blobs == nil ||
		k.MetadatenCodec == nil || k.URLCodec == nil || k.Schreiber == nil || k.Transaktion == nil
}

// Mock prüft den Request in der Reihenfolge von SPEC-195 Nr. 3 (der erste Verstoß entscheidet) und liefert die
// Antwort: $metadata (SPEC-198 Nr. 4), Collection, Entity und Anzahl (SPEC-196 bis SPEC-198); POST, PATCH, PUT und DELETE führt
// schreibe in einer Workspace-Transaktion aus (SPEC-192 bis SPEC-194).
func (m *Mocker) Mock(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
	if argumenteUngueltig(ctx, a) {
		return inbound.Antwort{}, ungueltig("Anfrage, die kein HTTP-Adapter liefert")
	}
	d, pfad, ok := dienstFuer(m.k.Dienste, a.Pfad)
	if !ok {
		return inbound.Antwort{}, mitCode(model.CodeE5005, nil)
	}
	ws, ok := einzelwert(a.Header, "x-recorder-workspace")
	if !ok {
		return inbound.Antwort{}, mitCode(model.CodeE5003, nil)
	}
	u, err := d.Codecs.URL.ParseURL(pfad, a.RawQuery)
	if err != nil {
		return inbound.Antwort{}, abbilden(err, outbound.GegenstandSonstiges)
	}
	if !slices.Contains(model.AlleRessourcenArten(), u.Ressource.Art) {
		return inbound.Antwort{}, mitCode(model.CodeE9001, nil)
	}
	if istMutation(a.Methode) {
		return m.schreibe(ctx, a, d, ws, u)
	}
	if a.Methode != "GET" && a.Methode != "HEAD" {
		return inbound.Antwort{}, nichtGelesen(d)
	}
	w, err := m.holeWorkspace(ctx, d, ws)
	if err != nil {
		return inbound.Antwort{}, err
	}
	s, err := m.holeService(ctx, d)
	if err != nil {
		return inbound.Antwort{}, err
	}
	if u.Ressource.Art != model.RessourceMetadaten {
		return m.lese(ctx, a, d, w, s, u)
	}
	return m.metadaten(ctx, a.Methode == "HEAD", d, s)
}

// argumenteUngueltig sagt, ob der Aufruf ein Programmfehler des Adapters ist (SPEC-195 Nr. 2).
func argumenteUngueltig(ctx context.Context, a inbound.Anfrage) bool {
	return ctx == nil || a.Methode == "" || a.Laenge < -1 || a.Body == nil && a.Laenge > 0
}

// holeService liest den Service genau einmal; ohne Metadaten ist er nicht verwendbar (SPEC-195 Nr. 9).
func (m *Mocker) holeService(ctx context.Context, d Dienst) (model.Service, error) {
	s, err := m.k.Metadaten.Get(ctx, d.ID)
	switch {
	case err != nil:
		return model.Service{}, abbilden(err, outbound.GegenstandDienst)
	case s.MetadataBlob == "":
		return model.Service{}, mitCode(model.CodeE6004, nil)
	}
	return s, nil
}

// nichtGelesen ist der Fehler einer Methode außer GET und HEAD (SPEC-195 Nr. 7): ODR-E6006 bei einem Dienst, dessen Satz nur
// lesend ist, sonst ODR-E6002.
func nichtGelesen(d Dienst) error {
	if d.Codecs.NurLesend {
		return mitCode(model.CodeE6006, &model.KompatFehler{Ursache: model.KompatMethode})
	}
	return mitCode(model.CodeE6002, nil)
}

// holeWorkspace liest den Workspace genau einmal und prüft seinen Service (SPEC-195 Nr. 8).
func (m *Mocker) holeWorkspace(ctx context.Context, d Dienst, id string) (model.Workspace, error) {
	w, err := m.k.Workspaces.Get(ctx, id)
	switch {
	case err != nil:
		return model.Workspace{}, abbilden(err, outbound.GegenstandWorkspace)
	case w.ServiceID != d.ID:
		return model.Workspace{}, mitCode(model.CodeE5004, nil)
	}
	return w, nil
}
