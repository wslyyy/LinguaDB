# 🧠 LinguaDB
⚡一个类似llama_index的极简GO版本框架(Qdrant + Embedding + Openai + Gin)，本地知识库QA应用后端⚡

## Show

## 👋 How to start

### 准备数据
├── data  
│ &emsp;&emsp;└── filename  
│ &emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;   ├── subtitle1.txt  
│ &emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;   ├── subtitle2.txt  
│ &emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;   ├── subtitle3.txt  
│ &emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;   ├── subtitle4.txt  
│ &emsp;&emsp;└── filename2  
│ &emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;   ├── subtitle1.txt  
│ &emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;   ├── subtitle2.txt

### 启动Qdrant向量数据库
`docker run -p 6333:6333 -p 6334:6334
-v $(pwd)/qdrant_storage:/qdrant/storage
qdrant/qdrant`

### 编辑配置文件
`vim config.yaml`

### 启动服务
`go mod tidy`  
`go run main.go`

### API

#### 录入文档至文档库
[POST] http://127.0.0.1:8000/insert

```json
{
    "dbName":"xxx",
    "dirName":"xxx"
}
```

#### 删除文档库
[POST] http://127.0.0.1:8000/deleteDB

```json
{
    "dbName":"xxx"
}
```

#### 🥳 自然语言查询
[POST] http://127.0.0.1:8000/query
```json
{
  "dbName": "xxx",
  "useCache": false,
  "question": "xxx"
}
```

## ⏳ ToDoList
- [ ] 录入所有人工分块的文档
- [x] 暴露HTTP服务给客户端(Gin)
- [ ] 暴露RPC服务给客户端(Grpc)
- [x] 实现录入->批量向量化->批量写入向量库(Qdrant)->检出最佳嵌入Embedding(Cache?)->拼接prompts->ChatGPT总结的全流程逻辑（Go实现）
- [x] 问题-答案对入库(mysql/qdrant)，增加一级缓存，用户查询时计算问题和已有问题向量相似度，高于阈值直接取答案
- [x] 测试各类问题回答效果
- [ ] 根据问题能够自动判断是否使用缓存策略（TODO）

## 🤔 问题与思考
1. 最初版本文档划分块的算法同llama_index/langchain思想类似，根据token限制划分固定大小的chunk，导致嵌入时信息不全或不相关信息过多，可探索基于语义段落的划分算法或人工划分🐶
2. openai的embedding接口效果用来语义匹配还行，但是对于中文并不是很理想
3. 启用缓存具有badcase，在涉及计算推理时，由于缓存问题和提问问题很相似，只有数值的改变，但会命中缓存，不会走openai，导致回答错误

## 💤 优化方向
1. 优化 text_split 算法，使匹配出的结果作为上下文时能够提供更合理的推理/回答依据，采用人工分块，提升质量，节省token ☑️
2. 优化 设计理念，目前思路是问题和答案的向量化匹配，可以采用问题和问题向量化匹配并缓存，以此提升语义匹配效果 ☑️
3. 优化 embedding 模型，可以采用本地中文llm(TextToVec)，提升语义向量化的效果，使得语义匹配过程中能够匹配出最满足要求的文本段落作为上下文
4. 优化 LLM 模型，可以使用GPT4接口，使得给定提问相同情况下，得到更理想的推理/回答结果

