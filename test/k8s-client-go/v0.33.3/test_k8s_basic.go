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
	"context"
	"fmt"
	"log"
	"time"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	stopCh := make(chan struct{})
	defer close(stopCh)

	ctx := context.Background()
	clientset := CreateK8sClient(stopCh)

	podAddedCh := make(chan struct{})
	podUpdatedCh := make(chan struct{})
	podDeletedCh := make(chan struct{})

	InjectEventChannels(podAddedCh, podUpdatedCh, podDeletedCh)

	testPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "otel-demo-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "60"},
				},
			},
		},
	}

	_, err := clientset.CoreV1().Pods("default").Create(ctx, testPod, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("failed to create pod: %v", err)
	}

	select {
	case <-podAddedCh:
		fmt.Println("Pod added event received")
	case <-time.After(10 * time.Second):
		log.Fatal("timeout waiting for pod Added event")
	}

	select {
	case <-podUpdatedCh:
		fmt.Println("Pod updated event received")
	case <-time.After(10 * time.Second):
		log.Fatal("timeout waiting for pod Updated event")
	}

	deletePolicy := metav1.DeletePropagationForeground
	err = clientset.CoreV1().Pods("default").Delete(ctx, testPod.Name, metav1.DeleteOptions{
		GracePeriodSeconds: new(int64),
		PropagationPolicy:  &deletePolicy,
	})
	if err != nil {
		log.Fatalf("failed to delete test pod: %v", err)
	}

	select {
	case <-podDeletedCh:
		fmt.Println("Pod deleted event received")
	case <-time.After(10 * time.Second):
		log.Fatal("timeout waiting for pod Deleted event")
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		if len(stubs) == 0 {
			log.Fatal("No traces found")
		}

		var hasPodAdded, hasPodUpdated, hasPodDelete bool

		for _, stub := range stubs {
			for _, span := range stub {
				if span.Name != "k8s.informer.Pod.process" {
					continue
				}

				var podName, eventType string
				for _, attr := range span.Attributes {
					if attr.Key == "k8s.object.name" {
						podName = attr.Value.AsString()
					}
					if attr.Key == "k8s.event.type" {
						eventType = attr.Value.AsString()
					}
				}

				if podName == "otel-demo-pod" {
					switch eventType {
					case "Added":
						hasPodAdded = true
						verifier.VerifyK8sPodEventAttributes(span, "Added", "otel-demo-pod", "Pod", "default", "/v1")
					case "Updated":
						hasPodUpdated = true
						verifier.VerifyK8sPodEventAttributes(span, "Updated", "otel-demo-pod", "Pod", "default", "/v1")
					case "Deleted":
						hasPodDelete = true
						verifier.VerifyK8sPodEventAttributes(span, "Deleted", "otel-demo-pod", "Pod", "default", "/v1")
					}
				}
			}
		}

		if !hasPodAdded {
			log.Fatal("Expected 'Added' event for pod 'otel-demo-pod' not found")
		}
		if !hasPodUpdated {
			log.Fatal("Expected 'Updated' event for pod 'otel-demo-pod' not found")
		}
		if !hasPodDelete {
			log.Fatal("Expected 'Deleted' event for pod 'otel-demo-pod' not found")
		}

		log.Println("All expected events found: Added, Updated, Deleted")
	}, 11)
}
