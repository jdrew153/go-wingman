package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

type WMMessagingService struct {
	Con  *pgxpool.Pool
	Rcon *redis.Client
}

type ConversationModel struct {
	ConversationId string `json:"conversationId"`
	UserIdA        string `json:"userIdA"`
	UserIdB        string `json:"userIdB"`

	// lets use last message as the id of the message, so we can get the actual message
	LastMessage string `json:"lastMessage"`
}

type WMMessageRequest struct {
	ConversationId string `json:"conversationId"`
	SenderId       string `json:"senderId"`
	ReceiverId     string `json:"receiverId"`
	Content        string `json:"content"`
	ContentUrl     string `json:"contentUrl"`
}

type WMMessageResponse struct {
	MessageId      string `json:"messageId"`
	ConversationId string `json:"conversationId"`
	SenderId       string `json:"senderId"`
	ReceiverId     string `json:"receiverId"`
	Content        string `json:"content"`
	ContentUrl     string `json:"contentUrl"`
	TimeStamp      int64  `json:"timeStamp"`
}

func NewWMMessagingServices(p *pgxpool.Pool, r *redis.Client) *WMMessagingService {
	return &WMMessagingService{
		Con:  p,
		Rcon: r,
	}
}

//
//  The flow for messaging needs to go like this:
//  1. Create a conversation if there isn't one when a user goes to send the message
//  2. Find the conversation between users where users are either a or b
//  3. Find a list of all the messages and hydrate a user model on the conversation or the message, whichever is easier but is not slow
//

// SearchAndUpsertANewConversation
// This function will find a conversation between a and b or b and a and return the conversation id if found.
// Otherwise, a new conversation will be created.
// This function is meant to be used as the first filter when a user sends a message.
func (s *WMMessagingService) SearchAndUpsertANewConversation(userIdA string, userIdB string) (string, error) {
	query := `
		select * from conversations where exists(
		    select * from conversations where (user_id_a = $1 and user_id_b = $2) or (user_id_b = $1 and user_id_a = $2)
		)
	`

	var conversation ConversationModel

	err := s.Con.QueryRow(context.Background(), query, userIdA, userIdB).Scan(
		&conversation.ConversationId,
		&conversation.UserIdA,
		&conversation.UserIdB,
		&conversation.LastMessage,
	)
	//
	// I think if we get an error here that means no rows were found, so now we start the upsert process.
	//

	if err != nil {
		fmt.Println("failed finding a conversation, but carrying on..")
		fmt.Println(err.Error())
		id := uuid.New().String()
		var newConversationId string

		err := s.Con.QueryRow(context.Background(), `
			insert into conversations (conversation_id, user_id_a, user_id_b, last_message)
			values ($1, $2, $3, '') returning conversation_id
		`, id, userIdA, userIdB).Scan(&newConversationId)

		if err != nil {
			fmt.Println("failed trying to create new conversation entry")
			return "", err
		}

		return newConversationId, nil
	} else {
		return conversation.ConversationId, nil
	}
}

func (s *WMMessagingService) AddNewMessageToConversation(message WMMessageRequest) (WMMessageResponse, error) {

	var messageResponse WMMessageResponse

	id := uuid.New().String()
	timeStamp := time.Now().UnixMilli()

	err := s.Con.QueryRow(context.Background(), `
		insert into messages values ($1, $2, $3, $4, $5, $6, $7) returning *
	`, id, message.ConversationId, message.SenderId, message.ReceiverId,
		message.Content, message.ContentUrl, timeStamp).Scan(
		&messageResponse.MessageId, &messageResponse.ConversationId,
		&messageResponse.SenderId, &messageResponse.ReceiverId, &messageResponse.Content,
		&messageResponse.ContentUrl, &messageResponse.TimeStamp,
	)

	if err != nil {
		fmt.Println("failed to create the message ")
		return WMMessageResponse{}, err
	}

	var conversation ConversationModel

	err = s.Con.QueryRow(context.Background(), `
		update conversations set last_message = $2 where conversation_id = $1
	`, messageResponse.ConversationId, messageResponse.MessageId).Scan(&conversation.ConversationId, &conversation.UserIdA, &conversation.UserIdB,
		&conversation.LastMessage)

	if err != nil {
		println(err.Error())
	}

	fmt.Println(conversation)

	return messageResponse, nil
}

//
//  Goal is to return an array of conversations, and attached to the conversation, the array of messages between the users
//  The user making the request can either be userA or userB in the conversation table.
//

type HydratedConversationModel struct {
	ConversationId string  `json:"conversationId"`
	UserIdA        string  `json:"userIdA"`
	UserIdB        string  `json:"userIdB"`
	UserA          NewUser `json:"userA"`
	UserB          NewUser `json:"userB"`

	// lets use last message as the id of the message, so we can get the actual message
	LastMessage string `json:"lastMessage"`

	Messages []WMMessageResponse `json:"messages"`
}

func (s *WMMessagingService) FetchConversationsForUser(userId string) ([]HydratedConversationModel, error) {

	var conversations []HydratedConversationModel

	rows, err := s.Con.Query(context.Background(), `
		select * from conversations where user_id_a = $1 or user_id_b = $1
	`, userId)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var conversation ConversationModel

		err = rows.Scan(&conversation.ConversationId, &conversation.UserIdA, &conversation.UserIdB, &conversation.LastMessage)

		if err != nil {
			return nil, err
		}

		// fetching both users with two go funcs??

		userAChan := make(chan NewUser, 1)
		userBChan := make(chan NewUser, 1)

		wg := sync.WaitGroup{}

		wg.Add(1)
		go func() {
			defer wg.Done()
			var user NewUser

			err := s.Con.QueryRow(context.Background(), `
				select * from authuser where id = $1
			`, conversation.UserIdA).Scan(&user.Id, &user.Username, &user.Password, &user.Email,
				&user.EmailVerified, &user.Image, &user.Latitude, &user.Longitude, &user.PhoneNumber,
				&user.TwoFactorEnabled, &user.Bio, &user.WingmanNickname)

			if err != nil {
				fmt.Println(err.Error())
			}

			userAChan <- user
			close(userAChan)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			var user NewUser

			err := s.Con.QueryRow(context.Background(), `
				select * from authuser where id = $1
			`, conversation.UserIdB).Scan(&user.Id, &user.Username, &user.Password, &user.Email,
				&user.EmailVerified, &user.Image, &user.Latitude, &user.Longitude, &user.PhoneNumber,
				&user.TwoFactorEnabled, &user.Bio, &user.WingmanNickname)

			if err != nil {
				fmt.Println(err.Error())
			}

			userBChan <- user
			close(userBChan)
		}()

		messageRows, err := s.Con.Query(context.Background(), `
			select * from messages where conversation_id = $1
        `, conversation.ConversationId)

		if err != nil {
			return nil, err
		}
		var messages []WMMessageResponse

		for messageRows.Next() {
			var message WMMessageResponse
			err = messageRows.Scan(&message.MessageId, &message.ConversationId, &message.SenderId, &message.ReceiverId, &message.Content,
				&message.ContentUrl, &message.TimeStamp)

			if err != nil {
				return nil, err
			}

			messages = append(messages, message)
		}

		wg.Wait()

		userA := <-userAChan
		userB := <-userBChan

		conversations = append(conversations, HydratedConversationModel{
			ConversationId: conversation.ConversationId,
			UserIdA:        conversation.UserIdA,
			UserIdB:        conversation.UserIdB,
			UserA:          userA,
			UserB:          userB,
			LastMessage:    conversation.LastMessage,
			Messages:       messages,
		})
	}

	return conversations, nil
}
