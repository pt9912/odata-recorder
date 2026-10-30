# Benutzerhandbuch Standard

Beim Erstellen eines Benutzerhandbuchs für Software zählt vor allem eines: **Der Nutzer muss sein Ziel erreichen, ohne die Software oder den Entwickler verstehen zu müssen.** Ein gutes Handbuch ist kein Ablageort für Feature-Beschreibungen, sondern eine Arbeitsanleitung.

## 1. Zielgruppe klären

Vor dem Schreiben musst du wissen, für wen das Handbuch ist:

* Endanwender ohne technisches Vorwissen
* Administratoren
* Power-User
* Support-Mitarbeiter
* Entwickler oder Integratoren

Davon hängt ab, welche Begriffe du erklärst, wie detailliert du wirst und welche Aufgaben du beschreibst.

Schlecht:

> Der Benutzer kann über die Konfigurationsmaske die Parameter persistieren.

Besser:

> Klicken Sie auf **Speichern**, um die Einstellungen dauerhaft zu übernehmen.

## 2. Aufgaben statt Funktionen beschreiben

Viele Handbücher sind schlecht, weil sie die Oberfläche der Reihe nach erklären:

> Menüpunkt Datei
> Menüpunkt Bearbeiten
> Menüpunkt Einstellungen

Besser ist eine Struktur nach Nutzerzielen:

* Erste Anmeldung
* Passwort ändern
* Neues Projekt anlegen
* Daten importieren
* Bericht exportieren
* Fehler beheben
* Benutzer verwalten

Der Nutzer fragt nicht: „Was macht dieser Button?“, sondern: „Wie erledige ich meine Aufgabe?“

## 3. Klare Struktur verwenden

Ein typisches Benutzerhandbuch sollte enthalten:

1. **Einleitung**

   * Zweck der Software
   * Zielgruppe des Handbuchs
   * Voraussetzungen

2. **Installation oder Zugriff**

   * Systemanforderungen
   * Installation
   * Login
   * Rollen und Rechte

3. **Erste Schritte**

   * Schneller Einstieg
   * Beispielablauf
   * Wichtigste Bedienkonzepte

4. **Bedienungsanleitung nach Aufgaben**

   * Schritt-für-Schritt-Anleitungen
   * Screenshots
   * Hinweise
   * Ergebnis der Aktion

5. **Konfiguration**

   * Einstellungen
   * Benutzerrollen
   * Schnittstellen
   * Import/Export

6. **Fehlerbehebung**

   * Häufige Probleme
   * Fehlermeldungen
   * Ursachen
   * Lösungen

7. **FAQ**

   * Kurze Antworten auf typische Fragen

8. **Glossar**

   * Fachbegriffe
   * Abkürzungen
   * Produktspezifische Begriffe

9. **Anhang**

   * Tastenkürzel
   * Formate
   * Grenzwerte
   * Kontakt zum Support

## 4. Einheitliche Sprache nutzen

Sprich den Nutzer direkt und eindeutig an.

Gut:

> Klicken Sie auf **Neues Projekt**.

Nicht gut:

> Man kann dann gegebenenfalls ein neues Projekt erstellen.

Wichtig:

* Ein Begriff = immer derselbe Begriff
* Button-Namen exakt wie in der Software schreiben
* Keine unnötigen Fremdwörter
* Keine langen Schachtelsätze
* Keine Entwicklerbegriffe, wenn sie für Nutzer nicht nötig sind

## 5. Schritt-für-Schritt-Anleitungen sauber schreiben

Eine gute Anleitung hat immer:

* Ausgangssituation
* Ziel
* Voraussetzungen
* nummerierte Schritte
* erwartetes Ergebnis
* Hinweise auf mögliche Fehler

Beispiel:

```markdown
## Neues Projekt anlegen

### Voraussetzung

Sie sind angemeldet und haben die Rolle **Projektmanager**.

### Vorgehen

1. Öffnen Sie im Hauptmenü den Bereich **Projekte**.
2. Klicken Sie auf **Neues Projekt**.
3. Geben Sie einen Projektnamen ein.
4. Wählen Sie einen Projektstatus aus.
5. Klicken Sie auf **Speichern**.

### Ergebnis

Das neue Projekt wird angelegt und erscheint in der Projektübersicht.
```

## 6. Screenshots sinnvoll einsetzen

Screenshots helfen, aber sie dürfen nicht die Anleitung ersetzen.

Achte auf:

* aktuelle Oberfläche
* keine echten Kundendaten
* keine Passwörter, Tokens, E-Mail-Adressen oder internen URLs
* Markierungen nur dort, wo sie helfen
* einheitliche Bildgröße
* gute Lesbarkeit
* Screenshots nach jedem UI-Update prüfen

Nicht jeder Klick braucht einen Screenshot. Zu viele Bilder machen das Handbuch schwer wartbar.

## 7. Rollen und Rechte erklären

Gerade bei Business-Software ist wichtig:

* Wer darf was sehen?
* Wer darf was bearbeiten?
* Warum fehlt ein Button?
* Welche Rolle braucht man für welche Aufgabe?

Beispiel:

```markdown
Hinweis: Die Schaltfläche **Benutzer löschen** wird nur angezeigt, wenn Sie die Rolle **Administrator** besitzen.
```

## 8. Fehlermeldungen und Problemfälle dokumentieren

Ein gutes Handbuch behandelt nicht nur den Idealfall.

Dokumentiere:

* häufige Fehlermeldungen
* ungültige Eingaben
* fehlende Rechte
* Netzwerkprobleme
* Importfehler
* Konflikte bei gleichzeitiger Bearbeitung
* abgelaufene Sitzungen

Gute Struktur:

```markdown
## Fehler: "Datei konnte nicht importiert werden"

### Ursache

Die Datei hat nicht das erwartete CSV-Format.

### Lösung

1. Öffnen Sie die Datei in einem Texteditor.
2. Prüfen Sie, ob die erste Zeile die Spaltenüberschriften enthält.
3. Speichern Sie die Datei erneut im UTF-8-Format.
4. Starten Sie den Import erneut.
```

## 9. Version und Aktualität sicherstellen

Ein Handbuch veraltet schnell. Deshalb müssen diese Angaben rein:

* Software-Version
* Handbuch-Version
* Änderungsdatum
* Autor oder Team
* Changelog oder Änderungshistorie (je Version eine Zeile in Betreibersicht, ohne interne Kennungen)
* Gültigkeitsbereich

Beispiel:

```markdown
Software-Version: 2.4.1
Handbuch-Version: 1.3
Stand: 15.06.2026
```

Ohne Versionsbezug ist ein Handbuch bei produktiver Software schnell wertlos.

## 10. Rechtliche und organisatorische Punkte beachten

Je nach Software können wichtig sein:

* Datenschutz
* Umgang mit personenbezogenen Daten
* Lizenzhinweise
* Sicherheitswarnungen
* Barrierefreiheit
* Aufbewahrungsfristen
* Exportkontrolle
* Haftungsausschlüsse
* interne Compliance-Vorgaben

Besonders bei Admin-Handbüchern solltest du keine sensiblen Details veröffentlichen, die Angreifern helfen könnten, etwa interne Servernamen, Tokens, genaue Sicherheitskonfigurationen oder Produktionszugänge.

## 11. Barrierefreiheit berücksichtigen

Ein Handbuch sollte auch ohne perfekte Sicht, Maus oder Spezialwissen nutzbar sein.

Achte auf:

* klare Überschriftenstruktur
* Alternativtexte für Bilder
* ausreichende Kontraste
* keine Informationen nur über Farbe vermitteln
* Tabellen sparsam und sauber verwenden
* PDF mit Lesezeichen und sinnvoller Struktur exportieren
* Suchfunktion nutzbar machen

## 12. Wartbarkeit einplanen

Das wird oft unterschätzt. Ein Benutzerhandbuch ist kein einmaliges Dokument, sondern Teil des Produkts.

Praktisch sinnvoll:

* Handbuch in Markdown, AsciiDoc oder einem Docs-System pflegen
* Screenshots automatisiert oder halbautomatisiert aktualisieren
* Doku-Version an Software-Version koppeln
* Review-Prozess mit Releases verbinden
* Verantwortliche Person oder Team festlegen
* veraltete Abschnitte markieren oder entfernen

## 13. Qualität prüfen

Vor Veröffentlichung sollte geprüft werden:

* Sind alle beschriebenen Funktionen in der Software vorhanden?
* Stimmen Button-Namen, Menüs und Screenshots?
* Kann ein neuer Nutzer die Schritte ausführen?
* Sind Fehlermeldungen korrekt beschrieben?
* Gibt es tote Links?
* Sind sensible Daten entfernt?
* Ist die Sprache einheitlich?
* Gibt es eine PDF- oder Online-Version?
* Funktioniert die Suche?

Der beste Test: Gib das Handbuch jemandem, der die Software nicht kennt. Wenn diese Person die Aufgabe ohne Rückfragen erledigt, ist die Anleitung gut.

## 14. Empfohlene Vorlage

```markdown
# Benutzerhandbuch: <Produktname>

Version: <Handbuch-Version>  
Software-Version: <Software-Version>  
Stand: <Datum>

## 1. Einleitung

### Zweck der Software
### Zielgruppe
### Voraussetzungen

## 2. Erste Schritte

### Anmeldung
### Überblick über die Oberfläche
### Grundlegende Bedienung

## 3. Aufgaben ausführen

### Aufgabe 1
#### Voraussetzung
#### Vorgehen
#### Ergebnis
#### Hinweise

### Aufgabe 2
#### Voraussetzung
#### Vorgehen
#### Ergebnis
#### Hinweise

## 4. Einstellungen

## 5. Rollen und Rechte

## 6. Import und Export

## 7. Fehlerbehebung

## 8. FAQ

## 9. Glossar

## 10. Support und Kontakt

## 11. Änderungshistorie
```

Das letzte Kapitel führt je Version eine Zeile in Betreibersicht: Es sagt, was sich für die Leserin oder den Leser ändert. Es nennt keine internen Kennungen, keine Namen von Arbeitspaketen und keine Links in die Planungsunterlagen.

## Kurz gesagt

Ein gutes Benutzerhandbuch ist:

**zielgruppenorientiert, aufgabenbasiert, aktuell, eindeutig, suchbar, bebildert, barrierearm und wartbar.**

Der häufigste Fehler ist, zu sehr aus Entwicklersicht zu schreiben. Schreibe nicht, was die Software kann. Schreibe, wie der Nutzer damit seine Arbeit erledigt.

## Anpassung für dieses Repo

Dieser Abschnitt gilt für das Benutzerhandbuch des OData Recorder.
Er ist der einzige Teil dieser Datei, der von der Quelle des Standards abweicht; die Abschnitte 1 bis 14 oben bleiben unverändert, damit der Abgleich mit der Quelle möglich bleibt.
Wo dieser Abschnitt und ein Abschnitt oben sich widersprechen, gilt dieser.
Der Standard ist Regelwerk, kein Handbuch: die Regel und das Gate unten gelten für das Handbuch, nicht für diese Datei.

### Regel: nur der Ist-Zustand

Das Handbuch beschreibt ausschließlich den Ist-Zustand des gebauten Binary der angegebenen Software-Version.
Es enthält nichts Geplantes und keine Ausblicke, keine Bezüge auf Roadmap, Wellen, Slices, ADRs oder Kennungen der Spezifikation und keine Chronik (kein Änderungsverlauf, kein „neu in“, kein „bisher“, kein „seit Version“).
Ein Kapitel entsteht, wenn das Binary den Gegenstand zeigt.
Verhalten des Binary steht im Handbuch als Beispiel: Der Aufruf, die Ausgabe je Strom und der Ausgangswert stehen in einer Datei unter docs/user/beispiele, das Handbuch zeigt sie in einem umzäunten Block mit einer Bindungszeile davor, und `make handbuch-check` fährt jede Datei gegen das gebaute Binary und hält den Block gleich der Datei.
Host-Befehle stehen als Inline-Code, und `make handbuch-check` prüft, dass das Ziel von `make` existiert.
Das Format der Beispiel-Dateien, die Wörter und Kennungen, die das Handbuch nicht enthält, und die Grenzen der Prüfung stehen im Vertrag des Gates (Ordner `harness/vertraege/`).

### Entscheidung je Abschnitt

| Abschnitt | Entscheidung | Festlegung für dieses Repo |
|---|---|---|
| 1 Zielgruppe klären | gilt angepasst | Zielgruppen sind Anwendungsentwickler und Testautomatisierung/QA, die den Recorder einsetzen, und das Plattform-/CI-Team, das ihn baut und betreibt. Ein Handbuch bedient beide; die Kapitel sind nach Aufgaben benannt, Aufgaben des Betriebs sind eigene Kapitel. Endanwender ohne Vorwissen und Support-Mitarbeiter sind keine Zielgruppe. Das Handbuch führt einen Begriff an der Stelle ein, an der ein Kapitel ihn braucht. |
| 2 Aufgaben statt Funktionen | gilt | Die Beispiele der Liste (Anmeldung, Passwort, Projekt) gehören zu einer Oberfläche; jedes Kapitel nennt seine Aufgabe selbst, aus Sicht dessen, der das Binary baut, aufruft oder eine Meldung deutet. |
| 3 Klare Struktur verwenden | gilt angepasst | Es stehen nur Kapitel, deren Aussagen das gebaute Binary belegt; eine Überschrift als Platzhalter gibt es nicht. „Installation oder Zugriff“ ist „Bauen“ (Voraussetzungen und Bauen, ohne Login und Rollen), „Erste Schritte“ ist „Aufruf“, Screenshots entfallen (Abschnitt 6), „Konfiguration“ führt Einstellungen und Schnittstellen ohne Benutzerrollen. „Fehlerbehebung“ gilt. „FAQ“ gilt nicht: eine Frage ist eine Aufgabe oder ein Eintrag der Fehlerbehebung, damit ein Satz an einer Stelle steht. „Glossar“ steht ab dem ersten Begriff, den ein Kapitel braucht und das Binary zeigt. Im Anhang entfallen Tastenkürzel und Kontakt zum Support; Formate und Grenzwerte stehen, sobald das Binary sie zeigt. Überschriften tragen keine Nummer, damit ein neues Kapitel keines umbenennt; Verweise nennen den Titel. |
| 4 Einheitliche Sprache nutzen | gilt angepasst | Die Anrede, ein Begriff in einer Bedeutung, keine langen Schachtelsätze und keine unnötigen Fremdwörter gelten. Statt „Button-Namen exakt wie in der Software“ gilt: Namen von Kommandos, Optionen und Schlüsseln und der Wortlaut von Meldungen stehen exakt, wie das Binary sie schreibt, in Inline-Code oder im Beispiel. Begriffe der Kommandozeile (Binary, Image, Standardausgabe, Fehlerausgabe, Ausgangswert) darf das Handbuch benutzen, weil die Zielgruppe sie kennt. |
| 5 Schritt-für-Schritt-Anleitungen | gilt angepasst | Der Aufbau (Ausgangssituation, Ziel, Voraussetzungen, nummerierte Schritte, erwartetes Ergebnis, Hinweise auf Fehler) gilt. Die Voraussetzungen nennen Docker und GNU `make` statt Anmeldung und Rolle. Ein Schritt ist ein Host-Befehl als Inline-Code oder der Aufruf des Binary als Beispiel; das erwartete Ergebnis eines Aufrufs ist die Ausgabe im Beispiel. |
| 6 Screenshots | gilt nicht | Das Binary hat keine Oberfläche. An die Stelle des Screenshots tritt das Beispiel mit Aufruf, Ausgabe je Strom und Ausgangswert, gefahren gegen das gebaute Binary. |
| 7 Rollen und Rechte | gilt nicht | Das gebaute Binary kennt keine Anmeldung und keine Rollen. Was es nicht zeigt, steht nicht im Handbuch, auch nicht als Hinweis, dass es fehlt. |
| 8 Fehlermeldungen und Problemfälle | gilt angepasst | Der Aufbau (Meldung im Wortlaut, Ursache, Lösung) gilt für jede Meldung, die das Binary schreibt, und zeigt sie als Beispiel. Die Liste des Standards (fehlende Rechte, Importfehler, Konflikte, abgelaufene Sitzungen) gehört zu einer Oberfläche; ein Eintrag steht nur für eine Meldung, die das Binary erzeugt. |
| 9 Version und Aktualität | gilt angepasst | Der Kopf trägt die Software-Version (die Ausgabe von `version` des gebauten Binary), den Gültigkeitsbereich (das Binary, das `make build` erzeugt) und den Stand (der Tag, an dem das Handbuch gegen das gebaute Binary der genannten Software-Version geprüft wurde, Format JJJJ-MM-TT). Es entfallen die Handbuch-Version (sie sagt nichts, was Software-Version und Stand nicht sagen, und ein Zählstand ist ein Verlauf), Autor oder Team (die Autorschaft hält die Versionsverwaltung; ein Name im Kopf wäre eine Zusage, die niemand pflegt), das Änderungsdatum (der Stand ersetzt es) und der Changelog oder die Änderungshistorie (ein Verlauf; was sich zwischen Versionen ändert, hält die Versionsverwaltung). `make handbuch-check` hält die Software-Version gegen das Binary und prüft die Form des Stands. |
| 10 Rechtliche und organisatorische Punkte | gilt angepasst | Ein Kapitel zu Geheimnissen, Sicherheitswarnungen oder Aufbewahrung steht, sobald das Binary einen solchen Gegenstand zeigt. Der Hinweis, keine Details zu nennen, die Angreifern helfen (interne Servernamen, Tokens, genaue Sicherheitskonfiguration), gilt unverändert. Lizenzhinweise und Haftungsausschlüsse trägt die Lizenz des Repos, nicht das Handbuch; Exportkontrolle und Compliance-Vorgaben gelten nicht. |
| 11 Barrierefreiheit berücksichtigen | gilt angepasst | Eine klare Überschriftenstruktur, sparsame und saubere Tabellen und keine Information nur über Farbe gelten. Alternativtexte, Kontraste und PDF-Lesezeichen entfallen: Das Handbuch enthält keine Bilder und liegt nur als Markdown vor; die Suche leistet der Betrachter. |
| 12 Wartbarkeit einplanen | gilt angepasst | Markdown im Repo gilt. Die Doku-Version ist über den Kopf an die Software-Version gekoppelt, und `make handbuch-check` prüft die Kopplung bei jedem `make gates`; das Review der Änderung ist der Review-Prozess. Die Pflege von Screenshots und eine benannte verantwortliche Person entfallen (Abschnitt 9). Veraltete Abschnitte werden entfernt, nicht markiert: eine Markierung wäre Chronik. |
| 13 Qualität prüfen | gilt angepasst | Das Gate belegt die Aufrufe, die Version und die Ziele von `make`; Kommandonamen, Schlüssel und Meldungen stimmen, wenn das Beispiel läuft. Tote Links fängt `make docs-check`. Button-Namen, Menüs, Screenshots, PDF- oder Online-Version und Suche entfallen. Der beste Test (eine Person ohne Vorwissen führt die Aufgaben allein nach dem Handbuch aus) gilt als Probelauf vor der Abnahme des Handbuchs. |
| 14 Empfohlene Vorlage | gilt angepasst | Der Kopf folgt Abschnitt 9. Die Kapitel der Vorlage gelten mit diesen Ausnahmen: „Anmeldung“ und „Überblick über die Oberfläche“ gelten nicht, ebenso „Rollen und Rechte“, „FAQ“, „Support und Kontakt“ und „Änderungshistorie“; „Import und Export“ steht nur, wenn das Binary eine solche Aufgabe zeigt; „Zweck der Software“ steht nur, soweit das gebaute Binary ihn zeigt. Der Absatz unter der Vorlage (je Version eine Zeile im letzten Kapitel) gilt nicht. |
