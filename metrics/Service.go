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
	"log"
	"net/http"

	"github.com/pborman/uuid"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

const metricSchema = "MetricIndex"
const dataSchema = "DataIndex"

type Service struct {
	origin      string
	metricIndex elasticsearch.IIndex
	dataIndex   elasticsearch.IIndex
	metricDB    *MetricDB
	dataDB      *DataDB
}

func (service *Service) Init(
	sys *piazza.SystemConfig,
	metricIndex elasticsearch.IIndex,
	dataIndex elasticsearch.IIndex) error {

	var err error

	/***
	err = esIndex.Delete()
	if err != nil {
		log.Fatal(err)
	}
	if esIndex.IndexExists() {
		log.Fatal("index still exists")
	}
	err = esIndex.Create()
	if err != nil {
		log.Fatal(err)
	}
	***/

	err = service.makeMetricIndex(metricIndex)
	if err != nil {
		return err
	}

	err = service.makeDataIndex(dataIndex)
	if err != nil {
		return err
	}

	service.metricDB, err = NewMetricDB(service, metricIndex)
	if err != nil {
		return err
	}

	service.dataDB, err = NewDataDB(service, dataIndex)
	if err != nil {
		return err
	}

	service.origin = string(sys.Name)

	return nil
}

func (service *Service) makeMetricIndex(metricIndex elasticsearch.IIndex) error {
	ok, err := metricIndex.IndexExists()
	if err != nil {
		return err
	}
	if !ok {
		log.Printf("Creating index: %s", metricIndex.IndexName())
		err = metricIndex.Create("")
		if err != nil {
			log.Fatal(err)
		}
	}

	ok, err = metricIndex.TypeExists(metricSchema)
	if err != nil {
		return err
	}
	if !ok {
		//log.Printf("Creating type: %s", metricSchema)

		metricMapping :=
			`{
			"MetricIndex":{
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
		}`

		err = metricIndex.SetMapping(metricSchema, piazza.JsonString(metricMapping))
		if err != nil {
			log.Printf("Init: %s", err.Error())
			return err
		}
	}

	service.metricIndex = metricIndex

	return nil
}

func (service *Service) makeDataIndex(dataIndex elasticsearch.IIndex) error {
	ok, err := dataIndex.IndexExists()
	if err != nil {
		return err
	}
	if !ok {
		log.Printf("Creating index: %s", dataIndex.IndexName())
		err = dataIndex.Create("")
		if err != nil {
			log.Fatal(err)
		}
	}

	ok, err = dataIndex.TypeExists(dataSchema)
	if err != nil {
		return err
	}
	if !ok {
		//log.Printf("Creating type: %s", dataSchema)

		dataMapping :=
			`{
			"DataIndex":{
				"properties": {
					"timeStamp": {
						"type": "long",
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
		}`

		err = dataIndex.SetMapping(dataSchema, piazza.JsonString(dataMapping))
		if err != nil {
			log.Printf("LoggerService.Init: %s", err.Error())
			return err
		}
	}

	service.dataIndex = dataIndex

	return nil
}

func (service *Service) newOKResponse(obj interface{}) *piazza.JsonResponse {
	resp := &piazza.JsonResponse{StatusCode: http.StatusOK, Data: obj}
	err := resp.SetType()
	if err != nil {
		return service.newInternalErrorResponse(err)
	}
	return resp
}

func (service *Service) newStatusCreatedResponse(obj interface{}) *piazza.JsonResponse {
	resp := &piazza.JsonResponse{StatusCode: http.StatusCreated, Data: obj}
	err := resp.SetType()
	if err != nil {
		return service.newInternalErrorResponse(err)
	}
	return resp
}

func (service *Service) newInternalErrorResponse(err error) *piazza.JsonResponse {
	return &piazza.JsonResponse{
		StatusCode: http.StatusInternalServerError,
		Message:    err.Error(),
		Origin:     service.origin,
	}
}

func (service *Service) newBadRequestResponse(err error) *piazza.JsonResponse {
	return &piazza.JsonResponse{
		StatusCode: http.StatusBadRequest,
		Message:    err.Error(),
		Origin:     service.origin,
	}
}

func (service *Service) newNotFoundResponse(err error) *piazza.JsonResponse {
	return &piazza.JsonResponse{
		StatusCode: http.StatusNotFound,
		Message:    err.Error(),
		Origin:     service.origin,
	}
}

func (service *Service) newIdent() (piazza.Ident, error) {
	s := uuid.New() // TODO
	//log.Printf("allocated new metric/data id: %s", s)
	return piazza.Ident(s), nil
}

//---------------------------------------------------------------------

func (service *Service) GetRoot() *piazza.JsonResponse {
	resp := &piazza.JsonResponse{
		StatusCode: 200,
		Data:       "Hi. I'm pz-metrics.",
	}

	err := resp.SetType()
	if err != nil {
		return service.newInternalErrorResponse(err)
	}

	return resp
}

//---------------------------------------------------------------------

func (service *Service) PostMetric(metric *Metric) *piazza.JsonResponse {

	id, err := service.newIdent()
	if err != nil {
		return service.newInternalErrorResponse(err)
	}
	id2, err := service.metricDB.PostData(metric, id)
	if err != nil || id != id2 {
		return service.newInternalErrorResponse(err)
	}

	metric.ID = id

	resp := &piazza.JsonResponse{
		StatusCode: http.StatusOK,
		Data:       metric,
	}

	err = resp.SetType()
	if err != nil {
		return service.newInternalErrorResponse(err)
	}

	return resp
}

func (service *Service) GetMetrics(params *piazza.HttpQueryParams) *piazza.JsonResponse {
	format, err := piazza.NewJsonPagination(params)
	if err != nil {
		return service.newBadRequestResponse(err)
	}

	var totalHits int64
	var metrics []Metric

	metrics, totalHits, err = service.metricDB.GetAll(format)
	if err != nil {
		return service.newBadRequestResponse(err)
	}

	if metrics == nil {
		return service.newInternalErrorResponse(errors.New("getallmetrics returned nil"))
	}
	resp := service.newOKResponse(metrics)

	if totalHits > 0 {
		format.Count = int(totalHits)
		resp.Pagination = format
	}

	return resp
}

func (service *Service) GetMetric(id piazza.Ident) *piazza.JsonResponse {
	metric, found, err := service.metricDB.GetOne(id)
	if !found {
		return service.newNotFoundResponse(err)
	}
	if err != nil {
		return service.newBadRequestResponse(err)
	}
	return service.newOKResponse(metric)
}

func (service *Service) DeleteMetric(id piazza.Ident) *piazza.JsonResponse {
	ok, err := service.metricDB.DeleteByID(id)
	if !ok {
		return service.newNotFoundResponse(err)
	}
	if err != nil {
		return service.newBadRequestResponse(err)
	}

	return service.newOKResponse(nil)
}

//---------------------------------------------------------------------

func (service *Service) PostData(data *Data) *piazza.JsonResponse {

	id, err := service.newIdent()
	if err != nil {
		return service.newInternalErrorResponse(err)
	}

	_, err = service.dataDB.PostData(data, id)
	if err != nil {
		return service.newInternalErrorResponse(err)
	}

	data.ID = id

	resp := &piazza.JsonResponse{
		StatusCode: http.StatusOK,
		Data:       data,
	}

	err = resp.SetType()
	if err != nil {
		return service.newInternalErrorResponse(err)
	}

	return resp
}

func (service *Service) GetData(id piazza.Ident) *piazza.JsonResponse {
	//log.Printf("Service.GetData: %s", id.String())

	metric, found, err := service.dataDB.GetOne(id)
	if !found {
		return service.newNotFoundResponse(err)
	}
	if err != nil {
		return service.newBadRequestResponse(err)
	}
	return service.newOKResponse(metric)
}

func (service *Service) DeleteData(id piazza.Ident) *piazza.JsonResponse {
	ok, err := service.dataDB.DeleteByID(id)
	if !ok {
		return service.newNotFoundResponse(err)
	}
	if err != nil {
		return service.newBadRequestResponse(err)
	}

	return service.newOKResponse(nil)
}

//---------------------------------------------------------------------

func (service *Service) GetReport(id piazza.Ident, req *ReportRequest) *piazza.JsonResponse {
	//log.Printf("Service.GetReport(%s, %#v)", id, req)

	stats, err := service.dataDB.GetStats(id, req)
	if err != nil {
		return service.newInternalErrorResponse(err)
	}

	return service.newOKResponse(stats)
}
