// mutdrv fährt eine Mutationsliste (Format SPEC-210 Nr. 2) auf dem Baum im Arbeitsverzeichnis:
// je Mutation alt -> neu (genau einmal), nur der genannte Test, danach wird die Datei zurückgeschrieben.
// Es gibt keine Zeitgrenze von 2 Minuten für das Paket; der Treiber ersetzt die Nullprobe durch einen Lauf
// der genannten Tests auf dem unveränderten Stand.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type mut struct {
	Name, Datei, Alt, Neu, Test, Klasse string
}

func last() string {
	b, _ := os.ReadFile("/proc/loadavg")
	f := strings.Fields(string(b))
	if len(f) == 0 {
		return "?"
	}
	return f[0]
}

func lauf(paket, re string) (string, time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	t0 := time.Now()
	c := exec.CommandContext(ctx, "go", "test", "-count=1", "-run", re, paket)
	var aus bytes.Buffer
	c.Stdout, c.Stderr = &aus, &aus
	_ = c.Run()
	return aus.String(), time.Since(t0)
}

func main() {
	roh, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("Liste:", err)
		os.Exit(2)
	}
	var l struct {
		Mutationen []mut `json:"mutationen"`
	}
	if err := json.Unmarshal(roh, &l); err != nil {
		fmt.Println("Liste:", err)
		os.Exit(2)
	}
	pakete := map[string]string{}
	var namen []string
	for _, m := range l.Mutationen {
		if _, da := pakete[m.Test]; !da {
			pakete[m.Test] = m.Datei
			namen = append(namen, m.Test)
		}
	}
	paket := "./src/bootstrap/"
	if len(os.Args) > 2 {
		paket = os.Args[2]
	}
	out, d := lauf(paket, namen[0])
	fmt.Printf("nullprobe tests=%d dauer=%.1f ok=%v\n", len(namen), d.Seconds(), strings.Contains(out, "\nok ") || strings.HasPrefix(out, "ok "))
	if !(strings.Contains(out, "\nok ") || strings.HasPrefix(out, "ok ")) {
		fmt.Println(out)
		os.Exit(2)
	}
	n := map[string]int{}
	for _, m := range l.Mutationen {
		orig, err := os.ReadFile(m.Datei)
		if err != nil {
			fmt.Println(m.Name, "Datei fehlt")
			continue
		}
		if c := strings.Count(string(orig), m.Alt); c != 1 {
			fmt.Printf("mutation %s ergebnis=eingabe alt-trifft=%d\n", m.Name, c)
			n["eingabe"]++
			continue
		}
		l0 := last()
		if err := os.WriteFile(m.Datei, []byte(strings.Replace(string(orig), m.Alt, m.Neu, 1)), 0o644); err != nil {
			fmt.Println(m.Name, err)
			continue
		}
		out, d := lauf(paket, m.Test)
		_ = os.WriteFile(m.Datei, orig, 0o644)
		e := "abbruch"
		switch {
		case strings.Contains(out, "[build failed]") || strings.Contains(out, "[setup failed]"):
			e = "uebersetzungsfehler"
		case strings.Contains(out, "--- FAIL: "):
			e = "rot"
		case strings.Contains(out, "\nok ") || strings.HasPrefix(out, "ok "):
			e = "gruen"
		}
		n[e]++
		fmt.Printf("mutation %s ergebnis=%s test=%s dauer=%.1f last=%s\n", m.Name, e, m.Test, d.Seconds(), l0)
		k := 0
		for _, z := range strings.Split(out, "\n") {
			if k < 4 && (strings.HasPrefix(z, "--- FAIL") || strings.Contains(z, "_test.go:") || strings.Contains(z, "build failed") || strings.HasPrefix(z, "# ") || strings.Contains(z, ".go:") && e == "uebersetzungsfehler") {
				fmt.Printf("    %s\n", strings.TrimSpace(z))
				k++
			}
		}
	}
	fmt.Printf("zusammenfassung %v\n", n)
}
