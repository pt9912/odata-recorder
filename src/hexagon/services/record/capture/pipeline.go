// Package capture entkoppelt den Proxy-Strom von der Aufzeichnung: ein Tee mit begrenzter
// Queue, begrenzte Dekompression, der Redaktor und der CaptureWriter (SPEC-119). Das Paket
// importiert nur Domain und Outbound Ports.
package capture

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Konfig ist die Konfiguration einer Pipeline (SPEC-119 Nr. 1).
type Konfig struct {
	Blobs          outbound.BlobRepository
	Pfade          []model.JSONPfad
	MaxBytes       int64
	QueueBytes     int64
	Frist          time.Duration
	MaxVerhaeltnis int64
}

// Standard liefert die Konfiguration mit den Werten aus SPEC-053: Queue 16 MiB, Frist 30 s,
// Verhältnis 20:1; maxBytes ist MAX_CAPTURE_BODY.
func Standard(blobs outbound.BlobRepository, pfade []model.JSONPfad, maxBytes int64) Konfig {
	return Konfig{
		Blobs:          blobs,
		Pfade:          pfade,
		MaxBytes:       maxBytes,
		QueueBytes:     16 << 20,
		Frist:          30 * time.Second,
		MaxVerhaeltnis: 20,
	}
}

// Pipeline eröffnet Aufnahmen; sie hält keinen Zustand zwischen ihnen (SPEC-119 Nr. 14).
type Pipeline struct {
	konfig Konfig
	// start hält die Goroutine jeder Aufnahme bis zu seinem Schließen an; nil in jedem Lauf
	// außer im Test (export_test.go).
	start <-chan struct{}
}

func ungueltig(format string, a ...any) error {
	return errors.Join(model.ErrUngueltig, fmt.Errorf(format, a...))
}

// New prüft die Konfiguration (SPEC-119 Nr. 1); eine ungültige ergibt model.ErrUngueltig.
func New(k Konfig) (*Pipeline, error) {
	switch {
	case k.Blobs == nil:
		return nil, ungueltig("Blobs fehlt")
	case k.MaxBytes < 1 || k.QueueBytes < 1 || k.MaxVerhaeltnis < 1:
		return nil, ungueltig("MaxBytes, QueueBytes und MaxVerhaeltnis sind mindestens 1")
	case k.Frist <= 0:
		return nil, ungueltig("Frist ist größer als 0")
	case slices.Contains(k.Pfade, model.JSONPfad{}):
		return nil, ungueltig("Pfadmuster ist der Nullwert")
	}
	k.Pfade = slices.Clone(k.Pfade)
	return &Pipeline{konfig: k}, nil
}

// Eingang beschreibt den Body, den Tee aufnimmt: ContentType und ContentEncoding sind alle
// Werte des jeweiligen Headers, Metadata setzt der Aufrufer wie bei model.BodyEingang und nie
// für einen Request (SPEC-119 Nr. 2).
type Eingang struct {
	Richtung        Richtung
	ContentType     []string
	ContentEncoding []string
	Metadata        bool
}

func pruefeEingang(ctx context.Context, src io.Reader, ein Eingang) error {
	switch {
	case ctx == nil:
		return ungueltig("ctx fehlt")
	case src == nil:
		return ungueltig("src fehlt")
	case ein.Richtung != Request && ein.Richtung != Response:
		return ungueltig("Richtung ist weder Request noch Response")
	case ein.Metadata && ein.Richtung == Request:
		return ungueltig("Metadata gilt nie für einen Request")
	}
	return nil
}

// Tee eröffnet die Aufnahme eines Bodys (SPEC-119 Nr. 2). Der Aufrufer liest den Body aus der
// Aufnahme statt aus src. Bei einem Fehler wird nichts gelesen und keine Goroutine gestartet.
func (p *Pipeline) Tee(ctx context.Context, src io.Reader, ein Eingang) (*Aufnahme, error) {
	if err := pruefeEingang(ctx, src, ein); err != nil {
		return nil, err
	}
	cctx, cancel := context.WithCancelCause(ctx)
	fctx, stop := context.WithTimeoutCause(cctx, p.konfig.Frist, errFrist)
	a := &Aufnahme{
		p:        p,
		src:      src,
		q:        neueQueue(p.konfig.QueueBytes, p.konfig.MaxBytes, func() { cancel(errUeberlauf) }),
		richtung: ein.Richtung,
		eingang:  model.BodyEingang{ContentType: slices.Clone(ein.ContentType), Metadata: ein.Metadata},
		fertig:   make(chan struct{}),
		abbruch: func(ursache error) {
			cancel(ursache)
			stop()
		},
	}
	if kette, ok := parseKette(ein.ContentEncoding); ok {
		a.kette = kette
	} else {
		// Der Body bleibt ungelesen: die Queue nimmt nichts an, die Goroutine meldet das Ergebnis.
		a.q.zu = true
		v := verworfen(ein.Richtung, befund{ausgang: Verstoss, grund: GrundEncoding, ursache: errEncoding})
		a.vorab = &v
	}
	go a.arbeite(fctx)
	return a, nil
}
