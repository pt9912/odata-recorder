package support

import "github.com/pt9912/odata-recorder/src/hexagon/model"

// AuditLog hält fest, was ein Dienst gelöscht hat und welcher Import scheiterte (SPEC-234,
// LH-RB-04, LH-FA-13). Append hat keinen ctx, ist für gleichzeitigen Aufruf sicher und gerät nie
// in Panik; es liefert model.ErrUngueltig bei einem Verstoß gegen Art, Feld oder Wertform
// (nichts geschrieben) oder den Fehler der Ablage.
type AuditLog interface {
	Append(e Eintrag) error
}

// AuditArt ist die Art eines Eintrags (SPEC-234 Nr. 2); die Menge ist geschlossen. Der Name
// weicht vom Vertrag ab, weil Art den Typ der Attribute von Telemetry trägt.
type AuditArt string

// Die vier Arten (SPEC-234 Nr. 2).
const (
	ArtSessionGeloescht     AuditArt = "session_geloescht"
	ArtSessionAufbewahrung  AuditArt = "session_aufbewahrung"
	ArtWorkspaceGeloescht   AuditArt = "workspace_geloescht"
	ArtImportFehlgeschlagen AuditArt = "import_fehlgeschlagen"
)

// Eintrag ist ein Audit-Eintrag. Je Art ist jedes Feld außer KorrelationsID Pflicht (nichtleer)
// oder verboten (leer), siehe Arten; KorrelationsID ist bei jeder Art optional (SPEC-234 Nr. 2, 7).
type Eintrag struct {
	Art            AuditArt
	ServiceID      string
	SessionID      string
	WorkspaceID    string
	ImportID       string
	ErrorCode      model.Code
	KorrelationsID string
}

// Feld benennt ein Feld von Eintrag, das je Art Pflicht oder verboten ist (SPEC-234 Nr. 2).
type Feld string

// Die fünf Felder der Tabelle in der Folge ihrer Spalten.
const (
	FeldServiceID   Feld = "service_id"
	FeldSessionID   Feld = "session_id"
	FeldWorkspaceID Feld = "workspace_id"
	FeldImportID    Feld = "import_id"
	FeldErrorCode   Feld = "error_code"
)

// ArtSpec nennt die Pflichtfelder einer Art; jedes andere Feld ist verboten.
type ArtSpec struct {
	Art     AuditArt
	Pflicht []Feld
}

// Arten liefert die Tabelle von SPEC-234 Nr. 2 als Daten, je Aufruf eine neue Kopie.
func Arten() []ArtSpec {
	return []ArtSpec{
		{ArtSessionGeloescht, []Feld{FeldServiceID, FeldSessionID}},
		{ArtSessionAufbewahrung, []Feld{FeldServiceID, FeldSessionID}},
		{ArtWorkspaceGeloescht, []Feld{FeldWorkspaceID}},
		{ArtImportFehlgeschlagen, []Feld{FeldServiceID, FeldImportID, FeldErrorCode}},
	}
}
