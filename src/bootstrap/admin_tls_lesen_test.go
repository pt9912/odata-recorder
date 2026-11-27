package bootstrap_test

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/bootstrap"
)

// Die Zeile eines ungültigen Flags als Literal (SPEC-299 Nr. 5, SPEC-110 Nr. 3).
const zeileFlag = "ODR-E2001 ADMIN_ALLOW_PLAINTEXT: Wahrheitswert muss true oder false sein"

// pruefProtokoll hält Eingabe, Zeit und Weg jedes Aufrufs der Prüfung an seiner Stelle fest; mit echt ruft es
// pruefeAdminTransport, ohne echt meldet es keinen Verstoß und keinen Weg.
type pruefProtokoll struct {
	echt     bool
	eingaben []bootstrap.AdminEingabe
	zeiten   []time.Time
	wege     []bootstrap.AdminWeg
}

func (p *pruefProtokoll) pruefe(e bootstrap.AdminEingabe, jetzt time.Time) (bootstrap.AdminTransport, []bootstrap.TLSVerstoss) {
	var t bootstrap.AdminTransport
	var vs []bootstrap.TLSVerstoss
	if p.echt {
		t, vs = bootstrap.PruefeAdminTransport(e, jetzt)
	}
	p.eingaben, p.zeiten, p.wege = append(p.eingaben, e), append(p.zeiten, jetzt), append(p.wege, t.Weg)
	return t, vs
}

// konfigZeilen sind die Zeilen eines *KonfigFehler, nil ohne Fehler.
func konfigZeilen(t *testing.T, err error) []string {
	t.Helper()
	if err == nil {
		return nil
	}
	kf, ok := err.(*bootstrap.KonfigFehler)
	if !ok {
		t.Fatalf("Fehlertyp %T: %v", err, err)
	}
	return kf.Zeilen
}

// buildMit baut k mit p an der Stelle Build, fährt den Handle sofort herunter und liefert seinen Transport und die Zeilen.
func buildMit(t *testing.T, p *pruefProtokoll, k bootstrap.Konfiguration) (bootstrap.AdminTransport, []string) {
	t.Helper()
	b := bootstrap.StandardBauteile()
	bootstrap.SetzePruefung(&b, p.pruefe)
	h, err := bootstrap.BuildMit(context.Background(), k, nil, b)
	if err != nil {
		return bootstrap.AdminTransport{}, konfigZeilen(t, err)
	}
	defer func() { _ = h.Herunterfahren(context.Background()) }()
	return bootstrap.AdminTransportVon(h), nil
}

// TestAdminEingabe (C01): Feld {Listen, CertPfad, KeyPfad, Klartext} × Zustand {gesetzt mit Leerraum am Rand, nur Leerraum, leer,
// ungesetzt} × 4 Modi = 64 Zellen. Je Zelle geht dieselbe Umgebung durch die Naht am Laden und die aus ihr geladene Konfiguration
// durch die Naht in Build; die drei anderen Felder tragen eigene gesetzte Werte. Orakel ist die ganze Eingabe als Literal, am Laden
// die übergebene Zeit des Starts, in Build eine Zeit der Uhr während des Aufrufs. Die aufzeichnende Prüfung meldet keinen Verstoß,
// damit jede Umgebung eine Konfiguration ergibt.
func TestAdminEingabe(t *testing.T) {
	const listen, cert, key = "10.1.2.3:4711", "/pfad/cert-4711.pem", "/pfad/key-4712.pem"
	start := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	namen := []string{"ADMIN_LISTEN", "ADMIN_TLS_CERT", "ADMIN_TLS_KEY", "ADMIN_ALLOW_PLAINTEXT"}
	werte := []string{listen, cert, key, "true"}
	// leer setzt den Wert der Eingabe eines ungesetzten Feldes; ADMIN_LISTEN trägt dann den Default.
	leer := []func(*bootstrap.AdminEingabe){
		func(e *bootstrap.AdminEingabe) { e.Listen = "127.0.0.1:8081" }, func(e *bootstrap.AdminEingabe) { e.CertPfad = "" },
		func(e *bootstrap.AdminEingabe) { e.KeyPfad = "" }, func(e *bootstrap.AdminEingabe) { e.Klartext = false },
	}
	zellen := 0
	for _, modus := range vierModi() {
		for i, name := range namen {
			for _, zustand := range []string{"gesetzt", " \t\r\n", "", "ungesetzt"} {
				zellen++
				var env []string
				for j := range namen {
					if j != i {
						env = append(env, namen[j]+"="+werte[j])
					}
				}
				soll := bootstrap.AdminEingabe{Listen: listen, CertPfad: cert, KeyPfad: key, Klartext: true, FormGueltig: true}
				if zustand == "gesetzt" {
					env = append(env, name+"= \t"+werte[i]+"\r\n")
				} else {
					leer[i](&soll)
					if zustand != "ungesetzt" {
						env = append(env, name+"="+zustand)
					}
				}
				laden, bauen := &pruefProtokoll{}, &pruefProtokoll{}
				k, err := bootstrap.LadeKonfigurationMitPruefung(adminEnv(modus, filepath.Join(t.TempDir(), "d"), env...), start, laden.pruefe)
				vorher := time.Now()
				_, zeilen := buildMit(t, bauen, k)
				if err != nil || zeilen != nil {
					t.Fatalf("%s %s %q: %v, Build %q", modus, name, zustand, err, zeilen)
				}
				if !slices.Equal(laden.eingaben, []bootstrap.AdminEingabe{soll}) || !laden.zeiten[0].Equal(start) {
					t.Errorf("%s %s %q, Laden: %+v zu %v, erwartet %+v", modus, name, zustand, laden.eingaben, laden.zeiten, soll)
				}
				if !slices.Equal(bauen.eingaben, []bootstrap.AdminEingabe{soll}) ||
					bauen.zeiten[0].Before(vorher.Add(-time.Second)) || bauen.zeiten[0].After(time.Now().Add(time.Second)) {
					t.Errorf("%s %s %q, Build: %+v zu %v, erwartet %+v", modus, name, zustand, bauen.eingaben, bauen.zeiten, soll)
				}
			}
		}
	}
	if zellen != 64 {
		t.Errorf("%d Zellen, erwartet 64", zellen)
	}
}

// TestAdminFlag (C02, FormGueltig des Flags): M-Flag, 11 Werte von ADMIN_ALLOW_PLAINTEXT × 4 Modi = 44 Zellen, mit ADMIN_LISTEN
// außerhalb von Loopback ohne Paar. Wahr: Weg Klartext ohne Zeile; falsch: die Zeile des Zusammenhangs; ungültig: allein die Zeile
// am Flag.
func TestAdminFlag(t *testing.T) {
	const wahr, falsch, ungueltig = "wahr", "falsch", "ungültig"
	werte := []struct {
		env []string
		art string
	}{
		{[]string{"ADMIN_ALLOW_PLAINTEXT=true"}, wahr}, {[]string{"ADMIN_ALLOW_PLAINTEXT= \ttrue\r\n"}, wahr},
		{[]string{"ADMIN_ALLOW_PLAINTEXT=false"}, falsch}, {nil, falsch}, {[]string{"ADMIN_ALLOW_PLAINTEXT="}, falsch},
		{[]string{"ADMIN_ALLOW_PLAINTEXT= \t "}, falsch},
		{[]string{"ADMIN_ALLOW_PLAINTEXT=TRUE"}, ungueltig}, {[]string{"ADMIN_ALLOW_PLAINTEXT=True"}, ungueltig},
		{[]string{"ADMIN_ALLOW_PLAINTEXT=1"}, ungueltig}, {[]string{"ADMIN_ALLOW_PLAINTEXT=yes"}, ungueltig},
		{[]string{"ADMIN_ALLOW_PLAINTEXT=on"}, ungueltig},
	}
	zellen := 0
	for _, modus := range vierModi() {
		for _, w := range werte {
			zellen++
			p := &pruefProtokoll{echt: true}
			env := adminEnv(modus, filepath.Join(t.TempDir(), "d"), append([]string{"ADMIN_LISTEN=0.0.0.0:8081"}, w.env...)...)
			k, err := bootstrap.LadeKonfigurationMitPruefung(env, time.Now(), p.pruefe)
			soll, weg := []string{zeileZusammenhang}, bootstrap.WegKeiner
			switch w.art {
			case wahr:
				soll, weg = nil, bootstrap.WegKlartext
			case ungueltig:
				soll = []string{zeileFlag}
			}
			if zeilen := konfigZeilen(t, err); !slices.Equal(zeilen, soll) || !slices.Equal(p.wege, []bootstrap.AdminWeg{weg}) ||
				k.AdminAllowPlaintext != (w.art == wahr) {
				t.Errorf("%s %q: Zeilen %q, Wege %v, Feld %v; erwartet %q, Weg %v", modus, w.env, zeilen, p.wege, k.AdminAllowPlaintext, soll, weg)
			}
		}
	}
	if zellen != 44 || len(werte) != 11 {
		t.Errorf("%d Zellen aus %d Werten, erwartet 44 aus 11", zellen, len(werte))
	}
}

// TestAdminNamenAusKonfig (C08, SPEC-299 Nr. 1): kein Zeichenketten-Literal von admintls.go enthält einen der drei Namen; Kommentare
// zählen nicht. Dass die Namen als Literal in einer Datei konfig*.go stehen, hält `make katalog-check`.
func TestAdminNamenAusKonfig(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "admintls.go", nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	literale := 0
	ast.Inspect(f, func(n ast.Node) bool {
		if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			literale++
			for _, name := range []string{"ADMIN_TLS_CERT", "ADMIN_TLS_KEY", "ADMIN_ALLOW_PLAINTEXT"} {
				if strings.Contains(lit.Value, name) {
					t.Errorf("admintls.go führt den Namen %s als eigenes Literal %s", name, lit.Value)
				}
			}
		}
		return true
	})
	if literale < 5 {
		t.Errorf("nur %d Literale in admintls.go gelesen", literale)
	}
}

// TestAdminLadenNebenFrueheren (SPEC-299 Nr. 6, 8 a, SPEC-110 Nr. 8, 12): ein Verstoß jeder Stufe, die das Laden vor der Prüfung des
// Transports meldet (12), × Verstoß des Transports {Zusammenhang, Paar fehlt, Flag ungültig} × 4 Modi = 144 Zellen: eine Meldung
// trägt die Schlüssel beider, nach Schlüssel sortiert, und die Zeilen des Transports.
func TestAdminLadenNebenFrueheren(t *testing.T) {
	fehlt := filepath.Join(t.TempDir(), "fehlt.pem")
	frueher := []struct{ env, schluessel []string }{
		{[]string{"MODE=Record"}, []string{"MODE"}}, {[]string{"LISTEN=x"}, []string{"LISTEN"}},
		{[]string{"ADMIN_X=1"}, []string{"ADMIN_X"}}, {[]string{"REDACT_HEADERS=bad name"}, []string{"REDACT_HEADERS"}},
		{[]string{"UPSTREAM_TLS_VERIFY=yes"}, []string{"UPSTREAM_TLS_VERIFY"}}, {[]string{"LISTEN_TLS_X=1"}, []string{"LISTEN_TLS_X"}},
		{[]string{"HYBRID_FALLBACK=x"}, []string{"HYBRID_FALLBACK"}}, {[]string{"SERVICE_ID=a b"}, []string{"SERVICE_ID"}},
		{[]string{"MODE=hybrid", "SERVICE_ODATA_VERSION=2.0"}, []string{"MODE"}},
		{[]string{"PUBLIC_BASE_URL=ftp://h"}, []string{"PUBLIC_BASE_URL"}},
		{[]string{"UPSTREAM_CA_FILE=" + fehlt}, []string{"UPSTREAM_CA_FILE"}},
		{[]string{"LISTEN_TLS_CERT=" + fehlt, "LISTEN_TLS_KEY=" + fehlt}, []string{"LISTEN_TLS_CERT", "LISTEN_TLS_KEY"}},
	}
	transport := []struct{ env, zeilen []string }{
		{[]string{"ADMIN_LISTEN=0.0.0.0:8081"}, []string{zeileZusammenhang}},
		{[]string{"ADMIN_TLS_CERT=" + fehlt, "ADMIN_TLS_KEY=" + fehlt}, []string{zeileAdminCertFehlt, zeileAdminKeyFehlt}},
		{[]string{"ADMIN_ALLOW_PLAINTEXT=yes"}, []string{zeileFlag}},
	}
	zellen := 0
	for _, modus := range vierModi() {
		for _, f := range frueher {
			for _, tr := range transport {
				zellen++
				_, zeilen := lade(t, adminEnv(modus, filepath.Join(t.TempDir(), "d"), slices.Concat(f.env, tr.env)...)...)
				soll := slices.Concat(f.schluessel, schluessel(t, tr.zeilen))
				slices.Sort(soll)
				fehlen := slices.ContainsFunc(tr.zeilen, func(z string) bool { return !slices.Contains(zeilen, z) })
				if !slices.Equal(schluessel(t, zeilen), soll) || fehlen {
					t.Errorf("%s %q %q: %q, erwartet die Schlüssel %q und %q", modus, f.env, tr.env, zeilen, soll, tr.zeilen)
				}
			}
		}
	}
	if zellen != 144 {
		t.Errorf("%d Zellen, erwartet 144", zellen)
	}
}

// TestAdminKeineAusgabeAmStrom (SPEC-299 Nr. 12): keine Datei, die die Pfade und das Flag des Transports trägt (6 Dateien), ruft print
// oder println, nennt os.NewFile oder importiert syscall; so schreibt keine an os.Stdout und os.Stderr vorbei auf die Deskriptoren.
func TestAdminKeineAusgabeAmStrom(t *testing.T) {
	for _, datei := range []string{"admintls.go", "konfig.go", "konfig_admin.go", "wurzel.go", "lebenszyklus.go", "start.go"} {
		f, err := parser.ParseFile(token.NewFileSet(), datei, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}
		for _, i := range f.Imports {
			if i.Path.Value == `"syscall"` {
				t.Errorf("%s importiert syscall", datei)
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.CallExpr:
				if id, ok := x.Fun.(*ast.Ident); ok && (id.Name == "print" || id.Name == "println") {
					t.Errorf("%s ruft %s", datei, id.Name)
				}
			case *ast.SelectorExpr:
				if id, ok := x.X.(*ast.Ident); ok && id.Name == "os" && x.Sel.Name == "NewFile" {
					t.Errorf("%s nennt os.NewFile", datei)
				}
			}
			return true
		})
	}
}
