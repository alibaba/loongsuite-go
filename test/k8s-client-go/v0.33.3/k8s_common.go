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

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	globalPodAddedCh   chan struct{}
	globalPodUpdatedCh chan struct{}
	globalPodDeletedCh chan struct{}
)

func InjectEventChannels(add, update, del chan struct{}) {
	globalPodAddedCh = add
	globalPodUpdatedCh = update
	globalPodDeletedCh = del
}

func CreateK8sClient(stopCh chan struct{}) *kubernetes.Clientset {
	kubeconfigYaml := []byte(os.Getenv("KUBECONFIG"))
	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigYaml)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		panic(err)
	}

	factory := informers.NewSharedInformerFactory(clientset, 30*time.Second)
	podInformer := factory.Core().V1().Pods().Informer()

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			fmt.Printf("[Informer] Pod added: %s/%s\n", pod.Namespace, pod.Name)
			if pod.Name == "otel-demo-pod" && globalPodAddedCh != nil {
				close(globalPodAddedCh)
				globalPodAddedCh = nil
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newPod := newObj.(*corev1.Pod)
			fmt.Printf("[Informer] Pod updated: %s/%s\n", newPod.Namespace, newPod.Name)
			if newPod.Name == "otel-demo-pod" && globalPodUpdatedCh != nil {
				close(globalPodUpdatedCh)
				globalPodUpdatedCh = nil
			}
		},
		DeleteFunc: func(obj interface{}) {
			var pod *corev1.Pod
			switch t := obj.(type) {
			case *corev1.Pod:
				pod = t
			case cache.DeletedFinalStateUnknown:
				var ok bool
				pod, ok = t.Obj.(*corev1.Pod)
				if !ok {
					fmt.Println("DeletedFinalStateUnknown contains non-Pod object")
					return
				}
			default:
				fmt.Println("Unknown type in DeleteFunc")
				return
			}
			fmt.Printf("[Informer] Pod deleted: %s/%s\n", pod.Namespace, pod.Name)
			if pod.Name == "otel-demo-pod" && globalPodDeletedCh != nil {
				close(globalPodDeletedCh)
				globalPodDeletedCh = nil
			}
		},
	})

	go factory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh, podInformer.HasSynced) {
		log.Fatalf("failed to sync informer cache")
	}

	return clientset
}
