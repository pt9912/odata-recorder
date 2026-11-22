package admin_test

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// inspMitglied ist ein Feld von inbound.ExchangeZeile (Name im Code) mit dem Schlüssel im JSON (SPEC-293 Nr. 6): die Liste
// steht in der Folge des Vertrags, nicht in der des Typs.
type inspMitglied struct{ feld, schluessel string }

func inspMitglieder() []inspMitglied {
	return []inspMitglied{
		{"ID", "id"}, {"SequenceNo", "sequence_no"}, {"Method", "method"}, {"Path", "path"}, {"StartedAt", "started_at"},
		{"FinishedAt", "finished_at"}, {"KeyRedacted", "key_redacted"}, {"RequestBody", "request_body_state"},
		{"ResponseBody", "response_body_state"}, {"ResponseStatus", "response_status"}, {"Capture", "capture_status"},
		{"Integrity", "body_integrity"}, {"Equivalence", "replay_equivalence"}, {"RequestRedact", "request_redaction_status"},
		{"ResponseRedact", "response_redaction_status"},
	}
}

// inspLit ist ein Wert eines Feldes mit seinem JSON als Literal.
type inspLit struct {
	wert any
	json string
}

// inspWert liefert den Wert des Feldes für die Zeile n (1 bis 9, jede Zeile hat in jedem Feld einen eigenen Wert) und sein JSON
// als Literal des Vertrags.
func inspWert(feld string, n int) inspLit {
	s := strconv.Itoa(n)
	praefix := map[string]string{
		"ID": "ex-", "Method": "M", "Path": "/p/", "RequestBody": "rb", "ResponseBody": "sb", "Capture": "cs", "Integrity": "bi",
		"Equivalence": "re", "RequestRedact": "rr", "ResponseRedact": "sr",
	}
	switch feld {
	case "SequenceNo":
		return inspLit{100 + n, strconv.Itoa(100 + n)}
	case "ResponseStatus":
		return inspLit{200 + n, strconv.Itoa(200 + n)}
	case "KeyRedacted":
		return inspLit{n%2 == 1, strconv.FormatBool(n%2 == 1)}
	case "StartedAt":
		return inspLit{time.Date(2026, 11, 6, 10, n, 0, 100000000+n, time.UTC), `"2026-11-06T10:0` + s + `:00.10000000` + s + `Z"`}
	case "FinishedAt":
		return inspLit{time.Date(2026, 11, 7, 11, n, 5, 200000000+n, time.UTC), `"2026-11-07T11:0` + s + `:05.20000000` + s + `Z"`}
	}
	return inspLit{praefix[feld] + s, `"` + praefix[feld] + s + `"`}
}

// inspZeile baut die Zeile des Dienstes und ihr JSON: jedes Feld aus der Zeile n, außer den Feldern in abw.
func inspZeile(n int, abw map[string]inspLit) (inbound.ExchangeZeile, string) {
	var z inbound.ExchangeZeile
	var teile []string
	for _, m := range inspMitglieder() {
		l, da := abw[m.feld]
		if !da {
			l = inspWert(m.feld, n)
		}
		f := reflect.ValueOf(&z).Elem().FieldByName(m.feld)
		if f.Kind() == reflect.String {
			f.SetString(reflect.ValueOf(l.wert).String())
		} else {
			f.Set(reflect.ValueOf(l.wert))
		}
		teile = append(teile, `"`+m.schluessel+`":`+l.json)
	}
	return z, "{" + strings.Join(teile, ",") + "}"
}

// inspSeiteJSON ist der Body der Seite aus den JSON-Zeilen als Literal des Vertrags.
func inspSeiteJSON(zeilen []string, limit, total int, next string) string {
	return `{"exchanges":[` + strings.Join(zeilen, ",") + `],"limit":` + strconv.Itoa(limit) + `,"total":` + strconv.Itoa(total) + `,"next":` + next + `}`
}

// inspHole sendet `GET <pfad>?<query>` an einen Server über einem Fake, den einstellen vorbereitet, und prüft die
// Erfolgsantwort (200, Header, Content-Length); der Body ist der Rückgabewert.
func inspHole(t *testing.T, name, pfad, query string, einstellen func(f *inspektionFake)) string {
	t.Helper()
	u, f, _ := inspUmgebung(t)
	einstellen(f)
	rec := u.rufe(inspReq(pfad, query))
	if rec.Code != 200 {
		t.Errorf("%s: Status %d, Body %q", name, rec.Code, rec.Body.String())
	}
	pruefeErfolgsHeader(t, name, rec)
	return rec.Body.String()
}

// TestI10FelderDerExchanges hält I10 (SPEC-293 Nr. 6): `exchanges` antwortet 200 mit den vier Feldern der Seite und den 15 Feldern der
// Zeile in der Folge des Vertrags. M-Felder: jedes der 15 Felder mit eigenem Wert (Zeile 1 mit je einem Feld aus Zeile 2: 15
// Fälle, ein vertauschtes oder fehlendes Feld ändert den Body) und alle Felder verschieden (Zeilen 1 bis 3); `next` 0 ist `null`,
// 7 die Zahl; 0 (nil und leer), 1 und 3 Zeilen. Orakel: Namen und Folge als Literale. Zählprüfung: NumField gleich 15 und 4.
func TestI10FelderDerExchanges(t *testing.T) {
	gleichAuf(t, "Felder von ExchangeZeile und ExchangeSeite", []int{reflect.TypeFor[inbound.ExchangeZeile]().NumField(), reflect.TypeFor[inbound.ExchangeSeite]().NumField()}, []int{15, 4})
	gleichAuf(t, "Zahl der Mitglieder", len(inspMitglieder()), 15)
	for _, m := range inspMitglieder() {
		if _, da := reflect.TypeFor[inbound.ExchangeZeile]().FieldByName(m.feld); !da {
			t.Errorf("ExchangeZeile hat kein Feld %s", m.feld)
		}
	}
	for _, m := range inspMitglieder() {
		z, js := inspZeile(1, map[string]inspLit{m.feld: inspWert(m.feld, 2)})
		got := inspHole(t, m.feld, exchangesZiel("s"), "", func(f *inspektionFake) {
			f.seite = inbound.ExchangeSeite{Eintraege: []inbound.ExchangeZeile{z}, Limit: 25, Gesamt: 99, Weiter: 7}
		})
		if soll := inspSeiteJSON([]string{js}, 25, 99, "7"); got != soll {
			t.Errorf("Feld %s: Body %q, erwartet %q", m.feld, got, soll)
		}
	}
	var zeilen []inbound.ExchangeZeile
	var js []string
	for n := 1; n <= 3; n++ {
		z, j := inspZeile(n, nil)
		zeilen, js = append(zeilen, z), append(js, j)
	}
	faelle := []struct {
		name   string
		zeilen []inbound.ExchangeZeile
		js     []string
		weiter int
		next   string
	}{
		{"nil", nil, nil, 0, "null"}, {"leer", []inbound.ExchangeZeile{}, nil, 0, "null"}, {"eine", zeilen[:1], js[:1], 7, "7"},
		{"drei, next 0", zeilen, js, 0, "null"}, {"drei, next 7", zeilen, js, 7, "7"}, {"drei, next größter int", zeilen, js, math.MaxInt, strconv.Itoa(math.MaxInt)},
		// Die Folge des Dienstes gilt: weder auf- noch absteigend in einem Feld der Zeilen (die Route sortiert nicht).
		{"Folge 2, 3, 1", []inbound.ExchangeZeile{zeilen[1], zeilen[2], zeilen[0]}, []string{js[1], js[2], js[0]}, 7, "7"},
		{"Folge 3, 1, 2", []inbound.ExchangeZeile{zeilen[2], zeilen[0], zeilen[1]}, []string{js[2], js[0], js[1]}, 7, "7"},
	}
	for _, f := range faelle {
		for _, lt := range [][2]int{{25, 99}, {0, 0}, {1, 3}} {
			got := inspHole(t, f.name, exchangesZiel("s"), "", func(fk *inspektionFake) {
				fk.seite = inbound.ExchangeSeite{Eintraege: f.zeilen, Limit: lt[0], Gesamt: lt[1], Weiter: f.weiter}
			})
			if soll := inspSeiteJSON(f.js, lt[0], lt[1], f.next); got != soll {
				t.Errorf("%s mit limit %d und total %d: Body %q, erwartet %q", f.name, lt[0], lt[1], got, soll)
			}
		}
	}
	gleichAuf(t, "Zahl der Fälle", len(faelle), 8)
}

// TestI11WerteDerZeile hält I11 (SPEC-293 Nr. 6; SPEC-292 Nr. 7): M-Zeit, die 7 Werte von Teil 1 an `started_at` und an
// `finished_at` (14 Fälle, die Zeit als Zeichenkette des Vertrags, die Nullzeit `null`), `response_status` 0 (`null`) und 200 und
// 1, `key_redacted` beide Werte, `path` mit 6 Zeichenformen (`<&>`, Anführungszeichen, Backslash, Byte 0x01, Byte 0xFF als
// U+FFFD, `ä`; aus Bytewerten, dekodiert verglichen); 14 + 3 + 2 + 6 = 25 Fälle. Zählprüfung: 2 Zeitfelder in ExchangeZeile.
func TestI11WerteDerZeile(t *testing.T) {
	zeitfelder := 0
	for i := range reflect.TypeFor[inbound.ExchangeZeile]().NumField() {
		if reflect.TypeFor[inbound.ExchangeZeile]().Field(i).Type == reflect.TypeFor[time.Time]() {
			zeitfelder++
		}
	}
	gleichAuf(t, "Zeitfelder und Werte", []int{zeitfelder, len(zeitFaelle())}, []int{2, 7})
	faelle := 0
	pruefe := func(name string, abw map[string]inspLit) {
		faelle++
		z, js := inspZeile(1, abw)
		got := inspHole(t, name, exchangesZiel("s"), "", func(f *inspektionFake) {
			f.seite = inbound.ExchangeSeite{Eintraege: []inbound.ExchangeZeile{z}, Limit: 1, Gesamt: 1}
		})
		if soll := inspSeiteJSON([]string{js}, 1, 1, "null"); got != soll {
			t.Errorf("%s: Body %q, erwartet %q", name, got, soll)
		}
	}
	for _, z := range zeitFaelle() {
		for _, feld := range []string{"StartedAt", "FinishedAt"} {
			pruefe(z.name+" an "+feld, map[string]inspLit{feld: {z.zeit, z.soll}})
		}
	}
	for _, s := range []int{0, 200, 1} {
		js := strconv.Itoa(s)
		if s == 0 {
			js = "null"
		}
		pruefe("response_status "+js, map[string]inspLit{"ResponseStatus": {s, js}})
	}
	for _, b := range []bool{true, false} {
		pruefe("key_redacted "+strconv.FormatBool(b), map[string]inspLit{"KeyRedacted": {b, strconv.FormatBool(b)}})
	}
	pfade := []string{"<&>", `"`, rueck(), string([]byte{0x01}), string([]byte{0xFF}), string([]byte{0xC3, 0xA4})}
	for _, p := range pfade {
		faelle++
		z, _ := inspZeile(1, map[string]inspLit{"Path": {p, ""}})
		got := inspHole(t, "Pfad", exchangesZiel("s"), "", func(f *inspektionFake) {
			f.seite = inbound.ExchangeSeite{Eintraege: []inbound.ExchangeZeile{z}}
		})
		var d struct{ Exchanges []struct{ Path string } }
		soll := strings.ToValidUTF8(p, string([]byte{0xEF, 0xBF, 0xBD}))
		if err := json.Unmarshal([]byte(got), &d); err != nil || len(d.Exchanges) != 1 || d.Exchanges[0].Path != soll {
			t.Errorf("Pfad %q: dekodiert %+v (%v), erwartet %q", p, d, err, soll)
		}
	}
	gleichAuf(t, "Fälle", faelle, 25)
}

// TestI10ZeichenInJedemTextfeld hält I10 und I11 für die Klasse „jedes Zeichen in jedem Textfeld“ (SPEC-292 Nr. 7 a): jedes der
// 10 Felder von ExchangeZeile mit Art `string` (ID, Method, Path und die 7 Zustände) mit jedem der 256 Bytes (`a<Byte>z`, ein
// Byte ab 0x80 ist allein kein UTF-8 und wird U+FFFD): 2560 Fälle, der dekodierte Wert des Feldes ist das Orakel. Dazu jeder
// Codepunkt außer den Surrogaten in `path` (1112064) in einem Request.
func TestI10ZeichenInJedemTextfeld(t *testing.T) {
	ersatz := string([]byte{0xEF, 0xBF, 0xBD})
	schluessel := map[string]string{}
	for _, m := range inspMitglieder() {
		schluessel[m.feld] = m.schluessel
	}
	var textfelder []string
	for i := range reflect.TypeFor[inbound.ExchangeZeile]().NumField() {
		if f := reflect.TypeFor[inbound.ExchangeZeile]().Field(i); f.Type.Kind() == reflect.String {
			textfelder = append(textfelder, f.Name)
		}
	}
	gleichAuf(t, "Zahl der Textfelder", len(textfelder), 10)
	u, f, _ := inspUmgebung(t)
	sende := func(feld, wert string) map[string]any {
		z, _ := inspZeile(1, nil)
		reflect.ValueOf(&z).Elem().FieldByName(feld).SetString(wert)
		f.seite = inbound.ExchangeSeite{Eintraege: []inbound.ExchangeZeile{z}}
		rec := u.rufe(inspReq(exchangesZiel("s"), ""))
		var d struct{ Exchanges []map[string]any }
		if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil || len(d.Exchanges) != 1 {
			t.Fatalf("Feld %s: Status %d, die Antwort ist kein JSON mit einer Zeile: %v", feld, rec.Code, err)
		}
		return d.Exchanges[0]
	}
	for _, feld := range textfelder {
		for b := range 256 {
			roh := "a" + string([]byte{byte(b)}) + "z"
			soll := roh
			if b >= 0x80 {
				soll = "a" + ersatz + "z"
			}
			if got := sende(feld, roh)[schluessel[feld]]; got != soll {
				t.Errorf("Feld %s, Byte 0x%02x: dekodiert %q, erwartet %q", feld, b, got, soll)
			}
		}
	}
	var alle []byte
	for r := rune(0); r <= 0x10FFFF; r++ {
		if r < 0xD800 || r > 0xDFFF {
			alle = append(alle, string(r)...)
		}
	}
	if got := sende("Path", string(alle))["path"]; got != string(alle) {
		t.Error("die Codepunkte kamen nicht unverändert in path an")
	}
}

// inspTextZustaende sind die Zustände eines Textwertes, die ein Feld unscheinbar machen: leer, nur Leerraum (Leerzeichen; Tab, CR und
// LF), Leerraum am Anfang, am Ende und an beiden Rändern.
func inspTextZustaende() []string { return []string{"", " ", "\t\r\n", " x", "x ", " x "} }

// inspZustandsZahl ist die Zahl der Zustände, die ein Feld der Art k kennt: Text 6, eine Zahl 0 und ein Wahrheitswert false je einen;
// jede andere Art (Zeit, Liste, Abbildung) hat andere Fänger.
func inspZustandsZahl(k reflect.Kind) int {
	switch k {
	case reflect.String:
		return len(inspTextZustaende())
	case reflect.Int, reflect.Bool:
		return 1
	}
	return 0
}

// inspSetzeZustand setzt das Feld v auf den Zustand z (Text: der z-te, Zahl: 0, Wahrheitswert: false).
func inspSetzeZustand(v reflect.Value, z int) {
	switch v.Kind() {
	case reflect.String:
		v.SetString(inspTextZustaende()[z])
	case reflect.Int:
		v.SetInt(0)
	case reflect.Bool:
		v.SetBool(false)
	}
}

// inspPruefeZustand prüft, dass das Objekt m das Mitglied key trägt und sein Wert der Zustand z des Feldes der Art k ist: ein Text
// dekodiert gleich dem Zustand (ob ein Zeichen als Escape oder roh steht, ist nicht zugesagt), eine Zahl das Literal `0` (nullbar: `null`),
// ein Wahrheitswert `false`.
func inspPruefeZustand(t *testing.T, name string, m map[string]json.RawMessage, key string, k reflect.Kind, z int, nullbar bool) {
	t.Helper()
	roh, da := m[key]
	if !da {
		t.Errorf("%s: das Mitglied %s fehlt", name, key)
		return
	}
	switch k {
	case reflect.String:
		var s string
		if err := json.Unmarshal(roh, &s); err != nil || s != inspTextZustaende()[z] {
			t.Errorf("%s: %s ist %s (%v), erwartet den Text %q", name, key, roh, err, inspTextZustaende()[z])
		}
	case reflect.Int:
		if soll := map[bool]string{false: "0", true: "null"}[nullbar]; string(roh) != soll {
			t.Errorf("%s: %s ist %s, erwartet %s", name, key, roh, soll)
		}
	case reflect.Bool:
		if string(roh) != "false" {
			t.Errorf("%s: %s ist %s, erwartet false", name, key, roh)
		}
	}
}

// inspRohVon zerlegt den Body eines Objekts in die Rohwerte seiner Mitglieder.
func inspRohVon(t *testing.T, name, body string) map[string]json.RawMessage {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Errorf("%s: der Body %q ist kein Objekt: %v", name, body, err)
	}
	return m
}

// inspZaehleArten liefert, wie viele Felder des Typs Text, Zahl, Wahrheitswert und eine andere Art haben (Zählprüfung gegen die
// Quelle, ohne feste Zahl der Felder).
func inspZaehleArten(typ reflect.Type) (text, zahl, wahr, andere int) {
	for i := range typ.NumField() {
		switch typ.Field(i).Type.Kind() {
		case reflect.String:
			text++
		case reflect.Int:
			zahl++
		case reflect.Bool:
			wahr++
		default:
			andere++
		}
	}
	return
}

// TestI10LeerUndNullInJedemFeld hält „kein Feld fehlt“ (SPEC-292 Nr. 7 a) und „Werte unverändert“ (SPEC-293 Nr. 6 bis 8) für den leeren
// Wert, die Zahl 0, false und Leerraum. Menge: jedes Feld der Typen `ExchangeZeile` (15), `ExchangeSeite` (4), `CoverageEintrag` (3) und
// `Diagnose` (3) aus dem Typ (reflect) mal die Zustände seiner Art: Text 6 (leer, nur Leerraum, Leerraum am Anfang, am Ende, an
// beiden Rändern), Zahl 0, Wahrheitswert false; dazu die Einträge von `settings` mit jedem der 6 Zustände im Wert und im Schlüssel.
// Jede Antwort trägt das Feld mit dem Zustand (Orakel: Literale `0`, `null` für `response_status` und `next`, `false`, der dekodierte
// Text) und keine anderen Felder. Zählprüfung gegen den Typ: Fälle gleich Textfelder mal 6 plus Zahlen plus Wahrheitswerte (ohne
// `evidence`, dessen Leerraum TestI12LeerraumAmRand führt), die übrigen Felder sind Zeiten, Listen und Abbildungen mit eigenem Fänger.
func TestI10LeerUndNullInJedemFeld(t *testing.T) {
	schluessel := map[string]string{}
	for _, m := range inspMitglieder() {
		schluessel[m.feld] = m.schluessel
	}
	schluesselSeite := map[string]string{"Eintraege": "exchanges", "Limit": "limit", "Gesamt": "total", "Weiter": "next"}
	schluesselCov := map[string]string{"EntitySet": "entity_set", "Status": "status", "Evidenz": "evidence"}
	nullbar := map[string]bool{"ResponseStatus": true, "Weiter": true}
	zustaende := len(inspTextZustaende())

	// ExchangeZeile: 15 Felder, ein Fall je Feld und Zustand.
	typ, faelle := reflect.TypeFor[inbound.ExchangeZeile](), 0
	for i := range typ.NumField() {
		f := typ.Field(i)
		if schluessel[f.Name] == "" {
			t.Errorf("ExchangeZeile.%s hat keinen Schlüssel im Vertrag", f.Name)
		}
		for z := range inspZustandsZahl(f.Type.Kind()) {
			faelle++
			zeile, _ := inspZeile(1, nil)
			inspSetzeZustand(reflect.ValueOf(&zeile).Elem().Field(i), z)
			got := inspHole(t, "ExchangeZeile."+f.Name, exchangesZiel("s"), "", func(fk *inspektionFake) {
				fk.seite = inbound.ExchangeSeite{Eintraege: []inbound.ExchangeZeile{zeile}, Limit: 1, Gesamt: 1}
			})
			var seite struct{ Exchanges []map[string]json.RawMessage }
			if err := json.Unmarshal([]byte(got), &seite); err != nil || len(seite.Exchanges) != 1 || len(seite.Exchanges[0]) != 15 {
				t.Errorf("ExchangeZeile.%s: Body %q (%v)", f.Name, got, err)
				continue
			}
			inspPruefeZustand(t, "ExchangeZeile."+f.Name, seite.Exchanges[0], schluessel[f.Name], f.Type.Kind(), z, nullbar[f.Name])
		}
	}
	text, zahl, wahr, andere := inspZaehleArten(typ)
	gleichAuf(t, "ExchangeZeile: Fälle", faelle, text*zustaende+zahl+wahr)
	gleichAuf(t, "ExchangeZeile: Zeitfelder", andere, 2)

	// ExchangeSeite: die Zahlen Limit, Gesamt und Weiter auf 0, die anderen mit eigenem Wert.
	typ, faelle = reflect.TypeFor[inbound.ExchangeSeite](), 0
	for i := range typ.NumField() {
		f := typ.Field(i)
		for z := range inspZustandsZahl(f.Type.Kind()) {
			faelle++
			seite := inbound.ExchangeSeite{Limit: 25, Gesamt: 99, Weiter: 7}
			inspSetzeZustand(reflect.ValueOf(&seite).Elem().Field(i), z)
			got := inspHole(t, "ExchangeSeite."+f.Name, exchangesZiel("s"), "", func(fk *inspektionFake) { fk.seite = seite })
			m := inspRohVon(t, "ExchangeSeite."+f.Name, got)
			if len(m) != 4 {
				t.Errorf("ExchangeSeite.%s: %d Mitglieder, erwartet 4", f.Name, len(m))
			}
			inspPruefeZustand(t, "ExchangeSeite."+f.Name, m, schluesselSeite[f.Name], f.Type.Kind(), z, nullbar[f.Name])
		}
	}
	text, zahl, wahr, andere = inspZaehleArten(typ)
	gleichAuf(t, "ExchangeSeite: Fälle", faelle, text*zustaende+zahl+wahr)
	gleichAuf(t, "ExchangeSeite: Liste", andere, 1)

	// CoverageEintrag: die Textfelder entity_set und status; evidence führt TestI12LeerraumAmRand.
	typ, faelle = reflect.TypeFor[inbound.CoverageEintrag](), 0
	for i := range typ.NumField() {
		f := typ.Field(i)
		if f.Name == "Evidenz" {
			continue
		}
		for z := range inspZustandsZahl(f.Type.Kind()) {
			faelle++
			e, _ := inspEintrag(1, `{}`)
			inspSetzeZustand(reflect.ValueOf(&e).Elem().Field(i), z)
			got := inspHole(t, "CoverageEintrag."+f.Name, inspCoverage, "baseline_id=b", func(fk *inspektionFake) { fk.eintraege = []inbound.CoverageEintrag{e} })
			var d struct{ Entries []map[string]json.RawMessage }
			if err := json.Unmarshal([]byte(got), &d); err != nil || len(d.Entries) != 1 || len(d.Entries[0]) != 3 {
				t.Errorf("CoverageEintrag.%s: Body %q (%v)", f.Name, got, err)
				continue
			}
			inspPruefeZustand(t, "CoverageEintrag."+f.Name, d.Entries[0], schluesselCov[f.Name], f.Type.Kind(), z, false)
		}
	}
	text, _, _, andere = inspZaehleArten(typ)
	gleichAuf(t, "CoverageEintrag: Fälle und andere Felder", []int{faelle, andere}, []int{(text - 1) * zustaende, 0})

	// Diagnose: mode (Text), settings (Abbildung: jeder Zustand im Wert und im Schlüssel) und die Liste errors (TestI14ZeichenInDiagnose
	// führt count 0).
	typ, faelle = reflect.TypeFor[inbound.Diagnose](), 0
	for i := range typ.NumField() {
		f := typ.Field(i)
		switch f.Type.Kind() {
		case reflect.String:
			for z := range zustaende {
				faelle++
				d := inbound.Diagnose{}
				inspSetzeZustand(reflect.ValueOf(&d).Elem().Field(i), z)
				got := inspHole(t, "Diagnose."+f.Name, inspDiagnostics, "", func(fk *inspektionFake) { fk.diagnose = d })
				m := inspRohVon(t, "Diagnose."+f.Name, got)
				if len(m) != 3 {
					t.Errorf("Diagnose.%s: %d Mitglieder, erwartet 3", f.Name, len(m))
				}
				inspPruefeZustand(t, "Diagnose."+f.Name, m, "mode", reflect.String, z, false)
			}
		case reflect.Map:
			for _, z := range inspTextZustaende() {
				for _, kv := range [][2]string{{"k", z}, {z, "v"}} {
					faelle++
					got := inspHole(t, "Diagnose.Einstellungen", inspDiagnostics, "", func(fk *inspektionFake) {
						fk.diagnose = inbound.Diagnose{Einstellungen: map[string]string{kv[0]: kv[1]}}
					})
					var d struct{ Settings map[string]string }
					if err := json.Unmarshal([]byte(got), &d); err != nil || len(d.Settings) != 1 {
						t.Errorf("settings %q: Body %q (%v)", kv, got, err)
					} else if v, da := d.Settings[kv[0]]; !da || v != kv[1] {
						t.Errorf("settings %q: Body %q, erwartet den Eintrag unverändert", kv, got)
					}
				}
			}
		}
	}
	gleichAuf(t, "Diagnose: Fälle", faelle, zustaende+zustaende*2)
}

// inspEintrag ist ein Eintrag der Coverage und sein JSON als Literal des Vertrags.
func inspEintrag(n int, evidenz string) (inbound.CoverageEintrag, string) {
	s := strconv.Itoa(n)
	return inbound.CoverageEintrag{EntitySet: "Set" + s, Status: model.CoverageStatus("st" + s), Evidenz: evidenz},
		`{"entity_set":"Set` + s + `","status":"st` + s + `","evidence":` + evidenz + `}`
}

// inspEvidenzen sind die gültigen Formen von M-Evidenz (SPEC-293 Nr. 7), die ersten 6 aus dem Plan, danach die Klasse
// „jeder gültige JSON-Wert“: Zeichen aus Bytewerten (U+2028 als UTF-8, eine Escape-Folge im Text), 64 Ebenen, 64 KiB, Leerraum
// zwischen den Token, jede Art des Wertes.
func inspEvidenzen() []string {
	return []string{
		`{"a":1}`, `{"t":"<&>"}`, `{"t":"` + string([]byte{0xE2, 0x80, 0xA8}) + `"}`, `{"t":"` + esc("0041") + rueck() + `n"}`,
		strings.Repeat("[", 64) + strings.Repeat("]", 64), `{"big":"` + strings.Repeat("a", 64*1024) + `"}`,
		`{ "a" : [ 1 , 2 ] }`, "{\n\t\"a\": 1\n}", `1`, `"x"`, `null`, `true`, `[]`, `{}`, `-0.5e+10`, `"` + esc("003c") + `"`, `[1,{"b":[null]}]`,
	}
}

// inspEvidenzImBody sendet einen Eintrag je Evidenz in der Folge der Liste und prüft den Body gegen das Literal des Vertrags: jede
// Evidenz steht mit denselben Bytes darin (der Body gleich dem Literal und bytes.Contains).
func inspEvidenzImBody(t *testing.T, name string, evidenzen ...string) {
	t.Helper()
	var es []inbound.CoverageEintrag
	var js []string
	for i, ev := range evidenzen {
		e, j := inspEintrag(i+1, ev)
		es, js = append(es, e), append(js, j)
	}
	got := inspHole(t, name, inspCoverage, "baseline_id=b", func(f *inspektionFake) { f.eintraege = es })
	if soll := `{"entries":[` + strings.Join(js, ",") + `]}`; got != soll {
		t.Errorf("%s: der Body trägt die Evidenz nicht mit denselben Bytes (Länge %d, erwartet %d)", name, len(got), len(soll))
	}
	for _, ev := range evidenzen {
		if !strings.Contains(got, ev) {
			t.Errorf("%s: der Body enthält die Evidenz der Länge %d nicht", name, len(ev))
		}
	}
}

// utfOrte sind die vier Orte, an denen eine Zeichenkette im JSON steht (Achse *Ort im JSON*, SPEC-293 Nr. 7): Wert eines Mitglieds,
// Schlüssel eines Mitglieds, Element eines Arrays, Wurzel-Zeichenkette; die Funktion setzt die Zeichenkette (ohne Anführungszeichen)
// ein.
func utfOrte() []func(s string) string {
	return []func(string) string{
		func(s string) string { return `{"a":"` + s + `"}` }, func(s string) string { return `{"` + s + `":1}` },
		func(s string) string { return `["` + s + `"]` }, func(s string) string { return `"` + s + `"` },
	}
}

// utfStellen sind die Stellen der Folge in der Zeichenkette (Achse *Stelle in der Zeichenkette*): Anfang, Mitte, Ende, jeweils mit
// dem Canary als Nachbarn (ASCII).
func utfStellen() []func(b string) string {
	return []func(string) string{
		func(b string) string { return b + canary() }, func(b string) string { return canary() + b + canary() },
		func(b string) string { return canary() + b },
	}
}

// utfGueltigeFolgen sind die gültigen Folgen der Menge *UTF-8 der Evidenz* (Achse *Folge*): die kleinste und die größte je Länge 2, 3
// und 4, aus Bytewerten.
func utfGueltigeFolgen() [][]byte {
	return [][]byte{{0xC2, 0x80}, {0xDF, 0xBF}, {0xE0, 0xA0, 0x80}, {0xED, 0x9F, 0xBF}, {0xEE, 0x80, 0x80}, {0xEF, 0xBF, 0xBF}, {0xF0, 0x90, 0x80, 0x80}, {0xF4, 0x8F, 0xBF, 0xBF}}
}

// utfUngueltigeFolgen sind die ungültigen Folgen: jedes der 128 Bytes 0x80 bis 0xFF allein und 6 Folgen jenseits der Grenze (überlang
// `C0 80`, `C1 BF`, `E0 9F BF`, `F0 8F BF BF`, Ersatzzeichen `ED A0 80`, über U+10FFFF `F4 90 80 80`), die gültige Nachbarn haben.
func utfUngueltigeFolgen() [][]byte {
	out := [][]byte{{0xC0, 0x80}, {0xC1, 0xBF}, {0xE0, 0x9F, 0xBF}, {0xED, 0xA0, 0x80}, {0xF0, 0x8F, 0xBF, 0xBF}, {0xF4, 0x90, 0x80, 0x80}}
	for b := 0x80; b <= 0xFF; b++ {
		out = append(out, []byte{byte(b)})
	}
	return out
}

// TestI12Coverage hält I12 (SPEC-293 Nr. 7): `coverage` antwortet 200 mit `{"entries":[…]}`, je Eintrag die drei Felder `entity_set`,
// `status`, `evidence` in dieser Folge, in der Folge des Dienstes; 0 (nil und leer), 1 und 3 Einträge. M-Evidenz gültig: jede der
// 17 Formen steht mit denselben Bytes im Body (Orakel: der Body gleich dem Literal und bytes.Contains). M-UTF-8 gültig: 8 Folgen
// (Achse *Folge*) mal 4 Orte im JSON in der Mitte der Zeichenkette gleich 32 Fälle, dasselbe Orakel. Zählprüfung: NumField von
// CoverageEintrag gleich 3.
func TestI12Coverage(t *testing.T) {
	gleichAuf(t, "Felder von CoverageEintrag", reflect.TypeFor[inbound.CoverageEintrag]().NumField(), 3)
	evidenzen := inspEvidenzen()
	gleichAuf(t, "Zahl der Evidenzen", len(evidenzen), 17)
	for _, ev := range evidenzen {
		inspEvidenzImBody(t, "Evidenz", ev)
	}
	gueltig, utfFaelle := utfGueltigeFolgen(), 0
	gleichAuf(t, "Zahl der gültigen Folgen und der Orte", []int{len(gueltig), len(utfOrte())}, []int{8, 4})
	for _, b := range gueltig {
		for _, ort := range utfOrte() {
			utfFaelle++
			inspEvidenzImBody(t, "UTF-8 gültig", ort(utfStellen()[1](string(b))))
		}
	}
	gleichAuf(t, "Fälle von M-UTF-8 gültig", utfFaelle, 32)
	var es []inbound.CoverageEintrag
	var js []string
	for n := 1; n <= 3; n++ {
		e, j := inspEintrag(n, `{"n":`+strconv.Itoa(n)+`}`)
		es, js = append(es, e), append(js, j)
	}
	faelle := []struct {
		name string
		es   []inbound.CoverageEintrag
		soll string
	}{
		{"nil", nil, `{"entries":[]}`}, {"leer", []inbound.CoverageEintrag{}, `{"entries":[]}`}, {"eine", es[:1], `{"entries":[` + js[0] + `]}`},
		{"drei", es, `{"entries":[` + strings.Join(js, ",") + `]}`},
		{"drei umgekehrt", []inbound.CoverageEintrag{es[2], es[1], es[0]}, `{"entries":[` + js[2] + "," + js[1] + "," + js[0] + `]}`},
		{"Folge 2, 3, 1", []inbound.CoverageEintrag{es[1], es[2], es[0]}, `{"entries":[` + js[1] + "," + js[2] + "," + js[0] + `]}`},
		{"Folge 3, 1, 2", []inbound.CoverageEintrag{es[2], es[0], es[1]}, `{"entries":[` + js[2] + "," + js[0] + "," + js[1] + `]}`},
	}
	for _, f := range faelle {
		if got := inspHole(t, f.name, inspCoverage, "baseline_id=b", func(fk *inspektionFake) { fk.eintraege = f.es }); got != f.soll {
			t.Errorf("%s: Body %q, erwartet %q", f.name, got, f.soll)
		}
	}
	ersatz := string([]byte{0xEF, 0xBF, 0xBD})
	schluessel := map[string]string{"EntitySet": "entity_set", "Status": "status"}
	for _, feld := range []string{"EntitySet", "Status"} {
		for b := range 256 {
			roh := "a" + string([]byte{byte(b)}) + "z"
			soll := roh
			if b >= 0x80 {
				soll = "a" + ersatz + "z"
			}
			e, _ := inspEintrag(1, `{}`)
			reflect.ValueOf(&e).Elem().FieldByName(feld).SetString(roh)
			got := inspHole(t, feld, inspCoverage, "baseline_id=b", func(f *inspektionFake) { f.eintraege = []inbound.CoverageEintrag{e} })
			var d struct{ Entries []map[string]any }
			if err := json.Unmarshal([]byte(got), &d); err != nil || len(d.Entries) != 1 || d.Entries[0][schluessel[feld]] != soll {
				t.Errorf("Feld %s, Byte 0x%02x: %q (%v), erwartet dekodiert %q", feld, b, got, err, soll)
			}
		}
	}
}

// TestI12LeerraumAmRand hält „mit denselben Bytes“ (SPEC-293 Nr. 7) für Leerraum am Rand der Evidenz, den json.Valid zulässt. M-Rand:
// jede der 17 gültigen Formen von inspEvidenzen mal 3 Positionen (am Anfang, am Ende, an beiden Rändern) mal 5 Folgen von Leerraum
// (Leerzeichen, Tab, LF, CR und `CR LF Tab Leerzeichen`) gleich 255 Evidenzen, je an 4 Stellen der Liste (allein, erster, mittlerer,
// letzter von drei Einträgen) gleich 1020 Requests; Orakel: der Body gleich dem Literal mit dem Leerraum darin.
func TestI12LeerraumAmRand(t *testing.T) {
	formen, raender := inspEvidenzen(), []string{" ", "\t", "\n", "\r", "\r\n\t "}
	gleichAuf(t, "Zahl der Formen und Ränder", []int{len(formen), len(raender)}, []int{17, 5})
	evidenzen, faelle := 0, 0
	for _, ev := range formen {
		for _, r := range raender {
			for _, mit := range []string{r + ev, ev + r, r + ev + r} {
				evidenzen++
				g1, g3 := `{"n":1}`, `{"n":3}`
				for _, liste := range [][]string{{mit}, {mit, g1, g3}, {g1, mit, g3}, {g1, g3, mit}} {
					faelle++
					inspEvidenzImBody(t, "Leerraum am Rand", liste...)
				}
			}
		}
	}
	gleichAuf(t, "Evidenzen und Requests", []int{evidenzen, faelle}, []int{255, 1020})
}

// inspUngueltigeEvidenzen sind die Formen von M-Evidenz ungültig (SPEC-293 Nr. 7): die 4 des Plans (leer, `{`, `{} {}`, `x`) und
// die Klasse „kein gültiges JSON“ aus der Grammatik: jedes echte Anfangsstück eines Dokuments (abgeschnitten), das Dokument
// zweimal (mit und ohne Leerraum dazwischen), mit Text dahinter und mit einem BOM davor.
func inspUngueltigeEvidenzen() []string {
	dok := `{"a":[1,"b"],"c":null}`
	out := []string{"", "{", "{} {}", "x"}
	for i := 1; i < len(dok); i++ {
		out = append(out, dok[:i])
	}
	return append(out, dok+dok, dok+" "+dok, dok+" x", dok+"x", string([]byte{0xEF, 0xBB, 0xBF})+dok, " ", `{"a":1,}`, `[1,]`)
}

// inspUngueltigerEintrag prüft die Liste es, die einen Eintrag mit ungültiger Evidenz trägt (ein Canary steht in ihm): 500
// `ODR-E9001` ohne einen Teil der Antwort, genau ein Zählen, kein Audit, der Fehler des Handlers nennt UTF-8 und keinen Wert, zwei Aufrufe
// von Coverage (Request und Handler).
func inspUngueltigerEintrag(t *testing.T, name string, es []inbound.CoverageEintrag) {
	t.Helper()
	u, f, routen := inspUmgebung(t)
	f.eintraege = es
	rec := u.rufe(inspReq(inspCoverage, "baseline_id=b"))
	pruefeFehlerAntwort(t, name, rec, 500, objekt500, "")
	if codes, _ := zaehlungenVon(u.tele); !slices.Equal(codes, []string{"ODR-E9001"}) || u.audit.Aufrufe() != 0 {
		t.Errorf("%s: CountError %q, %d Aufrufe von Audit", name, codes, u.audit.Aufrufe())
	}
	ant, err := routen[1].Handler(context.Background(), admin.Anfrage{RawQuery: "baseline_id=b"})
	if err == nil || ant.Status != 0 || len(ant.Body) != 0 || strings.Contains(err.Error(), canary()) || !strings.Contains(err.Error(), "UTF-8") {
		t.Errorf("%s: Handler liefert Fehler %v und Antwort %d %q (der Text nennt UTF-8 und keinen Wert)", name, err, ant.Status, ant.Body)
	}
	if n := f.zahlen(); n[inspOpCoverage] != 2 || n[inspOpExchanges] != 0 || n[inspOpDiagnose] != 0 {
		t.Errorf("%s: Zähler %v, erwartet 2 Aufrufe von Coverage", name, n)
	}
}

// TestI13UngueltigeEvidenz hält I13 (SPEC-293 Nr. 7): jede ungültige Evidenz ist 500 `ODR-E9001` ohne einen Teil der Antwort, an
// jeder Stelle der Liste (erster, mittlerer, letzter von drei Einträgen) und allein; der Fehler des Handlers trägt keinen Code
// und keinen Wert (ein Canary in der Evidenz steht in keiner Ausgabe). Zählprüfung: 4 + 21 + 8 Formen.
func TestI13UngueltigeEvidenz(t *testing.T) {
	formen := inspUngueltigeEvidenzen()
	gleichAuf(t, "Zahl der Formen", len(formen), 4+21+8)
	gut1, _ := inspEintrag(1, `{"n":1}`)
	gut3, _ := inspEintrag(3, `{"n":3}`)
	faelle := 0
	for _, ev := range slices.Concat(formen, []string{`{"` + canary()}) {
		schlecht, _ := inspEintrag(2, ev)
		for i, es := range [][]inbound.CoverageEintrag{{schlecht}, {schlecht, gut1, gut3}, {gut1, schlecht, gut3}, {gut1, gut3, schlecht}} {
			faelle++
			inspUngueltigerEintrag(t, "Evidenz Nr. "+strconv.Itoa(i), es)
		}
	}
	gleichAuf(t, "Fälle", faelle, (4+21+8+1)*4)
}

// TestI13UngueltigesUTF8 hält I13 für die Menge *UTF-8 der Evidenz* (SPEC-293 Nr. 7): gültiges JSON mit einem Byte oder einer Folge, die
// kein UTF-8 ist, ist 500 `ODR-E9001` wie jede ungültige Evidenz. Achsen: *Folge* (jedes der 128 Bytes 0x80 bis 0xFF allein und 6 Folgen
// jenseits der Grenze, 134) mal *Stelle in der Zeichenkette* (Anfang, Mitte, Ende, 3) mal *Ort im JSON* (Wert, Schlüssel, Element
// eines Arrays, Wurzel, 4) gleich 1608 Einträge allein in der Liste; *Stelle in der Liste*: die Vertreter `0x80`, `0xFF`, `C0 80`,
// `ED A0 80`, `F4 90 80 80` je als erster, mittlerer, letzter von drei Einträgen gleich 15. Die Bytes entstehen aus Bytewerten; das Orakel
// sind Status, Code und Body als Literale, nie eine Prüfung auf UTF-8.
func TestI13UngueltigesUTF8(t *testing.T) {
	folgen := utfUngueltigeFolgen()
	gleichAuf(t, "Zahl der Folgen, Stellen und Orte", []int{len(folgen), len(utfStellen()), len(utfOrte())}, []int{128 + 6, 3, 4})
	faelle := 0
	for _, b := range folgen {
		for _, stelle := range utfStellen() {
			for _, ort := range utfOrte() {
				faelle++
				schlecht, _ := inspEintrag(2, ort(stelle(string(b))))
				inspUngueltigerEintrag(t, "UTF-8 ungültig", []inbound.CoverageEintrag{schlecht})
			}
		}
	}
	gleichAuf(t, "Fälle allein in der Liste", faelle, 1608)
	gut1, _ := inspEintrag(1, `{"n":1}`)
	gut3, _ := inspEintrag(3, `{"n":3}`)
	vertreter := [][]byte{{0x80}, {0xFF}, {0xC0, 0x80}, {0xED, 0xA0, 0x80}, {0xF4, 0x90, 0x80, 0x80}}
	faelle = 0
	for _, b := range vertreter {
		schlecht, _ := inspEintrag(2, utfOrte()[0](utfStellen()[1](string(b))))
		for _, es := range [][]inbound.CoverageEintrag{{schlecht, gut1, gut3}, {gut1, schlecht, gut3}, {gut1, gut3, schlecht}} {
			faelle++
			inspUngueltigerEintrag(t, "UTF-8 ungültig in der Liste", es)
		}
	}
	gleichAuf(t, "Fälle in der Liste", faelle, 15)
}

// TestI14Diagnose hält I14 (SPEC-293 Nr. 8): `diagnostics` antwortet 200 mit `mode`, `settings` (Schlüssel aufsteigend nach den
// Bytes, ohne Eintrag `{}` und nie `null`) und `errors` in der Folge des Dienstes mit `code` und `count` über den ganzen Bereich
// von uint64 (ohne Verlust); `mode` mit 4 Werten, `settings` mit 0, 1, 3 Einträgen und nil, `errors` mit einem Eintrag je Code von
// model.AlleCodes() und nil. Zählprüfung: NumField von Diagnose gleich 3, von FehlerZaehler gleich 2.
func TestI14Diagnose(t *testing.T) {
	gleichAuf(t, "Felder von Diagnose und FehlerZaehler", []int{reflect.TypeFor[inbound.Diagnose]().NumField(), reflect.TypeFor[inbound.FehlerZaehler]().NumField()}, []int{3, 2})
	gleichAuf(t, "Text eines Codes", model.CodeE2102.String(), "ODR-E2102")
	zaehler := []uint64{0, 1, 9007199254740993, math.MaxInt64, math.MaxUint64}
	var fehler []inbound.FehlerZaehler
	var js []string
	for i, c := range model.AlleCodes() {
		n := zaehler[i%len(zaehler)]
		fehler = append(fehler, inbound.FehlerZaehler{Code: c, Anzahl: n})
		js = append(js, `{"code":"`+c.String()+`","count":`+strconv.FormatUint(n, 10)+`}`)
	}
	gleichAuf(t, "Zahl der Codes", len(fehler), len(model.AlleCodes()))
	ae := string([]byte{0xC3, 0xA4})
	faelle := []struct {
		name     string
		einstell map[string]string
		fehler   []inbound.FehlerZaehler
		soll     string
	}{
		{"settings nil", nil, nil, `"settings":{},"errors":[]`}, {"settings leer", map[string]string{}, nil, `"settings":{},"errors":[]`},
		{"ein Eintrag", map[string]string{"a": "1"}, nil, `"settings":{"a":"1"},"errors":[]`},
		{"drei, Bytefolge", map[string]string{"b": "2", "a": "1", "B": "3"}, nil, `"settings":{"B":"3","a":"1","b":"2"},"errors":[]`},
		{"mehr Schlüssel", map[string]string{ae: "7", "a_b": "6", "ab": "5", "_x": "4", "Z": "3", "a": "2", "": "1"}, nil,
			`"settings":{"":"1","Z":"3","_x":"4","a":"2","a_b":"6","ab":"5","` + ae + `":"7"},"errors":[]`},
		{"alle Codes", nil, fehler, `"settings":{},"errors":[` + strings.Join(js, ",") + `]`},
		{"ein Zähler", nil, fehler[:1], `"settings":{},"errors":[` + js[0] + `]`},
	}
	for _, f := range faelle {
		for _, mode := range []string{"record", "replay", "mock", "hybrid"} {
			got := inspHole(t, f.name, inspDiagnostics, "", func(fk *inspektionFake) {
				fk.diagnose = inbound.Diagnose{Modus: mode, Einstellungen: f.einstell, Fehler: f.fehler}
			})
			if soll := `{"mode":"` + mode + `",` + f.soll + `}`; got != soll {
				t.Errorf("%s, Modus %s: Body %q, erwartet %q", f.name, mode, got, soll)
			}
		}
	}
}

// TestI14ZeichenInDiagnose hält I14 für jedes Zeichen in jedem Textfeld: jedes der 256 Bytes (`a<Byte>z`) in `mode`, im Wert und
// im Schlüssel von `settings` (256 Requests mit je drei Stellen; ein Byte ab 0x80 ist U+FFFD); dazu 188 Zahlen von `count` ohne
// Verlust: jede Zehnerpotenz bis 10^19 mit ihren Nachbarn und jeder Wert 2^k und 2^k - 1, k von 0 bis 63 (Orakel: strconv in der Fixture).
func TestI14ZeichenInDiagnose(t *testing.T) {
	ersatz := string([]byte{0xEF, 0xBF, 0xBD})
	for b := range 256 {
		roh := "a" + string([]byte{byte(b)}) + "z"
		soll := roh
		if b >= 0x80 {
			soll = "a" + ersatz + "z"
		}
		got := inspHole(t, "Byte", inspDiagnostics, "", func(f *inspektionFake) {
			f.diagnose = inbound.Diagnose{Modus: roh, Einstellungen: map[string]string{roh: roh}}
		})
		var d struct {
			Mode     string
			Settings map[string]string
		}
		if err := json.Unmarshal([]byte(got), &d); err != nil || d.Mode != soll || len(d.Settings) != 1 || d.Settings[soll] != soll {
			t.Errorf("Byte 0x%02x: %q (%v), erwartet %q", b, got, err, soll)
		}
	}
	var zahlen []uint64
	for k := range 64 {
		zahlen = append(zahlen, 1<<k, 1<<k-1)
	}
	for p := uint64(1); p <= 1e19; p *= 10 {
		zahlen = append(zahlen, p, p-1, p+1)
		if p == 1e19 {
			break
		}
	}
	for _, n := range zahlen {
		got := inspHole(t, "count", inspDiagnostics, "", func(f *inspektionFake) {
			f.diagnose = inbound.Diagnose{Fehler: []inbound.FehlerZaehler{{Code: model.CodeE2102, Anzahl: n}}}
		})
		if soll := `{"mode":"","settings":{},"errors":[{"code":"ODR-E2102","count":` + strconv.FormatUint(n, 10) + `}]}`; got != soll {
			t.Errorf("count %d: Body %q, erwartet %q", n, got, soll)
		}
	}
}

// TestI15SettingsUngefiltert hält I15 (SPEC-293 Nr. 8, 11; akzeptiertes Negativ): die Route filtert `settings` nicht, ein
// Schlüssel außerhalb jeder Allowlist erscheint unverändert; gegen den Inhalt gibt es keinen Fänger (die Allowlist gehört dem
// Dienst, `SPEC-236` Nr. 7).
func TestI15SettingsUngefiltert(t *testing.T) {
	got := inspHole(t, "UPSTREAM_URL", inspDiagnostics, "", func(f *inspektionFake) {
		f.diagnose = inbound.Diagnose{Modus: "record", Einstellungen: map[string]string{"UPSTREAM_URL": "http://geheim.example"}}
	})
	if soll := `{"mode":"record","settings":{"UPSTREAM_URL":"http://geheim.example"},"errors":[]}`; got != soll {
		t.Errorf("Body %q, erwartet %q", got, soll)
	}
}

// inspErfolg ist die Erfolgsantwort einer Route für einen Fake, den einstellen vorbereitet, mit dem Body als Literal.
type inspErfolg struct {
	name, pfad, query string
	einstellen        func(f *inspektionFake)
	body              string
}

func inspErfolge() []inspErfolg {
	z, js := inspZeile(1, nil)
	e, ejs := inspEintrag(1, `{"a":1}`)
	return []inspErfolg{
		{"exchanges", exchangesZiel("s"), "limit=1", func(f *inspektionFake) {
			f.seite = inbound.ExchangeSeite{Eintraege: []inbound.ExchangeZeile{z}, Limit: 25, Gesamt: 99, Weiter: 7}
		}, inspSeiteJSON([]string{js}, 25, 99, "7")},
		{"coverage", inspCoverage, "baseline_id=b", func(f *inspektionFake) { f.eintraege = []inbound.CoverageEintrag{e} }, `{"entries":[` + ejs + `]}`},
		{"diagnostics", inspDiagnostics, "", func(f *inspektionFake) {
			f.diagnose = inbound.Diagnose{Modus: "record", Einstellungen: map[string]string{"a": "b"}, Fehler: []inbound.FehlerZaehler{{Code: model.CodeE2102, Anzahl: 3}}}
		}, `{"mode":"record","settings":{"a":"b"},"errors":[{"code":"ODR-E2102","count":3}]}`},
	}
}

// inspFehlerZellen liefert die Zellen von M-Fehler je Route (SPEC-293 Nr. 9): die 8 allgemeinen Ausgänge und die Codes der Tabelle;
// `diagnostics` hat keinen Code.
func inspFehlerZellen() map[string][]fehlerFall {
	gemeinsam := []fehlerFall{codeFall(model.CodeE3001, 500, "ODR-E3001", objektE3001), codeFall(model.CodeE9001, 500, "ODR-E9001", objekt500)}
	return map[string][]fehlerFall{
		"exchanges":   slices.Concat(allgemeineAusgaenge(), []fehlerFall{codeFall(model.CodeE7003, 404, "ODR-E7003", objektE7003)}, gemeinsam),
		"coverage":    slices.Concat(allgemeineAusgaenge(), []fehlerFall{codeFall(model.CodeE7002, 404, "ODR-E7002", objektE7002)}, gemeinsam),
		"diagnostics": allgemeineAusgaenge(),
	}
}

// TestI16I17Fehler hält I16 und I17 (SPEC-293 Nr. 9; SPEC-292 Nr. 8): M-Fehler, 8 allgemeine Ausgänge je Route und die Codes der
// Tabelle (30 Zellen) mit Status, Code und Body als Literale, genau einem Zählen je Fehlerantwort und keinem bei einem
// Kontextfehler (Verbindung abgebrochen, nichts geschrieben); M-Audit: dazu je Route ein Erfolg (33 Zellen), in jeder zählt
// `FakeAudit.Aufrufe()` 0. Der Dienst liefert bei „Wert und E7003“ beides, der Fehler gilt.
func TestI16I17Fehler(t *testing.T) {
	zellen := 0
	for _, route := range inspErfolge() {
		erfolg := fehlerFall{name: "Erfolg", status: 200, objekt: route.body}
		for _, f := range slices.Concat(inspFehlerZellen()[route.name], []fehlerFall{erfolg}) {
			zellen++
			name := route.name + " " + f.name
			u, fake, _ := inspUmgebung(t)
			if f.mitWert || f.err == nil {
				route.einstellen(fake)
			}
			fake.err = f.err
			rec := httptest.NewRecorder()
			r := inspReq(route.pfad, route.query)
			abbruch := abbruchVon(func() { u.srv.ServeHTTP(rec, r) })
			codes, andere := zaehlungenVon(u.tele)
			var soll []string
			switch {
			case f.abbruch:
				if abbruch != http.ErrAbortHandler || rec.Body.Len() != 0 || rec.Result().StatusCode != 200 {
					t.Errorf("%s: Panik %v, Body %q: erwartet den Abbruch der Verbindung ohne Antwort", name, abbruch, rec.Body.String())
				}
			case f.err == nil:
				if abbruch != nil || rec.Code != f.status || rec.Body.String() != f.objekt {
					t.Errorf("%s: Antwort %d %q, erwartet %d %q", name, rec.Code, rec.Body.String(), f.status, f.objekt)
				}
				pruefeErfolgsHeader(t, name, rec)
			default:
				soll = []string{f.code}
				if abbruch != nil {
					t.Errorf("%s: Panik %v", name, abbruch)
				}
				pruefeFehlerAntwort(t, name, rec, f.status, f.objekt, "")
			}
			if !slices.Equal(codes, soll) || andere != 0 {
				t.Errorf("%s: CountError %q, andere Aufrufe %d, erwartet %q", name, codes, andere, soll)
			}
			if n := u.audit.Aufrufe(); n != 0 {
				t.Errorf("%s: %d Aufrufe von Audit, erwartet 0", name, n)
			}
		}
	}
	gleichAuf(t, "Zellen von M-Audit", zellen, 33)
}

// TestI16Zahlen hält M-Fehler gegen die Quelle: 30 Zellen gleich 3 × 8 + 6, jeder Code der Tabelle von SPEC-293 Nr. 9 steht in
// model.AlleCodes() mit dem Status des Literals, und das Fehlerobjekt des Literals nennt die message des Codes.
func TestI16Zahlen(t *testing.T) {
	z := inspFehlerZellen()
	gleichAuf(t, "Zellen", len(z["exchanges"])+len(z["coverage"])+len(z["diagnostics"]), 30)
	gleichAuf(t, "Zellen gleich 3 × 8 + 6", 30, 3*8+(3+3))
	for route, fs := range z {
		for _, f := range fs {
			fe, ok := f.err.(*inbound.Fehler)
			if !ok || f.abbruch || f.name == "Wert und E7003" || f.name == "verpackter E7003" {
				continue
			}
			c, bekannt := model.ParseCode(f.code)
			if !bekannt || !slices.Contains(model.AlleCodes(), c) || c != fe.Code || !c.StatusErlaubt(f.status) || fe.Status != f.status ||
				!strings.Contains(f.objekt, `"message":"`+c.Message()+`"`) {
				t.Errorf("%s %s: Code %q, Status %d passt nicht zur Tabelle der Codes", route, f.name, f.code, f.status)
			}
		}
	}
}

// TestI16Unveraendert hält SPEC-293 Nr. 9 an den Handlern: jeder Fehler des Dienstes geht unverändert (dasselbe Objekt, nicht
// verpackt) und ohne Antwort zurück, auch wenn der Dienst Wert und Fehler liefert (30 Zellen); der Status der Erfolgsantwort liegt
// von 200 bis 299 und der Body ist nicht leer (`SPEC-238` Nr. 15 d).
func TestI16Unveraendert(t *testing.T) {
	zellen := 0
	for i, route := range inspErfolge() {
		for _, f := range inspFehlerZellen()[route.name] {
			zellen++
			_, fake, routen := inspUmgebung(t)
			route.einstellen(fake)
			fake.err = f.err
			ant, err := routen[i].Handler(context.Background(), admin.Anfrage{RawQuery: route.query, Pfadwerte: map[string]string{"id": "s"}})
			if err != f.err || ant.Status != 0 || len(ant.Body) != 0 {
				t.Errorf("%s %s: Handler liefert %v und Antwort %d %q, erwartet denselben Fehler und keine Antwort", route.name, f.name, err, ant.Status, ant.Body)
			}
		}
		_, fake, routen := inspUmgebung(t)
		route.einstellen(fake)
		ant, err := routen[i].Handler(context.Background(), admin.Anfrage{RawQuery: route.query, Pfadwerte: map[string]string{"id": "s"}})
		if err != nil || ant.Status < 200 || ant.Status > 299 || len(ant.Body) == 0 {
			t.Errorf("%s: Erfolg liefert Status %d, Body %q, Fehler %v", route.name, ant.Status, ant.Body, err)
		}
	}
	gleichAuf(t, "Zellen", zellen, 30)
}

// inspSchritt ist ein Request der Folge von M-Folge: was der Dienst liefert (einstellen und err) und die Antwort als Literal.
type inspSchritt struct {
	name       string
	einstellen func(f *inspektionFake)
	err        error
	status     int
	body       string
}

// inspZweiterErfolg ist eine zweite Erfolgsantwort je Route, die sich in jedem Feld von der ersten (inspErfolge) unterscheidet.
func inspZweiterErfolg(route string) (einstellen func(f *inspektionFake), body string) {
	switch route {
	case "exchanges":
		z2, j2 := inspZeile(2, nil)
		z3, j3 := inspZeile(3, nil)
		return func(f *inspektionFake) {
			f.seite = inbound.ExchangeSeite{Eintraege: []inbound.ExchangeZeile{z2, z3}, Limit: 3, Gesamt: 4}
		}, inspSeiteJSON([]string{j2, j3}, 3, 4, "null")
	case "coverage":
		e2, j2 := inspEintrag(2, `{"n":2}`)
		e3, j3 := inspEintrag(3, `[]`)
		return func(f *inspektionFake) { f.eintraege = []inbound.CoverageEintrag{e2, e3} }, `{"entries":[` + j2 + `,` + j3 + `]}`
	}
	return func(f *inspektionFake) {
		f.diagnose = inbound.Diagnose{Modus: "replay", Fehler: []inbound.FehlerZaehler{{Code: model.CodeE3001, Anzahl: math.MaxUint64}}}
	}, `{"mode":"replay","settings":{},"errors":[{"code":"ODR-E3001","count":18446744073709551615}]}`
}

// inspSchritte sind die 6 Schritte von M-Folge einer Route: zwei verschiedene Erfolge, ein Dienst ohne Inhalt, ein Fehler mit Code,
// ein Fehler ohne Code und der erste Erfolg noch einmal.
func inspSchritte(route inspErfolg) []inspSchritt {
	leer := map[string]string{"exchanges": inspLeereSeite, "coverage": inspLeereCoverage, "diagnostics": inspLeereDiagnose}[route.name]
	zweiter, zweiterBody := inspZweiterErfolg(route.name)
	nichts := func(*inspektionFake) {}
	return []inspSchritt{
		{"X", route.einstellen, nil, 200, route.body}, {"Y", zweiter, nil, 200, zweiterBody}, {"leer", nichts, nil, 200, leer},
		{"E3001", nichts, &inbound.Fehler{Code: model.CodeE3001, Status: 500}, 500, objektE3001},
		{"ohne Code", nichts, errors.New("x"), 500, objekt500}, {"X nochmal", route.einstellen, nil, 200, route.body},
	}
}

// TestI10WiederholungLiestNeu hält SPEC-293 Nr. 10 (jeder Aufruf liest den Stand neu, die Routen halten keinen Zustand) als Wirkung ohne
// Datenrennen: M-Folge ist jedes geordnete Paar aus den 6 Schritten (36 Paare je Route, auch derselbe Schritt zweimal), je auf
// demselben Server mit dem Fake neu eingestellt; jede Antwort ist die zu ihrem Schritt, nie die eines früheren (ein Zwischenspeicher
// einer Antwort oder eines Fehlers wird rot). Zählprüfung: 3 Routen × 36 Paare.
func TestI10WiederholungLiestNeu(t *testing.T) {
	paare := 0
	for _, route := range inspErfolge() {
		schritte := inspSchritte(route)
		gleichAuf(t, route.name+": Zahl der Schritte", len(schritte), 6)
		for _, x := range schritte {
			for _, y := range schritte {
				paare++
				u, fake, _ := inspUmgebung(t)
				for _, s := range []inspSchritt{x, y} {
					fake.seite, fake.eintraege, fake.diagnose, fake.err = inbound.ExchangeSeite{}, nil, inbound.Diagnose{}, s.err
					s.einstellen(fake)
					if rec := u.rufe(inspReq(route.pfad, route.query)); rec.Code != s.status || rec.Body.String() != s.body {
						t.Errorf("%s: %s dann %s, Schritt %s: %d %q, erwartet %d %q", route.name, x.name, y.name, s.name, rec.Code, rec.Body.String(), s.status, s.body)
					}
				}
				if n := fake.gesamt(); n != 2 {
					t.Errorf("%s: %s dann %s: %d Aufrufe des Dienstes, erwartet 2", route.name, x.name, y.name, n)
				}
			}
		}
	}
	gleichAuf(t, "Zahl der Paare", paare, 3*6*6)
}

// TestI16KeinWertAusDemRequest hält SPEC-293 Nr. 11 (die Antworten tragen nur die Felder von Nr. 6 bis 8): ein Canary in der Query (als
// Wert eines unbekannten Schlüssels und als Rest einer ungültigen Query), im Header, im Cookie und im Body steht in keinem Byte von Status,
// Header und Body der Antwort, bei Erfolg, bei einem Fehler mit Code (`ODR-E7003`, `ODR-E7002`) und bei einer Abweisung (400); 3 Routen
// mal 3 Ausgänge mal 2 Formen der Query.
func TestI16KeinWertAusDemRequest(t *testing.T) {
	kopf := map[string][]string{"X-Canary": {canary()}, "Cookie": {canary()}}
	faelle := 0
	for _, route := range inspErfolge() {
		for _, fehler := range []error{nil, inspE7003(), inspE7002()} {
			for _, query := range []string{route.query + "&x=" + canary(), route.query + "&%zz" + canary()} {
				faelle++
				u, fake, _ := inspUmgebung(t)
				route.einstellen(fake)
				fake.err = fehler
				rec := u.rufe(anfrageMit("GET", route.pfad+"?"+query, kopf, strings.NewReader(canary())))
				if strings.Contains(antwortText(rec), canary()) {
					t.Errorf("%s mit Query %q: der Canary steht in der Antwort: %s", route.name, query, antwortText(rec))
				}
			}
		}
	}
	gleichAuf(t, "Fälle", faelle, 3*3*2)
}

// TestI16RahmenSchreibtDasAbweisen hält SPEC-238 Nr. 15 b und c für die drei Routen: liefert der Dienst `ODR-E2101` mit 401, gibt die
// Route den Fehler zurück, und allein der Rahmen schreibt den Eintrag `admin_abgewiesen` und zählt einmal.
func TestI16RahmenSchreibtDasAbweisen(t *testing.T) {
	for _, route := range inspErfolge() {
		u, fake, _ := inspUmgebung(t)
		fake.err = &inbound.Fehler{Code: model.CodeE2101, Status: 401}
		rec := u.rufe(inspReq(route.pfad, route.query))
		pruefeFehlerAntwort(t, route.name, rec, 401, objekt401, "")
		gleichAuf(t, route.name+": Einträge", u.audit.Eintraege(),
			[]support.Eintrag{{Art: "admin_abgewiesen", ErrorCode: model.CodeE2101, KorrelationsID: rec.Header().Get("X-Correlation-Id")}})
		if codes, andere := zaehlungenVon(u.tele); !slices.Equal(codes, []string{"ODR-E2101"}) || andere != 0 {
			t.Errorf("%s: CountError %q, andere Aufrufe %d", route.name, codes, andere)
		}
	}
}
