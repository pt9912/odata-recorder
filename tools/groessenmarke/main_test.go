package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMesspunktZeile(t *testing.T) {
	// a = 3, b = 5, f = 7, p = 11: S = 255 + 400 + 1400 + 11 = 2066; der Block nennt 999.
	plan := planMit(tabelleText(3, 5), blockB("produktcode: 11", "fakes: 7", "schaetzung: 999", "marke: 1400", "schnitt-grund: Grund"))
	for _, m := range []int{13, 1401} {
		r := pr(plan, messungB(m))
		felder := []struct{ k, w string }{
			{"zeilen", "8"}, {"nebenlaeufig", "3"}, {"fakes", "7"}, {"produktcode", "11"},
			{"schaetzung", "2066"}, {"marke", "1400"}, {"messung", map[int]string{13: "13", 1401: "1401"}[m]},
		}
		zeile := ""
		for _, z := range strings.Split(r.out, "\n") {
			if strings.HasPrefix(z, "Messpunkt p.md: ") {
				zeile = z
			}
		}
		erwartet := "Messpunkt p.md:"
		for _, f := range felder {
			erwartet += " " + f.k + "=" + f.w
			if got := feld(r.out, f.k); got != f.w {
				t.Errorf("Messung %d: %s=%s statt %s", m, f.k, got, f.w)
			}
		}
		if zeile != erwartet {
			t.Errorf("Messung %d: Zeile %q statt %q", m, zeile, erwartet)
		}
		if m == 1401 && r.code != 1 {
			t.Errorf("Messung %d: Exit %d statt 1 (Zeile auch bei Befund)", m, r.code)
		}
	}
}

func TestExitUndStroeme(t *testing.T) {
	ok := pr(planB(), messungB(1100))
	ok.soll(t, "ok", 0, "p.md: ok, Messung 1100 von Marke 1100", "Messpunkt p.md:")
	if ok.err != "" {
		t.Errorf("ok: stderr %q", ok.err)
	}
	hinweis := pr(planB("schaetzung: 799"), messungB(5))
	hinweis.soll(t, "ok mit Hinweis", 0, "p.md: Hinweis: Tabelle ergibt 800, Schätzung 799", "p.md: ok, Messung 5 von Marke 1100")
	if hinweis.err != "" || strings.Index(hinweis.out, "Hinweis") > strings.Index(hinweis.out, "ok, Messung") {
		t.Errorf("Hinweis: stderr %q, stdout %q", hinweis.err, hinweis.out)
	}
	eins := pr(planB(), messungB(1101))
	eins.soll(t, "1 Befund", 1, "Messpunkt p.md:", "!ok, Messung")
	if strings.Count(eins.err, "\n") != 1 || strings.Contains(eins.out, "ohne Eintrag") || !strings.HasPrefix(eins.err, "groessenmarke p.md: Messung 1101") {
		t.Errorf("1 Befund: stdout %q, stderr %q", eins.out, eins.err)
	}
	plan, m := befundPlan([5]bool{true, true, true, false, true})
	vier := pr(plan, m)
	vier.soll(t, "4 Befunde", 1, "!ok, Messung")
	if strings.Count(vier.err, "\n") != 4 || strings.Contains(vier.out, "weicht") || strings.Contains(vier.out, "Schnittfrage") {
		t.Errorf("4 Befunde: stdout %q, stderr %q", vier.out, vier.err)
	}
	fehler := []struct{ name, plan, messung, text string }{
		{"Tabelle fehlt", strings.Replace(planB(), "| Nr. |", "| Nummer |", 1), messungB(0), "Eingabe: Tabelle fehlt"},
		{"Block mehrdeutig", planB() + "```groessenmarke\n```\n", messungB(0), "Eingabe: Block mehrdeutig"},
		{"Zahl ungültig", planB("marke: 11,00"), messungB(0), "Eingabe: Zahl ungültig: marke"},
		{"Messung fremd", planB(), kopfText("abcdef0", spitzeB, "src/"), "Eingabe: " + fremdePlan},
		{"Messung unlesbar", planB(), "unsinn", "Eingabe: " + nichtLesbar},
	}
	for _, f := range fehler {
		r := pr(f.plan, f.messung)
		r.soll(t, f.name, 2, "groessenmarke p.md: "+f.text)
		if r.out != "" {
			t.Errorf("%s: stdout %q", f.name, r.out)
		}
	}
}

// lauf ruft run mit den Argumenten.
func lauf(args ...string) ergebnis {
	var o, e bytes.Buffer
	c := run(args, &o, &e)
	return ergebnis{c, o.String(), e.String()}
}

func schreibe(t *testing.T, dir, name, inhalt string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(inhalt), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestAufruf(t *testing.T) {
	d := t.TempDir()
	plan := schreibe(t, d, "slice-x.md", planB())
	m := schreibe(t, d, "m", messungB(0))
	formen := [][]string{
		nil, {"unbekannt", plan}, {"abfrage"}, {"abfrage", plan, "x"}, {"abfrage", plan, m}, {"pruefen", plan}, {"pruefen"},
		{"pruefen", plan, m, "x"}, {"abfrage", ""}, {"pruefen", "", m}, {"pruefen", plan, ""}, {"", plan},
	}
	for _, a := range formen {
		r := lauf(a...)
		r.soll(t, strings.Join(a, " "), 2, "Aufruf: groessenmarke abfrage <Plan>", "groessenmarke pruefen <Plan> <Messung>")
		if r.out != "" {
			t.Errorf("%q: stdout %q", a, r.out)
		}
	}
	fehlt := filepath.Join(d, "gibt-es-nicht")
	lauf("abfrage", fehlt).soll(t, "Plan fehlt abfrage", 2, "Plan fehlt: "+fehlt)
	lauf("pruefen", fehlt, m).soll(t, "Plan fehlt pruefen", 2, "Plan fehlt: "+fehlt)
	lauf("pruefen", fehlt, fehlt).soll(t, "beides fehlt", 2, "Plan fehlt: "+fehlt, "!Messung fehlt")
	lauf("pruefen", plan, fehlt).soll(t, "Messung fehlt", 2, "Messung fehlt: "+fehlt, "!Plan fehlt")
	lauf("pruefen", d, m).soll(t, "Plan ist ein Verzeichnis", 2, "Plan fehlt: "+d)
	lauf("abfrage", plan).soll(t, "abfrage", 0, "messen\nbasis "+basisB+"\npfad src/\n")
	lauf("pruefen", plan, m).soll(t, "pruefen", 0, "ok, Messung 0 von Marke 1100")
}

// stand ist der Inhalt eines Verzeichnisses: Name und Bytes jeder Datei.
func stand(t *testing.T, dir string) map[string]string {
	t.Helper()
	es, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, e := range es {
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		out[e.Name()] = string(b)
	}
	return out
}

func gleich(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || w != v {
			return false
		}
	}
	return true
}

func TestSchreibtNichts(t *testing.T) {
	t.Chdir(t.TempDir())
	d := t.TempDir()
	plaene := []struct{ plan, messung string }{
		{planB(), messungB(0)}, {planB(), messungB(1101)}, {planB("marke: 1.1"), messungB(0)},
		{planB("-basis", "-schaetzung", "-marke"), messungB(0)}, {"# P\n", messungB(0)},
	}
	for i, p := range plaene {
		schreibe(t, d, "slice-"+string(rune('a'+i))+".md", p.plan)
		schreibe(t, d, "m"+string(rune('a'+i)), p.messung)
	}
	vorher, vorherCwd := stand(t, d), stand(t, ".")
	n := 0
	for i := range plaene {
		lauf("abfrage", filepath.Join(d, "slice-"+string(rune('a'+i))+".md"))
		lauf("pruefen", filepath.Join(d, "slice-"+string(rune('a'+i))+".md"), filepath.Join(d, "m"+string(rune('a'+i))))
		n++
	}
	if n != 5 || !gleich(vorher, stand(t, d)) || !gleich(vorherCwd, stand(t, ".")) {
		t.Errorf("%d Pläne; Verzeichnisinhalt verändert", n)
	}
}

func TestPruefenOhneVollstaendigenBlock(t *testing.T) {
	pr(planB("-basis", "-schaetzung", "-marke"), messungB(0)).soll(t, "ausstehend", 2, "Eingabe: kein vollständiger Block")
	pr("# P\n", messungB(0)).soll(t, "kein Block", 2, "Eingabe: kein vollständiger Block")
}

func TestWiederholung(t *testing.T) {
	plan, m := befundPlan([5]bool{true, true, true, false, true})
	erste := pr(plan, m)
	for i := 0; i < 50; i++ {
		if r := pr(plan, m); r != erste {
			t.Fatalf("Lauf %d weicht ab:\n%v\nstatt\n%v", i, r, erste)
		}
	}
	a := ab(planB())
	for i := 0; i < 20; i++ {
		if r := ab(planB()); r != a {
			t.Fatalf("abfrage, Lauf %d weicht ab", i)
		}
	}
}
