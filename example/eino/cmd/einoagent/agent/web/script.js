document.addEventListener('DOMContentLoaded', () => {
    const messageInput = document.getElementById('message-input');
    const sendButton = document.getElementById('send-button');
    const chatMessages = document.getElementById('chat-messages');
    const logMessages = document.getElementById('log-messages');
    const chatHistory = document.getElementById('chat-history');
    const rightPanel = document.getElementById('right-panel');
    const togglePanel = document.getElementById('toggle-panel');
    const newChatButton = document.getElementById('new-chat');

    let chatId = uuidv4();
    let currentConversation = null;
    let abortController = null;  // 用于取消请求

    // 创建取消按钮
    const cancelButton = document.createElement('button');
    cancelButton.id = 'cancel-button';
    cancelButton.className = 'w-8 h-8 rounded-full bg-red-500 text-white flex items-center justify-center hover:bg-red-600 hidden absolute left-1/2 -translate-x-1/2 -top-10';
    cancelButton.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>';
    messageInput.parentElement.style.position = 'relative';
    messageInput.parentElement.appendChild(cancelButton);

    // 取消按钮点击事件
    cancelButton.addEventListener('click', () => {
        if (abortController) {
            abortController.abort();
            abortController = null;
        }

        // 隐藏取消按钮，显示发送按钮
        cancelButton.classList.add('hidden');
        sendButton.classList.remove('hidden');
        // 重新启用输入框
        messageInput.disabled = false;
        sendButton.disabled = false;
        sendButton.classList.remove('opacity-50');
    });

    // 配置 marked
    marked.setOptions({
        highlight: function(code, language) {
            if (Prism.languages[language]) {
                return Prism.highlight(code, Prism.languages[language], language);
            }
            return code;
        },
        breaks: true,
        gfm: true
    });

    // 处理消息内容的函数
    function processMessageContent(content) {
        return content;  // 直接返回原始内容
    }

    // 添加复制按钮到代码块
    function addCopyButtons() {
        // 只选择包含 code 标签的 pre 元素
        document.querySelectorAll('pre code').forEach(code => {
            const pre = code.parentElement;
            if (pre.classList.contains('copy-button-added')) {
                return;
            }
            pre.classList.add('copy-button-added');

            const button = document.createElement('button');
            button.className = 'copy-button';
            button.textContent = 'Copy';

            button.addEventListener('click', async () => {
                try {
                    await navigator.clipboard.writeText(code.textContent);
                    button.textContent = 'Copied!';
                    button.classList.add('copied');
                    setTimeout(() => {
                        button.textContent = 'Copy';
                        button.classList.remove('copied');
                    }, 2000);
                } catch (err) {
                    console.error('Failed to copy:', err);
                    button.textContent = 'Failed';
                    setTimeout(() => {
                        button.textContent = 'Copy';
                    }, 2000);
                }
            });

            pre.insertBefore(button, pre.firstChild);
        });
    }

    // 面板控制
    document.querySelectorAll('.panel-toggle').forEach(button => {
        button.addEventListener('click', () => {
            const target = document.getElementById(button.dataset.target);
            const panel = button.closest('.panel');
            const icon = button.querySelector('svg');

            if (target.id === 'task-content') {
                // Task panel 使用高度控制
                if (panel.style.flex === '0 1 48px') {
                    panel.style.flex = '1 1 60%';
                    icon.style.transform = 'rotate(180deg)';
                } else {
                    panel.style.flex = '0 1 48px';
                    icon.style.transform = 'rotate(0deg)';
                }
            } else {
                // Log panel 高度收缩
                if (panel.style.flex === '0 1 48px') {
                    panel.style.flex = '1 1 300px';
                    icon.style.transform = 'rotate(180deg)';
                } else {
                    panel.style.flex = '0 1 48px';
                    icon.style.transform = 'rotate(0deg)';
                }
            }
        });
    });

    // 右侧面板切换
    togglePanel.addEventListener('click', () => {
        rightPanel.classList.toggle('w-0');
        rightPanel.classList.toggle('w-[500px]');
        rightPanel.classList.toggle('opacity-0');
        rightPanel.classList.toggle('opacity-100');
    });

    // 新建对话
    newChatButton.addEventListener('click', () => {
        chatId = uuidv4();
        currentConversation = null;
        chatMessages.innerHTML = '';
        messageInput.value = '';

        // 创建新的历史记录项
        const historyItem = document.createElement('div');
        historyItem.className = 'chat-item p-3 hover:bg-gray-100 cursor-pointer rounded-lg mb-2 transition-colors flex justify-between items-start';
        historyItem.dataset.chatId = chatId;
        historyItem.innerHTML = `
            <div class="flex-1 min-w-0 mr-2" onclick="event.stopPropagation()">
                <div class="font-medium text-gray-900 truncate">Empty</div>
                <div class="text-sm text-gray-500">ID: ${chatId.substring(0, 8)}...</div>
            </div>
            <button class="delete-chat p-1 hover:bg-red-100 rounded-lg transition-colors" onclick="event.stopPropagation()">
                <svg class="w-5 h-5 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
            </button>
        `;

        // 添加删除按钮事件
        const deleteButton = historyItem.querySelector('.delete-chat');
        deleteButton.addEventListener('click', (e) => {
            e.stopPropagation();
            deleteConversation(chatId, historyItem);
        });

        // 添加点击事件
        historyItem.querySelector('.flex-1').addEventListener('click', () => loadConversation(chatId));

        // 将新对话添加到列表顶部
        if (chatHistory.firstChild) {
            chatHistory.insertBefore(historyItem, chatHistory.firstChild);
        } else {
            chatHistory.appendChild(historyItem);
        }

        highlightCurrentChat();
    });

    // 设置 SSE 日志监听
    let isAutoScrollLog = true;

    function connectLogStream() {
        console.log('Connecting to log stream...');
        const logSource = new EventSource('/agent/api/log');

        logSource.onmessage = (event) => {
            const logMessage = event.data;
            const wasAtBottom = isAutoScrollLog;

            // 创建新的日志行
            const logLine = document.createElement('div');
            logLine.className = 'log-line';
            logLine.textContent = logMessage;
            logMessages.appendChild(logLine);

            // 保持最新的1000行日志
            const maxLogLines = 1000;
            while (logMessages.children.length > maxLogLines) {
                logMessages.removeChild(logMessages.firstChild);
            }

            // 如果之前在底部，则自动滚动到新消息
            if (wasAtBottom) {
                logMessages.scrollTop = logMessages.scrollHeight;
            }
        };

        logSource.onerror = (error) => {
            console.error('Log SSE Error:', error);
            logSource.close();

            // 3秒后尝试重连
            console.log('Reconnecting in 3 seconds...');
            setTimeout(connectLogStream, 3000);
        };

        logSource.onopen = () => {
            console.log('Log stream connected');
        };
    }

    // 监听日志面板的滚动事件
    logMessages.addEventListener('scroll', () => {
        const scrollBottom = logMessages.scrollHeight - logMessages.clientHeight;
        isAutoScrollLog = Math.abs(scrollBottom - logMessages.scrollTop) < 2;
    });

    // 初始连接
    connectLogStream();

    function highlightCurrentChat() {
        document.querySelectorAll('.chat-item').forEach(item => {
            item.classList.remove('bg-blue-50');
        });
        const currentItem = document.querySelector(`[data-chat-id="${chatId}"]`);
        if (currentItem) {
            currentItem.classList.add('bg-blue-50');
        }
    }

    // 删除对话
    async function deleteConversation(id, element) {
        try {
            const response = await fetch(`/agent/api/history?id=${id}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                element.remove();
                if (id === chatId) {
                    chatId = uuidv4();
                    currentConversation = null;
                    chatMessages.innerHTML = '';
                }
            } else {
                console.error('Failed to delete conversation');
            }
        } catch (error) {
            console.error('Error deleting conversation:', error);
        }
    }

    // 加载对话
    async function loadConversation(id) {
        try {
            const response = await fetch(`/agent/api/history?id=${id}`);
            const data = await response.json();

            if (data.conversation) {
                chatId = id;
                currentConversation = data.conversation;
                chatMessages.innerHTML = '';

                data.conversation.messages.forEach(msg => {
                    appendMessage(msg.content, msg.role === 'user', false);
                });

                highlightCurrentChat();
            }
        } catch (error) {
            console.error('Error loading conversation:', error);
        }
    }

    // 添加消息到聊天区域
    function appendMessage(content, isUser, animate = true) {
        const processedContent = processMessageContent(content);
        const messageDiv = document.createElement('div');
        messageDiv.className = 'flex items-start gap-3 mb-4';

        // 添加头像
        const avatar = document.createElement('div');
        avatar.className = 'w-8 h-8 flex items-center justify-center rounded-full bg-gray-100 flex-shrink-0';
        avatar.textContent = isUser ? '🧑‍💻' : '🤖';
        messageDiv.appendChild(avatar);

        // 消息内容
        const contentDiv = document.createElement('div');
        contentDiv.className = `message markdown-body rounded-lg p-4 ${isUser ? 'bg-gray-100' : 'bg-gray-50'}`;

        if (!animate || isUser) {
            contentDiv.innerHTML = marked.parse(processedContent);
            addCopyButtons();
        } else {
            const typingDiv = document.createElement('div');
            contentDiv.appendChild(typingDiv);

            new Typed(typingDiv, {
                strings: [marked.parse(processedContent)],
                typeSpeed: 20,
                showCursor: false,
                onComplete: () => addCopyButtons()
            });
        }

        messageDiv.appendChild(contentDiv);
        chatMessages.appendChild(messageDiv);
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }

    async function sendMessage() {
        const message = messageInput.value.trim();
        if (!message) return;

        // 如果是新对话的第一条消息，更新历史记录标题
        const historyItem = document.querySelector(`[data-chat-id="${chatId}"]`);
        if (historyItem && historyItem.querySelector('.font-medium').textContent === 'Empty') {
            historyItem.querySelector('.font-medium').textContent = message;
        }

        appendMessage(message, true);
        messageInput.value = '';

        // 禁用输入框和发送按钮，显示取消按钮
        messageInput.disabled = true;
        sendButton.disabled = true;
        sendButton.classList.add('opacity-50');
        sendButton.classList.add('hidden');
        cancelButton.classList.remove('hidden');

        try {
            console.log('Starting chat with ID:', chatId);

            let hasReceivedMessage = false;
            let currentMessageDiv = null;
            let contentDiv = null;
            let accumulatedContent = '';
            let isFirstChunk = true;
            let lastRenderTime = 0;

            // 创建新的 AbortController
            abortController = new AbortController();

            // 使用 fetch 替代 EventSource，添加 signal
            const response = await fetch(`/agent/api/chat?id=${chatId}&message=${encodeURIComponent(message)}`, {
                signal: abortController.signal
            });

            // 检查响应状态
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const reader = response.body.getReader();
            const decoder = new TextDecoder();
            let buffer = '';  // 用于存储不完整的 SSE 消息

            try {
                while (true) {
                    const {value, done} = await reader.read();
                    if (done) break;

                    // 解码新的数据块并添加到缓冲区
                    buffer += decoder.decode(value, {stream: true});

                    // 按行分割并处理每一行
                    const lines = buffer.split(/\r\n|\r|\n/);
                    // 保留最后一个可能不完整的行
                    buffer = lines.pop() || '';

                    for (const line of lines) {
                        // 解析 SSE 格式的行
                        if (line.startsWith('data:')) {
                            // 保留 data: 后的所有内容，包括前导空格
                            const rawData = line.slice(5);  // 直接截取 'data:' 后的内容
                            // console.log(`Raw SSE data: |${rawData}|`);
                            hasReceivedMessage = true;

                            // 如果是第一个 chunk，创建新的消息框
                            if (isFirstChunk) {
                                const messageDiv = document.createElement('div');
                                messageDiv.className = 'flex items-start gap-3 mb-4';

                                // 添加头像
                                const avatar = document.createElement('div');
                                avatar.className = 'w-8 h-8 flex items-center justify-center rounded-full bg-gray-100 flex-shrink-0';
                                avatar.textContent = '🤖';
                                messageDiv.appendChild(avatar);

                                // 消息内容
                                contentDiv = document.createElement('div');
                                contentDiv.className = 'message markdown-body rounded-lg p-4 bg-gray-50';
                                messageDiv.appendChild(contentDiv);
                                chatMessages.appendChild(messageDiv);

                                currentMessageDiv = contentDiv;
                                isFirstChunk = false;
                                accumulatedContent = rawData;
                            } else {
                                if (rawData === '') {
                                    // 如果是空数据，添加换行符
                                    accumulatedContent += '\n';
                                } else {
                                    // 否则直接拼接数据
                                    accumulatedContent += rawData;
                                }
                            }

                            // 限制渲染频率
                            const now = Date.now();
                            if (now - lastRenderTime >= 100) {
                                renderContent();
                            } else {
                                clearTimeout(window.renderTimeout);
                                window.renderTimeout = setTimeout(renderContent, 100 - (now - lastRenderTime));
                            }
                        }
                    }
                }

                function renderContent() {
                    currentMessageDiv.innerHTML = marked.parse(accumulatedContent);
                    addCopyButtons();
                    chatMessages.scrollTop = chatMessages.scrollHeight;
                    lastRenderTime = Date.now();
                }

                // 请求完成后，隐藏取消按钮，显示发送按钮
                cancelButton.classList.add('hidden');
                sendButton.classList.remove('hidden');
                abortController = null;

            } finally {
                // 确保读取器被正确关闭
                reader.cancel();
            }

        } catch (error) {
            console.error('Error sending message:', error);
            if (error.name === 'AbortError') {
            } else {
                appendMessage('Error: Failed to send message. Please try again.', false);
            }

            abortController = null;
        } finally {
            messageInput.disabled = false;
            sendButton.disabled = false;
            sendButton.classList.remove('opacity-50');
            sendButton.classList.remove('hidden');
            cancelButton.classList.add('hidden');
        }
    }

    sendButton.addEventListener('click', sendMessage);
    messageInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            sendMessage();
        }
    });

    // 加载历史对话列表
    fetch('/agent/api/history')
        .then(response => response.json())
        .then(data => {
            if (data.ids && data.ids.length > 0) {
                chatHistory.innerHTML = ''; // 清空现有历史
                let firstConversationId = null;

                const loadPromises = data.ids.map(id =>
                    fetch(`/agent/api/history?id=${id}`)
                        .then(response => response.json())
                        .then(convData => {
                            if (convData.conversation) {
                                const firstMessage = convData.conversation.messages.length > 0
                                    ? convData.conversation.messages[0].content
                                    : 'Empty';

                                const historyItem = document.createElement('div');
                                historyItem.className = 'chat-item p-3 hover:bg-gray-100 cursor-pointer rounded-lg mb-2 transition-colors flex justify-between items-start';
                                historyItem.dataset.chatId = id;
                                historyItem.innerHTML = `
                                    <div class="flex-1 min-w-0 mr-2" onclick="event.stopPropagation()">
                                        <div class="font-medium text-gray-900 truncate">${firstMessage}</div>
                                        <div class="text-sm text-gray-500">ID: ${id.substring(0, 8)}...</div>
                                    </div>
                                    <button class="delete-chat p-1 hover:bg-red-100 rounded-lg transition-colors" onclick="event.stopPropagation()">
                                        <svg class="w-5 h-5 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                        </svg>
                                    </button>
                                `;

                                const deleteButton = historyItem.querySelector('.delete-chat');
                                deleteButton.addEventListener('click', (e) => {
                                    e.stopPropagation();
                                    deleteConversation(id, historyItem);
                                });

                                historyItem.querySelector('.flex-1').addEventListener('click', () => loadConversation(id));
                                chatHistory.appendChild(historyItem);

                                // 记录第一个对话的ID
                                if (firstConversationId === null) {
                                    firstConversationId = id;
                                }
                            }
                        })
                );

                // 等待所有历史记录加载完成后，加载第一个对话
                Promise.all(loadPromises).then(() => {
                    if (firstConversationId) {
                        loadConversation(firstConversationId);
                    }
                });
            }
        })
        .catch(error => console.error('Error loading history:', error));
}); 