package bootstrap_test

import (
	"context"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
)

// etagLiteral ist der ETag zu einer Version als Literal des Vertrags (SPEC-280 Nr. 1).
func etagLiteral(v int) string { return "W/" + strconv.Quote(strconv.Itoa(v)) }

// etagVonAntwort liest die Kopfzeile ETag einer Antwort (leer, wenn sie fehlt).
func etagVonAntwort(t *testing.T, ist abIst) string {
	t.Helper()
	if ist.F != nil || ist.Err != nil {
		return ""
	}
	h := ist.Header["Etag"]
	if len(h) > 1 {
		t.Fatalf("mehrere ETag: %v", h)
	}
	if len(h) == 0 {
		return ""
	}
	return h[0]
}

// etagStand liest den ETag von Products(1) bis Products(5) in dieser Folge (leer für einen Eintrag ohne Antwort).
func etagStand(t *testing.T, m *mock.Mocker, ws string) []string {
	t.Helper()
	var r []string
	for n := 1; n <= 5; n++ {
		r = append(r, etagVonAntwort(t, abSende(context.Background(), m, ws, abG("/Products("+strconv.Itoa(n)+")", ""))))
	}
	return r
}

// TestEtagResetStelltHer (E21, LH-QA-02): hundert Läufe aus Änderung, Löschen, Reset, Lesen: der ETag jedes Eintrags gleicht dem des ersten Lesens
// (Version 0), und der gelöschte Eintrag ist nach dem Reset wieder sichtbar.
func TestEtagResetStelltHer(t *testing.T) {
	u := abFrisch(t)
	m, _ := abEtagMocks(t, u)
	ctx := context.Background()
	erst := etagStand(t, m, "ws-1")
	if want := []string{etagLiteral(0), etagLiteral(0), etagLiteral(0), etagLiteral(0), etagLiteral(0)}; !slices.Equal(erst, want) {
		t.Fatalf("erstes Lesen: %v, erwartet %v", erst, want)
	}
	for i := 0; i < 100; i++ {
		if ist := abSende(ctx, m, "ws-1", abM("PATCH", "/Products(1)", `{"Name":"x"}`)); etagVonAntwort(t, ist) != etagLiteral(1) {
			t.Fatalf("Lauf %d: PATCH %+v", i, ist)
		}
		if ist := abSende(ctx, m, "ws-1", abM("DELETE", "/Products(2)", "")); ist.Status != 204 {
			t.Fatalf("Lauf %d: DELETE %+v", i, ist)
		}
		if ist := abSende(ctx, m, "ws-1", abG("/Products(2)", "")); ist.F == nil {
			t.Fatalf("Lauf %d: Tombstone sichtbar", i)
		}
		if _, err := u.ws.Reset(ctx, "ws-1"); err != nil {
			t.Fatal(err)
		}
		if got := etagStand(t, m, "ws-1"); !slices.Equal(got, erst) {
			t.Fatalf("Lauf %d: ETags nach Reset %v, erstes Lesen %v", i, got, erst)
		}
	}
}

// TestEtagRestoreStelltHer (E22, LH-QA-02): Restore stellt die Version des Snapshots her: ein später geänderter, ein später gelöschter und ein später
// angelegter Eintrag.
func TestEtagRestoreStelltHer(t *testing.T) {
	u := abFrisch(t)
	m, _ := abEtagMocks(t, u)
	ctx := context.Background()
	for _, a := range []abAnfrage{abM("PATCH", "/Products(2)", `{"Name":"a"}`), abM("PATCH", "/Products(2)", `{"Name":"b"}`), abM("PATCH", "/Products(1)", `{"Name":"c"}`)} {
		if ist := abSende(ctx, m, "ws-1", a); ist.Status != 204 {
			t.Fatalf("Aufbau %+v", ist)
		}
	}
	vorher := etagStand(t, m, "ws-1")
	if want := []string{etagLiteral(1), etagLiteral(2), etagLiteral(0), etagLiteral(0), etagLiteral(0)}; !slices.Equal(vorher, want) {
		t.Fatalf("Stand beim Anlegen des Snapshots: %v, erwartet %v", vorher, want)
	}
	sn, err := u.snap.Create(ctx, "ws-1", "s1")
	abMust(t, err)
	for _, a := range []abAnfrage{abM("PATCH", "/Products(1)", `{"Name":"d"}`), abM("DELETE", "/Products(2)", ""),
		abM("POST", "/Products", `{"Id":20,"Name":"neu"}`)} {
		if ist := abSende(ctx, m, "ws-1", a); ist.F != nil || ist.Err != nil {
			t.Fatalf("Änderung nach dem Snapshot %+v: %+v", a, ist)
		}
	}
	if mitte := etagStand(t, m, "ws-1"); slices.Equal(mitte, vorher) {
		t.Fatalf("die Änderungen nach dem Snapshot ändern keinen ETag: %v", mitte)
	}
	if ist := abSende(ctx, m, "ws-1", abG("/Products(20)", "")); etagVonAntwort(t, ist) != etagLiteral(1) {
		t.Fatalf("Products(20) vor Restore: %+v", ist)
	}
	if _, err := u.ws.Restore(ctx, "ws-1", sn.ID); err != nil {
		t.Fatal(err)
	}
	if got := etagStand(t, m, "ws-1"); !slices.Equal(got, vorher) {
		t.Errorf("ETags nach Restore %v, beim Anlegen %v", got, vorher)
	}
	if ist := abSende(ctx, m, "ws-1", abG("/Products(20)", "")); ist.F == nil {
		t.Errorf("der später angelegte Eintrag ist nach Restore sichtbar: %+v", ist)
	}
}

// TestEtagGleichzeitigSQLite (E18, SPEC-280 Nr. 5) über der echten SQLite-Transaktion: n gleichzeitige PATCH derselben Entity (n = 2, 8, 32, je fünf
// frische Aufbauten) geben genau die ETags der Versionen 1 bis n, jeden einmal; der Endstand ist der von n.
func TestEtagGleichzeitigSQLite(t *testing.T) {
	for _, n := range []int{2, 8, 32} {
		for wdh := 0; wdh < 5; wdh++ {
			u := abFrisch(t)
			m, _ := abEtagMocks(t, u)
			var mu sync.Mutex
			var etags []string
			p := &abParallel{N: n}
			p.lauf(func(k int) {
				ist := abSende(context.Background(), m, "ws-1", abM("PATCH", "/Products(1)", `{"Name":"n`+strconv.Itoa(k)+`"}`))
				mu.Lock()
				defer mu.Unlock()
				if ist.Status != 204 {
					t.Errorf("n=%d: PATCH %d: Status %d, Fehler %v", n, k, ist.Status, ist.Err)
					return
				}
				etags = append(etags, etagVonAntwort(t, ist))
			})
			var want []string
			for v := 1; v <= n; v++ {
				want = append(want, etagLiteral(v))
			}
			sort.Strings(etags)
			sort.Strings(want)
			if !slices.Equal(etags, want) {
				t.Fatalf("n=%d Wiederholung %d: ETags %v, erwartet %v", n, wdh, etags, want)
			}
			if got := etagVonAntwort(t, abSende(context.Background(), m, "ws-1", abG("/Products(1)", ""))); got != etagLiteral(n) {
				t.Errorf("n=%d: Endstand %q, erwartet %q", n, got, etagLiteral(n))
			}
		}
	}
}

// TestEtagAcceptEchterSchreiber (E11): mit dem echten Schreiber ändert Accept den ETag nie: der Header gleicht sich über sieben Accept-Werte, das Mitglied
// fehlt allein in der Stufe none, und ohne Mitglied ist der Body der des Dienstes ohne Etag.
func TestEtagAcceptEchterSchreiber(t *testing.T) {
	u := abFrisch(t)
	mit, _ := abEtagMocks(t, u)
	accepts := []string{"", "application/json", "application/json;odata.metadata=minimal", "application/json;odata.metadata=full",
		"application/json;odata.metadata=none", "application/json;IEEE754Compatible=true", "text/html"}
	for _, pfad := range []string{"/Products(3)", "/Products"} {
		var koepfe []string
		for _, acc := range accepts {
			a := abG(pfad, "")
			if acc != "" {
				a = abGH(pfad, "", acc)
			}
			ist := abSende(context.Background(), mit, "ws-1", a)
			ohne := u.sende(a)
			if ist.F != nil || ohne.F != nil {
				// Accept ohne gangbare Form: beide Dienste antworten gleich mit demselben Fehler.
				if ist.F == nil || ohne.F == nil || ist.F.Code != ohne.F.Code {
					t.Errorf("%s Accept %q: Fehler %v, ohne Etag %v", pfad, acc, ist.F, ohne.F)
				}
				continue
			}
			bef, mitglieder := abOhneEtag(ist)
			if bef.Body != ohne.Body {
				t.Errorf("%s Accept %q: Body ohne ETag %q, Dienst ohne Etag %q", pfad, acc, bef.Body, ohne.Body)
			}
			none := strings.Contains(acc, "odata.metadata=none")
			if none && mitglieder != 0 || !none && mitglieder == 0 {
				t.Errorf("%s Accept %q: %d Mitglieder @odata.etag", pfad, acc, mitglieder)
			}
			koepfe = append(koepfe, etagVonAntwort(t, ist))
		}
		if pfad == "/Products(3)" {
			for _, k := range koepfe {
				if k != etagLiteral(0) {
					t.Errorf("%s: ETag %q, erwartet %q", pfad, k, etagLiteral(0))
				}
			}
		}
		if pfad == "/Products" && slices.ContainsFunc(koepfe, func(k string) bool { return k != "" }) {
			t.Errorf("%s: die Collection trägt die Kopfzeile ETag: %v", pfad, koepfe)
		}
	}
}
