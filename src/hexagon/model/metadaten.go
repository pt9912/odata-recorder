package model

import (
	"errors"
	"strconv"
)

// FundArt ist die Art eines Fundes im Validierungsbericht (SPEC-125 Nr. 7, 12); der Grund eines
// Fundes ist durch die Art bestimmt.
type FundArt string

// Die vier Arten.
const (
	FundPropertyTyp FundArt = "property_typ"
	FundAnnotation  FundArt = "annotation"
	FundKonstrukt   FundArt = "konstrukt"
	FundErweiterung FundArt = "erweiterung"
)

// Fund ist ein Element des Dokuments, das das Modell gültig lässt und das der Leser nicht auswertet.
type Fund struct {
	Art  FundArt
	Ort  string
	Name string
}

// ValidierungsBericht ist der Bericht zu einem gültigen Dokument (LH-FA-06.a); Funde ist dann nie nil.
type ValidierungsBericht struct {
	Funde []Fund
}

// MetadatenUrsache ist die Ursache eines MetadatenFehler (SPEC-125 Nr. 13).
type MetadatenUrsache string

// Die dreizehn Ursachen.
const (
	UrsacheSyntax          MetadatenUrsache = "syntax"
	UrsacheKodierung       MetadatenUrsache = "kodierung"
	UrsacheDoctype         MetadatenUrsache = "doctype"
	UrsacheWurzel          MetadatenUrsache = "wurzel"
	UrsacheStruktur        MetadatenUrsache = "struktur"
	UrsacheVersion         MetadatenUrsache = "version"
	UrsacheOhneSchluessel  MetadatenUrsache = "ohne_schluessel"
	UrsacheNichtAufloesbar MetadatenUrsache = "nicht_aufloesbar"
	UrsacheSchluesseltyp   MetadatenUrsache = "schluesseltyp"
	UrsacheZuGross         MetadatenUrsache = "zu_gross"
	UrsacheTiefe           MetadatenUrsache = "tiefe"
	UrsacheKette           MetadatenUrsache = "kette"
	UrsacheZuVieleMembers  MetadatenUrsache = "zu_viele_properties"
)

// ErrMetadatenUngueltig erkennt jeden MetadatenFehler über errors.Is; die Abbildung macht daraus ODR-E6004.
var ErrMetadatenUngueltig = errors.New("model: Metadaten ungültig")

// MetadatenFehler ist der Fehler des Metadatenlesers (SPEC-125 Nr. 13): ODR-E6004 mit Ursache und Ort.
type MetadatenFehler struct {
	Ursache MetadatenUrsache
	Ort     string
}

// Error nennt den Code und die Ursache; der Ort steht in der Schreibweise von strconv.Quote dahinter,
// denn er enthält Text des Dokuments und darf kein rohes Steuerzeichen tragen.
func (e *MetadatenFehler) Error() string {
	s := "ODR-E6004: Metadaten ungültig oder nicht unterstützt: " + string(e.Ursache)
	if e.Ort != "" {
		s += " " + strconv.Quote(e.Ort)
	}
	return s
}

// Code liefert ODR-E6004.
func (*MetadatenFehler) Code() Code { return CodeE6004 }

// Unwrap macht ErrMetadatenUngueltig erreichbar.
func (*MetadatenFehler) Unwrap() error { return ErrMetadatenUngueltig }
