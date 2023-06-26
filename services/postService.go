package services

import (
	"context"
	"fmt"
	"github.com/jdrew153/lib"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostStorage struct {
	Con *pgxpool.Pool
}

type Post struct {
	PostId       string  `json:"postId"`
	UserId       string  `json:"userId"`
	ImageUrl     string  `json:"imageUrl"`
	TimeStamp    string  `json:"timeStamp"`
	Caption      string  `json:"caption"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	LocationName string  `json:"locationName"`
}

type UserPostModel struct {
	Id               string `json:"id"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	Email            string `json:"email"`
	EmailVerified    bool   `json:"emailVerified"`
	Image            string `json:"image"`
	Latitude         string `json:"latitude"`
	Longitude        string `json:"longitude"`
	PhoneNumber      string `json:"phoneNumber"`
	TwoFactorEnabled string `json:"twoFactorEnabled"`
	Bio              string `json:"bio"`
	WingmanNickname  string `json:"wingmanNickname"`

	Posts []Post `json:"posts"`
}

type NewPostRequest struct {
	UserId       string  `json:"userId"`
	ImageUrl     string  `json:"imageUrl"`
	Caption      string  `json:"caption"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	LocationName string  `json:"locationName"`
}

// ----> Basically the new post service function is the constructor for the post service class

func NewPostService(con *pgxpool.Pool) *PostStorage {
	return &PostStorage{
		Con: con,
	}
}

func (p *PostStorage) CreatePost(post NewPostRequest) (Post, error) {
	var returnPost Post
	postId := uuid.New().String()
	timeStamp := time.Now().Unix()

	err := p.Con.QueryRow(context.Background(), "insert into posts(post_id, user_id, image_url, time_stamp, caption, latitude, longitude, location_name) values($1, $2, $3, $4, $5, $6, $7, $8) returning *",
		postId, post.UserId, post.ImageUrl,
		timeStamp, post.Caption, post.Latitude, post.Longitude, post.LocationName).Scan(&returnPost.PostId, &returnPost.UserId, &returnPost.ImageUrl,
		&returnPost.TimeStamp, &returnPost.Caption, &returnPost.Latitude, &returnPost.Longitude, &returnPost.LocationName)

	if err != nil {
		fmt.Println(err)
		return Post{}, err
	}

	return returnPost, nil
}

func (p *PostStorage) GetPostsByUserId(userId string) ([]Post, error) {
	var posts []Post

	rows, err := p.Con.Query(context.Background(), "select * from posts where user_id = $1", userId)

	if err != nil {
		return []Post{}, err
	}

	for rows.Next() {
		var post Post

		err := rows.Scan(&post.PostId, &post.UserId, &post.ImageUrl, &post.TimeStamp, &post.Caption, &post.Latitude, &post.Longitude, &post.LocationName)

		if err != nil {
			return []Post{}, err
		}

		posts = append(posts, post)
	}

	return posts, nil
}

//
//  Get all posts then sort list by closest location to user who made request..
//

type PostLocationDistanceModel struct {
	Post     Post    `json:"post"`
	Distance float64 `json:"distance"`
}

type UserLocationRequestModel struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (p *PostStorage) GetPostsByLocation(userLocation UserLocationRequestModel) ([]PostLocationDistanceModel, error) {

	var posts []PostLocationDistanceModel

	rows, err := p.Con.Query(context.Background(), `
		select * from posts
	`)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var post Post

		err := rows.Scan(&post.PostId, &post.UserId, &post.ImageUrl, &post.TimeStamp,
			&post.Caption, &post.Latitude, &post.Longitude, &post.LocationName)

		if err != nil {
			return nil, err
		}

		if post.Latitude == 0 || post.Longitude == 0 || len(post.LocationName) == 0 {
			fmt.Println("breaking, lat, long, or location name are their default values...")

		} else {
			distance := lib.GetDistanceFromCoords(userLocation.Latitude, userLocation.Longitude, post.Latitude, post.Longitude)

			posts = append(posts, PostLocationDistanceModel{Post: post, Distance: distance})
		}

	}
	fmt.Println(posts)
	return posts, nil
}
