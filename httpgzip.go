/*
gziphandler provides an http handler wrapper to transparently
 add gzip compression. It will sniff the content type based on the
 uncompressed data if necessary.
*/
package gziphandler

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

var compressibleMimes = map[string]bool{
	// Compressible types from https://www.fastly.com/blog/new-gzip-settings-and-deciding-what-compress
	"text/html":                     true,
	"application/x-javascript":      true,
	"text/css":                      true,
	"application/javascript":        true,
	"text/javascript":               true,
	"text/plain":                    true,
	"text/xml":                      true,
	"application/json":              true,
	"application/vnd.ms-fontobject": true,
	"application/x-font-opentype":   true,
	"application/x-font-truetype":   true,
	"application/x-font-ttf":        true,
	"application/xml":               true,
	"font/eot":                      true,
	"font/opentype":                 true,
	"font/otf":                      true,
	"image/svg+xml":                 true,
	"image/vnd.microsoft.icon":      true,
	// other types
	"application/vnd.geo+json": true,
	"application/cap+xml":      true,
	"text/csv":                 true,
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", http.DetectContentType(b))
	}

	// Find the base content type (remove any trailing parameters like the version)
	contentType := w.Header().Get("Content-Type")
	i := strings.Index(contentType, ";")
	if i > 0 {
		contentType = contentType[0:i]
	}

	contentType = strings.TrimSpace(contentType)

	switch compressibleMimes[contentType] {
	case true:
		w.Header().Set("Content-Encoding", "gzip")
		return w.Writer.Write(b)
	default:
		return w.ResponseWriter.Write(b)
	}

}

// Wrap a http.Handler to support transparent gzip encoding.
func GzipHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept-Encoding")

		switch strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		case true:
			gz := gzip.NewWriter(w)
			defer gz.Close()
			h.ServeHTTP(&gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
			return
		default:
			h.ServeHTTP(w, r)
			return
		}
	})
}
