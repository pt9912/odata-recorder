package bootstrap

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"unicode/utf8"
)

// maxServicesFile ist die größte lesbare SERVICES_FILE (SPEC-110 Nr. 9).
const maxServicesFile = 1 << 20

// felderDienst sind die Felder eines Eintrags der SERVICES_FILE.
func felderDienst() []string { return []string{"id", "service_root", "upstream_url"} }

// optionalerWahrheitswert liest einen Wahrheitswert und meldet, ob er gesetzt ist.
func (p *pruefer) optionalerWahrheitswert(schluessel string) (wert, gesetzt bool) {
	if _, ok := p.wert(schluessel); !ok {
		return false, false
	}
	return p.wahrheitswert(schluessel, false), true
}

// liesTLS prüft die Widersprüche der Zertifikatsprüfung (SPEC-110 Nr. 10).
func (p *pruefer) liesTLS(k *Konfiguration) {
	verify, verifyGesetzt := p.optionalerWahrheitswert("UPSTREAM_TLS_VERIFY")
	skip := p.wahrheitswert("UPSTREAM_TLS_INSECURE_SKIP_VERIFY", false)
	switch {
	case verifyGesetzt && !verify && !skip:
		p.melde("UPSTREAM_TLS_VERIFY", "false braucht UPSTREAM_TLS_INSECURE_SKIP_VERIFY=true")
	case verifyGesetzt && verify && skip:
		p.melde("UPSTREAM_TLS_INSECURE_SKIP_VERIFY", "widerspricht UPSTREAM_TLS_VERIFY=true")
	}
	k.UpstreamTLSVerify = !skip && (!verifyGesetzt || verify)
}

// pruefeCADatei verlangt eine lesbare Datei mit mindestens einem PEM-Zertifikat (SPEC-110 Nr. 11).
func (p *pruefer) pruefeCADatei(k *Konfiguration) {
	if k.UpstreamCAFile == "" {
		return
	}
	roh, err := os.ReadFile(k.UpstreamCAFile)
	switch {
	case err != nil:
		p.melde("UPSTREAM_CA_FILE", "Datei nicht lesbar")
	case !x509.NewCertPool().AppendCertsFromPEM(roh):
		p.melde("UPSTREAM_CA_FILE", "Datei enthält kein PEM-Zertifikat")
	}
}

// liesDienste bildet die Services aus SERVICES_FILE oder den Einzelschlüsseln (SPEC-110 Nr. 9).
func (p *pruefer) liesDienste(k *Konfiguration) {
	record := k.Modus == "record"
	datei, hatDatei := p.wert("SERVICES_FILE")
	if hatDatei {
		for _, schluessel := range []string{"SERVICE_ID", "SERVICE_ROOT_PATH", "UPSTREAM_URL"} {
			if _, gesetzt := p.wert(schluessel); gesetzt {
				p.melde(schluessel, "zugleich mit SERVICES_FILE gesetzt")
			}
		}
		k.Dienste = p.dateiDienste(datei, record)
		p.pruefeDienste(k.Dienste)
		return
	}
	k.Dienste = []Dienst{p.einzelDienst(record)}
}

func (p *pruefer) einzelDienst(record bool) Dienst {
	d := Dienst{ID: "default", Root: "/"}
	if w, ok := p.wert("SERVICE_ID"); ok {
		if !istDienstID(w) {
			p.melde("SERVICE_ID", "1 bis 64 Zeichen aus Buchstaben, Ziffern, ., _ und -")
		} else {
			d.ID = w
		}
	}
	if w, ok := p.wert("SERVICE_ROOT_PATH"); ok {
		if !istRoot(w) {
			p.melde("SERVICE_ROOT_PATH", "muss / sein oder Segmente /a/b ohne abschließenden Schrägstrich")
		} else {
			d.Root = w
		}
	}
	d.UpstreamURL = p.upstreamURL("UPSTREAM_URL", p.env["UPSTREAM_URL"], record)
	return d
}

// upstreamURL prüft eine URL; fehlt sie, ist das im Modus record ein Verstoß.
func (p *pruefer) upstreamURL(schluessel, wert string, record bool) string {
	wert = strings.TrimSpace(wert)
	if wert == "" {
		if record {
			p.melde(schluessel, "im Modus record erforderlich")
		}
		return ""
	}
	u, ok := normalisiereURL(wert)
	if !ok {
		p.melde(schluessel, "keine absolute http- oder https-URL ohne Zugangsdaten, Query und Fragment")
		return ""
	}
	return u
}

// dateiDienste liest SERVICES_FILE; jede Verletzung ist ein Verstoß, nie der Wert.
func (p *pruefer) dateiDienste(pfad string, record bool) []Dienst {
	roh, grund := leseBegrenzt(pfad)
	if grund != "" {
		p.melde("SERVICES_FILE", grund)
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(roh))
	var eintraege []json.RawMessage
	if err := dec.Decode(&eintraege); err != nil {
		p.melde("SERVICES_FILE", "kein JSON-Array")
		return nil
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		p.melde("SERVICES_FILE", "Inhalt nach dem Array")
		return nil
	}
	if len(eintraege) == 0 {
		p.melde("SERVICES_FILE", "das Array ist leer")
		return nil
	}
	dienste := make([]Dienst, len(eintraege))
	for i, e := range eintraege {
		dienste[i] = p.dateiEintrag(i, e, record)
	}
	return dienste
}

func leseBegrenzt(pfad string) ([]byte, string) {
	f, err := os.Open(pfad)
	if err != nil {
		return nil, "Datei nicht lesbar"
	}
	defer func() { _ = f.Close() }()
	roh, err := io.ReadAll(io.LimitReader(f, maxServicesFile+1))
	switch {
	case err != nil:
		return nil, "Datei nicht lesbar"
	case len(roh) > maxServicesFile:
		return nil, "Datei größer als 1 MiB"
	case !utf8.Valid(roh):
		return nil, "Datei ist kein UTF-8"
	}
	return roh, ""
}

// dateiEintrag prüft ein Objekt der Datei; ein fehlerhaftes Feld bleibt leer und ist gemeldet.
func (p *pruefer) dateiEintrag(i int, roh json.RawMessage, record bool) Dienst {
	eintrag := fmt.Sprintf("SERVICES_FILE[%d]", i)
	var felder map[string]json.RawMessage
	if err := json.Unmarshal(roh, &felder); err != nil || felder == nil {
		p.melde(eintrag, "Eintrag ist kein Objekt")
		return Dienst{}
	}
	for name := range felder {
		if !slices.Contains(felderDienst(), name) {
			p.melde(eintrag, "unbekanntes Feld")
		}
	}
	text := func(feld string) string {
		var s *string
		if _, da := felder[feld]; !da {
			return ""
		}
		if err := json.Unmarshal(felder[feld], &s); err != nil {
			p.melde(eintrag+"."+feld, "muss eine Zeichenkette sein")
			return ""
		}
		if s == nil {
			return ""
		}
		return strings.TrimSpace(*s)
	}
	var d Dienst
	d.ID = p.pflichtfeld(eintrag+".id", text("id"), istDienstID, "1 bis 64 Zeichen aus Buchstaben, Ziffern, ., _ und -")
	d.Root = p.pflichtfeld(eintrag+".service_root", text("service_root"), istRoot, "muss / sein oder Segmente /a/b ohne abschließenden Schrägstrich")
	d.UpstreamURL = p.upstreamURL(eintrag+".upstream_url", text("upstream_url"), record)
	return d
}

func (p *pruefer) pflichtfeld(schluessel, wert string, gueltig func(string) bool, grund string) string {
	switch {
	case wert == "":
		p.melde(schluessel, "Pflichtfeld fehlt")
	case !gueltig(wert):
		p.melde(schluessel, grund)
	default:
		return wert
	}
	return ""
}

// pruefeDienste meldet doppelte `id` und überlappende Roots (SPEC-110 Nr. 9).
func (p *pruefer) pruefeDienste(dienste []Dienst) {
	for j, d := range dienste {
		for i := 0; i < j; i++ {
			if d.ID != "" && d.ID == dienste[i].ID {
				p.melde(fmt.Sprintf("SERVICES_FILE[%d].id", j), "doppelte id")
			}
			if d.Root != "" && dienste[i].Root != "" && ueberlappen(d.Root, dienste[i].Root) {
				p.melde(fmt.Sprintf("SERVICES_FILE[%d].service_root", j), "überlappt den Root eines anderen Services")
			}
		}
	}
}

// ueberlappen: einer der Roots geht dem anderen an einer Segmentgrenze voraus, oder sie sind gleich.
func ueberlappen(a, b string) bool {
	sa, sb := segmente(a), segmente(b)
	n := min(len(sa), len(sb))
	return slices.Equal(sa[:n], sb[:n])
}

func segmente(root string) []string {
	if root == "/" {
		return nil
	}
	return strings.Split(strings.TrimPrefix(root, "/"), "/")
}
