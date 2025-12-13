// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"testing"
)

const (
	mqttModuleName = "mqtt"
)

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("mqtt_basic-test", mqttModuleName, "2.6.4", "", "1.18", "", TestMQTTBasic),
		NewGeneralTestCase("mqtt_publisher-test", mqttModuleName, "2.6.4", "", "1.18", "", TestMQTTPublisher),
		NewGeneralTestCase("mqtt_subscriber-test", mqttModuleName, "2.6.4", "", "1.18", "", TestMQTTSubscriber),
	)
}

func TestMQTTBasic(t *testing.T, env ...string) {
	UseApp("mqtt/v2.6.4")

	RunGoBuild(t, "go", "build", "-o", "test_mqtt_basic", "test_mqtt_basic.go", "base.go")

	RunApp(t, "test_mqtt_basic", env...)
}

func TestMQTTPublisher(t *testing.T, env ...string) {
	UseApp("mqtt/v2.6.4")

	RunGoBuild(t, "go", "build", "-o", "test_mqtt_publisher", "test_mqtt_publisher.go", "base.go")

	RunApp(t, "test_mqtt_publisher", env...)
}

func TestMQTTSubscriber(t *testing.T, env ...string) {
	UseApp("mqtt/v2.6.4")

	RunGoBuild(t, "go", "build", "-o", "test_mqtt_subscriber", "test_mqtt_subscriber.go", "base.go")

	RunApp(t, "test_mqtt_subscriber", env...)
}
