package bootstrap_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/csdl"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodec"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/maintenance"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
	"github.com/pt9912/odata-recorder/src/hexagon/services/snapshot"
	"github.com/pt9912/odata-recorder/src/hexagon/services/workspace"
)

// Absturztest (SPEC-048 Nr. 1 bis 3, LH-QA-03): je Haltepunkt führt ein Kindprozess (das eigene Test-Binary) genau eine Operation aus und
// sendet sich am Haltepunkt SIGKILL; der Test öffnet Datei und Ablage neu, ruft Startup und vergleicht mit den Orakeln vorher und nachher.
// Haltepunkte sind Dekoratoren von WorkspaceTransactor, BlobRepository (samt Writer) und SnapshotRepository, die der Test in die Konfig der
// Dienste gibt.

// Umgebung des Kindprozesses; absturzEnvStelle ist nur im Kindprozess gesetzt.
const (
	absturzEnvOp     = "ABSTURZ_OP"
	absturzEnvStelle = "ABSTURZ_STELLE"
	absturzEnvDB     = "ABSTURZ_DB"
	absturzEnvBlobs  = "ABSTURZ_BLOBS"
	absturzEnvMarke  = "ABSTURZ_MARKE"
)

// Operationen und Stellen der Haltepunkte (SPEC-048 Nr. 2); absturzNie ist keine Stelle eines Dekorators (Kontrolle des Endes).
const (
	absturzMutation = "mutation"
	absturzReset    = "reset"
	absturzRestore  = "restore"
	absturzSnapshot = "snapshot"

	absturzVorCommit        = "vor-commit"
	absturzNachCommit       = "nach-commit"
	absturzVorBegin         = "vor-begin"
	absturzImWrite          = "im-write"
	absturzNachWriterCommit = "nach-writer-commit"
	absturzNachCreate       = "nach-create"
	absturzNie              = "nie"
)

// absturzFrist begrenzt den Kindprozess; sein eigenes -test.timeout liegt darunter.
const absturzFrist = 2 * time.Minute

// absturzHalt ist ein Haltepunkt: Operation, Stelle, ob nach dem Neustart der Stand nachher gilt, und die Zahl der Blobs, die Startup entfernt.
type absturzHalt struct {
	op, stelle string
	nachher    bool
	entfernt   int
}

// absturzHaltepunkte ist die Tabelle von SPEC-048 Nr. 2.
func absturzHaltepunkte() []absturzHalt {
	var l []absturzHalt
	for _, op := range []string{absturzMutation, absturzReset, absturzRestore} {
		l = append(l, absturzHalt{op: op, stelle: absturzVorCommit}, absturzHalt{op: op, stelle: absturzNachCommit, nachher: true})
	}
	return append(l,
		absturzHalt{op: absturzSnapshot, stelle: absturzVorBegin},
		absturzHalt{op: absturzSnapshot, stelle: absturzImWrite},
		absturzHalt{op: absturzSnapshot, stelle: absturzNachWriterCommit, entfernt: 1},
		absturzHalt{op: absturzSnapshot, stelle: absturzNachCreate, nachher: true})
}

// absturzAusloeser beendet den Prozess an seiner Stelle; der Nullwert tut nichts.
type absturzAusloeser struct{ stelle, marke string }

// an schreibt an der Stelle des Auslösers die Marke (Stelle und Befund der Stelle) und sendet dem eigenen Prozess SIGKILL.
func (a absturzAusloeser) an(stelle, befund string) {
	if a.stelle != stelle {
		return
	}
	if err := os.WriteFile(a.marke, []byte(stelle+" "+befund), 0o600); err != nil {
		panic(err)
	}
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGKILL); err != nil {
		panic(err)
	}
	select {}
}

// absturzTx hält in der Transaktion nach f (also nach dem letzten Schreiben) und nach dem COMMIT, vor der Rückkehr zum Dienst. Die Marke vor
// dem COMMIT nennt die Revision von ws-1, die die Transaktion sieht.
type absturzTx struct {
	unten outbound.WorkspaceTransactor
	halt  absturzAusloeser
}

func (d absturzTx) WithWorkspaceTransaction(ctx context.Context, f func(outbound.WorkspaceStores) error) error {
	err := d.unten.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
		if err := f(s); err != nil || d.halt.stelle != absturzVorCommit {
			return err
		}
		w, err := s.Workspaces.Get(ctx, "ws-1")
		if err != nil {
			return err
		}
		d.halt.an(absturzVorCommit, "revision="+strconv.FormatInt(w.Revision, 10))
		return nil
	})
	if err == nil {
		d.halt.an(absturzNachCommit, "")
	}
	return err
}

// absturzBlobs hält vor Begin; sein Writer hält im Write und nach dem Commit.
type absturzBlobs struct {
	outbound.BlobRepository
	halt absturzAusloeser
}

func (d absturzBlobs) Begin(ctx context.Context, grenze int64) (outbound.CaptureWriter, error) {
	d.halt.an(absturzVorBegin, "")
	w, err := d.BlobRepository.Begin(ctx, grenze)
	if err != nil {
		return w, err
	}
	return absturzWriter{CaptureWriter: w, halt: d.halt}, nil
}

type absturzWriter struct {
	outbound.CaptureWriter
	halt absturzAusloeser
}

// Write schreibt an der Stelle im-write die erste Hälfte der Bytes und hält danach; die Marke nennt geschriebene und gesamte Zahl.
func (w absturzWriter) Write(p []byte) (int, error) {
	if w.halt.stelle == absturzImWrite {
		n, err := w.CaptureWriter.Write(p[:len(p)/2])
		if err != nil {
			return n, err
		}
		w.halt.an(absturzImWrite, fmt.Sprintf("%d %d", n, len(p)))
	}
	return w.CaptureWriter.Write(p)
}

func (w absturzWriter) Commit() (string, int64, error) {
	ref, n, err := w.CaptureWriter.Commit()
	if err == nil {
		w.halt.an(absturzNachWriterCommit, ref)
	}
	return ref, n, err
}

// absturzSnapshots hält nach dem Anlegen des Workspace-Snapshots, vor der Rückkehr.
type absturzSnapshots struct {
	outbound.SnapshotRepository
	halt absturzAusloeser
}

func (d absturzSnapshots) CreateWorkspace(ctx context.Context, s model.WorkspaceSnapshot) error {
	err := d.SnapshotRepository.CreateWorkspace(ctx, s)
	if err == nil {
		d.halt.an(absturzNachCreate, s.ID)
	}
	return err
}

// absturzWartung ist der Wartungsdienst über der Umgebung u.
func absturzWartung(u *abUmgebung) (*maintenance.Wartung, error) {
	return maintenance.New(maintenance.Konfig{Sessions: u.db.Sessions(), Dienste: u.db.Metadata(), Speicher: u.db.Maintenance(), Blobs: u.waechter,
		Uhr: u.uhr, Log: slog.New(slog.NewTextHandler(io.Discard, nil)), Audit: supporttest.NewFakeAudit()})
}

// absturzPatch ändert den Namen eines Produkts von ws-1 über den Mock m; erwartet ist 204.
func absturzPatch(ctx context.Context, m *mock.Mocker, schluessel, name string) error {
	ist := abSende(ctx, m, "ws-1", abAnfrage{Methode: "PATCH", Pfad: "/Products(" + schluessel + ")", Body: `{"Name":"` + name + `"}`})
	if ist.Err != nil || ist.Status != 204 {
		return fmt.Errorf("PATCH Products(%s): Status %d, Fehler %v", schluessel, ist.Status, ist.Err)
	}
	return nil
}

// absturzMock führt die Mutation über einen Mock mit der Transaktion tx aus.
func absturzMock(ctx context.Context, u *abUmgebung, tx outbound.WorkspaceTransactor) error {
	m, err := mock.New(mock.Konfig{
		Dienste: []mock.Dienst{{ID: "svc", Root: "/odata"}}, Metadaten: u.db.Metadata(), Workspaces: u.db.Workspaces(), Entities: u.db.Entities(),
		Coverage: u.db.Coverage(), Blobs: u.store, MetadatenCodec: csdl.New(), URLCodec: urlparser.New(urlparser.WithFilter(urlparser.NewFilterLeser())),
		Schreiber: jsoncodec.NewSchreiber(), Eingabe: jsoncodec.NewEingabe(), Transaktion: tx, PublicBaseURL: "http://recorder.test",
		MaxBytes: 8388608, Telemetrie: telemetrieFuerTest(),
	})
	if err != nil {
		return err
	}
	return absturzPatch(ctx, m, "1", "Absturz")
}

// absturzWorkspace führt Reset oder Restore des Snapshots vor über einen Workspace-Dienst mit der Transaktion tx aus.
func absturzWorkspace(ctx context.Context, u *abUmgebung, op string, tx outbound.WorkspaceTransactor) error {
	wartung, err := absturzWartung(u)
	if err != nil {
		return err
	}
	ws, err := workspace.New(workspace.Konfig{Workspaces: u.db.Workspaces(), Transaktion: tx, Snapshots: u.db.Snapshots(), Blobs: u.waechter,
		Dienste: u.db.Metadata(), Wartung: wartung, Uhr: u.uhr, IDs: u.ids, Log: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Audit: supporttest.NewFakeAudit()})
	if err != nil {
		return err
	}
	if op == absturzReset {
		_, err = ws.Reset(ctx, "ws-1")
		return err
	}
	l, err := u.db.Snapshots().ListWorkspace(ctx, "ws-1")
	if err != nil {
		return err
	}
	i := slices.IndexFunc(l, func(s model.WorkspaceSnapshot) bool { return s.Name == "vor" })
	if i < 0 {
		return errors.New("Snapshot vor fehlt")
	}
	_, err = ws.Restore(ctx, "ws-1", l[i].ID)
	return err
}

// absturzErzeuge erzeugt den Snapshot neu über einen Snapshot-Dienst mit den Dekoratoren. Die Generatoren eines neu geöffneten Aufbaus beginnen
// bei 1: die erste Kennung trägt der Snapshot vor und wird hier verbraucht; die Uhr steht eine Stunde nach der des Aufbaus.
func absturzErzeuge(ctx context.Context, u *abUmgebung, halt absturzAusloeser) error {
	u.ids.SnapshotID()
	u.uhr.Advance(time.Hour)
	d, err := snapshot.New(snapshot.Konfig{Workspaces: u.db.Workspaces(), Entities: u.db.Entities(),
		Snapshots: absturzSnapshots{SnapshotRepository: u.db.Snapshots(), halt: halt},
		Blobs:     absturzBlobs{BlobRepository: u.waechter, halt: halt}, Uhr: u.uhr, IDs: u.ids})
	if err != nil {
		return err
	}
	_, err = d.Create(ctx, "ws-1", "neu")
	return err
}

// absturzFuehreAus führt die Operation op mit den Dekoratoren des Auslösers halt aus.
func absturzFuehreAus(ctx context.Context, u *abUmgebung, op string, halt absturzAusloeser) error {
	tx := absturzTx{unten: u.db, halt: halt}
	switch op {
	case absturzMutation:
		return absturzMock(ctx, u, tx)
	case absturzReset, absturzRestore:
		return absturzWorkspace(ctx, u, op, tx)
	case absturzSnapshot:
		return absturzErzeuge(ctx, u, halt)
	}
	return fmt.Errorf("unbekannte Operation %q", op)
}

// absturzAufbau legt die Ausgangsdaten in einer frischen Datei und Ablage an (SPEC-048 Nr. 1): ws-1 und ws-2 je mit fortgeschrittenem Cursor,
// eine Änderung an ws-1, der Snapshot vor, eine weitere Änderung; danach ist die Datenbank geschlossen.
func absturzAufbau(t *testing.T) (dbPfad, blobPfad string) {
	t.Helper()
	ctx := context.Background()
	dbPfad, blobPfad = filepath.Join(t.TempDir(), "r.db"), t.TempDir()
	u := abOeffne(t, dbPfad, blobPfad)
	u.ausgangsdaten()
	u.neuerWorkspace("ws-1")
	u.neuerWorkspace("ws-2")
	u.cursor("ws-1", 3)
	u.cursor("ws-2", 2)
	abMust(t, absturzPatch(ctx, u.mock, "1", "Eins"))
	_, err := u.snap.Create(ctx, "ws-1", "vor")
	abMust(t, err)
	abMust(t, absturzPatch(ctx, u.mock, "2", "Zwei"))
	u.schliesse()
	return dbPfad, blobPfad
}

// absturzBild ist, was nach dem Neustart verglichen wird: Entity-State, Revision, Cursors und Snapshots beider Workspaces.
type absturzBild struct {
	WS1, WS2       abStand
	Liste1, Liste2 []model.WorkspaceSnapshot
}

func absturzLies(ctx context.Context, u *abUmgebung) (b absturzBild, err error) {
	if b.WS1, err = abLeseStand(ctx, u.db, "ws-1"); err != nil {
		return b, err
	}
	if b.WS2, err = abLeseStand(ctx, u.db, "ws-2"); err != nil {
		return b, err
	}
	if b.Liste1, err = u.db.Snapshots().ListWorkspace(ctx, "ws-1"); err != nil {
		return b, err
	}
	b.Liste2, err = u.db.Snapshots().ListWorkspace(ctx, "ws-2")
	return b, err
}

// absturzOrakel ist das Bild eines Aufbaus ohne Absturz: bei leerem op vorher, sonst nach der vollständig ausgeführten Operation.
func absturzOrakel(t *testing.T, op string) absturzBild {
	t.Helper()
	ctx := context.Background()
	dbPfad, blobPfad := absturzAufbau(t)
	u := abOeffne(t, dbPfad, blobPfad)
	if op != "" {
		abMust(t, absturzFuehreAus(ctx, u, op, absturzAusloeser{}))
	}
	u = u.neustart()
	b, err := absturzLies(ctx, u)
	abMust(t, err)
	u.schliesse()
	return b
}

// absturzVergleiche nennt die Unterschiede zweier Bilder (leer: gleich).
func absturzVergleiche(ist, soll absturzBild) []string {
	var d []string
	for _, x := range abStandGleich(ist.WS1, soll.WS1) {
		d = append(d, "ws-1: "+x)
	}
	for _, x := range abStandGleich(ist.WS2, soll.WS2) {
		d = append(d, "ws-2: "+x)
	}
	if !reflect.DeepEqual(ist.Liste1, soll.Liste1) {
		d = append(d, fmt.Sprintf("ws-1 Snapshots: %+v gegen %+v", ist.Liste1, soll.Liste1))
	}
	if !reflect.DeepEqual(ist.Liste2, soll.Liste2) {
		d = append(d, fmt.Sprintf("ws-2 Snapshots: %+v gegen %+v", ist.Liste2, soll.Liste2))
	}
	return d
}

// absturzStartBefunde wertet Fehler und Bericht von Startup (SPEC-048 Nr. 3); entfernt ist die erwartete Zahl entfernter Blobs.
func absturzStartBefunde(b inbound.Wartungsbericht, err error, entfernt int) []string {
	var v []string
	if err != nil {
		v = append(v, fmt.Sprintf("Startup: Fehler %v", err))
	}
	for _, l := range []struct {
		name  string
		liste []string
	}{{"Befunde", b.Befunde}, {"BlobsFehlend", b.BlobsFehlend}, {"BlobsBeschaedigt", b.BlobsBeschaedigt}} {
		if len(l.liste) != 0 {
			v = append(v, fmt.Sprintf("Startup: %s %v", l.name, l.liste))
		}
	}
	if b.BlobsEntfernt != entfernt {
		v = append(v, fmt.Sprintf("Startup: BlobsEntfernt %d, erwartet %d", b.BlobsEntfernt, entfernt))
	}
	return v
}

// absturzAblage nennt jede Datei der Ablage, die kein referenzierter Blob ist (eine temporäre Datei liegt nie unter dem Pfad eines Blobs).
func absturzAblage(blobPfad string, referenziert []string) []string {
	var v []string
	err := filepath.WalkDir(blobPfad, func(pfad string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(blobPfad, pfad)
		if err != nil {
			return err
		}
		if filepath.Dir(rel) != "sha256" || !slices.Contains(referenziert, "sha256:"+d.Name()) {
			v = append(v, "Ablage: Datei ohne Referenz "+rel)
		}
		return nil
	})
	if err != nil {
		v = append(v, "Ablage: "+err.Error())
	}
	return v
}

// absturzWiederherstellen stellt jeden Snapshot von ws-1 wieder her, den jüngsten zuerst und vor zuletzt, und führt dann eine Mutation aus.
func absturzWiederherstellen(ctx context.Context, u *abUmgebung, liste []model.WorkspaceSnapshot) []string {
	var v []string
	vorGesehen := false
	for i := len(liste) - 1; i >= 0; i-- {
		vorGesehen = vorGesehen || liste[i].Name == "vor"
		if _, err := u.ws.Restore(ctx, "ws-1", liste[i].ID); err != nil {
			v = append(v, fmt.Sprintf("Restore %s: %v", liste[i].Name, err))
		}
	}
	if !vorGesehen {
		v = append(v, "Restore vor: der Snapshot fehlt")
	}
	if err := absturzPatch(ctx, u.mock, "3", "Danach"); err != nil {
		v = append(v, "weitere Mutation: "+err.Error())
	}
	return v
}

// absturzPruefe nennt jede Verletzung von SPEC-048 Nr. 3 an der neu geöffneten Umgebung u (leer: keine): Startup, der Vergleich mit soll, die
// Dateien der Ablage, Restore und eine weitere Mutation.
func absturzPruefe(ctx context.Context, u *abUmgebung, soll absturzBild, entfernt int) []string {
	wartung, err := absturzWartung(u)
	if err != nil {
		return []string{err.Error()}
	}
	bericht, err := wartung.Startup(ctx)
	v := absturzStartBefunde(bericht, err, entfernt)
	ist, err := absturzLies(ctx, u)
	if err != nil {
		return append(v, "Lesen: "+err.Error())
	}
	v = append(v, absturzVergleiche(ist, soll)...)
	refs, err := u.db.Maintenance().ReferencedBlobs(ctx)
	if err != nil {
		return append(v, "ReferencedBlobs: "+err.Error())
	}
	v = append(v, absturzAblage(u.blobPfad, refs)...)
	return append(v, absturzWiederherstellen(ctx, u, ist.Liste1)...)
}

// absturzDateien sind die Dateien eines Unterverzeichnisses der Ablage mit ihrer Größe.
func absturzDateien(t *testing.T, blobPfad, unter string) map[string]int64 {
	t.Helper()
	l, err := os.ReadDir(filepath.Join(blobPfad, unter))
	abMust(t, err)
	m := map[string]int64{}
	for _, e := range l {
		i, err := e.Info()
		abMust(t, err)
		m[e.Name()] = i.Size()
	}
	return m
}

// absturzLage nennt, vor dem Neustart gelesen, die Verstöße gegen die Lage, die der Haltepunkt hinterlässt: vor-commit sah die Transaktion die
// Revision nach der Operation; vor Begin ist nichts geschrieben; im Write liegt genau eine temporäre Datei mit einem Teil der Bytes; nach dem
// Commit des Writers und nach Create liegt ein Blob mehr und keine temporäre Datei. nach-commit belegt der Vergleich.
func absturzLage(t *testing.T, stelle, befund, blobPfad string, blobsVorher int, revisionNachher int64) []string {
	t.Helper()
	sha, tmp := absturzDateien(t, blobPfad, "sha256"), absturzDateien(t, blobPfad, "tmp")
	lage := fmt.Sprintf("Lage %s: %d Blobs (vorher %d), temporär %v, Befund %q", stelle, len(sha), blobsVorher, tmp, befund)
	var ok bool
	switch stelle {
	case absturzVorCommit:
		ok = befund == "revision="+strconv.FormatInt(revisionNachher, 10)
	case absturzNachCommit:
		ok = true
	case absturzVorBegin:
		ok = len(sha) == blobsVorher && len(tmp) == 0
	case absturzImWrite:
		var n, gesamt int64
		_, err := fmt.Sscan(befund, &n, &gesamt)
		ok = err == nil && n > 0 && n < gesamt && len(sha) == blobsVorher && len(tmp) == 1
		for _, groesse := range tmp {
			ok = ok && groesse == n
		}
	case absturzNachWriterCommit, absturzNachCreate:
		ok = len(sha) == blobsVorher+1 && len(tmp) == 0
	}
	if ok {
		return nil
	}
	return []string{lage}
}

// absturzStarte führt den Kindprozess für op und stelle aus. Der Fehlertext ist leer, wenn der Prozess innerhalb der Frist durch SIGKILL
// endete und die Marke der Stelle hinterließ; befund ist der Rest der Marke.
func absturzStarte(t *testing.T, op, stelle, dbPfad, blobPfad string) (befund, fehler string) {
	t.Helper()
	marke := filepath.Join(t.TempDir(), "marke")
	ctx, abbruch := context.WithTimeout(context.Background(), absturzFrist)
	defer abbruch()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestAbsturz$", "-test.count=1", "-test.timeout="+(absturzFrist-20*time.Second).String())
	// Der Race-Detektor wartet beim regulären Ende eines Prozesses ohne atexit_sleep_ms=0 eine Sekunde.
	cmd.Env = append(os.Environ(), absturzEnvOp+"="+op, absturzEnvStelle+"="+stelle, absturzEnvDB+"="+dbPfad, absturzEnvBlobs+"="+blobPfad,
		absturzEnvMarke+"="+marke, "GORACE="+os.Getenv("GORACE")+" atexit_sleep_ms=0")
	cmd.WaitDelay = 10 * time.Second
	ausgabe, err := cmd.CombinedOutput()
	var ende *exec.ExitError
	if ctx.Err() != nil {
		return "", fmt.Sprintf("Haltepunkt nicht erreicht: Kindprozess nach der Frist beendet\n%s", ausgabe)
	}
	if !errors.As(err, &ende) {
		return "", fmt.Sprintf("Haltepunkt nicht erreicht: Ende %v ohne SIGKILL\n%s", err, ausgabe)
	}
	if status, ok := ende.Sys().(syscall.WaitStatus); !ok || !status.Signaled() || status.Signal() != syscall.SIGKILL {
		return "", fmt.Sprintf("Haltepunkt nicht erreicht: Ende %v ohne SIGKILL\n%s", err, ausgabe)
	}
	b, err := os.ReadFile(marke)
	if err != nil {
		return "", fmt.Sprintf("SIGKILL ohne Marke: %v\n%s", err, ausgabe)
	}
	befund, ok := strings.CutPrefix(string(b), stelle+" ")
	if !ok {
		return "", fmt.Sprintf("Marke %q nennt nicht die Stelle %s", b, stelle)
	}
	return befund, ""
}

// absturzKind ist der Kindprozess: er öffnet Datei und Ablage der Umgebung und führt genau eine Operation aus.
func absturzKind(t *testing.T, stelle string) {
	u := abOeffne(t, os.Getenv(absturzEnvDB), os.Getenv(absturzEnvBlobs))
	halt := absturzAusloeser{stelle: stelle, marke: os.Getenv(absturzEnvMarke)}
	abMust(t, absturzFuehreAus(context.Background(), u, os.Getenv(absturzEnvOp), halt))
}

// absturzLaufe fährt einen Haltepunkt: Aufbau, Kindprozess, Lage, Neustart, Prüfung gegen die Orakel.
func absturzLaufe(t *testing.T, h absturzHalt, vorher, nachher absturzBild) {
	t.Helper()
	dbPfad, blobPfad := absturzAufbau(t)
	blobsVorher := len(absturzDateien(t, blobPfad, "sha256"))
	befund, fehler := absturzStarte(t, h.op, h.stelle, dbPfad, blobPfad)
	if fehler != "" {
		t.Fatal(fehler)
	}
	for _, v := range absturzLage(t, h.stelle, befund, blobPfad, blobsVorher, nachher.WS1.Revision) {
		t.Error(v)
	}
	soll := vorher
	if h.nachher {
		soll.WS1, soll.Liste1 = nachher.WS1, nachher.Liste1
	}
	for _, v := range absturzPruefe(context.Background(), abOeffne(t, dbPfad, blobPfad), soll, h.entfernt) {
		t.Error(v)
	}
}

// TestAbsturz (T1 bis T17): die 10 Haltepunkte von SPEC-048 Nr. 2 mit der Prüfung nach jedem. Derselbe Test ist mit gesetzter Umgebung der
// Kindprozess.
func TestAbsturz(t *testing.T) {
	if stelle := os.Getenv(absturzEnvStelle); stelle != "" {
		absturzKind(t, stelle)
		return
	}
	vorher := absturzOrakel(t, "")
	if len(vorher.WS1.Cursors) != 1 || vorher.WS1.Cursors[0].Position != 3 || len(vorher.WS2.Cursors) != 1 || len(vorher.Liste1) != 1 {
		t.Fatalf("Orakel vorher ohne die Ausgangsdaten: %+v", vorher)
	}
	gefahren := map[string]bool{}
	for _, h := range absturzHaltepunkte() {
		name := h.op + "-" + h.stelle
		t.Run(name, func(t *testing.T) {
			gefahren[name] = true
			nachher := absturzOrakel(t, h.op)
			if len(absturzVergleiche(nachher, vorher)) == 0 {
				t.Fatal("das Orakel nachher gleicht vorher")
			}
			absturzLaufe(t, h, vorher, nachher)
		})
	}
	if len(gefahren) != 10 {
		t.Errorf("%d Haltepunkte gefahren, erwartet 10", len(gefahren))
	}
}

// TestAbsturzEndeOhneHaltepunkt (T12): ein Kindprozess, der seine Stelle nicht erreicht, endet regulär und gilt als nicht erreicht.
func TestAbsturzEndeOhneHaltepunkt(t *testing.T) {
	for _, op := range []string{absturzMutation, absturzReset, absturzRestore, absturzSnapshot} {
		dbPfad, blobPfad := absturzAufbau(t)
		if _, fehler := absturzStarte(t, op, absturzNie, dbPfad, blobPfad); !strings.Contains(fehler, "Haltepunkt nicht erreicht: Ende <nil>") {
			t.Errorf("%s ohne Haltepunkt: %q", op, fehler)
		}
	}
}

// TestAbsturzStartBefunde (T13): Fehler, jede Befundliste und die Zahl der entfernten Blobs ergeben je eine Verletzung.
func TestAbsturzStartBefunde(t *testing.T) {
	if v := absturzStartBefunde(inbound.Wartungsbericht{BlobsEntfernt: 1}, nil, 1); len(v) != 0 {
		t.Errorf("ohne Befund: %v", v)
	}
	x := []string{"x"}
	for soll, f := range map[string]struct {
		b   inbound.Wartungsbericht
		err error
	}{
		"Startup: Fehler":           {err: errors.New("x")},
		"Startup: Befunde":          {b: inbound.Wartungsbericht{Befunde: x}},
		"Startup: BlobsFehlend":     {b: inbound.Wartungsbericht{BlobsFehlend: x}},
		"Startup: BlobsBeschaedigt": {b: inbound.Wartungsbericht{BlobsBeschaedigt: x}},
		"Startup: BlobsEntfernt":    {b: inbound.Wartungsbericht{BlobsEntfernt: 1}},
	} {
		if v := absturzStartBefunde(f.b, f.err, 0); len(v) != 1 || !strings.HasPrefix(v[0], soll) {
			t.Errorf("%s: %v", soll, v)
		}
	}
}

// absturzSchaden ist eine Verletzung, die der Test selbst an einer neu geöffneten Umgebung herstellt, und was die Prüfung dazu nennen muss.
type absturzSchaden struct {
	name  string
	tu    func(u *abUmgebung, vorBlob string) error
	nennt []string
}

// absturzBlobDatei ist die Datei einer Blob-Referenz.
func absturzBlobDatei(u *abUmgebung, ref string) string {
	return filepath.Join(u.blobPfad, "sha256", strings.TrimPrefix(ref, "sha256:"))
}

// absturzZusatz legt am Workspace id einen weiteren Snapshot an; die erste Kennung trägt vor.
func absturzZusatz(u *abUmgebung, id string) error {
	u.ids.SnapshotID()
	_, err := u.snap.Create(context.Background(), id, "zuviel")
	return err
}

// absturzRevision erhöht die Revision von ws-1 ohne Änderung am Entity-State.
func absturzRevision(u *abUmgebung) error {
	ctx := context.Background()
	w, err := u.db.Workspaces().Get(ctx, "ws-1")
	if err != nil {
		return err
	}
	_, err = u.db.Workspaces().Advance(ctx, "ws-1", w.Revision)
	return err
}

// absturzSchaeden ist je ein Schaden zu jedem Teil von absturzPruefe.
func absturzSchaeden() []absturzSchaden {
	meta := sha256.Sum256([]byte(abCSDL))
	return []absturzSchaden{
		{"Blob von vor fehlt", func(u *abUmgebung, vor string) error { return os.Remove(absturzBlobDatei(u, vor)) },
			[]string{"Startup: BlobsFehlend", "Restore vor"}},
		{"Blob von vor beschädigt", func(u *abUmgebung, vor string) error {
			return os.WriteFile(absturzBlobDatei(u, vor), []byte("x"), 0o600)
		}, []string{"Startup: BlobsBeschaedigt", "Restore vor"}},
		{"Metadaten fehlen", func(u *abUmgebung, _ string) error {
			return os.Remove(absturzBlobDatei(u, "sha256:"+hex.EncodeToString(meta[:])))
		}, []string{"Startup: BlobsFehlend", "weitere Mutation"}},
		{"temporäre Datei", func(u *abUmgebung, _ string) error {
			return os.WriteFile(filepath.Join(u.blobPfad, "tmp", "rest.tmp"), []byte("x"), 0o600)
		}, []string{"Ablage: Datei ohne Referenz tmp/rest.tmp"}},
		{"verwaister Blob", func(u *abUmgebung, _ string) error { u.blob([]byte("verwaist")); return nil }, []string{"Startup: BlobsEntfernt 1"}},
		{"Eintrag von ws-1", func(u *abUmgebung, _ string) error { return absturzPatch(context.Background(), u.mock, "4", "X") },
			[]string{"ws-1: Eintrag Products/4"}},
		{"Revision von ws-1", func(u *abUmgebung, _ string) error { return absturzRevision(u) }, []string{"ws-1: Revision"}},
		{"Cursor von ws-1", func(u *abUmgebung, _ string) error { u.cursor("ws-1", 4); return nil }, []string{"ws-1: Cursors"}},
		{"Cursor von ws-2", func(u *abUmgebung, _ string) error { u.cursor("ws-2", 4); return nil }, []string{"ws-2: Cursors"}},
		{"Snapshot von ws-1", func(u *abUmgebung, _ string) error { return absturzZusatz(u, "ws-1") }, []string{"ws-1 Snapshots"}},
		{"Snapshot von ws-2", func(u *abUmgebung, _ string) error { return absturzZusatz(u, "ws-2") }, []string{"ws-2 Snapshots"}},
	}
}

// TestAbsturzPruefungErkennt (T13 bis T17): die Prüfung ist ohne Schaden leer und nennt jeden Schaden, den der Test herstellt.
func TestAbsturzPruefungErkennt(t *testing.T) {
	soll := absturzOrakel(t, "")
	for _, s := range append([]absturzSchaden{{name: "ohne Schaden"}}, absturzSchaeden()...) {
		t.Run(s.name, func(t *testing.T) {
			dbPfad, blobPfad := absturzAufbau(t)
			u := abOeffne(t, dbPfad, blobPfad)
			if s.tu != nil {
				l, err := u.db.Snapshots().ListWorkspace(context.Background(), "ws-1")
				abMust(t, err)
				abMust(t, s.tu(u, l[0].StateBlob))
			}
			v := absturzPruefe(context.Background(), u, soll, 0)
			if s.tu == nil && len(v) != 0 {
				t.Errorf("Verletzungen ohne Schaden: %v", v)
			}
			for _, n := range s.nennt {
				if !slices.ContainsFunc(v, func(x string) bool { return strings.HasPrefix(x, n) }) {
					t.Errorf("die Prüfung nennt %q nicht: %v", n, v)
				}
			}
		})
	}
}
