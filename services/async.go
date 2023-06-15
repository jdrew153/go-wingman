package services

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"sync"
	"time"
)

type AsyncService struct {
	Con *pgxpool.Pool
}

func NewAsyncService(con *pgxpool.Pool) *AsyncService {
	return &AsyncService{
		con,
	}
}

//
//  Continuation from sync services by getting lists of userIds that we want to filter out from the main
//  user pool we request from.
//

func (s *AsyncService) UserIdsWhoUserHasSentMatchRequestTo(ctx context.Context, userId string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)

	defer cancel()

	// Channel for error handling
	errorChan := make(chan error, 1)

	// Channel for list of user ids to be returned
	userIdsChan := make(chan map[string]string, 1)

	go func() {
		returnableUserIds := make(map[string]string)
		rows, err := s.Con.Query(context.Background(), `
		select user_id_b from matches where user_id_a = $1 
	`, userId)

		if err != nil {
			errorChan <- err

			return
		}

		for rows.Next() {
			var userId string
			err = rows.Scan(&userId)

			if err != nil {
				errorChan <- err

				return
			}

			returnableUserIds[userId] = userId
		}
		fmt.Println("obtained data from db...")
		userIdsChan <- returnableUserIds

	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("timed out")
	case err := <-errorChan:
		fmt.Println("function returned an error")
		return nil, err
	case userIds := <-userIdsChan:
		fmt.Println("function is returning a good result")
		return userIds, nil
	}

}

//
//  Filter out settled match requests.., Where the user is either a or b and the match status = 'mutual'
//

func (s *AsyncService) FilterSettledMatchRequestsAsync(ctx context.Context, userId string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*600)

	defer cancel()
	errorChan := make(chan error, 1)
	settledMatchRequestChan := make(chan map[string]string, 1)

	go func() {
		settledMatchRequestsIveSent := make(map[string]string)
		rows, err := s.Con.Query(context.Background(), `
			select user_id_b as user_id from matches where user_id_a = $1 and match_status = 'mutual'
			union
			select user_id_a as user_id from matches where user_id_b = $1 and match_status = 'mutual'
	
	`, userId)

		if err != nil {
			errorChan <- err

			return
		}

		for rows.Next() {
			var userId string

			err = rows.Scan(&userId)
			if err != nil {
				errorChan <- err

				return
			}
			fmt.Println(userId)
			settledMatchRequestsIveSent[userId] = userId
		}

		settledMatchRequestChan <- settledMatchRequestsIveSent

	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("timed out")
	case err := <-errorChan:
		return nil, err
	case settledMatchRequests := <-settledMatchRequestChan:
		return settledMatchRequests, nil
	}

}

func (s *AsyncService) CreateUserContextAsync(ctx context.Context, user NewUser) (UserContext, error) {

	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*400)

	wg := sync.WaitGroup{}

	defer cancel()

	errorChan := make(chan error, 1)

	postsChan := make(chan []Post, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var posts []Post
		rows, err := s.Con.Query(ctx, `
			select * from posts where user_id = $1
		`, user.Id)

		if err != nil {
			errorChan <- err
		}

		for rows.Next() {
			var post Post
			err = rows.Scan(&post.PostId,
				&post.UserId, &post.ImageUrl,
				&post.TimeStamp, &post.Caption)

			if err != nil {
				errorChan <- err
			}
			posts = append(posts, post)
		}

		postsChan <- posts
		close(postsChan)
	}()

	interestsChan := make(chan []Interest, 1)

	wg.Add(1)
	go func() {
		wg.Done()
		var interests []Interest
		rows, err := s.Con.Query(ctx, `
			select * from interests where user_id = $1
		`, user.Id)

		if err != nil {
			errorChan <- err
		}

		for rows.Next() {
			var interest Interest

			err = rows.Scan(&interest.InterestId, &interest.UserId, &interest.Interest)

			if err != nil {
				errorChan <- err
			}

			interests = append(interests, interest)
		}

		interestsChan <- interests
		close(interestsChan)
	}()

	wg.Wait()

	for {
		select {
		case <-ctx.Done():
			return UserContext{}, fmt.Errorf("timed out")
		case err := <-errorChan:
			return UserContext{}, err
		case posts := <-postsChan:
			interests := <-interestsChan

			return UserContext{
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
				IsPotentialMatch: false,
				Posts:            posts,
				Interests:        interests,
			}, nil
		}
	}
}
