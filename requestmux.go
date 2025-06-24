package handler

import (
	"bytes"
	"fmt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"sync"
)

func RequestMux(mainTarget *httputil.ReverseProxy, shadowTarget *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Configure the "http.route" for the HTTP instrumentation.
		otelhttp.WithRouteTag(r.URL.Path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read request body", http.StatusInternalServerError)
				return
			}
			rShadow := r.Clone(r.Context())

			//rww := NewResponseWriterWrapper()

			// clone body
			r.Body = io.NopCloser(bytes.NewBuffer(body))
			rShadow.Body = io.NopCloser(bytes.NewBuffer(body))

			respMainRecorder := httptest.NewRecorder()
			respShadowRecorder := httptest.NewRecorder()
			//respMainRecorder := NewResponseWriterWrapper()
			//respShadowRecorder := NewResponseWriterWrapper()

			// proxy request to the target
			wg := sync.WaitGroup{}
			wg.Add(2)

			go func() {
				mainTarget.ServeHTTP(respMainRecorder, r)
				wg.Done()
			}()
			go func() {
				shadowTarget.ServeHTTP(respShadowRecorder, rShadow)
				wg.Done()
			}()
			wg.Wait()
			go func() {
				var err error
				//respMain, err := http.ReadResponse(bufio.NewReader(respMainRecorder.body), r)
				//if err != nil {
				//	slog.Warn("read response from main", "error", err)
				//}
				//respShadow, err := http.ReadResponse(bufio.NewReader(respShadowRecorder.body), rShadow)
				//if err != nil {
				//	slog.Warn("read response from shadow", "error", err)
				//}

				respMain := NewResponse(respMainRecorder)
				resShadow := NewResponse(respShadowRecorder)

				diff, err := ResponseDiff{}.Compare(respMain, resShadow)
				if err != nil {
					slog.Error("response diff", "error", err)
				}
				slog.Info("response diff", "diff", diff)
			}()

			w.WriteHeader(respMainRecorder.Code)
			_, err = w.Write(respMainRecorder.Body.Bytes())
			if err != nil {
				slog.Error("request mux", "error", err)
			}
		})).ServeHTTP(w, r)

	}
}

// ResponseWriterWrapper struct is used to log the response
type ResponseWriterWrapper struct {
	body       *bytes.Buffer
	statusCode int
	header     http.Header
}

// NewResponseWriterWrapper static function creates a wrapper for the http.ResponseWriter
func NewResponseWriterWrapper() *ResponseWriterWrapper {
	var buf bytes.Buffer
	var statusCode = 200
	return &ResponseWriterWrapper{
		body:       &buf,
		statusCode: statusCode,
		header:     make(http.Header),
	}
}

func (rww *ResponseWriterWrapper) Write(buf []byte) (int, error) {
	size, err := rww.body.Write(buf)

	//if rww.w != nil {
	//	size, err = rww.w.Write(buf)
	//}

	return size, err
}

// Header function overwrites the http.ResponseWriter Header() function
func (rww *ResponseWriterWrapper) Header() http.Header {
	//if rww.w != nil {
	//	return rww.w.Header()
	//}

	return rww.header
}

// WriteHeader function overwrites the http.ResponseWriter WriteHeader() function
func (rww *ResponseWriterWrapper) WriteHeader(statusCode int) {
	rww.statusCode = statusCode
	//rww.w.WriteHeader(statusCode)
}

func (rww *ResponseWriterWrapper) String() string {
	var buf bytes.Buffer

	buf.WriteString("Response:")

	buf.WriteString("Headers:")
	for k, v := range rww.Header() {
		buf.WriteString(fmt.Sprintf("%s: %v", k, v))
	}

	buf.WriteString(fmt.Sprintf(" Status Code: %d", rww.statusCode))

	buf.WriteString("Body")
	buf.WriteString(rww.body.String())
	return buf.String()
}
