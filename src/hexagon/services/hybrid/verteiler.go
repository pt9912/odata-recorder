// Package hybrid ist der Hybrid-Dienst: er setzt HybridUseCase um (SPEC-284). Er importiert Domain, die Inbound Ports
// von Mock und Replay, den Outbound Port UpstreamClient, den Querschnitts-Port FallbackTelemetry (SPEC-321), das Paket
// policy und die Standardbibliothek, keinen Adapter und keinen anderen Dienst.
package hybrid

import (
	"context"
	"slices"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
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
// policy.New. Telemetrie ist der Adapter des Prozesses; nil ist ungültig. Der Dienst ruft von ihr StartSpan und
// CountDiscarded und am Span Set und End, nie Log und nie CountError (SPEC-321 Nr. 1).
type Konfig struct {
	Dienste    []Dienst
	Mock       inbound.MockUseCase
	Replay     inbound.ReplayUseCase
	Upstream   outbound.UpstreamClient
	Policy     *policy.Policy
	MaxBytes   int64
	Telemetrie support.FallbackTelemetry
}

// Verteiler setzt HybridUseCase um. Er hält nach New nur seine Konfig, die er nicht ändert, keine Sperre und keine
// Goroutine (SPEC-284 Nr. 13); die Liste der Dienste ist eine Kopie.
type Verteiler struct {
	dienste    []Dienst
	mock       inbound.MockUseCase
	replay     inbound.ReplayUseCase
	upstream   outbound.UpstreamClient
	policy     *policy.Policy
	maxBytes   int64
	telemetrie support.FallbackTelemetry
}

// New prüft die Konfiguration (SPEC-284 Nr. 1, SPEC-321 Nr. 1) und ruft keinen Port; Form und Überlappung der Roots und
// die Form der UpstreamURL prüft sie nicht.
func New(k Konfig) (*Verteiler, error) {
	switch {
	case len(k.Dienste) == 0:
		return nil, ungueltig("kein Dienst")
	case slices.ContainsFunc(k.Dienste, func(d Dienst) bool { return d.Root == "" }):
		return nil, ungueltig("Dienst ohne Root")
	case k.Mock == nil || k.Replay == nil || k.Upstream == nil || k.Policy == nil || k.Telemetrie == nil:
		return nil, ungueltig("Port oder Policy fehlt")
	case k.MaxBytes < 1:
		return nil, ungueltig("MaxBytes ist kleiner als 1")
	}
	return &Verteiler{
		dienste: slices.Clone(k.Dienste), mock: k.Mock, replay: k.Replay, upstream: k.Upstream, policy: k.Policy,
		maxBytes: k.MaxBytes, telemetrie: k.Telemetrie,
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
		return v.replayFallback(ctx, a, st, ziel, odataPfad, code)
	case policy.Upstream:
		return v.upstreamFallback(ctx, a, st.upstream, ziel, odataPfad, code)
	}
	return inbound.Antwort{}, err
}

// replayFallback ruft R im Span der Stufe; liefert R ODR-E5001 und gibt die Policy dafür den Upstream frei, folgt U
// nach dem Ende dieses Spans (SPEC-284 Nr. 3 d, SPEC-321 Nr. 3 b). verworfen ist der Code des Fehlers von M.
func (v *Verteiler) replayFallback(ctx context.Context, a inbound.Anfrage, st stufen, ziel, odataPfad string, verworfen model.Code) (inbound.Antwort, error) {
	antwort, err := v.stufe(ctx, wertReplay, verworfen, func(ctx context.Context) (inbound.Antwort, error) {
		antwort, err := v.replay.Replay(ctx, st.replay)
		if err != nil {
			verwirf(antwort)
			return inbound.Antwort{}, err
		}
		return gekennzeichnet(antwort, wertReplay), nil
	})
	if err != nil {
		if c, ok := codeVon(err); ok && c == model.CodeE5001 &&
			v.policy.Entscheide(policy.Anfrage{Methode: a.Methode, Pfad: a.Pfad, Ziel: ziel, Code: c}) == policy.Upstream {
			return v.upstreamFallback(ctx, a, st.upstream, ziel, odataPfad, c)
		}
		return inbound.Antwort{}, err
	}
	return antwort, nil
}
