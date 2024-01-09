package server

import (
	"context"
	openai "github.com/sashabaranov/go-openai"
	"log"
	"time"
)

type OpenAIResponse struct {
	Response string `json:"response"`
	Tokens   int    `json:"tokens"`
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func CallOpenAI(client *openai.Client, prompt string, model string, instructions string, maxTokens int) (string, int, error) {
	// set request details
	temperature := float32(0.1)
	topP := float32(1.0)
	frequencyPenalty := float32(0.0)
	presencePenalty := float32(0.6)
	stop := []string{"Human:", "AI:"}

	var assistantMessage string
	var tokens int
	var err error
	assistantMessage, tokens, err = useChatCompletionAPI(client, prompt, model, instructions, temperature,
			maxTokens, topP, frequencyPenalty, presencePenalty, stop)

	return assistantMessage, tokens, err
}

func useChatCompletionAPI(client *openai.Client, prompt, modelParam string, instructions string, temperature float32, maxTokens int, topP float32, frequencyPenalty, presencePenalty float32, stop []string) (string, int, error) {
	messages := []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: instructions,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:            modelParam,
			Messages:         messages,
			Temperature:      temperature,
			MaxTokens:        maxTokens,
			TopP:             topP,
			FrequencyPenalty: frequencyPenalty,
			PresencePenalty:  presencePenalty,
			Stop:             stop,
		},
	)

	if err != nil {
		return "", 0, err
	}

	return resp.Choices[0].Message.Content, resp.Usage.TotalTokens, nil
}

func callEmbeddingAPIWithRetry(client *openai.Client, texts []string, embedModel openai.EmbeddingModel,
	maxRetries int) (*openai.EmbeddingResponse, error) {
	var err error
	var res openai.EmbeddingResponse
	log.Println("长度: ", len(texts))
	log.Println("即将embeddings的内容是: ")
	log.Println(texts)
	for i := 0; i < maxRetries; i++ {
		res, err = client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
			Input: texts,
			Model: embedModel,
		})

		if err == nil {
			return &res, nil
		}

		time.Sleep(5 * time.Second)
	}

	return nil, err
}

func GetEmbedding(client *openai.Client, text string, embedModel openai.EmbeddingModel) ([]float32, error) {
	res, err := callEmbeddingAPIWithRetry(client, []string{text}, embedModel, 3)
	if err != nil {
		return nil, err
	}

	return res.Data[0].Embedding, nil
}

func GetEmbeddings(client *openai.Client, chunks []Chunk, batchSize int,
	embedModel openai.EmbeddingModel) ([][]float32, error) {
	embeddings := make([][]float32, 0, len(chunks))

	for i := 0; i < len(chunks); i += batchSize {
		iEnd := min(len(chunks), i+batchSize)

		texts := make([]string, 0, iEnd-i)
		for _, chunk := range chunks[i:iEnd] {
			texts = append(texts, chunk.Text)
		}

		log.Println("[getEmbeddings] Feeding texts to Openai to get embedding...\n", texts)

		res, err := callEmbeddingAPIWithRetry(client, texts, embedModel, 3)
		if err != nil {
			return nil, err
		}

		embeds := make([][]float32, len(res.Data))
		for i, record := range res.Data {
			embeds[i] = record.Embedding
		}

		embeddings = append(embeddings, embeds...)
	}

	return embeddings, nil
}
