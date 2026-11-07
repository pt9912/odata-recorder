package mock

import (
	"slices"
	"strconv"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// bildeSchluessel bildet `Entity.Schluessel` aus den Teilen der URL (SPEC-196 Nr. 7 a): die Namen sind genau die Keys des
// Typs (jeder einmal, ein unbenannter Teil nur bei einem Schlüsselteil), das Literal passt zum EDM-Typ; der Text ist bei einem
// Teil sein Text, sonst `<Name>=<Text>` durch `,` getrennt in der Folge der Keys.
func bildeSchluessel(typ model.EntityType, teile []model.SchluesselTeil) (string, bool) {
	if len(teile) != len(typ.Keys) {
		return "", false
	}
	texte := make([]string, len(typ.Keys))
	gesetzt := make([]bool, len(typ.Keys))
	for _, t := range teile {
		name := t.Name
		if name == "" && len(typ.Keys) == 1 {
			name = typ.Keys[0]
		}
		i := slices.Index(typ.Keys, name)
		if i < 0 || gesetzt[i] || !literalPasst(typ, name, t.Literal) {
			return "", false
		}
		texte[i], gesetzt[i] = t.Literal.Text, true
	}
	if len(texte) == 1 {
		return texte[0], true
	}
	for i, k := range typ.Keys {
		texte[i] = k + "=" + texte[i]
	}
	return strings.Join(texte, ","), true
}

// literalPasst sagt, ob das Literal zum EDM-Typ der Property name passt (SPEC-196 Nr. 7 a).
func literalPasst(typ model.EntityType, name string, l model.Literal) bool {
	i := slices.IndexFunc(typ.Properties, func(p model.Property) bool { return p.Name == name })
	if i < 0 {
		return false
	}
	switch typ.Properties[i].EdmType {
	case "Edm.String":
		return l.Art == model.LiteralZeichenkette
	case "Edm.Boolean":
		return l.Art == model.LiteralWahrheitswert
	case "Edm.Int32":
		return l.Art == model.LiteralZahl && ganz(l.Text, 32)
	case "Edm.Int64":
		return l.Art == model.LiteralZahl && ganz(l.Text, 64)
	case "Edm.Decimal":
		return l.Art == model.LiteralZahl
	case "Edm.Guid":
		return l.Art == model.LiteralGUID
	case "Edm.Date":
		return l.Art == model.LiteralDatum
	case "Edm.DateTimeOffset":
		return l.Art == model.LiteralZeitpunkt
	}
	return false
}

// ganz sagt, ob text eine ganze Zahl im Bereich von bits Bits ist.
func ganz(text string, bits int) bool {
	_, err := strconv.ParseInt(text, 10, bits)
	return err == nil
}
