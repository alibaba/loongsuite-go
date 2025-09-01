package k8s_client_go

import (
	"context"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"time"
)

type K8sOtelEventHandler struct {
	ctx    context.Context
	origin cache.ResourceEventHandler
}

func NewK8sOtelEventHandler(ctx context.Context, origin cache.ResourceEventHandler) K8sOtelEventHandler {
	return K8sOtelEventHandler{ctx: ctx, origin: origin}
}

func (k K8sOtelEventHandler) OnAdd(obj interface{}, isInInitialList bool) {
	eventInfo := abstractK8sEventInfo(obj, "Added")
	ctx := k8sClientGoEventInstrumenter.Start(k.ctx, eventInfo)
	k.origin.OnAdd(obj, isInInitialList)
	duration := time.Since(eventInfo.startTime)
	eventInfo.processingTime = duration.Microseconds()
	k8sClientGoEventInstrumenter.End(ctx, eventInfo, eventInfo, nil)
}

func (k K8sOtelEventHandler) OnUpdate(oldObj, newObj interface{}) {
	eventInfo := abstractK8sEventInfo(newObj, "Updated")
	ctx := k8sClientGoEventInstrumenter.Start(k.ctx, eventInfo)
	k8sClientGoEventInstrumenter.End(ctx, eventInfo, eventInfo, nil)
	k.origin.OnUpdate(oldObj, newObj)
	duration := time.Since(eventInfo.startTime)
	eventInfo.processingTime = duration.Microseconds()
	k8sClientGoEventInstrumenter.End(ctx, eventInfo, eventInfo, nil)
}

func (k K8sOtelEventHandler) OnDelete(obj interface{}) {
	eventInfo := abstractK8sEventInfo(obj, "Deleted")
	ctx := k8sClientGoEventInstrumenter.Start(k.ctx, eventInfo)
	k8sClientGoEventInstrumenter.End(ctx, eventInfo, eventInfo, nil)
	k.origin.OnDelete(obj)
	duration := time.Since(eventInfo.startTime)
	eventInfo.processingTime = duration.Microseconds()
	k8sClientGoEventInstrumenter.End(ctx, eventInfo, eventInfo, nil)
}

func abstractK8sEventInfo(obj interface{}, eventType string) k8sEventInfo {
	eventInfo := k8sEventInfo{}
	eventInfo.startTime = time.Now()
	eventInfo.eventType = eventType
	if m, err := meta.Accessor(obj); err == nil {
		eventInfo.name = m.GetName()
		eventInfo.namespace = m.GetNamespace()
		eventInfo.eventUID = string(m.GetUID())
		eventInfo.resourceVersion = m.GetResourceVersion()
	}
	if gvks, _, err := scheme.Scheme.ObjectKinds(obj.(runtime.Object)); err == nil && len(gvks) > 0 {
		gvk := gvks[0]
		eventInfo.apiVersion = gvk.Version
		eventInfo.apiVersion = gvk.Group + "/" + gvk.Version
		eventInfo.kind = gvk.Kind
	}
	return eventInfo
}
