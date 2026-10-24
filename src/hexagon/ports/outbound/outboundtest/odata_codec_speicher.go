package outboundtest

import (
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// l1Grenze ist die Obergrenze des Heaps über dem Stand vor dem Lesen für ein gültiges Ergebnis aus rund 4 MiB Body.
const l1Grenze = 256 << 20

// spitze fährt f und liefert den höchsten Stand des Heaps, den ein Abtaster dabei sah, über dem Stand vor dem Lauf.
func spitze(f func()) uint64 {
	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	vorher := ms.HeapAlloc
	var hoechst atomic.Uint64
	ende := make(chan struct{})
	fertig := make(chan struct{})
	go func() {
		defer close(fertig)
		tick := time.NewTicker(time.Millisecond)
		defer tick.Stop()
		var m runtime.MemStats
		for {
			select {
			case <-ende:
				return
			case <-tick.C:
			}
			runtime.ReadMemStats(&m)
			if m.HeapAlloc > hoechst.Load() {
				hoechst.Store(m.HeapAlloc)
			}
		}
	}()
	f()
	close(ende)
	<-fertig
	if h := hoechst.Load(); h > vorher {
		return h - vorher
	}
	return 0
}

// jL1: der Speicher des Ergebnisses wächst mit dem Body, nicht schneller (SPEC-126 Nr. 5, Keine Grenze): ein gültiger Body
// aus 100 000 kleinen Elementen und einer aus einem Element mit 100 000 kurzen Properties eines offenen Typs.
func jL1(t *testing.T, s *jsuite) {
	var elems strings.Builder
	for i := range 100000 {
		if i > 0 {
			elems.WriteByte(',')
		}
		fmt.Fprintf(&elems, `{"ID":%d,"Name":"name-%d","Price":1.5}`, i, i)
	}
	viele := jfall{name: "L1 100000 Elemente", m: "R", set: "Products", body: []byte(`{"@odata.context":"$metadata#Products","value":[` + elems.String() + `]}`), grenze: 8 << 20}
	var members strings.Builder
	for i := range 100000 {
		fmt.Fprintf(&members, `,"p%d":%d`, i, i)
	}
	offen := jfall{name: "L1 100000 Properties", m: "Z", set: "Animals", body: []byte(`{"@odata.context":"$metadata#Animals","value":[{"@odata.type":"#Zoo.Mammal","Id":1` + members.String() + `}]}`), grenze: 8 << 20}
	for _, f := range []jfall{viele, offen} {
		var r jergebnis
		hoch := spitze(func() { r = s.lauf(f) })
		if r.err != nil {
			t.Errorf("%s: %v", f.name, r.err)
		}
		t.Logf("%s: %d Bytes, Spitze des Heaps %d MiB", f.name, len(f.body), hoch>>20)
		if hoch > l1Grenze {
			t.Errorf("%s: Spitze des Heaps %d MiB über %d MiB", f.name, hoch>>20, l1Grenze>>20)
		}
		if f.m == "R" && len(r.seite.Entities) != 100000 {
			t.Errorf("%s: %d Entities", f.name, len(r.seite.Entities))
		}
		if f.m == "Z" && (len(r.seite.Entities) != 1 || len(r.seite.Entities[0].Uebersprungen) != 100000 || r.seite.Entities[0].Typ != "Zoo.Mammal") {
			t.Errorf("%s: %d Entities, Ergebnis %+v", f.name, len(r.seite.Entities), len(r.seite.Entities))
		}
	}
}
