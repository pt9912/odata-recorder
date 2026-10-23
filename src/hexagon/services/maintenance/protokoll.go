package maintenance

import (
	"context"
	"log/slog"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// protokolliere schreibt einen Eintrag mit festem Text und den Attributen aus SPEC-123 Nr. 11;
// kein Attribut trägt einen Fehlertext oder den Text eines Befunds.
func (w *Wartung) protokolliere(ctx context.Context, stufe slog.Level, text, ereignis string, c model.Code, extra ...slog.Attr) {
	attrs := append([]slog.Attr{slog.String("ereignis", ereignis), slog.String("error_code", c.String())}, extra...)
	w.k.Log.LogAttrs(ctx, stufe, text, attrs...)
}

func (w *Wartung) protokolliereIntegritaet(ctx context.Context, befunde int) {
	w.protokolliere(ctx, slog.LevelError, "wartung: Integritätsprüfung mit Befund", "integritaet", model.CodeE3001,
		slog.Int("befunde", befunde))
}

func (w *Wartung) protokolliereUebersprungen(ctx context.Context, dienstID, sessionID string) {
	w.protokolliere(ctx, slog.LevelWarn, "wartung: Session als Quelle einer Baseline nicht gelöscht", "aufbewahrung_uebersprungen",
		model.CodeE7004, slog.String("service_id", dienstID), slog.String("session_id", sessionID))
}

// protokolliereBlob schreibt blob_fehlt oder blob_beschaedigt.
func (w *Wartung) protokolliereBlob(ctx context.Context, ereignis, ref string) {
	text := "wartung: referenzierte Blob-Datei fehlt"
	if ereignis == "blob_beschaedigt" {
		text = "wartung: referenzierte Blob-Datei beschädigt"
	}
	w.protokolliere(ctx, slog.LevelError, text, ereignis, model.CodeE3001, slog.String("blob", ref))
}
