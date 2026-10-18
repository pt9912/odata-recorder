# Erfassungsschicht — die Feldliste und ihre Grenzen

Dieses Dokument ist **werkzeug-erzeugt**: es ist der Ausdruck der Erfassungsschicht über ihr
eigenes Schema. Ein Feld, das erfasst wird, steht in der Tabelle; einen Eintrag der Tabelle
ohne erfasstes Feld gibt es nicht. Beide entstehen aus derselben Quelle — deshalb kann die
Liste nicht gegen die Erfassung driften.

**Ein erneuter Lauf des Werkzeugs schreibt diese Datei kanonisch neu.** Änderungen von Hand
gehen dabei verloren; sie gehören in ein eigenes Dokument daneben.

## Was erfasst wird

Je Werkzeug-Aufruf eines Agenten-Laufs entsteht **eine** Zeile JSON in einem gitignorierten
Zustands-Bereich unterhalb von `.harness/`, ein Strom je Paar aus Sitzung und Agent.

**Das Schema ist geschlossen.** Erfasst wird ausschließlich, was die Tabelle unten führt; ein
Feld einer künftigen Werkzeug-Fassung wird nicht still mitgeschrieben. Von Argument-Werten
wandert **nie der Inhalt**, sondern eine Ableitung: der Pfad, die Größe, ein
Fingerabdruck-Präfix, das Programm-Token, die Anzahl der Argumente. Ein Werkzeug, das die
Erfassung nicht namentlich führt, gibt **nur seinen Namen und seinen Status** preis.

**Die Liste gilt auch dann, wenn gerade nichts erfasst wird.** Die Erfassung läuft über ein
Programm, das gitignored liegt: ein frischer Klon dieses Repos hat es nicht, dieses Dokument
schon. Dann sagt die Tabelle, was erfasst **würde**, sobald ein erneuter Lauf des Werkzeugs
das Programm wieder ablegt.

## Feldliste

**Pflicht** heißt: das Feld steht in jeder Zeile. Eine leere Liste `[]` ist dort eine Aussage —
keiner — und kein fehlender Wert. Die Kennzeichnung *nicht bekannt* tragen abschließend diese Felder:
`cache_creation_input_tokens`, `cache_read_input_tokens`, `agent_role`, `branch`, `commit`, `slice`,
`requirement`, `adr`. Liefert die Quelle ihren Wert nicht, steht statt seiner `nicht bekannt:` und die
Quelle, die ihn nicht liefert — nie `0`, nie `""`, nie `[]`. Für die übrigen Pflichtfelder gilt sie
nicht; im Haupt-Kontext stehen etwa `agent`, `agent_type`, `tool_use_id` und `event` als `""`.
**Optional** heißt: das Feld fehlt, wo es nichts zu sagen gibt.

| Feld | Pflicht | Wonach gefragt wird |
|---|---|---|
| `seq` | Pflicht | Fehlt eine Zeile? — je Strom vergeben und steigend, damit eine Lücke sichtbar wird |
| `ts` | Pflicht | Wann geschah es? |
| `event` | Pflicht | Welches Ereignis löste die Zeile aus — Nachlauf oder Fehlschlag? |
| `tool` | Pflicht | Welches Werkzeug lief? |
| `tool_use_id` | Pflicht | Welche Ereignisse gehören zu einem Aufruf? |
| `session` | Pflicht | Welcher Lauf war es? — zusammen mit `agent` der Strom |
| `agent` | Pflicht | Welcher Agent innerhalb des Laufs? — zusammen mit `session` der Strom |
| `agent_type` | Pflicht | Welche Art Lauf? — der Typ des laufenden Agenten, roh übernommen |
| `agent_role` | Pflicht | Welche Rolle verursachte den Zugriff? — besetzt, wenn `agent_type` eine kanonische Rolle nennt, sonst `nicht bekannt: agent_type` |
| `slice` | Pflicht | Auf wessen Rechnung lief der Zugriff? — aus dem Lifecycle-Verzeichnis abgeleitet, Liste; `[]` heißt kein Slice, ein unlesbares Verzeichnis die Kennzeichnung |
| `requirement` | Pflicht | Gegen welche Anforderung? — aus dem Bezug-Block der laufenden Slices, Liste; `[]` heißt kein Bezug, eine unlesbare Slice-Datei die Kennzeichnung |
| `adr` | Pflicht | Auf wessen Entscheidung? — aus demselben Bezug-Block, Liste; `[]` und Kennzeichnung wie bei `requirement` |
| `branch` | Pflicht | Zu welchem Zweig gehört der Zugriff? — aus dem git-Zustand abgeleitet, sonst `nicht bekannt: .git/HEAD` |
| `commit` | Pflicht | Zu welchem Stand gehört der Zugriff? — aus dem git-Zustand abgeleitet, sonst `nicht bekannt: .git/HEAD` |
| `status` | Pflicht | Ging es gut? |
| `rule_version` | Pflicht | Unter welcher Fassung der Erfassungsregel entstand die Zeile? — eine Ganzzahl, die mit jedem Bedeutungswechsel eines Feldes steigt; eine Zeile ohne dieses Feld hat eine nicht bekannte Fassung |
| `permission_mode` | Optional | Unter welcher Berechtigungs-Lage lief der Aufruf? |
| `path` | Optional | Was wurde gelesen oder geschrieben? — der Pfad, nie der Inhalt, und nur bei namentlich geführten Datei-Werkzeugen |
| `bytes` | Optional | Wie groß ist die geschriebene Datei? — aus dem Dateisystem, nie aus der Payload |
| `sha256_16` | Optional | Hat sich der Inhalt geändert? — ein Fingerabdruck-Präfix aus dem Dateisystem, nie der Inhalt selbst |
| `program` | Optional | Welches Programm lief? — das erste Wort des ausgeführten Segments, nie das der ganzen Kommandozeile |
| `argc` | Optional | Wie viele Argumente hatte es? — die Anzahl, nie die Argumente |
| `duration_ms` | Optional | Wie lange dauerte der Aufruf, wie der Hook ihn sieht? |
| `result_bytes` | Optional | Wie groß war das Ergebnis? — die Länge, nie der Inhalt |
| `spawned_role` | Optional | Welche Rolle lief im Subagenten? — aus dem Ergebnis, gegen die kanonischen Namen normalisiert |
| `input_tokens` | Optional | Wie viele Eingabe-Token verbrauchte der Subagenten-Lauf? |
| `output_tokens` | Optional | Wie viele Ausgabe-Token verbrauchte er? |
| `cache_creation_input_tokens` | Pflicht | Zahlte der Lauf den Cache? — ohne Zähler im Ergebnis `nicht bekannt: tool_response.usage` |
| `cache_read_input_tokens` | Pflicht | Nutzte der Lauf den Cache? — ohne Zähler im Ergebnis `nicht bekannt: tool_response.usage` |
| `total_tokens` | Optional | Wie groß war der Subagenten-Lauf insgesamt? — die Summe, die das Werkzeug selbst ausweist |
| `total_duration_ms` | Optional | Wie lange lief der Subagent selbst? — nicht `duration_ms`, das den Aufruf misst |
| `total_tool_use_count` | Optional | Wie viele Werkzeug-Aufrufe verursachte der Subagent? |
| `model_version` | Optional | Welches Modell verursachte die Kosten? — strukturell begrenzt; was die Gestalt eines Bezeichners nicht hat, wird verworfen |

## Verfügbarkeit und Aufbewahrung

**Den Cache-Status liefert nur ein Subagenten-Aufruf im Vordergrund.** Die zwei
Cache-Zähler tragen eine Zahl nur im Span eines solchen Aufrufs, und die Zahl ist die des
Subagenten-Laufs. Jeder andere Span und ein Aufruf, dessen Ergebnis keine Zähler führt, trägt in
beiden Feldern `nicht bekannt: tool_response.usage`; die Felder stehen trotzdem in jeder Zeile.
Der Cache des Haupt-Kontexts selbst steht in keinem Span.

**Eine PR-Nummer steht bewusst nicht im Schema.** Sie lebt bei der Forge, und die
Erfassung läuft je Werkzeug-Aufruf, ohne Netz und ohne `gh`. An ihrer Stelle stehen `branch`
und `commit`, abgeleitet aus `.git/HEAD`; über sie schlägt eine Auswertung den PR nach. Das ist
eine Ableitung, keine Erfüllung: liegt zum Zweig kein PR vor, bleibt die Frage offen.

**Der Haupt-Kontext trägt keine Zahl.** Die Token-Zähler und die drei
Gesamtwerte stehen ausschließlich im Ergebnis eines Subagenten-Aufrufs, und den Haupt-Kontext
umschließt keiner. `result_bytes` und `duration_ms` sind Größen eines Aufrufs, keine Token;
geschätzt wird nicht. Jede Token-Bilanz aus diesen Zeilen ist eine Bilanz über
Subagenten-Läufe: ihr Nenner ist nicht der Verbrauch des Laufs, und ein Prozentsatz daraus ist
ein Anteil an der erfassten Teilmenge.

**Der Bestand wird nie nebenbei geräumt.** Die Erfassung hängt ausschließlich an;
Altbestände bleiben auch beim ersten Span einer Sitzung liegen. Aufgeräumt wird ausdrücklich mit
`make span-clean`, das den ganzen Bestand entfernt. Ein Werkzeug, das Sitzungs-Kennungen
wiederverwendet, mischt zwei Läufe in einer Datei.

## Grenzen, die kein Sensor hält

Sie gelten auch dann, wenn niemand eine Auswertung ruft — deshalb stehen sie hier und nicht
nur in deren Ausgabe.

**Über die Aufrufform des Agenten-Werkzeugs führt diese Ebene keinen Wächter.**
Das Feld `agent_role` besetzt sich genau dann, wenn der Agenten-Typ eine der sechs kanonischen
Rollen nennt — `planner`, `architect`, `implementer`, `reviewer`, `verifier`, `validator`. Wer
seine Typen umbenennt, bekommt `nicht bekannt: agent_type`, und **das heißt unbekannt, nie rollenlos**.
Kein Gate und kein Hook erzwingt, dass Rollen-Arbeit unter ihrem Rollen-Typ läuft; die
Rollen-Achse ruht hier auf Disziplin.

**Die Verbrauchs-Zähler kommen aus der Mechanik des Agenten-Werkzeugs nicht.**
Die Token- und Cache-Zähler tragen einen Wert nur, wenn das Werkzeug sie im Ergebnis eines
Subagenten-Aufrufs mitliefert; ein im Hintergrund gestarteter Lauf liefert sie nicht. Dann
fehlen die Token-Zähler, und die Cache-Zähler tragen `nicht bekannt: tool_response.usage`. **Kein
Lauf dieses Repos führt sie herbei** — das ist keine Eigenschaft dieses Aufbaus, sondern der
Mechanik. Ein Bestand ohne Zähler ist deshalb der Normalfall und kein Defekt.

**Über den Schutz des Bestands ist wenig zugesagt.** Er ist **gitignored** und **nicht
verschlüsselt**. Seine Dateien entstehen **nur für den Eigentümer lesbar** (Modus 0600); das
Verzeichnis selbst bleibt für andere auflistbar. Und **Pfadnamen sind nicht als unkritisch
zugesagt** — sie stehen als `path` in der Zeile, und ein Pfad kann selbst die Aussage sein, die
niemand teilen wollte. Wer den Bestand weitergibt, gibt beides weiter.
