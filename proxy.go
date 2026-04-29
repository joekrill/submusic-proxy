package sanitizer

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

// TODO: intercept and change the URL parameter "c" to be more consistent (unchanging between SubMusic versions

type SanitizerProxy struct {
	logger *slog.Logger
	proxy  *httputil.ReverseProxy
}

type SanitizerProxyConfig = func(*SanitizerProxy) error

func WithLogger(logger *slog.Logger) SanitizerProxyConfig {
	return func(app *SanitizerProxy) error {
		app.logger = logger
		return nil
	}
}

func NewSanitizerProxy(targetUrl *url.URL, options ...SanitizerProxyConfig) (*SanitizerProxy, error) {
	p := &SanitizerProxy{}

	for _, op := range options {
		err := op(p)
		if err != nil {
			return nil, err
		}
	}

	if p.logger == nil {
		p.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	p.proxy = httputil.NewSingleHostReverseProxy(targetUrl)
	p.proxy.ModifyResponse = p.modifyResponse()
	p.proxy.ErrorHandler = p.errorHandler()

	return p, nil
}

func (s *SanitizerProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	s.proxy.ServeHTTP(rw, req)
}

func (s SanitizerProxy) errorHandler() func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, req *http.Request, err error) {
		s.logger.Warn(fmt.Sprintf("error proxying response: %v", err))
	}
}

// Decodes the response body, handling any content-encoding
func (s SanitizerProxy) decodeResponseBody(resp *http.Response) ([]byte, error) {
	var reader io.ReadCloser
	var err error

	contentEncoding := resp.Header.Get("Content-Encoding")
	s.logger.Debug("decoding response body", "content-encoding", contentEncoding)
	switch contentEncoding {
	case "gzip", "x-gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to decode gzip-compressed body: %w", err)
		}
	case "br":
		brReader := brotli.NewReader(resp.Body)
		reader = io.NopCloser(brReader)
	case "zstd":
		decoder, err := zstd.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to decode zstd-compressed body: %w", err)
		}
		reader = io.NopCloser(decoder)
	case "":
		reader = resp.Body
	default:
		return nil, fmt.Errorf("unsupported Content-Encoding: %s", contentEncoding)
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// REVIEW: Do we need to close resp.Body explicitly?
	defer reader.Close()

	return body, err
}

func (s SanitizerProxy) modifyResponse() func(*http.Response) error {
	return func(resp *http.Response) error {
		path := resp.Request.URL.Path

		if !CanSanitize(path) {
			s.logger.Debug("skipping unknown path", "path", path)
			return nil
		}

		// SubMusic passes a client ID in the URL string, which we check so that
		// we only sanitize responses coming from SubMusic clients
		clientId := resp.Request.URL.Query().Get("c")
		if !strings.HasPrefix(clientId, "SubMusic") {
			s.logger.Debug("skipping unknown client", "client-id", clientId)
			return nil
		}

		s.logger.Debug("processing path", "path", path, "client-id", clientId)

		body, err := s.decodeResponseBody(resp)
		if err != nil {
			s.logger.Error(err.Error(), "path", path, "error", err)
			return nil
		}
		s.logger.Debug("decoded response body", "path", path, "length", len(body))

		sanitizedBody, err := sanitizeResponse(path, body)

		if err != nil {
			s.logger.Error("failed to marshal sanitized response", "path", path, "error", err)
			return nil
		}

		originalSize := int64(len(body))
		sanitizedSize := int64(len(sanitizedBody))
		reduction := ((float64(originalSize) - float64(sanitizedSize)) / float64(originalSize)) * 100

		s.logger.Info(fmt.Sprintf("sanitized %s from %d to %d bytes (%.1f%%)", path, originalSize, sanitizedSize, reduction),
			"path", path,
			"original-size", originalSize,
			"sanitized-size", sanitizedSize,
			"reduction", reduction,
		)

		resp.Body = io.NopCloser(bytes.NewReader(sanitizedBody))
		resp.ContentLength = sanitizedSize

		// TODO: Recompress?
		resp.Header.Del("Content-Encoding")

		return nil
	}
}
