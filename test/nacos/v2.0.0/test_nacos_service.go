// Copyright (c) 2024 Alibaba Group Holding Ltd.
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

package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/util"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func main() {
	port := 8848
	if os.Getenv("NACOS_PORT") != "" {
		port, _ = strconv.Atoi(os.Getenv("NACOS_PORT"))
	}
	//create ServerConfig
	sc := []constant.ServerConfig{
		*constant.NewServerConfig("127.0.0.1", uint64(port), constant.WithContextPath("/nacos")),
	}

	//create ClientConfig
	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(""),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("debug"),
	)

	// create naming client
	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)

	if err != nil {
		panic(err)
	}

	//Register
	ExampleServiceClient_RegisterServiceInstance(client, vo.RegisterInstanceParam{
		Ip:          "10.0.0.10",
		Port:        8848,
		ServiceName: "demo.go",
		GroupName:   "group-a",
		ClusterName: "cluster-a",
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"idc": "shanghai"},
	})

	//DeRegister
	ExampleServiceClient_DeRegisterServiceInstance(client, vo.DeregisterInstanceParam{
		Ip:          "10.0.0.10",
		Port:        8848,
		ServiceName: "demo.go",
		GroupName:   "group-a",
		Cluster:     "cluster-a",
		Ephemeral:   true, //it must be true
	})

	//Register
	ExampleServiceClient_RegisterServiceInstance(client, vo.RegisterInstanceParam{
		Ip:          "10.0.0.10",
		Port:        8848,
		ServiceName: "demo.go",
		GroupName:   "group-a",
		ClusterName: "cluster-a",
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"idc": "shanghai"},
	})

	time.Sleep(1 * time.Second)

	//Get service with serviceName, groupName , clusters
	ExampleServiceClient_GetService(client, vo.GetServiceParam{
		ServiceName: "demo.go",
		GroupName:   "group-a",
		Clusters:    []string{"cluster-a"},
	})

	//SelectAllInstance
	//GroupName=DEFAULT_GROUP
	ExampleServiceClient_SelectAllInstances(client, vo.SelectAllInstancesParam{
		ServiceName: "demo.go",
		GroupName:   "group-a",
		Clusters:    []string{"cluster-a"},
	})

	//SelectInstances only return the instances of healthy=${HealthyOnly},enable=true and weight>0
	//ClusterName=DEFAULT,GroupName=DEFAULT_GROUP
	ExampleServiceClient_SelectInstances(client, vo.SelectInstancesParam{
		ServiceName: "demo.go",
		GroupName:   "group-a",
		Clusters:    []string{"cluster-a"},
	})

	//SelectOneHealthyInstance return one instance by WRR strategy for load balance
	//And the instance should be health=true,enable=true and weight>0
	//ClusterName=DEFAULT,GroupName=DEFAULT_GROUP
	ExampleServiceClient_SelectOneHealthyInstance(client, vo.SelectOneHealthInstanceParam{
		ServiceName: "demo.go",
		GroupName:   "group-a",
		Clusters:    []string{"cluster-a"},
	})

	//Subscribe key=serviceName+groupName+cluster
	//Note:We call add multiple SubscribeCallback with the same key.
	param := &vo.SubscribeParam{
		ServiceName: "demo.go",
		GroupName:   "group-a",
		SubscribeCallback: func(services []model.Instance, err error) {
			fmt.Printf("callback return services:%s \n\n", util.ToJsonString(services))
		},
	}
	ExampleServiceClient_Subscribe(client, param)

	ExampleServiceClient_RegisterServiceInstance(client, vo.RegisterInstanceParam{
		Ip:          "10.0.0.10",
		Port:        8848,
		ServiceName: "demo.go",
		GroupName:   "group-a",
		ClusterName: "cluster-a",
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"idc": "beijing"},
	})
	//wait for client pull change from server
	time.Sleep(3 * time.Second)

	ExampleServiceClient_UpdateServiceInstance(client, vo.UpdateInstanceParam{
		Ip:          "10.0.0.11", //update ip
		Port:        8848,
		ServiceName: "demo.go",
		GroupName:   "group-a",
		ClusterName: "cluster-a",
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"idc": "beijing1"}, //update metadata
	})

	verifier.WaitAndAssertMetrics(map[string]func(metricdata.ResourceMetrics){
		"nacos.client.serviceinfo.size": func(metrics metricdata.ResourceMetrics) {
			if len(metrics.ScopeMetrics) == 0 {
				panic("should not be empty metrics")
			}
		},
		"nacos.client.dombeat.size": func(metrics metricdata.ResourceMetrics) {
			if len(metrics.ScopeMetrics) == 0 {
				panic("should not be empty metrics")
			}
		},
		"nacos.client.naming.request.duration": func(metrics metricdata.ResourceMetrics) {
			if len(metrics.ScopeMetrics) == 0 {
				panic("should not be empty metrics")
			}
			point := metrics.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[float64])
			if point.DataPoints[0].Count <= 0 {
				panic("nacos.client.naming.request.duration should not be negative")
			}
		},
	})
}

func ExampleServiceClient_RegisterServiceInstance(client naming_client.INamingClient, param vo.RegisterInstanceParam) {
	success, _ := client.RegisterInstance(param)
	fmt.Printf("RegisterServiceInstance,param:%+v,result:%+v \n\n", param, success)
}

func ExampleServiceClient_DeRegisterServiceInstance(client naming_client.INamingClient, param vo.DeregisterInstanceParam) {
	success, _ := client.DeregisterInstance(param)
	fmt.Printf("DeRegisterServiceInstance,param:%+v,result:%+v \n\n", param, success)
}

func ExampleServiceClient_UpdateServiceInstance(client naming_client.INamingClient, param vo.UpdateInstanceParam) {
	success, _ := client.UpdateInstance(param)
	fmt.Printf("UpdateServiceInstance,param:%+v,result:%+v \n\n", param, success)
}

func ExampleServiceClient_GetService(client naming_client.INamingClient, param vo.GetServiceParam) {
	service, _ := client.GetService(param)
	fmt.Printf("GetService,param:%+v, result:%+v \n\n", param, service)
}

func ExampleServiceClient_SelectAllInstances(client naming_client.INamingClient, param vo.SelectAllInstancesParam) {
	instances, _ := client.SelectAllInstances(param)
	fmt.Printf("SelectAllInstance,param:%+v, result:%+v \n\n", param, instances)
}

func ExampleServiceClient_SelectInstances(client naming_client.INamingClient, param vo.SelectInstancesParam) {
	instances, _ := client.SelectInstances(param)
	fmt.Printf("SelectInstances,param:%+v, result:%+v \n\n", param, instances)
}

func ExampleServiceClient_SelectOneHealthyInstance(client naming_client.INamingClient, param vo.SelectOneHealthInstanceParam) {
	instances, _ := client.SelectOneHealthyInstance(param)
	fmt.Printf("SelectOneHealthyInstance,param:%+v, result:%+v \n\n", param, instances)
}

func ExampleServiceClient_Subscribe(client naming_client.INamingClient, param *vo.SubscribeParam) {
	client.Subscribe(param)
}
