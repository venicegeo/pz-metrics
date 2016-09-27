// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"net/http"

	"github.com/venicegeo/pz-gocommon/gocommon"
)

//---------------------------------------------------------------------

type Client struct {
	url            string
	serviceName    piazza.ServiceName
	serviceAddress string
	h              piazza.Http
}

//---------------------------------------------------------------------

func NewClient(sys *piazza.SystemConfig) (*Client, error) {
	var err error

	url, err := sys.GetURL(piazza.PzMetrics)
	if err != nil {
		return nil, err
	}

	service := &Client{
		url:            url,
		serviceName:    sys.Name,
		serviceAddress: sys.Address,
		h: piazza.Http{
			BaseUrl: url,
			//Preflight:  piazza.SimplePreflight,
			//Postflight: piazza.SimplePostflight,
		},
	}

	err = sys.WaitForService(piazza.PzMetrics)
	if err != nil {
		return nil, err
	}

	return service, nil
}

func NewClient2(url string) (*Client, error) {

	service := &Client{
		url:            url,
		serviceName:    "notset",
		serviceAddress: "0.0.0.0",
		h: piazza.Http{
			BaseUrl: url,
			//Preflight:  preflight,
			//Postflight: postflight,
		},
	}

	return service, nil
}

//---------------------------------------------------------------------

// stolen from pz-workflow TODO

func (c *Client) getObject(endpoint string, out interface{}) error {

	h := piazza.Http{BaseUrl: c.url}

	resp := h.PzGet(endpoint)
	if resp.IsError() {
		return resp.ToError()
	}

	if resp.StatusCode != http.StatusOK {
		return resp.ToError()
	}

	err := resp.ExtractData(out)
	return err
}

func (c *Client) getObject2(endpoint string, input interface{}, out interface{}) error {

	h := piazza.Http{BaseUrl: c.url}

	resp := h.PzGet2(endpoint, input)
	if resp.IsError() {
		return resp.ToError()
	}

	if resp.StatusCode != http.StatusOK {
		return resp.ToError()
	}

	err := resp.ExtractData(out)
	return err
}

func (c *Client) postObject(obj interface{}, endpoint string, out interface{}) error {
	h := piazza.Http{BaseUrl: c.url}
	resp := h.PzPost(endpoint, obj)
	if resp.IsError() {
		return resp.ToError()
	}

	if resp.StatusCode != http.StatusCreated &&
		resp.StatusCode != http.StatusOK {
		return resp.ToError()
	}

	err := resp.ExtractData(out)
	return err
}

func (c *Client) putObject(obj interface{}, endpoint string, out interface{}) error {
	h := piazza.Http{BaseUrl: c.url}

	resp := h.PzPut(endpoint, obj)
	if resp.IsError() {
		return resp.ToError()
	}

	if resp.StatusCode != http.StatusOK {
		return resp.ToError()
	}

	err := resp.ExtractData(out)
	return err
}

func (c *Client) deleteObject(endpoint string) error {
	h := piazza.Http{BaseUrl: c.url}
	resp := h.PzDelete(endpoint)
	if resp.IsError() {
		return resp.ToError()
	}
	if resp.StatusCode != http.StatusOK {
		return resp.ToError()
	}

	return nil
}

//---------------------------------------------------------------------

func (c *Client) GetVersion() (*piazza.Version, error) {
	jresp := c.h.PzGet("/version")
	if jresp.IsError() {
		return nil, jresp.ToError()
	}

	var version piazza.Version
	err := jresp.ExtractData(&version)
	if err != nil {
		return nil, err
	}

	return &version, nil
}

//---------------------------------------------------------------------

func (c *Client) PostMetric(metric *Metric) (*Metric, error) {
	out := &Metric{}
	err := c.postObject(metric, "/metric", out)
	return out, err
}

func (c *Client) GetAllMetrics() (*[]Metric, error) {
	out := &[]Metric{}
	err := c.getObject("/metric", out)
	return out, err
}

func (c *Client) GetMetric(id piazza.Ident) (*Metric, error) {
	out := &Metric{}
	err := c.getObject("/metric/"+id.String(), out)
	return out, err
}

func (c *Client) DeleteMetric(id piazza.Ident) error {
	err := c.deleteObject("/metric/" + id.String())
	return err
}

//---------------------------------------------------------------------

func (c *Client) PostData(data *Data) (*Data, error) {
	out := &Data{}
	err := c.postObject(data, "/data", out)
	return out, err
}

func (c *Client) GetData(id piazza.Ident) (*Data, error) {
	out := &Data{}
	err := c.getObject("/data/"+id.String(), out)
	return out, err
}

func (c *Client) DeleteData(id piazza.Ident) error {
	err := c.deleteObject("/data/" + id.String())
	return err
}

//---------------------------------------------------------------------

func (c *Client) GetReport(id piazza.Ident, req *ReportRequest) (*Report, error) {
	out := &Report{}
	err := c.getObject2("/report/"+id.String(), req, out)
	//log.Printf("stats2: %#v", out)
	return out, err
}
