// Package fixture trägt das Metadaten-Dokument des Referenzdienstes der Stufen von tests/e2e und den
// Hash, den der Übergang beim Import nennt (SPEC-037 Nr. 6): beide Seiten lesen dieselben Bytes.
package fixture

import (
	"crypto/sha256"
	"encoding/hex"
)

// CSDL ist das Dokument, das der Referenzdienst als `$metadata` liefert.
const CSDL = `<?xml version="1.0" encoding="utf-8"?>
<edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx">
  <edmx:DataServices>
    <Schema Namespace="Ref" xmlns="http://docs.oasis-open.org/odata/ns/edm">
      <EntityType Name="Product">
        <Key><PropertyRef Name="Id"/></Key>
        <Property Name="Id" Type="Edm.Int32" Nullable="false"/>
        <Property Name="Name" Type="Edm.String"/>
      </EntityType>
      <EntityContainer Name="Container">
        <EntitySet Name="Products" EntityType="Ref.Product"/>
      </EntityContainer>
    </Schema>
  </edmx:DataServices>
</edmx:Edmx>
`

// Hash ist der MetadataHash des Imports: SHA-256 der Bytes von CSDL.
func Hash() string {
	s := sha256.Sum256([]byte(CSDL))
	return "sha256:" + hex.EncodeToString(s[:])
}
