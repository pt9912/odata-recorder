// Package hybrid ist der Hybrid-Dienst: er setzt HybridUseCase um (SPEC-284). Er importiert Domain, die Inbound Ports
// von Mock und Replay, den Outbound Port UpstreamClient, das Paket policy und die Standardbibliothek, keinen Adapter und
// keinen anderen Dienst.
package hybrid

import (
	"context"
	"slices"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

var _ inbound.HybridUseCase = (*Verteiler)(nil)

// Dienst ist ein Service der Konfiguration: sein Root und seine UpstreamURL, die der Dienst als Ziel an die Policy und
// an Senden reicht (SPEC-284 Nr. 1, 4); eine leere UpstreamURL gibt die Policy nie frei.
type Dienst struct {
	Root        string
	UpstreamURL string
}

// Konfig ist die Konfiguration des Dienstes (SPEC-284 Nr. 1); MaxBytes ist MAX_CAPTURE_BODY, Policy das Ergebnis von
// policy.New.
type Konfig struct {
	Dienste  []Dienst
	Mock     inbound.MockUseCase
	Replay   inbound.ReplayUseCase
	Upstream outbound.UpstreamClient
	Policy   *policy.Policy
	MaxBytes int64
}

// Verteiler setzt HybridUseCase um. Er hält nach New nur seine Konfig, die er nicht ändert, keine Sperre und keine
// Goroutine (SPEC-284 Nr. 13); die Liste der Dienste ist eine Kopie.
type Verteiler struct {
	dienste  []Dienst
	mock     inbound.MockUseCase
	replay   inbound.ReplayUseCase
	upstream outbound.UpstreamClient
	policy   *policy.Policy
	maxBytes int64
}

// New prüft die Konfiguration (SPEC-284 Nr. 1) und ruft keinen Port; Form und Überlappung der Roots und die Form der
// UpstreamURL prüft sie nicht.
func New(k Konfig) (*Verteiler, error) {
	switch {
	case len(k.Dienste) == 0:
		return nil, ungueltig("kein Dienst")
	case slices.ContainsFunc(k.Dienste, func(d Dienst) bool { return d.Root == "" }):
		return nil, ungueltig("Dienst ohne Root")
	case k.Mock == nil || k.Replay == nil || k.Upstream == nil || k.Policy == nil:
		return nil, ungueltig("Port oder Policy fehlt")
	case k.MaxBytes < 1:
		return nil, ungueltig("MaxBytes ist kleiner als 1")
	}
	return &Verteiler{
		dienste: slices.Clone(k.Dienste), mock: k.Mock, replay: k.Replay, upstream: k.Upstream, policy: k.Policy,
		maxBytes: k.MaxBytes,
	}, nil
}

// Hybrid ruft den Mock und bei einem freigegebenen Code des Fehlers den Replay- oder Upstream-Fallback, je Stufe
// höchstens einmal (SPEC-284 Nr. 3); der Fehler von Mock und Replay kehrt als derselbe Wert zurück (Nr. 11).
func (v *Verteiler) Hybrid(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
	if ctx == nil || a.Methode == "" || a.Laenge < -1 || a.Body == nil && a.Laenge > 0 {
		return inbound.Antwort{}, ungueltig("Anfrage, die kein HTTP-Adapter liefert")
	}
	ziel, odataPfad := v.waehle(a.Pfad)
	st := v.vorbereiten(a, ziel)
	antwort, err := v.mock.Mock(ctx, st.mock)
	if err == nil {
		return antwort, nil
	}
	verwirf(antwort)
	code, ok := codeVon(err)
	if !ok || !st.fallback {
		return inbound.Antwort{}, err
	}
	switch v.policy.Entscheide(policy.Anfrage{Methode: a.Methode, Pfad: a.Pfad, Ziel: ziel, Code: code}) {
	case policy.Replay:
		return v.replayFallback(ctx, a, st, ziel, odataPfad)
	case policy.Upstream:
		return v.upstreamFallback(ctx, a, st.upstream, ziel, odataPfad)
	}
	return inbound.Antwort{}, err
}

// replayFallback ruft R; liefert R ODR-E5001 und gibt die Policy dafür den Upstream frei, folgt U (SPEC-284 Nr. 3 d).
func (v *Verteiler) replayFallback(ctx context.Context, a inbound.Anfrage, st stufen, ziel, odataPfad string) (inbound.Antwort, error) {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return inbound.Antwort{}, ctxErr
	}
	antwort, err := v.replay.Replay(ctx, st.replay)
	if err != nil {
		verwirf(antwort)
		if c, ok := codeVon(err); ok && c == model.CodeE5001 &&
			v.policy.Entscheide(policy.Anfrage{Methode: a.Methode, Pfad: a.Pfad, Ziel: ziel, Code: c}) == policy.Upstream {
			return v.upstreamFallback(ctx, a, st.upstream, ziel, odataPfad)
		}
		return inbound.Antwort{}, err
	}
	return gekennzeichnet(antwort, wertReplay), nil
}
