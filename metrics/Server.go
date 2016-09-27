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

	"github.com/gin-gonic/gin"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

type Server struct {
	service *Service
	Routes  []piazza.RouteData
}

const Version = "1.0.0"

func (server *Server) handleGetRoot(c *gin.Context) {
	resp := server.service.GetRoot()
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleGetVersion(c *gin.Context) {
	version := piazza.Version{Version: Version}
	resp := &piazza.JsonResponse{StatusCode: http.StatusOK, Data: version}
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handlePostMetric(c *gin.Context) {
	var metric Metric
	err := c.BindJSON(&metric)
	if err != nil {
		resp := &piazza.JsonResponse{StatusCode: http.StatusBadRequest, Message: err.Error()}
		piazza.GinReturnJson(c, resp)
	}
	resp := server.service.PostMetric(&metric)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleGetMetrics(c *gin.Context) {
	params := piazza.NewQueryParams(c.Request)
	resp := server.service.GetMetrics(params)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleGetMetric(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.GetMetric(id)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleDeleteMetric(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.DeleteMetric(id)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handlePostData(c *gin.Context) {
	var data Data
	err := c.BindJSON(&data)
	if err != nil {
		resp := &piazza.JsonResponse{StatusCode: http.StatusBadRequest, Message: err.Error()}
		piazza.GinReturnJson(c, resp)
	}
	resp := server.service.PostData(&data)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleGetData(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	//log.Printf("Server.handleGetData: %s", id.String())
	resp := server.service.GetData(id)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleDeleteData(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.DeleteData(id)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleGetReport(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	var req ReportRequest
	err := c.BindJSON(&req)
	if err != nil {
		resp := &piazza.JsonResponse{StatusCode: http.StatusBadRequest, Message: err.Error()}
		piazza.GinReturnJson(c, resp)
	}
	resp := server.service.GetReport(id, &req)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) Init(service *Service) {
	server.service = service

	server.Routes = []piazza.RouteData{
		{Verb: "GET", Path: "/", Handler: server.handleGetRoot},
		{Verb: "GET", Path: "/version", Handler: server.handleGetVersion},

		{Verb: "GET", Path: "/metric", Handler: server.handleGetMetrics},
		{Verb: "POST", Path: "/metric", Handler: server.handlePostMetric},

		{Verb: "GET", Path: "/metric/:id", Handler: server.handleGetMetric},
		{Verb: "DELETE", Path: "/metric/:id", Handler: server.handleDeleteMetric},

		{Verb: "POST", Path: "/data", Handler: server.handlePostData},
		{Verb: "GET", Path: "/data/:id", Handler: server.handleGetData},
		{Verb: "DELETE", Path: "/data/:id", Handler: server.handleDeleteData},

		{Verb: "GET", Path: "/report/:id", Handler: server.handleGetReport},
	}
}
