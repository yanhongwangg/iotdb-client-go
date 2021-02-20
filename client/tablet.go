/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package client

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"sort"
)

type MeasurementSchema struct {
	Measurement string
	DataType    TSDataType
	Encoding    TSEncoding
	Compressor  TSCompressionType
	Properties  map[string]string
}

type Tablet struct {
	deviceId           string
	measurementSchemas []*MeasurementSchema
	timestamps         []int64
	values             []interface{}
	rowCount           int
}

func (t *Tablet) SetTimestamp(timestamp int64, rowIndex int) {
	t.timestamps[rowIndex] = timestamp
}

func (t *Tablet) SetValueAt(value interface{}, columnIndex, rowIndex int) error {
	if value == nil {
		return errors.New("Illegal argument value can't be nil")
	}

	if columnIndex < 0 || columnIndex > len(t.measurementSchemas) {
		return fmt.Errorf("Illegal argument columnIndex %d", columnIndex)
	}

	if rowIndex < 0 || rowIndex > int(t.rowCount) {
		return fmt.Errorf("Illegal argument rowIndex %d", rowIndex)
	}

	switch t.measurementSchemas[columnIndex].DataType {
	case BOOLEAN:
		values := t.values[columnIndex].([]bool)
		switch value.(type) {
		case bool:
			values[rowIndex] = value.(bool)
		case *bool:
			values[rowIndex] = *value.(*bool)
		default:
			return fmt.Errorf("Illegal argument value %v %v", value, reflect.TypeOf(value))
		}
	case INT32:
		values := t.values[columnIndex].([]int32)
		switch value.(type) {
		case int32:
			values[rowIndex] = value.(int32)
		case *int32:
			values[rowIndex] = *value.(*int32)
		default:
			return fmt.Errorf("Illegal argument value %v %v", value, reflect.TypeOf(value))
		}
	case INT64:
		values := t.values[columnIndex].([]int64)
		switch value.(type) {
		case int64:
			values[rowIndex] = value.(int64)
		case *int64:
			values[rowIndex] = *value.(*int64)
		default:
			return fmt.Errorf("Illegal argument value %v %v", value, reflect.TypeOf(value))
		}
	case FLOAT:
		values := t.values[columnIndex].([]float32)
		switch value.(type) {
		case float32:
			values[rowIndex] = value.(float32)
		case *float32:
			values[rowIndex] = *value.(*float32)
		default:
			return fmt.Errorf("Illegal argument value %v %v", value, reflect.TypeOf(value))
		}
	case DOUBLE:
		values := t.values[columnIndex].([]float64)
		switch value.(type) {
		case float64:
			values[rowIndex] = value.(float64)
		case *float64:
			values[rowIndex] = *value.(*float64)
		default:
			return fmt.Errorf("Illegal argument value %v %v", value, reflect.TypeOf(value))
		}
	case TEXT:
		values := t.values[columnIndex].([]string)
		switch value.(type) {
		case string:
			values[rowIndex] = value.(string)
		case []byte:
			values[rowIndex] = string(value.([]byte))
		default:
			return fmt.Errorf("Illegal argument value %v %v", value, reflect.TypeOf(value))
		}
	}
	return nil
}

func (t *Tablet) GetRowCount() int {
	return t.rowCount
}

func (t *Tablet) GetValueAt(columnIndex, rowIndex int) (interface{}, error) {
	if columnIndex < 0 || columnIndex > len(t.measurementSchemas) {
		return nil, fmt.Errorf("Illegal argument columnIndex %d", columnIndex)
	}

	if rowIndex < 0 || rowIndex > int(t.rowCount) {
		return nil, fmt.Errorf("Illegal argument rowIndex %d", rowIndex)
	}

	schema := t.measurementSchemas[columnIndex]
	switch schema.DataType {
	case BOOLEAN:
		return t.values[columnIndex].([]bool)[rowIndex], nil
	case INT32:
		return t.values[columnIndex].([]int32)[rowIndex], nil
	case INT64:
		return t.values[columnIndex].([]int64)[rowIndex], nil
	case FLOAT:
		return t.values[columnIndex].([]float32)[rowIndex], nil
	case DOUBLE:
		return t.values[columnIndex].([]float64)[rowIndex], nil
	case TEXT:
		return t.values[columnIndex].([]string)[rowIndex], nil
	default:
		return nil, fmt.Errorf("Illegal datatype %v", schema.DataType)
	}
}

func (t *Tablet) GetTimestampBytes() []byte {
	buff := &bytes.Buffer{}
	for _, v := range t.timestamps {
		binary.Write(buff, binary.BigEndian, v)
	}
	return buff.Bytes()
}

func (t *Tablet) GetMeasurements() []string {
	measurements := make([]string, len(t.measurementSchemas))
	for i, s := range t.measurementSchemas {
		measurements[i] = s.Measurement
	}
	return measurements
}

func (t *Tablet) getDataTypes() []int32 {
	types := make([]int32, len(t.measurementSchemas))
	for i, s := range t.measurementSchemas {
		types[i] = int32(s.DataType)
	}
	return types
}

func (t *Tablet) getValuesBytes() ([]byte, error) {
	buff := &bytes.Buffer{}
	for i, schema := range t.measurementSchemas {
		switch schema.DataType {
		case BOOLEAN:
			binary.Write(buff, binary.BigEndian, t.values[i].([]bool))
		case INT32:
			binary.Write(buff, binary.BigEndian, t.values[i].([]int32))
		case INT64:
			binary.Write(buff, binary.BigEndian, t.values[i].([]int64))
		case FLOAT:
			binary.Write(buff, binary.BigEndian, t.values[i].([]float32))
		case DOUBLE:
			binary.Write(buff, binary.BigEndian, t.values[i].([]float64))
		case TEXT:
			for _, s := range t.values[i].([]string) {
				binary.Write(buff, binary.BigEndian, int32(len(s)))
				binary.Write(buff, binary.BigEndian, []byte(s))
			}
		default:
			return nil, fmt.Errorf("Illegal datatype %v", schema.DataType)
		}
	}
	return buff.Bytes(), nil
}

func (t *Tablet) Sort() error {
	var timeIndexs = make(map[int64]int, t.rowCount)
	for index := range t.timestamps {
		timeIndexs[t.timestamps[index]] = index
	}
	var keys []int64
	for timeValue := range timeIndexs {
		keys = append(keys, timeValue)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	t.timestamps = keys
	for index := range t.values {
		sortValue, _ := sortList(t.values[index], t.getDataTypes()[index], timeIndexs, t.timestamps)
		if sortValue != nil {
			t.values[index] = sortValue
		} else {
			return fmt.Errorf("Illegal datatype %v", t.getDataTypes()[index])
		}
	}
	return nil
}

func sortList(valueList interface{}, dataType int32, timeIndexs map[int64]int, timeStamps []int64) (interface{}, error){
	switch dataType {
	case 0:
		boolValues := valueList.([]bool)
		sortedValues := make([]bool, len(boolValues))
		for index := range sortedValues {
			sortedValues[index] = boolValues[timeIndexs[timeStamps[index]]]
		}
		return sortedValues, nil
	case 1:
		intValues := valueList.([]int32)
		sortedValues := make([]int32, len(intValues))
		for index := range sortedValues {
			sortedValues[index] = intValues[timeIndexs[timeStamps[index]]]
		}
		return sortedValues, nil
	case 2:
		longValues := valueList.([]int64)
		sortedValues := make([]int64, len(longValues))
		for index := range sortedValues {
			sortedValues[index] = longValues[timeIndexs[timeStamps[index]]]
		}
		return sortedValues, nil
	case 3:
		floatValues := valueList.([]float32)
		sortedValues := make([]float32, len(floatValues))
		for index := range sortedValues {
			sortedValues[index] = floatValues[timeIndexs[timeStamps[index]]]
		}
		return sortedValues, nil
	case 4:
		doubleValues := valueList.([]float64)
		sortedValues := make([]float64, len(doubleValues))
		for index := range sortedValues {
			sortedValues[index] = doubleValues[timeIndexs[timeStamps[index]]]
		}
		return sortedValues, nil
	case 5:
		stringValues := valueList.([]string)
		sortedValues := make([]string, len(stringValues))
		for index := range sortedValues {
			sortedValues[index] = stringValues[timeIndexs[timeStamps[index]]]
		}
		return sortedValues,nil
	default:
		return nil, fmt.Errorf("Illegal datatype %v", dataType)
	}
}


func NewTablet(deviceId string, measurementSchemas []*MeasurementSchema, rowCount int) (*Tablet, error) {
	tablet := &Tablet{
		deviceId:           deviceId,
		measurementSchemas: measurementSchemas,
		rowCount:           rowCount,
	}
	tablet.timestamps = make([]int64, rowCount)
	tablet.values = make([]interface{}, len(measurementSchemas))
	for i, schema := range tablet.measurementSchemas {
		switch schema.DataType {
		case BOOLEAN:
			tablet.values[i] = make([]bool, rowCount)
		case INT32:
			tablet.values[i] = make([]int32, rowCount)
		case INT64:
			tablet.values[i] = make([]int64, rowCount)
		case FLOAT:
			tablet.values[i] = make([]float32, rowCount)
		case DOUBLE:
			tablet.values[i] = make([]float64, rowCount)
		case TEXT:
			tablet.values[i] = make([]string, rowCount)
		default:
			return nil, fmt.Errorf("Illegal datatype %v", schema.DataType)
		}
	}
	return tablet, nil
}
