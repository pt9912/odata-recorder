package main

import (
	"context"
	"errors"
	"fmt"
	"io"
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
	aufruf     = "Aufruf: uebergang import|zeige <DATA_DIR> <Session-ID>"
)

// lauf führt `import` oder `zeige` aus; ein fehlendes oder leeres Argument ist Exit 2, jeder andere Fehler
// Exit 1 mit dem Fehlercode auf stderr.
func lauf(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) != 3 || args[1] == "" || args[2] == "" || (args[0] != "import" && args[0] != "zeige") {
		_, _ = fmt.Fprintln(stderr, aufruf)
		return exitAufruf
	}
	s, err := session(ctx, args[1], args[2])
	if err == nil && args[0] == "import" {
		err = importiere(ctx, stdout, args[1], s)
	} else if err == nil {
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

// handle baut das Handle im Modus record über dir für den Service der Session, ohne Hochfahren: es öffnet
// weder eine Session noch einen Listener, und der Upstream wird nie gerufen.
func handle(ctx context.Context, dir, dienst string) (*bootstrap.Handle, error) {
	k, err := bootstrap.LadeKonfiguration([]string{
		"MODE=record", "LISTEN=127.0.0.1:8080", "DATA_DIR=" + dir, "SERVICE_ID=" + dienst,
		"SERVICE_ROOT_PATH=/odata", "UPSTREAM_URL=http://127.0.0.1:9/odata",
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
