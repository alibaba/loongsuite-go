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
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockK8sEventSpanNameGetter struct{}

func (m mockK8sEventSpanNameGetter) GetK8sNamespace(request testK8sEventRequest) string {
	return request.Namespace
}

func (m mockK8sEventSpanNameGetter) GetK8sObjectName(request testK8sEventRequest) string {
	return request.ObjectName
}

func (m mockK8sEventSpanNameGetter) GetK8sObjectResourceVersion(request testK8sEventRequest) string {
	return request.ResourceVersion
}

func (m mockK8sEventSpanNameGetter) GetK8sObjectAPIVersion(request testK8sEventRequest) string {
	return request.APIVersion
}

func (m mockK8sEventSpanNameGetter) GetK8sObjectKind(request testK8sEventRequest) string {
	return request.Kind
}

func (m mockK8sEventSpanNameGetter) GetK8sEventType(request testK8sEventRequest) string {
	return request.EventType
}

func (m mockK8sEventSpanNameGetter) GetK8sEventUID(request testK8sEventRequest) string {
	return request.EventUID
}

func (m mockK8sEventSpanNameGetter) GetK8sEventProcessingTime(response testK8sEventResponse) int64 {
	return response.ProcessingTime
}

func (m mockK8sEventSpanNameGetter) GetK8sEventStartTime(request testK8sEventRequest) int64 {
	return request.StartTime
}

type mockK8sEventsSpanNameGetter struct{}

func (m mockK8sEventsSpanNameGetter) GetK8sEventsIsInInitialList(request testK8sEventsRequest) bool {
	return request.IsInitialList
}

func (m mockK8sEventsSpanNameGetter) GetK8sEventsCount(request testK8sEventsRequest) int {
	return request.EventsCount
}

func TestK8sEventSpanNameExtractorWithKind(t *testing.T) {
	extractor := K8sEventSpanNameExtractor[testK8sEventRequest, testK8sEventResponse]{
		Getter: mockK8sEventSpanNameGetter{},
	}

	request := testK8sEventRequest{
		Kind: "Pod",
	}

	spanName := extractor.Extract(request)
	assert.Equal(t, "k8s.informer.Pod.process", spanName)
}

func TestK8sEventSpanNameExtractorWithDeployment(t *testing.T) {
	extractor := K8sEventSpanNameExtractor[testK8sEventRequest, testK8sEventResponse]{
		Getter: mockK8sEventSpanNameGetter{},
	}

	request := testK8sEventRequest{
		Kind: "Deployment",
	}

	spanName := extractor.Extract(request)
	assert.Equal(t, "k8s.informer.Deployment.process", spanName)
}

func TestK8sEventSpanNameExtractorWithService(t *testing.T) {
	extractor := K8sEventSpanNameExtractor[testK8sEventRequest, testK8sEventResponse]{
		Getter: mockK8sEventSpanNameGetter{},
	}

	request := testK8sEventRequest{
		Kind: "Service",
	}

	spanName := extractor.Extract(request)
	assert.Equal(t, "k8s.informer.Service.process", spanName)
}

func TestK8sEventSpanNameExtractorEmptyKind(t *testing.T) {
	extractor := K8sEventSpanNameExtractor[testK8sEventRequest, testK8sEventResponse]{
		Getter: mockK8sEventSpanNameGetter{},
	}

	request := testK8sEventRequest{
		Kind: "",
	}

	spanName := extractor.Extract(request)
	assert.Equal(t, "k8s.informer.event.process", spanName)
}

func TestK8sEventSpanNameExtractorComplexKind(t *testing.T) {
	extractor := K8sEventSpanNameExtractor[testK8sEventRequest, testK8sEventResponse]{
		Getter: mockK8sEventSpanNameGetter{},
	}

	request := testK8sEventRequest{
		Kind: "CustomResourceDefinition",
	}

	spanName := extractor.Extract(request)
	assert.Equal(t, "k8s.informer.CustomResourceDefinition.process", spanName)
}

func TestK8sEventsSpanNameExtractorInitialListTrue(t *testing.T) {
	extractor := K8sEventsSpanNameExtractor[testK8sEventsRequest, testK8sEventResponseEmpty]{
		Getter: mockK8sEventsSpanNameGetter{},
	}

	request := testK8sEventsRequest{
		IsInitialList: true,
		EventsCount:   5,
	}

	spanName := extractor.Extract(request)
	assert.Equal(t, "k8s.informer.initial_list.process", spanName)
}

func TestK8sEventsSpanNameExtractorInitialListFalse(t *testing.T) {
	extractor := K8sEventsSpanNameExtractor[testK8sEventsRequest, testK8sEventResponseEmpty]{
		Getter: mockK8sEventsSpanNameGetter{},
	}

	request := testK8sEventsRequest{
		IsInitialList: false,
		EventsCount:   10,
	}

	spanName := extractor.Extract(request)
	assert.Equal(t, "k8s.informer.events.process", spanName)
}

func TestK8sEventsSpanNameExtractorZeroCount(t *testing.T) {
	extractor := K8sEventsSpanNameExtractor[testK8sEventsRequest, testK8sEventResponseEmpty]{
		Getter: mockK8sEventsSpanNameGetter{},
	}

	request := testK8sEventsRequest{
		IsInitialList: false,
		EventsCount:   0,
	}

	spanName := extractor.Extract(request)
	assert.Equal(t, "k8s.informer.events.process", spanName)
}

func TestK8sEventsSpanNameExtractorEmptyRequest(t *testing.T) {
	extractor := K8sEventsSpanNameExtractor[testK8sEventsRequest, testK8sEventResponseEmpty]{
		Getter: mockK8sEventsSpanNameGetter{},
	}

	request := testK8sEventsRequest{}

	spanName := extractor.Extract(request)
	assert.Equal(t, "k8s.informer.events.process", spanName)
}

func TestK8sEventSpanNameExtractorCaseSensitivity(t *testing.T) {
	extractor := K8sEventSpanNameExtractor[testK8sEventRequest, testK8sEventResponse]{
		Getter: mockK8sEventSpanNameGetter{},
	}

	request := testK8sEventRequest{
		Kind: "pod", // lowercase
	}

	spanName := extractor.Extract(request)
	assert.Equal(t, "k8s.informer.pod.process", spanName)
}

func TestK8sEventSpanNameExtractorWithNumbers(t *testing.T) {
	extractor := K8sEventSpanNameExtractor[testK8sEventRequest, testK8sEventResponse]{
		Getter: mockK8sEventSpanNameGetter{},
	}

	request := testK8sEventRequest{
		Kind: "Pod123",
	}

	spanName := extractor.Extract(request)
	assert.Equal(t, "k8s.informer.Pod123.process", spanName)
}

func TestK8sEventsSpanNameExtractorInitialListTrueZeroCount(t *testing.T) {
	extractor := K8sEventsSpanNameExtractor[testK8sEventsRequest, testK8sEventResponseEmpty]{
		Getter: mockK8sEventsSpanNameGetter{},
	}

	request := testK8sEventsRequest{
		IsInitialList: true,
		EventsCount:   0,
	}

	spanName := extractor.Extract(request)
	assert.Equal(t, "k8s.informer.initial_list.process", spanName)
}
