package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jdrew153/lib"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type UserStorage struct {
	Con  *pgxpool.Pool
	RCon *redis.Client
	KP   *kafka.Writer
}

func NewUserService(con *pgxpool.Pool, rcon *redis.Client, kp *kafka.Writer) *UserStorage {
	return &UserStorage{
		Con:  con,
		RCon: rcon,
		KP:   kp,
	}
}

type NewUser struct {
	Id            string `json:"id"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
	Image         string `json:"image"`
	Latitude      string `json:"latitude"`
	Longitude     string `json:"longitude"`
}

func (s *UserStorage) CacheUser(data NewUser) (NewUser, error) {
	var user NewUser

	id := uuid.New().String()
	hashedPass, err := lib.Hash(data.Password)

	if err != nil {
		fmt.Println("Error hashing password")
		return NewUser{}, err
	}

	rUser := &NewUser{
		Id:            id,
		Username:      data.Username,
		Password:      hashedPass,
		Email:         data.Email,
		EmailVerified: data.EmailVerified,
		Image:         data.Image,
		Latitude:      data.Latitude,
		Longitude:     data.Longitude,
	}

	b, err := json.Marshal(rUser)

	println("b", string(b))
	if err != nil {
		return NewUser{}, err
	}

	g := strings.Trim(string(b), "\"")

	err = s.RCon.Set(context.Background(), id, g, 0).Err()

	if err != nil {
		println("err setting data in redis")
		println(err.Error())
		return NewUser{}, err
	}

	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("%s-%s", id, data.Email)),
		Value: b,
	}

	go func() {
		err = s.KP.WriteMessages(context.Background(), msg)

		if err != nil {
			println("err writing message to kafka", err.Error())
		}
	}()

	val, err := s.RCon.Get(context.Background(), id).Result()

	if err != nil {
		println("err getting data in redis")
		return NewUser{}, err
	}

	err = json.Unmarshal([]byte(val), &user)

	if err != nil {
		println("err unmarshalling data from redis")
		return NewUser{}, err
	}

	return user, nil

}

func (s *UserStorage) CreateNewUser(data NewUser) (NewUser, error) {
	var user NewUser

	hashedPass, err := lib.Hash(data.Password)
	fmt.Println("hashedPass", hashedPass)
	id := uuid.New().String()

	if err != nil {
		fmt.Println("Error hashing password")
		return NewUser{}, err
	}

	err = s.Con.QueryRow(context.Background(), "INSERT INTO authuser (id, username, password, email, email_verified, image, latitude, longitude) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, username, password, email, email_verified, image, latitude, longitude", id, data.Username, hashedPass, data.Email, data.EmailVerified, data.Image, data.Latitude, data.Longitude).Scan(&user.Id, &user.Username, &user.Password, &user.Email, &user.EmailVerified, &user.Image, &user.Latitude, &user.Longitude)

	if err != nil {
		return NewUser{}, err
	}

	go func() {

		json, err := json.Marshal(data)

		if err != nil {
			println("err marshalling data for kafka")
			return
		}

		msg := kafka.Message{
			Key:   []byte(fmt.Sprintf("%s-%s", id, data.Email)),
			Value: json,
		}

		err = s.KP.WriteMessages(context.Background(), msg)

		if err != nil {
			println("err writing message to kafka", err.Error())
		}
	}()

	return user, nil
}
