// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
)

type testK8sEventRequest struct {
	EventType       string
	EventUID        string
	Namespace       string
	ObjectName      string
	ResourceVersion string
	APIVersion      string
	Kind            string
	StartTime       int64
}

type testK8sEventResponse struct {
	ProcessingTime int64
}

type testK8sEventsRequest struct {
	IsInitialList bool
	EventsCount   int
}

type testK8sEventResponseEmpty struct{}

type mockK8sEventAttrsGetter struct{}

func (m mockK8sEventAttrsGetter) GetK8sNamespace(request testK8sEventRequest) string {
	return request.Namespace
}

func (m mockK8sEventAttrsGetter) GetK8sObjectName(request testK8sEventRequest) string {
	return request.ObjectName
}

func (m mockK8sEventAttrsGetter) GetK8sObjectResourceVersion(request testK8sEventRequest) string {
	return request.ResourceVersion
}

func (m mockK8sEventAttrsGetter) GetK8sObjectAPIVersion(request testK8sEventRequest) string {
	return request.APIVersion
}

func (m mockK8sEventAttrsGetter) GetK8sObjectKind(request testK8sEventRequest) string {
	return request.Kind
}

func (m mockK8sEventAttrsGetter) GetK8sEventType(request testK8sEventRequest) string {
	return request.EventType
}

func (m mockK8sEventAttrsGetter) GetK8sEventUID(request testK8sEventRequest) string {
	return request.EventUID
}

func (m mockK8sEventAttrsGetter) GetK8sEventProcessingTime(response testK8sEventResponse) int64 {
	return response.ProcessingTime
}

func (m mockK8sEventAttrsGetter) GetK8sEventStartTime(request testK8sEventRequest) int64 {
	return request.StartTime
}

type mockK8sEventsAttrsGetter struct{}

func (m mockK8sEventsAttrsGetter) GetK8sEventsIsInInitialList(request testK8sEventsRequest) bool {
	return request.IsInitialList
}

func (m mockK8sEventsAttrsGetter) GetK8sEventsCount(request testK8sEventsRequest) int {
	return request.EventsCount
}

func TestK8sEventAttrsExtractorOnStart(t *testing.T) {
	extractor := K8sEventAttrsExtractor[testK8sEventRequest, testK8sEventResponse, mockK8sEventAttrsGetter]{
		Getter: mockK8sEventAttrsGetter{},
	}

	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()

	request := testK8sEventRequest{
		EventType:       "Normal",
		EventUID:        "test-uid-123",
		Namespace:       "default",
		ObjectName:      "test-pod",
		ResourceVersion: "123456",
		APIVersion:      "v1",
		Kind:            "Pod",
		StartTime:       1234567890,
	}

	attrs, _ = extractor.OnStart(attrs, parentContext, request)

	assert.Equal(t, 8, len(attrs))

	expectedAttrs := []attribute.KeyValue{
		attribute.String("k8s.event.type", "Normal"),
		attribute.String("k8s.event.uid", "test-uid-123"),
		attribute.String("k8s.namespace.name", "default"),
		attribute.String("k8s.object.name", "test-pod"),
		attribute.String("k8s.object.resource_version", "123456"),
		attribute.String("k8s.object.api_version", "v1"),
		attribute.String("k8s.object.kind", "Pod"),
		attribute.Int64("k8s.event.start_time", 1234567890),
	}

	for i, expected := range expectedAttrs {
		assert.Equal(t, expected.Key, attrs[i].Key)
		assert.Equal(t, expected.Value, attrs[i].Value)
	}
}

func TestK8sEventAttrsExtractorOnStartEmptyValues(t *testing.T) {
	extractor := K8sEventAttrsExtractor[testK8sEventRequest, testK8sEventResponse, mockK8sEventAttrsGetter]{
		Getter: mockK8sEventAttrsGetter{},
	}

	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()

	request := testK8sEventRequest{
		EventType:       "",
		EventUID:        "",
		Namespace:       "",
		ObjectName:      "",
		ResourceVersion: "",
		APIVersion:      "",
		Kind:            "",
		StartTime:       0,
	}

	attrs, _ = extractor.OnStart(attrs, parentContext, request)

	assert.Equal(t, 8, len(attrs))

	expectedAttrs := []attribute.KeyValue{
		attribute.String("k8s.event.type", ""),
		attribute.String("k8s.event.uid", ""),
		attribute.String("k8s.namespace.name", ""),
		attribute.String("k8s.object.name", ""),
		attribute.String("k8s.object.resource_version", ""),
		attribute.String("k8s.object.api_version", ""),
		attribute.String("k8s.object.kind", ""),
		attribute.Int64("k8s.event.start_time", 0),
	}

	for i, expected := range expectedAttrs {
		assert.Equal(t, expected.Key, attrs[i].Key)
		assert.Equal(t, expected.Value, attrs[i].Value)
	}
}

func TestK8sEventAttrsExtractorOnEnd(t *testing.T) {
	extractor := K8sEventAttrsExtractor[testK8sEventRequest, testK8sEventResponse, mockK8sEventAttrsGetter]{
		Getter: mockK8sEventAttrsGetter{},
	}

	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()

	request := testK8sEventRequest{}
	response := testK8sEventResponse{ProcessingTime: 150}

	attrs, _ = extractor.OnEnd(attrs, parentContext, request, response, nil)

	assert.Equal(t, 1, len(attrs))
	assert.Equal(t, attribute.Key("k8s.event.process_duration_ms"), attrs[0].Key)
	assert.Equal(t, int64(150), attrs[0].Value.AsInt64())
}

func TestK8sEventAttrsExtractorOnEndZeroProcessingTime(t *testing.T) {
	extractor := K8sEventAttrsExtractor[testK8sEventRequest, testK8sEventResponse, mockK8sEventAttrsGetter]{
		Getter: mockK8sEventAttrsGetter{},
	}

	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()

	request := testK8sEventRequest{}
	response := testK8sEventResponse{ProcessingTime: 0}

	attrs, _ = extractor.OnEnd(attrs, parentContext, request, response, nil)

	assert.Equal(t, 1, len(attrs))
	assert.Equal(t, attribute.Key("k8s.event.process_duration_ms"), attrs[0].Key)
	assert.Equal(t, int64(0), attrs[0].Value.AsInt64())
}

func TestK8sEventsAttrsExtractorOnStart(t *testing.T) {
	extractor := K8sEventsAttrsExtractor[testK8sEventsRequest, testK8sEventResponseEmpty, mockK8sEventsAttrsGetter]{
		Getter: mockK8sEventsAttrsGetter{},
	}

	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()

	request := testK8sEventsRequest{
		IsInitialList: true,
		EventsCount:   5,
	}

	attrs, _ = extractor.OnStart(attrs, parentContext, request)

	assert.Equal(t, 2, len(attrs))

	expectedAttrs := []attribute.KeyValue{
		attribute.Bool("k8s.events.is_initial_list", true),
		attribute.Int("k8s.events.count", 5),
	}

	for i, expected := range expectedAttrs {
		assert.Equal(t, expected.Key, attrs[i].Key)
		assert.Equal(t, expected.Value, attrs[i].Value)
	}
}

func TestK8sEventsAttrsExtractorOnStartFalseInitialList(t *testing.T) {
	extractor := K8sEventsAttrsExtractor[testK8sEventsRequest, testK8sEventResponseEmpty, mockK8sEventsAttrsGetter]{
		Getter: mockK8sEventsAttrsGetter{},
	}

	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()

	request := testK8sEventsRequest{
		IsInitialList: false,
		EventsCount:   10,
	}

	attrs, _ = extractor.OnStart(attrs, parentContext, request)

	assert.Equal(t, 2, len(attrs))

	expectedAttrs := []attribute.KeyValue{
		attribute.Bool("k8s.events.is_initial_list", false),
		attribute.Int("k8s.events.count", 10),
	}

	for i, expected := range expectedAttrs {
		assert.Equal(t, expected.Key, attrs[i].Key)
		assert.Equal(t, expected.Value, attrs[i].Value)
	}
}

func TestK8sEventsAttrsExtractorOnStartZeroCount(t *testing.T) {
	extractor := K8sEventsAttrsExtractor[testK8sEventsRequest, testK8sEventResponseEmpty, mockK8sEventsAttrsGetter]{
		Getter: mockK8sEventsAttrsGetter{},
	}

	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()

	request := testK8sEventsRequest{
		IsInitialList: false,
		EventsCount:   0,
	}

	attrs, _ = extractor.OnStart(attrs, parentContext, request)

	assert.Equal(t, 2, len(attrs))

	expectedAttrs := []attribute.KeyValue{
		attribute.Bool("k8s.events.is_initial_list", false),
		attribute.Int("k8s.events.count", 0),
	}

	for i, expected := range expectedAttrs {
		assert.Equal(t, expected.Key, attrs[i].Key)
		assert.Equal(t, expected.Value, attrs[i].Value)
	}
}

func TestK8sEventsAttrsExtractorOnEnd(t *testing.T) {
	extractor := K8sEventsAttrsExtractor[testK8sEventsRequest, testK8sEventResponseEmpty, mockK8sEventsAttrsGetter]{
		Getter: mockK8sEventsAttrsGetter{},
	}

	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()

	request := testK8sEventsRequest{}
	response := testK8sEventResponseEmpty{}

	attrs, _ = extractor.OnEnd(attrs, parentContext, request, response, nil)

	assert.Equal(t, 0, len(attrs))
}

func TestK8sEventAttrsExtractorWithExistingAttributes(t *testing.T) {
	extractor := K8sEventAttrsExtractor[testK8sEventRequest, testK8sEventResponse, mockK8sEventAttrsGetter]{
		Getter: mockK8sEventAttrsGetter{},
	}

	attrs := []attribute.KeyValue{
		attribute.String("existing.attr", "existing-value"),
	}
	parentContext := context.Background()

	request := testK8sEventRequest{
		EventType: "Warning",
		Namespace: "test-namespace",
	}

	attrs, _ = extractor.OnStart(attrs, parentContext, request)

	assert.Equal(t, 9, len(attrs))
	assert.Equal(t, attribute.String("existing.attr", "existing-value"), attrs[0])
	assert.Equal(t, attribute.String("k8s.event.type", "Warning"), attrs[1])
	assert.Equal(t, attribute.String("k8s.namespace.name", "test-namespace"), attrs[3])
}

func TestK8sEventsAttrsExtractorWithExistingAttributes(t *testing.T) {
	extractor := K8sEventsAttrsExtractor[testK8sEventsRequest, testK8sEventResponseEmpty, mockK8sEventsAttrsGetter]{
		Getter: mockK8sEventsAttrsGetter{},
	}

	attrs := []attribute.KeyValue{
		attribute.String("existing.attr", "existing-value"),
	}
	parentContext := context.Background()

	request := testK8sEventsRequest{
		IsInitialList: true,
		EventsCount:   3,
	}

	attrs, _ = extractor.OnStart(attrs, parentContext, request)

	assert.Equal(t, 3, len(attrs))
	assert.Equal(t, attribute.String("existing.attr", "existing-value"), attrs[0])
	assert.Equal(t, attribute.Bool("k8s.events.is_initial_list", true), attrs[1])
	assert.Equal(t, attribute.Int("k8s.events.count", 3), attrs[2])
}
