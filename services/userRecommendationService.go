package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jdrew153/lib"
)

type UserRecommendationService struct {
	Con *pgx.Conn
}

func NewUserRecommendationService(con *pgx.Conn) *UserRecommendationService {
	return &UserRecommendationService{Con: con}
}

type UserRecommendationModel struct {
	RecommendationId string   `json:"recommendation_id"`
	UserId       string   `json:"user_id"`
	UserIdRecommended string `json:"user_id_recommended"`
	InterestSimilarity float64 `json:"interest_similarity"`
	Distance float64 `json:"distance"`
	MatchSimilarity float64 `json:"match_similarity"`
	LikeSimilarity float64 `json:"like_similarity"`
	IsPotentialMatch bool `json:"is_potential_match"`     
}


func (s *UserRecommendationService) CalcInterestSimilarityBtwnUsers(intersestArrOne []string, interestArrTwo []string) float64 {
	strOne := strings.Join(intersestArrOne, ",")
	strTwo := strings.Join(interestArrTwo, ",")

	return lib.CalcTextSimilarity(strOne, strTwo)
}


func (s *UserRecommendationService) CalcSimilarityBetweenMatches(userId string, recommendedUserId string) (float64, error) {

	//
	// Collect initial users interests
	//

	var initialUsersInterests []string

	rows, err := s.Con.Query(context.Background(), `
		select interest from interests where user_id = $1
	`, userId)

	if err != nil {
		fmt.Println(err)
		return -1, err
	}

	for rows.Next() {
		var interest Interest

		err = rows.Scan(&interest.InterestId, &interest.Interest, &interest.UserId)

		if err != nil {
			fmt.Println(err)
			return -1, err
		}

		initialUsersInterests = append(initialUsersInterests, interest)
	}

	//
	// Find recommended users matches interests using GetMatchesInterests func
	//

	matchesInterests, err := s.GetMatchesInterests(recommendedUserId)

	if err != nil {
		fmt.Println(err)
		return -1, err
	}

	//
	// Iterate through map, get list of interests, compare to initial users interests and keep running tab of similarity to calc average at the end
	//

	var totalSimilarity float64 = 0

	for _, interests := range matchesInterests {
		totalSimilarity += s.CalcInterestSimilarityBtwnUsers(initialUsersInterests, interests)
	}

}


func (s *UserRecommendationService) GetMatchesInterests(userId string) (map[string][]string, error) {
	matchesInterests := make(map[string][]string)

	rows, err := s.Con.Query(context.Background(), `
		select interest, user_id from interests i left join matches m on m.user_id_b = i.user_id
		where m.user_id_a = $1 and match_status = 'mutual' 
	`, userId)

	if err != nil {
		fmt.Println(err)
		//
		// I think if no results are found, it will error out above... So we need to check if userId is user_id_a
		//	
		rows, err := s.Con.Query(context.Background(), `
		select interest, user_id from interests i left join matches m on m.user_id_a = i.user_id
		where m.user_id_b = $1 and match_status = 'mutual' 
	`, userId)

		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		
		for rows.Next() {
			var interest Interest
			err = rows.Scan(&interest.InterestId, &interest.Interest, &interest.UserId)
			if err != nil {
				fmt.Println(err)
				// 
				// This could return nil and err, but be okay because no matches were found...
				//
				return nil, err
			}
			val, err := matchesInterests[interest.UserId] 

			if err {
				matchesInterests[interest.UserId] = []string{interest.Interest}
			} else {
				matchesInterests[interest.UserId] = append(val, interest.Interest)
			}
		}

		return matchesInterests, nil
	}

	for rows.Next() {
		var interest Interest
		err = rows.Scan(&interest.InterestId, &interest.Interest, &interest.UserId)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		val, err := matchesInterests[interest.UserId] 

		if err {
			matchesInterests[interest.UserId] = []string{interest.Interest}
		} else {
			matchesInterests[interest.UserId] = append(val, interest.Interest)
		}
	}

	return matchesInterests, nil
}