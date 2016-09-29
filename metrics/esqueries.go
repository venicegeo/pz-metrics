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

import "time"

func newExtendedStatsAggsQuery(fieldName string) map[string]interface{} {
	m := map[string]interface{}{
		"extended_stats": map[string]interface{}{
			"field": fieldName,
		},
	}
	return m
}

func newRangeQuery(fieldName string, start time.Time, stop time.Time) map[string]interface{} {
	m := map[string]interface{}{
		fieldName: map[string]interface{}{
			"gte": start.Format(time.RFC3339),
			"lt":  stop.Format(time.RFC3339),
		},
	}
	return m
}

func newTermQuery(fieldName string, value interface{}) map[string]interface{} {
	m := map[string]interface{}{
		fieldName: value,
	}
	return m
}

func newPercentilesAggsQuery(field string, value string) map[string]interface{} {
	m := map[string]interface{}{
		"percentiles": map[string]interface{}{
			field: value,
		},
	}
	return m
}

func newDateHistogramAggsQuery(dateFieldName string, interval string, valueFieldName string) map[string]interface{} {
	m := map[string]interface{}{
		"date_histogram": map[string]interface{}{
			"field":         dateFieldName,
			"interval":      interval,
			"format":        "strict_date_time",
			"min_doc_count": 0,
		},
		"aggs": map[string]interface{}{
			"bucket_stats": map[string]interface{}{
				"stats": map[string]interface{}{
					"field": valueFieldName,
				},
			},
		},
	}
	return m
}

// {k1:v1, k2:v2, ...} ==> [{k1:v1},{k2:v2}]
func newAndFilter(items map[string]interface{}) []interface{} {
	m := []interface{}{}
	for k, v := range items {
		mm := map[string]interface{}{k: v}
		m = append(m, mm)
	}
	return m
}
