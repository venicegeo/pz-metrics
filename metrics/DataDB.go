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
	"encoding/json"
	"fmt"
	"time"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

type DataDB struct {
	*ResourceDB
	mapping string
}

const DataDBMapping string = "Data"

const DataIndexSettings = `
{
        "mappings": {
            "Data": {
                "properties": {
					"timestamp": {
						"type": "date",
						"store": true,
						"index": "not_analyzed"
					},
					"metricId": {
						"type": "string",
						"store": true,
						"index": "not_analyzed"
					},
					"value": {
						"type": "double",
						"store": true,
						"index": "not_analyzed"
					}
                }
            }
        }
}`

func NewDataDB(service *Service, esi elasticsearch.IIndex) (*DataDB, error) {
	rdb, err := NewResourceDB(service, esi, DataIndexSettings)
	if err != nil {
		return nil, err
	}
	ardb := DataDB{ResourceDB: rdb, mapping: DataDBMapping}
	return &ardb, nil
}

func (db *DataDB) PostData(obj interface{}, id piazza.Ident) (piazza.Ident, error) {

	indexResult, err := db.Esi.PostData(db.mapping, id.String(), obj)
	if err != nil {
		return piazza.NoIdent, LoggedError("DataDB.PostData failed: %s", err)
	}
	if !indexResult.Created {
		return piazza.NoIdent, LoggedError("DataDB.PostData failed: not created")
	}

	return id, nil
}

func (db *DataDB) GetAll(format *piazza.JsonPagination) ([]Data, int64, error) {
	datas := []Data{}

	exists, err := db.Esi.TypeExists(db.mapping)
	if err != nil {
		return datas, 0, err
	}
	if !exists {
		return datas, 0, nil
	}

	searchResult, err := db.Esi.FilterByMatchAll(db.mapping, format)
	if err != nil {
		return nil, 0, LoggedError("DataDB.GetAll failed: %s", err)
	}
	if searchResult == nil {
		return nil, 0, LoggedError("DataDB.GetAll failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		for _, hit := range *searchResult.GetHits() {
			var data Data
			err := json.Unmarshal(*hit.Source, &data)
			if err != nil {
				return nil, 0, err
			}
			datas = append(datas, data)
		}
	}

	return datas, searchResult.TotalHits(), nil
}

func (db *DataDB) GetOne(id piazza.Ident) (*Data, bool, error) {
	//log.Printf("DataDB.GetOne: %s %s", id.String(), db.mapping)
	getResult, err := db.Esi.GetByID(db.mapping, id.String())
	if err != nil {
		return nil, false, fmt.Errorf("DataDB.GetOne failed: %s", err)
	}
	if getResult == nil {
		return nil, true, fmt.Errorf("DataDB.GetOne failed: %s no getResult", id.String())
	}

	src := getResult.Source
	var data Data
	err = json.Unmarshal(*src, &data)
	if err != nil {
		return nil, getResult.Found, err
	}

	return &data, getResult.Found, nil
}

func (db *DataDB) DeleteByID(id piazza.Ident) (bool, error) {
	deleteResult, err := db.Esi.DeleteByID(db.mapping, string(id))
	if err != nil {
		return deleteResult.Found, fmt.Errorf("DataDB.DeleteById failed: %s", err)
	}
	if deleteResult == nil {
		return false, fmt.Errorf("DataDB.DeleteById failed: no deleteResult")
	}

	if !deleteResult.Found {
		return false, fmt.Errorf("DataDB.DeleteById failed: not found")
	}

	return deleteResult.Found, nil
}

func (db *DataDB) GetStats(id piazza.Ident, req *ReportRequest) (*Report, error) {
	indexName := db.Esi.IndexName()
	//log.Printf("DataDB.GetStats: %s %s", id.String(), indexName)

	command := "/_search?search_type=count"
	endpoint := fmt.Sprintf("/%s%s", indexName, command)

	newExtendedStatsAggsQuery := func(fieldName string) map[string]interface{} {
		m := map[string]interface{}{
			"extended_stats": map[string]interface{}{
				"field": fieldName,
			},
		}
		return m
	}

	newRangeQuery := func(fieldName string, start time.Time, stop time.Time) map[string]interface{} {
		m := map[string]interface{}{
			fieldName: map[string]interface{}{
				"gte": start.Format(time.RFC3339),
				"lt":  stop.Format(time.RFC3339),
			},
		}
		return m
	}

	newTermQuery := func(fieldName string, value interface{}) map[string]interface{} {
		m := map[string]interface{}{
			fieldName: value,
		}
		return m
	}

	newPercentilesAggsQuery := func(field string, value string) map[string]interface{} {
		m := map[string]interface{}{
			"percentiles": map[string]interface{}{
				field: value,
			},
		}
		return m
	}

	newDateHistogramAggsQuery := func(dateFieldName string, interval string, valueFieldName string) map[string]interface{} {
		m := map[string]interface{}{
			"date_histogram": map[string]interface{}{
				"field":         dateFieldName,
				"interval":      interval,
				"format":        "strict_date_time",
				"min_doc_count": 0,
			},
			"aggs": map[string]interface{}{
				"typelog": map[string]interface{}{
					"stats": map[string]interface{}{
						"field": valueFieldName,
					},
				},
			},
		}
		return m
	}

	in := &map[string]interface{}{
		"aggs": map[string]interface{}{
			"foo": map[string]interface{}{
				"filter": map[string]interface{}{
					"and": []interface{}{
						map[string]interface{}{
							"term": newTermQuery("metricId", id.String()),
						},
						map[string]interface{}{
							"range": newRangeQuery("timestamp", req.Start, req.End),
						},
					},
				},
				"aggs": map[string]interface{}{
					"stats_report": newExtendedStatsAggsQuery("value"),
					"percs_report": newPercentilesAggsQuery("field", "value"),
					"hist_report":  newDateHistogramAggsQuery("timestamp", req.Interval, "value"),
				},
			},
		},
	}

	out := &map[string]interface{}{}
	err := db.Esi.DirectAccess("GET", endpoint, in, out)
	if err != nil {
		return nil, err
	}

	// get the error, if there is one
	f_error := func(i interface{}) (interface{}, error) {
		ii := i.(*map[string]interface{})
		if ii == nil {
			return nil, nil
		}
		s := (*ii)["error"]
		return s, nil
	}

	// given an interface, convert to a map, find the given field, and return the value as a map
	f_map := func(i interface{}, field string) (map[string]interface{}, error) {
		ii := i.(map[string]interface{})
		if ii == nil {
			return nil, fmt.Errorf("Failed to extract field %s, because source is not map", field)
		}
		iii := ii[field]
		m := iii.(map[string]interface{})
		if m == nil {
			return nil, fmt.Errorf("Failed to extract field %s, because value is not map", field)
		}
		return m, nil
	}

	m, err := f_error(out)
	if err != nil {
		return nil, err
	}
	if m != nil {
		return nil, fmt.Errorf("ERROR ERROR %#v", m)
	}

	var stats, percs, histo interface{}
	{
		outerAggs, err := f_map(*out, "aggregations")
		if err != nil {
			return nil, err
		}

		foo, err := f_map(outerAggs, "foo")
		if err != nil {
			return nil, err
		}

		stats, err = f_map(foo, "stats_report")
		if err != nil {
			return nil, err
		}

		percs, err = f_map(foo, "percs_report")
		if err != nil {
			return nil, err
		}

		histo, err = f_map(foo, "hist_report")
		if err != nil {
			return nil, err
		}
	}

	report := &Report{
		MetricID:    id,
		Statistics:  StatisticsReport{Data: stats},
		Percentiles: PercentilesReport{Data: percs},
		Histogram:   HistogramReport{Data: histo},
	}

	return report, nil
}
