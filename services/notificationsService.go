package services

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type NotificationStorage struct {
	RCon *redis.Client
}

func NewNotificationService(rcon *redis.Client) *NotificationStorage {
	return &NotificationStorage{
		RCon: rcon,
	}
}

type NotificationRequest struct {
	UserId string `json:"userId"`
	DeviceId string `json:"deviceId"`
}

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

	val, err := s.RCon.Get(context.Background(), fmt.Sprintf("notifications-%s", string(userId))).Result()

	if err != nil {
		return "", err
	}

	return val, nil

}