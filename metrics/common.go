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
	"errors"
	"fmt"
	"log"
	"time"

	piazza "github.com/venicegeo/pz-gocommon/gocommon"
)

type Units string

const (
	UnitSeconds      Units = "Seconds"
	UnitMilliseconds Units = "Milliseconds"
	UnitCount        Units = "Count"
	UnitBytes        Units = "Bytes"
	UnitSquareYards  Units = "SquareYards"
	UnitBooleans     Units = "Booleans"
	UnitStrings      Units = "Strings"
)

type Metric struct {
	ID          piazza.Ident `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Units       Units        `json:"units"`
}

type Data struct {
	ID        piazza.Ident `json:"id"`
	MetricID  piazza.Ident `json:"metricId"`
	Timestamp time.Time    `json:"timestamp"`
	Value     float64      `json:"value"`
}

type ReportRequest struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`

	// year, quarter, month, week, day, hour, minute, second
	// or fractiosn thereof, e.g. "0.001s"
	Interval string `json:"interval"`
}

type StatisticsReport struct {
	Data interface{} `json:"data"`
}

type PercentilesReport struct {
	Data interface{} `json:"data"`
}

type HistogramReport struct {
	Data interface{} `json:"data"`
}

type Report struct {
	MetricID    piazza.Ident      `json:"metricId"`
	Statistics  StatisticsReport  `json:"statistics"`
	Percentiles PercentilesReport `json:"percentiles"`
	Histogram   HistogramReport   `json:"histogram"`
}

//---------------------------------------------------------------------------

func LoggedError(mssg string, args ...interface{}) error {
	str := fmt.Sprintf(mssg, args...)
	log.Print(str)
	return errors.New(str)
}

//---------------------------------------------------------------------------

func init() {
	piazza.JsonResponseDataTypes["metrics.Metric"] = "metricsmetric"
	piazza.JsonResponseDataTypes["*metrics.Metric"] = "metricsmetric"
	piazza.JsonResponseDataTypes["[]metrics.Metric"] = "metricsmetric-list"
	piazza.JsonResponseDataTypes["metrics.Data"] = "metricsdata"
	piazza.JsonResponseDataTypes["*metrics.Data"] = "metricsdata"
	piazza.JsonResponseDataTypes["[]metrics.Data"] = "metricsdata-list"
	piazza.JsonResponseDataTypes["metrics.Report"] = "metricsreport"
	piazza.JsonResponseDataTypes["*metrics.Report"] = "metricsreport"
}
