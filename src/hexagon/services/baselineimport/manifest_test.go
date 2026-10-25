package baselineimport_test

import (
	"bytes"
	"encoding/json"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
)

// weltI1 ist die Session der Goldens: $metadata, die Collection Products mit zwei Entities, die
// Entity Orders(OrderID=7,LineNo=2) und zwei Exchanges, die der Import überspringt (fünf Exchanges).
func weltI1(t *testing.T) *welt {
	w := neu(t)
	b := w.standard()
	b.collection("Products", "", seite(produkt(1, "Milch", "1.5"), produkt(2, "Brot", "2")))
	b.entity("Orders", "OrderID=7,LineNo=2", ent(typBestell, "OrderID=7,LineNo=2", pw("OrderID", model.ArtInt32, "7"), pw("LineNo", model.ArtInt32, "2"), pw("Qty", model.ArtInt32, "3")))
	b.roh("/Products/$count", "")
	b.roh("/Products", "", func(e *model.Exchange) { e.Method = "POST" })
	return w
}

const (
	hashMeta  = "sha256:1a82d01d44a5d23f76f5192e02046eeefbd0c674f7841072f9a74de9d14cd3b4"
	hashMilch = "sha256:1933517318efffb6367223abc1bb533725f17187fc7e46dea7516c1d27fbdcea"
	hashBrot  = "sha256:c02a6c12716ecf5588876cca8c9f444134f7c221fbdda74c072e7635d9c64e89"
	hashOrder = "sha256:e170a6032ad3d27f742f6154b309a40ee46911079ba1debba036988e963f4d0a"
)

// teileI1 sind die Stücke der Goldens, in der Reihenfolge der Kanonform.
func coverageI1() string {
	return `[{"entity_set":"Orders","evidence":{"chains":[],"entities":1,"reason":"keine_ungefilterte_abfrage","responses":{"collection":0,"entity":1},"version":1},"status":"partial"},` +
		`{"entity_set":"Products","evidence":{"chains":[],"entities":2,"reason":"kette_nicht_belegt","responses":{"collection":1,"entity":0},"version":1},"status":"partial"},` +
		`{"entity_set":"Tags","evidence":{"chains":[],"entities":0,"reason":"keine_beobachtung","responses":{"collection":0,"entity":0},"version":1},"status":"unknown"}]`
}

func quellenI1() string {
	return `[{"elements":0,"entity_set":"","exchange_id":"exc-001","kind":"metadata","sequence_no":1,"session_id":"ses-1"},` +
		`{"elements":2,"entity_set":"Products","exchange_id":"exc-002","kind":"collection","sequence_no":2,"session_id":"ses-1"},` +
		`{"elements":1,"entity_set":"Orders","exchange_id":"exc-003","kind":"entity","sequence_no":3,"session_id":"ses-1"}]`
}

const snapI1 = "snp-00000000000000000000000000000001"

func manifestI1() string {
	return `{"baseline_snapshot_id":"` + snapI1 + `","coverage":` + coverageI1() + `,"created_at":"2026-10-25T12:00:00.000000000Z","entities":[` +
		`{"completeness":"complete","entity_set":"Orders","payload":{"LineNo":2,"OrderID":7,"Qty":3},"provenance":[{"exchange_id":"exc-003","hash":"` + hashOrder +
		`","projection":["LineNo","OrderID","Qty"],"session_id":"ses-1"}],"selected_properties":["LineNo","OrderID","Qty"],"typed_key":"OrderID=7,LineNo=2"},` +
		`{"completeness":"complete","entity_set":"Products","payload":{"Id":1,"Name":"Milch","Price":"1.5"},"provenance":[{"exchange_id":"exc-002","hash":"` + hashMilch +
		`","projection":["Id","Name","Price"],"session_id":"ses-1"}],"selected_properties":["Id","Name","Price"],"typed_key":"1"},` +
		`{"completeness":"complete","entity_set":"Products","payload":{"Id":2,"Name":"Brot","Price":"2"},"provenance":[{"exchange_id":"exc-002","hash":"` + hashBrot +
		`","projection":["Id","Name","Price"],"session_id":"ses-1"}],"selected_properties":["Id","Name","Price"],"typed_key":"2"}],` +
		`"entity_count":3,"import_id":"imp-1","manifest_version":1,"metadata_hash":"` + hashMeta + `","provenance":` + quellenI1() +
		`,"service_id":"catalog","snapshot_kind":"baseline","source_session_ids":["ses-1"]}`
}

func antwortI1() string {
	return `{"baseline_snapshot_id":"` + snapI1 + `","coverage":` + coverageI1() + `,"entity_count":3,"import_id":"imp-1","report":{"entities":{"distinct":3,"observations":3},` +
		`"exchanges":{"read":3,"skipped":[{"count":1,"reason":"methode"},{"count":1,"reason":"pfad"}],"total":5},"masked_properties":0,"metadata_hash":"` + hashMeta +
		`","provenance":` + quellenI1() + `,"service_id":"catalog","source_session_ids":["ses-1"],` +
		`"unsupported_features":{"entities":[{"entities":2,"entity_set":"Products","name":"Category","reason":"navigation"},` +
		`{"entities":2,"entity_set":"Products","name":"Rating","reason":"typ"}],"metadata":[]}}}`
}

// TestZeileI1Manifest: das Manifest in der Kanonform auf dem Referenzfall, unabhängig von der
// Reihenfolge, in der die Ablage die Exchanges liefert; der Antwortobjekt gleichfalls.
func TestZeileI1Manifest(t *testing.T) {
	w := weltI1(t)
	r := mustImport(t, w, anfrage())
	if got := string(w.blobs.writers[0].bytes()); got != manifestI1() {
		t.Errorf("Manifest:\n%s\nerwartet:\n%s", got, manifestI1())
	}
	if r.Import.ReportJSON != antwortI1() {
		t.Errorf("Antwortobjekt:\n%s\nerwartet:\n%s", r.Import.ReportJSON, antwortI1())
	}
	ordnungen := [][]int{{4, 3, 2, 1, 0}, {1, 2, 3, 4, 0}, {2, 0, 4, 1, 3}, {3, 4, 0, 1, 2}, {0, 2, 1, 4, 3}, {4, 0, 3, 2, 1}}
	for _, o := range ordnungen {
		w := weltI1(t)
		alt := slices.Clone(w.exchanges.m[sitzungID])
		for i, k := range o {
			w.exchanges.m[sitzungID][i] = alt[k]
		}
		r := mustImport(t, w, anfrage())
		if string(w.blobs.writers[0].bytes()) != manifestI1() || r.Import.ReportJSON != antwortI1() {
			t.Errorf("Ablage in der Reihenfolge %v liefert andere Bytes", o)
		}
	}
}

// TestZeileI2Kanonform: Strings, Zahlen, Schlüssel und die leere Liste mit Sollbytes von Hand (Menge 48).
func TestZeileI2Kanonform(t *testing.T) {
	n := 0
	pruefe := func(name string, v any, soll string) {
		t.Helper()
		n++
		got := baselineimport.Kanon(v)
		if string(got) != soll {
			t.Errorf("%s: %q, erwartet %q", name, got, soll)
		}
		var rueck any
		if err := json.Unmarshal(got, &rueck); err != nil {
			t.Errorf("%s: kein gültiges JSON: %v", name, err)
		}
		if s, ok := v.(string); ok && rueck != s {
			t.Errorf("%s: Rückweg %q", name, rueck)
		}
	}
	hex := "0123456789abcdef"
	for c := 0; c < 0x20; c++ {
		pruefe("Zeichen unter 0x20", string([]byte{byte(c)}), `"\u00`+string(hex[c>>4])+string(hex[c&15])+`"`)
	}
	for name, p := range map[string][2]string{
		"Anführungszeichen": {"\"", `"\""`}, "Backslash": {"\\", `"\\"`}, "Schrägstrich": {"/", `"/"`}, "DEL": {"\x7f", "\"\x7f\""},
		"U+0080": {"\xc2\x80", "\"\xc2\x80\""}, "U+2028": {"\xe2\x80\xa8", "\"\xe2\x80\xa8\""}, "U+2029": {"\xe2\x80\xa9", "\"\xe2\x80\xa9\""},
		"é": {"\xc3\xa9", "\"\xc3\xa9\""}, "U+1F600": {"\xf0\x9f\x98\x80", "\"\xf0\x9f\x98\x80\""}, "U+FFFD": {"\xef\xbf\xbd", "\"\xef\xbf\xbd\""},
	} {
		pruefe(name, p[0], p[1])
	}
	pruefe("kleiner als", "<>&", `"<>&"`)
	n-- // eine Ergänzung außerhalb der Menge
	pruefe("0", 0, "0")
	pruefe("1", 1, "1")
	pruefe("-1", -1, "-1")
	pruefe("MaxInt64", int64(math.MaxInt64), "9223372036854775807")
	pruefe("Schlüssel nach Bytes", map[string]any{"ab": 1, "é": 1, "Z": 1, "a": 1}, "{\"Z\":1,\"a\":1,\"ab\":1,\"\xc3\xa9\":1}")
	pruefe("leere Liste", []any{}, "[]")
	pruefe("Liste nil", []any(nil), "[]") // Ergänzung: eine Liste ohne Wert ist ebenfalls []
	n--
	// Ergänzung: Bytefolge, nicht UTF-16 oder Rune (U+FF61 vor U+10000).
	pruefe("Schlüssel U+FF61 und U+10000", map[string]any{"\xf0\x90\x80\x80": 1, "\xef\xbd\xa1": 1}, "{\"\xef\xbd\xa1\":1,\"\xf0\x90\x80\x80\":1}")
	n--
	if n != 48 {
		t.Errorf("Menge %d, erwartet 48", n)
	}
}

// TestZeileI3PayloadJeArt: der Payload je Art, als Bytes (Menge 10).
func TestZeileI3PayloadJeArt(t *testing.T) {
	w := neu(t)
	e := ent(typGeschenk, "1",
		pw("Id", model.ArtInt32, "42"),
		pw("S", model.ArtString, "\xc3\xa9\"\n"),
		pw("B", model.ArtBoolean, "true"),
		pw("I64", model.ArtInt64, "9007199254740993"),
		pw("Dec", model.ArtDecimal, "1.5"),
		pw("G", model.ArtGUID, "0f8fad5b-d9cb-469f-a165-70867728950e"),
		pw("D", model.ArtDate, "2020-01-01"),
		pw("DT", model.ArtDateTimeOffset, "2020-01-01T00:00:00Z"),
		pw("N", model.ArtNull, ""),
		pw("F", model.ArtBoolean, "false"),
	)
	w.standard().beobachtung("Products", false, e)
	mustImport(t, w, anfrage())
	soll := "\"payload\":{\"@odata.type\":\"#Shop.Gift\",\"B\":true,\"D\":\"2020-01-01\",\"DT\":\"2020-01-01T00:00:00Z\",\"Dec\":\"1.5\",\"F\":false," +
		"\"G\":\"0f8fad5b-d9cb-469f-a165-70867728950e\",\"I64\":\"9007199254740993\",\"Id\":42,\"N\":null,\"S\":\"\xc3\xa9\\\"\\u000a\"}"
	if got := w.blobs.writers[0].bytes(); !bytes.Contains(got, []byte(soll)) {
		t.Errorf("Payload nicht in den Bytes:\n%s\nerwartet:\n%s", got, soll)
	}
}

// TestZeileI4BlobSnapshotKennungZeit: Blob, Snapshot, Kennung und Zeit (Menge).
func TestZeileI4BlobSnapshotKennungZeit(t *testing.T) {
	var einzel []byte
	for _, stueck := range []int{0, -1} {
		w := neu(t)
		w.sessionR()
		if stueck < 0 {
			w.blobs.stueck = len(manifestLaenge(t))/3 + 1
		}
		r := mustImport(t, w, anfrage())
		wr := w.blobs.writers[0]
		pub := w.imports.pubs[0]
		if stueck == 0 {
			einzel = wr.bytes()
			if wr.aufruf != 1 {
				t.Errorf("Write %d-mal, erwartet 1", wr.aufruf)
			}
		} else {
			if wr.aufruf != 3 || !bytes.Equal(wr.bytes(), einzel) {
				t.Errorf("in drei Write: %d Aufrufe, gleiche Bytes %v", wr.aufruf, bytes.Equal(wr.bytes(), einzel))
			}
		}
		if g := w.p.argsVon("Blobs.Begin"); len(g) != 1 || g[0][0] != int64(math.MaxInt64) {
			t.Errorf("Begin %v", g)
		}
		if w.p.anzahl("Commit") != 1 || w.p.anzahl("IDs.SnapshotID") != 1 || w.p.anzahl("Uhr.Now") != 1 || w.p.anzahl("Abort") != 0 {
			t.Errorf("Commit %d, SnapshotID %d, Now %d, Abort %d", w.p.anzahl("Commit"), w.p.anzahl("IDs.SnapshotID"), w.p.anzahl("Uhr.Now"), w.p.anzahl("Abort"))
		}
		ref := refVon(wr.bytes())
		doc := w.exchanges.m[sitzungID][0].ResponseBlob
		if pub.Snapshot.StateBlob != ref || pub.Snapshot.StateHash != ref || pub.Snapshot.MetadataHash != doc {
			t.Errorf("Snapshot %+v", pub.Snapshot)
		}
		var m map[string]any
		_ = json.Unmarshal(wr.bytes(), &m)
		var rep map[string]any
		_ = json.Unmarshal([]byte(r.Import.ReportJSON), &rep)
		zeiten := []string{m["created_at"].(string)}
		if pub.Import.CreatedAt != pub.Snapshot.CreatedAt || !pub.Import.CreatedAt.Equal(t0()) || zeiten[0] != "2026-10-25T12:00:00.000000000Z" {
			t.Errorf("Zeiten: Import %v, Snapshot %v, Manifest %v", pub.Import.CreatedAt, pub.Snapshot.CreatedAt, zeiten[0])
		}
		ids := []string{m["baseline_snapshot_id"].(string), pub.Snapshot.ID, pub.Import.BaselineSnapshotID, rep["baseline_snapshot_id"].(string), r.Import.BaselineSnapshotID}
		for _, id := range ids {
			if id != ids[0] {
				t.Errorf("Kennungen verschieden: %v", ids)
			}
		}
	}
}

// manifestLaenge liefert die Bytes des Manifests des Referenzfalls sessionR (zur Wahl der Stückgröße).
func manifestLaenge(t *testing.T) []byte {
	w := neu(t)
	w.sessionR()
	mustImport(t, w, anfrage())
	return w.blobs.writers[0].bytes()
}

// TestZeileI5PublicationGueltig: die Publikation ist gültig und vollständig (Menge 3).
func TestZeileI5PublicationGueltig(t *testing.T) {
	modelle := map[string][]string{"leer": nil, "ein Set": {"Only"}, "drei Sets": {"Zebra", "Apple", "Mango"}}
	for name, sets := range modelle {
		w := neu(t)
		var es []model.EntitySet
		for _, s := range sets {
			es = append(es, model.EntitySet{Name: s, EntityType: typProdukt})
		}
		w.meta.vorgabe.modell = model.NeuesServiceModel("4.0", []string{"Shop"}, modellR().EntityTypes(), es, nil)
		w.sitzung().metadaten("<edmx/>")
		mustImport(t, w, anfrage())
		pub := w.imports.pubs[0]
		if err := pub.Validate(); err != nil {
			t.Errorf("%s: %v", name, err)
		}
		if !slices.Equal(pub.SourceSessionIDs, []string{sitzungID}) || len(pub.Coverage) != len(sets) {
			t.Errorf("%s: Quellen %v, Coverage %d", name, pub.SourceSessionIDs, len(pub.Coverage))
		}
		var namen []string
		for _, c := range pub.Coverage {
			namen = append(namen, c.EntitySet)
			if !bytes.Contains(w.blobs.writers[0].bytes(), []byte(`"evidence":`+c.EvidenceJSON+`,`)) {
				t.Errorf("%s: die Evidenz von %s steht nicht byte-gleich im Manifest", name, c.EntitySet)
			}
		}
		if !slices.IsSorted(namen) {
			t.Errorf("%s: Coverage nicht aufsteigend: %v", name, namen)
		}
	}
}

// TestZeileI6Antwortobjekt: das Objekt exchanges für 0, 1 und 3 Gründe des Überspringens.
func TestZeileI6Antwortobjekt(t *testing.T) {
	faelle := []struct {
		name  string
		pfade []string
		soll  string
	}{
		{"kein Grund", nil, `"exchanges":{"read":1,"skipped":[],"total":1}`},
		{"ein Grund", []string{"/Nothing"}, `"exchanges":{"read":1,"skipped":[{"count":1,"reason":"pfad"}],"total":2}`},
		{"drei Gründe", []string{"/Nothing", "POST", "404", "/Other"}, `"exchanges":{"read":1,"skipped":[{"count":1,"reason":"methode"},{"count":2,"reason":"pfad"},{"count":1,"reason":"status"}],"total":5}`},
	}
	for _, f := range faelle {
		w := neu(t)
		b := w.standard()
		for _, p := range f.pfade {
			switch p {
			case "POST":
				b.roh("/Products", "", func(e *model.Exchange) { e.Method = "POST" })
			case "404":
				b.roh("/Products", "", status(404))
			default:
				b.roh(p, "")
			}
		}
		r := mustImport(t, w, anfrage())
		if !strings.Contains(r.Import.ReportJSON, f.soll) {
			t.Errorf("%s: %s", f.name, r.Import.ReportJSON)
		}
		if !strings.Contains(r.Import.ReportJSON, `"provenance":[{"elements":0,"entity_set":"","exchange_id":"exc-001","kind":"metadata"`) {
			t.Errorf("%s: provenance fehlt", f.name)
		}
	}
}

// TestZeileI8LeererImport: ein Import ohne gelesene Antwort ist gültig (Menge 2).
func TestZeileI8LeererImport(t *testing.T) {
	w := neu(t)
	w.sitzung().metadaten("<edmx/>")
	r := mustImport(t, w, anfrage())
	m := manifest(t, w)
	if m["entity_count"].(float64) != 0 || w.p.anzahl("Metadaten.SetMetadata") != 1 {
		t.Errorf("entity_count %v, SetMetadata %d", m["entity_count"], w.p.anzahl("Metadaten.SetMetadata"))
	}
	for _, c := range m["coverage"].([]any) {
		if c.(map[string]any)["status"] != "unknown" {
			t.Errorf("Status %v", c)
		}
	}
	if len(m["coverage"].([]any)) != 3 || !bytes.Contains(w.blobs.writers[0].bytes(), []byte(`"entities":[]`)) || !strings.Contains(r.Import.ReportJSON, `"entity_count":0`) {
		t.Errorf("Manifest: %s", w.blobs.writers[0].bytes())
	}
	w = neu(t)
	w.meta.vorgabe.modell = model.NeuesServiceModel("4.0", []string{"Shop"}, nil, nil, nil)
	w.sitzung().metadaten("<edmx/>")
	r = mustImport(t, w, anfrage())
	if !bytes.Contains(w.blobs.writers[0].bytes(), []byte(`"coverage":[]`)) || !strings.Contains(r.Import.ReportJSON, `"coverage":[]`) {
		t.Errorf("coverage nicht []: %s", w.blobs.writers[0].bytes())
	}
}
