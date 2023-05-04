package initialization

import (
	"fmt"
	"github.com/sashabaranov/go-openai"
	"net/http"
	"net/url"
)

func NewOpenAIClient(config Config) *openai.Client {
	if config.HttpProxy == "" {
		clientToUse := openai.NewClient(config.OpenaiApiKeys)
		return clientToUse
	}
	clientconfig := openai.DefaultConfig(config.OpenaiApiKeys)
	proxyUrl, err := url.Parse(config.HttpProxy)
	if err != nil {
		fmt.Printf("NewOpenAIClient Error:%v\n", err)
		return nil
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}
	clientconfig.HTTPClient = &http.Client{
		Transport: transport,
	}
	clientToUseWithProxy := openai.NewClientWithConfig(clientconfig)
	return clientToUseWithProxy
}
