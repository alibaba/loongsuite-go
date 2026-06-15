## **angentAction模块**

github.com/tmc/langchaingo/agents下监听Executor下的doAction的方法，Executor作为agent的执行器，doAction方法为Executor调用agents下根据决策调用每个工具类的位置。所以agentAction监听的实际是agent对于每个工具使用，也就是agent的每一次动作。之所以没有监听工具模块，因为工具以接口方式实现过于细化而不可能一一监控。

## **chains模块**

github.com/tmc/langchaingo/chains下监听callChain方法，chains的call，run，predict最终都会到callChain处，callChain再调用对应的llm的call方法，当然这里个别会有例外，例如chains.NewConversation().Call()这个方法，就会脱离chain直接调用llms下的GenerateFromSinglePrompt直接对接模型获取消息。

## **Embed模块**

github.com/tmc/langchaingo/embeddings下监听了EmbedQuery和batchedEmbedOnEnter两个方法，目前对于由embeddings.NewEmbedder方式创建的嵌入器可以监听，但对于如voyageai.NewVoyageAI()等创建的方式无法监听

## **llmGenerateSingle模块**

github.com/tmc/langchaingo/llm下监听GenerateFromSinglePrompt方法，该方法用于调用具有单个字符串提示符的LLM，期望单个字符串响应。langchain-go中大部分模型接口的call方法都调用这个方法，再通过这个方法调用GenerateContent。因为后面实现了对于具体模型接口的监听，此块监听有点多余，但模型接口众多需要一一监听且还未实现完全，所以此模块作为预留备案

## **relevantDocuments模块**

github.com/tmc/langchaingo/vectorstores下监听GetRelevantDocuments方法，该方法作为Retriever 获取关联文档的方法，如果直接调用向量数据库自己本身的的SimilaritySearc方法是监听不到的。vectorstores.ToRetriever(db, 1).GetRelevantDocuments()这种方式才可以。

## **llm模型接口（目前只包监听了ollama和openai接口，其他后续补充）**

### ollama：
监听github.com/tmc/langchaingo/llms/ollama下GenerateContent方法目前模型response结果只统计TotalTokens值，request值根据填入而定

### openai：
监听github.com/tmc/langchaingo/llms/openai下GenerateContent方法目前response结果只统计TotalTokens 值、ReasoningTokens值，request值根据填入而定
