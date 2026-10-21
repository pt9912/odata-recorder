package capture

import (
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Richtung nennt, ob ein Body zum Request oder zur Response gehört. Der Nullwert ist keine Richtung.
type Richtung uint8

// Die beiden Richtungen (SPEC-119).
const (
	Request Richtung = iota + 1
	Response
)

// Ausgang ist der Ausgang der Aufnahme eines Bodys (SPEC-119 Nr. 10).
type Ausgang string

// Die Ausgänge: Vollstaendig ist ein erfasster Body, Abbruch das vorzeitige Ende des Austauschs,
// Verstoss die Verwerfung durch die Capture-Seite.
const (
	Vollstaendig Ausgang = "vollstaendig"
	Abbruch      Ausgang = "abbruch"
	Verstoss     Ausgang = "verstoss"
)

// Grund nennt die Ursache eines Abbruchs oder Verstoßes (SPEC-119 Nr. 10).
type Grund string

// Die Gründe aus SPEC-119 Nr. 10; die ersten acht sind die des Redaktors (SPEC-118 Nr. 7).
const (
	GrundTyp            Grund = "typ"
	GrundMismatch       Grund = "mismatch"
	GrundSyntax         Grund = "syntax"
	GrundUTF8           Grund = "utf8"
	GrundTiefe          Grund = "tiefe"
	GrundNichtstring    Grund = "nichtstring"
	GrundUnvollstaendig Grund = "unvollstaendig"
	GrundZuGross        Grund = "zu_gross"
	GrundEncoding       Grund = "encoding"
	GrundDekodierung    Grund = "dekodierung"
	GrundBombe          Grund = "bombe"
	GrundUeberlauf      Grund = "ueberlauf"
	GrundFrist          Grund = "frist"
	GrundAusgabeZuGross Grund = "ausgabe_zu_gross"
	GrundAblage         Grund = "ablage"
	GrundAbbruch        Grund = "abbruch"
)

// Ergebnis ist der Ausgang der Aufnahme eines Bodys (SPEC-119 Nr. 11). Fehler ist nur bei
// Abbruch und Verstoss gesetzt; Redigiert nur bei Richtung Request und Body present.
type Ergebnis struct {
	Richtung  Richtung
	Ausgang   Ausgang
	Grund     Grund
	Body      model.BodyState
	Blob      string
	Redaction model.RedactionStatus
	Integrity model.BodyIntegrity
	Fehler    *Fehler
	Redigiert []byte
}

// Fehler ist die Ursache eines Abbruchs oder Verstoßes: ODR-E2401 mit Grund. Sein Text nennt
// den Grund und nie ein Byte des Bodys, eines Headers oder eines Pfades; die Ursache bleibt über
// errors.Is erreichbar.
type Fehler struct {
	Grund   Grund
	ursache error
}

// Error nennt den Code und den Grund.
func (e *Fehler) Error() string {
	return model.CodeE2401.String() + ": Capture fehlgeschlagen: " + string(e.Grund)
}

// Code liefert ODR-E2401 (SPEC-109).
func (*Fehler) Code() model.Code { return model.CodeE2401 }

// Unwrap macht die Ursache erreichbar.
func (e *Fehler) Unwrap() error { return e.ursache }

// verworfen ist das Ergebnis eines Abbruchs oder Verstoßes: der Body ist verworfen, auch wenn
// kein Byte ankam (SPEC-119 Nr. 11).
func verworfen(r Richtung, b befund) Ergebnis {
	return Ergebnis{
		Richtung:  r,
		Ausgang:   b.ausgang,
		Grund:     b.grund,
		Body:      model.BodyDiscarded,
		Redaction: model.RedactionUnsafe,
		Integrity: model.IntegrityAbsent,
		Fehler:    &Fehler{Grund: b.grund, ursache: b.ursache},
	}
}

// ohneBody ist das Ergebnis eines Bodys ohne ein Byte.
func ohneBody(r Richtung) Ergebnis {
	return Ergebnis{
		Richtung:  r,
		Ausgang:   Vollstaendig,
		Body:      model.BodyAbsent,
		Redaction: model.RedactionNone,
		Integrity: model.IntegrityAbsent,
	}
}

// befund ist die Ursache eines Abbruchs oder Verstoßes, bevor sie ein Ergebnis wird.
type befund struct {
	ausgang Ausgang
	grund   Grund
	ursache error
}
