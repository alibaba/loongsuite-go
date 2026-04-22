package consumer

import (
	"context"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

//go:linkname consumeChannelOnEnter github.com/streadway/amqp.consumeChannelOnEnter
func consumeChannelOnEnter(call api.CallContext, ch *amqp.Channel, queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) {
	// no-op; wraps delivery channel on exit
}

//go:linkname consumeChannelOnExit github.com/streadway/amqp.consumeChannelOnExit
func consumeChannelOnExit(call api.CallContext, deliveries <-chan amqp.Delivery, err error) {
	if deliveries == nil {
		return
	}
	defer func() {
		if recover() != nil {
			call.SetReturnVal(0, deliveries)
		}
	}()

	wrpCh := make(chan amqp.Delivery)
	go func() {
		defer close(wrpCh)
		for d := range deliveries {
			spanStartAndEnd(d)
			wrpCh <- d
		}
	}()
	call.SetReturnVal(0, (<-chan amqp.Delivery)(wrpCh))
}

func spanStartAndEnd(delivery amqp.Delivery) {
	ctx := context.Background()
	if delivery.Headers != nil {
		ctx = otel.GetTextMapPropagator().Extract(ctx, amqpConsumerTextMapCarrier{headers: delivery.Headers})
	}
	opts := []oteltrace.SpanStartOption{oteltrace.WithSpanKind(oteltrace.SpanKindConsumer)}
	_, span := otel.Tracer("github.com/streadway/amqp").Start(ctx, delivery.Exchange+" receive", opts...)
	span.SetAttributes(
		semconv.MessagingSystemRabbitmq,
		semconv.MessagingDestinationName(delivery.Exchange),
		attribute.String("messaging.operation.name", "receive"),
		semconv.MessagingRabbitmqDestinationRoutingKey(delivery.RoutingKey),
	)
	span.End()
}
