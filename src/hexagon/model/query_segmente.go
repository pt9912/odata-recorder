package model

import "strings"

// QuerySegment ist ein nichtleeres Segment der Query: der Name bis zum ersten `=` (das ganze Segment ohne `=`) und der
// Wert dahinter; beide undekodiert.
type QuerySegment struct {
	Name, Wert string
}

// QuerySegmente zerlegt die Query ohne Dekodierung an `&`; `;` trennt nichts, leere Segmente entfallen, ein Text ohne
// nichtleeres Segment ergibt nil (SPEC-162 Nr. 1). Es ist die Regel für Segment, Name und Wert der Query im Import
// und im Parser der OData-URL.
func QuerySegmente(raw string) []QuerySegment {
	n := 0
	for rest := raw; rest != ""; {
		var seg string
		seg, rest, _ = strings.Cut(rest, "&")
		if seg != "" {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	out := make([]QuerySegment, 0, n)
	for rest := raw; rest != ""; {
		var seg string
		seg, rest, _ = strings.Cut(rest, "&")
		if seg == "" {
			continue
		}
		name, wert, _ := strings.Cut(seg, "=")
		out = append(out, QuerySegment{Name: name, Wert: wert})
	}
	return out
}
