package outboundtest

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Zeilen R1 und R2 prüfen die Antworten echter Dienste: das Olingo-Referenzszenario für V2 (Demodaten) und drei Antworten
// von V4. Die Dateien unter testdata/ sind die Rohdateien der Erfassung, Byte für Byte.

//go:embed testdata/v2_echt_*.json testdata/v4_fremd_*.json
var echtDaten embed.FS

// v2Datei liefert den Inhalt einer Datei unter testdata/; ein fehlender Inhalt ist ein Fehler des Aufbaus der Suite.
func v2Datei(name string) []byte {
	b, err := echtDaten.ReadFile("testdata/" + name)
	if err != nil {
		panic(fmt.Sprintf("Fixture %s: %v", name, err))
	}
	return b
}

// v2EchtHash sind der SHA-256 einer Rohdatei und ihre Größe in Bytes.
type v2EchtHash struct {
	hash  string
	bytes int
}

// v2EchtHashes liefert zu jeder Rohdatei ihren SHA-256 und ihre Größe.
func v2EchtHashes() map[string]v2EchtHash {
	return map[string]v2EchtHash{
		"v2_echt_14-top.json":         {"109526173fb866c081338e655303379a990cef8a58020b428d27b9f99a61019b", 2132},
		"v2_echt_20-inlinecount.json": {"de5de6fa9b05dcf000e379a484be5cd9e6d5f041876a8f91d5ba470d7a3c7774", 2146},
		"v2_echt_22-leer.json":        {"cda900e0bd7fb6118c2c63cc5f77ce2a6cce5524c695b77af4cbe82676a5bcc2", 20},
		"v2_echt_08-entity.json":      {"ff13bed873e035d32c875933821be535c2b37d84ca8bd060438fa8de68dd7f2a", 1062},
		"v2_echt_19-expand.json":      {"abb5493888680491ab4b44ca67af066b0a7bac894103f016a33b0df8530e0f38", 1299},
		"v2_echt_36-rooms.json":       {"68ec21f9c0a9e5f190532b314b1c3f868aeca6ba113c132ec9807b39af44dda8", 481},
		"v2_echt_35-teams.json":       {"d5cae4237461728e2af891ab6844f300d9c9bbc4b24b0f9b8fe38d534da03b42", 1017},
		"v2_echt_18-select.json":      {"87ef9da4cafb2ed2c41184cd585c2877ab223df023371186b0aee72d4b431958", 2561},
		"v2_echt_24-fehler.json":      {"d33cbc6f327222e9512b24d278e4522c319ce2ac5124cff56331a587345991e3", 94},
		"v4_fremd_olingo-04.json":     {"4f1a8cad8b32a97b1cf5ca10e70d7591e79f0d72dc00b7b7b5a7f5835910f196", 441},
		"v4_fremd_olingo-08.json":     {"33aa541982e0f8527a00bbfa1a4568ea29d4ad1c35314607d807fb1ff577428e", 121},
		"v4_fremd_net-21.json":        {"e792dd6e4e98fdca734c49e610aaf7e31f515f28ca8c94e33d8f408a9f8f67f9", 375},
	}
}

// Die erwarteten Entities der Echtfälle als Literale, nicht aus dem Leser gebildet (die Spalten sind die des Dienstes).
func echtEmployee(id, name, mgr, room, entry, image string) model.Entity {
	return model.Entity{Typ: "RefScenario.Employee", Schluessel: "'" + id + "'",
		Properties: []model.PropertyWert{pw("EmployeeId", model.ArtString, id), pw("EmployeeName", model.ArtString, name), pw("ManagerId", model.ArtString, mgr),
			pw("RoomId", model.ArtString, room), pw("TeamId", model.ArtString, "1"), pw("EntryDate", model.ArtDateTimeOffset, entry), pw("ImageUrl", model.ArtString, image)},
		Uebersprungen: []model.Uebersprungen{{Name: "Location", Grund: model.UebersprungTyp}, {Name: "Age", Grund: model.UebersprungTyp},
			{Name: "ne_Manager", Grund: model.UebersprungNavigation}, {Name: "ne_Team", Grund: model.UebersprungNavigation}, {Name: "ne_Room", Grund: model.UebersprungNavigation}}}
}

func echtZwei() []model.Entity {
	return []model.Entity{
		echtEmployee("1", "Walter Winter", "1", "1", "1999-01-01T00:00:00Z", "Employees('1')/$value"),
		echtEmployee("2", "Frederic Fall", "1", "2", "2003-07-01T00:00:00Z", "Employees('2')/$value"),
	}
}

func echtTeam(n, scrum string) model.Entity {
	return model.Entity{Typ: "RefScenario.Team", Schluessel: "'" + n + "'",
		Properties:    []model.PropertyWert{pw("Id", model.ArtString, n), pw("Name", model.ArtString, "Team "+n), pw("isScrumTeam", model.ArtBoolean, scrum)},
		Uebersprungen: []model.Uebersprungen{{Name: "nt_Employees", Grund: model.UebersprungNavigation}}}
}

// echtErwartet prüft die Entities einer Echtdatei gegen die Literale: die Seite oder die eine Entity.
func echtErwartet(name string, einzel bool, soll []model.Entity, anzahl model.Anzahl) func(*testing.T, jergebnis) {
	return func(t *testing.T, r jergebnis) {
		t.Helper()
		got := r.entities(jfall{einzel: einzel})
		if !reflect.DeepEqual(got, soll) {
			t.Errorf("%s: Entities %+v, erwartet %+v", name, got, soll)
		}
		if !einzel && r.seite.Anzahl != anzahl {
			t.Errorf("%s: Anzahl %+v, erwartet %+v", name, r.seite.Anzahl, anzahl)
		}
		if !einzel && r.seite.FolgeLink != (model.FolgeLink{}) {
			t.Errorf("%s: FolgeLink %+v, erwartet den Nullwert", name, r.seite.FolgeLink)
		}
	}
}

func v2FaelleR1() []jfall {
	raum := model.Entity{Typ: "RefScenario.Room", Schluessel: "'1'",
		Properties:    []model.PropertyWert{pw("Id", model.ArtString, "1"), pw("Name", model.ArtString, "Room 1")},
		Uebersprungen: []model.Uebersprungen{{Name: "Seats", Grund: model.UebersprungTyp}, {Name: "Version", Grund: model.UebersprungTyp}, {Name: "nr_Employees", Grund: model.UebersprungNavigation}, {Name: "nr_Building", Grund: model.UebersprungNavigation}}}
	ersterMitExpand := echtEmployee("1", "Walter Winter", "1", "1", "1999-01-01T00:00:00Z", "Employees('1')/$value")
	mk := func(name, datei, set string, einzel bool, soll []model.Entity, anzahl model.Anzahl) jfall {
		return jfall{name: "R1 " + name, m: "R", set: set, einzel: einzel, body: v2Datei(datei), nach: echtErwartet("R1 "+name, einzel, soll, anzahl)}
	}
	return []jfall{
		mk("14 top", "v2_echt_14-top.json", "Employees", false, echtZwei(), model.Anzahl{}),
		mk("20 inlinecount", "v2_echt_20-inlinecount.json", "Employees", false, echtZwei(), model.Anzahl{Vorhanden: true, Wert: 6}),
		mk("22 leer", "v2_echt_22-leer.json", "Employees", false, []model.Entity{}, model.Anzahl{}),
		mk("08 Entity", "v2_echt_08-entity.json", "Employees", true, []model.Entity{ersterMitExpand}, model.Anzahl{}),
		mk("19 expand", "v2_echt_19-expand.json", "Employees", true, []model.Entity{ersterMitExpand}, model.Anzahl{}),
		mk("36 rooms", "v2_echt_36-rooms.json", "Rooms", false, []model.Entity{raum}, model.Anzahl{}),
		mk("35 teams", "v2_echt_35-teams.json", "Teams", false, []model.Entity{echtTeam("1", "false"), echtTeam("2", "true"), echtTeam("3", "false")}, model.Anzahl{}),
	}
}

func v2FaelleR2() []jfall {
	st := model.DatenStruktur
	return []jfall{
		{name: "R2 18 select ohne Schlüssel", m: "R", set: "Employees", body: v2Datei("v2_echt_18-select.json"), ursache: model.DatenSchluessel, ort: "/d/results/0/EmployeeId"},
		{name: "R2 24 Fehlerobjekt (Entity)", m: "R", set: "Employees", einzel: true, body: v2Datei("v2_echt_24-fehler.json"), ursache: st, ort: "/d"},
		{name: "R2 24 Fehlerobjekt (Collection)", m: "R", set: "Employees", body: v2Datei("v2_echt_24-fehler.json"), ursache: st, ort: "/d"},
		{name: "R2 V4 olingo-04", m: "R", set: "Employees", body: v2Datei("v4_fremd_olingo-04.json"), ursache: st, ort: "/d"},
		{name: "R2 V4 olingo-08", m: "R", set: "Employees", einzel: true, body: v2Datei("v4_fremd_olingo-08.json"), ursache: st, ort: "/d"},
		{name: "R2 V4 net-21", m: "R", set: "Employees", body: v2Datei("v4_fremd_net-21.json"), ursache: st, ort: "/d"},
	}
}

func v2FaelleEcht() []jfall { return append(v2FaelleR1(), v2FaelleR2()...) }

func v2R1(t *testing.T, s *jsuite) {
	for name, soll := range v2EchtHashes() {
		b := v2Datei(name)
		if h := sha256.Sum256(b); hex.EncodeToString(h[:]) != soll.hash || len(b) != soll.bytes {
			t.Errorf("%s: SHA-256 %x und %d Bytes, erwartet %s und %d", name, h, len(b), soll.hash, soll.bytes)
		}
	}
	if len(v2EchtHashes()) != 12 {
		t.Errorf("R1: %d Dateien, erwartet 12", len(v2EchtHashes()))
	}
	s.pruefeAlle(t, v2FaelleR1())
}

func v2R2(t *testing.T, s *jsuite) { s.pruefeAlle(t, v2FaelleR2()) }
