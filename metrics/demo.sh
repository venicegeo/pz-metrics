#!/bin/sh

curl="curl -S -s"
url="http://localhost:50107"

metricJson='{
  "name": "MyCounter",
  "description": "this is my metric",
  "units": "Count"
}'

metricId=`$curl -XPOST -d "$metricJson" $url/metric | jq -r .data.id`
echo MetridId: $metricId

dataJson='{
  "metricId": "'"$metricId"'",
  "value": -1,
  "timestamp": "2016-09-28T23:49:54.000Z"
}'

dataId=`$curl -XPOST -d "$dataJson" $url/data | jq -r .data.id`
echo DataId: $dataId
