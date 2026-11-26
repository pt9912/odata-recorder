package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
)

// maxZeile ist die längste Zeile in Bytes ohne Zeilenende, die der Leser auswertet (SPEC-216 Nr. 10 a).
const maxZeile = 1 << 20

// spanZeile ist eine erkannte Zeile eines Spans (SPEC-216 Nr. 10 b); Mikros ist die Dauer in ganzen Mikrosekunden, Modus der Wert von
// recorder.mode (leer, wenn die Zeile ihn nicht trägt) und ParentID leer ohne parent_id.
type spanZeile struct {
	Name     string
	SpanID   string
	ParentID string
	Modus    string
	Mikros   int64
}

// mikrosekunden liest die Dauer `<Ziffern>.<drei Ziffern>` als ganze Mikrosekunden aus dem Text, nie als Gleitkommazahl.
func mikrosekunden(roh []byte) (int64, bool) {
	punkt := bytes.IndexByte(roh, '.')
	if punkt < 0 || len(roh)-punkt-1 != 3 {
		return 0, false
	}
	for i, c := range roh {
		if i != punkt && (c < '0' || c > '9') {
			return 0, false
		}
	}
	n, err := strconv.ParseInt(string(roh[:punkt])+string(roh[punkt+1:]), 10, 64)
	return n, err == nil
}

// textFeld liest das Feld key als JSON-Text; ok sagt, ob es besteht und ein Text ist (null ist keiner).
func textFeld(obj map[string]json.RawMessage, key string) (string, bool) {
	roh, ok := obj[key]
	if !ok || string(roh) == "null" {
		return "", false
	}
	var s string
	return s, json.Unmarshal(roh, &s) == nil
}

// leseSpan erkennt eine Span-Zeile: ein JSON-Objekt mit ereignis gleich span, msg, span_id und duration_ms; parent_id ist optional, ein Feld
// falscher Art macht die Zeile fremd. Grenze: vor dem { übergeht die Funktion allein Leerzeichen und Tab; ein CR davor nennt SPEC-216 nicht, und
// kein Test führt es.
func leseSpan(zeile []byte) (spanZeile, bool) {
	var obj map[string]json.RawMessage
	if kopf := bytes.TrimLeft(zeile, " \t"); len(kopf) == 0 || kopf[0] != '{' || json.Unmarshal(kopf, &obj) != nil {
		return spanZeile{}, false
	}
	ereignis, _ := textFeld(obj, "ereignis")
	name, ok2 := textFeld(obj, "msg")
	id, ok3 := textFeld(obj, "span_id")
	dauer, ok4 := mikrosekunden(obj["duration_ms"])
	if !ok2 || !ok3 || !ok4 || ereignis != "span" {
		return spanZeile{}, false
	}
	z := spanZeile{Name: name, SpanID: id, Mikros: dauer}
	if _, da := obj["parent_id"]; da {
		var ok bool
		if z.ParentID, ok = textFeld(obj, "parent_id"); !ok {
			return spanZeile{}, false
		}
	}
	z.Modus, _ = textFeld(obj, "recorder.mode")
	return z, true
}

// lies entleert r bis zum Ende ohne Unterbrechung und ruft für jede Span-Zeile verarbeite; jede andere Zeile zählt in fremd. Eine Zeile
// über maxZeile wird bis zum Zeilenende verworfen (der Puffer ist dann leer) und zählt fremd; der Speicher wächst nicht mit ihrer Länge. Eine
// letzte Zeile ohne Zeilenende ist eine Zeile. Ein anderer Fehler als das Ende des Stroms beendet das Lesen.
func lies(r io.Reader, verarbeite func(spanZeile)) (fremd int, err error) {
	br := bufio.NewReaderSize(r, 64<<10)
	var z zeilenPuffer
	for {
		teil, rerr := br.ReadSlice('\n')
		ende := rerr == nil
		if ende {
			teil = teil[:len(teil)-1]
		}
		z.haenge(teil)
		if errors.Is(rerr, bufio.ErrBufferFull) {
			continue
		}
		if ende || !z.leer() {
			if s, ok := leseSpan(z.puffer); ok {
				verarbeite(s)
			} else {
				fremd++
			}
			z.zuLang, z.puffer = false, z.puffer[:0]
		}
		if rerr != nil {
			if errors.Is(rerr, io.EOF) {
				return fremd, nil
			}
			return fremd, rerr
		}
	}
}

// zeilenPuffer sammelt die Stücke einer Zeile bis maxZeile; eine längere Zeile gibt den Puffer frei und merkt sich nur, dass sie zu lang war.
type zeilenPuffer struct {
	puffer []byte
	zuLang bool
}

func (z *zeilenPuffer) haenge(teil []byte) {
	switch {
	case z.zuLang:
	case len(z.puffer)+len(teil) > maxZeile:
		z.zuLang, z.puffer = true, z.puffer[:0]
	default:
		z.puffer = append(z.puffer, teil...)
	}
}

// leer sagt, dass seit dem letzten Zeilenende nichts gelesen wurde.
func (z *zeilenPuffer) leer() bool { return len(z.puffer) == 0 && !z.zuLang }
