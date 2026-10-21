package capture_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
)

// fall ist eine Zeile der Tabelle captureFaelle: Eingabe und Erwartung stehen von Hand daneben.
// Eine Zeile mit treibe hängt von der Zeit oder vom Anhalten einer Goroutine ab und gehört nicht
// zu den Zeilen, die TestZerlegungOhneWirkung zerlegt.
type fall struct {
	name string
	ein  capture.Eingang
	roh  []byte
	// ende ersetzt das io.EOF von src (Fehler nach den Rohbytes).
	ende    error
	opt     []func(*capture.Konfig)
	blobs   func(*fakeBlobs)
	treibe  func(t *testing.T, u *umg, f fall) capture.Ergebnis
	ausgang capture.Ausgang
	grund   capture.Grund
	// ursache ist die Ursache, die errors.Is am Fehler erreicht.
	ursache error
	// writer ist die Zahl der Writer, die Begin lieferte; commits die Zahl der Commit-Aufrufe.
	writer, commits int
	// Erwartung bei Vollstaendig.
	body      model.BodyState
	redaction model.RedactionStatus
	blob      string
}

// errCtxUrsache ist die Ursache, mit der die Zeile `abbruch ctx` den Kontext des Aufrufers beendet.
var errCtxUrsache = errors.New("kontext des aufrufers beendet")

// beliebig als writer heißt: die Zahl der Writer hängt vom Tempo der Goroutine ab und wird nicht geprüft.
const beliebig = -1

func xmlMetadata() capture.Eingang {
	return capture.Eingang{Richtung: capture.Response, ContentType: []string{"application/xml"}, Metadata: true}
}

func textPlain() capture.Eingang {
	return capture.Eingang{Richtung: capture.Response, ContentType: []string{"text/plain"}}
}

// mitGeheim ist ein Body, dessen geheimer Wert vor dem Verstoß steht und unter `password` redigiert wird.
func mitGeheim(rest string) []byte {
	return []byte(`{"password":"` + geheim + `",` + rest)
}

const maskiert = `{"password":"[REDACTED]",`

func captureFaelle() []fall {
	bombe := jsonString(1 << 20)
	g := gz(`{"password":"` + geheim + `","a":1}`)
	return []fall{
		{name: "gut", ein: antwort(), roh: []byte(`{"a":1}`), ausgang: capture.Vollstaendig, writer: 1, commits: 1,
			body: model.BodyPresent, redaction: model.RedactionNone, blob: `{"a":1}`},
		{name: "ersetzt", ein: antwort(), roh: mitGeheim(`"a":1}`), ausgang: capture.Vollstaendig, writer: 1, commits: 1,
			body: model.BodyPresent, redaction: model.RedactionApplied, blob: maskiert + `"a":1}`},
		{name: "leer", ein: antwort(), ausgang: capture.Vollstaendig, body: model.BodyAbsent, redaction: model.RedactionNone},
		{name: "gzip", ein: mitEncoding(antwort(), "gzip"), roh: g, ausgang: capture.Vollstaendig, writer: 1, commits: 1,
			body: model.BodyPresent, redaction: model.RedactionApplied, blob: maskiert + `"a":1}`},
		{name: "zlib", ein: mitEncoding(antwort(), "deflate"), roh: zl([]byte(`{"password":"` + geheim + `","a":1}`)),
			ausgang: capture.Vollstaendig, writer: 1, commits: 1,
			body: model.BodyPresent, redaction: model.RedactionApplied, blob: maskiert + `"a":1}`},
		{name: "typ", ein: textPlain(), roh: []byte(geheim), ausgang: capture.Verstoss, grund: capture.GrundTyp, ursache: model.ErrUnsicher},
		{name: "mismatch", ein: xmlMetadata(), roh: []byte(`{"password":"` + geheim + `"}`), ausgang: capture.Verstoss,
			grund: capture.GrundMismatch, ursache: model.ErrUnsicher},
		{name: "syntax", ein: antwort(), roh: mitGeheim(`"b":x`), ausgang: capture.Verstoss, grund: capture.GrundSyntax,
			ursache: model.ErrUnsicher, writer: 1},
		{name: "syntax im Request", ein: anfrage(), roh: mitGeheim(`"b":x`), ausgang: capture.Verstoss, grund: capture.GrundSyntax,
			ursache: model.ErrUnsicher, writer: 1},
		{name: "utf8", ein: antwort(), roh: mitGeheim("\"b\":[\"\xff\"]}"), ausgang: capture.Verstoss, grund: capture.GrundUTF8,
			ursache: model.ErrUnsicher, writer: 1},
		{name: "tiefe", ein: antwort(), roh: mitGeheim(`"b":` + wiederholt("[", 513)), ausgang: capture.Verstoss,
			grund: capture.GrundTiefe, ursache: model.ErrUnsicher, writer: 1},
		{name: "nichtstring", ein: antwort(), roh: mitGeheim(`"password":1}`), ausgang: capture.Verstoss,
			grund: capture.GrundNichtstring, ursache: model.ErrUnsicher, writer: 1},
		{name: "unvollstaendig", ein: antwort(), roh: mitGeheim(`"b":`), ausgang: capture.Verstoss,
			grund: capture.GrundUnvollstaendig, ursache: model.ErrUnsicher, writer: 1},
		{name: "zu_gross entpackt", ein: mitEncoding(antwort(), "gzip"), roh: gz(string(mitGeheim(`"b":"` + wiederholt("x", 200) + `"}`))),
			opt: []func(*capture.Konfig){mitMax(100), mitVerh(1000000)}, ausgang: capture.Verstoss, grund: capture.GrundZuGross,
			ursache: model.ErrUnsicher, writer: 1},
		{name: "zu_gross identity", ein: antwort(), roh: []byte(jsonString(60)),
			opt: []func(*capture.Konfig){mitMax(50)}, ausgang: capture.Verstoss, grund: capture.GrundZuGross,
			writer: 1},
		{name: "zu_gross roh", ein: mitEncoding(antwort(), "gzip"), roh: []byte(wiederholt(string(gz("")), 1000)),
			opt: []func(*capture.Konfig){mitMax(500)}, ausgang: capture.Verstoss, grund: capture.GrundZuGross},
		{name: "encoding", ein: mitEncoding(antwort(), "br"), roh: mitGeheim(`"a":1}`), ausgang: capture.Verstoss, grund: capture.GrundEncoding},
		{name: "dekodierung CRC", ein: mitEncoding(antwort(), "gzip"), roh: aendere(g, -8), ausgang: capture.Verstoss,
			grund: capture.GrundDekodierung, writer: 1},
		{name: "dekodierung Nachlauf", ein: mitEncoding(antwort(), "deflate"),
			roh: append(zl([]byte(`{"password":"`+geheim+`","a":1}`)), 'x'), ausgang: capture.Verstoss, grund: capture.GrundDekodierung, writer: 1},
		{name: "bombe", ein: mitEncoding(antwort(), "gzip"), roh: gz(bombe), opt: []func(*capture.Konfig){mitMax(8 << 20)},
			ausgang: capture.Verstoss, grund: capture.GrundBombe, writer: 1},
		{name: "ueberlauf", ein: antwort(), roh: mitGeheim(wiederholt(" ", 200)), opt: []func(*capture.Konfig){mitQueue(100)},
			treibe: treibeUeberlauf, ausgang: capture.Verstoss, grund: capture.GrundUeberlauf},
		{name: "frist", ein: antwort(), roh: mitGeheim(`"a":`), opt: []func(*capture.Konfig){mitFrist(30 * time.Millisecond)},
			treibe: treibeFrist, ausgang: capture.Verstoss, grund: capture.GrundFrist, ursache: context.DeadlineExceeded, writer: beliebig},
		{name: "ausgabe_zu_gross", ein: antwort(), roh: []byte(`{"password":""}`), opt: []func(*capture.Konfig){mitMax(15)},
			ausgang: capture.Verstoss, grund: capture.GrundAusgabeZuGross, ursache: outbound.ErrZuGross, writer: 1},
		{name: "ablage Begin", ein: antwort(), roh: mitGeheim(`"a":1}`), blobs: func(b *fakeBlobs) { b.beginFehler = errDisk },
			ausgang: capture.Verstoss, grund: capture.GrundAblage, ursache: errDisk},
		{name: "ablage Commit", ein: antwort(), roh: mitGeheim(`"a":1}`), blobs: func(b *fakeBlobs) { b.commitFehler = errDisk },
			ausgang: capture.Verstoss, grund: capture.GrundAblage, ursache: errDisk, writer: 1, commits: 1},
		{name: "abbruch src", ein: antwort(), roh: mitGeheim(`"a":`), ende: errQuelle, ausgang: capture.Abbruch,
			grund: capture.GrundAbbruch, ursache: errQuelle, writer: 1},
		{name: "abbruch Abbrechen", ein: antwort(), roh: mitGeheim(`"a":`), treibe: treibeAbbrechen, ausgang: capture.Abbruch,
			grund: capture.GrundAbbruch, ursache: context.Canceled},
		{name: "abbruch ctx", ein: antwort(), roh: mitGeheim(`"a":`), treibe: treibeCtx, ausgang: capture.Abbruch,
			grund: capture.GrundAbbruch, ursache: errCtxUrsache},
	}
}

func treibeUeberlauf(t *testing.T, u *umg, f fall) capture.Ergebnis {
	t.Helper()
	frei := u.pipe().AnhaltenAmStart()
	a := u.tee(quelle(string(f.roh)), f.ein)
	_, _ = io.Copy(io.Discard, a)
	frei()
	return abwarten(t, a)
}

func treibeFrist(t *testing.T, u *umg, f fall) capture.Ergebnis {
	t.Helper()
	s, ankunft, _ := blockiertSrc(t, string(f.roh))
	a := u.tee(s, f.ein)
	leseInGoroutine(a)
	warteAufAnkunft(t, ankunft)
	return abwarten(t, a)
}

func treibeAbbrechen(t *testing.T, u *umg, f fall) capture.Ergebnis {
	t.Helper()
	frei := u.pipe().AnhaltenAmStart()
	a := u.tee(quelle(), f.ein)
	a.Abbrechen()
	frei()
	return abwarten(t, a)
}

func treibeCtx(t *testing.T, u *umg, f fall) capture.Ergebnis {
	t.Helper()
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(errCtxUrsache)
	a, err := u.pipe().Tee(ctx, u.umhuellt(quelle()), f.ein)
	if err != nil {
		t.Fatal(err)
	}
	return abwarten(t, a)
}

// fuehreAus fährt eine Zeile und liefert die Umgebung und das Ergebnis; aufbau richtet die
// Umgebung vor dem Lauf zusätzlich ein.
func fuehreAus(t *testing.T, f fall, aufbau ...func(*umg)) (*umg, capture.Ergebnis) {
	t.Helper()
	u := neu(t, f.opt...)
	if f.blobs != nil {
		f.blobs(u.blobs)
	}
	for _, a := range aufbau {
		a(u)
	}
	if f.treibe != nil {
		return u, f.treibe(t, u, f)
	}
	src := quelle(string(f.roh))
	src.ende = f.ende
	return u, u.lauf(src, f.ein)
}
