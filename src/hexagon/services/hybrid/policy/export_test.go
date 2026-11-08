package policy

// Hilfen für die externen Tests: die geparsten Formen der Einträge als einfache Werte.

// Segment ist die lesbare Form eines Segments des Musters.
type Segment struct {
	Art          int
	Anfang, Ende string
}

// Arten eines Segments.
const (
	SegLiteral = int(segLiteral)
	SegStern   = int(segStern)
	SegRest    = int(segRest)
)

// Namen der beiden Kategorien.
const (
	KategorieUnsupported = kategorieUnsupported
	KategorieCoverage    = kategorieCoverage
)

func segmenteVon(segs []segment) []Segment {
	if segs == nil {
		return nil
	}
	aus := make([]Segment, len(segs))
	for i, s := range segs {
		aus[i] = Segment{Art: int(s.art), Anfang: s.anfang, Ende: s.ende}
	}
	return aus
}

// ParseHost liefert Name, Port und Ursache eines Eintrags von Hosts.
func ParseHost(e string) (name string, port int, ursache string) {
	h, u := parseHost(e)
	return h.name, h.port, u
}

// ParseFreigabe liefert Methode, Segmente und Ursache eines Eintrags von Schreibfreigaben.
func ParseFreigabe(e string) (methode string, segs []Segment, ursache string) {
	f, u := parseFreigabe(e)
	return f.methode, segmenteVon(f.segmente), u
}

// ParseMuster liefert die Segmente eines Musters.
func ParseMuster(m string) ([]Segment, bool) {
	segs, ok := parseMuster(m)
	return segmenteVon(segs), ok
}

// ZielName liest den Namen eines Ziels aus dem Host mit Klammern (roh) und dem Hostnamen ohne sie.
func ZielName(roh, host string) (string, bool) { return zielName(roh, host) }

// PruefeKategorie liefert die Ursache eines Eintrags von ReplayKategorien.
func PruefeKategorie(e string) string { return pruefeKategorie(e) }

// ReplayFlags liefert, welche Kategorien die Policy hält.
func (p *Policy) ReplayFlags() (unsupported, coverage bool) {
	return p.replayUnsupported, p.replayCoverage
}
