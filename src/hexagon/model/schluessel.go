package model

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"slices"
	"strings"
)

// KodierungsModus ist MATCH_INVALID_ENCODING (SPEC-081).
type KodierungsModus string

// Werte von MATCH_INVALID_ENCODING.
const (
	KodierungLiteral KodierungsModus = "literal"
	KodierungReject  KodierungsModus = "reject"
)

// BodyModus ist MATCH_BODY_MODE (SPEC-082).
type BodyModus string

// Werte von MATCH_BODY_MODE.
const (
	BodyBytes         BodyModus = "bytes"
	BodyJSONCanonical BodyModus = "json-canonical"
)

// SchluesselEingabe ist der unredigierte Request, aus dem BildeSchluessel den Key bildet
// (SPEC-116). Body trägt die Bytes des bereits redigierten Request-Bodys.
type SchluesselEingabe struct {
	Method           string
	ServiceID        string
	Path             string
	RawQuery         string
	Header           map[string][]string
	Body             []byte
	RequestRedaction RedactionStatus
}

// SchluesselRegeln sind die wirksamen Mengen und Aufzählungen der Konfiguration (SPEC-116).
type SchluesselRegeln struct {
	VaryHeaders       []string
	VolatileHeaders   []string
	RedactHeaders     []string
	RedactQueryParams []string
	InvalidEncoding   KodierungsModus
	BodyMode          BodyModus
}

// Schluessel ist das Ergebnis von BildeSchluessel: der Key und die persistierbare, redigierte
// Darstellung des Requests (SPEC-116 Nr. 8).
type Schluessel struct {
	MatchKey    string
	KeyRedacted bool
	RawQuery    string
	Header      map[string][]string
}

const schluesselMarke = "odr-match-v1"

// Arten der Query-Paare und des Bodys im Tupel (SPEC-116 Nr. 4, 6).
const (
	paarSchluessel byte = 0
	paarWert       byte = 1
	paarRoh        byte = 2

	bodyLeer      byte = 0
	bodyBytes     byte = 1
	bodyRoh       byte = 2
	bodyKanonisch byte = 3
)

// BildeSchluessel bildet den Match-Key und die redigierte Darstellung in einem Aufruf
// (SPEC-116). Fehler: ErrUngueltig für Regeln, RequestRedaction, Methode, Service und die Form
// des Pfades; ErrKodierung bei KodierungReject. Kein Fehler trägt einen Wert des Requests.
func BildeSchluessel(in SchluesselEingabe, regeln SchluesselRegeln) (Schluessel, error) {
	if err := pruefeEingabe(in, regeln); err != nil {
		return Schluessel{}, err
	}
	reject := regeln.InvalidEncoding == KodierungReject
	pfad, pfadOK := normalisierePfad(in.Path)
	if !pfadOK && reject {
		return Schluessel{}, ErrKodierung
	}
	rq, geaendert := RedigiereQuery(in.RawQuery, regeln.RedactQueryParams)
	paare, queryOK := zerlegePaare(rq)
	if !queryOK && reject {
		return Schluessel{}, ErrKodierung
	}
	var b tupel
	b.text(schluesselMarke)
	b.text(in.Method)
	b.text(in.ServiceID)
	b.text(pfad)
	b.paare(paare)
	b.vary(in.Header, regeln)
	b.body(in.Body, regeln.BodyMode)
	sum := sha256.Sum256(b.buf.Bytes())
	return Schluessel{
		MatchKey:    "sha256:" + hex.EncodeToString(sum[:]),
		KeyRedacted: geaendert || in.RequestRedaction == RedactionApplied,
		RawQuery:    rq,
		Header:      RedigiereHeader(in.Header, regeln.RedactHeaders, regeln.RedactQueryParams),
	}, nil
}

// pruefeEingabe prüft in der Reihenfolge von SPEC-116 Nr. 9: Aufzählungen der Regeln,
// RequestRedaction, Method, ServiceID, Form des Pfades.
func pruefeEingabe(in SchluesselEingabe, regeln SchluesselRegeln) error {
	switch {
	case regeln.InvalidEncoding != KodierungLiteral && regeln.InvalidEncoding != KodierungReject:
		return ungueltig("MATCH_INVALID_ENCODING ist unbekannt")
	case regeln.BodyMode != BodyBytes && regeln.BodyMode != BodyJSONCanonical:
		return ungueltig("MATCH_BODY_MODE ist unbekannt")
	case in.RequestRedaction != RedactionNone && in.RequestRedaction != RedactionApplied:
		return ungueltig("redaction.request erlaubt keinen Key")
	case in.Method == "":
		return ungueltig("Method fehlt")
	case in.ServiceID == "":
		return ungueltig("ServiceID fehlt")
	case !pfadFormGueltig(in.Path):
		return ungueltig("Pfad hat eine ungültige Form")
	}
	return nil
}

func pfadFormGueltig(p string) bool {
	if p == "" || p[0] != '/' {
		return false
	}
	for i := 0; i < len(p); i++ {
		if c := p[i]; c < 0x20 || c == 0x7f || c == '?' || c == '#' {
			return false
		}
	}
	return true
}

func unreserviert(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' ||
		c == '-' || c == '.' || c == '_' || c == '~'
}

// normalisierePfad fasst allein %XX an (SPEC-116 Nr. 3) und meldet, ob jedes % zwei Hex-Ziffern
// trägt; ein anderes % bleibt stehen.
func normalisierePfad(p string) (string, bool) {
	var sb strings.Builder
	gueltig := true
	for i := 0; i < len(p); i++ {
		c := p[i]
		if c != '%' {
			sb.WriteByte(c)
			continue
		}
		if i+2 >= len(p) || hexWert(p[i+1]) < 0 || hexWert(p[i+2]) < 0 {
			gueltig = false
			sb.WriteByte(c)
			continue
		}
		w := byte(hexWert(p[i+1])<<4 | hexWert(p[i+2]))
		if unreserviert(w) {
			sb.WriteByte(w)
		} else {
			sb.WriteString("%" + strings.ToUpper(p[i+1:i+3]))
		}
		i += 2
	}
	return sb.String(), gueltig
}

type paar struct {
	art  byte
	k, v string
}

// zerlegePaare liefert die sortierte Multimenge der Paare aus dem redigierten raw_query und
// meldet, ob jedes Paar gültig kodiert ist; ein ungültiges Paar ist ein Paar der Art 2 (Nr. 4).
func zerlegePaare(rq string) ([]paar, bool) {
	var out []paar
	gueltig := true
	for _, s := range strings.Split(rq, "&") {
		if s == "" {
			continue
		}
		k, v, hatWert := strings.Cut(s, "=")
		dk, okK := dekodiereProzent(k)
		dv, okV := dekodiereProzent(v)
		switch {
		case !okK || !okV:
			gueltig = false
			out = append(out, paar{art: paarRoh, k: s})
		case hatWert:
			out = append(out, paar{art: paarWert, k: dk, v: dv})
		default:
			out = append(out, paar{art: paarSchluessel, k: dk})
		}
	}
	slices.SortFunc(out, func(a, b paar) int {
		if a.art != b.art {
			return int(a.art) - int(b.art)
		}
		if c := strings.Compare(a.k, b.k); c != 0 {
			return c
		}
		return strings.Compare(a.v, b.v)
	})
	return out, gueltig
}

type tupel struct{ buf bytes.Buffer }

func (t *tupel) zahl(n int) {
	t.buf.Write(binary.BigEndian.AppendUint32(nil, uint32(n)))
}

func (t *tupel) text(s string) {
	t.zahl(len(s))
	t.buf.WriteString(s)
}

func (t *tupel) paare(ps []paar) {
	t.zahl(len(ps))
	for _, p := range ps {
		t.buf.WriteByte(p.art)
		t.text(p.k)
		if p.art == paarWert {
			t.text(p.v)
		}
	}
}

// vary kodiert die Namen aus MATCH_VARY_HEADERS ohne die flüchtigen und die redigierten, die
// der Request mit mindestens einem Wert trägt (SPEC-116 Nr. 5).
func (t *tupel) vary(h map[string][]string, regeln SchluesselRegeln) {
	ausgenommen := mengeKlein(append(slices.Clone(regeln.VolatileHeaders), regeln.RedactHeaders...))
	var namen []string
	for _, n := range regeln.VaryHeaders {
		k := kanonischerName(n)
		if !ausgenommen[asciiKlein(k)] && !slices.Contains(namen, k) {
			namen = append(namen, k)
		}
	}
	slices.Sort(namen)
	zusammen := fasseNamenZusammen(h)
	var vorhanden []string
	for _, n := range namen {
		if len(zusammen[n]) > 0 {
			vorhanden = append(vorhanden, n)
		}
	}
	t.zahl(len(vorhanden))
	for _, n := range vorhanden {
		t.text(n)
		t.zahl(len(zusammen[n]))
		for _, w := range zusammen[n] {
			t.text(strings.Trim(w, " \t"))
		}
	}
}

func (t *tupel) body(b []byte, modus BodyModus) {
	if len(b) == 0 {
		t.buf.WriteByte(bodyLeer)
		return
	}
	art, daten := bodyBytes, b
	if modus == BodyJSONCanonical {
		if kanon, ok := kanonischesJSON(b); ok {
			art, daten = bodyKanonisch, kanon
		} else {
			art = bodyRoh
		}
	}
	sum := sha256.Sum256(daten)
	t.buf.WriteByte(art)
	t.buf.Write(sum[:])
}
