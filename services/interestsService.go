package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Interest struct {
	InterestId string `json:"interestId"`
	Interest   string `json:"interest"`
	UserId     string `json:"userId"`
}


type InterestStorage struct {
	Con *pgxpool.Pool
}

func InterestService(con *pgxpool.Pool) *InterestStorage {
	return &InterestStorage{
		Con: con,
	}
}

func (i *InterestStorage) writeNewInterest(interest Interest) (Interest, error) {

	var returnInterest Interest
	id := uuid.New().String()
	 err := i.Con.QueryRow(context.Background(), "insert into interests(interest_id, interest, user_id) values($1, $2, $3)", 
	 id, interest.Interest, interest.UserId).Scan(&returnInterest.InterestId, 
		&returnInterest.Interest, &returnInterest.UserId)

	if err != nil {
		return Interest{}, err
	}

	return returnInterest, nil
}

func (s *InterestStorage) BatchCreateInterests(interests []string, userId string) ([]Interest) {

	var returnInterests []Interest

	for i := 0; i < len(interests); i++ {
		go func(x int)  {
			newInterest := Interest{
				Interest: interests[x],
				UserId: userId,
			}
			interest, err := s.writeNewInterest(newInterest)

			if err != nil {
				println(err.Error())
				return 
			}
			returnInterests = append(returnInterests, interest)
		}(i)
	}

	return returnInterests
}