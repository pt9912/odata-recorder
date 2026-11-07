package model

import (
	"errors"
	"strconv"
	"unicode/utf8"
)

// EingabeUrsache ist die Ursache eines EingabeFehler (SPEC-180 Nr. 3).
type EingabeUrsache string

// Die fünfzehn Ursachen der Klasse ungueltig: fünf des Lesens, content_type vor dem Lesen, neun des Deutens.
const (
	EingabeSyntax             EingabeUrsache = "syntax"
	EingabeKodierung          EingabeUrsache = "kodierung"
	EingabeTiefe              EingabeUrsache = "tiefe"
	EingabeZuGross            EingabeUrsache = "zu_gross"
	EingabeDoppelterName      EingabeUrsache = "doppelter_name"
	EingabeContentType        EingabeUrsache = "content_type"
	EingabeStruktur           EingabeUrsache = "struktur"
	EingabeODataTyp           EingabeUrsache = "odata_typ"
	EingabeUnbekannteProperty EingabeUrsache = "unbekannte_property"
	EingabeTyp                EingabeUrsache = "typ"
	EingabeWert               EingabeUrsache = "wert"
	EingabeNull               EingabeUrsache = "null"
	EingabeZahlZuLang         EingabeUrsache = "zahl_zu_lang"
	EingabeSchluessel         EingabeUrsache = "schluessel"
	EingabeFehlend            EingabeUrsache = "fehlend"
)

// Die vier Ursachen der Klasse nicht_unterstuetzt.
const (
	EingabePropertyTyp EingabeUrsache = "property_typ"
	EingabeNavigation  EingabeUrsache = "navigation"
	EingabeBind        EingabeUrsache = "bind"
	EingabeOffen       EingabeUrsache = "offen"
)

// AlleEingabeUrsachen liefert die neunzehn Ursachen in der Reihenfolge der Spezifikation.
func AlleEingabeUrsachen() []EingabeUrsache {
	return []EingabeUrsache{
		EingabeSyntax, EingabeKodierung, EingabeTiefe, EingabeZuGross, EingabeDoppelterName, EingabeContentType,
		EingabeStruktur, EingabeODataTyp, EingabeUnbekannteProperty, EingabeTyp, EingabeWert, EingabeNull,
		EingabeZahlZuLang, EingabeSchluessel, EingabeFehlend,
		EingabePropertyTyp, EingabeNavigation, EingabeBind, EingabeOffen,
	}
}

// NichtUnterstuetzt ist wahr für die vier Ursachen der Klasse nicht_unterstuetzt.
func (u EingabeUrsache) NichtUnterstuetzt() bool {
	return u == EingabePropertyTyp || u == EingabeNavigation || u == EingabeBind || u == EingabeOffen
}

// ErrEingabeUngueltig erkennt jeden EingabeFehler der Klasse ungueltig; die Abbildung macht daraus ODR-E6007.
var ErrEingabeUngueltig = errors.New("model: Request-Body ungültig")

// ErrEingabeNichtUnterstuetzt erkennt jeden EingabeFehler der Klasse nicht_unterstuetzt; die Abbildung macht daraus ODR-E6002.
var ErrEingabeNichtUnterstuetzt = errors.New("model: Request-Body nicht unterstützt")

// EingabeFehler ist der Fehler des Lesers der Client-Bodies (SPEC-180 Nr. 3): Ursache und Ort, nie ein Wert des Bodys.
type EingabeFehler struct {
	Ursache EingabeUrsache
	Ort     string
}

// ortGrenze ist die Länge in Bytes, auf die Error() den Ort kürzt (SPEC-180 Nr. 3).
const ortGrenze = 256

// Error nennt den Code und die Ursache; der Ort steht in der Schreibweise von strconv.Quote dahinter, denn er enthält Text des
// Dokuments und darf kein rohes Steuerzeichen tragen. Der Ort ist auf höchstens 256 Bytes am Rand eines UTF-8-Zeichens
// gekürzt, nach dem schließenden Anführungszeichen folgt `...`, wenn gekürzt wurde; das Feld Ort bleibt ungekürzt.
func (e *EingabeFehler) Error() string {
	s := e.Code().String() + ": Request-Body ungültig: "
	if e.Ursache.NichtUnterstuetzt() {
		s = e.Code().String() + ": Request-Body nicht unterstützt: "
	}
	s += string(e.Ursache)
	if e.Ort != "" {
		ort, gekuerzt := kuerzeOrt(e.Ort)
		s += " " + strconv.Quote(ort)
		if gekuerzt {
			s += "..."
		}
	}
	return s
}

// kuerzeOrt liefert den längsten Anfang von ort mit höchstens ortGrenze Bytes, der nicht in einem UTF-8-Zeichen endet.
func kuerzeOrt(ort string) (string, bool) {
	if len(ort) <= ortGrenze {
		return ort, false
	}
	n := ortGrenze
	for n > 0 && !utf8.RuneStart(ort[n]) {
		n--
	}
	return ort[:n], true
}

// Code liefert ODR-E6002 für die Klasse nicht_unterstuetzt, sonst ODR-E6007.
func (e *EingabeFehler) Code() Code {
	if e.Ursache.NichtUnterstuetzt() {
		return CodeE6002
	}
	return CodeE6007
}

// Unwrap macht den Sentinel der Klasse erreichbar, nie den der anderen.
func (e *EingabeFehler) Unwrap() error {
	if e.Ursache.NichtUnterstuetzt() {
		return ErrEingabeNichtUnterstuetzt
	}
	return ErrEingabeUngueltig
}
