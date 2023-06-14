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
	M    *lib.Mailer
}

func NewUserService(
	con *pgxpool.Pool,
	rcon *redis.Client,
	kp *kafka.Writer,
	m *lib.Mailer) *UserStorage {
	return &UserStorage{
		Con:  con,
		RCon: rcon,
		KP:   kp,
		M:    m,
	}
}

type NewUser struct {
	Id               string `json:"id"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	Email            string `json:"email"`
	EmailVerified    bool   `json:"emailVerified"`
	Image            string `json:"image"`
	Latitude         string `json:"latitude"`
	Longitude        string `json:"longitude"`
	PhoneNumber      string `json:"phoneNumber"`
	TwoFactorEnabled string   `json:"twoFactorEnabled"`
	Bio              string `json:"bio"`
	WingmanNickname  string `json:"wingmanNickname"`
}

type UserSessionResponse struct {
	Id          string  `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	Image       string  `json:"image"`
	Latitude    string  `json:"latitude"`
	Longitude   string  `json:"longitude"`
	PhoneNumber string  `json:"phoneNumber"`
	Session     Session `json:"session"`
}

func (s *UserStorage) CacheUser(data NewUser) (NewUser, error) {
	var user NewUser

	id := uuid.New().String()
	hashedPass, err := lib.Hash(data.Password)

	if err != nil {
		fmt.Println("Error hashing password")
		return NewUser{}, err
	}

	/////////  TODO - Add region to user schema, based
	////////  on region, write user to another specific table** for that region
	////////  so that feeds can be generated based on region

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

func (s *UserStorage) CreateNewUser(data NewUser) (UserSessionResponse, error) {
	var user UserSessionResponse

	hashedPass, err := lib.Hash(data.Password)
	fmt.Println("hashedPass", hashedPass)
	id := uuid.New().String()

	if err != nil {
		fmt.Println("Error hashing password")
		return UserSessionResponse{}, err
	}

	err = s.Con.QueryRow(context.Background(), "INSERT INTO authuser (id, username, password, email, email_verified, image, latitude, longitude, two_factor_enabled, bio, wingman_nickname) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id, username, email, image, latitude, longitude", id, data.Username, hashedPass, data.Email, data.EmailVerified, data.Image, data.Latitude, data.Longitude, data.TwoFactorEnabled, data.Bio, data.WingmanNickname).Scan(&user.Id, &user.Username, &user.Email,
		&user.Image, &user.Latitude, &user.Longitude)

	if err != nil {
		return UserSessionResponse{}, err
	}

	go func() {
		s.M.SendMail(user.Email, "Welcome to Wingman!", user.Username)
	}()

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

func (s *UserStorage) GetAllUsers(userId string) ([]UserPostModel, error) {

	/// do not include the user who made the request in the results

	results := s.retrieveCachedPosts("test")

	if len(results) == 0 {
		fmt.Println("fetching from db...")
		users := make(map[string]*UserPostModel)

		rows, err := s.Con.Query(context.Background(), "select * from authuser a inner join posts p on a.id = p.user_id where a.id != $1", userId)

		if err != nil {
			return []UserPostModel{}, err
		}

		for rows.Next() {
			var user NewUser
			var post Post

			err = rows.Scan(&user.Id, &user.Username, &user.Password, &user.Email,
				&user.EmailVerified, &user.Image, &user.Latitude, &user.Longitude, &user.PhoneNumber,
				&user.TwoFactorEnabled, &user.Bio, &user.WingmanNickname, &post.PostId, &post.UserId,
				&post.ImageUrl, &post.TimeStamp, &post.Caption)

			if err != nil {
				fmt.Println(err.Error())
				return []UserPostModel{}, err
			}

			if existingUser, ok := users[user.Email]; ok {
				existingUser.Posts = append(existingUser.Posts, post)
			} else {
				users[user.Email] = &UserPostModel{
					Id:               user.Id,
					Username:         user.Username,
					Password:         user.Password,
					Email:            user.Email,
					EmailVerified:    user.EmailVerified,
					Image:            user.Image,
					Latitude:         user.Latitude,
					Longitude:        user.Longitude,
					PhoneNumber:      user.PhoneNumber,
					TwoFactorEnabled: user.TwoFactorEnabled,
					Bio:              user.Bio,
					WingmanNickname:  user.WingmanNickname,
					Posts:            []Post{post},
				}
			}
		}

		// Convert the map values to a slice
		usersSlice := make([]UserPostModel, 0, len(users))
		for _, user := range users {
			usersSlice = append(usersSlice, *user)
		}

		d, err := json.Marshal(usersSlice)
		if err != nil {
			fmt.Println(err.Error())
		}

		s.RCon.Set(context.Background(), "test", string(d), 0)

		return usersSlice, nil
	}

	fmt.Println("fetching from cache...")
	return results, nil
}

func (s *UserStorage) retrieveCachedPosts(key string) []UserPostModel {
	var users []UserPostModel

	val, err := s.RCon.Get(context.Background(), key).Result()

	if err != nil {
		fmt.Println(err.Error())
		return []UserPostModel{}
	}

	err = json.Unmarshal([]byte(val), &users)

	if err != nil {
		fmt.Println(err.Error())
		return []UserPostModel{}
	}

	return users
}
