package telemetry

import (
	"math"
	"slices"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// form ist die Wertform eines Schlüssels (SPEC-230 Nr. 3).
type form int

const (
	formToken form = iota + 1
	formKennung
	formName
	formMethode
	formStatus
	formZahl
	formFehlercode
	formDauer
	formBlob
)

// regel ist Art und Wertform eines Schlüssels.
type regel struct {
	art  support.Art
	form form
}

// operation benennt die Menge der Schlüssel: opLog oder der Name eines Spans.
type operation string

const opLog operation = ""

// logRegel liefert die Regel der 13 Schlüssel der Menge Log (SPEC-230 Nr. 3).
func logRegel(key string) (regel, bool) {
	switch key {
	case "mode", "ereignis", "richtung", "grund":
		return regel{support.ArtString, formToken}, true
	case "service_id", "session_id", "exchange_id", "workspace_id", "correlation_id":
		return regel{support.ArtString, formKennung}, true
	case "error_code":
		return regel{support.ArtString, formFehlercode}, true
	case "duration_ms":
		return regel{support.ArtFloat, formDauer}, true
	case "blob":
		return regel{support.ArtString, formBlob}, true
	case "befunde":
		return regel{support.ArtInt, formZahl}, true
	}
	return regel{}, false
}

// spanRegel liefert die Regel der elf Pflicht-Attribute der fünf Spans (SPEC-230 Nr. 3).
func spanRegel(key string) (regel, bool) {
	switch key {
	case "http.request.method":
		return regel{support.ArtString, formMethode}, true
	case "http.response.status_code":
		return regel{support.ArtInt, formStatus}, true
	case "recorder.mode", "recorder.capture_status", "recorder.match_strategy",
		"recorder.match_result", "recorder.coverage_status", "recorder.mutation_type":
		return regel{support.ArtString, formToken}, true
	case "recorder.session_id", "recorder.workspace_id":
		return regel{support.ArtString, formKennung}, true
	case "recorder.entity_set":
		return regel{support.ArtString, formName}, true
	}
	return regel{}, false
}

// schluessel liefert die Regel von key, wenn key in der Menge von op liegt (SPEC-230 Nr. 2).
func schluessel(op operation, key string) (regel, bool) {
	if op == opLog {
		return logRegel(key)
	}
	for _, s := range support.Spans() {
		if string(s.Name) == string(op) && slices.Contains(s.Pflicht, key) {
			return spanRegel(key)
		}
	}
	return regel{}, false
}

// filtern wendet SPEC-230 Nr. 4 an: je Attribut höchstens ein Verwurf, Schlüssel vor Wert; der
// erste angenommene Wert eines Schlüssels gilt. Es liefert die angenommenen Attribute und die
// Verwürfe in der Reihenfolge der Attribute.
func filtern(op operation, attrs []support.Attr) ([]support.Attr, []support.Verwurf) {
	var angenommen []support.Attr
	var verworfen []support.Verwurf
	belegt := map[string]bool{}
	for _, a := range attrs {
		r, ok := schluessel(op, a.Key)
		if !ok || belegt[a.Key] {
			verworfen = append(verworfen, support.VerwurfSchluessel)
			continue
		}
		if !wertGueltig(r, a) {
			verworfen = append(verworfen, support.VerwurfWert)
			continue
		}
		belegt[a.Key] = true
		angenommen = append(angenommen, a)
	}
	return angenommen, verworfen
}

// wertGueltig prüft Art und Form des Werts von a gegen r.
func wertGueltig(r regel, a support.Attr) bool {
	if a.Art != r.art {
		return false
	}
	switch r.form {
	case formToken:
		return token(a.Text)
	case formKennung:
		return nurZeichen(a.Text, 64, kennungByte)
	case formName:
		return name(a.Text)
	case formMethode:
		return nurZeichen(a.Text, 16, grossbuchstabe)
	case formStatus:
		return a.Ganz >= 100 && a.Ganz <= 599
	case formZahl:
		return a.Ganz >= 0
	case formFehlercode:
		_, ok := model.ParseCode(a.Text)
		return ok
	case formDauer:
		return a.Gleit >= 0 && !math.Signbit(a.Gleit) && !math.IsInf(a.Gleit, 0)
	case formBlob:
		return blobReferenz(a.Text)
	}
	return false
}

// nurZeichen meldet, ob s aus 1 bis laenge Bytes besteht, die alle ok erfüllen; ein Byte ab 0x80
// erfüllt keine der Funktionen.
func nurZeichen(s string, laenge int, ok func(byte) bool) bool {
	if len(s) == 0 || len(s) > laenge {
		return false
	}
	for i := range len(s) {
		if !ok(s[i]) {
			return false
		}
	}
	return true
}

func kleinbuchstabe(c byte) bool { return c >= 'a' && c <= 'z' }

func grossbuchstabe(c byte) bool { return c >= 'A' && c <= 'Z' }

func ziffer(c byte) bool { return c >= '0' && c <= '9' }

func hexKlein(c byte) bool { return ziffer(c) || (c >= 'a' && c <= 'f') }

func kennungByte(c byte) bool {
	return kleinbuchstabe(c) || grossbuchstabe(c) || ziffer(c) || c == '.' || c == '_' || c == '-'
}

// token: 1 bis 32 Zeichen aus a-z, 0-9, _; das erste aus a-z.
func token(s string) bool {
	return nurZeichen(s, 32, func(c byte) bool { return kleinbuchstabe(c) || ziffer(c) || c == '_' }) &&
		kleinbuchstabe(s[0])
}

// name: 1 bis 128 Zeichen aus A-Z, a-z, 0-9, _, .; das erste ein Buchstabe oder _.
func name(s string) bool {
	return nurZeichen(s, 128, func(c byte) bool { return kleinbuchstabe(c) || grossbuchstabe(c) || ziffer(c) || c == '_' || c == '.' }) &&
		(kleinbuchstabe(s[0]) || grossbuchstabe(s[0]) || s[0] == '_')
}

const blobPraefix = "sha256:"

// blobReferenz: sha256: und 64 Zeichen aus 0-9a-f.
func blobReferenz(s string) bool {
	if len(s) != len(blobPraefix)+64 || s[:len(blobPraefix)] != blobPraefix {
		return false
	}
	return nurZeichen(s[len(blobPraefix):], 64, hexKlein)
}
