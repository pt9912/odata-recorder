package bootstrap

import (
	"cmp"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/clock"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// ExitKonfiguration ist der Exit-Code, wenn die Startkonfiguration ungültig ist (SPEC-110 Nr. 12).
const ExitKonfiguration = 1

// Dienst ist ein konfigurierter Service (LH-FA-04.c). UpstreamURL ist leer, wenn keine gesetzt ist. Version ist die OData-Version
// des Services, 4.0 oder 2.0 (LH-FA-17.a); die geladene Konfiguration trägt den wirksamen Wert, nie einen leeren (SPEC-110 Nr. 17 a),
// und Build liest in einer von Hand gebauten Konfiguration eine leere Version als 4.0 (SPEC-139 Nr. 2).
type Dienst struct {
	ID          string
	Root        string
	UpstreamURL string
	Version     string
}

// Konfiguration trägt die typisierten Werte der Startkonfiguration (SPEC-110); die Services
// erhalten sie über ihre Konstruktoren, nie die Umgebung. Listen sind Mengen: sortiert, ohne Doppelte.
type Konfiguration struct {
	Modus  string
	Listen string
	// ListenTLSCert und ListenTLSKey sind die Pfade von LISTEN_TLS_CERT und LISTEN_TLS_KEY, leer heißt ungesetzt; die Konfiguration
	// trägt nie den Inhalt der Dateien (SPEC-291 Nr. 7 a, 8).
	ListenTLSCert string
	ListenTLSKey  string
	// AdminListen und AdminTokenFile sind ADMIN_LISTEN und ADMIN_TOKEN_FILE, nie der Inhalt der Datei; ein leeres AdminListen in
	// einem Literal heißt: kein Admin-Listener (SPEC-300 Nr. 2, 11).
	AdminListen          string
	AdminTokenFile       string
	DataDir              string
	Dienste              []Dienst
	ReplayStrategy       string
	ReplayAssignment     string
	ReplayRedactedBodies bool
	MatchVaryHeaders     []string
	MatchInvalidEncoding string
	MatchBodyMode        string
	RedactHeaders        []string
	RedactQueryParams    []string
	RedactJSONPaths      []string
	MaxCaptureBody       int64
	UpstreamTimeout      time.Duration
	UpstreamCAFile       string
	UpstreamTLSVerify    bool
	ShutdownTimeout      time.Duration
	RetentionMaxAge      time.Duration
	SQLiteBusyTimeout    time.Duration
	VolatileHeaders      []string
	LinkRewrite          string
	// StrictCoverage und PublicBaseURL wirken in den Modi mock und hybrid (SPEC-110 Nr. 14); PublicBaseURL ist in den
	// anderen Modi leer, wenn der Schlüssel fehlt.
	StrictCoverage bool
	PublicBaseURL  string
	// EtagEnabled ist ETAG_ENABLED (SPEC-110 Nr. 15, SPEC-282): in jedem Modus geprüft, in den Modi mock und hybrid nur für
	// Dienste der Version 4.0 wirksam (SPEC-280 Nr. 2).
	EtagEnabled bool
	// Fallback sind die Werte von HYBRID_FALLBACK und der drei Listen (SPEC-110 Nr. 16, SPEC-261), in jedem Modus gelesen; die
	// Listen stehen in der Folge der Eingabe, mit Doppelten (der Eintrag der Meldung zählt sie mit), und NeuePolicy baut daraus
	// die Policy.
	Fallback policy.Werte
}

// KonfigFehler sammelt alle Verstöße der Startkonfiguration, je Zeile
// `ODR-E2001 <Schlüssel>: <Grund>`, nach Schlüssel (Bytefolge), dann Grund sortiert, gleiche Zeilen einmal; sie nennt nie einen Wert (SPEC-110 Nr. 12).
type KonfigFehler struct {
	Zeilen []string
}

// Error liefert die Zeilen, durch Zeilenumbruch getrennt.
func (e *KonfigFehler) Error() string { return strings.Join(e.Zeilen, "\n") }

// Code liefert ODR-E2001.
func (*KonfigFehler) Code() model.Code { return model.CodeE2001 }

type verstoss struct{ schluessel, grund string }

// pruefer hält die Umgebung und sammelt die Verstöße.
type pruefer struct {
	env        map[string]string
	verstoesse []verstoss
}

func (p *pruefer) melde(schluessel, grund string) {
	p.verstoesse = append(p.verstoesse, verstoss{schluessel, grund})
}

// leerraum sind die Zeichen, die die Konfiguration am Rand eines Wertes, eines Listeneintrags und eines Feldes der Services-Datei
// entfernt: Leerzeichen, Tabulator, LF und CR, kein anderes Zeichen (SPEC-110 Nr. 1, 6).
const leerraum = " \t\n\r"

// kappe entfernt leerraum an beiden Rändern von s.
func kappe(s string) string { return strings.Trim(s, leerraum) }

// wert liefert den Wert ohne Leerraum am Rand; leer gilt als ungesetzt (SPEC-110 Nr. 1).
func (p *pruefer) wert(schluessel string) (string, bool) {
	w := kappe(p.env[schluessel])
	return w, w != ""
}

func umgebung(environ []string) map[string]string {
	env := make(map[string]string, len(environ))
	for _, e := range environ {
		if k, v, ok := strings.Cut(e, "="); ok && k != "" {
			env[k] = v
		}
	}
	return env
}

func (p *pruefer) fehler() error {
	if len(p.verstoesse) == 0 {
		return nil
	}
	sortiert := slices.Clone(p.verstoesse)
	slices.SortFunc(sortiert, func(a, b verstoss) int {
		return cmp.Or(strings.Compare(a.schluessel, b.schluessel), strings.Compare(a.grund, b.grund))
	})
	sortiert = slices.Compact(sortiert)
	zeilen := make([]string, len(sortiert))
	for i, v := range sortiert {
		zeilen[i] = fmt.Sprintf("%s %s: %s", model.CodeE2001, v.schluessel, v.grund)
	}
	return &KonfigFehler{Zeilen: zeilen}
}

// LadeKonfiguration liest die Startkonfiguration aus environ (Einträge `NAME=Wert`) und aus
// SERVICES_FILE, prüft sie nach SPEC-110 und legt DATA_DIR an, wenn alles andere besteht. Der
// Fehler ist ein *KonfigFehler mit allen Verstößen. Dieselbe Eingabe ergibt dasselbe Ergebnis, außer
// bei einem Zertifikat von LISTEN_TLS_CERT, dessen Gültigkeit an der Zeit des Aufrufs hängt (SPEC-291 Nr. 3).
func LadeKonfiguration(environ []string) (Konfiguration, error) {
	return ladeKonfigurationZu(environ, clock.NewSystemuhr().Now())
}

// ladeKonfigurationZu ist LadeKonfiguration mit der Zeit des Starts jetzt.
func ladeKonfigurationZu(environ []string, jetzt time.Time) (Konfiguration, error) {
	p := &pruefer{env: umgebung(environ)}
	var k Konfiguration
	p.liesModus(&k)
	p.liesEinfache(&k)
	p.liesAdmin(&k, jetzt)
	p.liesMengen(&k)
	p.liesTLS(&k)
	p.liesListenTLS(&k)
	p.liesHybrid(&k)
	p.liesDienste(&k)
	p.pruefeVersionModus(&k)
	p.liesPublicBaseURL(&k)
	p.pruefeCADatei(&k)
	p.pruefeListenTLS(&k, jetzt)
	if len(p.verstoesse) == 0 {
		p.legeDatenverzeichnisAn(k.DataDir)
	}
	if err := p.fehler(); err != nil {
		return Konfiguration{}, err
	}
	return k, nil
}

// LadeOderMelde lädt die Konfiguration und schreibt die Verstöße nach stderr; der zweite Wert ist
// 0 oder ExitKonfiguration (SPEC-110 Nr. 12).
func LadeOderMelde(environ []string, stderr io.Writer) (Konfiguration, int) {
	k, err := LadeKonfiguration(environ)
	if err != nil {
		if stderr != nil {
			_, _ = fmt.Fprintln(stderr, err.Error())
		}
		return Konfiguration{}, ExitKonfiguration
	}
	return k, exitOK
}

func (p *pruefer) legeDatenverzeichnisAn(dir string) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		p.melde("DATA_DIR", "Verzeichnis nicht anlegbar oder kein Verzeichnis")
	}
}
