package model

import (
	"errors"
	"strconv"
)

// URLUrsache ist die Ursache eines URLFehler (SPEC-160 Nr. 4).
type URLUrsache string

// Die zehn Ursachen der Klasse `ungueltig` (ODR-E6005).
const (
	URLZuLang        URLUrsache = "zu_lang"
	URLZuViele       URLUrsache = "zu_viele"
	URLKodierung     URLUrsache = "kodierung"
	URLSegment       URLUrsache = "segment"
	URLPfadForm      URLUrsache = "pfad_form"
	URLKey           URLUrsache = "key"
	URLLiteral       URLUrsache = "literal"
	URLOptionName    URLUrsache = "option_name"
	URLOptionDoppelt URLUrsache = "option_doppelt"
	URLWert          URLUrsache = "wert"
)

// Die sechs Ursachen der Klasse `nicht_unterstuetzt` (ODR-E6001).
const (
	URLServiceDokument         URLUrsache = "service_dokument"
	URLPfadNichtUnterstuetzt   URLUrsache = "pfad_nicht_unterstuetzt"
	URLLiteralTyp              URLUrsache = "literal_typ"
	URLOptionNichtUnterstuetzt URLUrsache = "option_nicht_unterstuetzt"
	URLFilter                  URLUrsache = "filter"
	URLWertNichtUnterstuetzt   URLUrsache = "wert_nicht_unterstuetzt"
)

// AlleURLUrsachen liefert die sechzehn Ursachen, zuerst die der Klasse `ungueltig`.
func AlleURLUrsachen() []URLUrsache {
	return []URLUrsache{
		URLZuLang, URLZuViele, URLKodierung, URLSegment, URLPfadForm, URLKey, URLLiteral, URLOptionName, URLOptionDoppelt, URLWert,
		URLServiceDokument, URLPfadNichtUnterstuetzt, URLLiteralTyp, URLOptionNichtUnterstuetzt, URLFilter, URLWertNichtUnterstuetzt,
	}
}

// NichtUnterstuetzt meldet, ob die Ursache zur Klasse `nicht_unterstuetzt` gehört (gültig, aber nicht unterstützt).
func (u URLUrsache) NichtUnterstuetzt() bool {
	switch u {
	case URLServiceDokument, URLPfadNichtUnterstuetzt, URLLiteralTyp, URLOptionNichtUnterstuetzt, URLFilter, URLWertNichtUnterstuetzt:
		return true
	}
	return false
}

// Gueltig meldet, ob u eine der sechzehn Ursachen ist.
func (u URLUrsache) Gueltig() bool {
	for _, v := range AlleURLUrsachen() {
		if u == v {
			return true
		}
	}
	return false
}

// ErrURLUngueltig erkennt jeden URLFehler der Klasse `ungueltig`; die Abbildung macht daraus ODR-E6005.
var ErrURLUngueltig = errors.New("model: OData-URL ungültig")

// ErrURLNichtUnterstuetzt erkennt jeden URLFehler der Klasse `nicht_unterstuetzt`; die Abbildung macht daraus ODR-E6001.
var ErrURLNichtUnterstuetzt = errors.New("model: OData-URL nicht unterstützt")

// URLFehler ist der Fehler des Lesers der OData-URL (SPEC-160 Nr. 4): Ursache und Ort, nie ein Teil der Eingabe.
type URLFehler struct {
	Ursache URLUrsache
	Ort     string
}

// Error nennt den Code und die Ursache; der Ort steht in der Schreibweise von strconv.Quote dahinter.
func (e *URLFehler) Error() string {
	s := "ODR-E6005: OData-URL ungültig: "
	if e.Ursache.NichtUnterstuetzt() {
		s = "ODR-E6001: OData-URL nicht unterstützt: "
	}
	s += string(e.Ursache)
	if e.Ort != "" {
		s += " " + strconv.Quote(e.Ort)
	}
	return s
}

// Code liefert ODR-E6001 für die Klasse `nicht_unterstuetzt`, sonst ODR-E6005.
func (e *URLFehler) Code() Code {
	if e.Ursache.NichtUnterstuetzt() {
		return CodeE6001
	}
	return CodeE6005
}

// Unwrap macht die Sentinel der Klasse erreichbar, nie die der anderen.
func (e *URLFehler) Unwrap() error {
	if e.Ursache.NichtUnterstuetzt() {
		return ErrURLNichtUnterstuetzt
	}
	return ErrURLUngueltig
}
