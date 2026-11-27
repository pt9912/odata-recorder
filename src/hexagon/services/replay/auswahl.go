package replay

import (
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// waehle wählt aus liste den Exchange für key nach der Strategie s (SPEC-135 Nr. 9); Kandidaten sind allein die Exchanges
// mit dem Key, eine Ordnung der Liste setzt die Auswahl nicht voraus. hoechste ist die größte SequenceNo der Session,
// verbraucht sind die SequenceNo in consumed, position ist die Position des Cursors. Der Fehler trägt den Code; ohne Fehler
// ist der Exchange gewählt.
func waehle(s model.ReplayStrategy, key string, liste []model.Exchange, hoechste int, redigierteBodies bool, consumed []int64, position int64) (model.Exchange, error) {
	var treffer []model.Exchange
	for _, e := range liste {
		if gleicherKey(e, key) {
			treffer = append(treffer, e)
		}
	}
	if len(treffer) == 0 {
		return model.Exchange{}, mitCode(model.CodeE5001, nil)
	}
	switch s {
	case model.StrategySequential:
		return waehleSequential(treffer, redigierteBodies, consumed)
	case model.StrategyStrict:
		return waehleStrict(treffer, hoechste, redigierteBodies, position)
	}
	return waehleErster(s == model.StrategyLast, treffer, redigierteBodies)
}

// gleicherKey sagt, ob der Key des Exchange byte-gleich key ist; ein leerer MatchKey passt nie.
func gleicherKey(e model.Exchange, key string) bool {
	return e.MatchKey != "" && e.MatchKey == key
}

func verbraucht(consumed []int64, seq int) bool {
	for _, c := range consumed {
		if c == int64(seq) {
			return true
		}
	}
	return false
}

// waehleSequential wählt unter den unverbrauchten replayfähigen Kandidaten den mit der kleinsten SequenceNo.
func waehleSequential(treffer []model.Exchange, redigierteBodies bool, consumed []int64) (model.Exchange, error) {
	var (
		gewaehlt     model.Exchange
		gefunden     bool
		unverbraucht bool
	)
	for _, e := range treffer {
		if verbraucht(consumed, e.SequenceNo) {
			continue
		}
		unverbraucht = true
		if model.Replayfaehig(e, redigierteBodies) && (!gefunden || e.SequenceNo < gewaehlt.SequenceNo) {
			gewaehlt, gefunden = e, true
		}
	}
	switch {
	case gefunden:
		return gewaehlt, nil
	case unverbraucht:
		return model.Exchange{}, mitCode(model.CodeE5008, nil)
	}
	return model.Exchange{}, mitCode(model.CodeE5002, nil)
}

// waehleStrict wählt allein den Exchange der Session mit SequenceNo gleich position+1. Über hoechste ist die Session
// erschöpft; sonst gibt es die Nummer, und ist sie kein Kandidat in treffer, trägt sie einen anderen Key.
func waehleStrict(treffer []model.Exchange, hoechste int, redigierteBodies bool, position int64) (model.Exchange, error) {
	if position+1 > int64(hoechste) {
		return model.Exchange{}, mitCode(model.CodeE5002, nil)
	}
	for _, e := range treffer {
		if int64(e.SequenceNo) != position+1 {
			continue
		}
		if !model.Replayfaehig(e, redigierteBodies) {
			return model.Exchange{}, mitCode(model.CodeE5008, nil)
		}
		return e, nil
	}
	return model.Exchange{}, mitCode(model.CodeE5007, nil)
}

// waehleErster wählt unter den replayfähigen Kandidaten den mit der kleinsten, bei letzter die größten SequenceNo.
func waehleErster(letzter bool, treffer []model.Exchange, redigierteBodies bool) (model.Exchange, error) {
	var (
		gewaehlt model.Exchange
		gefunden bool
	)
	for _, e := range treffer {
		if !model.Replayfaehig(e, redigierteBodies) {
			continue
		}
		if !gefunden || letzter && e.SequenceNo > gewaehlt.SequenceNo || !letzter && e.SequenceNo < gewaehlt.SequenceNo {
			gewaehlt, gefunden = e, true
		}
	}
	if !gefunden {
		return model.Exchange{}, mitCode(model.CodeE5008, nil)
	}
	return gewaehlt, nil
}
