package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type WMNotificationStorage struct {
	Con  *pgxpool.Pool
	RCon *redis.Client
}

func NewWMNotificationStorage(con *pgxpool.Pool, r *redis.Client) *WMNotificationStorage {
	return &WMNotificationStorage{
		Con:  con,
		RCon: r,
	}
}

type WMNotification struct {
	NotificationID   string `json:"notificationId"`
	NotifiedUserId   string `json:"notifiedUserId"`
	NotifierUserId   string `json:"notifierUserId"`
	NotificationType string `json:"notificationType"`
	MessageId        string `json:"messageId"`
	MatchId          string `json:"matchId"`
	AckStatus        string `json:"ackStatus"`
	TimeStamp        int64  `json:"timeStamp"`
}

type WMNotificationRequest struct {
	NotifiedUserId   string `json:"notifiedUserId"`
	NotifierUserId   string `json:"notifierUserId"`
	NotificationType string `json:"notificationType"`
	MessageId        string `json:"messageId"`
	MatchId          string `json:"matchId"`
	AckStatus        string `json:"ackStatus"`
	TimeStamp        int64  `json:"timeStamp"`
}

const (
	MessageNotification = "messageNotification"
	MatchNotification   = "matchNotification"
)


/// -------------------  Notification ------------------- ///
///
///		Take in notificaiton request, handle based on notification type
///


func (s *WMNotificationStorage) CreateNotification(notification WMNotificationRequest) (WMNotification, error) {
	id := uuid.New().String()
	var newNotification WMNotification

	err := s.Con.QueryRow(context.Background(), `
       Insert into wNotification (notification_id, notified_user_id, notifier_user_id, notification_type, message_id, match_id, ack_status, time_stamp)	      
	`, id, notification.NotifiedUserId, notification.NotifierUserId, notification.NotificationType, notification.MessageId, 
	notification.MatchId, notification.AckStatus, notification.TimeStamp).Scan(&newNotification.NotificationID, &newNotification.NotifiedUserId, &newNotification.NotifierUserId, 
		&newNotification.NotificationType, &newNotification.MessageId, &newNotification.MatchId, 
		&newNotification.AckStatus, &newNotification.TimeStamp)

	if err != nil {
		return WMNotification{}, err
	}

	// based on notificaiton type, send different apns notification type
	
	return newNotification, nil
	
}