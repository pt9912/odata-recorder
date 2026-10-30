package workspace_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/workspace"
)

func t0() time.Time { return time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC) }

// welt ist ein Dienst mit allen Fakes; die Baseline "bsl-1" des Service "svc" liegt vor.
type welt struct {
	t       *testing.T
	p       *prot
	st      *store
	snaps   *snapshots
	blobs   *blobs
	dienste *dienste
	wartung *wartung
	uhr     zaehlUhr
	ids     zaehlIDs
	konfig  workspace.Konfig
	v       *workspace.Verwaltung
}

func neu(t *testing.T) *welt {
	t.Helper()
	p := neuesProt(t)
	w := &welt{t: t, p: p, st: neuerStore(p)}
	t.Cleanup(func() {
		if n := p.verboten.Load(); n != 0 {
			t.Errorf("%d Operationen gerufen, die der Dienst nie rufen darf", n)
		}
	})
	w.snaps = &snapshots{p: p, baselines: map[string]model.BaselineSnapshot{}}
	w.blobs = &blobs{p: p, inhalt: map[string][]byte{}, fehlerBei: -1}
	w.dienste = &dienste{p: p, alle: []model.Service{{ID: "svc"}}}
	w.wartung = &wartung{p: p}
	w.uhr = zaehlUhr{p: p, u: supporttest.NewFestUhr(t0())}
	w.ids = zaehlIDs{p: p, g: supporttest.NewFestIDs()}
	w.konfig = workspace.Konfig{
		Workspaces: w.st.workspaces(), Transaktion: w.st, Snapshots: w.snaps, Blobs: w.blobs, Dienste: w.dienste,
		Wartung: w.wartung, Uhr: w.uhr, IDs: w.ids, Log: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	v, err := workspace.New(w.konfig)
	if err != nil {
		t.Fatal(err)
	}
	w.v = v
	w.baseline("bsl-1", "svc", standardEnts())
	p.leeren()
	return w
}

// baselineBytes legt die Baseline id des Service svc mit dem Manifest m ab (Snapshot und Blob).
func (w *welt) baselineBytes(id, svc string, m []byte) string {
	sum := sha256.Sum256(m)
	ref := "sha256:" + hex.EncodeToString(sum[:])
	w.blobs.inhalt[ref] = m
	w.snaps.baselines[id] = model.BaselineSnapshot{ID: id, ServiceID: svc, CreatedAt: t0(), StateBlob: ref, StateHash: ref, MetadataHash: ref}
	return ref
}

func (w *welt) baseline(id, svc string, es []ent) string {
	return w.baselineBytes(id, svc, []byte(schreibe(manifestBaum(id, svc, es))))
}

// ent ist ein Eintrag der Baseline; payload ist der JSON-Text, wie er im Manifest steht.
type ent struct {
	set, key, payload string
	props             []string
	compl             string
}

func standardEnts() []ent {
	return []ent{
		{"Products", "Id=1", `{"Id":"1","Name":"Milk"}`, []string{"Id", "Name"}, "complete"},
		{"Products", "Id=2", `{"Id":"2"}`, []string{"Id"}, "partial"},
		{"Categories", "Id=1", `{"Id":"1","Title":"Food"}`, []string{"Id", "Title"}, "complete"},
	}
}

// erwarteterState ist der Entity-State, den die Anlage aus es für den Workspace id erzeugt.
func erwarteterState(id string, es []ent) []model.EntityState {
	r := make([]model.EntityState, 0, len(es))
	for _, e := range es {
		r = append(r, model.EntityState{WorkspaceID: id, EntitySet: e.set, EntityKey: e.key, Payload: e.payload,
			KnownProperties: append([]string{}, e.props...), Completeness: model.Completeness(e.compl)})
	}
	return r
}

type roh string
type obj = map[string]any

// schreibe schreibt v als JSON in der Kanonform von SPEC-127 Nr. 14 (Mitglieder aufsteigend, ohne
// Leerraum, Steuerzeichen als \u00xx); roh steht unverändert da. Der Schreiber ist vom Dienst unabhängig.
func schreibe(v any) string {
	var b strings.Builder
	schreibeIn(&b, v)
	return b.String()
}

func schreibeIn(b *strings.Builder, v any) {
	switch x := v.(type) {
	case nil:
		b.WriteString("null")
	case roh:
		b.WriteString(string(x))
	case string:
		b.WriteByte('"')
		for i := 0; i < len(x); i++ {
			c := x[i]
			switch {
			case c == '"':
				b.WriteString(`\"`)
			case c == '\\':
				b.WriteString(`\\`)
			case c < 0x20:
				fmt.Fprintf(b, "\\u%04x", c)
			default:
				b.WriteByte(c)
			}
		}
		b.WriteByte('"')
	case int:
		b.WriteString(strconv.Itoa(x))
	case bool:
		b.WriteString(strconv.FormatBool(x))
	case []any:
		b.WriteByte('[')
		for i, e := range x {
			if i > 0 {
				b.WriteByte(',')
			}
			schreibeIn(b, e)
		}
		b.WriteByte(']')
	case obj:
		namen := make([]string, 0, len(x))
		for k := range x {
			namen = append(namen, k)
		}
		sort.Strings(namen)
		b.WriteByte('{')
		for i, k := range namen {
			if i > 0 {
				b.WriteByte(',')
			}
			schreibeIn(b, k)
			b.WriteByte(':')
			schreibeIn(b, x[k])
		}
		b.WriteByte('}')
	default:
		panic(fmt.Sprintf("schreibe: %T", v))
	}
}

func eintragBaum(e ent) obj {
	props := make([]any, 0, len(e.props))
	for _, p := range e.props {
		props = append(props, p)
	}
	return obj{"completeness": e.compl, "entity_set": e.set, "payload": roh(e.payload), "provenance": []any{},
		"selected_properties": props, "typed_key": e.key}
}

// manifestBaum ist das Manifest der Baseline id des Service svc in der Form von SPEC-127 Nr. 14 a.
func manifestBaum(id, svc string, es []ent) obj {
	l := make([]any, 0, len(es))
	for _, e := range es {
		l = append(l, eintragBaum(e))
	}
	return obj{"baseline_snapshot_id": id, "coverage": []any{}, "created_at": "2026-10-20T10:00:00.000000000Z",
		"entities": l, "entity_count": len(es), "import_id": "imp-1", "manifest_version": 1,
		"metadata_hash": "sha256:" + strings.Repeat("a", 64), "provenance": []any{}, "service_id": svc,
		"snapshot_kind": "baseline", "source_session_ids": []any{"ses-1"}}
}

// kopf liefert das Glied name des Baums.
func eintraege(m obj) []any { return m["entities"].([]any) }

func esc(rest string) string { return "\\" + rest }

// Fehlerprüfungen.

func codeVon(err error) (*inbound.Fehler, bool) {
	var f *inbound.Fehler
	return f, errors.As(err, &f)
}

// pruefeCode: err ist ein *inbound.Fehler mit dem Code c, dem einen Status des Codes und der Ursache u.
func pruefeCode(t *testing.T, err error, c model.Code, u error) {
	t.Helper()
	f, ok := codeVon(err)
	if !ok {
		t.Errorf("kein *Fehler mit Code: %v", err)
		return
	}
	if f.Code != c || f.Status != c.Stati()[0] {
		t.Errorf("Code %v Status %d, erwartet %v", f.Code, f.Status, c)
	}
	if u != nil && !errors.Is(err, u) {
		t.Errorf("Ursache %v nicht erreichbar in %v", u, err)
	}
}

// pruefeOhneCode: err trägt ziel und keinen *inbound.Fehler.
func pruefeOhneCode(t *testing.T, err, ziel error) {
	t.Helper()
	if err == nil || !errors.Is(err, ziel) {
		t.Errorf("Fehler %v, erwartet %v", err, ziel)
		return
	}
	if _, ok := codeVon(err); ok {
		t.Errorf("%v trägt einen Code", err)
	}
}

// tabellenFehler nennt jeden Fehler, den die Tabelle der Abbildung benennt, und die Platte als einen
// ohne Zeile; TestTabellenFehlerVollstaendig hält die Liste gegen die Tabelle.
type tabellenFehler struct {
	name string
	err  error
}

var errPlatte = errors.New("Platte voll")

func tabellenFehlerListe() []tabellenFehler {
	return []tabellenFehler{
		{"ErrBlobFehlt", outbound.ErrBlobFehlt}, {"ErrBlobBeschaedigt", outbound.ErrBlobBeschaedigt},
		{"ErrBeschaedigt", outbound.ErrBeschaedigt}, {"ErrPfad", outbound.ErrPfad}, {"ErrZuGross", outbound.ErrZuGross},
		{"ErrUpstreamNichtErreichbar", outbound.ErrUpstreamNichtErreichbar}, {"ErrUpstreamZeit", outbound.ErrUpstreamZeit},
		{"ErrFremderService", outbound.ErrFremderService}, {"ErrDienstUnbekannt", outbound.ErrDienstUnbekannt},
		{"ErrFalscheArt", outbound.ErrFalscheArt}, {"ErrReferenziert", outbound.ErrReferenziert},
		{"ErrSessionGeschlossen", outbound.ErrSessionGeschlossen}, {"ErrUngueltigerUebergang", outbound.ErrUngueltigerUebergang},
		{"ErrNichtGefunden", outbound.ErrNichtGefunden}, {"ErrBereitsVorhanden", outbound.ErrBereitsVorhanden},
		{"ErrUngueltig", model.ErrUngueltig}, {"ErrMetadatenUngueltig", model.ErrMetadatenUngueltig},
		{"ErrDatenUngueltig", model.ErrDatenUngueltig}, {"ErrUngueltigeSequenz", outbound.ErrUngueltigeSequenz},
		{"ErrAbgeschlossen", outbound.ErrAbgeschlossen}, {"ErrRevisionVeraltet", outbound.ErrRevisionVeraltet},
		{"Platte (ohne Zeile)", errPlatte},
	}
}

func TestTabellenFehlerVollstaendig(t *testing.T) {
	liste := tabellenFehlerListe()
	namen := outbound.EintraegeAbbildung()
	if len(liste) != len(namen)+1 {
		t.Fatalf("%d Fehler in der Liste, %d in der Tabelle (+ einer ohne Zeile)", len(liste), len(namen))
	}
	hat := map[string]bool{}
	for _, z := range liste {
		hat[z.name] = true
	}
	for _, n := range namen {
		if !hat[n] {
			t.Errorf("%s fehlt in der Liste", n)
		}
	}
}

// kontextFehler sind die vier Formen eines Kontextfehlers, die ein Port liefern kann (Z07).
func kontextFehler() map[string]error {
	return map[string]error{
		"Canceled": context.Canceled, "DeadlineExceeded": context.DeadlineExceeded,
		"Join mit ErrBeschaedigt": errors.Join(context.Canceled, outbound.ErrBeschaedigt),
		"Join mit ErrFalscheArt":  errors.Join(context.DeadlineExceeded, outbound.ErrFalscheArt),
	}
}

// zustaendeGleich vergleicht zwei Zustände des Stores.
func zustaendeGleich(a, b zustand) bool {
	norm := func(z zustand) zustand {
		k := z.kopie()
		for id, cs := range k.cur {
			for i := range cs {
				if len(cs[i].Consumed) == 0 {
					cs[i].Consumed = nil
				}
			}
			k.cur[id] = cs
		}
		return k
	}
	return reflect.DeepEqual(norm(a), norm(b))
}

// vorbelegen legt den Workspace id der Baseline bsl-1 mit Revision rev und dem State der Baseline an.
func (w *welt) vorbelegen(id string, rev int64) {
	z := w.st.stand()
	z.ws[id] = model.Workspace{ID: id, ServiceID: "svc", BaselineSnapshotID: "bsl-1", Revision: rev, CreatedAt: t0()}
	z.ent[id] = erwarteterState(id, standardEnts())
	w.st.setze(z)
}

// veraendern legt Mutation, Tombstone, Cursor und Workspace-Snapshot in den Workspace id.
func (w *welt) veraendern(id string) {
	z := w.st.stand()
	es := z.ent[id]
	es[0].Payload = `{"Id":"1","Name":"Geaendert"}`
	es[0].Version = 3
	es[1].Deleted = true
	es[1].Version = 1
	es = append(es, model.EntityState{WorkspaceID: id, EntitySet: "Products", EntityKey: "Id=99", Payload: `{"Id":"99"}`, Completeness: model.EntityComplete, Version: 1})
	z.ent[id] = es
	z.cur[id] = []model.ReplayCursor{
		{WorkspaceID: id, ReplaySessionID: "rep-1", RecordingSessionID: "ses-1", ServiceID: "svc", Strategy: model.StrategySequential, Position: 2, Consumed: []int64{1, 2}},
		{WorkspaceID: id, ReplaySessionID: "rep-2", RecordingSessionID: "ses-1", ServiceID: "svc", Strategy: model.StrategyStrict, Position: 1, Consumed: []int64{5}},
	}
	z.snaps[id] = []string{"snp-ws-1"}
	w.st.setze(z)
}
