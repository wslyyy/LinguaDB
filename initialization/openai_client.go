package initialization

import (
	"crypto/tls"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/net/proxy"
	"log"
	"net/http"
	"time"
)

func NewOpenAIClient(config Config) *openai.Client {
	if config.HttpProxy == "" {
		clientToUse := openai.NewClient(config.OpenaiApiKeys)
		return clientToUse
	}
	clientconfig := openai.DefaultConfig(config.OpenaiApiKeys)
	/*
	proxyUrl, err := url.Parse(config.HttpProxy)
	if err != nil {
		fmt.Printf("NewOpenAIClient Error:%v\n", err)
		return nil
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}

	 */
	dialer, err := proxy.SOCKS5("tcp", "50.18.207.90:41897", &proxy.Auth{
		User:     "root",
		Password: "root",
	}, proxy.Direct)
	if err != nil {
		log.Printf("socks5 init err: %s", err.Error())
		return nil
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Dial:            dialer.Dial,
	}

	clientconfig.HTTPClient = &http.Client{
		Transport: transport,
		Timeout: 110 * time.Second,
	}
	clientToUseWithProxy := openai.NewClientWithConfig(clientconfig)
	return clientToUseWithProxy
}
