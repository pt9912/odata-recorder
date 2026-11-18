package bootstrap

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// ursachenText ist der Grund je Ursache eines policy.KonfigFehler (SPEC-110 Nr. 16 c).
func ursachenText(ursache string) string {
	return map[string]string{
		policy.UrsacheLeer:      "leerer Eintrag",
		policy.UrsacheTrenner:   "Methode und Muster nicht durch ein Leerzeichen getrennt",
		policy.UrsacheMethode:   "Methode ist keine mutierende Methode",
		policy.UrsacheMuster:    "Pfadmuster ungültig",
		policy.UrsacheHost:      "Host ungültig",
		policy.UrsacheKategorie: "Kategorie unbekannt; erlaubt sind unsupported und coverage_insufficient",
		policy.UrsacheOhneHosts: "enabled verlangt mindestens einen Eintrag in FALLBACK_ALLOW_HOSTS",
	}[ursache]
}

// liesHybrid liest HYBRID_FALLBACK und die drei Listen des Modus hybrid (SPEC-110 Nr. 16), in jedem Modus; die Form der
// Einträge und die Zusammenhänge prüft policy.New, jeden Fehler des Join macht der Root zu einem Verstoß.
// Ein ungültiges HYBRID_FALLBACK lässt Upstream falsch: es gilt weder als enabled noch als disabled.
func (p *pruefer) liesHybrid(k *Konfiguration) {
	k.Fallback = policy.Werte{
		Upstream:         p.aufzaehlung("HYBRID_FALLBACK", "disabled", "disabled", "enabled") == "enabled",
		Hosts:            p.eintraege("FALLBACK_ALLOW_HOSTS"),
		Schreibfreigaben: p.eintraege("FALLBACK_WRITE_ALLOW"),
		ReplayKategorien: p.eintraege("HYBRID_REPLAY_FALLBACK"),
	}
	_, err := policy.New(k.Fallback)
	p.meldeKonfigFehler(err)
}

// eintraege zerlegt eine Liste nach SPEC-110 Nr. 6: Komma, Leerraum um einen Eintrag entfernt; jeder Eintrag bleibt, auch ein
// leerer und ein doppelter (SPEC-110 Nr. 16 b). Eine Liste ohne Text ist leer.
func (p *pruefer) eintraege(schluessel string) []string {
	w, ok := p.wert(schluessel)
	if !ok {
		return nil
	}
	e := strings.Split(w, ",")
	for i := range e {
		e[i] = kappe(e[i])
	}
	return e
}

// meldeKonfigFehler macht jeden *policy.KonfigFehler im Baum von err, auch jeden Fehler eines errors.Join, zu einem Verstoß
// (SPEC-110 Nr. 16 c); der Grund nennt nie den Wert eines Eintrags.
func (p *pruefer) meldeKonfigFehler(err error) {
	switch e := err.(type) {
	case interface{ Unwrap() []error }:
		for _, x := range e.Unwrap() {
			p.meldeKonfigFehler(x)
		}
	case *policy.KonfigFehler:
		grund := ursachenText(e.Ursache)
		if e.Eintrag > 0 {
			grund = fmt.Sprintf("Eintrag %d: %s", e.Eintrag, grund)
		}
		p.melde(e.Schluessel, grund)
	}
}

// NeuePolicy baut die Fallback-Policy aus den gelesenen Werten (SPEC-261 Nr. 10 b). Ein Fehler von policy.New ist ein
// *KonfigFehler (ODR-E2001); die Policy ist dann nil.
func (k Konfiguration) NeuePolicy() (*policy.Policy, error) {
	pol, err := policy.New(k.Fallback)
	if err == nil {
		return pol, nil
	}
	p := &pruefer{}
	p.meldeKonfigFehler(err)
	return nil, cmp.Or(p.fehler(), err)
}
