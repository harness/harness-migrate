// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package harness

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
)

var (
	ErrDuplicate    = errors.New("Resource already exists")
	ErrNotFound     = errors.New("Resource not found")
	ErrUnauthorized = errors.New("Unauthorized")
	ErrForbidden    = errors.New("Forbidden")
)

// helper function to make an http request
func Do(rawurl, method string, setAuth func(h *http.Header), in, out interface{}, tracing bool) error {
	body, err := Open(rawurl, method, setAuth, in, out, tracing)
	if err != nil {
		return err
	}
	defer body.Close()
	if out != nil {
		return json.NewDecoder(body).Decode(out)
	}
	return nil
}

// helper function to open an http request
func Open(rawurl, method string, setAuth func(h *http.Header), in, out interface{}, tracing bool) (io.ReadCloser, error) {
	uri, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, uri.String(), nil)
	if err != nil {
		return nil, err
	}

	setAuth(&req.Header)

	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "curl/7.79.1")
	if in != nil {
		if buf, ok := in.(*bytes.Buffer); ok {
			req.Body = ioutil.NopCloser(buf)
			req.ContentLength = int64(buf.Len())
		} else {
			decoded, derr := json.Marshal(in)
			if derr != nil {
				return nil, derr
			}
			buf := bytes.NewBuffer(decoded)
			req.Body = ioutil.NopCloser(buf)
			req.ContentLength = int64(len(decoded))
			req.Header.Set("Content-Length", strconv.Itoa(len(decoded)))
			req.Header.Set("Content-Type", "application/json")
		}
	}

	// if tracing enabled, dump the request body.
	if tracing {
		dump, _ := httputil.DumpRequest(req, true)
		os.Stdout.Write(dump)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// if tracing enabled, dump the response body.
	if tracing {
		dump, _ := httputil.DumpResponse(resp, true)
		os.Stdout.Write(dump)
	}

	if resp.StatusCode < 299 {
		return resp.Body, nil
	}

	switch resp.StatusCode {
	case 401:
		return nil, ErrUnauthorized
	case 403:
		return nil, ErrForbidden
	case 404:
		return nil, ErrNotFound
	case 409:
		return nil, ErrDuplicate
	default:
		defer resp.Body.Close()
		out, _ := ioutil.ReadAll(resp.Body)
		// attempt to unmarshal the error into the
		// custom Error structure.
		resperr := new(Error)
		if jsonerr := json.Unmarshal(out, resperr); jsonerr == nil {
			return nil, resperr
		}
		// else return the error body as a string
		return nil, fmt.Errorf("client error %d: %s", resp.StatusCode, string(out))
	}
}
