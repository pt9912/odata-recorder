package jsoncodec

import "strings"

// form ist das Ergebnis der Aushandlung (SPEC-169 Nr. 3): die Stufe und die Zahlform.
type form struct {
	keine        bool // Stufe `none`, sonst `minimal`
	zeichenkette bool // Zahlform `Zeichenkette` (`IEEE754Compatible=true`), sonst `Zahl`
}

// aushandeln wertet die Werte der Kopfzeilen `Accept` aus: der erste Bereich mit dem Medientyp `application/json`
// bestimmt die Form, jede Eingabe ergibt eine; ein Gewicht und weitere Bereiche gelten nicht.
func aushandeln(accept []string) form {
	for _, wert := range accept {
		for _, bereich := range strings.Split(wert, ",") {
			teile := strings.Split(bereich, ";")
			if !strings.EqualFold(randTrim(teile[0]), "application/json") {
				continue
			}
			return parameter(teile[1:])
		}
	}
	return form{}
}

func randTrim(s string) string { return strings.Trim(s, " \t") }

// parameter liest `odata.metadata` und `IEEE754Compatible`, jeweils das erste Vorkommen.
func parameter(teile []string) form {
	var f form
	var metadata, ieee bool
	for _, p := range teile {
		name, wert, _ := strings.Cut(p, "=")
		name, wert = randTrim(name), randTrim(wert)
		switch {
		case !metadata && strings.EqualFold(name, "odata.metadata"):
			metadata, f.keine = true, strings.EqualFold(wert, "none")
		case !ieee && strings.EqualFold(name, "IEEE754Compatible"):
			ieee, f.zeichenkette = true, strings.EqualFold(wert, "true")
		}
	}
	return f
}

// contentType ist der Wert von `Content-Type` von Entity und Collection (SPEC-169 Nr. 3).
func (f form) contentType() string {
	stufe := "minimal"
	if f.keine {
		stufe = "none"
	}
	if f.zeichenkette {
		return "application/json;odata.metadata=" + stufe + ";IEEE754Compatible=true;charset=utf-8"
	}
	return "application/json;odata.metadata=" + stufe + ";charset=utf-8"
}
