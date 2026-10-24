package model

// WertArt ist die Art eines gelesenen Wertes (SPEC-126 Nr. 10); für die acht EDM-Typen steht die Art des Typs,
// ArtNull für `null` und ArtRedigiert für den Marker.
type WertArt string

// Die zehn Arten.
const (
	ArtNull           WertArt = "null"
	ArtString         WertArt = "string"
	ArtBoolean        WertArt = "boolean"
	ArtInt32          WertArt = "int32"
	ArtInt64          WertArt = "int64"
	ArtDecimal        WertArt = "decimal"
	ArtGUID           WertArt = "guid"
	ArtDate           WertArt = "date"
	ArtDateTimeOffset WertArt = "datetimeoffset"
	ArtRedigiert      WertArt = "redigiert"
)

// Wert ist ein gelesener Wert; Text ist seine kanonische Schreibweise (SPEC-126 Nr. 10), bei ArtNull und
// ArtRedigiert leer.
type Wert struct {
	Art  WertArt
	Text string
}

// PropertyWert ist eine Property, die der Body nennt, mit ihrem Wert.
type PropertyWert struct {
	Name string
	Wert Wert
}

// UebersprungGrund ist der Grund, aus dem der Leser den Wert eines Mitglieds nicht deutet (SPEC-126 Nr. 9).
type UebersprungGrund string

// Die drei Gründe.
const (
	UebersprungTyp        UebersprungGrund = "typ"
	UebersprungNavigation UebersprungGrund = "navigation"
	UebersprungOffen      UebersprungGrund = "offen"
)

// Uebersprungen ist ein Mitglied, dessen Wert der Leser nicht übernimmt.
type Uebersprungen struct {
	Name  string
	Grund UebersprungGrund
}

// Entity ist eine gelesene Entity (SPEC-126 Nr. 9 bis 12): Typ ist der qualifizierte Name, Schluessel die
// Literalform der URL ohne Klammern. Properties und Uebersprungen stehen in der Reihenfolge des Dokuments und
// sind nie nil, wenn der Leser die Entity geliefert hat.
type Entity struct {
	Typ           string
	Schluessel    string
	Properties    []PropertyWert
	Uebersprungen []Uebersprungen
}

// FolgeLink ist der Nachweis `@odata.nextLink`: der rohe Text, oder Redigiert, wenn der Marker an seiner Stelle steht.
type FolgeLink struct {
	Vorhanden bool
	Redigiert bool
	Wert      string
}

// Anzahl ist der Nachweis `@odata.count`.
type Anzahl struct {
	Vorhanden bool
	Wert      int64
}

// EntitySeite ist die gelesene Collection; Entities ist bei leerem `value` leer und nicht nil.
type EntitySeite struct {
	Entities  []Entity
	FolgeLink FolgeLink
	Anzahl    Anzahl
}
