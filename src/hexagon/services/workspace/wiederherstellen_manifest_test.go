package workspace_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// kanarienWert steht in Schlüsseln und Payloads der Manifeste dieser Datei; er darf in keinem Fehlertext stehen
// (SPEC-129 Nr. 20: der Ort ohne Wert).
const kanarienWert = "CANARY-7f3a"

// kanarienEintraege ist der State der Manifest-Tests: fünf Einträge, jeder Schlüssel und jeder Payload trägt den Wert.
func kanarienEintraege() []model.EntityState {
	es := fuenfEintraege("ws-1")
	for i := range es {
		es[i].EntityKey += "-" + kanarienWert
		es[i].Payload = fmt.Sprintf(`{"c":"%s","n":%d}`, kanarienWert, i)
	}
	return es
}

// fehltMarke steht für ein Glied, das fehlt.
type fehltMarke struct{}

// manifestMit liefert das Manifest der Kanarien-Einträge, in dem das Glied name des Kopfes (idx < 0) oder des Eintrags
// idx auf wert gesetzt ist; fehltMarke entfernt das Glied.
func manifestMit(idx int, name string, wert any) []byte {
	baum := snapshotManifestBaum("svc", kanarienEintraege())
	ziel := baum
	if idx >= 0 {
		ziel = baum["entities"].([]any)[idx].(obj)
	}
	if _, fehlt := wert.(fehltMarke); fehlt {
		delete(ziel, name)
	} else {
		ziel[name] = wert
	}
	return []byte(schreibe(baum))
}

func manifestGut() []byte { return []byte(schreibe(snapshotManifestBaum("svc", kanarienEintraege()))) }

// mfall ist ein Fall der Manifest-Tests: die Regel aus SPEC-129 Nr. 20, der er angehört, und die Bytes.
type mfall struct {
	name, regel string
	b           []byte
}

// restoreMitManifest legt b als Manifest des Snapshots snp-1 ab und ruft Restore.
func restoreMitManifest(t *testing.T, b []byte) (*welt, error) {
	t.Helper()
	w := neu(t)
	w.vorbelegen("ws-1", 5)
	w.snapshotBytesAblegen("ws-1", "snp-1", b)
	w.p.leeren()
	_, err := restoreLauf(w)
	return w, err
}

// pruefeVerstoss: ODR-E3001 (500) mit ErrBeschaedigt, ohne den Wert des Kanarienvogels, ohne Transaktion, State unverändert.
func pruefeVerstoss(t *testing.T, f mfall) {
	t.Helper()
	w := neu(t)
	w.vorbelegen("ws-1", 5)
	w.snapshotBytesAblegen("ws-1", "snp-1", f.b)
	vorher := w.st.stand()
	w.p.leeren()
	_, err := restoreLauf(w)
	pruefeCode(t, err, model.CodeE3001, outbound.ErrBeschaedigt)
	if fe, ok := codeVon(err); ok {
		if fe.Status != 500 || fe.Ursache == nil {
			t.Errorf("%s: Status %d, Ursache %v", f.name, fe.Status, fe.Ursache)
		} else if text := err.Error() + " " + fe.Ursache.Error(); strings.Contains(text, kanarienWert) {
			t.Errorf("%s: der Wert steht im Fehlertext: %s", f.name, text)
		}
	}
	if w.p.anzahl(lblTx) != 0 || !zustaendeGleich(vorher, w.st.stand()) {
		t.Errorf("%s: Transaktion %d oder Zustand geändert", f.name, w.p.anzahl(lblTx))
	}
}

// syntaxFaelle: M6 a, Zeichen vor und hinter dem Manifest: gültig sind Leerzeichen, Tab, LF und CR (vier Zeichen, vorn
// und hinten = 8); Verstöße sind FF, BOM, NBSP und NUL (vorn und hinten = 8) und acht Formen des Dokuments (16). Die
// Bytes entstehen aus Bytewerten.
func syntaxFaelle() (gueltig [][]byte, verstoesse []mfall) {
	gut := manifestGut()
	for _, z := range [][]byte{{0x20}, {0x09}, {0x0A}, {0x0D}} {
		gueltig = append(gueltig, append(append([]byte{}, z...), gut...), append(append([]byte{}, gut...), z...))
	}
	for i, z := range [][]byte{{0x0C}, {0xEF, 0xBB, 0xBF}, {0xC2, 0xA0}, {0x00}} {
		n := []string{"FF", "BOM", "NBSP", "NUL"}[i]
		verstoesse = append(verstoesse, mfall{n + " vorn", "a", append(append([]byte{}, z...), gut...)},
			mfall{n + " hinten", "a", append(append([]byte{}, gut...), z...)})
	}
	verstoesse = append(verstoesse,
		mfall{"leer", "a", []byte{}},
		mfall{"nicht geschlossen", "a", gut[:len(gut)-1]},
		mfall{"zwei Werte", "a", append(append(append([]byte{}, gut...), ' '), gut...)},
		// Das Byte 0xFF steht in einem Glied, das der Leser sonst nicht liest: nur die Prüfung der Syntax fängt es.
		mfall{"Byte 0xFF", "a", bytes.Replace(gut, []byte(`"workspace"}`), append([]byte(`"workspace","zzz":"`), 0xFF, '"', '}'), 1)},
		mfall{"Array", "a", []byte{'[', ']'}},
		mfall{"null", "a", []byte{'n', 'u', 'l', 'l'}},
		mfall{"Zahl", "a", []byte{'1'}},
		mfall{"Text hinter dem Objekt", "a", append(append([]byte{}, gut...), 'x')},
	)
	return gueltig, verstoesse
}

// TestManifestSyntax (T41): M6 a.
func TestManifestSyntax(t *testing.T) {
	gueltig, verstoesse := syntaxFaelle()
	if len(gueltig) != 8 || len(verstoesse) != 16 {
		t.Fatalf("Zählprüfung: gültig %d, Verstoß %d", len(gueltig), len(verstoesse))
	}
	for i, b := range gueltig {
		if w, err := restoreMitManifest(t, b); err != nil || len(w.st.stand().ent["ws-1"]) != 5 {
			t.Errorf("gültig %d: %v", i, err)
		}
	}
	for _, f := range verstoesse {
		pruefeVerstoss(t, f)
	}
}

// kopfFaelle: M6 b: Glied 5 × Form 3 (fehlt, falscher Typ, falscher Wert) = 15, dazu die Schreibweisen der Ganzzahl
// von manifest_version und entity_count (je 5).
func kopfFaelle() []mfall {
	var r []mfall
	typ := map[string]any{"manifest_version": roh("true"), "snapshot_kind": 5, "service_id": 5, "metadata_hash": 5, "entity_count": "5"}
	wert := map[string]any{"manifest_version": 2, "snapshot_kind": "baseline", "service_id": "anders", "metadata_hash": "sha256:" + strings.Repeat("f", 64), "entity_count": 4}
	for _, g := range []string{"manifest_version", "snapshot_kind", "service_id", "metadata_hash", "entity_count"} {
		r = append(r, mfall{g + " fehlt", "b:" + g, manifestMit(-1, g, fehltMarke{})},
			mfall{g + " Typ", "b:" + g, manifestMit(-1, g, typ[g])},
			mfall{g + " Wert", "b:" + g, manifestMit(-1, g, wert[g])})
	}
	for _, s := range []string{"01", "1.0", "1e0", "-1", `"1"`} {
		r = append(r, mfall{"manifest_version " + s, "b:manifest_version", manifestMit(-1, "manifest_version", roh(s))})
	}
	for _, s := range []string{"05", "5.0", "5e0", "-5", `"5"`} {
		r = append(r, mfall{"entity_count " + s, "b:entity_count", manifestMit(-1, "entity_count", roh(s))})
	}
	return r
}

// TestManifestKopf (T42): M6 b.
func TestManifestKopf(t *testing.T) {
	fs := kopfFaelle()
	if len(fs) != 25 {
		t.Fatalf("Zählprüfung: %d Fälle", len(fs))
	}
	for _, f := range fs {
		pruefeVerstoss(t, f)
	}
}

// eintragFaelle: M6 c und e: die sieben Glieder eines Eintrags und entities, je mit fehlt, Typ, leer, Wert; dazu die
// Invarianten von EntityState (e).
func eintragFaelle() []mfall {
	var r []mfall
	e := func(g, name string, wert any) {
		r = append(r, mfall{g + " " + name, "c:" + g, manifestMit(2, g, wert)})
	}
	f := fehltMarke{}
	e("completeness", "fehlt", f)
	e("completeness", "Typ", 5)
	e("completeness", "leer", "")
	e("completeness", "Wert", "full")
	e("deleted", "fehlt", f)
	e("deleted", "Typ", "true")
	e("deleted", "null", roh("null"))
	e("deleted", "Zahl", roh("1"))
	e("entity_set", "fehlt", f)
	e("entity_set", "Typ", 5)
	e("entity_set", "leer", "")
	e("payload", "fehlt", f)
	e("selected_properties", "fehlt", f)
	e("selected_properties", "Typ", "Id")
	e("selected_properties", "null", roh("null"))
	e("selected_properties", "leeres Element", []any{""})
	e("selected_properties", "doppelt", []any{"Id", "Id"})
	e("selected_properties", "kein String", []any{roh("1")})
	e("typed_key", "fehlt", f)
	e("typed_key", "Typ", 5)
	e("typed_key", "leer", "")
	e("version", "fehlt", f)
	e("version", "Typ", "1")
	e("version", "Bruch", roh("1.5"))
	e("version", "Exponent", roh("1e2"))
	e("version", "über int64", roh("9223372036854775808"))
	e("version", "null", roh("null"))
	// (e) Validate(): eine negative Version ist ein Ganzzahl im Bereich, aber kein gültiger Eintrag.
	r = append(r, mfall{"version negativ", "e", manifestMit(2, "version", roh("-1"))})
	for _, n := range []struct {
		name string
		w    any
	}{{"fehlt", fehltMarke{}}, {"Typ", obj{}}, {"null", roh("null")}} {
		r = append(r, mfall{"entities " + n.name, "c:entities", manifestMit(-1, "entities", n.w)})
	}
	r = append(r, mfall{"Element kein Objekt", "c:entities", manifestMit(-1, "entities", []any{roh("5"), roh("5"), roh("5"), roh("5"), roh("5")})})
	return r
}

// TestManifestEintrag (T43): M6 c, e, dazu die Gutfälle `payload: null`, leere Properties und die größte Version.
func TestManifestEintrag(t *testing.T) {
	fs := eintragFaelle()
	if len(fs) != 32 {
		t.Fatalf("Zählprüfung: %d Fälle", len(fs))
	}
	for _, f := range fs {
		pruefeVerstoss(t, f)
	}
	gut := map[string][]byte{
		"payload null":           manifestMit(2, "payload", roh("null")),
		"Properties leer":        manifestMit(2, "selected_properties", []any{}),
		"größte Version":         manifestMit(2, "version", roh("9223372036854775807")),
		"Version null":           manifestMit(2, "version", roh("0")),
		"deleted true":           manifestMit(2, "deleted", true),
		"Menge und Schlüssel ok": manifestMit(2, "entity_set", "Zwei"),
	}
	for name, b := range gut {
		w, err := restoreMitManifest(t, b)
		if err != nil || len(w.st.stand().ent["ws-1"]) != 5 {
			t.Errorf("Gutfall %s: %v", name, err)
		}
	}
	w, err := restoreMitManifest(t, gut["payload null"])
	if err != nil || w.st.stand().ent["ws-1"][2].Payload != "null" {
		t.Errorf("payload null: %v", err)
	}
}

// doppeltFaelle: M6 d: das Paar (entity_set, typed_key) zweimal an den Positionen (erste zwei, erster und letzter,
// mittlerer und letzter) bei fünf Einträgen.
func doppeltFaelle() []mfall {
	var r []mfall
	for _, p := range [][2]int{{0, 1}, {0, 4}, {2, 4}} {
		baum := snapshotManifestBaum("svc", kanarienEintraege())
		es := baum["entities"].([]any)
		es[p[1]].(obj)["entity_set"] = es[p[0]].(obj)["entity_set"]
		es[p[1]].(obj)["typed_key"] = es[p[0]].(obj)["typed_key"]
		r = append(r, mfall{fmt.Sprintf("doppelt %d und %d", p[0], p[1]), "d", []byte(schreibe(baum))})
	}
	return r
}

// TestManifestDoppelt (T44): M6 d; derselbe Schlüssel in zwei Mengen und in anderer Schreibweise gelten.
func TestManifestDoppelt(t *testing.T) {
	fs := doppeltFaelle()
	if len(fs) != 3 {
		t.Fatalf("Zählprüfung: %d Fälle", len(fs))
	}
	for _, f := range fs {
		pruefeVerstoss(t, f)
	}
	baum := snapshotManifestBaum("svc", kanarienEintraege())
	es := baum["entities"].([]any)
	es[1].(obj)["typed_key"] = es[0].(obj)["typed_key"] // zwei Mengen, derselbe Schlüssel
	es[4].(obj)["typed_key"] = strings.ToLower(es[2].(obj)["typed_key"].(string))
	es[4].(obj)["entity_set"] = es[2].(obj)["entity_set"] // dieselbe Menge, andere Schreibweise
	if w, err := restoreMitManifest(t, []byte(schreibe(baum))); err != nil || len(w.st.stand().ent["ws-1"]) != 5 {
		t.Errorf("Gutfälle: %v", err)
	}
}

// TestManifestVerstoss (T45): jeder Verstoß ist ODR-E3001 (500) mit ErrBeschaedigt, der Ort ohne Wert, nichts geändert,
// keine Transaktion; die Schleife läuft über alle Fälle von T41 bis T44, die Zählprüfung gleicht die Regeln mit der Liste
// 16 ab.
func TestManifestVerstoss(t *testing.T) {
	_, syntax := syntaxFaelle()
	alle := append(append(append(syntax, kopfFaelle()...), eintragFaelle()...), doppeltFaelle()...)
	if len(alle) != 16+25+32+3 {
		t.Fatalf("Zählprüfung: %d Fälle", len(alle))
	}
	regeln := map[string]bool{}
	for _, f := range alle {
		regeln[f.regel] = true
		pruefeVerstoss(t, f)
	}
	if len(regeln) != 16 {
		t.Errorf("Regeln %d, erwartet 16 (a, b×5, c×8, d, e): %v", len(regeln), regeln)
	}
}
