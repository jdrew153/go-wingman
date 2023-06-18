package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"time"
)

type MessageService struct {
	RCon *redis.Client
}

func NewMessageService(r *redis.Client) *MessageService {
	return &MessageService{
		RCon: r,
	}
}

type MessageRequestModel struct {
	UserIdA         string `json:"userIdA"`
	UserIdB         string `json:"userIdB"`
	UserIdBImageUrl string `json:"userIdBImageUrl"`
	UserIdBUsername string `json:"userIdBUsername"`
	Content         string `json:"content"`
}

type MessageResponseModel struct {
	MessageId       string `json:"messageId"`
	UserIdA         string `json:"userIdA"`
	UserIdB         string `json:"userIdB"`
	UserIdBImageUrl string `json:"userIdBImageUrl"`
	UserIdBUsername string `json:"userIdBUsername"`
	Content         string `json:"content"`
	Timestamp       int64  `json:"timestamp"`
}

//
//  Using redis, need one key that holds conversations between users and others that hold the messages
//  between userA and userB. This key will hold a list of userIds that the user has conversations with.
//

func (s *MessageService) CreateConversationKeyForUser(userIdA string, userIdB string) error {
	key := fmt.Sprintf("conversations-%s", userIdA)

	// need to check if userIdB is already in list

	exists, err := s.RCon.Exists(context.Background(), key).Result()

	if err != nil {
		return err
	}

	found := false

	if exists == 1 {
		results, err := s.RCon.LRange(context.Background(), key, 0, -1).Result()

		if err != nil {
			fmt.Println("error at exists check..")
			return err
		}

		for _, id := range results {
			if userIdB == id {
				found = true
			}
		}

	}

	if found {
		return nil
	} else {
		err = s.RCon.RPush(context.Background(), key, userIdB).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

//
//  This function creates key between users to hold messages between userIdA and userIdB. If the key exists,
//  update the message list with the new message...
//

func (s *MessageService) SendNewMessage(message MessageRequestModel) (MessageResponseModel, error) {
	key := fmt.Sprintf("messages-%s-%s", message.UserIdA, message.UserIdB)

	// Create the message
	response := MessageResponseModel{
		MessageId:       uuid.NewString(),
		UserIdA:         message.UserIdA,
		UserIdB:         message.UserIdB,
		UserIdBImageUrl: message.UserIdBImageUrl,
		UserIdBUsername: message.UserIdBUsername,
		Content:         message.Content,
		Timestamp:       time.Now().UnixMilli(),
	}

	// Marshal the message to JSON
	d, err := json.Marshal(response)
	if err != nil {
		return MessageResponseModel{}, err
	}

	err = s.RCon.RPush(context.Background(), key, d).Err()
	if err != nil {
		return MessageResponseModel{}, err
	}

	fmt.Println(response)

	return response, nil
}

//
//  Use this function to find a list of the users that the user in question has,
//  use the list of user ids to find the messages using the key format messages-userIdA-userIdB.
//  If there are no messages, return an empty list, not nil
//

func (s *MessageService) GetMessagesForUser(userId string) (map[string][]MessageResponseModel, error) {
	// Conversation key
	key := fmt.Sprintf("conversations-%s", userId)

	exists, err := s.RCon.Exists(context.Background(), key).Result()

	if err != nil {
		return nil, err
	}

	if exists == 0 {
		// Not necessarily an error state. The user has no conversations at this point.
		return nil, nil
	} else {
		// At this point, the user does have conversations. Need to get the list

		results, err := s.RCon.LRange(context.Background(), key, 0, -1).Result()

		if err != nil {
			return nil, err
		}

		returnMap := make(map[string][]MessageResponseModel)

		for _, result := range results {

			messagesKey := fmt.Sprintf("messages-%s-%s", userId, result)

			messages, err := s.RCon.LRange(context.Background(), messagesKey, 0, -1).Result()

			if err != nil {
				return nil, err
			}

			var messageResponses []MessageResponseModel

			for _, msg := range messages {
				var messageResponse MessageResponseModel
				err := json.Unmarshal([]byte(msg), &messageResponse)

				if err != nil {
					return nil, err
				}

				messageResponses = append(messageResponses, messageResponse)
			}

			returnMap[result] = messageResponses
		}

		return returnMap, nil
	}
}
