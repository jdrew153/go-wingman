package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type MatchStorage struct {
	Con  *pgxpool.Pool
	RCon *redis.Client
}

func NewMatchService(con *pgxpool.Pool, rcon *redis.Client) *MatchStorage {
	return &MatchStorage{
		Con:  con,
		RCon: rcon,
	}
}

type MatchRequest struct {
	UserIdA string `json:"userIdA"`
	UserIdB string `json:"userIdB"`
}

////
////   Note: UserIdA is the user who initiated the match request.
////   UserIdB is the user who received the match request.
////
////

type MatchResponse struct {
	MatchId     string `json:"matchId"`
	UserIdA     string `json:"userIdA"`
	UserIdB     string `json:"userIdB"`
	MatchStatus string `json:"matchStatus"`
	TimeStamp   int64  `json:"timeStamp"`
}

//// Use these values for the match status enum.....

const (
	Pending  = "pending"
	Rejected = "rejected"
	Mutual   = "mutual"
)

func (s *MatchStorage) CreateNewMatch(request MatchRequest) (MatchResponse, error) {

	var match MatchResponse

	fmt.Println("Creating new match: ", request)

	id := uuid.New().String()
	timeStamp := time.Now().Unix()
	matchStatus := Pending

	err := s.Con.QueryRow(context.Background(), "INSERT INTO matches (match_id, user_id_a, user_id_b, match_status, time_stamp) VALUES ($1, $2, $3, $4, $5) RETURNING *",
		id, request.UserIdA, request.UserIdB, matchStatus, timeStamp).Scan(&match.MatchId, &match.UserIdA,
		&match.UserIdB, &match.MatchStatus, &match.TimeStamp)

	if err != nil {
		fmt.Println("Error inserting match: ", err.Error())
		return match, err
	}

	return match, nil
}

func (s *MatchStorage) FindAndUpdateMatchStatus(userIdA string, userIdB string, updatedStatus string) (MatchResponse, error) {
	var match MatchResponse

	err := s.Con.QueryRow(
		context.Background(),
		"UPDATE matches SET match_status = $1 WHERE user_id_a = $2 AND user_id_b = $3 RETURNING *", updatedStatus, userIdA, userIdB,
	).Scan(&match.MatchId, &match.UserIdA, &match.UserIdB, &match.MatchStatus, &match.TimeStamp)

	/// TODO - Send off notification to userIdA based on updated status value

	if err != nil {
		fmt.Println("Error updating match status: ", err.Error())
		// try other combo??? or eventually place match ids in the stored user defaults / context on the client
		err = s.Con.QueryRow(context.Background(), "UPDATE matches SET match_status = $1 WHERE user_id_a = $2 AND user_id_b = $3 RETURNING *", updatedStatus, userIdB, userIdA).Scan(&match.MatchId, &match.UserIdA, &match.UserIdB, &match.MatchStatus, &match.TimeStamp)

		if err != nil {
			fmt.Println("Error updating match status: ", err.Error())
			return match, err
		}

		return match, nil
	}

	return match, nil
}
