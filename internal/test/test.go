package test

import (
	"bytes"
	"context"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-checkout/internal/dispatcher/common"
	httpEcho "github.com/paysuper/paysuper-checkout/pkg/http"
	"github.com/stretchr/testify/assert"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// EchoReqResCaller
type EchoReqResCaller struct {
	dispatcher      httpEcho.Dispatcher
	middlewareSetUp *MiddlewareTestUp
}

// NewTestRequest
func NewTestRequest(dispatcher httpEcho.Dispatcher, mw *MiddlewareTestUp) *EchoReqResCaller {
	return &EchoReqResCaller{
		dispatcher:      dispatcher,
		middlewareSetUp: mw,
	}
}

type Middleware interface {
	Post(middleware ...echo.MiddlewareFunc)
	Pre(middleware ...echo.MiddlewareFunc)
}

type MiddlewareTestUp struct {
	preMiddleware []echo.MiddlewareFunc
	useMiddleware []echo.MiddlewareFunc
}

type QueryBuilder struct {
	c            *EchoReqResCaller
	method, path string
	init         []func(request *http.Request, middleware Middleware)
	params       []string
	query        url.Values
	body         io.Reader
	cookie       []*http.Cookie
	headers      map[string]string
}

// Params
func (s *QueryBuilder) Params(params ...string) *QueryBuilder {
	s.params = append(s.params, params...)
	return s
}

// SetCookie
func (s *QueryBuilder) AddCookie(cookie *http.Cookie) *QueryBuilder {
	s.cookie = append(s.cookie, cookie)
	return s
}

// SetQueryParam
func (s *QueryBuilder) SetQueryParam(key, value string) *QueryBuilder {
	s.query.Set(key, value)
	return s
}

// SetQueryParams
func (s *QueryBuilder) SetQueryParams(values url.Values) *QueryBuilder {
	for key, value := range values {
		for _, val := range value {
			s.query.Add(key, val)
		}
	}
	return s
}

// Init
func (s *QueryBuilder) Init(init func(*http.Request, Middleware)) *QueryBuilder {
	s.init = append(s.init, init)
	return s
}

// BodyString
func (s *QueryBuilder) BodyString(body string) *QueryBuilder {
	s.body = strings.NewReader(body)
	return s
}

// BodyBytes
func (s *QueryBuilder) BodyBytes(body []byte) *QueryBuilder {
	s.body = bytes.NewReader(body)
	return s
}

// Method
func (s *QueryBuilder) Method(method string) *QueryBuilder {
	s.method = method
	return s
}

// Path
func (s *QueryBuilder) Path(path string) *QueryBuilder {
	s.path = path
	return s
}

// Body
func (s *QueryBuilder) Body(body io.Reader) *QueryBuilder {
	s.body = body
	return s
}

// AddHeader
func (s *QueryBuilder) AddHeader(name, value string) *QueryBuilder {
	if len(s.headers) == 0 {
		s.headers = make(map[string]string)
	}

	s.headers[name] = value
	return s
}

// SetHeaders
func (s *QueryBuilder) SetHeaders(headers map[string]string) *QueryBuilder {
	if len(headers) > 0 {
		for name, value := range headers {
			s.AddHeader(name, value)
		}
	}

	return s
}

// Exec
func (s *QueryBuilder) Exec(t *testing.T) (*httptest.ResponseRecorder, error) {

	r := strings.NewReplacer(s.params...)
	s.path = r.Replace(s.path)

	u, err := url.Parse(s.path)
	assert.NoError(t, err)

	u.RawQuery = s.query.Encode()

	return s.c.Request(s.method, u.String(), s.body, func(request *http.Request, middleware Middleware) {
		for _, cookie := range s.cookie {
			request.AddCookie(cookie)
		}
		for _, init := range s.init {
			init(request, middleware)
		}
		for name, value := range s.headers {
			request.Header.Set(name, value)
		}
	})
}

// ExecFileUpload
func (s *QueryBuilder) ExecFileUpload(t *testing.T, params map[string]string, paramName, path string) (*httptest.ResponseRecorder, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := file.Close(); err != nil {
			return
		}
	}()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))

	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	err = writer.Close()

	if err != nil {
		return nil, err
	}

	s.Init(func(request *http.Request, middleware Middleware) {
		request.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	})

	s.Method(http.MethodPost)
	s.Body(body)

	return s.Exec(t)
}

// NewQueryBuilder
func NewQueryBuilder(c *EchoReqResCaller) *QueryBuilder {
	return &QueryBuilder{c: c, query: make(url.Values)}
}

// Post
func (m *MiddlewareTestUp) Post(middleware ...echo.MiddlewareFunc) {
	m.useMiddleware = middleware
}

// Pre
func (m *MiddlewareTestUp) Pre(middleware ...echo.MiddlewareFunc) {
	m.preMiddleware = middleware
}

// ListUse
func (m *MiddlewareTestUp) ListUse() []echo.MiddlewareFunc {
	return m.useMiddleware
}

// ListPre
func (m *MiddlewareTestUp) ListPre() []echo.MiddlewareFunc {
	return m.preMiddleware
}

type MiddlewareHandlerFunc func(echo.Context) (echo.Context, error)

// Request
func (c *EchoReqResCaller) Request(method, target string, body io.Reader, init func(*http.Request, Middleware)) (resRec *httptest.ResponseRecorder, err error) {
	he := echo.New()
	req := httptest.NewRequest(method, target, body)
	if init == nil {
		panic("request init function should be present")
	}
	init(req, c.middlewareSetUp)
	he.Pre(c.middlewareSetUp.ListPre()...)
	he.Use(c.middlewareSetUp.ListUse()...)
	resRec = httptest.NewRecorder()
	if err = c.dispatcher.Dispatch(he); err != nil {
		return
	}
	//
	he.HTTPErrorHandler = func(e error, context echo.Context) {
		err = e
		he.DefaultHTTPErrorHandler(e, context)
	}
	//
	he.ServeHTTP(resRec, req)
	return
}

// Send
func (c *EchoReqResCaller) Builder() *QueryBuilder {
	return NewQueryBuilder(c)
}

// DefaultSettings
func DefaultSettings() map[string]interface{} {
	return map[string]interface{}{
		"dispatcher": map[string]interface{}{
			"global": map[string]interface{}{
				"paymentFormJsLibraryUrl":      "unknown",
				"customerTokenCookiesLifetime": "2592000s",
				"CookieDomain":                 "localhost",
				"orderInlineFormUrlMask":       "http://localhost",
			},
		},
	}
}

// SetUp
func SetUp(settings map[string]interface{}, services common.Services, setUp func(*TestSet, Middleware) common.Handlers) (*EchoReqResCaller, error) {
	middlewareSetUp := &MiddlewareTestUp{}
	testSet, _, e := BuildTestSet(
		context.Background(),
		settings,
		services,
		nil,
	)
	if e != nil {
		return nil, e
	}
	d, _, e := BuildDispatcher(
		context.Background(),
		settings,
		services,
		setUp(testSet, middlewareSetUp),
		nil,
	)
	if e != nil {
		return nil, e
	}
	return NewTestRequest(d, middlewareSetUp), e
}

// ReqInitJSON
func ReqInitJSON() func(request *http.Request, middleware Middleware) {
	return func(request *http.Request, middleware Middleware) {
		request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
}

// ReqInitMultipartForm
func ReqInitMultipartForm() func(request *http.Request, middleware Middleware) {
	return func(request *http.Request, middleware Middleware) {
		request.Header.Set(echo.HeaderContentType, echo.MIMEMultipartForm)
	}
}

// ReqInitApplicationForm
func ReqInitApplicationForm() func(request *http.Request, middleware Middleware) {
	return func(request *http.Request, middleware Middleware) {
		request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	}
}

// ReqInitXML
func ReqInitXML() func(request *http.Request, middleware Middleware) {
	return func(request *http.Request, middleware Middleware) {
		request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationXML)
	}
}
