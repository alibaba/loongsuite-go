## **consume module**

Listen to the send method of the consumers class under github.com/rabbitmq/amqp091-go, as the message is not put into the channel at this point. This module only monitors message reception. For specific processing of received messages, you can add custom monitoring in your own project as needed.
## **publish module**

Monitor the PublishWithDeferredConfirm method of the Channel class in github.com/rabbitmq/amqp091-go. The Headers field serves as a medium for passing traceparent to associate publish and consume. When users need to link publish and consume traces, they must ensure that the Headers field in the amqp.Publishing parameters is not nil when calling publish.

 