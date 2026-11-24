package main

import "math/rand/v2"

// Zwecke der Ströme: der Strom der Nutzlast trägt Set und Id oberhalb von 32 Bit, diese drei darunter.
const (
	stromPaare  uint64 = 1
	stromReplay uint64 = 2
	stromMock   uint64 = 3
)

const (
	gruppen        = 20
	nutzlastKlein  = 200
	nutzlastMittel = 8 << 10
	nutzlastGross  = 60 << 10
	// nutzlastZeichen ist die Zahl der Zeichen a-z, die ein gezogener Wert trägt (26 hoch 13 liegt unter 2 hoch 64).
	nutzlastZeichen = 13
)

// strom liefert den Generator eines Zwecks aus dem Startwert (SPEC-216 Nr. 6 c): allein PCG, nie der globale Generator.
func strom(startwert, zweck uint64) *rand.Rand { return rand.New(rand.NewPCG(startwert, zweck)) }

// ziehe liefert eine Zahl von 0 bis k-1.
func ziehe(r *rand.Rand, k int) int { return int(r.Uint64() % uint64(k)) }

// permutation liefert die Zahlen von 0 bis n-1 in einer Folge, die allein von Startwert und Zweck abhängt.
func permutation(startwert, zweck uint64, n int) []int {
	r := strom(startwert, zweck)
	p := make([]int, n)
	for i := range p {
		j := ziehe(r, i+1)
		p[i] = p[j]
		p[j] = i
	}
	return p
}

// nutzlastLaenge ist die Länge der Nutzlast der Entity id: 200 Bytes für 0 bis 6 modulo 10, 8 KiB für 7 und 8, 60 KiB für 9.
func nutzlastLaenge(id int) int {
	switch id % 10 {
	case 7, 8:
		return nutzlastMittel
	case 9:
		return nutzlastGross
	}
	return nutzlastKlein
}

// nutzlast liefert die Zeichen a-z der Entity (set, id) aus dem Startwert; sie hängen nicht von der Reihenfolge der Abfragen ab.
func nutzlast(startwert uint64, set, id int) []byte {
	r := strom(startwert, uint64(set)<<32|uint64(id))
	b := make([]byte, nutzlastLaenge(id))
	var block [nutzlastZeichen]byte
	for i := 0; i < len(b); i += nutzlastZeichen {
		z := r.Uint64()
		for j := range block {
			block[j] = 'a' + byte(z%26)
			z /= 26
		}
		copy(b[i:], block[:])
	}
	return b
}

// paare liefert alle Paare (Set, Id) des Profils in der Folge der Permutation des Startwerts.
func paare(p Profil, startwert uint64) [][2]int {
	perm := permutation(startwert, stromPaare, p.Sets*p.Entities)
	out := make([][2]int, len(perm))
	for i, k := range perm {
		out[i] = [2]int{k/p.Entities + 1, k%p.Entities + 1}
	}
	return out
}
