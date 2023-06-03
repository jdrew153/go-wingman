package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type SessionStorage struct {
	Con *pgxpool.Pool
	R *redis.Client
}

func NewSessionService(con *pgxpool.Pool, r *redis.Client) *SessionStorage {
	return &SessionStorage{
		Con: con,
		R: r,
	}
}

type Session struct {
	Id string `json:"id"`
	UserId string `json:"userId"`
	Expires string `json:"expires"`
	Dead bool `json:"dead"`
}

func (s *SessionStorage) CreateSession(userId string) (Session, error) {
	var newSession Session
	id := uuid.New().String()

	expires :=  time.Now().Unix() + 60 * 60 * 24 * 7

	err := s.Con.QueryRow(context.Background(), "insert into authsession(id, user_id, expires, dead) values($1, $2, $3, $4) returning *", id, userId, expires, false).Scan(&newSession.Id, &newSession.UserId, &newSession.Expires, &newSession.Dead)

	if err != nil {
		return Session{}, err
	}

	return newSession, nil
}
	