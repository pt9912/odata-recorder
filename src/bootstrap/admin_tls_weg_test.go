package bootstrap_test

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/tlstest"
)

// TestAdminUmhuellung (C05): 3 Wege {TLS mit gültigem Paar, Loopback, Klartext} × 4 Modi = 12 Zellen an Build und Hochfahren. Nur der
// Weg TLS umhüllt; die Dateien des Paars sind nach Build gelöscht, kein späterer Aufruf lädt sie. Jeder Bind geht auf 127.0.0.1.
func TestAdminUmhuellung(t *testing.T) {
	w := neueWerkstatt(t)
	zellen := 0
	for _, modus := range vierModi() {
		for _, weg := range []bootstrap.AdminWeg{bootstrap.WegTLS, bootstrap.WegLoopback, bootstrap.WegKlartext} {
			zellen++
			k := listenerKonfig(t.TempDir(), modus)
			switch weg {
			case bootstrap.WegTLS:
				k.AdminTLSCert, k.AdminTLSKey, _ = w.tlsPaar(w.wurzel, tlstest.Optionen{})
			case bootstrap.WegKlartext:
				k.AdminListen, k.AdminAllowPlaintext = "0.0.0.0:0", true
			}
			b := bootstrap.StandardBauteile()
			echt := b.Listen
			b.Listen = func(ctx context.Context, network, _ string) (net.Listener, error) {
				return echt(ctx, network, "127.0.0.1:0")
			}
			h, err := bootstrap.BuildMit(context.Background(), k, nil, b)
			if err != nil {
				t.Fatalf("%s Weg %d: %v", modus, weg, err)
			}
			t.Cleanup(func() { _ = h.Herunterfahren(context.Background()) })
			_, _ = os.Remove(k.AdminTLSCert), os.Remove(k.AdminTLSKey)
			if err := h.Hochfahren(context.Background()); err != nil {
				t.Fatalf("%s Weg %d: %v", modus, weg, err)
			}
			pruefeTransportAm(t, fmt.Sprintf("%s Weg %d", modus, weg), h.AdminAdresse(), weg == bootstrap.WegTLS, w.wurzel)
		}
	}
	if zellen != 12 {
		t.Errorf("%d Zellen, erwartet 12", zellen)
	}
}

// wegBild ist die Form eines Bildes: Eingabe und Weg der Prüfung, Weg und TLS des Handle, die Zeilen und die Warnung.
const wegBild = "Eingaben %+v, geprüft %v, Weg %d, TLS %v, Zeilen %q, Warnung %q"

// bildWie baut k mit der echten Prüfung und liefert sein Bild.
func bildWie(t *testing.T, k bootstrap.Konfiguration) string {
	t.Helper()
	p := &pruefProtokoll{echt: true}
	tr, zeilen := buildMit(t, p, k)
	var warnung strings.Builder
	bootstrap.AdminWarne(tr, slog.New(slog.NewTextHandler(&warnung, nil)))
	_, rest, _ := strings.Cut(strings.TrimSpace(warnung.String()), " ")
	return fmt.Sprintf(wegBild, p.eingaben, p.wege, tr.Weg, tr.TLS != nil, zeilen, rest)
}

// TestAdminWegJeModus (C07): M-Weg, 8 Hosts × Paar × Flag = 32 Zellen von SPEC-299 Nr. 3, × 4 Modi = 128 Zellen an der Naht in Build.
// Orakel je Zelle für jeden Modus dasselbe Bild nach der Tabelle: ohne Weg die Zeile des Zusammenhangs, im Weg Klartext die Warnung.
func TestAdminWegJeModus(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key := w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut))
	hosts := []struct {
		listen   string
		loopback bool
	}{
		{"127.0.0.1:8081", true}, {"[::1]:8081", true}, {"localhost:8081", true},
		{"0.0.0.0:8081", false}, {"[::]:8081", false}, {":8081", false}, {"recorder.internal:8081", false}, {"10.0.0.1:8081", false},
	}
	zellen := 0
	for _, host := range hosts {
		for _, paar := range []bool{false, true} {
			for _, flag := range []bool{false, true} {
				e := bootstrap.AdminEingabe{Listen: host.listen, Klartext: flag, FormGueltig: true}
				if paar {
					e.CertPfad, e.KeyPfad = cert, key
				}
				weg, zeilen, warnung := bootstrap.WegKeiner, []string{zeileZusammenhang}, ""
				switch {
				case paar:
					weg, zeilen = bootstrap.WegTLS, nil
				case host.loopback:
					weg, zeilen = bootstrap.WegLoopback, nil
				case flag:
					weg, zeilen, warnung = bootstrap.WegKlartext, nil, zeileWarnung
				}
				soll := fmt.Sprintf(wegBild, []bootstrap.AdminEingabe{e}, []bootstrap.AdminWeg{weg}, weg, paar, zeilen, warnung)
				for _, modus := range vierModi() {
					zellen++
					k := listenerKonfig(t.TempDir(), modus)
					k.AdminListen, k.AdminTLSCert, k.AdminTLSKey, k.AdminAllowPlaintext = host.listen, e.CertPfad, e.KeyPfad, flag
					if bild := bildWie(t, k); bild != soll {
						t.Errorf("%s Paar %v Flag %v, %s: %s, erwartet %s", host.listen, paar, flag, modus, bild, soll)
					}
				}
			}
		}
	}
	if zellen != 128 {
		t.Errorf("%d Zellen, erwartet 128", zellen)
	}
}
