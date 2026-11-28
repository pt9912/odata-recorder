package batchcodec

import (
	"bytes"
	"strconv"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// maxTeile ist die Zahl der Teile der Hülle und aller Changesets zusammen (SPEC-183 Nr. 7).
const maxTeile = 1000

// zerleger zählt die Teile eines Batch über beide Ebenen.
type zerleger struct{ teile int }

// zerlege liest die Hülle; jeder Verstoß verwirft das ganze Ergebnis.
func zerlege(daten []byte, grenze string) (model.Batch, error) {
	z := &zerleger{}
	inhalte, err := z.ebene(daten, grenze, "")
	if err != nil {
		return model.Batch{}, err
	}
	teile := make([]model.BatchTeil, 0, len(inhalte))
	for i, inhalt := range inhalte {
		t, err := z.teil(inhalt, strconv.Itoa(i+1))
		if err != nil {
			return model.Batch{}, err
		}
		teile = append(teile, t)
	}
	return model.Batch{Teile: teile}, nil
}

// ebene liefert die Inhalte einer Ebene: zuerst der Rahmen, dann die Zahl der Teile (SPEC-182 Nr. 5 f).
func (z *zerleger) ebene(inhalt []byte, grenze, ort string) ([][]byte, error) {
	inhalte, anzahl, geschlossen := rahmen(inhalt, grenze, maxTeile)
	if !geschlossen {
		return nil, fehler(model.BatchRahmen, ort)
	}
	z.teile += anzahl
	if z.teile > maxTeile {
		return nil, fehler(model.BatchZuViele, ort)
	}
	return inhalte, nil
}

// teil liest einen Teil der Hülle: ein Einzelrequest oder ein Changeset.
func (z *zerleger) teil(inhalt []byte, ort string) (model.BatchTeil, error) {
	k, rest, changeset, err := teilArt(inhalt, ort, false)
	if err != nil {
		return model.BatchTeil{}, err
	}
	if changeset {
		return z.changeset(k.typ.wert, rest, ort)
	}
	r, err := request(k, rest, ort)
	if err != nil {
		return model.BatchTeil{}, err
	}
	return model.BatchTeil{Requests: []model.BatchRequest{r}}, nil
}

// changeset liest den Rest eines Teils als Ebene mit der Grenzzeichenfolge seines Content-Type (SPEC-183 Nr. 5).
func (z *zerleger) changeset(contentType string, inhalt []byte, ort string) (model.BatchTeil, error) {
	_, parameter := medientyp(contentType)
	grenze, ok := grenzzeichen(parameter)
	if !ok {
		return model.BatchTeil{}, fehler(model.BatchBoundary, ort)
	}
	inhalte, err := z.ebene(inhalt, grenze, ort)
	if err != nil {
		return model.BatchTeil{}, err
	}
	requests := make([]model.BatchRequest, 0, len(inhalte))
	for i, in := range inhalte {
		o := ort + "/" + strconv.Itoa(i+1)
		k, rest, _, err := teilArt(in, o, true)
		if err != nil {
			return model.BatchTeil{}, err
		}
		r, err := request(k, rest, o)
		if err != nil {
			return model.BatchTeil{}, err
		}
		requests = append(requests, r)
	}
	return model.BatchTeil{Changeset: true, Requests: requests}, nil
}

// feld ist eine Kopfzeile eines Teils, die höchstens einmal steht.
type feld struct {
	wert string
	da   bool
}

func (f *feld) setze(wert []byte) bool {
	if f.da {
		return false
	}
	f.wert, f.da = string(wert), true
	return true
}

// teilKopf sind die Kopfzeilen eines Teils mit Wirkung (SPEC-183 Nr. 4).
type teilKopf struct{ typ, kodierung, contentID feld }

// setze nimmt eine Kopfzeile des Teils auf; falsch ist die Wiederholung einer Kopfzeile mit Wirkung.
func (k *teilKopf) setze(name, wert []byte) bool {
	switch n := string(name); {
	case gleichASCII(n, "content-type"):
		return k.typ.setze(wert)
	case gleichASCII(n, "content-transfer-encoding"):
		return k.kodierung.setze(wert)
	case gleichASCII(n, "content-id"):
		return k.contentID.setze(wert)
	}
	return true
}

// teilArt liest den Block eines Teils und bestimmt seine Art (SPEC-183 Nr. 4 a und b): Content-Type, dann
// Content-Transfer-Encoding. changeset ist nur in der Hülle wahr.
func teilArt(inhalt []byte, ort string, imChangeset bool) (k teilKopf, rest []byte, changeset bool, err error) {
	rest, ursache := block(inhalt, k.setze)
	if ursache != "" {
		return k, nil, false, fehler(ursache, ort)
	}
	typ, _ := medientyp(k.typ.wert)
	switch {
	case k.typ.wert == "":
		return k, nil, false, fehler(model.BatchContentType, ort)
	case gleichASCII(typ, typMultipart) && imChangeset:
		return k, nil, false, fehler(model.BatchVerschachtelt, ort)
	case gleichASCII(typ, typMultipart):
		changeset = true
	case !gleichASCII(typ, typHTTP):
		return k, nil, false, fehler(model.BatchTeilTyp, ort)
	}
	if k.kodierung.da && !kodierungGueltig(k.kodierung.wert) {
		return k, nil, false, fehler(model.BatchTransferKodierung, ort)
	}
	return k, rest, changeset, nil
}

// kodierungGueltig ist wahr für die Transfer-Kodierungen, die den Inhalt unverändert lassen.
func kodierungGueltig(wert string) bool {
	return gleichASCII(wert, "binary") || gleichASCII(wert, "7bit") || gleichASCII(wert, "8bit")
}

// request liest den Einzelrequest eines Teils (SPEC-183 Nr. 4 c und 6): Content-ID, Anfragezeile, Kopfzeilen,
// Transfer-Encoding, Body. Body ist eine Kopie, nil ohne Byte.
func request(k teilKopf, inhalt []byte, ort string) (model.BatchRequest, error) {
	if k.contentID.da && !model.ContentIDGueltig(k.contentID.wert) {
		return model.BatchRequest{}, fehler(model.BatchContentID, ort)
	}
	erste, danach, _ := bytes.Cut(inhalt, []byte(zeilenende))
	methode, ziel, ok := anfragezeile(erste)
	if !ok {
		return model.BatchRequest{}, fehler(model.BatchAnfragezeile, ort)
	}
	r := model.BatchRequest{ContentID: k.contentID.wert, Methode: string(methode), Ziel: string(ziel), Header: []model.BatchKopf{}}
	transfer := false
	body, ursache := block(danach, func(name, wert []byte) bool {
		kopf := model.BatchKopf{Name: string(name), Wert: string(wert)}
		transfer = transfer || gleichASCII(kopf.Name, "transfer-encoding")
		r.Header = append(r.Header, kopf)
		return true
	})
	switch {
	case ursache != "":
		return model.BatchRequest{}, fehler(ursache, ort)
	case transfer:
		return model.BatchRequest{}, fehler(model.BatchTransferKodierung, ort)
	case len(body) > 0:
		r.Body = bytes.Clone(body)
	}
	return r, nil
}

// anfragezeile zerlegt `Methode SP Ziel SP HTTP/1.1` mit genau zwei SP: die Methode ein Token, das Ziel eine nichtleere Folge aus
// 0x21 bis 0x7E und 0x80 bis 0xFF.
func anfragezeile(zeile []byte) (methode, ziel []byte, ok bool) {
	methode, rest, _ := bytes.Cut(zeile, []byte(" "))
	ziel, version, _ := bytes.Cut(rest, []byte(" "))
	if !istToken(methode) || len(ziel) == 0 || string(version) != "HTTP/1.1" {
		return nil, nil, false
	}
	for _, c := range ziel {
		if c < 0x21 || c == 0x7F {
			return nil, nil, false
		}
	}
	return methode, ziel, true
}
