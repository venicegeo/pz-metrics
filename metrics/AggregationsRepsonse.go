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

import "fmt"

type StdDeviationBounds struct {
	Lower float64 `json:"lower"`
	Upper float64 `json:"upper"`
}

type StatsReport struct {
	Count              int64              `json:"count"`
	Max                float64            `json:"max"`
	Min                float64            `json:"min"`
	Avg                float64            `json:"avg"`
	Sum                float64            `json:"sum"`
	SumOfSquares       float64            `json:"sum_of_squares"`
	Variance           float64            `json:"variance"`
	StdDeviation       float64            `json:"std_deviation"`
	StdDeviationBounds StdDeviationBounds `json:"std_deviation_bounds"`
}

func (d *StatsReport) String() string {
	s := `  Count: %d
  Min: %f
  Max: %f
  Avg: %f
  Sum: %f
  SumOfSquares: %f
  Variance: %f
  StdDeviation: %f
  StdDeviation.Lower: %f
  StdDeviation.Upper: %f
`
	return fmt.Sprintf(s, d.Count, d.Min, d.Max, d.Avg, d.Sum,
		d.SumOfSquares, d.Variance, d.StdDeviation,
		d.StdDeviationBounds.Lower, d.StdDeviationBounds.Upper)
}

type PercsReport struct {
	Values map[string]float64 `json:"values"`
}

func (d *PercsReport) String() string {
	s := `  1%%: %f
  5%%: %f
  25%%: %f
  50%%: %f
  75%%: %f
  95%%: %f
  99%%: %f
`
	return fmt.Sprintf(s, d.Values["1.0"], d.Values["5.0"],
		d.Values["25.0"], d.Values["50.0"], d.Values["75.0"],
		d.Values["95.0"], d.Values["99.0"])
}

type BucketStats struct {
	Count int64   `json:"count"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Avg   float64 `json:"avg"`
	Sum   float64 `json:"sum"`
}

func (b *BucketStats) String() string {
	s := `count: %d, min: %f, max: %f, avg: %f`
	return fmt.Sprintf(s, b.Count, b.Min, b.Max, b.Avg)
}

type DateBucket struct {
	Key         float64     `json:"key"`
	KeyAsString string      `json:"key_as_string"`
	BucketStats BucketStats `json:"bucket_stats"`
	DocCount    int         `json:"doc_count"`
}

type ValueBucket struct {
	Key         float64     `json:"key"`
	BucketStats BucketStats `json:"bucket_stats"`
	DocCount    int         `json:"doc_count"`
}

type ByDateBucket []DateBucket

func (a ByDateBucket) Len() int           { return len(a) }
func (a ByDateBucket) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDateBucket) Less(i, j int) bool { return a[i].KeyAsString < a[j].KeyAsString }

func (b *DateBucket) String() string {
	s := `      Key: %s
      Count: %d
      Stats: %s`
	return fmt.Sprintf(s, b.KeyAsString, b.DocCount, b.BucketStats.String())
}

type ByValueBucket []ValueBucket

func (a ByValueBucket) Len() int           { return len(a) }
func (a ByValueBucket) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByValueBucket) Less(i, j int) bool { return a[i].Key < a[j].Key }

func (b *ValueBucket) String() string {
	s := `      Key: %f
      Count: %d
      Stats: %s`
	return fmt.Sprintf(s, b.Key, b.DocCount, b.BucketStats.String())
}

type DateHistReport struct {
	Buckets []DateBucket `json:"buckets"`
}

func (d *DateHistReport) String() string {
	s := fmt.Sprintf("  Buckets:\n")
	for i, b := range d.Buckets {
		t := b.String()
		s += fmt.Sprintf("    #%d:\n%s\n", i, t)
	}
	return s
}

type ValueHistReport struct {
	Buckets []ValueBucket `json:"buckets"`
}

func (d *ValueHistReport) String() string {
	s := fmt.Sprintf("  Buckets:\n")
	for i, b := range d.Buckets {
		t := b.String()
		s += fmt.Sprintf("    #%d:\n%s\n", i, t)
	}
	return s
}

type FullReport struct {
	StatsReport     StatsReport     `json:"stats_report"`
	PercsReport     PercsReport     `json:"percs_report"`
	DateHistReport  DateHistReport  `json:"date_hist_report"`
	ValueHistReport ValueHistReport `json:"value_hist_report"`
}

func (d *FullReport) String() string {
	return fmt.Sprintf("STATISTICS:\n%s\nPERCENTILES:\n%s\nDATE-HISTOGRAM:\n%s\nVALUE-HISTOGRAM:\n%s\n",
		d.StatsReport.String(), d.PercsReport.String(),
		d.DateHistReport.String(),
		d.ValueHistReport.String())
}

type Aggregations struct {
	FullReport FullReport `json:"full_report"`
}

type RootCause struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

type ErrorResponse struct {
	RootCause []*RootCause `json:"root_cause"`
}

type AggsResponse struct {
	Error        *ErrorResponse `json:"error"`
	Aggregations Aggregations   `json:"aggregations"`
}
