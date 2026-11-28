package batchcodec_test

import (
	"context"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/batchcodec"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Steuerzeichen der Fixtures aus Bytewerten (AGENTS.md §3.18).
const (
	cr   = string(rune(0x0D))
	lf   = string(rune(0x0A))
	crlf = cr + lf
	ht   = string(rune(0x09))
	nul  = string(rune(0x00))
)

// geheim steht in den Verstoß-Fixtures und darf in keinem Fehler erscheinen (SPEC-182 Nr. 4).
const geheim = "GEHEIMTEXT"

const (
	ctBatch   = "multipart/mixed; boundary=b"
	zeileCTH  = "Content-Type: application/http"
	grenzeAll = 1 << 24
)

// block verbindet Kopfzeilen und eine leere Zeile mit dem Rest eines Teils oder Requests.
func block(zeilen []string, rest string) string {
	s := ""
	for _, z := range zeilen {
		s += z + crlf
	}
	return s + crlf + rest
}

// anfrage ist der Rest eines Teils: Anfragezeile, Kopfzeilen des Requests, leere Zeile, Body.
func anfrage(zeile string, kopf []string, body string) string {
	return zeile + crlf + block(kopf, body)
}

// httpTeil ist ein Teil mit application/http und einer Anfrage ohne Body.
func httpTeil(zeile string) string { return block([]string{zeileCTH}, anfrage(zeile, nil, "")) }

// idTeil ist ein Teil mit Content-ID.
func idTeil(id, zeile string) string {
	return block([]string{zeileCTH, "Content-ID: " + id}, anfrage(zeile, nil, ""))
}

// ebene fügt Teile mit den Grenzzeilen der Zeichenfolge g zusammen, ohne CRLF hinter der schließenden Zeile.
func ebene(g string, teile ...string) string {
	var s strings.Builder
	for _, t := range teile {
		s.WriteString("--" + g + crlf + t + crlf)
	}
	s.WriteString("--" + g + "--")
	return s.String()
}

// huelle ist der Body mit der Grenzzeichenfolge b.
func huelle(teile ...string) string { return ebene("b", teile...) + crlf }

// changeset ist ein Teil mit multipart/mixed und der Grenzzeichenfolge g.
func changeset(g string, teile ...string) string {
	return block([]string{"Content-Type: multipart/mixed; boundary=" + g}, ebene(g, teile...))
}

func lese(body string) (model.Batch, error) {
	return leseMit(strings.NewReader(body), []string{ctBatch}, grenzeAll)
}

func leseMit(src io.Reader, ct []string, grenze int64) (model.Batch, error) {
	return batchcodec.NewLeser().LeseBatch(context.Background(), src, ct, grenze)
}

// req ist der erwartete Request ohne Kopfzeilen und Body.
func req(methode, ziel string) model.BatchRequest {
	return model.BatchRequest{Methode: methode, Ziel: ziel, Header: []model.BatchKopf{}}
}

func einzel(r model.BatchRequest) model.BatchTeil {
	return model.BatchTeil{Requests: []model.BatchRequest{r}}
}

func menge(requests ...model.BatchRequest) model.BatchTeil {
	if requests == nil {
		requests = []model.BatchRequest{}
	}
	return model.BatchTeil{Changeset: true, Requests: requests}
}

func ergebnis(teile ...model.BatchTeil) model.Batch {
	if teile == nil {
		teile = []model.BatchTeil{}
	}
	return model.Batch{Teile: teile}
}

func wantOK(t *testing.T, was string, got model.Batch, err error, want model.Batch) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: Fehler %v", was, err)
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s: %#v, erwartet %#v", was, got, want)
	}
}

// wantFehler verlangt genau den BatchFehler mit Ursache und Ort, den Nullwert als Ergebnis und kein Byte der Marke in Error() und
// Ort.
func wantFehler(t *testing.T, was string, got model.Batch, err error, u model.BatchUrsache, ort string) {
	t.Helper()
	f, ok := err.(*model.BatchFehler)
	if !ok || f.Ursache != u || f.Ort != ort {
		t.Errorf("%s: Fehler %v, erwartet %s %q", was, err, u, ort)
		return
	}
	if !reflect.DeepEqual(got, model.Batch{}) {
		t.Errorf("%s: Ergebnis %#v neben dem Fehler", was, got)
	}
	if strings.Contains(f.Error(), geheim) || strings.Contains(f.Ort, geheim) {
		t.Errorf("%s: der Fehler trägt Bytes der Eingabe: %q", was, f.Error())
	}
}

// jedeSchreibweise ruft f mit jeder Groß- und Kleinschreibung der ASCII-Buchstaben von s und liefert die Zahl der Aufrufe.
func jedeSchreibweise(s string, f func(string)) int {
	var stellen []int
	for i := range len(s) {
		if c := s[i] | 0x20; c >= 'a' && c <= 'z' {
			stellen = append(stellen, i)
		}
	}
	b := []byte(strings.ToLower(s))
	for maske := range 1 << len(stellen) {
		for k, i := range stellen {
			b[i] = b[i] | 0x20
			if maske&(1<<k) != 0 {
				b[i] &^= 0x20
			}
		}
		f(string(b))
	}
	return 1 << len(stellen)
}

// quelle ist ein Strom, der jeden Aufruf jeder Schnittstelle des Pakets io zählt und jede angeforderte Länge festhält.
type quelle struct {
	data        []byte
	pos         int
	stueck      []int
	hoechstens  int
	eofMitDaten bool
	fehler      error
	fehlerBei   int
	mitDaten    bool
	beim        func(k int)

	reads, nachEnde int
	ende            bool
	angefordert     []int
	geliefert       []int
	andere          []string
}

func neueQuelle(s string) *quelle { return &quelle{data: []byte(s)} }

func (q *quelle) Read(p []byte) (int, error) {
	if q.ende {
		q.nachEnde++
		return 0, io.EOF
	}
	q.reads++
	if q.beim != nil {
		q.beim(q.reads)
	}
	q.angefordert = append(q.angefordert, len(p))
	grenze := len(q.data)
	if q.fehler != nil {
		grenze = q.fehlerBei
	}
	n := min(len(p), grenze-q.pos)
	switch {
	case len(q.stueck) > 0:
		n = min(n, q.stueck[0])
		leer := q.stueck[0] == 0
		q.stueck = q.stueck[1:]
		if leer {
			q.geliefert = append(q.geliefert, 0)
			return 0, nil
		}
	case q.hoechstens > 0:
		n = min(n, q.hoechstens)
	}
	copy(p, q.data[q.pos:q.pos+n])
	q.pos += n
	q.geliefert = append(q.geliefert, n)
	if q.pos < grenze {
		return n, nil
	}
	if q.fehler != nil && (q.mitDaten || n == 0) {
		q.ende = true
		return n, q.fehler
	}
	if q.fehler == nil && (q.eofMitDaten || n == 0) {
		q.ende = true
		return n, io.EOF
	}
	return n, nil
}

func (q *quelle) merke(name string) { q.andere = append(q.andere, name) }

func (q *quelle) Close() error                     { q.merke("Close"); return nil }
func (q *quelle) WriteTo(io.Writer) (int64, error) { q.merke("WriteTo"); return 0, io.EOF }
func (q *quelle) ReadByte() (byte, error)          { q.merke("ReadByte"); return 0, io.EOF }
func (q *quelle) UnreadByte() error                { q.merke("UnreadByte"); return nil }
func (q *quelle) ReadRune() (rune, int, error)     { q.merke("ReadRune"); return 0, 0, io.EOF }
func (q *quelle) UnreadRune() error                { q.merke("UnreadRune"); return nil }
func (q *quelle) Seek(int64, int) (int64, error)   { q.merke("Seek"); return 0, nil }
func (q *quelle) ReadAt([]byte, int64) (int, error) {
	q.merke("ReadAt")
	return 0, io.EOF
}

// beispiel ist das Beispiel aus SPEC-183 Nr. 8 mit den Grenzzeichenfolgen batch_1 und cs_1.
func beispiel() string {
	zeilen := []string{
		"--batch_1", zeileCTH, "Content-Transfer-Encoding: binary", "", "GET Products(1) HTTP/1.1", "Host: recorder.test", "",
		"--batch_1", "Content-Type: multipart/mixed; boundary=cs_1", "",
		"--cs_1", zeileCTH, "Content-ID: 1", "", "POST Products HTTP/1.1", "Content-Type: application/json", "", `{"Id":9}`,
		"--cs_1", zeileCTH, "Content-ID: 2", "", "PATCH $1 HTTP/1.1", "Content-Type: application/json", "", `{"Name":"Nail"}`,
		"--cs_1--",
		"--batch_1--",
	}
	return strings.Join(zeilen, crlf) + crlf
}

func ctBeispiel() []string { return []string{"multipart/mixed; boundary=batch_1"} }

func beispielErgebnis() model.Batch {
	json := []model.BatchKopf{{Name: "Content-Type", Wert: "application/json"}}
	get := req("GET", "Products(1)")
	get.Header = []model.BatchKopf{{Name: "Host", Wert: "recorder.test"}}
	post := model.BatchRequest{ContentID: "1", Methode: "POST", Ziel: "Products", Header: json, Body: []byte(`{"Id":9}`)}
	patch := model.BatchRequest{ContentID: "2", Methode: "PATCH", Ziel: "$1", Header: json, Body: []byte(`{"Name":"Nail"}`)}
	return ergebnis(einzel(get), menge(post, patch))
}

func esB(bytes ...byte) string { return string(bytes) }
