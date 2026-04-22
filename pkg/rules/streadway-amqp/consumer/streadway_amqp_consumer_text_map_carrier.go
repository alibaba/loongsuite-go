package consumer

import "github.com/streadway/amqp"

type amqpConsumerTextMapCarrier struct {
	headers amqp.Table
}

func (rtmc amqpConsumerTextMapCarrier) Get(key string) string {
	if v, ok := rtmc.headers[key]; ok {
		if vs, ok2 := v.(string); ok2 {
			return vs
		}
	}
	return ""
}

func (rtmc amqpConsumerTextMapCarrier) Set(key string, value string) {
	rtmc.headers[key] = value
}

func (rtmc amqpConsumerTextMapCarrier) Keys() []string {
	keys := make([]string, 0, len(rtmc.headers))
	for k := range rtmc.headers {
		keys = append(keys, k)
	}
	return keys
}
