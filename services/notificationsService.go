package services

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
)

type NotificationStorage struct {
	RCon *redis.Client
	APNS *apns2.Client
}

func NewNotificationService(rcon *redis.Client, apns *apns2.Client) *NotificationStorage {
	return &NotificationStorage{
		RCon: rcon,
		APNS: apns,
	}
}

type NotificationRequest struct {
	UserId string `json:"userId"`
	DeviceId string `json:"deviceId"`
}

// APNS portion of notification handling...

func (s *NotificationStorage) CreateNotificationPair(pair NotificationRequest) (NotificationRequest, error) {
	fmt.Println("pair", pair)
    if (pair.UserId == "" || pair.DeviceId == "") {
		return NotificationRequest{}, fmt.Errorf("invalid request body")
	}

	
	err := s.RCon.Set(context.Background(), fmt.Sprintf("notifications-%s", string(pair.UserId)),
	string(pair.DeviceId), 0).Err()

	if err != nil {
		return NotificationRequest{}, err
	}

	val, err := s.RCon.Get(context.Background(), fmt.Sprintf("notifications-%s", string(pair.UserId))).Result()

	if err != nil {
		return NotificationRequest{}, err
	}

	response := NotificationRequest{
		UserId: pair.UserId,
		DeviceId: val,
	}
	fmt.Println("device key --->", val)

	return response, nil

}

func (s *NotificationStorage) RetriveDeviceTokenForUser(userId string) (string, error) {
	fmt.Println("userId", userId)
	val, err := s.RCon.Get(context.Background(), fmt.Sprintf("notifications-%s", string(userId))).Result()

	if err != nil {
		return "", err
	}

	return val, nil

}


func (s *NotificationStorage) CreateAPNSNotification(message string, userId string) (*apns2.Response, error) {
	deviceToken, err := s.RetriveDeviceTokenForUser(userId)

	if err != nil {
		return nil, err
	}

	payload := payload.NewPayload().Alert(message).Badge(0).Sound("default").InterruptionLevel("active")

	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       "lmn.Joshie",
		Payload:     payload,
	}

	res, err := s.APNS.Push(notification)

	if err != nil {
		return nil, err
	}

	return res, nil
}	