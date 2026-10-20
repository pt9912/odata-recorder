**Vorgang:** slice-gegenproben-build-test
**Fund:** Die Verdrahtung der Dockerfile-Stufen `test`, `build` und `pruefung` mit `pruefung` hat keinen dauerhaften Fänger (`RUN true` bleibt grün), ebenso der Eintrag `tools/pruefung` in der Prüfsumme; beides steht als Grenze in `SPEC-036` und `SPEC-102`.
**Stand:** offen
