package lingua

import (
	"LinguaDB/Prompt"
	"LinguaDB/initialization"
	"LinguaDB/server"
	"LinguaDB/storage"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"log"
)

func Query(config initialization.Config, Q string, useCache bool, dbName string) (string, string, error) {
	qdrant := server.NewQdrant(config.QdrantAddrGrpc, config.QdrantAddrHttp, dbName, uint64(1536))
	clientToUse := initialization.NewOpenAIClient(config)

	QEmbedding, err := server.GetEmbedding(clientToUse, Q, openai.AdaEmbeddingV2)
	if err != nil {
		return "", "", err
	}

	qaStorage := storage.QAStorage{
		Q:          Q,
		QEmbedding: QEmbedding,
		A:          "",
		AEmbedding: nil,
		Title:      "",
		SubTitle:   "",
		DbName:     "QA-" + dbName,
	}
	if useCache {
		// 查询QA缓存
		cacheRep, err := qaStorage.GetQAStorage(config)
		if err != nil {
			return "", "", err
		}

		// 命中QA缓存
		if cacheRep != nil {
			qaStorage.A = cacheRep.Result[0].Payload.A
			qaStorage.Title = cacheRep.Result[0].Payload.File_name
			qaStorage.SubTitle = cacheRep.Result[0].Payload.Sub_title
			return qaStorage.A, qaStorage.Title, nil
		}
	}
	// 未命中QA缓存
	res, err := qdrant.SearchHttp(qaStorage.QEmbedding)
	if err != nil {
		log.Println("[QuestionHandler ERR] SearchHttp error\n", err.Error())
		return "", "", err
	}

	if len(res.Result) == 0 {
		log.Println("[QuestionHandler] 未查询到匹配段落（db为空情况，请录入文档）")
		return "未查询到匹配段落（db为空情况，请录入文档）", "", nil
	}
	if res.Result[0].Score <= 0.79 {
		log.Println("[QuestionHandler] 相似度:\n", res.Result[0].Score)
		log.Println("[QuestionHandler] 匹配段落:\n", res.Result[0].Payload.Text)
		return "不好意思呀，你的问题我回答不了", "", nil
	}

	qaStorage.Title = res.Result[0].Payload.File_name
	qaStorage.SubTitle = res.Result[0].Payload.Sub_title

	HRPrompt := Prompt.HRPrompt{
		Question:    qaStorage.Q,
		ContextText: res.Result[0].Payload.Text,
	}

	prompt, err := HRPrompt.BuildPrompt()

	if err != nil {
		log.Println("[QuestionHandler ERR] BuildPrompt error\n", err.Error())
		return "", "", err
	}

	model := openai.GPT3Dot5Turbo
	log.Printf("[QuestionHandler] Sending OpenAI api request...\nPrompt:%s\n", prompt)

	openAIResponse, tokens, err := server.CallOpenAI(clientToUse, prompt, model,
		"You are an internal knowledge base robot of an enterprise, answering questions based on the context I provided, always using Chinese.",
		512)

	if err != nil {
		log.Println("[QuestionHandler ERR] OpenAI answer questions request error\n", err.Error())
		return "", "", err
	}

	response := server.OpenAIResponse{Response: openAIResponse, Tokens: tokens}

	log.Println("[QuestionHandler] OpenAI response:\n", response)
	log.Println("[QuestionHandler] 相似度:\n", res.Result[0].Score)

	if useCache {
		// 持久化到mysql(暂时不用)
		qaStorage.A = response.Response
		//qaStorage.SaveToMysqlStorage()

		// 持久化到qdrant
		log.Println("[QuestionHandler] SaveToQdrantStorage:\n", qaStorage)
		qaStorage.SaveToQdrantStorage(config)
	}

	return response.Response, qaStorage.Title, nil
}

func ShouldInsertCache(config initialization.Config, qaStorage storage.QAStorage) {
	// 持久化到qdrant
	log.Println("[QuestionHandler] ShouldInsertCache:\n", qaStorage)
	qaStorage.SaveToQdrantStorage(config)
}

func LoadDOC(config initialization.Config, dbName string, dirName string) error {
	qdrant := server.NewQdrant(config.QdrantAddrGrpc, config.QdrantAddrHttp, dbName, uint64(1536))
	clientToUse := initialization.NewOpenAIClient(config)

	chunks, err := server.CreateChunk(dirName)
	if err != nil {
		errMsg := fmt.Sprintf("Error getting chunks: %v", err)
		log.Println("[Chunks ERR]", errMsg)
		return err
	}
	embeddings, err := server.GetEmbeddings(clientToUse, chunks, 2, openai.AdaEmbeddingV2)
	if err != nil {
		errMsg := fmt.Sprintf("Error getting embeddings: %v", err)
		log.Println("[Embeddings ERR]", errMsg)
		return err
	}
	fmt.Printf("Total chunks: %d\n", len(chunks))
	fmt.Printf("Total embeddings: %d\n", len(embeddings))
	fmt.Printf("Embeddings length: %d\n", len(embeddings[0]))

	err = qdrant.UpsertEmbeddingsToQdrant(embeddings, chunks)
	if err != nil {
		return err
	}
	return nil
}

func DeleteDB(config initialization.Config, dbName string) error {
	qdrant := server.NewQdrant(config.QdrantAddrGrpc, config.QdrantAddrHttp, dbName, uint64(1536))
	err := qdrant.DeleteCollection()
	if err != nil {
		return err
	}
	return nil
}
