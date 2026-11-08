package snapshot

import (
	"bytes"
	"slices"
	"strconv"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const hexKlein = "0123456789abcdef"

// manifest bildet die Bytes des Manifests eines Workspace-Snapshots (SPEC-129 Nr. 20): Kanonform ohne Leerraum,
// Mitglieder aufsteigend nach den Bytes ihrer Namen, Einträge aufsteigend nach (entity_set, typed_key), payload
// byte-gleich. Die Folge des Ports gilt nicht; nichts im Manifest hängt von Zeit, Namen, Kennung oder Revision ab.
func manifest(serviceID, metadataHash string, es []model.EntityState) []byte {
	sortiert := slices.Clone(es)
	slices.SortFunc(sortiert, func(a, b model.EntityState) int {
		if c := bytes.Compare([]byte(a.EntitySet), []byte(b.EntitySet)); c != 0 {
			return c
		}
		return bytes.Compare([]byte(a.EntityKey), []byte(b.EntityKey))
	})
	var b bytes.Buffer
	b.WriteString(`{"entities":[`)
	for i, e := range sortiert {
		if i > 0 {
			b.WriteByte(',')
		}
		eintrag(&b, e)
	}
	b.WriteString(`],"entity_count":`)
	b.WriteString(strconv.Itoa(len(sortiert)))
	b.WriteString(`,"manifest_version":1,"metadata_hash":`)
	text(&b, metadataHash)
	b.WriteString(`,"service_id":`)
	text(&b, serviceID)
	b.WriteString(`,"snapshot_kind":"workspace"}`)
	return b.Bytes()
}

// eintrag schreibt einen Eintrag von entities (SPEC-129 Nr. 20): sieben Glieder, Properties aufsteigend.
func eintrag(b *bytes.Buffer, e model.EntityState) {
	b.WriteString(`{"completeness":`)
	text(b, string(e.Completeness))
	b.WriteString(`,"deleted":`)
	b.WriteString(strconv.FormatBool(e.Deleted))
	b.WriteString(`,"entity_set":`)
	text(b, e.EntitySet)
	b.WriteString(`,"payload":`)
	b.WriteString(e.Payload)
	b.WriteString(`,"selected_properties":[`)
	props := slices.Clone(e.KnownProperties)
	slices.Sort(props)
	for i, p := range props {
		if i > 0 {
			b.WriteByte(',')
		}
		text(b, p)
	}
	b.WriteString(`],"typed_key":`)
	text(b, e.EntityKey)
	b.WriteString(`,"version":`)
	b.WriteString(strconv.FormatInt(e.Version, 10))
	b.WriteByte('}')
}

// text schreibt s als Zeichenkette der Kanonform: `\"` und `\\`, `\u00xx` mit kleinen Hexziffern für jedes Byte
// unter 0x20, sonst unverändert.
func text(b *bytes.Buffer, s string) {
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '"' || c == '\\':
			b.WriteByte('\\')
			b.WriteByte(c)
		case c < 0x20:
			b.WriteString(`\u00`)
			b.WriteByte(hexKlein[c>>4])
			b.WriteByte(hexKlein[c&15])
		default:
			b.WriteByte(c)
		}
	}
	b.WriteByte('"')
}
