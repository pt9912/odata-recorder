package baselineimport_test

import (
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
)

// linkFall ist ein Fall der Auflösung des Links (SPEC-127 Nr. 12 e): die erste Seite nennt link, die
// zweite trägt den Key von GET /Products?<query2>; end ist das Ende der Kette.
type linkFall struct {
	name     string
	upstream string
	link     string
	query2   string
	reject   bool
	end      string
}

const (
	basisURL = "https://up.example/odata/v4"
	folge    = "$skiptoken=1"
)

func linkFaelle() []linkFall {
	var out []linkFall
	pos := func(name, link string) { out = append(out, linkFall{name, basisURL, link, folge, false, "proven"}) }
	neg := func(name, link string) {
		out = append(out, linkFall{name, basisURL, link, folge, false, "link_nicht_aufloesbar"})
	}
	pos("absolut", "https://up.example/odata/v4/Products?$skiptoken=1")
	pos("absolut mit Port 443", "https://up.example:443/odata/v4/Products?$skiptoken=1")
	pos("Host groß", "https://UP.EXAMPLE/odata/v4/Products?$skiptoken=1")
	pos("relativ", "Products?$skiptoken=1")
	pos("pfadabsolut", "/odata/v4/Products?$skiptoken=1")
	pos("nur Query", "?$skiptoken=1")
	pos("Punkt", "./Products?$skiptoken=1")
	pos("Punkt Punkt", "../v4/Products?$skiptoken=1")
	pos("Netzpfad", "//up.example/odata/v4/Products?$skiptoken=1")
	pos("Fragment", "Products?$skiptoken=1#frag")
	out = append(out, linkFall{"%7e gegen %7E", basisURL, "Products?$skiptoken=a%7eb", "$skiptoken=a%7Eb", false, "proven"})
	pos("Escape im Pfad", "Prod%75cts?$skiptoken=1")
	out = append(out, linkFall{"ungültige Kodierung bei literal", basisURL, "Products?a=%zz", "a=%zz", false, "proven"})
	neg("fremder Host", "https://other.example/odata/v4/Products?$skiptoken=1")
	neg("Port 8443", "https://up.example:8443/odata/v4/Products?$skiptoken=1")
	neg("Port 80", "http://up.example/odata/v4/Products?$skiptoken=1")
	neg("Pfad v42", "/odata/v42/Products?$skiptoken=1")
	neg("Pfad ohne v4", "/odata/Products?$skiptoken=1")
	neg("ftp", "ftp://up.example/odata/v4/Products?$skiptoken=1")
	neg("mailto", "mailto:x@y")
	neg("Netzpfad fremd", "//other.example/odata/v4/Products?$skiptoken=1")
	out = append(out, linkFall{"ungültige Kodierung bei reject", basisURL, "Products?a=%zz", "a=%zz", true, "link_nicht_aufloesbar"})
	neg("Steuerzeichen", "Products\x01?$skiptoken=1")
	neg("javascript", "javascript:alert(1)")
	out = append(out, linkFall{"Pfad gleich dem Präfix", basisURL, "/odata/v4?$skiptoken=1", folge, false, "seite_fehlt"})
	const ohnePfad = "https://up.example"
	out = append(out,
		linkFall{"ohne Pfad absolut", ohnePfad, "https://up.example/Products?$skiptoken=1", folge, false, "proven"},
		linkFall{"ohne Pfad relativ", ohnePfad, "Products?$skiptoken=1", folge, false, "proven"},
		linkFall{"ohne Pfad pfadabsolut", ohnePfad, "/Products?$skiptoken=1", folge, false, "proven"},
		linkFall{"ohne Pfad fremd", ohnePfad, "https://other.example/Products?$skiptoken=1", folge, false, "link_nicht_aufloesbar"},
	)
	return out
}

// TestZeileH6AufloesungDesLinks: 29 Formen (13 positive, 11 nicht auflösbare, eine fehlende Seite, vier
// ohne Pfad der Upstream-URL) mit handbestimmten Soll-Keys, dazu elf Formen außerhalb der Menge.
func TestZeileH6AufloesungDesLinks(t *testing.T) {
	faelle := linkFaelle()
	if len(faelle) != 29 {
		t.Fatalf("Menge %d, erwartet 29", len(faelle))
	}
	// Ergänzung außerhalb der 29 Formen: ein maskierter Schrägstrich im Präfix, ein fremdes Schema bei gleichem
	// Port, Punktsegmente eines absoluten Links, ein Steuerzeichen in der Query und ein Link, der nur ein Fragment
	// nennt (die Seite selbst, mit ihrer Query: eine Schleife).
	faelle = append(faelle,
		linkFall{"Escape des Schrägstrichs im Präfix", basisURL, "/odata/v4%2FProducts?$skiptoken=1", folge, false, "link_nicht_aufloesbar"},
		linkFall{"ftp mit Port 443", basisURL, "ftp://up.example:443/odata/v4/Products?$skiptoken=1", folge, false, "link_nicht_aufloesbar"},
		linkFall{"absolut mit Punktsegmenten", basisURL, "https://up.example/odata/v4/x/../Products?$skiptoken=1", folge, false, "proven"},
		linkFall{"Steuerzeichen in der Query", basisURL, "Products?$skiptoken=1\x01", folge, false, "link_nicht_aufloesbar"},
		linkFall{"nur ein Fragment", basisURL, "#frag", folge, false, "schleife"},
		linkFall{"zwei Ebenen höher", basisURL, "../../v4/Products?$skiptoken=1", folge, false, "link_nicht_aufloesbar"},
		linkFall{"Port nicht numerisch", basisURL, "https://up.example:abc/odata/v4/Products?$skiptoken=1", folge, false, "link_nicht_aufloesbar"},
		linkFall{"Schema ungültig", basisURL, "1x://up.example/odata/v4/Products?$skiptoken=1", folge, false, "link_nicht_aufloesbar"},
		linkFall{"Punkt am Ende", basisURL, "/odata/v4/.", folge, false, "seite_fehlt"},
		linkFall{"IPv6 absolut", "http://[::1]:8080/odata", "http://[::1]:8080/odata/Products?$skiptoken=1", folge, false, "proven"},
		linkFall{"IPv6 fremder Port", "http://[::1]:8080/odata", "http://[::1]:8081/odata/Products?$skiptoken=1", folge, false, "link_nicht_aufloesbar"},
	)
	for _, f := range faelle {
		startQuery := map[string]string{"nur ein Fragment": "$format=json"}[f.name]
		w := neu(t)
		w.konfig.Dienste = []baselineimport.Dienst{{ID: dienstID, UpstreamURL: f.upstream}}
		if f.reject {
			w.konfig.Regeln.InvalidEncoding = model.KodierungReject
		}
		b := w.standard()
		b.collection("Products", startQuery, mitLink(seite(produkt(1, "A", "1")), f.link))
		b.collection("Products", f.query2, seite(produkt(2, "A", "1")))
		mustImport(t, w, anfrage())
		status, grund, ketten := ketteVon(t, w)
		if endeVon(ketten[0]) != f.end {
			t.Errorf("%s: Ende %s, erwartet %s", f.name, endeVon(ketten[0]), f.end)
		}
		if (f.end == "proven") != (status == "complete") || (f.end == "proven") != (grund == "kette_belegt") {
			t.Errorf("%s: %s/%s", f.name, status, grund)
		}
	}
}

// TestZeileH6UpstreamMitSchraegstrich: ein abschließender Schrägstrich der Upstream-URL entfällt
// (SPEC-127 Nr. 1), und mit ihm gelingt der Präfix des Pfades.
func TestZeileH6UpstreamMitSchraegstrich(t *testing.T) {
	for _, up := range []string{"https://up.example/odata/v4/", "https://up.example/odata/v4"} {
		w := neu(t)
		w.konfig.Dienste = []baselineimport.Dienst{{ID: dienstID, UpstreamURL: up}}
		b := w.standard()
		b.collection("Products", "", mitLink(seite(produkt(1, "A", "1")), "https://up.example/odata/v4/Products?$skiptoken=1"))
		b.collection("Products", folge, seite(produkt(2, "A", "1")))
		mustImport(t, w, anfrage())
		if _, _, k := ketteVon(t, w); endeVon(k[0]) != "proven" {
			t.Errorf("%s: %v", up, k)
		}
	}
}
