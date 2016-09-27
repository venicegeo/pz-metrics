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

type MetricDB struct {
	*ResourceDB
	mapping string
}

const MetricDBMapping string = "Metric"

const MetricIndexSettings = `
{
        "mappings": {
            "Metric": {
				"properties": {
					"name": {
						"type": "string",
						"store": true,
						"index": "not_analyzed"
					},
					"description": {
						"type": "string",
						"store": true,
						"index": "not_analyzed"
					},
					"units": {
						"type": "string",
						"store": true,
						"index": "not_analyzed"
					},
					"datatype": {
						"type": "string",
						"store": true,
						"index": "not_analyzed"
					}
				}
            }
        }
}`

func NewMetricDB(service *Service, esi elasticsearch.IIndex) (*MetricDB, error) {
	rdb, err := NewResourceDB(service, esi, MetricIndexSettings)
	if err != nil {
		return nil, err
	}
	ardb := MetricDB{ResourceDB: rdb, mapping: MetricDBMapping}
	return &ardb, nil
}

func (db *MetricDB) PostData(obj interface{}, id piazza.Ident) (piazza.Ident, error) {
	indexResult, err := db.Esi.PostData(db.mapping, id.String(), obj)
	if err != nil {
		return piazza.NoIdent, LoggedError("MetricDB.PostData failed: %s", err)
	}
	if !indexResult.Created {
		return piazza.NoIdent, LoggedError("MetricDB.PostData failed: not created")
	}

	return id, nil
}

func (db *MetricDB) GetAll(format *piazza.JsonPagination) ([]Metric, int64, error) {
	metrics := []Metric{}
	exists, err := db.Esi.TypeExists(db.mapping)
	if err != nil {
		return metrics, 0, err
	}
	if !exists {
		return metrics, 0, nil
	}

	searchResult, err := db.Esi.FilterByMatchAll(db.mapping, format)
	if err != nil {
		return nil, 0, LoggedError("MetricDB.GetAll failed: %s", err)
	}
	if searchResult == nil {
		return nil, 0, LoggedError("MetricDB.GetAll failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		for _, hit := range *searchResult.GetHits() {
			var metric Metric
			err := json.Unmarshal(*hit.Source, &metric)
			if err != nil {
				return nil, 0, err
			}
			metrics = append(metrics, metric)
		}
	}

	return metrics, searchResult.TotalHits(), nil
}

func (db *MetricDB) GetOne(id piazza.Ident) (*Metric, bool, error) {
	getResult, err := db.Esi.GetByID(db.mapping, id.String())
	if err != nil {
		return nil, false, fmt.Errorf("MetricDB.GetOne failed: %s", err)
	}
	if getResult == nil {
		return nil, true, fmt.Errorf("MetricDB.GetOne failed: %s no getResult", id.String())
	}

	src := getResult.Source
	var metric Metric
	err = json.Unmarshal(*src, &metric)
	if err != nil {
		return nil, getResult.Found, err
	}

	return &metric, getResult.Found, nil
}

func (db *MetricDB) DeleteByID(id piazza.Ident) (bool, error) {
	deleteResult, err := db.Esi.DeleteByID(db.mapping, string(id))
	if err != nil {
		return deleteResult.Found, fmt.Errorf("MetricDB.DeleteById failed: %s", err)
	}
	if deleteResult == nil {
		return false, fmt.Errorf("MetricDB.DeleteById failed: no deleteResult")
	}

	if !deleteResult.Found {
		return false, fmt.Errorf("MetricDB.DeleteById failed: not found")
	}

	return deleteResult.Found, nil
}
