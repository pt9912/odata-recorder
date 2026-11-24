package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	wurzel  = "/odata"
	jsonTyp = "application/json;odata.metadata=minimal"
)

// referenz ist der Referenzdienst des Aufbaus (SPEC-216 Nr. 6 b): `$metadata`, die Collections `Set<n>` in Seiten und die Einzel-Entity
// `Set<n>(<Id>)`. Jede Antwort folgt aus Profil und Startwert; basis ist die Adresse des Dienstes ohne Pfad, die der `@odata.nextLink` nennt.
type referenz struct {
	p         Profil
	startwert uint64
	basis     string
}

// csdl ist das Dokument `$metadata`: ein EntityType Eintrag und je Set ein EntitySet.
func (r referenz) csdl() []byte {
	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="utf-8"?>
<edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx">
  <edmx:DataServices>
    <Schema Namespace="Lasttest" xmlns="http://docs.oasis-open.org/odata/ns/edm">
      <EntityType Name="Eintrag">
        <Key><PropertyRef Name="Id"/></Key>
        <Property Name="Id" Type="Edm.Int32" Nullable="false"/>
        <Property Name="Name" Type="Edm.String"/>
        <Property Name="Gruppe" Type="Edm.Int32"/>
        <Property Name="Nutzlast" Type="Edm.String"/>
      </EntityType>
      <EntityContainer Name="Container">
`)
	for set := 1; set <= r.p.Sets; set++ {
		fmt.Fprintf(&b, "        <EntitySet Name=\"Set%d\" EntityType=\"Lasttest.Eintrag\"/>\n", set)
	}
	b.WriteString("      </EntityContainer>\n    </Schema>\n  </edmx:DataServices>\n</edmx:Edmx>\n")
	return b.Bytes()
}

// csdlHash ist der MetadataHash des Imports: `sha256:` und der SHA-256 der Bytes von csdl.
func (r referenz) csdlHash() string {
	s := sha256.Sum256(r.csdl())
	return "sha256:" + hex.EncodeToString(s[:])
}

func (r referenz) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	rest, ok := strings.CutPrefix(req.URL.Path, wurzel+"/")
	switch {
	case !ok || req.Method != http.MethodGet:
	case rest == "$metadata" && req.URL.RawQuery == "":
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write(r.csdl())
		return
	case r.menge(w, req, rest):
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`{"error":{"code":"404","message":"nicht gefunden"}}`))
}

// zahl liest text als Ganzzahl von 1 bis hoechst in der Schreibweise ohne führende Null.
func zahl(text string, hoechst int) (int, bool) {
	n, err := strconv.Atoi(text)
	return n, err == nil && strconv.Itoa(n) == text && n >= 1 && n <= hoechst
}

// menge bedient `Set<n>`, `Set<n>?$skiptoken=<k>` und `Set<n>(<Id>)`; sie meldet, ob sie geantwortet hat.
func (r referenz) menge(w http.ResponseWriter, req *http.Request, rest string) bool {
	name, arg, einzel := strings.Cut(rest, "(")
	text, mitPraefix := strings.CutPrefix(name, "Set")
	set, gueltig := zahl(text, r.p.Sets)
	if !mitPraefix || !gueltig {
		return false
	}
	var b bytes.Buffer
	if einzel {
		idText, mitKlammer := strings.CutSuffix(arg, ")")
		id, gueltig := zahl(idText, r.p.Entities)
		if !mitKlammer || !gueltig || req.URL.RawQuery != "" {
			return false
		}
		r.entity(&b, set, id, fmt.Sprintf(`"@odata.context":"%s%s/$metadata#Set%d/$entity",`, r.basis, wurzel, set))
	} else {
		von, gueltig := r.skiptoken(req.URL.RawQuery)
		if !gueltig {
			return false
		}
		r.seite(&b, set, von)
	}
	w.Header().Set("Content-Type", jsonTyp)
	w.Header().Set("OData-Version", "4.0")
	_, _ = w.Write(b.Bytes())
	return true
}

// skiptoken liest die Zahl der Entities vor der Seite: keine Query ist die erste Seite, `$skiptoken=<k>` ein Vielfaches der Seitengröße.
func (r referenz) skiptoken(query string) (int, bool) {
	if query == "" {
		return 0, true
	}
	text, ok := strings.CutPrefix(query, "$skiptoken=")
	k, gueltig := zahl(text, r.p.Entities-1)
	return k, ok && gueltig && k%r.p.SeitenGroesse == 0
}

// seite schreibt die Seite mit den Entities von+1 bis zur Seitengröße; jede außer der letzten nennt den `@odata.nextLink` der nächsten.
func (r referenz) seite(b *bytes.Buffer, set, von int) {
	bis := min(von+r.p.SeitenGroesse, r.p.Entities)
	fmt.Fprintf(b, `{"@odata.context":"%s%s/$metadata#Set%d",`, r.basis, wurzel, set)
	if bis < r.p.Entities {
		fmt.Fprintf(b, `"@odata.nextLink":"%s%s/Set%d?$skiptoken=%d",`, r.basis, wurzel, set, bis)
	}
	b.WriteString(`"value":[`)
	for id := von + 1; id <= bis; id++ {
		if id > von+1 {
			b.WriteByte(',')
		}
		r.entity(b, set, id, "")
	}
	b.WriteString("]}")
}

// entity schreibt das Objekt der Entity; kopf steht vor den vier Feldern.
func (r referenz) entity(b *bytes.Buffer, set, id int, kopf string) {
	fmt.Fprintf(b, `{%s"Id":%d,"Name":"Set%d-%d","Gruppe":%d,"Nutzlast":"`, kopf, id, set, id, id%gruppen)
	b.Write(nutzlast(r.startwert, set, id))
	b.WriteString(`"}`)
}

// starteReferenz startet den Referenzdienst auf adresse; umhuellen ersetzt den Handler (nil: keine Ersetzung). Die Funktion beendet den Server.
func starteReferenz(ctx context.Context, adresse string, p Profil, startwert uint64, umhuellen func(http.Handler) http.Handler) (referenz, func(), error) {
	var lc net.ListenConfig
	l, err := lc.Listen(ctx, "tcp", adresse)
	if err != nil {
		return referenz{}, nil, err
	}
	r := referenz{p: p, startwert: startwert, basis: "http://" + l.Addr().String()}
	var h http.Handler = r
	if umhuellen != nil {
		h = umhuellen(h)
	}
	srv := &http.Server{Handler: h, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = srv.Serve(l) }()
	return r, func() { _ = srv.Close() }, nil
}
