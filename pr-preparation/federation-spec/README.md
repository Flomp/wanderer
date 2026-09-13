# Vorschlag: Föderation als überprüfbarer Vertrag

Stand: 13. September 2026. **Entwurf zur Diskussion, keine beschlossene Architektur und keine Implementierung.** Die technischen Kapitel sind auf Englisch formuliert, damit sie als Grundlage für Maintainer-Reviews, Issues und spätere PRs dienen können.

Die Empfehlung ist, das Sollverhalten zuerst verbindlich zu beschreiben und danach innerhalb der bestehenden Anwendung klare Zuständigkeiten einzuführen. PocketBase, ActivityPub und Meilisearch können dabei erhalten bleiben. Die bereits vorbereiteten Sicherheitskorrekturen sollten unabhängig davon geprüft und veröffentlicht werden.

**Empfohlener nächster Umfang: nur P0–P2.** D-01 bis D-03 abstimmen, vorhandene Regressionstests zu wiederverwendbaren Fixtures bündeln und Import sowie Ausgabe klar trennen. P3–P8 sind ein möglicher späterer Ausbau, für den jeweils Nutzen, Aufwand und Betreuung geklärt werden müssen. Die vollständige Zielarchitektur und alle 52 Tests sind keine Voraussetzung für diesen ersten Schritt.

## Warum diese Vorschläge

Die untersuchten Fehler konzentrieren sich auf Grenzen, an denen heute mehrere Aufgaben zusammenfallen:

| Beobachtung | Daraus abgeleitetes Problem | Vorgeschlagene Grenze |
| --- | --- | --- |
| Ein PocketBase-Record trägt importierte Daten, interne Expands und die HTTP-Antwort | Herkunft und Berechtigung der Daten sind am Ausgabeort schwer erkennbar | Importmodell, gespeicherter Zustand, Suchprojektion und Antwort getrennt behandeln |
| „Full sync“ hängt von den angeforderten Expands ab; fehlgeschlagene Dateidownloads blockieren das Abschlussflag nicht | Vollständigkeit hat keinen stabilen fachlichen Inhalt | Feste Datenpakete; Dateiverfügbarkeit separat |
| Öffentliche Records, lokale Freigaben, Tokens, Remote-Akteure und interne Imports durchlaufen verschiedene Prüfungen | Ein korrekt geschützter Einstieg schützt nicht automatisch die übrigen Einstiege | Gemeinsame Zugriffsmatrix mit expliziter Umsetzung pro API-, Datei- und Suchpfad |
| Hintergrundarbeit verändert dieselben Records oder verarbeitet verspätete Daten | Ältere Arbeit kann eine neuere Entscheidung überschreiben | Stabile Antworten; später lokale Generationen und wiederaufnehmbare Arbeit |
| Empfang, Datenänderung, Feed und Versand sind eng verbunden | Wiederholung und Teilerfolg haben keine durchgängig sichtbaren Regeln | Atomare lokale Änderungen und dauerhafte Aufträge für externe Wirkungen |

Das ist eine Architekturdiagnose aus den geprüften Pfaden, kein Beweis, dass sämtliche Föderationsfehler dieselbe Ursache haben. Die anonymen Leseregeln und die manipulierbaren Share-Ziele waren auch eigenständige Fehler der lokalen Records-API.

Die Nutzerdokumentation beschreibt bereits öffentliche, ausschließlich lesbare Freigaben zwischen Instanzen. Lücken bestehen vor allem bei Cache-Alter, Vollständigkeit, Privatschaltung, Löschung, Wiederholungen und der Durchsetzung über alle Ausgabewege. Konkrete Belege und Grenzen der Analyse stehen in [08-evidence.md](08-evidence.md).

## Dokumente und Lesereihenfolge

| Dokument | Inhalt | Besonders relevant für |
| --- | --- | --- |
| [01 — Behavior and access](01-behavior-and-access.md) | Rollen, Zugriffsmatrix, Tokens, Kinder, Dateien, Suche, Statistik, Freigaben | Produktverhalten und Sicherheitsreview |
| [02 — Replica lifecycle](02-replica-lifecycle.md) | Zustandsmodell, Vollständigkeit, HTTP-Ergebnisse, Cache-Alter, Konkurrenz, Migration | Backend und Verfügbarkeit |
| [03 — Events and delivery](03-events-and-delivery.md) | Create/Update/Delete/Announce/Undo, Empfänger, Widerruf, Wiederholungen | Protokoll und Zustellung |
| [04 — Architecture and data](04-architecture-and-data.md) | Komponenten, Datenhoheit pro Feld, Schnittstellen, Transaktionen und Ausgabegrenzen | Schrittweises Refactoring |
| [05 — Acceptance tests](05-acceptance-tests.md) | Kleine Auswahl für P1/P2; Katalog mit 52 Szenarien für spätere Ausbauoptionen | Review und gezielte Abnahme |
| [06 — Rollout plan](06-rollout-plan.md) | Begrenzte erste Etappe P0–P2; optionaler Ausbau P3–P8 | Umsetzung und Aufwandseinschätzung |
| [07 — Decisions](07-decisions.md) | 16 offene Entscheidungen mit Empfehlung, Alternativen und Auswirkungen | Maintainer-Entscheidungen |
| [08 — Evidence](08-evidence.md) | Referenzstand, konkrete Codebelege, bestehende Dokumentation und Prüfgrenzen | Nachvollziehbarkeit |

Für eine erste Besprechung reichen diese Übersicht, die Matrix in Kapitel 01, das Zustandsmodell in Kapitel 02 und die Entscheidungen D-01 bis D-03. Für eine Implementierung werden zusätzlich die betroffenen Kapitel und Abnahmeszenarien benötigt.

## Empfohlenes Sollverhalten

Die folgenden Punkte beschreiben die längerfristige Zielrichtung. P0–P2 implementieren daraus die bestehenden Import-/Ausgabegrenzen; sie führen noch kein neues Cache-Zeitfenster, Zustandsmodell, Such-Gateway oder Zustellsystem ein.

### 1. Öffentliche Föderation klar von lokalen Berechtigungen trennen

Die erste Ausbaustufe bleibt auf öffentliche Remote-Inhalte beschränkt. Eine Signatur belegt, welcher Akteur eine Nachricht geschickt hat; sie erteilt keinen Zugriff auf private Trails. Eine Remote-Freigabe empfiehlt einen öffentlichen Inhalt an einen Empfänger. Lokale View-/Edit-Freigaben und Trail-Tokens sind eigene Berechtigungsmechanismen.

Jeder Trail einer Liste wird einzeln geprüft. Eine öffentliche Liste macht ihre privaten Einträge nicht öffentlich. Dasselbe gilt für Dateien, Expands, Suchergebnisse, Kartenbegrenzungen und Statistiken. Bei Beiträgen auf inzwischen privaten Trails muss separat entschieden werden, welche Verwaltungsrechte der jeweilige Beitragsautor behält.

### 2. Sichtbarkeit, Datenumfang und laufende Arbeit getrennt speichern

| Dimension | Vorgeschlagene Zustände | Aussage |
| --- | --- | --- |
| Öffentliche Verfügbarkeit | `unverified`, `public`, `withheld`, `deleted` | Welche belastbare Aussage über die öffentliche Verfügbarkeit vorliegt |
| Gespeicherter Datenumfang | `stub`, `metadata_ready`, `detail_ready` | Welche definierte Datenmenge vorhanden ist |
| Fetch-Auftrag | `idle`, `queued`, `running`, `backoff`, `failed` | Ob und wie gerade nachgeladen wird |
| Einzelne Datei | beispielsweise `missing`, `present`, `failed` | Ob genau diese Version der Datei vorliegt |

Ein gespeicherter GPX-Track macht einen privaten Trail nicht lesbar. Ein vollständiges Metadatenpaket verspricht nicht, dass sämtliche Fotos und Diskussionen offline verfügbar sind. Eine Liste ist vollständig geladen, wenn ihre definierte öffentliche Mitgliederprojektion vorliegt; dafür müssen nicht alle enthaltenen Trails vollständig heruntergeladen werden.

### 3. Verfügbarkeit bei Ausfällen ausdrücklich festlegen

Als Diskussionsgrundlage schlage ich vor: Nach **einer Stunde** eine erneute Prüfung anstoßen; bei vorübergehendem Ausfall einen zuvor öffentlich bestätigten Stand verwenden, solange die letzte Bestätigung **weniger als sieben Tage** zurückliegt. Ab sieben Tagen wird er bis zur erneuten Bestätigung ausgeblendet.

Das sind **neue, noch nicht akzeptierte Produktwerte**. Eine vertrauenswürdige Ablehnung durch den Ursprung wirkt sofort; das Zeitfenster erlaubt kein Ignorieren bekannter Privatschaltungen. Ein Actor-Refresh, eine eingehende Nachricht oder ein lokaler Datenbankeintrag verlängert die Bestätigung nicht. Der Nachteil ist ausdrücklich benannt: Dauerhaft verschwundene Ursprungsinstanzen verschwinden irgendwann auch aus den normalen Remote-Ansichten.

### 4. Privatschaltung und Löschung unterscheiden

`withheld` unterdrückt die Ausgabe, lässt aber eine spätere, neu bestätigte Veröffentlichung zu. `deleted` behält eine minimale Löschmarkierung und ist für dieselbe Objekt-IRI endgültig. Alte Fetches, Downloads und Nachrichten dürfen diese Entscheidungen nicht überschreiben.

Für kooperierende Wanderer-Instanzen beschreibt der Entwurf eine optionale, gesondert auszuhandelnde Withdrawal-Erweiterung. Ihr zusätzlicher Nutzen ist auf unterstützende Gegenstellen beschränkt; deshalb gehört sie nicht zur nächsten Etappe. Ältere und fremde Gegenstellen brauchen weiterhin Revalidierung oder ihre eigene Policy. Bereits ausgelieferte Kopien können nicht zurückgerufen werden. Die Details und die Abgrenzung zu ActivityPub-Update/Delete stehen in Kapitel 03 und D-05.

### 5. Lokale Änderung und externe Wirkung zuverlässig verbinden

Eine erfolgreiche lokale Änderung soll bedeuten, dass Daten und notwendige Aufträge gespeichert sind. Versand und Indexierung folgen durch wiederaufnehmbare Worker. Doppelte Zustellung darf keine doppelten Beziehungen oder Feed-Einträge erzeugen. Verspätete Arbeit muss prüfen, ob ihr zugrunde liegender Zustand noch gilt.

P2 beschränkt sich auf explizit erlaubte Importfelder und isolierte Antworten. Die Fehlertaxonomie aus ARC-011 gehört zu P3+. Dauerhafte Versandjobs und eine Protokollerweiterung sind ebenfalls spätere, getrennte Arbeitspakete.

## Was zuerst entschieden werden sollte

Alle 16 Punkte in [07-decisions.md](07-decisions.md) sind offen. Nicht alle müssen für das erste Refactoring beschlossen werden.

1. **D-01: Umfang.** Bleibt private Zusammenarbeit auf derselben Instanz? Empfehlung: ja, für diese Ausbaustufe.
2. **D-02: Vollständigkeit.** Welcher Datenumfang ist „synchronisiert“, und welche Dateien benötigt „offline verfügbar“?
3. **D-03: Ausfallverhalten.** Welche maximale Zeit ohne öffentliche Bestätigung ist akzeptabel? Vorschlag: sieben Tage mit den beschriebenen Verfügbarkeitseinbußen.
4. **D-07/D-10: Ausgabewege und Kompatibilität.** Wie werden Dateien, direkte Suchzugriffe, IDs, Counts und bestehende API-Antworten in denselben Vertrag aufgenommen?
5. **D-05/D-06/D-08: Lebenszyklus und Versand.** Welche Widerrufs- und Empfängeränderungen werden unterstützt, und welche Zustellgarantie wird versprochen?

D-04 sowie D-11 bis D-16 behandeln gezielt Rechte an eigenen Beiträgen, Aufbewahrung, Ressourcenlimits, Token-Schreibrechte, Profilprivatsphäre, Remote-Statistiken und die atomare lokale Listenfreigabe. Sie verhindern, dass solche Produktentscheidungen unbemerkt in einem Refactoring getroffen werden.

## Konkreter Umsetzungsvorschlag

1. **P0:** D-01 bis D-03 abstimmen und vorhandenes Verhalten von späteren Produktänderungen trennen. Die isolierten Sicherheitskorrekturen bleiben unabhängig.
2. **P1:** Bestehende Regressionstests mit echtem Schema, registrierten Hooks und tatsächlichen Routen als kleine Fixture-Grundlage nutzen.
3. **P2:** Erlaubte Importfelder explizit machen und Antwortprojektionen isolieren. Erfolgreiche API-Antworten und vorhandene Zugriffsprüfungen erhalten; bewusste Fehlerkorrekturen separat benennen und testen.
4. Danach den Nutzen prüfen und nur ausgewählte Folgepakete beauftragen.

Wanderer wird hier als ein Backend-Prozess mit SQLite betrachtet. Auch darin können HTTP-Anfragen und Goroutines konkurrieren. Dafür braucht es stabile Antworten und gegebenenfalls lokale Zusammenführung paralleler Fetches; bei späteren Jobs ist Wiederaufnahme nach Neustart entscheidend. Verteilte Worker und ablaufende Mehrprozess-Leases sind keine Anforderung dieser Etappe.

Die Suche benötigt eine gesonderte Aufwandsschätzung: Der Browser nutzt bereits die SvelteKit-Routen unter `/api/v1/search/...`; diese fragen Meilisearch mit einem Tenant-Token ab. Der vorhandene Proxy bietet einen Ansatzpunkt für die Erweiterung. Ein sofort wirksamer Replica-Status verlangt aber zusätzliche Prüfungen für Treffer, Counts, Facetten und Kartenabfragen. Außerdem müssen mögliche Direktzugriffe auf Meilisearch berücksichtigt werden: Der Token ist im Browser lesbar; ob er den Proxy umgehen kann, hängt von der Erreichbarkeit des Suchdienstes ab. Ein asynchrones Re-Indexieren liefert keine sofortige Sperrgarantie.

[Kapitel 06](06-rollout-plan.md) enthält die konkreten Grenzen der ersten Etappe und die späteren Optionen. Neue Zwei-Instanzen-Infrastruktur wird erst für entsprechende Föderationsänderungen benötigt. P3–P8 zusammen sind ein umfangreiches Folgeprogramm, kein bereits vereinbarter Arbeitsplan; eine belastbare Zeitplanung setzt eine Auswahl und benannte Betreuung voraus.

## Status und Verwendung

- `EXISTING` bezeichnet belegtes Verhalten am angegebenen Referenzstand.
- `PROPOSED` und normative `MUST`/`SHOULD` beschreiben den vorgeschlagenen Zielvertrag, keine bereits vorhandene Garantie.
- `OPEN` und D-01…D-16 markieren noch erforderliche Entscheidungen. Eine Empfehlung ist keine erteilte Freigabe.
- Die T-001…T-052 sind spezifizierte Abnahmeszenarien, keine in diesem Auftrag implementierten oder erfolgreich ausgeführten Tests.

Referenz ist der Integrationscommit `865251d49dc6c1d734cccb85ee51c680c62d35a6`, gesichert durch den Tag `spec/federation-baseline-2026-09-13`. Wie dieser Belegstand abgerufen wird, steht in [08-evidence.md](08-evidence.md). Die spätere deterministische Hintergrund-Sync-Regression aus `ddd1929d8` wird gesondert berücksichtigt. Der Dokumentationsbranch basiert auf `dev`; der analysierte Integrationsstand enthält zusätzlich die geprüften Fixes. Die Dokumente beschreiben Vorschläge und enthalten keine Laufzeitimplementierung.
