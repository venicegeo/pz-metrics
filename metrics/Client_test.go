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
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	piazza "github.com/venicegeo/pz-gocommon/gocommon"
)

func sleep() {
	time.Sleep(1 * time.Second)
}

type LoggerTester struct {
	suite.Suite

	sys    *piazza.SystemConfig
	client *Client

	genericServer *piazza.GenericServer

	metricIndex *elasticsearch.Index
	dataIndex   *elasticsearch.Index
}

func (suite *LoggerTester) SetupSuite() {}

func (suite *LoggerTester) TearDownSuite() {}

func (suite *LoggerTester) setupFixture() {
	t := suite.T()
	assert := assert.New(t)

	var required []piazza.ServiceName
	required = []piazza.ServiceName{piazza.PzElasticSearch}
	sys, err := piazza.NewSystemConfig(piazza.PzMetrics, required)
	assert.NoError(err)
	suite.sys = sys

	metricIndex, err := elasticsearch.NewIndex(sys, "metricstest$", MetricIndexSettings)
	assert.NoError(err)
	suite.metricIndex = metricIndex
	//log.Printf("New index: %s", metricIndex.IndexName())

	dataIndex, err := elasticsearch.NewIndex(sys, "datastest$", DataIndexSettings)
	assert.NoError(err)
	suite.dataIndex = dataIndex
	//log.Printf("New index: %s", dataIndex.IndexName())

	service := &Service{}
	err = service.Init(sys, metricIndex, dataIndex)
	assert.NoError(err)

	server := &Server{}
	server.Init(service)

	suite.genericServer = &piazza.GenericServer{Sys: sys}

	err = suite.genericServer.Configure(server.Routes)
	if err != nil {
		log.Fatal(err)
	}

	_, err = suite.genericServer.Start()
	assert.NoError(err)

	//	x := sys.Address
	//	assert.NoError(err)
	//	sys.AddService(piazza.PzMetrics, x)
	client, err := NewClient(sys)
	assert.NoError(err)
	suite.client = client
}

func (suite *LoggerTester) teardownFixture() {
	err := suite.genericServer.Stop()
	if err != nil {
		panic(err)
	}

	err = suite.metricIndex.Close()
	if err != nil {
		panic(err)
	}

	err = suite.metricIndex.Delete()
	if err != nil {
		panic(err)
	}

	err = suite.dataIndex.Close()
	if err != nil {
		panic(err)
	}

	err = suite.dataIndex.Delete()
	if err != nil {
		panic(err)
	}
}

func now() string {
	return time.Now().Format(time.RFC3339)
}

func TestRunSuite(t *testing.T) {
	s := &LoggerTester{}
	suite.Run(t, s)
}

func (suite *LoggerTester) Test00DirectAccess() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	out := &map[string]interface{}{}
	err := suite.dataIndex.DirectAccess("GET", "", nil, out)
	assert.NoError(err)
	assert.NotNil(out)
	assert.Equal("You Know, for Search", (*out)["tagline"])
	//log.Printf("** %#v", out)
}

func (suite *LoggerTester) Test01Metric() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	metric := &Metric{
		Name:        "MyCounter1",
		Description: "my first metric",
		Units:       UnitCount,
	}
	resp, err := client.PostMetric(metric)
	assert.NoError(err)

	_, err = client.GetMetric(resp.ID)
	assert.NoError(err)

	_, err = client.GetMetric("badid")
	assert.Error(err)

	err = client.DeleteMetric(resp.ID)
	assert.NoError(err)

	_, err = client.GetMetric(resp.ID)
	assert.Error(err)
}

func (suite *LoggerTester) Test02Data() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	var err error

	data := Data{
		Value:     17,
		Timestamp: now(),
	}

	_, err = client.GetData("badid")
	assert.Error(err)

	resp, err := client.PostData(&data)
	assert.NoError(err)

	id := resp.ID

	sleep()

	_, err = client.GetData(id)
	assert.NoError(err)

	err = client.DeleteData(id)
	assert.NoError(err)

	sleep()

	_, err = client.GetData(id)
	assert.Error(err)
}

func (suite *LoggerTester) Test03Report() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	//	defer suite.teardownFixture()

	metric := &Metric{
		Name:        "MyCounter2",
		Description: "my second metric",
		Units:       UnitCount,
	}

	resp, err := suite.client.PostMetric(metric)
	assert.NoError(err)

	metricId := resp.ID

	start := time.Now()

	for i := 0; i < 10; i++ {
		data := Data{
			MetricID:  metricId,
			Value:     -1,
			Timestamp: now(),
		}

		_, err := suite.client.PostData(&data)
		assert.NoError(err)
	}

	sleep()

	for i := 0; i < 10; i++ {
		data := Data{
			MetricID:  metricId,
			Value:     555,
			Timestamp: now(),
		}

		_, err := suite.client.PostData(&data)
		assert.NoError(err)
	}

	sleep()

	for i := 0; i < 100; i++ {
		data := Data{
			MetricID:  metricId,
			Value:     50,
			Timestamp: now(),
		}

		_, err := suite.client.PostData(&data)
		assert.NoError(err)
	}

	{
		metric3 := &Metric{
			Name:        "MyCounter3",
			Description: "my third metric",
			Units:       UnitCount,
		}

		resp3, err := suite.client.PostMetric(metric3)
		assert.NoError(err)

		metricId3 := resp3.ID
		_ = metricId3

		data3 := Data{
			MetricID:  metricId3,
			Value:     -9999,
			Timestamp: now(),
		}

		_, err = suite.client.PostData(&data3)

		data33 := Data{
			MetricID:  metricId3,
			Value:     8888,
			Timestamp: now(),
		}

		_, err = suite.client.PostData(&data33)
		assert.NoError(err)
	}

	stop := time.Now()
	sleep()

	req := &ReportRequest{
		Start:         start.Add(-1 * time.Second),
		End:           stop.Add(1 * time.Second),
		DateInterval:  "0.5s",
		ValueInterval: "100",
	}

	report, err := suite.client.GetReport(metricId, req)
	assert.NoError(err)
	assert.NotNil(report)
	assert.NotEmpty(report.String())

	log.Printf("%s", report)
}
