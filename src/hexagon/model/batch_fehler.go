package model

import "errors"

// BatchUrsache ist die Ursache eines BatchFehler (SPEC-182 Nr. 4).
type BatchUrsache string

// Die Ursachen, die der Leser von `$batch` meldet, in der Reihenfolge der Spezifikation.
const (
	BatchContentType       BatchUrsache = "content_type"
	BatchBoundary          BatchUrsache = "boundary"
	BatchZuGross           BatchUrsache = "zu_gross"
	BatchZuViele           BatchUrsache = "zu_viele"
	BatchRahmen            BatchUrsache = "rahmen"
	BatchKopfzeile         BatchUrsache = "kopfzeile"
	BatchContentID         BatchUrsache = "content_id"
	BatchAnfragezeile      BatchUrsache = "anfragezeile"
	BatchFormat            BatchUrsache = "format"
	BatchTeilTyp           BatchUrsache = "teil_typ"
	BatchVerschachtelt     BatchUrsache = "verschachtelt"
	BatchTransferKodierung BatchUrsache = "transfer_kodierung"
)

// Die Ursachen, die allein die Prüfung der Struktur meldet (SPEC-184).
const (
	BatchMethode                  BatchUrsache = "methode"
	BatchContentIDDoppelt         BatchUrsache = "content_id_doppelt"
	BatchReferenz                 BatchUrsache = "referenz"
	BatchMethodeNichtUnterstuetzt BatchUrsache = "methode_nicht_unterstuetzt"
)

// AlleBatchUrsachen liefert die sechzehn Ursachen: die zwölf des Lesers, dann die vier der Prüfung.
func AlleBatchUrsachen() []BatchUrsache {
	return []BatchUrsache{
		BatchContentType, BatchBoundary, BatchZuGross, BatchZuViele, BatchRahmen, BatchKopfzeile, BatchContentID, BatchAnfragezeile,
		BatchFormat, BatchTeilTyp, BatchVerschachtelt, BatchTransferKodierung,
		BatchMethode, BatchContentIDDoppelt, BatchReferenz, BatchMethodeNichtUnterstuetzt,
	}
}

// nichtUnterstuetzt ist wahr für die Ursachen der Klasse nicht_unterstuetzt.
func (u BatchUrsache) nichtUnterstuetzt() bool {
	switch u {
	case BatchFormat, BatchTeilTyp, BatchVerschachtelt, BatchTransferKodierung, BatchMethodeNichtUnterstuetzt:
		return true
	}
	return false
}

// ErrBatchUngueltig erkennt jeden BatchFehler der Klasse ungueltig; die Abbildung macht daraus ODR-E6007.
var ErrBatchUngueltig = errors.New("model: Batch ungültig")

// ErrBatchNichtUnterstuetzt erkennt jeden BatchFehler der Klasse nicht_unterstuetzt; die Abbildung macht daraus ODR-E6001.
var ErrBatchNichtUnterstuetzt = errors.New("model: Batch nicht unterstützt")

// BatchFehler ist der Fehler des Lesers und der Prüfung eines `$batch` (SPEC-182 Nr. 4). Ort ist leer für die Hülle, sonst ein
// Teilpfad aus 1-basierten Dezimalzahlen (`3`, `3/2`); er trägt wie die Meldung kein Byte der Eingabe.
type BatchFehler struct {
	Ursache BatchUrsache
	Ort     string
}

// Error nennt den Code und die Ursache, bei nichtleerem Ort gefolgt von einem Leerzeichen und ihm.
func (e *BatchFehler) Error() string {
	s := "ODR-E6007: Batch ungültig: "
	if e.Ursache.nichtUnterstuetzt() {
		s = "ODR-E6001: Batch nicht unterstützt: "
	}
	s += string(e.Ursache)
	if e.Ort != "" {
		s += " " + e.Ort
	}
	return s
}

// Code liefert ODR-E6001 für die Klasse nicht_unterstuetzt, sonst ODR-E6007.
func (e *BatchFehler) Code() Code {
	if e.Ursache.nichtUnterstuetzt() {
		return CodeE6001
	}
	return CodeE6007
}

// Unwrap macht den Sentinel der Klasse erreichbar, nie den der anderen.
func (e *BatchFehler) Unwrap() error {
	if e.Ursache.nichtUnterstuetzt() {
		return ErrBatchNichtUnterstuetzt
	}
	return ErrBatchUngueltig
}
