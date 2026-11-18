// ratedrv <liste.json> <paket> <N>: je Mutation der Liste wird das Testbinary ohne und mit -race gebaut; je N Prozesse (-test.count=1)
// mit dem Muster der Mutation, gezählt werden die roten. Danach wird die Datei zurückgeschrieben. Ein unveränderter Stand läuft als
// Nullprobe mit. Eine Mutation kann weitere Stellen derselben Datei nennen (Zusatz).
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type stelle struct{ Alt, Neu string }

type mut struct {
	Name, Datei, Alt, Neu, Test string
	Zusatz                      []stelle
}

func lauf(name string, args ...string) (string, error) {
	c := exec.Command(name, args...)
	var aus bytes.Buffer
	c.Stdout, c.Stderr = &aus, &aus
	err := c.Run()
	return aus.String(), err
}

func rate(bin, test string, n int) (rot int, racemeldung int) {
	for range n {
		out, err := lauf(bin, "-test.run", test, "-test.count=1")
		if err != nil {
			rot++
		}
		if strings.Contains(out, "WARNING: DATA RACE") {
			racemeldung++
		}
	}
	return
}

func main() {
	roh, _ := os.ReadFile(os.Args[1])
	var l struct {
		Mutationen []mut `json:"mutationen"`
	}
	if err := json.Unmarshal(roh, &l); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	paket := os.Args[2]
	n, _ := strconv.Atoi(os.Args[3])
	ms := append([]mut{{Name: "nullprobe"}}, l.Mutationen...)
	for _, m := range ms {
		var orig []byte
		if m.Datei != "" {
			orig, _ = os.ReadFile(m.Datei)
			neu := string(orig)
			ok := true
			for _, s := range append([]stelle{{m.Alt, m.Neu}}, m.Zusatz...) {
				if strings.Count(neu, s.Alt) != 1 {
					fmt.Println(m.Name, "alt trifft nicht genau einmal:", s.Alt)
					ok = false
					break
				}
				neu = strings.Replace(neu, s.Alt, s.Neu, 1)
			}
			if !ok {
				continue
			}
			_ = os.WriteFile(m.Datei, []byte(neu), 0o644)
		}
		test := m.Test
		if test == "" {
			test = l.Mutationen[0].Test
		}
		o1, e1 := lauf("go", "test", "-c", "-o", "/tmp/plain.test", paket)
		o2, e2 := lauf("go", "test", "-race", "-c", "-o", "/tmp/race.test", paket)
		if m.Datei != "" {
			_ = os.WriteFile(m.Datei, orig, 0o644)
		}
		if e1 != nil || e2 != nil {
			fmt.Println(m.Name, "bau", o1, o2)
			continue
		}
		// Die Testbinaries laufen im Verzeichnis ihres Pakets.
		_ = os.Chdir(strings.TrimPrefix(paket, "./"))
		r1, _ := rate("/tmp/plain.test", test, n)
		r2, rm := rate("/tmp/race.test", test, n)
		_ = os.Chdir("/work/baum")
		fmt.Printf("%s ohne Detektor rot %d von %d; mit Detektor rot %d von %d (Meldung DATA RACE in %d)\n", m.Name, r1, n, r2, n, rm)
	}
}
