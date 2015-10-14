package gorequest

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// some test cases are from gorequest
// added retry and timeout test cases
// testing for Get method
func TestHTTPGet(t *testing.T) {
	const case1Empty = "/"
	const case2SetHeader = "/set_header"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check method is GET before going to check other features
		if r.Method != GET {
			t.Errorf("Expected method %q; got %q", GET, r.Method)
		}
		if r.Header == nil {
			t.Errorf("Expected non-nil request Header")
		}
		switch r.URL.Path {
		default:
			t.Errorf("No testing for this case yet : %q", r.URL.Path)
		case case1Empty:
			t.Logf("case %v ", case1Empty)
		case case2SetHeader:
			t.Logf("case %v ", case2SetHeader)
			if r.Header.Get("API-Key") != "fookey" {
				t.Errorf("Expected 'API-Key' == %q; got %q", "fookey", r.Header.Get("API-Key"))
			}
		}
	}))

	defer ts.Close()
	c := New()
	c.Request().Get(ts.URL + case1Empty).
		End()

	c.Request().Get(ts.URL+case2SetHeader).
		Set("API-Key", "fookey").
		End()
}

// testing for POST method
func TestHTTPPost(t *testing.T) {
	const case1Empty = "/"
	const case2SetHeader = "/set_header"
	const case3SendJSON = "/send_json"
	const case4SendString = "/send_string"
	const case5IntegSendJSONString = "/integrationSendJSON_string"
	const case6SetQuery = "/set_query"
	const case7IntegSendJSONStruct = "/integrationSendJSON_struct"
	// Check that the number conversion should be converted as string not float64
	const case8SendJSONWithLongIDNumber = "/send_jsonWithLongIDNumber"
	const case9SendJSONStringResult = "/send_jsonStringResult"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check method is PATCH before going to check other features
		if r.Method != POST {
			t.Errorf("Expected method %q; got %q", POST, r.Method)
		}
		if r.Header == nil {
			t.Errorf("Expected non-nil request Header")
		}
		switch r.URL.Path {
		default:
			t.Errorf("No testing for this case yet : %q", r.URL.Path)
		case case1Empty:
			t.Logf("case %v ", case1Empty)
		case case2SetHeader:
			t.Logf("case %v ", case2SetHeader)
			if r.Header.Get("API-Key") != "fookey" {
				t.Errorf("Expected 'API-Key' == %q; got %q", "fookey", r.Header.Get("API-Key"))
			}
		case case3SendJSON:
			t.Logf("case %v ", case3SendJSON)
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			if string(body) != `{"query1":"test","query2":"test"}` {
				t.Error(`Expected Body with {"query1":"test","query2":"test"}`, "| but got", string(body))
			}
		case case4SendString:
			t.Logf("case %v ", case4SendString)
			if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
				t.Error("Expected Header Content-Type -> application/x-www-form-urlencoded", "| but got", r.Header.Get("Content-Type"))
			}
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			if string(body) != "query1=test&query2=test" {
				t.Error("Expected Body with \"query1=test&query2=test\"", "| but got", string(body))
			}
		case case5IntegSendJSONString:
			t.Logf("case %v ", case5IntegSendJSONString)
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			if string(body) != "query1=test&query2=test" {
				t.Error("Expected Body with \"query1=test&query2=test\"", "| but got", string(body))
			}
		case case6SetQuery:
			t.Logf("case %v ", case6SetQuery)
			v := r.URL.Query()
			if v["query1"][0] != "test" {
				t.Error("Expected query1:test", "| but got", v["query1"][0])
			}
			if v["query2"][0] != "test" {
				t.Error("Expected query2:test", "| but got", v["query2"][0])
			}
		case case7IntegSendJSONStruct:
			t.Logf("case %v ", case7IntegSendJSONStruct)
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			comparedBody := []byte(`{"Lower":{"Color":"green","Size":1.7},"Upper":{"Color":"red","Size":0},"a":"a","name":"Cindy"}`)
			if !bytes.Equal(body, comparedBody) {
				t.Errorf(`Expected correct json but got ` + string(body))
			}
		case case8SendJSONWithLongIDNumber:
			t.Logf("case %v ", case8SendJSONWithLongIDNumber)
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			if string(body) != `{"id":123456789,"name":"nemo"}` {
				t.Error(`Expected Body with {"id":123456789,"name":"nemo"}`, "| but got", string(body))
			}
		case case9SendJSONStringResult:
			t.Logf("case %v ", case9SendJSONStringResult)
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			if string(body) != `id=123456789&name=nemo` {
				t.Error(`Expected Body with "id=123456789&name=nemo"`, `| but got`, string(body))
			}
		}
	}))

	defer ts.Close()

	c := New()
	c.Request().Post(ts.URL + case1Empty).
		End()

	c.Request().Post(ts.URL+case2SetHeader).
		Set("API-Key", "fookey").
		End()

	c.Request().Post(ts.URL + case3SendJSON).
		Send(`{"query1":"test"}`).
		Send(`{"query2":"test"}`).
		End()

	c.Request().Post(ts.URL + case4SendString).
		Send("query1=test").
		Send("query2=test").
		End()

	c.Request().Post(ts.URL + case5IntegSendJSONString).
		Send("query1=test").
		Send(`{"query2":"test"}`).
		End()

	/* TODO: More testing post for application/x-www-form-urlencoded
	   post.query(json), post.query(string), post.send(json), post.send(string), post.query(both).send(both)
	*/
	c.Request().Post(ts.URL + case6SetQuery).
		Query("query1=test").
		Query("query2=test").
		End()
	// TODO:
	// 1. test normal struct
	// 2. test 2nd layer nested struct
	// 3. test struct pointer
	// 4. test lowercase won't be export to json
	// 5. test field tag change to json field name
	type Upper struct {
		Color string
		Size  int
		note  string
	}
	type Lower struct {
		Color string
		Size  float64
		note  string
	}
	type Style struct {
		Upper Upper
		Lower Lower
		Name  string `json:"name"`
	}
	myStyle := Style{Upper: Upper{Color: "red"}, Name: "Cindy", Lower: Lower{Color: "green", Size: 1.7}}
	c.Request().Post(ts.URL + case7IntegSendJSONStruct).
		Send(`{"a":"a"}`).
		Send(myStyle).
		End()

	c.Request().Post(ts.URL + case8SendJSONWithLongIDNumber).
		Send(`{"id":123456789, "name":"nemo"}`).
		End()

	c.Request().Post(ts.URL + case9SendJSONStringResult).
		Type("form").
		Send(`{"id":123456789, "name":"nemo"}`).
		End()
}

func TestHTTPEndStruct(t *testing.T) {
	var jsonBlob = []byte(`{"agoraName": "Platypus", "agoraNote": "Monotremata"}`)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(jsonBlob)
	}))
	defer ts.Close()

	type Order struct {
		AgoraName string `json: "agoraName"`
		AgoraNote string `json: "agoraNote"`
	}

	c := New()
	{
		order := &Order{}
		resp, bodyBytes, errs := c.Request().Get(ts.URL).EndStruct(order)
		if len(errs) > 0 {
			t.Errorf("Unexpected errors: %s", errs)
		}
		if resp.StatusCode != 200 {
			t.Errorf("Expected StatusCode=200, actual StatusCode=%v", resp.StatusCode)
		}
		if bodyBytes == nil {
			t.Errorf("Expected bodyBytes=%s, actual bodyBytes=%s", string(bodyBytes), string(jsonBlob))
		}
		if order.AgoraName != "Platypus" {
			t.Errorf("Expected AgoraName=Platypus, actual AgoraName=%v", order.AgoraName)
		}
		if order.AgoraNote != "Monotremata" {
			t.Errorf("Expected AgoraNote=Monotremata, actual AgoraNote=%v", order.AgoraNote)
		}
	}
}

// testing for Retry method
func TestHTTPRetry(t *testing.T) {
	const case1Empty = "/"
	var jsonBlob = []byte(`
		{"name": "Platypus", "order": "Monotremata"}
	`)
	tryTimes := 0
	maxRetryTimes := 5
	backoff := float64(1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check method is GET before going to check other features
		if r.Method != GET {
			t.Errorf("Expected method %q; got %q", GET, r.Method)
		}
		switch r.URL.Path {
		default:
			t.Errorf("No testing for this case yet : %q", r.URL.Path)
		case case1Empty:
			t.Logf("case %v ", case1Empty)
			tryTimes++
			w.WriteHeader(500)
			w.Write(jsonBlob)
		}
	}))

	defer ts.Close()
	c := New().Retry(maxRetryTimes, backoff)
	c.Request().Get(ts.URL + case1Empty).
		End()

	{
		if maxRetryTimes != tryTimes-1 {
			t.Errorf("Retry times dose not match setting, actual retryTimes is %v", tryTimes)
		}
	}
}

// testing for Timeout method
func TestHTTPTimeout(t *testing.T) {
	const case1Empty = "/"
	var jsonBlob = []byte(`
		{"name": "Platypus", "order": "Monotremata"}
	`)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check method is GET before going to check other features
		if r.Method != GET {
			t.Errorf("Expected method %q; got %q", GET, r.Method)
		}
		switch r.URL.Path {
		default:
			t.Errorf("No testing for this case yet : %q", r.URL.Path)
		case case1Empty:
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(200)
			w.Write(jsonBlob)
		}
	}))

	defer ts.Close()
	c := New().Timeout(100 * time.Millisecond)
	resp, body, errors := c.Request().Get(ts.URL + case1Empty).
		End()

	{
		if resp != nil {
			t.Error("Response returned with timeout")
		}
		if len(body) != 0 {
			t.Error("Body returned with timeout")
		}
		if len(errors) == 0 {
			t.Error("No timeout error returned")
		}
	}
}

// testing for Retry method with injected shouldRetry func
func TestHTTPInjectedRetry(t *testing.T) {
	const case1Empty = "/"
	var jsonBlob = []byte(`
		{"name": "Platypus", "order": "Monotremata"}
	`)
	tryTimes := 0
	maxRetryTimes := 5
	backoff := float64(1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check method is GET before going to check other features
		if r.Method != GET {
			t.Errorf("Expected method %q; got %q", GET, r.Method)
		}
		switch r.URL.Path {
		default:
			t.Errorf("No testing for this case yet : %q", r.URL.Path)
		case case1Empty:
			t.Logf("case %v ", case1Empty)
			tryTimes++
			w.WriteHeader(500)
			w.Write(jsonBlob)
		}
	}))

	defer ts.Close()

	shouldRetry := func(resp *http.Response, body []byte, errors []error) bool {
		return false
	}

	c := New().Retry(maxRetryTimes, backoff, shouldRetry)
	c.Request().Get(ts.URL + case1Empty).
		End()

	{
		if tryTimes != 1 {
			t.Errorf("Retry times dose not match setting, actual retryTimes is %v", tryTimes)
		}
	}
}
