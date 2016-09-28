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

	in := &map[string]interface{}{
		"aggs": map[string]interface{}{
			"stats_report": map[string]interface{}{
				"extended_stats": map[string]interface{}{
					"field": "value",
				},
			},
			"percs_report": map[string]interface{}{
				"percentiles": map[string]interface{}{
					"field": "value",
				},
			},
			"hist_report": map[string]interface{}{
				"date_histogram": map[string]interface{}{
					"field":         "timestamp",
					"interval":      req.Interval,
					"format":        "strict_date_time",
					"min_doc_count": 0,
				},
				"aggs": map[string]interface{}{
					"typelog": map[string]interface{}{
						"stats": map[string]interface{}{
							"field": "value",
						},
					},
				},
			},
		},
	}

	out := &map[string]interface{}{}
	err := db.Esi.DirectAccess("GET", endpoint, in, out)
	if err != nil {
		return nil, err
	}

	var stats, percs, histo interface{}
	{
		aggsx := (*out)["aggregations"]
		if aggsx == nil {
			return nil, fmt.Errorf("Failed to get aggsx")
		}
		aggs := aggsx.(map[string]interface{})
		if aggs == nil {
			return nil, fmt.Errorf("Failed to get aggs")
		}

		stats = aggs["stats_report"]
		if stats == nil {
			return nil, fmt.Errorf("Failed to get stats")
		}

		percs = aggs["percs_report"]
		if percs == nil {
			return nil, fmt.Errorf("Failed to get percs")
		}

		histo = aggs["hist_report"]
		if histo == nil {
			return nil, fmt.Errorf("Failed to get histo")
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
