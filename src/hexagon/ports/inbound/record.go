// Package inbound enthält die Inbound Ports: die Use Cases, die die Driving Adapter rufen, und
// ihren Fehlerwert.
package inbound

import (
	"context"
	"io"
	"time"
)

// RecordUseCase leitet im Modus record einen Request an den Upstream seines Service weiter und
// zeichnet den Austausch auf; Drain hält die Aufzeichnung einer Session an (SPEC-121).
type RecordUseCase interface {
	// Record kehrt mit dem Kopf der Upstream-Antwort zurück; Close des Bodys beendet den Exchange.
	Record(ctx context.Context, anfrage Anfrage) (Antwort, error)
	// Drain nimmt für die Session keinen Exchange mehr an und wartet höchstens frist auf die laufenden.
	Drain(ctx context.Context, sessionID string, frist time.Duration) (DrainBericht, error)
}

// Anfrage füllt der HTTP-Adapter unverändert (SPEC-121): Pfad mit seinen Escapes und ohne Query,
// RawQuery ohne `?`, Header ohne Host, Body nil heißt keiner, Laenge -1 unbekannt und 0 keiner.
type Anfrage struct {
	Methode        string
	Pfad           string
	RawQuery       string
	KorrelationsID string
	Header         map[string][]string
	Body           io.Reader
	Laenge         int64
}

// Antwort ist der Kopf der Upstream-Antwort und ihr Body; der Adapter ruft Close genau einmal.
type Antwort struct {
	Status int
	Header map[string][]string
	Body   io.ReadCloser
}

// DrainBericht zählt je Session über die Lebenszeit des Dienstes (SPEC-121 Nr. 13).
type DrainBericht struct {
	Abgebrochen     int
	NichtAngehaengt int
}
