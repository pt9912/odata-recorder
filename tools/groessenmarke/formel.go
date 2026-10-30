package main

import "fmt"

// Die Faktoren der Formel stehen in SPEC-128 Nr. 1 und 2; der Zuschlag 1,35 ist 135/100.
const (
	faktorNebenlaeufig = 85
	faktorAndere       = 80
	faktorFake         = 200
	zuschlagProzent    = 135
	schnittGrenze      = 300000
)

// schaetzung ist S = 85·a + 80·b + 200·f + p (SPEC-128 Nr. 1).
func schaetzung(a, b, f, p int64) int64 {
	return faktorNebenlaeufig*a + faktorAndere*b + faktorFake*f + p
}

// markeZu ist die kleinste volle Hundert, die 1,35·S nicht unterschreitet, ganzzahlig gerechnet.
func markeZu(s int64) int64 {
	return 100 * ((zuschlagProzent*s + 9999) / 10000)
}

// schnittfrage gilt genau dann, wenn 135·S über 300000 liegt.
func schnittfrage(s int64) bool {
	return zuschlagProzent*s > schnittGrenze
}

// aufHundert rundet m auf die nächste volle Hundert nach oben.
func aufHundert(m int64) int64 {
	return (m + 99) / 100 * 100
}

// pruefung sammelt, was die Prüfungen von SPEC-128 Nr. 5 melden.
type pruefung struct {
	hinweise []string
	befunde  []string
}

// pruefe wendet Nr. 5 auf einen vollständigen Plan und die Messung an; die Befunde stehen in der
// Reihenfolge Schnitt, Marke, Schätzung, Überschreitung, Grund.
func pruefe(p plan, messung int64) pruefung {
	var r pruefung
	b := p.blk
	if schnittfrage(b.schaetzung) && b.schnittGrund == "" {
		r.befunde = append(r.befunde, "Schnittfrage ohne Grund")
	}
	if erwartet := markeZu(b.schaetzung); b.marke != erwartet {
		r.befunde = append(r.befunde, fmt.Sprintf("Marke %d weicht von der Formel ab (Schätzung %d, erwartet %d)",
			b.marke, b.schaetzung, erwartet))
	}
	r.schaetzungGegenTabelle(p, messung)
	r.ueberschreitung(b, messung)
	for _, e := range b.ueberschreitung {
		if e.grund == "" {
			r.befunde = append(r.befunde, fmt.Sprintf("ueberschreitung bis %d ohne Grund", e.n))
		}
	}
	for _, e := range b.ersetzt {
		if e.grund == "" {
			r.befunde = append(r.befunde, fmt.Sprintf("ersetzt %d ohne Grund", e.n))
		}
	}
	if messung <= b.marke && len(b.ueberschreitung) > 0 {
		r.hinweise = append(r.hinweise, "Eintrag ohne Überschreitung")
	}
	return r
}

// schaetzungGegenTabelle vergleicht die Schätzung des Blocks mit der Tabelle (Nr. 5 c): vor dem
// ersten Code-Commit (Messung 0) ein Befund, im Code ein Hinweis.
func (r *pruefung) schaetzungGegenTabelle(p plan, messung int64) {
	s := p.tabellenSchaetzung()
	if p.blk.schaetzung == s {
		return
	}
	if messung >= 1 {
		r.hinweise = append(r.hinweise, fmt.Sprintf("Tabelle ergibt %d, Schätzung %d", s, p.blk.schaetzung))
		return
	}
	r.befunde = append(r.befunde, fmt.Sprintf("Schätzung %d weicht von der Tabelle ab (Tabelle ergibt %d)",
		p.blk.schaetzung, s))
}

// ueberschreitung prüft Nr. 5 d: über der Marke braucht die Messung einen Eintrag, dessen größtes
// N sie nicht unterschreitet; gleich der Marke ist keine Überschreitung.
func (r *pruefung) ueberschreitung(b block, messung int64) {
	if messung <= b.marke {
		return
	}
	zeile := fmt.Sprintf("eintragen: ueberschreitung: %d | <Grund>", aufHundert(messung))
	if len(b.ueberschreitung) == 0 {
		r.befunde = append(r.befunde, fmt.Sprintf("Messung %d über Marke %d ohne Eintrag; %s", messung, b.marke, zeile))
		return
	}
	var groesstes int64
	for _, e := range b.ueberschreitung {
		groesstes = max(groesstes, e.n)
	}
	if groesstes < messung {
		r.befunde = append(r.befunde, fmt.Sprintf("Messung %d über Marke %d: Eintrag reicht nicht (größtes N %d); %s",
			messung, b.marke, groesstes, zeile))
	}
}
