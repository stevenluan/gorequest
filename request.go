package gorequest

import (
	"encoding/json"
	golog "log"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"

	"golang.org/x/net/publicsuffix"
)

// Client represents the client with timeout and retry ability
// gorequest dose not provide retry, and it is not thread safe
// so this is the wrapper to add those missing functionality
// the actual http request is still being made by gorequest module.
type Client interface {
	Request() HTTPRequest
	Timeout(timeout time.Duration) Client
	Retry(
		maxRetries int,
		backoff float64,
		shouldRetry ...func(*http.Response, []byte, []error) bool,
	) Client
}

// HTTPRequest represents the http request. The methods are mostly from SuperAgent.
type HTTPRequest interface {
	Get(targetURL string) HTTPRequest
	Post(targetURL string) HTTPRequest
	Head(targetURL string) HTTPRequest
	Put(targetURL string) HTTPRequest
	Delete(targetURL string) HTTPRequest
	Patch(targetURL string) HTTPRequest
	Set(param string, value string) HTTPRequest
	SetHeaders(headers map[string]string) HTTPRequest
	Type(typeStr string) HTTPRequest
	Query(content interface{}) HTTPRequest
	Send(content interface{}) HTTPRequest
	End() (*http.Response, string, []error)
	EndBytes() (resp *http.Response, body []byte, errors []error)
	EndStruct(content interface{}) (*http.Response, []byte, []error)
	SetDebug(enabled bool) HTTPRequest
}

// clientImpl is the data structure to hold client configuration.
type clientImpl struct {
	client           *http.Client
	transport        *http.Transport
	maxRetries       int
	backoff          float64
	minRetryPeriod   time.Duration
	retryPeriodRange time.Duration
	shouldRetry      func(*http.Response, []byte, []error) bool
}

// New creates new client object.
// This should be called once only to reuse the http transport.
func New() Client {
	c := &clientImpl{
		client:           newHTTPClient(),
		transport:        &http.Transport{},
		maxRetries:       0,
		backoff:          float64(0),
		minRetryPeriod:   100 * time.Millisecond,
		retryPeriodRange: 200 * time.Millisecond,
		shouldRetry:      defaultShouldRetry,
	}
	return c
}

// Retry sets the retry configuration
func (c *clientImpl) Retry(
	maxRetries int,
	backoff float64,
	shouldRetry ...func(*http.Response, []byte, []error) bool,
) Client {
	c.maxRetries = maxRetries
	c.backoff = backoff
	if len(shouldRetry) > 0 {
		c.shouldRetry = shouldRetry[0]
	}

	return c
}

// Timeout sets the timeout configuration
func (c *clientImpl) Timeout(timeout time.Duration) Client {
	c.transport.Dial = func(network, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(network, addr, timeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(timeout))
		return conn, nil
	}
	return c
}

// HTTPRequest starts single http request
// This can be called multiple times
// Each returned request should be only used in single goroutine.
func (c *clientImpl) Request() HTTPRequest {
	return newRequest(
		c.client,
		c.transport,
		c.maxRetries,
		c.backoff,
		c.minRetryPeriod,
		c.retryPeriodRange,
		c.shouldRetry,
	)
}

// requestImpl is the data structure for single http request.
type requestImpl struct {
	superAgent       *SuperAgent
	maxRetries       int
	backoff          float64
	client           *http.Client
	transport        *http.Transport
	minRetryPeriod   time.Duration
	retryPeriodRange time.Duration
	shouldRetry      func(*http.Response, []byte, []error) bool
}

// newRequest creates object of HTTPRequest type
func newRequest(
	client *http.Client,
	transport *http.Transport,
	maxRetries int,
	backoff float64,
	minRetryPeriod time.Duration,
	retryPeriodRange time.Duration,
	shouldRetry func(*http.Response, []byte, []error) bool,
) HTTPRequest {
	superAgent := newSuperAgent(client, transport)
	r := &requestImpl{
		client:           client,
		transport:        transport,
		maxRetries:       maxRetries,
		backoff:          backoff,
		superAgent:       superAgent,
		minRetryPeriod:   minRetryPeriod,
		retryPeriodRange: retryPeriodRange,
		shouldRetry:      shouldRetry,
	}

	return r
}

// Get makes http Get request
func (r *requestImpl) Get(targetURL string) HTTPRequest {
	r.superAgent.Get(targetURL)
	return r
}

// Post makes http Post request
func (r *requestImpl) Post(targetURL string) HTTPRequest {
	r.superAgent.Post(targetURL)
	return r
}

// Put makes http Put request
func (r *requestImpl) Put(targetURL string) HTTPRequest {
	r.superAgent.Put(targetURL)
	return r
}

// Head makes http Head request
func (r *requestImpl) Head(targetURL string) HTTPRequest {
	r.superAgent.Head(targetURL)
	return r
}

// Delete makes http Delete request
func (r *requestImpl) Delete(targetURL string) HTTPRequest {
	r.superAgent.Delete(targetURL)
	return r
}

// Patch makes http Patch request
func (r *requestImpl) Patch(targetURL string) HTTPRequest {
	r.superAgent.Patch(targetURL)
	return r
}

// Set add http header to the request
func (r *requestImpl) Set(param string, value string) HTTPRequest {
	r.superAgent.Set(param, value)
	return r
}

// SetHeaders add http header to the request
func (r *requestImpl) SetHeaders(headers map[string]string) HTTPRequest {
	for key, value := range headers {
		r.superAgent.Set(key, value)
	}
	return r
}

// Type sets the response type of the request
func (r *requestImpl) Type(typeStr string) HTTPRequest {
	r.superAgent.Type(typeStr)
	return r
}

// Query sets the http get query params
func (r *requestImpl) Query(content interface{}) HTTPRequest {
	r.superAgent.Query(content)
	return r
}

// Send adds content to the body in http request
func (r *requestImpl) Send(content interface{}) HTTPRequest {
	r.superAgent.Send(content)
	return r
}

// End is the function to end the chain and fires the actual http request.
func (r *requestImpl) End() (resp *http.Response, body string, errors []error) {
	r.retryEnd(func() bool {
		var goResp Response
		goResp, body, errors = r.superAgent.End()
		resp = convert(goResp)
		return r.shouldRetry(resp, []byte(body), errors)
	})
	return
}

// EndBytes is the end function returns []byte as body.
func (r *requestImpl) EndBytes() (resp *http.Response, body []byte, errors []error) {
	r.retryEnd(func() bool {
		var goResp Response
		goResp, body, errors = r.superAgent.EndBytes()
		resp = convert(goResp)
		return r.shouldRetry(resp, body, errors)
	})
	return
}

// EndStruct is the end function unmarshal json body to data struct
func (r *requestImpl) EndStruct(content interface{}) (resp *http.Response, body []byte, errors []error) {
	r.retryEnd(func() bool {
		var goResp Response
		goResp, body, errors = r.superAgent.EndBytes()
		resp = convert(goResp)
		return r.shouldRetry(resp, body, errors)
	})
	err := json.Unmarshal(body, content)
	if err != nil {
		errors = append(errors, err)
	}
	return
}

// SetDebug to set the debug flag of SuperAgent
func (r *requestImpl) SetDebug(enabled bool) HTTPRequest {
	r.superAgent.SetDebug(enabled)
	return r
}

func (r *requestImpl) retryEnd(do func() bool) {
	retry := &retryActor{
		do:               do,
		backoff:          r.backoff,
		maxRetries:       r.maxRetries,
		minRetryPeriod:   r.minRetryPeriod,
		retryPeriodRange: r.retryPeriodRange,
	}
	retry.act()
}

func newHTTPClient() *http.Client {
	cookiejarOptions := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, _ := cookiejar.New(&cookiejarOptions)
	return &http.Client{Jar: jar}
}

func newSuperAgent(
	client *http.Client,
	transport *http.Transport,
) *SuperAgent {
	superAgent := &SuperAgent{
		TargetType: "json",
		Data:       make(map[string]interface{}),
		Header:     make(map[string]string),
		FormData:   url.Values{},
		QueryData:  url.Values{},
		Client:     client,
		Transport:  transport,
		Cookies:    make([]*http.Cookie, 0),
		Errors:     nil,
		BasicAuth:  struct{ Username, Password string }{},
		Debug:      false,
	}
	superAgent.SetLogger(golog.New(os.Stdout, "[gohttp]", golog.LstdFlags))
	return superAgent
}

func defaultShouldRetry(resp *http.Response, body []byte, errors []error) bool {
	if resp != nil {
		return resp.StatusCode > 499
	}
	return false
}

func convert(goResp Response) *http.Response {
	if goResp == nil {
		return nil
	}
	httpResp := http.Response(*goResp)
	return &httpResp
}

type retryActor struct {
	maxRetries       int
	backoff          float64
	minRetryPeriod   time.Duration
	retryPeriodRange time.Duration
	retries          int
	do               func() bool
	retryPeriod      time.Duration
}

func (r *retryActor) act() {
	for r.retries = 0; r.retries <= r.maxRetries; r.retries++ {
		if shouldRetry := r.do(); shouldRetry == false {
			return
		}
		time.Sleep(r.interval())
	}
}

func (r *retryActor) interval() time.Duration {
	if r.retries == 0 {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		r.retryPeriod = time.Duration(rnd.Int63n(int64(r.retryPeriodRange))) + r.minRetryPeriod
	} else {
		r.retryPeriod = time.Duration(int64(float64(r.retryPeriod) * r.backoff))
	}
	return r.retryPeriod
}
