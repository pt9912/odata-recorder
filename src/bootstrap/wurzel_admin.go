package bootstrap

// Dienste und Routen des Admin-Adapters (SPEC-300 Nr. 5 a bis c, Nr. 6): Build baut sie in jedem Modus und legt die Routen im
// Handle ab.

import (
	"context"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/inspection"
	"github.com/pt9912/odata-recorder/src/hexagon/services/session"
)

// ohneRecorder steht in den Modi ohne Recorder an seiner Stelle (SPEC-300 Nr. 5 a): er nimmt nichts auf.
type ohneRecorder struct{}

// Record ist model.ErrUngueltig.
func (ohneRecorder) Record(context.Context, inbound.Anfrage) (inbound.Antwort, error) {
	return inbound.Antwort{}, ungueltig("kein Recorder außerhalb des Modus record")
}

// Drain liefert den leeren Bericht ohne Fehler.
func (ohneRecorder) Drain(context.Context, string, time.Duration) (inbound.DrainBericht, error) {
	return inbound.DrainBericht{}, nil
}

// baueSessions baut den SessionUseCase; record ist im Modus record der Recorder, sonst ohneRecorder.
func (h *Handle) baueSessions(db *sqlite.DB, record inbound.RecordUseCase) error {
	var err error
	h.sessions, err = h.b.NeueSession(session.Konfig{
		Sessions: db.Sessions(), Dienste: h.dienste, Wartung: h.wartung, Uhr: h.uhr, IDs: h.ids, Log: slog.New(h.tel.Handler()),
		Record: record, Exchanges: db.Recording(), Frist: h.k.ShutdownTimeout, Audit: h.audit,
	})
	return err
}

// verdrahteAdmin baut außerhalb des Modus record den SessionUseCase, in jedem Modus die Inspektion, und legt die Routen ab.
func (h *Handle) verdrahteAdmin(db *sqlite.DB) error {
	if h.k.Modus != modusRecord {
		if err := h.baueSessions(db, ohneRecorder{}); err != nil {
			return err
		}
	}
	var err error
	h.inspektion, err = h.b.NeueInspektion(inspection.Konfig{
		Exchanges: db.Recording(), Coverage: db.Coverage(), Zaehler: h.tel, Modus: h.k.Modus, Einstellungen: einstellungen(h.k),
	})
	if err != nil {
		return err
	}
	h.routen, err = h.adminRouten()
	return err
}

// adminRouten ruft jeden Konstruktor einmal und fügt die Routen in der Folge Sessions, Inspektion, Imports, Workspaces
// zusammen; die der Imports gibt es nur im Modus record (SPEC-300 Nr. 5 c).
func (h *Handle) adminRouten() ([]admin.Route, error) {
	sitzungen, err := admin.SessionRouten(h.sessions, h.audit)
	if err != nil {
		return nil, err
	}
	inspektion, err := admin.InspektionRouten(h.inspektion)
	if err != nil {
		return nil, err
	}
	var importe []admin.Route
	if h.k.Modus == modusRecord {
		if importe, err = admin.ImportRouten(h.importer, h.audit); err != nil {
			return nil, err
		}
	}
	arbeitsraeume, err := admin.WorkspaceRouten(h.workspaces, h.snapshots, h.audit)
	if err != nil {
		return nil, err
	}
	return slices.Concat(sitzungen, inspektion, importe, arbeitsraeume), nil
}

// einstellungen sind die Einträge der Diagnose, einer je Schlüssel von inspection.Erlaubt, aus den Werten der Konfiguration
// (SPEC-300 Nr. 6): eine Dauer in ihren ganzen Millisekunden, eine leere Liste als none.
func einstellungen(k Konfiguration) map[string]string {
	dauer := func(d time.Duration) string { return strconv.FormatInt(d.Milliseconds(), 10) + "ms" }
	liste := func(eintraege []string) string {
		if len(eintraege) == 0 {
			return "none"
		}
		return strings.Join(eintraege, ",")
	}
	return map[string]string{
		"REPLAY_STRATEGY":                   k.ReplayStrategy,
		"REPLAY_ASSIGNMENT":                 k.ReplayAssignment,
		"REPLAY_REDACTED_BODIES":            strconv.FormatBool(k.ReplayRedactedBodies),
		"MATCH_VARY_HEADERS":                liste(k.MatchVaryHeaders),
		"MATCH_INVALID_ENCODING":            k.MatchInvalidEncoding,
		"MATCH_BODY_MODE":                   k.MatchBodyMode,
		"REDACT_HEADERS":                    liste(k.RedactHeaders),
		"REDACT_QUERY_PARAMS":               liste(k.RedactQueryParams),
		"REDACT_JSON_PATHS":                 liste(k.RedactJSONPaths),
		"MAX_CAPTURE_BODY":                  strconv.FormatInt(k.MaxCaptureBody, 10),
		"UPSTREAM_TIMEOUT":                  dauer(k.UpstreamTimeout),
		"UPSTREAM_TLS_VERIFY":               strconv.FormatBool(k.UpstreamTLSVerify),
		"UPSTREAM_TLS_INSECURE_SKIP_VERIFY": strconv.FormatBool(!k.UpstreamTLSVerify),
		"SHUTDOWN_TIMEOUT":                  dauer(k.ShutdownTimeout),
		"RETENTION_MAX_AGE":                 dauer(k.RetentionMaxAge),
		"SQLITE_BUSY_TIMEOUT":               dauer(k.SQLiteBusyTimeout),
		"VOLATILE_HEADERS":                  liste(k.VolatileHeaders),
		"LINK_REWRITE":                      k.LinkRewrite,
	}
}
