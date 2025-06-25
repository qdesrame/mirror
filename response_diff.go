package mirror

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"time"
)

const (
	ContentTypeJson = "application/json"
	ContentTypeXML  = "application/xml"
)

type Diff[T comparable] struct {
	Main   T
	Shadow T
}

type DiffMap[K comparable, V any] struct {
	Main   map[K]V
	Shadow map[K]V
}

type DiffMapResult[K comparable, V any] struct {
	Added   []MapEntry[K, V]
	Missing []MapEntry[K, V]
	Changed map[K]DiffValue[V]
	Equals  []MapEntry[K, V]
}

type DiffArrayResult[V any] DiffMapResult[int, V]

type DiffValue[V any] struct {
	Main   V
	Shadow V
}

type MapEntry[K comparable, V any] struct {
	Key   K
	Value V
}

func (d Diff[T]) Compare() bool {
	return d.Main != d.Shadow
}

type ResponseDiffResult struct {
	Status Diff[int]
	Proto  Diff[string]
	Body   BodyDiff
	Header HeaderDiffResult
}

type BodyDiff struct {
	Main              []byte
	MainContentType   string
	Shadow            []byte
	ShadowContentType string
}

type BodyDiffResult struct {
	Equals               bool
	JsonObjectDiffResult DiffMapResult[string, any]
	JsonArrayDiffResult  DiffArrayResult[any]
	XMLDiffResult        DiffMapResult[string, any]
}

func (b BodyDiff) Compare() (BodyDiffResult, error) {
	if len(b.Main) == 0 && len(b.Shadow) == 0 {
		return BodyDiffResult{Equals: true}, nil
	}

	result := BodyDiffResult{}

	if b.MainContentType == b.ShadowContentType {
		// TODO Here you would implement the logic to compare JSON or XML bodies.
		// For now, we just return an empty result.
		switch b.MainContentType {
		case ContentTypeJson:
			result.JsonObjectDiffResult = DiffMapResult[string, any]{}
			result.JsonArrayDiffResult = DiffArrayResult[any]{}
		case ContentTypeXML:
			result.XMLDiffResult = DiffMapResult[string, any]{}
		}
	} else {
		result.Equals = false
	}

	return result, nil
}

type HeaderDiff struct {
	Main   http.Header
	Shadow http.Header
}

type HeaderDiffResult = DiffMapResult[string, []string]

func (h HeaderDiff) Compare() (HeaderDiffResult, error) {
	if h.Main == nil || h.Shadow == nil {
		return HeaderDiffResult{}, nil
	}

	result := HeaderDiffResult{
		Added:   []MapEntry[string, []string]{},
		Missing: []MapEntry[string, []string]{},
		Changed: make(map[string]DiffValue[[]string]),
		Equals:  []MapEntry[string, []string]{},
	}

	for k, v := range h.Main {
		if sv, ok := h.Shadow[k]; ok {
			if len(v) != len(sv) || !slices.Equal(v, sv) {
				result.Changed[k] = DiffValue[[]string]{Main: v, Shadow: sv}
			} else {
				result.Equals = append(result.Equals, MapEntry[string, []string]{Key: k, Value: v})
			}
		} else {
			result.Added = append(result.Added, MapEntry[string, []string]{Key: k, Value: v})
		}
	}

	for k := range h.Shadow {
		if _, ok := h.Main[k]; !ok {
			result.Missing = append(result.Missing, MapEntry[string, []string]{Key: k, Value: h.Shadow[k]})
		}
	}

	return result, nil
}

type Request struct {
	Proto  string
	Method string
	Url    string
	Path   string
	Header http.Header
	Body   []byte
}
type Response struct {
	Proto  string
	Status string
	Code   int
	Header http.Header
	Body   []byte
}

func NewResponse(recorder *httptest.ResponseRecorder) Response {
	return Response{
		Proto:  recorder.Result().Proto,
		Status: recorder.Result().Status,
		Code:   recorder.Code,
		Header: recorder.Header(),
		Body:   recorder.Body.Bytes(),
	}
}

// Capture is our traffic data.
type Capture struct {
	ID  int
	Req Request
	Res Response
	// Elapsed time of the request, in milliseconds.
	Elapsed time.Duration
}

type (
	ResponseDiff struct {
	}
	ResponseDiffOpt struct {
		IgnoreHeaders []string
	}
)

func NewResponseDiff(opts ...ResponseDiffOpt) (ResponseDiffResult, error) {
	return ResponseDiffResult{}, nil
}

func (d ResponseDiff) Compare(response1 Response, response2 Response) (ResponseDiffResult, error) {
	//if response1 == nil || response2 == nil {
	//	return ResponseDiffResult{}, fmt.Errorf("nil response")
	//}

	if response1.Code != response2.Code {
		slog.Info("diff response code", "main", response1.Status, "shadow", response2.Status)
	}

	// Compare the two responses and return the differences
	// This is a placeholder for the actual comparison logic
	return ResponseDiffResult{}, nil
}
