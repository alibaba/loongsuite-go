## **angentAction module**

Listen to the doAction method under Executor in github.com/tmc/langchaingo/agents. As the executor of the agent, Executor calls the doAction method to invoke the corresponding tool classes under agents based on decision-making. Therefore, agentAction essentially listens to each action the agent takes when using a tool. The reason why the tool module is not monitored is that tools are implemented as interfaces, making them too granular to be individually monitored.

## **chains module**

Listen to the callChain method under github.com/tmc/langchaingo/chains. The call, run, and predict methods of chains will eventually reach callChain, which then calls the corresponding LLM’s call method. However, there are exceptions, such as the chains.NewConversation().Call() method, which bypasses the chain and directly calls GenerateFromSinglePrompt under llms to interact with the model and obtain a response.

## **Embed module**

Listen to the EmbedQuery and batchedEmbedOnEnter methods under github.com/tmc/langchaingo/embeddings. Currently, embedding instances created using embeddings.NewEmbedder can be monitored, but instances created using methods like voyageai.NewVoyageAI() cannot be monitored.

## **llmGenerateSingle module**

Listen to the GenerateFromSinglePrompt method under github.com/tmc/langchaingo/llm. This method is used to invoke an LLM with a single string prompt, expecting a single string response. In langchain-go, most model interface call methods invoke this method, which in turn calls GenerateContent. Since monitoring for specific model interfaces has been implemented, this monitoring is somewhat redundant. However, given the numerous model interfaces that require individual monitoring and the fact that implementation is not yet complete, this module is retained as a backup.

## **relevantDocuments module**

Listen to the GetRelevantDocuments method under github.com/tmc/langchaingo/vectorstores. This method is used by the Retriever to fetch relevant documents. If the vector database’s own SimilaritySearch method is called directly, it cannot be monitored. Only calls made using vectorstores.ToRetriever(db, 1).GetRelevantDocuments() can be detected.

## **LLM model interfaces (currently only monitoring ollama and openai interfaces).**

### ollama：

Listen to the GenerateContent method under github.com/tmc/langchaingo/llms/ollama. Currently, the model response results only track the TotalTokens value, while the request values depend on the input.

### openai：

Listen to the GenerateContent method under github.com/tmc/langchaingo/llms/openai. Currently, the response results only track the TotalTokens and ReasoningTokens values, while the request values depend on the input.