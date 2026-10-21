package record

import (
	"context"
	"io"
	"sync"
	"sync/atomic"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
)

// austausch ist ein angenommener Exchange von der Annahme bis zum Ende des Anhängens (SPEC-121 Nr. 9).
type austausch struct {
	r       *Recorder
	sz      *sessionZustand
	l       *lauf
	dienst  Dienst
	anfrage inbound.Anfrage
	vorab   model.Schluessel
	e       model.Exchange
	// client liefert den Fehler des Kontexts von Record (SPEC-121 Nr. 11).
	client func() error
	// req ist die Request-Aufnahme samt Merker für io.EOF; nil ohne Aufnahme (SPEC-121 Nr. 5).
	req    *eofLeser
	reqAuf *capture.Aufnahme
	// res ist die Response-Aufnahme; nil bei einer Antwort ohne Body nach HTTP.
	res    *capture.Aufnahme
	upBody io.ReadCloser
}

// eofLeser merkt sich, ob ein Read io.EOF lieferte, auch zusammen mit n > 0 (SPEC-121 Nr. 8).
type eofLeser struct {
	r   io.Reader
	eof atomic.Bool
}

func (l *eofLeser) Read(p []byte) (int, error) {
	n, err := l.r.Read(p)
	if err == io.EOF {
		l.eof.Store(true)
	}
	return n, err
}

// leite nimmt den Request-Body auf, sendet an den Upstream und eröffnet die Response-Aufnahme
// (SPEC-121 Nr. 4, 5, 7); actx ist der Kontext des Anhängens (Nr. 9).
func (x *austausch) leite(ctx, actx context.Context) (inbound.Antwort, error) {
	body, laenge, err := x.nimmRequestAuf(ctx)
	if err != nil {
		return inbound.Antwort{}, x.ohneTee(actx, err)
	}
	a := x.anfrage
	ant, err := x.r.k.Upstream.Senden(ctx, outbound.UpstreamAnfrage{
		Ziel: x.dienst.UpstreamURL, Methode: a.Methode, Pfad: x.e.Path, RawQuery: a.RawQuery,
		Header: weiterzuleiten(a.Header), Body: body, Laenge: laenge,
	})
	if err != nil {
		return inbound.Antwort{}, x.ohneAntwort(actx, err)
	}
	x.upBody = ant.Body
	x.e.ResponseStatus = ant.Status
	ohne := ohneHopByHop(ant.Header)
	x.e.ResponseHeaders = model.RedigiereHeader(ohne, x.r.k.Regeln.RedactHeaders, x.r.k.Regeln.RedactQueryParams)
	src := io.Reader(ant.Body)
	if hatBody(a.Methode, ant.Status) {
		x.res, err = x.r.k.Pipeline.Tee(ctx, ant.Body, capture.Eingang{
			Richtung:        capture.Response,
			ContentType:     werteVon(ant.Header, "content-type"),
			ContentEncoding: werteVon(ant.Header, "content-encoding"),
			Metadata:        x.e.Path == "/$metadata",
		})
		if err != nil {
			return inbound.Antwort{}, x.ohneTee(actx, err)
		}
		src = x.res
	}
	b := &antwortBody{src: &eofLeser{r: src}, ende: func(eof bool) { x.beende(actx, eof) }}
	return inbound.Antwort{Status: ant.Status, Header: ohne, Body: b}, nil
}

// nimmRequestAuf gibt den Body durch Tee; ohne Body oder bei Laenge 0 gibt es keine Aufnahme,
// und Senden erhält keinen Body und Laenge 0 (SPEC-121 Nr. 5).
func (x *austausch) nimmRequestAuf(ctx context.Context) (io.Reader, int64, error) {
	a := x.anfrage
	if a.Body == nil || a.Laenge == 0 {
		return nil, 0, nil
	}
	auf, err := x.r.k.Pipeline.Tee(ctx, a.Body, capture.Eingang{
		Richtung:        capture.Request,
		ContentType:     werteVon(a.Header, "content-type"),
		ContentEncoding: werteVon(a.Header, "content-encoding"),
	})
	if err != nil {
		return nil, 0, err
	}
	x.reqAuf, x.req = auf, &eofLeser{r: auf}
	return x.req, a.Laenge, nil
}

// hatBody meldet, ob eine Antwort nach HTTP einen Body hat: nicht auf HEAD (byte-gleich), nicht
// bei 1xx, 204 und 304 (SPEC-121 Nr. 5).
func hatBody(methode string, status int) bool {
	return methode != "HEAD" && (status < 100 || status > 199) && status != 204 && status != 304
}

// antwortBody ist der Body der Antwort an den Adapter: Read reicht den Strom unverändert durch,
// Close beendet den Exchange genau einmal und liefert nil (SPEC-121 Nr. 7).
type antwortBody struct {
	src    *eofLeser
	zu     atomic.Bool
	einmal sync.Once
	ende   func(eof bool)
}

func (b *antwortBody) Read(p []byte) (int, error) {
	if b.zu.Load() {
		return 0, io.ErrClosedPipe
	}
	return b.src.Read(p)
}

func (b *antwortBody) Close() error {
	b.einmal.Do(func() {
		b.zu.Store(true)
		b.ende(b.src.eof.Load())
	})
	return nil
}

// brichRequestAb ruft Abbrechen der Request-Aufnahme, wenn kein Read io.EOF lieferte (SPEC-121 Nr. 8b).
func (x *austausch) brichRequestAb() {
	if x.req != nil && !x.req.eof.Load() {
		x.reqAuf.Abbrechen()
	}
}

// requestErgebnis wartet auf die Request-Aufnahme; ohne Aufnahme ist das Ergebnis das ohne Body.
func (x *austausch) requestErgebnis() capture.Ergebnis {
	if x.reqAuf == nil {
		return ohneAufnahme(capture.Request)
	}
	return x.reqAuf.Ergebnis()
}

func ohneAufnahme(r capture.Richtung) capture.Ergebnis {
	return capture.Ergebnis{
		Richtung: r, Ausgang: capture.Vollstaendig,
		Body: model.BodyAbsent, Redaction: model.RedactionNone, Integrity: model.IntegrityAbsent,
	}
}

// beende ist das Ende bei Close in der Reihenfolge von SPEC-121 Nr. 8.
func (x *austausch) beende(actx context.Context, antwortEOF bool) {
	if x.res != nil && !antwortEOF {
		x.res.Abbrechen()
	}
	x.brichRequestAb()
	_ = x.upBody.Close()
	req := x.requestErgebnis()
	res := ohneAufnahme(capture.Response)
	if x.res != nil {
		res = x.res.Ergebnis()
	}
	x.e.FinishedAt = x.r.k.Uhr.Now()
	z, err := capture.Abschluss(req, res)
	x.protokolliereCapture(actx, req, res)
	x.schliesseAb(actx, z, req.Redigiert, err)
}

// ohneAntwort ist das Ende, wenn Senden scheitert (SPEC-121 Nr. 11): der Exchange wird ohne
// Antwort angehängt, der Fehler folgt der Rangfolge Kontext, Drain, Zeit, nicht erreichbar, andere.
func (x *austausch) ohneAntwort(actx context.Context, sendenFehler error) error {
	x.brichRequestAb()
	req := x.requestErgebnis()
	x.e.FinishedAt = x.r.k.Uhr.Now()
	x.protokolliereCapture(actx, req)
	x.schliesseAb(actx, ohneAntwort(req), req.Redigiert, nil)
	return x.fehlerOhneAntwort(actx, sendenFehler)
}

// ohneAntwort ist der Zustand eines Exchange ohne Antwort aus dem Ergebnis des Requests (SPEC-121 Nr. 11).
func ohneAntwort(req capture.Ergebnis) capture.Zustand {
	z := capture.Zustand{
		Capture:        model.CaptureIncomplete,
		KeyMoeglich:    req.Ausgang == capture.Vollstaendig,
		Integrity:      model.IntegrityAbsent,
		Equivalence:    model.EquivalenceNone,
		RequestBody:    req.Body,
		RequestBlob:    req.Blob,
		RequestRedact:  req.Redaction,
		ResponseBody:   model.BodyAbsent,
		ResponseRedact: model.RedactionNone,
	}
	if !z.KeyMoeglich {
		z.Capture = model.CaptureFailed
	}
	if req.Body == model.BodyPresent {
		z.Integrity = model.IntegrityVerified
	}
	return z
}

// ohneTee ist das Ende nach einem Fehler von Tee: eine laufende Request-Aufnahme endet, der Exchange
// ist nicht angehängt, Record liefert ODR-E9001. Dorthin führt nur ein Body nil aus Senden; kein Test
// erreicht den Zweig (SPEC-121 Ungeprüft).
func (x *austausch) ohneTee(actx context.Context, err error) error {
	x.brichRequestAb()
	x.requestErgebnis()
	x.schliesseAb(actx, capture.Zustand{}, nil, err)
	return mitCode(model.CodeE9001, err)
}

// schliesseAb bildet den Match-Key und hängt an, ist err nil; sonst und bei jedem Fehler dabei ist
// der Exchange nicht angehängt (SPEC-121 Nr. 6, 10). Danach läuft er nicht mehr. Die Tests erreichen
// von diesen Fehlern die des Anhängens; ein err aus Abschluss ist nicht erreichbar (SPEC-121 Nr. 8),
// ein Fehler des zweiten BildeSchluessel ebenso nicht (SPEC-121 Nr. 6).
func (x *austausch) schliesseAb(actx context.Context, z capture.Zustand, redigiert []byte, err error) {
	if err == nil {
		x.setzeZustand(z)
		err = x.bildeKey(z, redigiert)
	}
	if err == nil {
		err = x.r.haengeAn(actx, x.sz, x.e)
	}
	if err != nil {
		x.protokolliereAblage(actx, err)
	}
	x.r.endeLauf(x.sz, x.l, err != nil)
}

func (x *austausch) setzeZustand(z capture.Zustand) {
	x.e.Capture, x.e.Integrity, x.e.Equivalence = z.Capture, z.Integrity, z.Equivalence
	x.e.RequestBody, x.e.RequestBlob, x.e.RequestRedact = z.RequestBody, z.RequestBlob, z.RequestRedact
	x.e.ResponseBody, x.e.ResponseBlob, x.e.ResponseRedact = z.ResponseBody, z.ResponseBlob, z.ResponseRedact
}

// bildeKey bildet den Key aus dem redigierten Request-Body, wenn KeyMoeglich; sonst bleibt er
// leer und KeyRedacted ist der Wert ohne Body (SPEC-121 Nr. 6).
func (x *austausch) bildeKey(z capture.Zustand, redigiert []byte) error {
	if !z.KeyMoeglich {
		x.e.MatchKey, x.e.KeyRedacted = "", x.vorab.KeyRedacted
		return nil
	}
	a := x.anfrage
	k, err := model.BildeSchluessel(model.SchluesselEingabe{
		Method: a.Methode, ServiceID: x.dienst.ID, Path: x.e.Path, RawQuery: a.RawQuery, Header: a.Header,
		Body: redigiert, RequestRedaction: z.RequestRedact,
	}, x.r.k.Regeln)
	if err != nil {
		return err
	}
	x.e.MatchKey, x.e.KeyRedacted = k.MatchKey, k.KeyRedacted
	return nil
}
