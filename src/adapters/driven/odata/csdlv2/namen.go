package csdlv2

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Namensräume der Wurzel und der Elemente (SPEC-171 Nr. 2 bis 4).
const (
	nsEdmx   = "http://schemas.microsoft.com/ado/2007/06/edmx"
	nsEdmxV4 = "http://docs.oasis-open.org/odata/ns/edmx"
	nsMeta   = "http://schemas.microsoft.com/ado/2007/08/dataservices/metadata"
	nsEdm3   = "http://schemas.microsoft.com/ado/2009/11/edm"
	nsEdmV4  = "http://docs.oasis-open.org/odata/ns/edm"
	nsXML    = "http://www.w3.org/XML/1998/namespace"
)

// Grenzen des Lesers (SPEC-125 Nr. 5, SPEC-171 Nr. 1).
const (
	maxTiefe      = 64
	maxKette      = 64
	maxMembers    = 65536
	maxBindings   = 65536
	maxName       = 128
	maxNamensraum = 256
)

// istEdm sagt, ob ns einer der drei Namensräume von CSDL 1.0 bis 2.0 ist (Nr. 3).
func istEdm(ns string) bool {
	switch ns {
	case "http://schemas.microsoft.com/ado/2006/04/edm", "http://schemas.microsoft.com/ado/2007/05/edm", "http://schemas.microsoft.com/ado/2008/09/edm":
		return true
	}
	return false
}

// istSpaeteresEdm sagt, ob ns der Namensraum von CSDL 3.0 oder von V4 ist (Nr. 3: `version`).
func istSpaeteresEdm(ns string) bool { return ns == nsEdm3 || ns == nsEdmV4 }

// metaAttributOhneWirkung sind die drei Attribute des Namensraums `m`, die nie ein Fund sind (Nr. 4).
func metaAttributOhneWirkung(local string) bool {
	return slices.Contains([]string{"DataServiceVersion", "MaxDataServiceVersion", "IsDefaultEntityContainer"}, local)
}

// einfacherName prüft Nr. 8 b: 1 bis 128 Zeichen, Buchstabe oder `_`, dann Buchstaben, Ziffern, `_`.
func einfacherName(s string) bool {
	n := utf8.RuneCountInString(s)
	if n < 1 || n > maxName {
		return false
	}
	for i, r := range s {
		if r == '_' || unicode.IsLetter(r) || (i > 0 && unicode.IsDigit(r)) {
			continue
		}
		return false
	}
	return true
}

// namensraumGueltig prüft jedes Glied als einfachen Namen und die Länge des Ganzen.
func namensraumGueltig(s string) bool {
	if utf8.RuneCountInString(s) > maxNamensraum {
		return false
	}
	for _, g := range strings.Split(s, ".") {
		if !einfacherName(g) {
			return false
		}
	}
	return true
}

// bekannteEDMTypen sind die 15 Primitiven von Nr. 5.
func bekannteEDMTypen() []string {
	var out []string
	for _, p := range []string{
		"Binary", "Boolean", "Byte", "DateTime", "DateTimeOffset", "Decimal", "Double", "Guid",
		"Int16", "Int32", "Int64", "SByte", "Single", "String", "Time",
	} {
		out = append(out, "Edm."+p)
	}
	return out
}

// edmTypImModell bildet einen bekannten Namen auf den Typ im Modell ab (Nr. 5): `Edm.DateTime` auf
// `Edm.DateTimeOffset`, das `Edm.DateTimeOffset` des Dokuments auf einen Namen, der keiner der acht Typen ist.
func edmTypImModell(name string) string {
	switch name {
	case "Edm.DateTime":
		return "Edm.DateTimeOffset"
	case "Edm.DateTimeOffset":
		return "Edm.DateTimeOffset(V2)"
	}
	return name
}
