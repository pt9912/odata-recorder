package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// TestTextRoundTrip zeichnet Text-Antworten auf und spielt sie ab (LH-FA-04, LH-FA-05, LH-FA-14,
// SPEC-045): `/Products/$count` dreimal in derselben Folge, `/Products(1)/Name/$value` ohne Treffer
// des Redaktors und `/Products(1)/Password/$value` mit Treffer. Der Treffer wird nur bei
// REPLAY_REDACTED_BODIES=true abgespielt, als Marker; das Geheimnis erreicht weder einen Blob noch eine
// andere Datei unter DATA_DIR. Der Referenzdienst ist beim Abspielen beendet.
func TestTextRoundTrip(t *testing.T) {
	for _, option := range []string{"true", "false"} {
		t.Run("REPLAY_REDACTED_BODIES="+option, func(t *testing.T) { textRoundTrip(t, option == "true") })
	}
}

func textRoundTrip(t *testing.T, mitRedigiertem bool) {
	ctx := context.Background()
	ref := neuerReferenzdienst(t)
	dir := t.TempDir()
	kRec := konfig(t, "record", dir, ref.URL()+"/odata")
	kRep := konfig(t, "replay", dir, ref.URL()+"/odata", fmt.Sprintf("REPLAY_REDACTED_BODIES=%v", mitRedigiertem))
	rec := hochgefahren(ctx, t, kRec)

	sessions, err := rec.Sessions().List(ctx)
	if err != nil || len(sessions) != 1 {
		t.Fatalf("erste Session: %+v, %v", sessions, err)
	}
	sessionID := sessions[0].ID
	anfragen := []anfrage{
		{"/odata/$metadata", acceptXML},
		{"/odata/Products", acceptJSON},
		{"/odata/Products(1)/Name/$value", acceptText},
		{"/odata/Products(1)/Password/$value", acceptText},
		{"/odata/Products/$count", acceptText},
		{"/odata/Products/$count", acceptText},
		{"/odata/Products/$count", acceptText},
		{"/odata/Products(3)/Name/$value", acceptText},
	}
	rc := neuerClient(t, rec.Adresse())
	var aufgenommen []antwort
	for _, a := range anfragen {
		aufgenommen = append(aufgenommen, rc.hole(a, nil))
	}
	// Der Client sieht den Wert, wie der Upstream ihn sendet; der Referenzdienst kennt Produkt 3 nicht.
	for i, soll := range []struct {
		status int
		typ    string
		body   string
	}{
		{200, "application/xml", ""}, {200, "application/json;odata.metadata=minimal", ""},
		{200, "text/plain", "Anvil"}, {200, "text/plain", geheimC},
		{200, "text/plain", "7"}, {200, "text/plain", "11"}, {200, "text/plain", "13"},
		{404, "application/json", ""},
	} {
		w := aufgenommen[i]
		if w.status != soll.status || w.header.Get("Content-Type") != soll.typ || (soll.body != "" && string(w.body) != soll.body) {
			t.Fatalf("Aufnahme %d (%s): %d %q %q", i, anfragen[i].pfad, w.status, w.header.Get("Content-Type"), w.body)
		}
	}

	beendet, err := rec.Sessions().Beenden(ctx, sessionID)
	if err != nil || beendet.Status != model.SessionComplete {
		t.Fatalf("Beenden: %+v, %v", beendet, err)
	}
	erg, err := rec.Import().Import(ctx, inbound.ImportAnfrage{
		ImportID: "text-import", ServiceID: serviceID, SourceSessionIDs: []string{sessionID}, MetadataHash: csdlHash(),
	})
	if err != nil || !erg.Neu {
		t.Fatalf("Import: %+v, %v", erg, err)
	}
	ws, err := rec.Workspaces().Create(ctx, erg.Import.BaselineSnapshotID, "")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := rec.Herunterfahren(ctx); err != nil {
		t.Fatalf("Herunterfahren(record): %v", err)
	}

	// Kein Geheimnis in einer Datei; der Wert ohne Treffer steht im Blob, der Marker für den Treffer ebenso.
	_, funde := durchsuche(t, dir, []string{geheimA, geheimB, geheimC, "Anvil", model.Marker})
	gefunden := map[string]bool{}
	for _, f := range funde {
		switch {
		case f.geheimnis == geheimA || f.geheimnis == geheimB || f.geheimnis == geheimC:
			t.Errorf("Geheimnis %s in %s", f.geheimnis, f.datei)
		case strings.HasPrefix(f.datei, "blobs"+string(filepath.Separator)):
			gefunden[f.geheimnis] = true
		}
	}
	if !gefunden["Anvil"] || !gefunden[model.Marker] {
		t.Errorf("Blobs tragen Anvil %v und den Marker %v", gefunden["Anvil"], gefunden[model.Marker])
	}

	rep := hochgefahren(ctx, t, kRep)
	ref.srv.Close()
	pc := neuerClient(t, rep.Adresse())
	z := map[string]string{
		"X-Recorder-Workspace": ws.ID, "X-Recorder-Recording-Session": sessionID, "X-Recorder-Replay-Session": "rep-text",
	}
	for i, a := range anfragen {
		w := pc.hole(a, z)
		if strings.HasSuffix(a.pfad, "/Password/$value") {
			pruefeTreffer(t, w, mitRedigiertem)
			continue
		}
		if diff := gleich(kRep, aufgenommen[i], w); diff != "" {
			t.Errorf("Wiedergabe %d (%s) weicht von der Aufnahme ab: %s", i, a.pfad, diff)
		}
	}
	if err := rep.Herunterfahren(ctx); err != nil {
		t.Fatalf("Herunterfahren(replay): %v", err)
	}
	if _, funde := durchsuche(t, dir, []string{geheimC}); len(funde) != 0 {
		t.Errorf("Geheimnis nach dem Replay in %v", funde)
	}
}

// pruefeTreffer: der redigierte Wert wird bei true als Marker mit dem aufgezeichneten Typ ausgeliefert, bei
// false nicht (ODR-E5008, SPEC-045); das Geheimnis steht in keinem von beiden.
func pruefeTreffer(t *testing.T, w antwort, mitRedigiertem bool) {
	t.Helper()
	if bytes.Contains(w.body, []byte(geheimC)) {
		t.Errorf("Replay liefert das Geheimnis aus: %s", w.body)
	}
	if mitRedigiertem {
		if w.status != 200 || string(w.body) != model.Marker || w.header.Get("Content-Type") != "text/plain" {
			t.Errorf("Treffer mit REPLAY_REDACTED_BODIES=true: %d %q als %q", w.status, w.body, w.header.Get("Content-Type"))
		}
		return
	}
	if w.status == 200 || !bytes.Contains(w.body, []byte("ODR-E5008")) {
		t.Errorf("Treffer mit REPLAY_REDACTED_BODIES=false: %d %s", w.status, w.body)
	}
}
