// Package audit ist der Driven Adapter von support.AuditLog: Datei hängt je Eintrag eine
// JSON-Zeile an die Datei audit.jsonl an (SPEC-235, LH-RB-04, LH-FA-13).
package audit

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// DateiName ist der Name der Ablage im Datenverzeichnis (SPEC-235 Nr. 8).
const DateiName = "audit.jsonl"

// logText ist der feste Text des Log-Eintrags eines gescheiterten Append; Auswertungen suchen
// nach ihm (SPEC-235 Nr. 5).
const logText = "audit: Eintrag nicht geschrieben"

// zeitFormat ist UTC mit genau neun Nachkommastellen (SPEC-235 Nr. 3).
const zeitFormat = "2006-01-02T15:04:05.000000000Z"

var _ support.AuditLog = (*Datei)(nil)

// Config trägt die Ablage und die Quellen von Zeit und Korrelations-ID; alle vier Felder sind
// Pflicht (SPEC-235 Nr. 1).
type Config struct {
	Pfad string
	Uhr  support.Clock
	IDs  support.IDGenerator
	Log  *slog.Logger
}

// Datei legt Einträge als JSON-Zeilen in der Datei Config.Pfad ab. Die Sperre ist der einzige
// Zustand (SPEC-235 Nr. 7).
type Datei struct {
	cfg    Config
	sperre sync.Mutex
}

// New liefert model.ErrUngueltig und keinen Adapter bei einem leeren Pfad und einem nil-Wert in
// Uhr, IDs oder Log. New berührt die Datei nicht und legt kein Verzeichnis an.
func New(c Config) (*Datei, error) {
	if c.Pfad == "" || c.Uhr == nil || c.IDs == nil || c.Log == nil {
		return nil, fmt.Errorf("%w: audit.Config braucht Pfad, Uhr, IDs und Log", model.ErrUngueltig)
	}
	return &Datei{cfg: c}, nil
}

// Append prüft den Eintrag, nimmt die Sperre und hängt die Zeile an; erst nach Sync liefert es
// nil (SPEC-235 Nr. 4). Scheitert es, steht genau ein Eintrag im Log (Nr. 5).
func (d *Datei) Append(e support.Eintrag) error {
	if err := pruefe(e); err != nil {
		d.melde(e, false)
		return err
	}
	if err := d.anhaengen(e); err != nil {
		d.melde(e, true)
		return err
	}
	return nil
}

// anhaengen führt die Schritte b bis i von SPEC-235 Nr. 4 unter der Sperre aus.
func (d *Datei) anhaengen(e support.Eintrag) error {
	d.sperre.Lock()
	defer d.sperre.Unlock()
	jetzt := d.cfg.Uhr.Now().UTC()
	if jetzt.IsZero() || jetzt.Year() < 1 || jetzt.Year() > 9999 {
		return fmt.Errorf("%w: Zeit der Uhr außerhalb der Jahre 1 bis 9999", model.ErrUngueltig)
	}
	korr := e.KorrelationsID
	if korr == "" {
		korr = d.cfg.IDs.KorrelationsID()
	}
	return schreibe(d.cfg.Pfad, zeile(e, jetzt, korr))
}

// schreibe öffnet die Datei zum Anhängen, setzt ein fehlendes Zeilenende der letzten Zeile
// voran, schreibt in einem Aufruf und ruft Sync; ein Fehler des Schließens bleibt ohne Wirkung.
func schreibe(pfad string, z []byte) error {
	f, err := os.OpenFile(pfad, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	halb, err := endetOhneZeilenende(f)
	if err != nil {
		return err
	}
	if halb {
		z = append([]byte{'\n'}, z...)
	}
	n, err := f.Write(z)
	if err != nil {
		return err
	}
	if n < len(z) {
		return io.ErrShortWrite
	}
	return f.Sync()
}

// endetOhneZeilenende meldet, ob die Datei nicht leer ist und ihr letztes Byte nicht LF ist.
func endetOhneZeilenende(f *os.File) (bool, error) {
	info, err := f.Stat()
	if err != nil {
		return false, err
	}
	if info.Size() == 0 {
		return false, nil
	}
	var b [1]byte
	if _, err := f.ReadAt(b[:], info.Size()-1); err != nil {
		return false, err
	}
	return b[0] != '\n', nil
}

// melde schreibt den Log-Eintrag eines gescheiterten Append: nie den Fehlertext, den Pfad oder
// einen Wert eines ungültigen Eintrags (SPEC-235 Nr. 5, 6). Grund und Kennungen stehen nur bei
// einem Eintrag, der die Prüfung bestand.
func (d *Datei) melde(e support.Eintrag, gueltig bool) {
	attrs := []slog.Attr{
		slog.String("ereignis", "audit_fehlgeschlagen"),
		slog.String("error_code", model.CodeE9001.String()),
	}
	if gueltig {
		attrs = append(attrs, slog.String("grund", string(e.Art)))
		for _, kv := range [][2]string{{"service_id", e.ServiceID}, {"session_id", e.SessionID}, {"workspace_id", e.WorkspaceID}} {
			if kv[1] != "" {
				attrs = append(attrs, slog.String(kv[0], kv[1]))
			}
		}
	}
	d.cfg.Log.LogAttrs(context.Background(), slog.LevelError, logText, attrs...)
}

// zeile bildet die Zeile von SPEC-235 Nr. 3: feste Folge der Felder, ohne Leerraum und ohne
// Maskierung, beendet durch LF.
func zeile(e support.Eintrag, t time.Time, korr string) []byte {
	b := append([]byte(`{"timestamp":"`), t.AppendFormat(nil, zeitFormat)...)
	b = append(b, `","correlation_id":"`...)
	b = append(b, korr...)
	b = append(b, `","art":"`...)
	b = append(b, e.Art...)
	b = append(b, '"')
	for _, kv := range [][2]string{
		{"service_id", e.ServiceID}, {"session_id", e.SessionID},
		{"workspace_id", e.WorkspaceID}, {"import_id", e.ImportID},
	} {
		if kv[1] != "" {
			b = append(b, `,"`+kv[0]+`":"`+kv[1]+`"`...)
		}
	}
	if e.ErrorCode != 0 {
		b = append(b, `,"error_code":"`+e.ErrorCode.String()+`"`...)
	}
	return append(b, '}', '\n')
}

// pruefe prüft den Eintrag nach SPEC-234 Nr. 2 und 3; der Text des Fehlers nennt nie einen Wert.
func pruefe(e support.Eintrag) error {
	var pflicht []support.Feld
	bekannt := false
	for _, s := range support.Arten() {
		if s.Art == e.Art {
			pflicht, bekannt = s.Pflicht, true
		}
	}
	if !bekannt {
		return fmt.Errorf("%w: Art außerhalb der Menge", model.ErrUngueltig)
	}
	felder := []struct {
		feld    support.Feld
		gesetzt bool
		kennung string
	}{
		{support.FeldServiceID, e.ServiceID != "", e.ServiceID},
		{support.FeldSessionID, e.SessionID != "", e.SessionID},
		{support.FeldWorkspaceID, e.WorkspaceID != "", e.WorkspaceID},
		{support.FeldImportID, e.ImportID != "", e.ImportID},
		{support.FeldErrorCode, e.ErrorCode != 0, ""},
	}
	for _, f := range felder {
		if f.gesetzt != slices.Contains(pflicht, f.feld) {
			return fmt.Errorf("%w: Feld %s ist für die Art Pflicht oder verboten", model.ErrUngueltig, f.feld)
		}
		if f.kennung != "" && !istKennung(f.kennung) {
			return fmt.Errorf("%w: Feld %s hat nicht die Form einer Kennung", model.ErrUngueltig, f.feld)
		}
	}
	if e.KorrelationsID != "" && !istKennung(e.KorrelationsID) {
		return fmt.Errorf("%w: KorrelationsID hat nicht die Form einer Kennung", model.ErrUngueltig)
	}
	if e.ErrorCode != 0 && !e.ErrorCode.Gueltig() {
		return fmt.Errorf("%w: error_code ist kein Code", model.ErrUngueltig)
	}
	return nil
}

// istKennung prüft 1 bis 64 Bytes aus A-Z, a-z, 0-9, Punkt, Unterstrich und Bindestrich;
// ein Byte ab 0x80 gehört keiner Form an (SPEC-234 Nr. 3).
func istKennung(s string) bool {
	if len(s) < 1 || len(s) > 64 {
		return false
	}
	for i := range len(s) {
		c := s[i]
		ok := c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '.' || c == '_' || c == '-'
		if !ok {
			return false
		}
	}
	return true
}
