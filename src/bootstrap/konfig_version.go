package bootstrap

import "slices"

// schluesselVersion ist der Einzelschlüssel der Version des einen Services (SPEC-110 Nr. 17, SPEC-176), feldVersion das Feld
// desselben Werts je Eintrag der SERVICES_FILE.
const (
	schluesselVersion = "SERVICE_ODATA_VERSION"
	feldVersion       = "odata_version"
)

// Gründe der Verstöße von SPEC-110 Nr. 17; keiner nennt einen Wert.
const (
	grundVersion = "Wert nicht in der Aufzählung: " + version4 + ", " + version2
	grundHybrid  = "Modus hybrid für Version " + version2 + " nicht unterstützt"
)

// liesVersion liefert die wirksame Version zu einem Wert ohne Leerraum am Rand: leer gilt 4.0, 4.0 und 2.0 gelten, jeder andere
// Wert ist ein Verstoß am schluessel und liefert leer, damit er weder als 2.0 noch als 4.0 zählt (SPEC-110 Nr. 17 a und b).
func (p *pruefer) liesVersion(schluessel, wert string) string {
	switch wert {
	case "":
		return version4
	case version4, version2:
		return wert
	}
	p.melde(schluessel, grundVersion)
	return ""
}

// pruefeVersionModus meldet den Modus hybrid mit mindestens einem Service der Version 2.0 als einen Verstoß am Schlüssel MODE,
// gleich wie viele Services V2 sind; bei ungültigem MODE (leer) entfällt die Prüfung (SPEC-110 Nr. 8, 17 b).
func (p *pruefer) pruefeVersionModus(k *Konfiguration) {
	if k.Modus == modusHybrid && slices.ContainsFunc(k.Dienste, func(d Dienst) bool { return d.Version == version2 }) {
		p.melde("MODE", grundHybrid)
	}
}
