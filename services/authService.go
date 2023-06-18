package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jdrew153/lib"
	"github.com/redis/go-redis/v9"
)

type AuthService struct {
	Con            *pgxpool.Pool
	RCon           *redis.Client
	SessionService *SessionStorage
}

func NewAuthService(con *pgxpool.Pool, rcon *redis.Client, SessionService *SessionStorage) *AuthService {
	return &AuthService{
		Con:            con,
		RCon:           rcon,
		SessionService: SessionService,
	}
}

type AuthenticationModel struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthenticatedModelResponse struct {
	Id        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Image     string `json:"image"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`

	Interests []Interest `json:"interests"`

	Session Session `json:"session"`
}

func (s *AuthService) AuthenticateUser(auth AuthenticationModel) (AuthenticatedModelResponse, error) {
	var user NewUser

	err := s.Con.QueryRow(context.Background(), "select * from authuser where email = $1",
		auth.Email).Scan(&user.Id, &user.Username, &user.Password, &user.Email,
		&user.EmailVerified, &user.Image, &user.Latitude, &user.Longitude, &user.PhoneNumber,
		&user.TwoFactorEnabled, &user.Bio, &user.WingmanNickname)

	if err != nil {
		return AuthenticatedModelResponse{}, err
	}

	passCheck, err := lib.VerifyHash(user.Password, auth.Password)

	if err != nil {
		return AuthenticatedModelResponse{}, err
	}

	if !passCheck {
		return AuthenticatedModelResponse{}, errors.New("invalid password")
	}

	var interests []Interest

	rows, err := s.Con.Query(context.Background(), `
		select * from interests where user_id = $1	
	`, user.Id)

	if err != nil {
		fmt.Println("user has no interests set...")
	}

	for rows.Next() {
		var interest Interest

		err = rows.Scan(&interest.InterestId, &interest.Interest, &interest.UserId)

		if err != nil {
			fmt.Println("err...", err.Error())
		}

		interests = append(interests, interest)
	}

	session, err := s.SessionService.CreateSession(user.Id)

	if err != nil {
		return AuthenticatedModelResponse{}, err

	}

	return AuthenticatedModelResponse{
		Id:        user.Id,
		Username:  user.Username,
		Email:     user.Email,
		Image:     user.Image,
		Latitude:  user.Latitude,
		Longitude: user.Longitude,
		Interests: interests,
		Session:   session,
	}, nil
}
