# repo.mk — die Make-Targets dieses Repos.
#
# Diese Datei gehoert dem Repo. Der Bootstrap legt sie an, wo sie fehlt, und
# schreibt sie danach nie wieder; harness/mk/ und das Root-Makefile dagegen
# schreibt jeder Lauf neu. Das Root-Makefile bindet sie nach den Fragmenten
# unter harness/mk/ ein (-include repo.mk); fehlt sie, laeuft make ohne sie.
#
# Ein eigenes Target steht hier wie in jedem Makefile. Ein eigenes Gate haengt
# sich an die Gate-Kette mit
#   GATE_CHECKS += <target>
# und laeuft dann in `make gates`, vor dem Gate-Nachweis. Jedes Target dieser
# Datei, Gate oder nicht, bekommt seine Zeile in harness/README.md, Abschnitt
# Sensors: das Doku-Gate (.d-check.yml, Modul targets) meldet ein Target ohne
# Zeile als gate-undocumented.
#
# Eine weitere Make-Datei, die diese Datei per include einbindet, liest das
# Doku-Gate nur im Wurzelverzeichnis (Glob "*.mk" in makefiles:). Grenze: liegt
# sie in einem Unterverzeichnis, laufen ihre Targets mit make, das Doku-Gate
# sieht sie nicht und meldet nichts.
#
# Eine Vorgabe fuer einen Marker der Fragmente (dort mit ?= belegt) steht hier
# mit einfachem = und gewinnt, weil diese Datei nach den Fragmenten gelesen wird.
