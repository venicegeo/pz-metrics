#!/bin/sh

curl="curl -S -s"
url="http://localhost:55732"

metricJson='{
  "name": "MyCounter",
  "description": "this is my metric",
  "units": "Count"
}'

metricId=`$curl -XPOST -d "$metricJson" $url/metric | jq -r .data.id`
echo MetridId: $metricId

dataJson1='{
  "metricId": "'"$metricId"'",
  "value": 10,
  "timestamp": "2016-09-28T01:10:00.000Z"
}'
dataJson2='{
  "metricId": "'"$metricId"'",
  "value": 20,
  "timestamp": "2016-09-28T02:10:00.000Z"
}'
dataJson3='{
  "metricId": "'"$metricId"'",
  "value": 100,
  "timestamp": "2016-09-28T03:10:00.000Z"
}'
dataJson4='{
  "metricId": "'"$metricId"'",
  "value": 101,
  "timestamp": "2016-09-28T04:10:00.000Z"
}'

dataId=`$curl -XPOST -d "$dataJson1" $url/data | jq -r .data.id`
echo "DataId1: $dataId"
dataId=`$curl -XPOST -d "$dataJson2" $url/data | jq -r .data.id`
echo "DataId2: $dataId"
dataId=`$curl -XPOST -d "$dataJson3" $url/data | jq -r .data.id`
echo "DataId3: $dataId"
dataId=`$curl -XPOST -d "$dataJson4" $url/data | jq -r .data.id`
echo "DataId4: $dataId"

sleep 3

reportJson='{
    "start": "2016-09-28T00:00:00.000Z",
    "end":   "2016-09-28T12:06:00.000Z",
    "valueInterval": "10",
    "dateInterval": "1h"
}'

$curl -XGET -d "$reportJson" $url/report/$metricId
