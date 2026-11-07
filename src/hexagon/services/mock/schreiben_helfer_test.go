package mock_test

import (
	"io"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// Die Orakel der Mutationen nennen Statuscodes, Codes, Header und Texte als Literale aus dem Vertrag, nie als Konstanten des
// Pakets (SPEC-192 bis SPEC-194).
const (
	e6002 = model.CodeE6002
	e6003 = model.CodeE6003
	e6005 = model.CodeE6005
	e6006 = model.CodeE6006
	e6007 = model.CodeE6007
	e6008 = model.CodeE6008
	e6009 = model.CodeE6009
	e7001 = model.CodeE7001
	e9001 = model.CodeE9001
)

// alleProperties sind die neun Namen von Product in der Folge der Bytes (Big, Count, Day, Flag, Id, Name, Price, Ref, Ts).
func alleProperties() []string {
	r := alleNamen()
	slices.Sort(r)
	return r
}

// voll ist ein vollständiger Eintrag von Products in der Kanonform: Id und Name gesetzt, jede andere Property `null`, alle neun
// bekannt, Version 1.
func voll(id int, name string) model.EntityState {
	return model.EntityState{
		WorkspaceID: wsID, EntitySet: setProducts, EntityKey: strconv.Itoa(id), Completeness: model.EntityComplete, Version: 1,
		Payload:         `{"Big":null,"Count":null,"Day":null,"Flag":null,"Id":` + strconv.Itoa(id) + `,"Name":"` + name + `","Price":null,"Ref":null,"Ts":null}`,
		KnownProperties: alleProperties(),
	}
}

// schreibUmg ist ein Aufbau mit dem Testmodell, Coverage complete für Products und den drei vollen Einträgen 1 bis 3.
func schreibUmg(t *testing.T) *umg {
	t.Helper()
	u := neueUmg(t)
	u.c.mc.modell = testModell()
	u.co.setze(baselineID, setProducts, model.CoverageComplete)
	u.en.setze(setProducts, voll(1, "a"), voll(2, "b"), voll(3, "c"))
	return u
}

// pruefeFehler prüft Code, Status und Text eines Fehlers des Dienstes und wirft das Ergebnis weg.
func pruefeFehler(t *testing.T, kontext string, err error, code model.Code) {
	t.Helper()
	_ = erwarteFehler(t, kontext, err, code)
}

// pw ist ein Eintrag von EntityEingabe.Properties.
func pw(name string, art model.WertArt, text string) model.PropertyWert {
	return model.PropertyWert{Name: name, Wert: model.Wert{Art: art, Text: text}}
}

// eingabeIDName ist der Body mit Id und Name; der Schlüssel ist die Id.
func eingabeIDName(id, name string) model.EntityEingabe {
	return model.EntityEingabe{Typ: typProduct, Schluessel: id, Properties: []model.PropertyWert{pw("Id", model.ArtInt32, id), pw("Name", model.ArtString, name)}}
}

// setzeMutation setzt die Ressource des URL-Lesers auf die der Methode (POST an die Collection, sonst an die Entity mit dem
// Schlüssel id) und den Body, den der Fake des Lesers liefert.
func (u *umg) setzeMutation(methode, id string, e model.EntityEingabe) {
	if methode == "POST" {
		u.menge(model.Optionen{})
	} else {
		u.entity(model.Optionen{}, id)
	}
	u.ei.ergebnis = e
}

// mutiere ruft die Methode mit dem Body e; der Request trägt den Body "{}" und einen Content-Type.
func (u *umg) mutiere(methode, id string, e model.EntityEingabe) (inbound.Antwort, error) {
	u.t.Helper()
	u.setzeMutation(methode, id, e)
	a := anfrage(methode, dienstRoot+"/x")
	a.Body, a.Laenge = io.NopCloser(strings.NewReader("{}")), 2
	a.Header["Content-Type"] = []string{"application/json"}
	return u.rufe(a)
}

// zustand ist der gespeicherte Entity-State von Products als Karte Schlüssel → Eintrag.
func (u *umg) zustand() map[string]model.EntityState {
	r := map[string]model.EntityState{}
	for _, x := range u.en.inhalt()[setProducts] {
		r[x.EntityKey] = x
	}
	return r
}

// unveraendert prüft, dass der Entity-State und die Revision die des Aufbaus sind und kein schreibender Port lief (W61).
func (u *umg) unveraendert(vorher map[string][]model.EntityState, rev int64, kontext string) {
	u.t.Helper()
	nachher := u.en.inhalt()
	if len(nachher) != len(vorher) {
		u.t.Errorf("%s: Sets %d, vorher %d", kontext, len(nachher), len(vorher))
	}
	for set, liste := range vorher {
		if !slices.EqualFunc(liste, nachher[set], func(a, b model.EntityState) bool {
			return a.EntityKey == b.EntityKey && a.Payload == b.Payload && a.Version == b.Version && a.Deleted == b.Deleted &&
				a.Completeness == b.Completeness && slices.Equal(a.KnownProperties, b.KnownProperties)
		}) {
			u.t.Errorf("%s: Set %s verändert", kontext, set)
		}
	}
	if got := u.ws.revision(wsID); got != rev {
		u.t.Errorf("%s: Revision %d, vorher %d", kontext, got, rev)
	}
}

// ohneSchreiben prüft, dass Put und Advance nie liefen.
func (u *umg) ohneSchreiben(kontext string) {
	u.t.Helper()
	if u.p.anzahl("Entities.Put") != 0 || u.p.anzahl("Workspaces.Advance") != 0 {
		u.t.Errorf("%s: Put %d Advance %d, erwartet 0", kontext, u.p.anzahl("Entities.Put"), u.p.anzahl("Workspaces.Advance"))
	}
}
