package bootstrap

import (
	"cmp"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// ExitKonfiguration ist der Exit-Code, wenn die Startkonfiguration ungültig ist (SPEC-110 Nr. 12).
const ExitKonfiguration = 1

// Dienst ist ein konfigurierter Service (LH-FA-04.c). UpstreamURL ist leer, wenn keine gesetzt ist.
type Dienst struct {
	ID          string
	Root        string
	UpstreamURL string
}

// Konfiguration trägt die typisierten Werte der Startkonfiguration (SPEC-110); die Services
// erhalten sie über ihre Konstruktoren, nie die Umgebung. Listen sind Mengen: sortiert, ohne Doppelte.
type Konfiguration struct {
	Modus                string
	Listen               string
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

// wert liefert den Wert ohne Leerraum am Rand; leer gilt als ungesetzt (SPEC-110 Nr. 1).
func (p *pruefer) wert(schluessel string) (string, bool) {
	w := strings.TrimSpace(p.env[schluessel])
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
// Fehler ist ein *KonfigFehler mit allen Verstößen. Dieselbe Eingabe ergibt dasselbe Ergebnis.
func LadeKonfiguration(environ []string) (Konfiguration, error) {
	p := &pruefer{env: umgebung(environ)}
	var k Konfiguration
	p.nichtGeliefert()
	p.liesModus(&k)
	p.liesEinfache(&k)
	p.liesMengen(&k)
	p.liesTLS(&k)
	p.liesDienste(&k)
	p.pruefeCADatei(&k)
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
