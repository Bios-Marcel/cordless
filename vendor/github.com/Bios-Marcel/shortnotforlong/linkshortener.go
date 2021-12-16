package linkshortener

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Shortener offers a function to shorten a URL and redirect to the shortened
// URL as soon as a request comes in.
type Shortener struct {
	nextFreeIndex uint16
	shortenedUrls map[uint16]string
	port          int
	httpServer    *http.Server
}

// Shorten takes a url and returns a shortend version that redirects via the
// local webserver.
func (shortener *Shortener) Shorten(url string) (string, string) {
	suffix := getSuffix(url)
	for id, address := range shortener.shortenedUrls {
		if address == url {
			return shortener.formatShortenedURL(id), suffix
		}
	}

	newID := shortener.popNextIndex()
	shortener.shortenedUrls[newID] = url

	return shortener.formatShortenedURL(newID), suffix
}

func (shortener *Shortener) formatShortenedURL(id uint16) string {
	return fmt.Sprintf("http://localhost:%d/%d", shortener.port, id)
}

func getSuffix(url string) string {
	if strings.Count(url, "/") > 2 {
		return filepath.Ext(url)
	}
	return ""
}

// CalculateShortenedLength returns the length of the shortened URL without
// a suffix and the length of the suffix, which is 0 if the suffix isn't available.
func (shortener *Shortener) CalculateShortenedLength(url string) (int, int) {
	id := shortener.peekNextID()
	return len(shortener.formatShortenedURL(id)), len(getSuffix(url))
}

func (shortener *Shortener) popNextIndex() uint16 {
	shortener.nextFreeIndex = shortener.peekNextID()
	return shortener.nextFreeIndex
}

func (shortener *Shortener) peekNextID() uint16 {
	if shortener.nextFreeIndex >= math.MaxUint16 {
		return 0
	}
	return shortener.nextFreeIndex + 1
}

//RedirectHandler handles all the redirects for the Server.
type RedirectHandler struct {
	Shortener *Shortener
}

func (h RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var id = ""
	var idRegex = regexp.MustCompile(`/(\d*)`)
	var matches = idRegex.FindStringSubmatch(r.URL.Path)
	if len(matches) > 1 {
		id = matches[1]
	}

	idAsInt, convertError := strconv.ParseUint(id, 10, 16)
	if convertError != nil {
		http.NotFound(w, r)
	} else {
		url, contains := h.Shortener.shortenedUrls[uint16(idAsInt)]
		if contains {
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		} else {
			http.NotFound(w, r)
		}
	}
}

// Close closes the internal http server.
func (shortener *Shortener) Close() {
	shortener.httpServer.Shutdown(context.Background())
}

//NewShortener creates a new server that uses the given port.
func NewShortener(port int) *Shortener {
	shortener := &Shortener{
		shortenedUrls: make(map[uint16]string),
		port:          port,
	}

	handler := RedirectHandler{
		Shortener: shortener,
	}

	httpServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        handler,
		ReadTimeout:    1 * time.Second,
		WriteTimeout:   1 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	shortener.httpServer = httpServer

	return shortener
}

//Start servers the internal http server, blocks and returns an error on
//failure.
func (shortener *Shortener) Start() error {
	return shortener.httpServer.ListenAndServe()
}
