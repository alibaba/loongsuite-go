## **consume module**

监听github.com/rabbitmq/amqp091-go下consumers类的send方法，因为此处message未放入chan中。此模块只是监测消息接收，对于收到后消息的具体处理，可以根据需求在自己项目中加入自定义监测。

## **publish module**

监听github.com/rabbitmq/amqp091-go下Channel类的PublishWithDeferredConfirm方法。Headers作为传递traceparent的媒介用于关联publish和consume。当使用者需要使publish和consume的追踪关联时，在调用publish时要保证所传参数amqp.Publishing的Headers不能为nil。
