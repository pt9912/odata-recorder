package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/tests/e2e/fixture"
)

const (
	exitOK     = 0
	exitFehler = 1
	exitAufruf = 2
	aufruf     = "Aufruf: uebergang import|zeige <DATA_DIR> <Session-ID> | uebergang reset <DATA_DIR> <Workspace-ID>"
)

// lauf führt `import`, `zeige` oder `reset` aus; ein fehlendes oder leeres Argument ist Exit 2, jeder andere
// Fehler Exit 1 mit dem Fehlercode auf stderr.
func lauf(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) != 3 || args[1] == "" || args[2] == "" || (args[0] != "import" && args[0] != "zeige" && args[0] != "reset") {
		_, _ = fmt.Fprintln(stderr, aufruf)
		return exitAufruf
	}
	var err error
	if args[0] == "reset" {
		err = zuruecksetzen(ctx, stdout, args[1], args[2])
	} else if s, serr := session(ctx, args[1], args[2]); serr != nil {
		err = serr
	} else if args[0] == "import" {
		err = importiere(ctx, stdout, args[1], s)
	} else {
		err = zeige(ctx, stdout, args[1], s)
	}
	if err != nil {
		_, _ = fmt.Fprintln(stderr, meldung(err))
		return exitFehler
	}
	return exitOK
}

// meldung nennt den Code eines Fehlers des Handle, sonst dessen Text.
func meldung(err error) string {
	var f *inbound.Fehler
	if errors.As(err, &f) {
		return f.Error()
	}
	return err.Error()
}

// session liest die Session über ein Handle mit einem beliebigen Service: das Handle für den Import braucht
// den Service der Session in seiner Konfiguration, und eine unbekannte Session ist ODR-E7003.
func session(ctx context.Context, dir, id string) (s model.Session, err error) {
	h, err := handle(ctx, dir, "uebergang")
	if err != nil {
		return s, err
	}
	defer func() { err = errors.Join(err, h.Herunterfahren(ctx)) }()
	return h.Sessions().Get(ctx, id)
}

// upstreamVar nennt die Umgebungsvariable mit der UPSTREAM_URL der Aufnahme (Dienst samt Root): der Import
// löst den `@odata.nextLink` einer Seite gegen sie auf (`SPEC-127` Nr. 12 e), und nur eine aufgelöste Kette
// ist `complete`. Ohne sie gilt eine Adresse, die kein Link trifft.
const upstreamVar = "E2E_UPSTREAM_URL"

// handle baut das Handle im Modus record über dir für den Service der Session, ohne Hochfahren: es öffnet
// weder eine Session noch einen Listener, und der Upstream wird nie gerufen.
func handle(ctx context.Context, dir, dienst string) (*bootstrap.Handle, error) {
	upstream := os.Getenv(upstreamVar)
	if upstream == "" {
		upstream = "http://127.0.0.1:9/odata"
	}
	k, err := bootstrap.LadeKonfiguration([]string{
		"MODE=record", "LISTEN=127.0.0.1:8080", "DATA_DIR=" + dir, "SERVICE_ID=" + dienst,
		"SERVICE_ROOT_PATH=/odata", "UPSTREAM_URL=" + upstream,
	})
	if err != nil {
		return nil, err
	}
	k.Listen = "127.0.0.1:0"
	return bootstrap.Build(ctx, k, nil)
}

func importiere(ctx context.Context, stdout io.Writer, dir string, s model.Session) (err error) {
	h, err := handle(ctx, dir, s.ServiceID)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, h.Herunterfahren(ctx)) }()
	erg, err := h.Import().Import(ctx, inbound.ImportAnfrage{
		ImportID: "e2e-" + s.ID, ServiceID: s.ServiceID, SourceSessionIDs: []string{s.ID}, MetadataHash: fixture.Hash(),
	})
	if err != nil {
		return err
	}
	ws, err := h.Workspaces().Create(ctx, erg.Import.BaselineSnapshotID, "")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "workspace=%s\nsession=%s\n", ws.ID, s.ID)
	return err
}

// zuruecksetzen ruft Reset des Workspace id über dir und schreibt die Revision, die der Reset liefert. Ein
// DATA_DIR, das nicht besteht oder kein Verzeichnis ist, wird nicht angelegt (SPEC-037 Nr. 6): LadeKonfiguration
// legt es an, darum prüft der Befehl vorher.
func zuruecksetzen(ctx context.Context, stdout io.Writer, dir, id string) (err error) {
	if fi, serr := os.Stat(dir); serr != nil || !fi.IsDir() {
		return fmt.Errorf("DATA_DIR %s besteht nicht oder ist kein Verzeichnis", dir)
	}
	h, err := handle(ctx, dir, "uebergang")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, h.Herunterfahren(ctx)) }()
	ws, err := h.Workspaces().Reset(ctx, id)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "revision=%d\n", ws.Revision)
	return err
}

func zeige(ctx context.Context, stdout io.Writer, dir string, s model.Session) (err error) {
	db, err := sqlite.Open(ctx, filepath.Join(dir, "odata-recorder.db"), sqlite.Options{})
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	ex, err := db.Recording().ListBySession(ctx, s.ID)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "status=%s exchanges=%d\n", s.Status, len(ex))
	return err
}
