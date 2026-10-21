package capture_test

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
)

var (
	errQuelle = errors.New("quelle gescheitert")
	errDisk   = errors.New("platte voll")
	errDisk2  = errors.New("platte zweimal voll")
)

// geheim ist der Wert, der in keinem Fehler, keinem Ergebnisfeld und in keinem Byte stehen darf,
// das ein Writer je angeboten bekam.
const geheim = "S3CR3T-7f"

// zeitgrenze begrenzt jedes Warten eines Tests auf die Pipeline.
const zeitgrenze = 10 * time.Second

// --- Blob-Fake nach SPEC-106 Nr. 1 bis 4 --------------------------------------------------

// fakeBlobs ist ein BlobRepository im Speicher: Begin, Write, Commit und Abort mit Zählern,
// Fehlerinjektion an der n-ten Operation, Anhalten einer Operation mit einem Kanal und der
// Aufzeichnung jedes je angebotenen Bytes, auch abgebrochener Writer.
type fakeBlobs struct {
	mu sync.Mutex

	beginFehler  error
	writeFehler  map[int]error
	commitFehler error
	abortFehler  error
	laengeDelta  int64
	// haltWrite hält den n-ten Write (1-based über alle Writer) an: er meldet sich auf imWrite
	// und wartet auf freiWrite.
	haltWrite  int
	imWrite    chan struct{}
	freiWrite  chan struct{}
	haltCommit bool
	imCommit   chan struct{}
	freiCommit chan struct{}
	// haltBegin hält Begin an: er meldet sich auf imBegin und wartet auf freiBegin; danach gilt
	// beginFehler.
	haltBegin bool
	imBegin   chan struct{}
	freiBegin chan struct{}
	// warteAufCtx lässt jeden Write, warteBegin jedes Begin und warteCommit jedes Commit auf das
	// Ende des Kontexts warten, den Begin bekam, und dessen Fehler zurückgeben; die Aufrufe melden
	// sich vorher auf imWrite, imBegin und imCommit.
	warteAufCtx bool
	warteBegin  bool
	warteCommit bool
	// nachCtx hält einen auf den Kontext wartenden Write an, nachdem der Kontext endete: er
	// meldet das auf ctxEnde und wartet auf nachCtx.
	nachCtx chan struct{}
	ctxEnde chan struct{}
	// fehlerNachHalt gibt den Fehler von writeFehler erst zurück, nachdem ein angehaltener Write
	// freigegeben ist.
	fehlerNachHalt bool

	begins, commits, aborts, writeNr int
	grenzen                          []int64
	ctxs                             []context.Context
	committed                        map[string][]byte
	angeboten                        []byte
	meldungen                        []string
	nWriter                          int
}

func neueBlobs() *fakeBlobs {
	return &fakeBlobs{
		writeFehler: map[int]error{},
		committed:   map[string][]byte{},
		imWrite:     make(chan struct{}, 64),
		freiWrite:   make(chan struct{}),
		imCommit:    make(chan struct{}, 64),
		freiCommit:  make(chan struct{}),
		imBegin:     make(chan struct{}, 64),
		freiBegin:   make(chan struct{}),
		ctxEnde:     make(chan struct{}, 64),
	}
}

type fakeWriter struct {
	b *fakeBlobs
	// done und cerr tragen den Kontext von Begin.
	done   <-chan struct{}
	cerr   func() error
	grenze int64
	buf    []byte
	aktiv  atomic.Int32
	fertig atomic.Bool
	// abgebrochen ist der Writer nach ErrZuGross (SPEC-106 Nr. 2).
	abgebrochen bool
}

func refVon(b []byte) string {
	h := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(h[:])
}

func (b *fakeBlobs) melde(format string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.meldungen = append(b.meldungen, format)
}

func (b *fakeBlobs) Begin(ctx context.Context, grenze int64) (outbound.CaptureWriter, error) {
	b.mu.Lock()
	b.begins++
	b.grenzen = append(b.grenzen, grenze)
	b.ctxs = append(b.ctxs, ctx)
	fehler, halt, warte := b.beginFehler, b.haltBegin, b.warteBegin
	b.mu.Unlock()
	switch {
	case warte:
		b.imBegin <- struct{}{}
		<-ctx.Done()
		return nil, ctx.Err()
	case halt:
		b.imBegin <- struct{}{}
		<-b.freiBegin
	}
	if fehler != nil {
		return nil, fehler
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nWriter++
	return &fakeWriter{b: b, done: ctx.Done(), cerr: ctx.Err, grenze: grenze}, nil
}

func (b *fakeBlobs) Open(context.Context, string) (io.ReadCloser, error) {
	b.melde("Open gerufen")
	return nil, outbound.ErrBlobFehlt
}

func (b *fakeBlobs) Verify(context.Context, string) error {
	b.melde("Verify gerufen")
	return outbound.ErrBlobFehlt
}

func (b *fakeBlobs) List(context.Context) ([]string, error) {
	b.melde("List gerufen")
	return nil, nil
}

func (b *fakeBlobs) Remove(context.Context, string) error {
	b.melde("Remove gerufen")
	return nil
}

// betritt meldet einen Aufruf, der einen laufenden überlappt oder nach Commit oder Abort kommt.
func (w *fakeWriter) betritt(name string) func() {
	if w.aktiv.Add(1) > 1 {
		w.b.melde(name + " überlappt einen laufenden Aufruf")
	}
	if w.fertig.Load() {
		w.b.melde(name + " nach Commit oder Abort")
	}
	return func() { w.aktiv.Add(-1) }
}

func (w *fakeWriter) Write(p []byte) (int, error) {
	defer w.betritt("Write")()
	b := w.b
	b.mu.Lock()
	b.writeNr++
	nr := b.writeNr
	err := b.writeFehler[nr]
	halt, warte, spaet := b.haltWrite == nr, b.warteAufCtx, b.fehlerNachHalt && b.haltWrite == nr
	nachCtx := b.nachCtx
	b.angeboten = append(b.angeboten, p...)
	b.mu.Unlock()
	if err != nil && !spaet {
		return 0, err
	}
	if cerr := w.cerr(); cerr != nil {
		if warte && nachCtx != nil {
			b.ctxEnde <- struct{}{}
			<-nachCtx
		}
		return 0, cerr
	}
	if warte {
		b.imWrite <- struct{}{}
		<-w.done
		if nachCtx != nil {
			b.ctxEnde <- struct{}{}
			<-nachCtx
		}
		return 0, w.cerr()
	}
	if halt {
		b.imWrite <- struct{}{}
		<-b.freiWrite
	}
	if err != nil {
		return 0, err
	}
	if w.abgebrochen {
		return 0, outbound.ErrAbgeschlossen
	}
	if int64(len(w.buf)+len(p)) > w.grenze {
		w.abgebrochen, w.buf = true, nil
		return 0, outbound.ErrZuGross
	}
	w.buf = append(w.buf, p...)
	return len(p), nil
}

func (w *fakeWriter) Commit() (string, int64, error) {
	defer w.betritt("Commit")()
	b := w.b
	b.mu.Lock()
	b.commits++
	fehler, halt, delta, warte := b.commitFehler, b.haltCommit, b.laengeDelta, b.warteCommit
	b.mu.Unlock()
	if cerr := w.cerr(); cerr != nil {
		return "", 0, cerr
	}
	if warte {
		b.imCommit <- struct{}{}
		<-w.done
		return "", 0, w.cerr()
	}
	if halt {
		b.imCommit <- struct{}{}
		<-b.freiCommit
	}
	if w.abgebrochen {
		return "", 0, outbound.ErrAbgeschlossen
	}
	if fehler != nil {
		return "", 0, fehler
	}
	ref := refVon(w.buf)
	b.mu.Lock()
	b.committed[ref] = bytes.Clone(w.buf)
	b.mu.Unlock()
	w.fertig.Store(true)
	return ref, int64(len(w.buf)) + delta, nil
}

func (w *fakeWriter) Abort() error {
	defer w.betritt("Abort")()
	b := w.b
	b.mu.Lock()
	b.aborts++
	err := b.abortFehler
	b.mu.Unlock()
	w.fertig.Store(true)
	return err
}

type stand struct{ begins, commits, aborts, writer int }

func (b *fakeBlobs) stand() stand {
	b.mu.Lock()
	defer b.mu.Unlock()
	return stand{b.begins, b.commits, b.aborts, b.nWriter}
}

// inhalte liefert die Inhalte aller gelungenen Commits.
func (b *fakeBlobs) inhalte() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []string
	for _, v := range b.committed {
		out = append(out, string(v))
	}
	return out
}

func (b *fakeBlobs) inhalt(ref string) (string, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	v, ok := b.committed[ref]
	return string(v), ok
}

func (b *fakeBlobs) alleAngeboten() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.angeboten)
}

func (b *fakeBlobs) verstoesse() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]string(nil), b.meldungen...)
}

// --- skriptgesteuertes src -----------------------------------------------------------------

// schritt ist ein Read von src: daten (auch leer) und err kommen zurück; warte hält den Read
// vor der Lieferung an, ankunft meldet, dass er angekommen ist.
type schritt struct {
	daten   string
	err     error
	warte   <-chan struct{}
	ankunft chan<- struct{}
}

// skript liefert eine Folge von Reads und danach ende (io.EOF, wenn nicht gesetzt).
type skript struct {
	mu       sync.Mutex
	schritte []schritt
	i        int
	reads    int
	ende     error
}

func quelle(teile ...string) *skript {
	s := &skript{}
	for _, t := range teile {
		s.schritte = append(s.schritte, schritt{daten: t})
	}
	return s
}

// quelleMit endet nach den Teilen mit ende statt mit io.EOF.
func quelleMit(ende error, teile ...string) *skript {
	s := quelle(teile...)
	s.ende = ende
	return s
}

func quelleBytes(b []byte) *skript { return quelle(string(b)) }

// zerlegt liefert b in Reads zu je n Bytes.
func zerlegt(b []byte, n int) *skript {
	s := &skript{}
	for len(b) > 0 {
		k := min(n, len(b))
		s.schritte = append(s.schritte, schritt{daten: string(b[:k])})
		b = b[k:]
	}
	return s
}

// haltVor hängt einen Schritt an, der wartet, bis warte geschlossen ist, und dann daten liefert.
func (s *skript) haltVor(warte <-chan struct{}, ankunft chan<- struct{}, daten string, err error) *skript {
	s.schritte = append(s.schritte, schritt{daten: daten, err: err, warte: warte, ankunft: ankunft})
	return s
}

func (s *skript) anzahlReads() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reads
}

func (s *skript) Read(p []byte) (int, error) {
	s.mu.Lock()
	s.reads++
	if s.i >= len(s.schritte) {
		e := s.ende
		s.mu.Unlock()
		if e == nil {
			e = io.EOF
		}
		return 0, e
	}
	st := s.schritte[s.i]
	s.mu.Unlock()
	if st.ankunft != nil {
		st.ankunft <- struct{}{}
	}
	if st.warte != nil {
		<-st.warte
	}
	n := copy(p, st.daten)
	s.mu.Lock()
	defer s.mu.Unlock()
	if n < len(st.daten) {
		s.schritte[s.i].daten = st.daten[n:]
		return n, nil
	}
	s.i++
	return n, st.err
}

// --- Umgebung ------------------------------------------------------------------------------

// umg ist die Umgebung eines Tests: Blob-Fake, Konfiguration und die daraus gebaute Pipeline.
type umg struct {
	t     *testing.T
	blobs *fakeBlobs
	k     capture.Konfig
	p     *capture.Pipeline
}

func pfad(t *testing.T, text string) model.JSONPfad {
	t.Helper()
	p, err := model.ParseJSONPfad(text)
	if err != nil {
		t.Fatalf("ParseJSONPfad(%q): %v", text, err)
	}
	return p
}

// neu baut die Umgebung mit kleinen Werten: Pfad $..password, MaxBytes und QueueBytes 1 MiB,
// Frist 10 s, Verhältnis 20. Jede Meldung des Fakes über einen überlappenden oder späten Aufruf
// ist am Ende ein Fehler.
func neu(t *testing.T, opts ...func(*capture.Konfig)) *umg {
	t.Helper()
	b := neueBlobs()
	u := &umg{t: t, blobs: b}
	u.k = capture.Konfig{
		Blobs:          b,
		Pfade:          []model.JSONPfad{pfad(t, "$..password")},
		MaxBytes:       1 << 20,
		QueueBytes:     1 << 20,
		Frist:          zeitgrenze,
		MaxVerhaeltnis: 20,
	}
	for _, o := range opts {
		o(&u.k)
	}
	t.Cleanup(func() {
		for _, m := range b.verstoesse() {
			t.Errorf("Blob-Fake: %s", m)
		}
	})
	return u
}

func (u *umg) pipe() *capture.Pipeline {
	u.t.Helper()
	if u.p == nil {
		p, err := capture.New(u.k)
		if err != nil {
			u.t.Fatalf("New: %v", err)
		}
		u.p = p
	}
	return u.p
}

func (u *umg) tee(src io.Reader, ein capture.Eingang) *capture.Aufnahme {
	u.t.Helper()
	a, err := u.pipe().Tee(context.Background(), src, ein)
	if err != nil {
		u.t.Fatalf("Tee: %v", err)
	}
	return a
}

// lauf nimmt src auf, liest den Strom leer und liefert das Ergebnis.
func (u *umg) lauf(src io.Reader, ein capture.Eingang) capture.Ergebnis {
	u.t.Helper()
	a := u.tee(src, ein)
	_, _ = io.Copy(io.Discard, a)
	return abwarten(u.t, a)
}

// abwarten liefert das Ergebnis oder bricht den Test nach der Zeitgrenze ab.
func abwarten(t *testing.T, a *capture.Aufnahme) capture.Ergebnis {
	t.Helper()
	ch := make(chan capture.Ergebnis, 1)
	go func() { ch <- a.Ergebnis() }()
	select {
	case e := <-ch:
		return e
	case <-time.After(zeitgrenze):
		t.Fatal("Ergebnis() kehrt nicht zurück")
	}
	return capture.Ergebnis{}
}

func mitMax(n int64) func(*capture.Konfig)   { return func(k *capture.Konfig) { k.MaxBytes = n } }
func mitQueue(n int64) func(*capture.Konfig) { return func(k *capture.Konfig) { k.QueueBytes = n } }
func mitVerh(n int64) func(*capture.Konfig)  { return func(k *capture.Konfig) { k.MaxVerhaeltnis = n } }
func mitFrist(d time.Duration) func(*capture.Konfig) {
	return func(k *capture.Konfig) { k.Frist = d }
}
func ohnePfade() func(*capture.Konfig) { return func(k *capture.Konfig) { k.Pfade = nil } }

// --- Eingänge ------------------------------------------------------------------------------

func antwort() capture.Eingang {
	return capture.Eingang{Richtung: capture.Response, ContentType: []string{"application/json"}}
}

func anfrage() capture.Eingang {
	return capture.Eingang{Richtung: capture.Request, ContentType: []string{"application/json"}}
}

func mitEncoding(ein capture.Eingang, werte ...string) capture.Eingang {
	ein.ContentEncoding = werte
	return ein
}

// --- Kompression von Hand ------------------------------------------------------------------

func gz(s string) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, _ = w.Write([]byte(s))
	_ = w.Close()
	return b.Bytes()
}

func zl(b []byte) []byte {
	var out bytes.Buffer
	w := zlib.NewWriter(&out)
	_, _ = w.Write(b)
	_ = w.Close()
	return out.Bytes()
}

func rohDeflate(s string) []byte {
	var out bytes.Buffer
	w, _ := flate.NewWriter(&out, flate.DefaultCompression)
	_, _ = w.Write([]byte(s))
	_ = w.Close()
	return out.Bytes()
}

// --- Prüfungen -----------------------------------------------------------------------------

// verworfen prüft die Felder eines Ergebnisses mit Ausgang Abbruch oder Verstoss (SPEC-119 Nr. 11).
func verworfen(t *testing.T, e capture.Ergebnis, r capture.Richtung, aus capture.Ausgang, g capture.Grund) {
	t.Helper()
	if e.Richtung != r || e.Ausgang != aus || e.Grund != g {
		t.Errorf("Ergebnis = %v/%v/%q, erwartet %v/%v/%q", e.Richtung, e.Ausgang, e.Grund, r, aus, g)
	}
	if e.Body != model.BodyDiscarded || e.Blob != "" || e.Redaction != model.RedactionUnsafe ||
		e.Integrity != model.IntegrityAbsent || e.Redigiert != nil {
		t.Errorf("Felder eines verworfenen Bodys falsch: %+v", e)
	}
	if e.Fehler == nil {
		t.Fatalf("Fehler fehlt bei %v/%q", aus, g)
	}
	if e.Fehler.Grund != g || e.Fehler.Code() != model.CodeE2401 || e.Fehler.Code().String() != "ODR-E2401" {
		t.Errorf("Fehler = %q/%v, erwartet %q/ODR-E2401", e.Fehler.Grund, e.Fehler.Code(), g)
	}
	// Der Text nennt Code und Grund und sonst nichts (SPEC-119 Nr. 11).
	if want := "ODR-E2401: Capture fehlgeschlagen: " + string(g); e.Fehler.Error() != want {
		t.Errorf("Error() = %q, erwartet %q", e.Fehler.Error(), want)
	}
}

// vollstaendig prüft ein Ergebnis mit Ausgang Vollstaendig und Body present und liefert den Inhalt des Blobs.
func vollstaendig(t *testing.T, u *umg, e capture.Ergebnis, red model.RedactionStatus) string {
	t.Helper()
	if e.Ausgang != capture.Vollstaendig || e.Grund != "" || e.Fehler != nil {
		t.Fatalf("Ergebnis = %v/%q/%v, erwartet Vollstaendig", e.Ausgang, e.Grund, e.Fehler)
	}
	if e.Body != model.BodyPresent || e.Redaction != red || e.Integrity != model.IntegrityVerified {
		t.Errorf("Felder = %v/%v/%v, erwartet present/%v/verified", e.Body, e.Redaction, e.Integrity, red)
	}
	inhalt, ok := u.blobs.inhalt(e.Blob)
	if !ok {
		t.Fatalf("Blob %q ist nicht committet", e.Blob)
	}
	return inhalt
}

// ohneBody prüft ein Ergebnis mit Ausgang Vollstaendig und Body absent.
func ohneBody(t *testing.T, e capture.Ergebnis) {
	t.Helper()
	if e.Ausgang != capture.Vollstaendig || e.Body != model.BodyAbsent || e.Blob != "" ||
		e.Redaction != model.RedactionNone || e.Integrity != model.IntegrityAbsent || e.Fehler != nil || e.Redigiert != nil {
		t.Errorf("kein absent-Ergebnis: %+v", e)
	}
}

func wiederholt(s string, n int) string {
	var b bytes.Buffer
	for range n {
		b.WriteString(s)
	}
	return b.String()
}

func errorsIsUngueltig(err error) bool { return errors.Is(err, model.ErrUngueltig) }

// writes liefert die Zahl der Write-Aufrufe über alle Writer.
func (b *fakeBlobs) writes() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.writeNr
}

// gzMitRoh liefert einen gzip-Strom von s, der genau roh Bytes lang ist: ein Kommentar im Kopf
// füllt auf (länger als der Strom ohne Kommentar plus 2 Bytes: Zeichen und Abschluss).
func gzMitRoh(t *testing.T, s string, roh int) []byte {
	t.Helper()
	basis := gz(s)
	if roh == len(basis) {
		return basis
	}
	if roh < len(basis)+2 {
		t.Fatalf("gzMitRoh: %d Bytes sind zu kurz, der Strom ohne Kommentar hat %d", roh, len(basis))
	}
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Comment = wiederholt("c", roh-len(basis)-1)
	_, _ = w.Write([]byte(s))
	_ = w.Close()
	if b.Len() != roh {
		t.Fatalf("gzMitRoh: %d Bytes statt %d", b.Len(), roh)
	}
	return b.Bytes()
}

// kontexte liefert die Kontexte aller Begin-Aufrufe.
func (b *fakeBlobs) kontexte() []context.Context {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]context.Context(nil), b.ctxs...)
}
