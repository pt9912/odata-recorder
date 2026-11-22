package bootstrap

import (
	"slices"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/csdl"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/csdlv2"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodec"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodecv2"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparserv2"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Die zwei Versionen, die der Root kennt (LH-FA-17.a, SPEC-175 Nr. 3); satzTabelle führt je Version einen Eintrag. Außer dieser
// Datei und konfig_version.go nennt keine Datei des Roots eine Version.
const (
	version4 = "4.0"
	version2 = "2.0"
)

// satzVersion4 ist der Satz der Codecs der Version 4.0 (SPEC-139 Nr. 11 b).
func satzVersion4() outbound.CodecSatz {
	return outbound.CodecSatz{
		Version: version4, Metadaten: csdl.New(), Daten: jsoncodec.New(),
		URL:       urlparser.New(urlparser.WithFilter(urlparser.NewFilterLeser())),
		Schreiber: jsoncodec.NewSchreiber(), Eingabe: jsoncodec.NewEingabe(),
	}
}

// satzVersion2 ist der Satz der Codecs der Version 2.0: der Mock liest nur, und es gibt keine Eingabe (SPEC-139 Nr. 11 b, SPEC-174 Nr. 3).
func satzVersion2() outbound.CodecSatz {
	return outbound.CodecSatz{
		Version: version2, Metadaten: csdlv2.New(), Daten: jsoncodecv2.New(),
		URL:       urlparserv2.New(urlparserv2.WithFilter(urlparserv2.NewFilterLeser())),
		Schreiber: jsoncodecv2.NewSchreiber(), NurLesend: true,
	}
}

// satzTabelle ist die Tabelle Version → Satz mit genau zwei Einträgen; sie hat keinen Zustand und wird je Aufruf gebildet.
func satzTabelle() map[string]outbound.CodecSatz {
	return map[string]outbound.CodecSatz{version4: satzVersion4(), version2: satzVersion2()}
}

// etagFuerSatz ist `Dienst.Etag`: ETAG_ENABLED gilt nur für einen Dienst, dessen Satz die Version 4.0 hat; ein Dienst der
// Version 2.0 liefert keinen ETag (SPEC-280 Nr. 2).
func etagFuerSatz(satz outbound.CodecSatz, an bool) bool {
	return an && satz.Version == version4
}

// wirksameDienste liefert die Dienste mit der wirksamen Version: eine leere gilt 4.0 (SPEC-139 Nr. 2); der Aufrufer bleibt unverändert.
func wirksameDienste(dienste []Dienst) []Dienst {
	aus := slices.Clone(dienste)
	for i := range aus {
		if aus[i].Version == "" {
			aus[i].Version = version4
		}
	}
	return aus
}

// wahl ist die Wahl des Satzes je Dienst: Dienste trägt in der Reihenfolge der Dienste den Satz der Version des Dienstes, Rueckfall
// den Satz 4.0, den die Konfig von Mock und Import verlangt und der für keinen Dienst gilt (SPEC-139 Nr. 4, 11 b).
type wahl struct {
	Dienste   []outbound.CodecSatz
	Rueckfall outbound.CodecSatz
}

// waehle ist die eine Stelle, an der der Root nach der Version wählt (SPEC-175 Nr. 3). Eine Version ohne Eintrag und der Modus
// hybrid mit einem Dienst der Version 2.0 sind model.ErrUngueltig (SPEC-139 Nr. 2); kein Dienst erhält den Satz eines anderen.
func waehle(k Konfiguration) (wahl, error) {
	tabelle := satzTabelle()
	w := wahl{Dienste: make([]outbound.CodecSatz, len(k.Dienste)), Rueckfall: tabelle[version4]}
	for i, d := range k.Dienste {
		satz, ok := tabelle[d.Version]
		switch {
		case !ok:
			return wahl{}, ungueltig("Version eines Dienstes nicht lieferbar")
		case k.Modus == modusHybrid && d.Version == version2:
			return wahl{}, ungueltig(grundHybrid)
		}
		w.Dienste[i] = satz
	}
	return w, nil
}
