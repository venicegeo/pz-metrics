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
	"sort"

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

func (db *DataDB) GetStats(id piazza.Ident, req *ReportRequest) (*FullReport, error) {
	indexName := db.Esi.IndexName()
	//log.Printf("DataDB.GetStats: %s %s", id.String(), indexName)

	command := "/_search?search_type=count"
	endpoint := fmt.Sprintf("/%s%s", indexName, command)

	in := &map[string]interface{}{
		"aggs": map[string]interface{}{
			"full_report": map[string]interface{}{
				"filter": map[string]interface{}{
					"and": newAndFilter(
						map[string]interface{}{
							"term":  newTermQuery("metricId", id.String()),
							"range": newRangeQuery("timestamp", req.Start, req.End),
						},
					),
				},
				"aggs": map[string]interface{}{
					"stats_report": newExtendedStatsAggsQuery("value"),
					"percs_report": newPercentilesAggsQuery("field", "value"),
					"hist_report":  newDateHistogramAggsQuery("timestamp", req.Interval, "value"),
				},
			},
		},
	}

	out := &AggsResponse{}
	err := db.Esi.DirectAccess("GET", endpoint, in, out)
	if out.Error != nil && out.Error.RootCause != nil && (out.Error.RootCause)[0] != nil {
		return nil, fmt.Errorf("%#v", (out.Error.RootCause)[0])
	}

	if err != nil {
		return nil, err
	}

	sort.Sort(ByBucket(out.Aggregations.FullReport.HistReport.Buckets))

	return &out.Aggregations.FullReport, nil
}
