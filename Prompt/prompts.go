package Prompt

import (
	"fmt"
	"github.com/pkoukk/tiktoken-go"
)

type HRPrompt struct {
    ContextText string
	Question string
}

func (HP *HRPrompt) BuildPrompt() (string, error){
	tokenLimit := 3750
	DEFAULT_TEXT_QA_PROMPT_TMPL := fmt.Sprintf("Context information is below. \n---------------------\n%s\n---------------------\nGiven the context information and not prior knowledge, answer the question: %s\n", HP.ContextText, HP.Question)
	tke, err := tiktoken.EncodingForModel("gpt-3.5-turbo")
	if err != nil {
		return "", fmt.Errorf("getEncoding: %v", err)
	}

	questionToken := tke.Encode(HP.Question, nil, nil)
	promptToken := tke.Encode(DEFAULT_TEXT_QA_PROMPT_TMPL, nil, nil)
	promptTokenCount := len(promptToken)
	currentTokenCount := len(questionToken)

	if promptTokenCount >= tokenLimit || currentTokenCount >= tokenLimit {
		return "", fmt.Errorf("token too long: %v", currentTokenCount)
	}
	return DEFAULT_TEXT_QA_PROMPT_TMPL, nil
}
