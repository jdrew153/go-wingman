package services

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type WingmanService struct {
	G *openai.Client
}

func NewWingmanService(g *openai.Client) *WingmanService {
	return &WingmanService{
		G: g,
	}
}

func (s *WingmanService) CreateWingmanResponse(messages []openai.ChatCompletionMessage) openai.ChatCompletionMessage {

	systemMessages := []openai.ChatCompletionMessage{{

		Role:    openai.ChatMessageRoleSystem,
		Content: "You are a helpful dating app assistant who can provide pickup lines and provide advice to a user when asked.",
	},
	}

	systemMessages = append(systemMessages, messages...)

	resp, err := s.G.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: systemMessages,
		},
	)

	if err != nil {
		fmt.Println("Error", err)
	}

	return openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: resp.Choices[0].Message.Content,
	}
}
