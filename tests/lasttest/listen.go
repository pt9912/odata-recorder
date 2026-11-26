package main

import "fmt"

// Betriebsart ist die Betriebsart des Binary, gegen das eine Liste läuft.
type Betriebsart string

// Die Betriebsarten, die ein Binary fährt.
const (
	betriebReplay Betriebsart = "replay"
	betriebMock   Betriebsart = "mock"
)

// Anfrage ist eine GET-Anfrage der Last: Pfad, rohe (prozent-kodierte) Query und Header in fester Folge.
type Anfrage struct {
	Pfad   string
	Query  string
	Header [][2]string
}

// Ziel ist die Anfragezeile ohne Schema und Host.
func (a Anfrage) Ziel() string {
	if a.Query == "" {
		return a.Pfad
	}
	return a.Pfad + "?" + a.Query
}

const (
	headerAccept    = "Accept"
	headerWorkspace = "X-Recorder-Workspace"
	headerRecording = "X-Recorder-Recording-Session"
	headerReplay    = "X-Recorder-Replay-Session"
	acceptJSON      = "application/json"
	// sitzungLaenge ist die Zahl der Anfragen einer Replay-Session.
	sitzungLaenge = 100
)

// seitenZiel ist die Anfrage der j-ten Seite (ab 0) des Sets.
func seitenZiel(p Profil, set, j int) Anfrage {
	a := Anfrage{Pfad: fmt.Sprintf("%s/Set%d", wurzel, set)}
	if j > 0 {
		a.Query = fmt.Sprintf("$skiptoken=%d", j*p.SeitenGroesse)
	}
	return a
}

// lookupZiel ist die Anfrage der Einzel-Entity.
func lookupZiel(set, id int) Anfrage {
	return Anfrage{Pfad: fmt.Sprintf("%s/Set%d(%d)", wurzel, set, id)}
}

// aufbauAnfragen sind die Exchanges des Aufbaus (SPEC-216 Nr. 6 d) in der Folge der Aufnahme: `$metadata`, je Set die Seiten, danach
// die ersten Lookups der Permutation.
func aufbauAnfragen(p Profil, startwert uint64) []Anfrage {
	out := []Anfrage{{Pfad: wurzel + "/$metadata"}}
	for set := 1; set <= p.Sets; set++ {
		for j := range p.SeitenJeSet() {
			out = append(out, seitenZiel(p, set, j))
		}
	}
	for _, s := range paare(p, startwert)[:p.Lookups] {
		out = append(out, lookupZiel(s[0], s[1]))
	}
	return out
}

// listeReplay ist die Liste der Anfragen der Betriebsart replay für den Lauf (ab 1; SPEC-216 Nr. 7 a): Anfrage i nimmt das Element i modulo
// Exchanges einer Permutation der Anfragen des Aufbaus, mit den Headern des Workspace, der Recording-Session und einer Replay-Session, die
// alle 100 Anfragen wechselt.
func listeReplay(p Profil, startwert uint64, lauf int, workspace, recording string) []Anfrage {
	menge := aufbauAnfragen(p, startwert)
	perm := permutation(startwert, stromReplay, len(menge))
	out := make([]Anfrage, p.Anfragen())
	for i := range out {
		out[i] = menge[perm[i%len(menge)]]
		out[i].Header = [][2]string{
			{headerAccept, acceptJSON}, {headerWorkspace, workspace}, {headerRecording, recording},
			{headerReplay, fmt.Sprintf("lasttest-%s-%d-%d", betriebReplay, lauf, i/sitzungLaenge)},
		}
	}
	return out
}

// listeMock ist die Liste der Anfragen der Betriebsart mock (SPEC-216 Nr. 7 b): die Art der Anfrage i ist i modulo 10, 0 bis 6 ein
// Einzel-Lookup, 7 und 8 eine Collection mit `$filter` über die Gruppe und `$top=10`, 9 eine Collection mit `$top` von 10 oder 50; Set,
// Id, Gruppe und die Größe der Seite stammen aus dem Startwert.
func listeMock(p Profil, startwert uint64, workspace string) []Anfrage {
	r := strom(startwert, stromMock)
	out := make([]Anfrage, p.Anfragen())
	for i := range out {
		set, id, g, n := 1+ziehe(r, p.Sets), 1+ziehe(r, p.Entities), ziehe(r, gruppen), 10+40*ziehe(r, 2)
		switch art := i % 10; {
		case art < 7:
			out[i] = lookupZiel(set, id)
		case art < 9:
			out[i] = Anfrage{Pfad: fmt.Sprintf("%s/Set%d", wurzel, set), Query: fmt.Sprintf("$filter=Gruppe%%20eq%%20%d&$top=10", g)}
		default:
			out[i] = Anfrage{Pfad: fmt.Sprintf("%s/Set%d", wurzel, set), Query: fmt.Sprintf("$top=%d", n)}
		}
		out[i].Header = [][2]string{{headerWorkspace, workspace}, {headerAccept, acceptJSON}}
	}
	return out
}

// umgebung ist die Umgebung des Binary (SPEC-216 Nr. 9 a): allein diese fünf Schlüssel; listen ist `127.0.0.1:<Port>`.
func umgebung(b Betriebsart, listen string) []string {
	return []string{"MODE=" + string(b), "LISTEN=" + listen, "DATA_DIR=/daten", "SERVICE_ID=lasttest", "SERVICE_ROOT_PATH=" + wurzel}
}
