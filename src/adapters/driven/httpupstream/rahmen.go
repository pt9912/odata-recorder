package httpupstream

import "net/http"

// rahme setzt die Rahmung nach Laenge (SPEC-120 Nr. 4). Bei 0 bleibt Body nil: die Bibliothek
// liest nichts und setzt `Content-Length: 0` selbst bei POST, PUT und PATCH. Bei -1 steht
// `chunked` im Request, bei jeder Methode. Sonst geht die Länge als Content-Length, und die
// Bibliothek liest bis io.EOF und prüft die Länge.
func rahme(req *http.Request, laenge int64, l *leser) {
	if laenge == 0 {
		return
	}
	req.Body = l
	req.ContentLength = laenge
	if laenge < 0 {
		req.TransferEncoding = []string{"chunked"}
	}
}
