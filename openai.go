package main

import (
	"context"
	"time"

	"github.com/mengbin92/openai"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func goChatWithChan(reqCache *WeChatCache, respChan chan string, errChan chan error) {
	req := &openai.ChatCompletionRequset{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.Message{
			{Role: openai.ChatMessageRoleUser, Content: reqCache.Content},
		},
		Temperature: 0.9,
	}
	log.Info(req)

	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		errChan <- errors.Wrap(err, "get chat response from openai error")
		return
	}
	go func() {
	LOOP:
		err = cache.Set(context.Background(), reqCache.Key(), []byte(resp.Choices[0].Message.Content), viper.GetDuration("redis.expire")*time.Second).Err()
		if err != nil {
			log.Debugf("Set data error: %s", err.Error())
			goto LOOP
		}
	}()

	respChan <- resp.Choices[0].Message.Content
}
