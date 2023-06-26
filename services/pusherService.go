package services

import "github.com/pusher/pusher-http-go/v5"

type PusherService struct {
	Client *pusher.Client
}

func NewPusherService(p *pusher.Client) *PusherService {
	return &PusherService{
		Client: p,
	}
}

func (s *PusherService) Trigger(channel string, event string, data interface{}) error {

	return s.Client.Trigger(channel, event, data)

}
