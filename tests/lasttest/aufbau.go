package main

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// Gründe und Stellen eines gescheiterten Aufbaus: der Grund ist das Token der Ausgabe (SPEC-216 Nr. 13), die Stelle sagt, welche Prüfung
// oder welcher Schritt scheiterte, und erscheint in keiner Meldung.
const (
	grundAufbau = "aufbau"
	grundFrist  = "frist"

	stelleDaten    = "daten"
	stelleDienst   = "dienst"
	stelleAufnahme = "aufnahme"
	stelleStatus   = "status"
	stelleImport   = "import"
	stelleSession  = "session"
	stelleZahl     = "zahl"
	stelleCoverage = "coverage"
	stelleEntities = "entities"
	stelleFrist    = "frist"

	dienstID = "lasttest"
	dateiDB  = "odata-recorder.db"
)

// aufbauFehler ist `Aufbau gescheitert: <Token>` (SPEC-216 Nr. 6 f); die Meldung nennt keinen Wert aus Body oder Header.
type aufbauFehler struct {
	grund   string
	stelle  string
	ursache error
}

func (f *aufbauFehler) Error() string {
	if f.stelle == stelleDaten {
		return "Aufbau: DATA_DIR nicht leer"
	}
	return "Aufbau gescheitert: " + f.grund
}

func (f *aufbauFehler) Unwrap() error { return f.ursache }

// an meldet err als Fehler des Aufbaus an der Stelle; ein Fehler, der schon eine Stelle trägt, behält sie.
func an(stelle string, err error) error {
	var f *aufbauFehler
	if err == nil || errors.As(err, &f) {
		return err
	}
	return &aufbauFehler{grund: grundAufbau, stelle: stelle, ursache: err}
}

// handle ist, was der Aufbau vom Composition Root braucht, als Funktionen; baueEcht füllt sie aus *bootstrap.Handle, der Test ersetzt einzelne.
type handle struct {
	Hochfahren     func(ctx context.Context) error
	Herunterfahren func(ctx context.Context) error
	Adresse        func() string
	Sessions       func() inbound.SessionUseCase
	Import         func() inbound.ImportUseCase
	Workspaces     func() inbound.WorkspaceUseCase
}

// ablage ist die lesende Sicht auf die Datenbank des Aufbaus als Funktionen; oeffneAblage füllt sie aus dem Adapter sqlite, der Test ersetzt einzelne.
type ablage struct {
	Session   func(ctx context.Context, id string) (model.Session, error)
	Exchanges func(ctx context.Context, sessionID string) (int, error)
	Coverage  func(ctx context.Context, baselineID string) ([]model.Coverage, error)
	Entities  func(ctx context.Context, workspaceID, set string) (int, error)
	Close     func() error
}

// oeffneAblage öffnet die Datenbank in dir.
func oeffneAblage(ctx context.Context, dir string) (*ablage, error) {
	db, err := sqlite.Open(ctx, filepath.Join(dir, dateiDB), sqlite.Options{})
	if err != nil {
		return nil, err
	}
	return &ablage{
		Session:  func(ctx context.Context, id string) (model.Session, error) { return db.Sessions().Get(ctx, id) },
		Coverage: func(ctx context.Context, id string) ([]model.Coverage, error) { return db.Coverage().List(ctx, id) },
		Exchanges: func(ctx context.Context, id string) (int, error) {
			ex, err := db.Recording().ListBySession(ctx, id)
			return len(ex), err
		},
		Entities: func(ctx context.Context, workspaceID, set string) (int, error) {
			es, err := db.Entities().ListSet(ctx, workspaceID, set)
			return len(es), err
		},
		Close: db.Close,
	}, nil
}

// aufbauEingabe sind die Werte des Aufbaus. Adresse ist die des Referenzdienstes (leer: ein freier Port von 127.0.0.1); Baue ersetzt
// bootstrap.Build, Oeffne die Datenbank, Umhuellen den Handler des Referenzdienstes (je nil: keine Ersetzung), Dial den Verbindungsaufbau
// des Clients (nil: der des Prozesses); die letzten vier dienen dem Test.
type aufbauEingabe struct {
	Profil    Profil
	Startwert uint64
	Uhr       Uhr
	DataDir   string
	Adresse   string
	Baue      func(ctx context.Context, k bootstrap.Konfiguration) (*handle, error)
	Oeffne    func(ctx context.Context, dir string) (*ablage, error)
	Umhuellen func(http.Handler) http.Handler
	Dial      func(ctx context.Context, netz, adresse string) (net.Conn, error)
}

// aufbauErgebnis nennt, was der Aufbau angelegt hat.
type aufbauErgebnis struct {
	Session   string
	Baseline  string
	Workspace string
	Exchanges int
	Entities  int
}

// aufbauen baut die Fixtures (SPEC-216 Nr. 6 a): Aufnahme der Anfragen des Referenzdienstes über bootstrap.Build im Modus record, Import
// der Session und Workspace aus der Baseline, danach die Zählprüfung; alles innerhalb der Frist des Profils. Jeder Fehler ist ein
// *aufbauFehler, bei Überschreitung der Frist mit dem Grund frist, und liefert keinen Workspace.
func (e aufbauEingabe) aufbauen(ctx context.Context) (aufbauErgebnis, error) {
	if err := datenLeer(e.DataDir); err != nil {
		return aufbauErgebnis{}, &aufbauFehler{grund: grundAufbau, stelle: stelleDaten, ursache: err}
	}
	ctx, cancel := mitFrist(ctx, e.Uhr, e.Profil.FristAufbau)
	defer cancel()
	erg, err := e.ablauf(ctx)
	if err != nil && errors.Is(context.Cause(ctx), errFrist) {
		return aufbauErgebnis{}, &aufbauFehler{grund: grundFrist, stelle: stelleFrist, ursache: err}
	}
	return erg, err
}

// datenLeer meldet einen Fehler, wenn DATA_DIR besteht und kein leeres Verzeichnis ist, auch eine Datei (SPEC-216 Nr. 6 e); ein fehlendes
// legt Build an.
func datenLeer(dir string) error {
	eintraege, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil || len(eintraege) > 0 {
		return errors.Join(errors.New("DATA_DIR nicht leer"), err)
	}
	return nil
}

func (e aufbauEingabe) ablauf(ctx context.Context) (erg aufbauErgebnis, err error) {
	ref, schliesse, err := starteReferenz(ctx, cmp.Or(e.Adresse, "127.0.0.1:0"), e.Profil, e.Startwert, e.Umhuellen)
	if err != nil {
		return erg, an(stelleDienst, err)
	}
	defer schliesse()
	baue := e.Baue
	if baue == nil {
		baue = baueEcht
	}
	k, err := bootstrap.LadeKonfiguration([]string{
		"MODE=record", "LISTEN=127.0.0.1:8080", "DATA_DIR=" + e.DataDir, "SERVICE_ID=" + dienstID,
		"SERVICE_ROOT_PATH=" + wurzel, "UPSTREAM_URL=" + ref.basis + wurzel,
	})
	if err != nil {
		return erg, an(stelleDienst, err)
	}
	k.Listen = "127.0.0.1:0"
	if erg.Session, err = e.nimmAuf(ctx, baue, k); err != nil {
		return aufbauErgebnis{}, an(stelleAufnahme, err)
	}
	if erg.Baseline, erg.Workspace, err = importiere(ctx, baue, k, ref.csdlHash(), erg.Session); err != nil {
		return aufbauErgebnis{}, an(stelleImport, err)
	}
	if erg.Exchanges, erg.Entities, err = e.pruefe(ctx, erg); err != nil {
		return aufbauErgebnis{}, err
	}
	return erg, nil
}

// baueEcht ist bootstrap.Build als handle; ein Fehler liefert kein Handle.
func baueEcht(ctx context.Context, k bootstrap.Konfiguration) (*handle, error) {
	h, err := bootstrap.Build(ctx, k, nil)
	if err != nil {
		return nil, err
	}
	return &handle{
		Hochfahren: h.Hochfahren, Herunterfahren: h.Herunterfahren, Adresse: h.Adresse,
		Sessions: h.Sessions, Import: h.Import, Workspaces: h.Workspaces,
	}, nil
}

// nimmAuf sendet die Anfragen des Aufbaus an ein Handle im Modus record und fährt es herunter, wobei Herunterfahren die Session beendet; es
// liefert die Session. Herunterfahren läuft unter context.WithoutCancel, und sein Fehler tritt zum Fehler des Schritts (SPEC-216 Nr. 6 f).
func (e aufbauEingabe) nimmAuf(ctx context.Context, baue func(context.Context, bootstrap.Konfiguration) (*handle, error),
	k bootstrap.Konfiguration) (session string, err error) {
	h, err := baue(ctx, k)
	if err != nil {
		return "", err
	}
	defer func() { err = errors.Join(err, h.Herunterfahren(context.WithoutCancel(ctx))) }()
	if err = h.Hochfahren(ctx); err != nil {
		return "", err
	}
	if err = e.sende(ctx, "http://"+h.Adresse()); err != nil {
		return "", err
	}
	ss, err := h.Sessions().List(ctx)
	if err != nil || len(ss) != 1 {
		return "", an(stelleSession, errors.Join(fmt.Errorf("Sessions %d, erwartet 1", len(ss)), err))
	}
	return ss[0].ID, nil
}

// anfrage ist ein Schritt des Aufbaus: ein Ziel, bei einer Seite mit Fortsetzung über den `@odata.nextLink`.
type anfrage struct {
	ziel  string
	seite bool
}

// sende sendet die Anfragen von SPEC-216 Nr. 6 d nacheinander auf einer Verbindung an basis: `$metadata`, je Set die Seiten über den
// `@odata.nextLink` bis zur letzten, danach die ersten Paare der Permutation als Lookups.
func (e aufbauEingabe) sende(ctx context.Context, basis string) error {
	c := &http.Client{Transport: &http.Transport{DialContext: e.Dial}}
	defer c.CloseIdleConnections()
	todo := []anfrage{{ziel: wurzel + "/$metadata"}}
	for set := 1; set <= e.Profil.Sets; set++ {
		todo = append(todo, anfrage{ziel: seitenZiel(e.Profil, set, 0).Ziel(), seite: true})
	}
	for _, s := range paare(e.Profil, e.Startwert)[:e.Profil.Lookups] {
		todo = append(todo, anfrage{ziel: lookupZiel(s[0], s[1]).Ziel()})
	}
	for len(todo) > 0 {
		a := todo[0]
		todo = todo[1:]
		body, err := hole(ctx, c, basis+a.ziel)
		weiter := ""
		if err == nil && a.seite {
			weiter, err = folgelink(body)
		}
		if err != nil {
			return err
		}
		if weiter != "" {
			todo = append([]anfrage{{ziel: weiter, seite: true}}, todo...)
		}
	}
	return nil
}

// hole liest ziel mit `Accept: application/json`; jeder Status außer 200 ist ein Fehler an der Stelle status.
func hole(ctx context.Context, c *http.Client, ziel string) ([]byte, error) {
	var resp *http.Response
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ziel, nil)
	if err == nil {
		req.Header.Set(headerAccept, acceptJSON)
		resp, err = c.Do(req)
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, an(stelleStatus, fmt.Errorf("Status %d", resp.StatusCode))
	}
	return body, err
}

// folgelink liefert die Anfragezeile des `@odata.nextLink` der Seite, leer ohne ihn.
func folgelink(body []byte) (string, error) {
	var seite struct {
		Weiter string `json:"@odata.nextLink"`
	}
	err := json.Unmarshal(body, &seite)
	if err != nil || seite.Weiter == "" {
		return "", err
	}
	u, err := url.Parse(seite.Weiter)
	if err != nil {
		return "", err
	}
	return u.RequestURI(), nil
}

// importiere importiert die Session mit dem MetadataHash in einem Handle ohne Hochfahren und legt den Workspace aus der Baseline an; es
// liefert beide Kennungen.
func importiere(ctx context.Context, baue func(context.Context, bootstrap.Konfiguration) (*handle, error),
	k bootstrap.Konfiguration, metadataHash, session string) (baseline, workspace string, err error) {
	h, err := baue(ctx, k)
	if err != nil {
		return "", "", err
	}
	defer func() { err = errors.Join(err, h.Herunterfahren(context.WithoutCancel(ctx))) }()
	erg, err := h.Import().Import(ctx, inbound.ImportAnfrage{
		ImportID: dienstID, ServiceID: dienstID, SourceSessionIDs: []string{session}, MetadataHash: metadataHash,
	})
	if err != nil {
		return "", "", err
	}
	ws, err := h.Workspaces().Create(ctx, erg.Import.BaselineSnapshotID, "")
	return erg.Import.BaselineSnapshotID, ws.ID, err
}

// pruefe liest die Ablage lesend (SPEC-216 Nr. 6 e): die Session ist complete mit so vielen Exchanges, wie das Profil nennt, die Coverage
// der Baseline ist je Set complete, und der Workspace trägt je Set so viele Entities, wie das Profil nennt. Es liefert die Zahlen.
func (e aufbauEingabe) pruefe(ctx context.Context, a aufbauErgebnis) (exchanges, entities int, err error) {
	p := e.Profil
	oeffne := e.Oeffne
	if oeffne == nil {
		oeffne = oeffneAblage
	}
	db, err := oeffne(ctx, e.DataDir)
	if err != nil {
		return 0, 0, an(stelleSession, err)
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	s, err := db.Session(ctx, a.Session)
	if err != nil || s.Status != model.SessionComplete {
		return 0, 0, an(stelleSession, errors.Join(errors.New("Session nicht complete"), err))
	}
	ex, err := db.Exchanges(ctx, a.Session)
	if err != nil || ex != p.Exchanges {
		return 0, 0, an(stelleZahl, errors.Join(fmt.Errorf("Exchanges %d, erwartet %d", ex, p.Exchanges), err))
	}
	cov, err := db.Coverage(ctx, a.Baseline)
	if err != nil || len(cov) != p.Sets || unvollstaendig(cov) {
		return 0, 0, an(stelleCoverage, errors.Join(errors.New("Coverage nicht complete"), err))
	}
	for set := 1; set <= p.Sets; set++ {
		n, err := db.Entities(ctx, a.Workspace, fmt.Sprintf("Set%d", set))
		if err != nil || n != p.Entities {
			return 0, 0, an(stelleEntities, errors.Join(fmt.Errorf("Set%d: Entities %d, erwartet %d", set, n, p.Entities), err))
		}
		entities += n
	}
	return ex, entities, nil
}

// unvollstaendig meldet, ob ein Eintrag der Coverage nicht complete ist.
func unvollstaendig(cov []model.Coverage) bool {
	for _, c := range cov {
		if c.Status != model.CoverageComplete {
			return true
		}
	}
	return false
}
