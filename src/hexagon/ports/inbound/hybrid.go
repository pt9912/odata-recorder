package inbound

import "context"

// HybridUseCase beantwortet im Modus hybrid einen OData-Request: der Mock antwortet zuerst, und nur wenn er mit einem
// ODR-Code scheitert, den die Fallback-Policy freigibt, antwortet der Replay-Fallback (SPEC-284). Anfrage und Antwort
// sind die Werte von SPEC-121.
type HybridUseCase interface {
	// Hybrid liefert die Antwort des Mock, bei einem freigegebenen Fehler die des Replay-Fallbacks mit dem Header
	// X-Recorder-Fallback; der Adapter ruft Close des Bodys genau einmal. Ein Fehler mit Code ist ein *Fehler, das
	// Fehlerobjekt schreibt der Adapter.
	Hybrid(ctx context.Context, anfrage Anfrage) (Antwort, error)
}
