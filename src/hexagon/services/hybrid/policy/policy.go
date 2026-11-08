// Package policy prüft die Werte der Fallback-Policy und hält sie in geparster Form
// (LH-FA-12.a, SPEC-261). Das Paket importiert model und die Standardbibliothek, keinen Port
// und keinen Adapter; New ruft nichts, liest keine Umgebung und löst keinen Namen auf.
package policy

import (
	"errors"
	"fmt"
)

// Schlüssel der Einstellungen, die ein KonfigFehler nennt (SPEC-010, SPEC-062, SPEC-079, SPEC-080).
const (
	SchluesselUpstream         = "HYBRID_FALLBACK"
	SchluesselHosts            = "FALLBACK_ALLOW_HOSTS"
	SchluesselSchreibfreigaben = "FALLBACK_WRITE_ALLOW"
	SchluesselReplayKategorien = "HYBRID_REPLAY_FALLBACK"
)

// Ursachen eines KonfigFehler (SPEC-261 Nr. 5 a, 6 a, 7, 8 c).
const (
	UrsacheLeer      = "leer"
	UrsacheTrenner   = "trenner"
	UrsacheMethode   = "methode"
	UrsacheMuster    = "muster"
	UrsacheHost      = "host"
	UrsacheKategorie = "kategorie"
	UrsacheOhneHosts = "ohne-hosts"
)

// Werte sind die typisierten Werte der vier Schlüssel; die Listen sind die Einträge nach der
// Zerlegung durch den Root (SPEC-261, SPEC-110 Nr. 6).
type Werte struct {
	Upstream         bool
	Hosts            []string
	Schreibfreigaben []string
	ReplayKategorien []string
}

// KonfigFehler benennt den ersten Verstoß eines Schlüssels. Eintrag zählt ab 1, bei einem
// Zusammenhang ohne Eintrag ist er 0. Der Text nennt nie den Wert eines Eintrags (SPEC-261 Nr. 8 b).
type KonfigFehler struct {
	Schluessel string
	Eintrag    int
	Ursache    string
}

func (e *KonfigFehler) Error() string {
	return fmt.Sprintf("%s: Eintrag %d: %s", e.Schluessel, e.Eintrag, e.Ursache)
}

// Policy hält die geparsten Einträge; nach New ändert sich nichts mehr (SPEC-261 Nr. 1 c).
type Policy struct {
	upstream          bool
	hosts             []hostEintrag
	freigaben         []freigabe
	replayUnsupported bool
	replayCoverage    bool
}

const (
	kategorieUnsupported = "unsupported"
	kategorieCoverage    = "coverage_insufficient"
)

// pruefeKategorie liefert die Ursache für einen Eintrag von ReplayKategorien, leer bei einem
// gültigen (SPEC-261 Nr. 7).
func pruefeKategorie(e string) string {
	switch e {
	case kategorieUnsupported, kategorieCoverage:
		return ""
	case "":
		return UrsacheLeer
	}
	return UrsacheKategorie
}

// New prüft jeden Eintrag der drei Listen, auch bei Upstream aus, und liefert je Schlüssel den
// ersten Verstoß, danach den Zusammenhang, als errors.Join und ohne Policy (SPEC-261 Nr. 8).
func New(w Werte) (*Policy, error) {
	p := &Policy{upstream: w.Upstream}
	var fehler []error

	for i, e := range w.Hosts {
		h, ursache := parseHost(e)
		if ursache != "" {
			fehler = append(fehler, &KonfigFehler{SchluesselHosts, i + 1, ursache})
			break
		}
		p.hosts = append(p.hosts, h)
	}
	for i, e := range w.Schreibfreigaben {
		f, ursache := parseFreigabe(e)
		if ursache != "" {
			fehler = append(fehler, &KonfigFehler{SchluesselSchreibfreigaben, i + 1, ursache})
			break
		}
		p.freigaben = append(p.freigaben, f)
	}
	for i, e := range w.ReplayKategorien {
		ursache := pruefeKategorie(e)
		if ursache != "" {
			fehler = append(fehler, &KonfigFehler{SchluesselReplayKategorien, i + 1, ursache})
			break
		}
		if e == kategorieUnsupported {
			p.replayUnsupported = true
		} else {
			p.replayCoverage = true
		}
	}
	if w.Upstream && len(w.Hosts) == 0 {
		fehler = append(fehler, &KonfigFehler{SchluesselUpstream, 0, UrsacheOhneHosts})
	}
	if len(fehler) > 0 {
		return nil, errors.Join(fehler...)
	}
	return p, nil
}
