package replay_test

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/replay"
)

// alleOps sind die 34 Operationen der Ports Sessions (6), Exchanges (3), Blobs (5), Workspaces (5),
// Entities (5), Cursors (4), Snapshots (5) und Transaktion (1), wie das Protokoll sie nennt; die Telemetrie
// zählt der FakeTelemetry (fünf Operationen), zusammen 39.
func alleOps() []string {
	var ops []string
	for _, n := range []string{"Create", "Get", "SetStatus", "List", "FindOpen", "Delete"} {
		ops = append(ops, "Sessions."+n)
	}
	for _, n := range []string{"Append", "Get", "ListBySession"} {
		ops = append(ops, "Exchanges."+n)
	}
	for _, n := range []string{"Begin", "Open", "Verify", "List", "Remove"} {
		ops = append(ops, "Blobs."+n)
	}
	for _, n := range []string{"Create", "Get", "List", "Advance", "Delete"} {
		ops = append(ops, "tx Workspaces."+n)
	}
	for _, n := range []string{"Put", "Get", "List", "ListSet", "ReplaceAll"} {
		ops = append(ops, "tx Entities."+n)
	}
	for _, n := range []string{"Put", "Get", "List", "ResetAll"} {
		ops = append(ops, "tx Cursors."+n)
	}
	for _, n := range []string{"Get", "List", "CreateWorkspace", "GetWorkspace", "ListWorkspace"} {
		ops = append(ops, "tx Snapshots."+n)
	}
	return append(ops, "Transaktion.WithWorkspaceTransaction")
}

func telOps() []supporttest.Operation {
	return []supporttest.Operation{supporttest.OpLog, supporttest.OpStartSpan, supporttest.OpCountError, supporttest.OpSet, supporttest.OpEnd}
}

// TestNewKonfig hält Z1: 18 Verstöße und 5 Gutfälle, je Element ein Lauf.
func TestNewKonfig(t *testing.T) {
	type fall struct {
		name string
		mod  func(*replay.Konfig)
		gut  bool
	}
	var faelle []fall
	add := func(n string, gut bool, m func(*replay.Konfig)) { faelle = append(faelle, fall{n, m, gut}) }
	add("kein Dienst", false, func(k *replay.Konfig) { k.Dienste = nil })
	add("leere Kennung", false, func(k *replay.Konfig) { k.Dienste = []replay.Dienst{{ID: "", Root: "/a"}} })
	add("leeres Root", false, func(k *replay.Konfig) { k.Dienste = []replay.Dienst{{ID: "a", Root: ""}} })
	add("Sessions nil", false, func(k *replay.Konfig) { k.Sessions = nil })
	add("Exchanges nil", false, func(k *replay.Konfig) { k.Exchanges = nil })
	add("Blobs nil", false, func(k *replay.Konfig) { k.Blobs = nil })
	add("Transaktion nil", false, func(k *replay.Konfig) { k.Transaktion = nil })
	add("Telemetrie nil", false, func(k *replay.Konfig) { k.Telemetrie = nil })
	add("MaxBytes 0", false, func(k *replay.Konfig) { k.MaxBytes = 0 })
	add("MaxBytes -1", false, func(k *replay.Konfig) { k.MaxBytes = -1 })
	add("MaxBytes 1", true, func(k *replay.Konfig) { k.MaxBytes = 1 })
	for _, s := range []model.ReplayStrategy{"", "Sequential", "random"} {
		add("Strategie "+string(s), false, func(k *replay.Konfig) { k.Strategie = s })
	}
	for _, s := range []model.ReplayStrategy{model.StrategySequential, model.StrategyStrict, model.StrategyFirst, model.StrategyLast} {
		add("Strategie "+string(s), true, func(k *replay.Konfig) { k.Strategie = s })
	}
	for _, s := range []model.KodierungsModus{"", "x"} {
		add("InvalidEncoding "+string(s), false, func(k *replay.Konfig) { k.Regeln.InvalidEncoding = s })
	}
	for _, s := range []model.BodyModus{"", "x"} {
		add("BodyMode "+string(s), false, func(k *replay.Konfig) { k.Regeln.BodyMode = s })
	}
	add("Nullwert in Pfade", false, func(k *replay.Konfig) { k.Pfade = []model.JSONPfad{{}} })
	var schlecht, gut int
	for _, f := range faelle {
		u := neu(t)
		f.mod(&u.k)
		r, err := replay.New(u.k)
		if f.gut {
			gut++
			if err != nil || r == nil {
				t.Errorf("%s: Gutfall abgewiesen: %v", f.name, err)
			}
			continue
		}
		schlecht++
		if r != nil || !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%s: Dienst %v, Fehler %v", f.name, r, err)
		}
	}
	if schlecht != 18 || gut != 5 {
		t.Errorf("Menge: %d Verstöße, %d Gutfälle, erwartet 18 und 5", schlecht, gut)
	}
}

// TestNewRuftKeinenPort hält Z2: nach New zählt keine der 39 Operationen.
func TestNewRuftKeinenPort(t *testing.T) {
	u := neu(t)
	u.replayer()
	ops := alleOps()
	if len(ops) != 34 {
		t.Fatalf("%d Protokoll-Operationen, erwartet 34", len(ops))
	}
	for _, op := range ops {
		if n := u.p.anzahl(op); n != 0 {
			t.Errorf("%s: %d Aufrufe nach New", op, n)
		}
	}
	for _, op := range telOps() {
		if n := u.tel.Anzahl(op); n != 0 {
			t.Errorf("Telemetrie %d: %d Aufrufe nach New", op, n)
		}
	}
}

// TestArgumente hält Z3: fünf Verstöße ohne Port und ohne gelesenen Body, fünf Gutfälle.
func TestArgumente(t *testing.T) {
	type fall struct {
		name  string
		mod   func(*inbound.Anfrage)
		ctx   func() context.Context
		bad   bool
		body  bool
		laeng int64
	}
	ctx := func() context.Context { return context.Background() }
	var faelle []fall
	faelle = append(faelle, fall{name: "ctx nil", ctx: func() context.Context { return nil }, bad: true, laeng: -1})
	faelle = append(faelle, fall{name: "Methode leer", ctx: ctx, bad: true, laeng: -1, mod: func(a *inbound.Anfrage) { a.Methode = "" }})
	for _, l := range []int64{-2, -1, 0, 1} {
		for _, mitBody := range []bool{false, true} {
			bad := l == -2 || l == 1 && !mitBody
			faelle = append(faelle, fall{name: "Laenge " + strconv.FormatInt(l, 10) + " Body " + strconv.FormatBool(mitBody), ctx: ctx, bad: bad, body: mitBody, laeng: l})
		}
	}
	var schlecht, gut int
	for _, f := range faelle {
		u := neu(t)
		a := anf()
		a.Laenge = f.laeng
		var rd *zaehlLeser
		if f.body {
			rd = &zaehlLeser{daten: []byte("{}")}
			a.Body = rd
		}
		if f.mod != nil {
			f.mod(&a)
		}
		_, err := u.replay(f.ctx(), a)
		if f.bad {
			schlecht++
			if !errors.Is(err, model.ErrUngueltig) {
				t.Errorf("%s: Fehler %v", f.name, err)
			}
			if _, _, hat := fehlerCode(err); hat {
				t.Errorf("%s: Fehler mit Code", f.name)
			}
			if u.p.gesamt() != 0 || u.tel.Anzahl(supporttest.OpLog) != 0 {
				t.Errorf("%s: ein Port wurde gerufen", f.name)
			}
			if rd != nil && rd.gelesen != 0 {
				t.Errorf("%s: Body gelesen", f.name)
			}
			continue
		}
		gut++
		if errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%s: Gutfall abgewiesen: %v", f.name, err)
		}
	}
	if schlecht != 5 || gut != 5 {
		t.Errorf("Menge: %d Verstöße, %d Gutfälle, erwartet 5 und 5", schlecht, gut)
	}
}

// TestKeinUpstream hält Z45: die elf Felder des Vertrags bestehen mit Namen und Typ, kein Feld der Konfig ist ein
// Upstream-Port, und der Quelltext des Pakets (auch
// _test.go) importiert weder Adapter noch net/http noch einen anderen Dienst und nennt kein Symbol
// UpstreamClient.
func TestKeinUpstream(t *testing.T) {
	k := reflect.TypeOf(replay.Konfig{})
	// Die Konfig wächst additiv (SPEC-135 Nr. 1): die elf Felder des Vertrags bleiben mit Namen und Typ, weitere
	// Felder sind erlaubt und unterliegen derselben Prüfung auf Upstream-Ports.
	vertrag := map[string]string{
		"Dienste": "[]replay.Dienst", "Sessions": "outbound.SessionRepository", "Exchanges": "outbound.RecordingRepository",
		"Blobs": "outbound.BlobRepository", "Transaktion": "outbound.WorkspaceTransactor", "Regeln": "model.SchluesselRegeln",
		"Pfade": "[]model.JSONPfad", "MaxBytes": "int64", "Strategie": "model.ReplayStrategy", "RedigierteBodies": "bool",
		"Telemetrie": "support.Telemetry",
	}
	for name, typ := range vertrag {
		if f, da := k.FieldByName(name); !da || f.Type.String() != typ {
			t.Errorf("Feld %s fehlt oder hat einen anderen Typ (%v)", name, f.Type)
		}
	}
	for i := 0; i < k.NumField(); i++ {
		if strings.Contains(k.Field(i).Type.String(), "UpstreamClient") {
			t.Errorf("Feld %s ist ein Upstream-Port", k.Field(i).Name)
		}
	}
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	fset := token.NewFileSet()
	dateien := 0
	for _, e := range eintraege {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		dateien++
		for _, imp := range f.Imports {
			pfad, _ := strconv.Unquote(imp.Path.Value)
			eigen := strings.Contains(pfad, "/services/replay")
			if strings.Contains(pfad, "/adapters/") || pfad == "net/http" || strings.Contains(pfad, "/services/") && !eigen {
				t.Errorf("%s importiert %s", name, pfad)
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && id.Name == "UpstreamClient" {
				t.Errorf("%s nennt das Symbol UpstreamClient", name)
			}
			return true
		})
	}
	if dateien < 5 {
		t.Errorf("nur %d Dateien gelesen", dateien)
	}
}

// zaehlLeser ist ein Reader, der die gelesenen Bytes zählt.
type zaehlLeser struct {
	daten   []byte
	gelesen int
}

func (z *zaehlLeser) Read(p []byte) (int, error) {
	if len(z.daten) == 0 {
		return 0, io.EOF
	}
	n := copy(p, z.daten)
	z.daten = z.daten[n:]
	z.gelesen += n
	return n, nil
}
