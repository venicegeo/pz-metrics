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

package main

import (
	"log"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
	pzmetrics "github.com/venicegeo/pz-metrics/metrics"
)

func assertNoError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	var required []piazza.ServiceName
	required = []piazza.ServiceName{piazza.PzElasticSearch}
	sys, err := piazza.NewSystemConfig(piazza.PzMetrics, required)
	assertNoError(err)

	metricIndex, err := elasticsearch.NewIndex(sys, "metricstest$", pzmetrics.MetricIndexSettings)
	assertNoError(err)
	//log.Printf("New index: %s", metricIndex.IndexName())

	dataIndex, err := elasticsearch.NewIndex(sys, "datastest$", pzmetrics.DataIndexSettings)
	assertNoError(err)
	//log.Printf("New index: %s", dataIndex.IndexName())

	service := &pzmetrics.Service{}
	err = service.Init(sys, metricIndex, dataIndex)
	assertNoError(err)

	server := &pzmetrics.Server{}
	server.Init(service)

	genericServer := &piazza.GenericServer{Sys: sys}

	err = genericServer.Configure(server.Routes)
	if err != nil {
		log.Fatal(err)
	}

	done, err := genericServer.Start()
	assertNoError(err)

	//client, err := pzmetrics.NewClient(sys)
	//assertNoError(err)

	err = <-done
	if err != nil {
		log.Fatal(err)
	}
}
