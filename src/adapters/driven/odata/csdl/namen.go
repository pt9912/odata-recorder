package csdl

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Namensräume der Wurzel und der Elemente (SPEC-125 Nr. 6, 7).
const (
	nsEdm   = "http://docs.oasis-open.org/odata/ns/edm"
	nsEdmx  = "http://docs.oasis-open.org/odata/ns/edmx"
	nsEdmxV = "http://schemas.microsoft.com/ado/2007/06/edmx"
	nsXML   = "http://www.w3.org/XML/1998/namespace"
)

// Grenzen des Lesers (SPEC-125 Nr. 5).
const (
	maxTiefe      = 64
	maxKette      = 64
	maxMembers    = 65536
	maxName       = 128
	maxNamensraum = 256
)

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

// bekannteEDMTypen sind die 34 Namen von Nr. 10.
func bekannteEDMTypen() []string {
	var out []string
	for _, p := range []string{
		"Binary", "Boolean", "Byte", "Date", "DateTimeOffset", "Decimal", "Double", "Duration", "Guid",
		"Int16", "Int32", "Int64", "SByte", "Single", "Stream", "String", "TimeOfDay",
	} {
		out = append(out, "Edm."+p)
	}
	for _, g := range []string{"Geography", "Geometry"} {
		out = append(out, "Edm."+g)
		for _, e := range []string{"Point", "LineString", "Polygon", "MultiPoint", "MultiLineString", "MultiPolygon", "Collection"} {
			out = append(out, "Edm."+g+e)
		}
	}
	return append(out, "Edm.Untyped")
}
