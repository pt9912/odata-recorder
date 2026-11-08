// Package hybrid ist der Hybrid-Dienst: er setzt HybridUseCase um (SPEC-284). Er importiert Domain, die Inbound Ports
// von Mock und Replay, das Paket policy und die Standardbibliothek, keinen Adapter und keinen anderen Dienst.
package hybrid

import (
	"context"
	"slices"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

var _ inbound.HybridUseCase = (*Verteiler)(nil)

// Dienst ist ein Service der Konfiguration: sein Root (SPEC-284 Nr. 1).
type Dienst struct {
	Root string
}

// Konfig ist die Konfiguration des Dienstes (SPEC-284 Nr. 1); MaxBytes ist MAX_CAPTURE_BODY, Policy das Ergebnis von
// policy.New.
type Konfig struct {
	Dienste  []Dienst
	Mock     inbound.MockUseCase
	Replay   inbound.ReplayUseCase
	Policy   *policy.Policy
	MaxBytes int64
}

// Verteiler setzt HybridUseCase um. Er hält nach New nur seine Konfig, die er nicht ändert, keine Sperre und keine
// Goroutine (SPEC-284 Nr. 13).
type Verteiler struct {
	mock     inbound.MockUseCase
	replay   inbound.ReplayUseCase
	policy   *policy.Policy
	maxBytes int64
}

// New prüft die Konfiguration (SPEC-284 Nr. 1) und ruft keinen Port; Form und Überlappung der Roots prüft sie nicht.
func New(k Konfig) (*Verteiler, error) {
	switch {
	case len(k.Dienste) == 0:
		return nil, ungueltig("kein Dienst")
	case slices.ContainsFunc(k.Dienste, func(d Dienst) bool { return d.Root == "" }):
		return nil, ungueltig("Dienst ohne Root")
	case k.Mock == nil || k.Replay == nil || k.Policy == nil:
		return nil, ungueltig("Port oder Policy fehlt")
	case k.MaxBytes < 1:
		return nil, ungueltig("MaxBytes ist kleiner als 1")
	}
	return &Verteiler{mock: k.Mock, replay: k.Replay, policy: k.Policy, maxBytes: k.MaxBytes}, nil
}

// Hybrid ruft den Mock und bei einem freigegebenen Code des Fehlers den Replay-Fallback, je Stufe höchstens einmal
// (SPEC-284 Nr. 3); der Fehler einer Stufe kehrt als derselbe Wert zurück (Nr. 11).
func (v *Verteiler) Hybrid(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
	if ctx == nil || a.Methode == "" || a.Laenge < -1 || a.Body == nil && a.Laenge > 0 {
		return inbound.Antwort{}, ungueltig("Anfrage, die kein HTTP-Adapter liefert")
	}
	anfrageM, anfrageR, fallback := v.vorbereiten(a)
	antwort, err := v.mock.Mock(ctx, anfrageM)
	if err == nil {
		return antwort, nil
	}
	verwirf(antwort)
	code, ok := codeVon(err)
	if !ok || !fallback {
		return inbound.Antwort{}, err
	}
	if v.policy.Entscheide(policy.Anfrage{Methode: a.Methode, Pfad: a.Pfad, Code: code}) != policy.Replay {
		return inbound.Antwort{}, err
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return inbound.Antwort{}, ctxErr
	}
	antwort, err = v.replay.Replay(ctx, anfrageR)
	if err != nil {
		verwirf(antwort)
		return inbound.Antwort{}, err
	}
	return gekennzeichnet(antwort, wertReplay), nil
}
