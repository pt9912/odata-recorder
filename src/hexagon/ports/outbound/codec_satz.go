package outbound

// CodecSatz ist der Satz der Implementierungen einer OData-Version (SPEC-175 Nr. 3). Der Composition
// Root hält je Version einen Satz und gibt ihn jedem Dienst, der Codecs braucht; ein Dienst liest allein
// die Felder, die er braucht, und ein Feld, das er nicht braucht, darf nil sein. Ein Satz hat keinen
// Zustand und darf von mehreren Diensten geteilt werden.
type CodecSatz struct {
	// Version ist die OData-Version des Satzes (`4.0` oder `2.0`); kein Dienst verzweigt danach.
	Version string
	// Metadaten liest das $metadata der Version (SPEC-125, SPEC-171).
	Metadaten MetadataCodec
	// Daten liest die Antworten der Version (SPEC-126, SPEC-172).
	Daten ODataCodec
	// URL liest Pfad und Query der OData-URL der Version (SPEC-160, SPEC-173); der Import liest es nie.
	URL URLCodec
	// Schreiber erzeugt die Antworten der Version (SPEC-168, SPEC-169, SPEC-172 Nr. 6); der Import liest es nie.
	Schreiber ODataSchreiber
	// Eingabe liest die Client-Bodies der Version (SPEC-180, SPEC-181); der Import liest es nie.
	Eingabe ODataEingabe
	// NurLesend sagt, dass der Mock die Version nur liest; der Import liest es nicht.
	NurLesend bool
}
