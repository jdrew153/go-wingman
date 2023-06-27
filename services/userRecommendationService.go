package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"strconv"
	"strings"
	"sync"

	"github.com/jdrew153/lib"
)

type UserRecommendationService struct {
	Con *pgxpool.Pool
}

func NewUserRecommendationService(con *pgxpool.Pool) *UserRecommendationService {
	return &UserRecommendationService{Con: con}
}

type UserRecommendationModel struct {
	RecommendationId   string  `json:"recommendation_id"`
	UserId             string  `json:"user_id"`
	UserIdRecommended  string  `json:"user_id_recommended"`
	InterestSimilarity float64 `json:"interest_similarity"`
	Distance           float64 `json:"distance"`
	MatchSimilarity    float64 `json:"match_similarity"`
	LikeSimilarity     float64 `json:"like_similarity"`
	IsPotentialMatch   bool    `json:"is_potential_match"`
}

func (s *UserRecommendationService) BuildUserRecommendationModel(userId string) ([]UserRecommendationModel, error) {

	var recommendations []UserRecommendationModel

	//
	// Get list of users that is not the user who made the request.
	//
	rows, err := s.Con.Query(context.Background(), `
		select id from authuser where id != $1
	`, userId)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var recommendationUserId string

		err = rows.Scan(&recommendationUserId)

		if err != nil {
			return nil, err
		}

		//
		// Do the calc functions with go routines and wait group to sync after all calc funcs have finished
		//

		wg := sync.WaitGroup{}
		errorChan := make(chan error)
		matchSimilarityChan := make(chan float64, 1)

		wg.Add(1)
		go func() {
			fmt.Println("at wg 1")
			defer wg.Done()
			matchSimilarity, err := s.CalcSimilarityBetweenMatches(userId, recommendationUserId)

			if err != nil {
				fmt.Println("wg1", err)
				errorChan <- err
			}
			fmt.Println("wg 1 result", matchSimilarity)
			matchSimilarityChan <- matchSimilarity

		}()

		interestSimilarityChan := make(chan float64, 1)
		wg.Add(1)
		go func() {

			fmt.Println("at wg 2")
			wgGo := sync.WaitGroup{}
			errChanGo := make(chan error, 2)
			//requestingUserInterestArrChan := make(chan []string)
			//
			// Optimize later and remove this out of the loop because you only need to call this once...
			//
			var requestingUserInterestArrTest []string
			wgGo.Add(1)
			go func() {

				requestingUserInterests, err := s.FindInterestsByUserId(userId)

				if err != nil {
					fmt.Println("wg2", err)
					errChanGo <- err

				}
				fmt.Println("wg 2 result", requestingUserInterests)
				requestingUserInterestArrTest = append(requestingUserInterestArrTest, requestingUserInterests...)
				//requestingUserInterestArrChan <- requestingUserInterests
				wgGo.Done()
			}()

			//recommendedUserInterestArrChan := make(chan []string)
			var recommendedUserInterestArrTest []string
			wgGo.Add(1)
			go func() {
				fmt.Println("at wg 3")

				recommendedUserInterests, err := s.FindInterestsByUserId(recommendationUserId)

				if err != nil {
					fmt.Println("wg3", err)
					errChanGo <- err

				}
				fmt.Println("wg 3 result", recommendedUserInterests)
				recommendedUserInterestArrTest = append(recommendedUserInterestArrTest, recommendedUserInterests...)
				//recommendedUserInterestArrChan <- recommendedUserInterests
				wgGo.Done()
			}()

			go func() {
				for potentialError := range errChanGo {
					errorChan <- potentialError
				}
				close(errorChan)
			}()
			fmt.Println("before wg 6 end")

			wgGo.Wait()
			fmt.Println("after wg 6 end")

			//
			// Got both users interests, now time to get similarity
			//
			//requestingUserInterestArr := <-requestingUserInterestArrChan
			//recommendedUserInterestArr := <-recommendedUserInterestArrChan

			result := s.CalcInterestSimilarityBtwnUsers(requestingUserInterestArrTest, recommendedUserInterestArrTest)
			fmt.Println("wg6 result", result)
			interestSimilarityChan <- result
			wg.Done()

		}()

		isPotentialMatchResultChan := make(chan bool)
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println("at wg 4")
			result, err := s.IsPotentialMatch(userId, recommendationUserId)

			if err != nil {
				fmt.Println("wg4", err)
				errorChan <- err
			}
			fmt.Println("wg 4 result", result)
			isPotentialMatchResultChan <- result

		}()

		distanceResultChan := make(chan float64)
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println("at wg5")
			result, err := s.CalcDistanceBetweenUsersGPT(userId, recommendationUserId)

			if err != nil {
				fmt.Println("wg5", err)
				errorChan <- err
			}
			fmt.Println("wg 5 result", result)
			distanceResultChan <- result
		}()

		matchSimilarity := <-matchSimilarityChan
		fmt.Println(matchSimilarity)
		interestSimilarity := <-interestSimilarityChan
		fmt.Println(interestSimilarity)
		isPotentialMatchResult := <-isPotentialMatchResultChan
		fmt.Println(isPotentialMatchResult)
		distanceResult := <-distanceResultChan
		fmt.Println(distanceResult)

		wg.Wait()

		if len(errorChan) > 0 {
			potentialErr := <-errorChan

			return nil, potentialErr
		}

		recommendation := UserRecommendationModel{
			RecommendationId:   uuid.New().String(),
			UserId:             userId,
			UserIdRecommended:  recommendationUserId,
			InterestSimilarity: interestSimilarity,
			Distance:           distanceResult, // need to resolve this
			MatchSimilarity:    matchSimilarity,
			LikeSimilarity:     0, // need to come up  with something for this....
			IsPotentialMatch:   isPotentialMatchResult,
		}

		recommendations = append(recommendations, recommendation)

		_, err = s.SaveNewUserRecommendationModel(recommendation)

		if err != nil {
			return nil, err
		}

	}

	return recommendations, nil
}

func (s *UserRecommendationService) CalcInterestSimilarityBtwnUsers(intersestArrOne []string, interestArrTwo []string) float64 {
	strOne := strings.Join(intersestArrOne, ",")
	strTwo := strings.Join(interestArrTwo, ",")

	result := lib.CalcTextSimilarity(strOne, strTwo)
	fmt.Println("interest similarity calc", result)
	return result
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
		fmt.Println("ERROR	executing query wg 1", err)
		return -1, err
	}

	for rows.Next() {
		var interest string

		err = rows.Scan(&interest)

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

	fmt.Println(initialUsersInterests)

	for _, interests := range matchesInterests {

		totalSimilarity += s.CalcInterestSimilarityBtwnUsers(initialUsersInterests, interests)
	}

	if len(matchesInterests) == 0 {
		return -1, nil
	}

	return totalSimilarity / float64(len(matchesInterests)), nil

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
	fmt.Println(matchesInterests)
	return matchesInterests, nil
}

func (s *UserRecommendationService) SaveNewUserRecommendationModel(model UserRecommendationModel) (bool, error) {
	var recommendationResult string

	err := s.Con.QueryRow(context.Background(), `
   		insert 
   		into user_recommendations  
   		(recommendation_id, user_id, user_id_recommended, interest_similarity, distance, match_similarity, like_similarity, is_potential_match)
   		values ($1,$2,$3, $4, $5, $6, $7, $8) returning recommendation_id
	`, model.RecommendationId, model.UserId, model.UserIdRecommended, model.InterestSimilarity,
		model.Distance, model.MatchSimilarity, model.LikeSimilarity, model.IsPotentialMatch).Scan(&recommendationResult)

	if err != nil {
		return false, err
	}

	return len(recommendationResult) > 0, nil
}

//
// Helper funcs to get info that may be in other services...
//

func (s *UserRecommendationService) IsPotentialMatch(requestingUserId string, recommendedUserId string) (bool, error) {

	result := 0

	err := s.Con.QueryRow(context.Background(), `
		select count(*) from matches where user_id_a = $1 and user_id_b = $2 and match_status = 'pending'
	`, recommendedUserId, requestingUserId).Scan(&result)

	if err != nil {
		return false, err
	}

	return result > 0, nil
}

func (s *UserRecommendationService) FindInterestsByUserId(userId string) ([]string, error) {
	var interests []string

	rows, err := s.Con.Query(context.Background(), `
    	select interest from interests where user_id = $1
    `, userId)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var interest string

		err = rows.Scan(&interest)

		if err != nil {
			return nil, err
		}

		interests = append(interests, interest)
	}

	return interests, nil
}

type LocationModel struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

func (s *UserRecommendationService) CalcDistanceBetweenUsers(requestingUserId string, recommendedUserId string) (float64, error) {

	wg := &sync.WaitGroup{}
	errorChan := make(chan error)
	requestingUserCoordsChan := make(chan UserLocationRequestModel)
	recommendedUserCoordsChan := make(chan UserLocationRequestModel)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		var requestingUserLocation LocationModel
		err := s.Con.QueryRow(context.Background(), `
			select latitude, longitude from authuser where id = $1
		`, requestingUserId).Scan(&requestingUserLocation.Latitude, &requestingUserLocation.Longitude)

		if err != nil {
			fmt.Println("err at wg 1 in calc dist func", err)
			errorChan <- err
		}

		lat, err := strconv.ParseFloat(requestingUserLocation.Latitude, 64)

		if err != nil {
			fmt.Println("error conv str to float in wg1", err)
			errorChan <- err
		}

		long, err := strconv.ParseFloat(requestingUserLocation.Longitude, 64)

		if err != nil {
			fmt.Println("error conv str to float in wg1", err)
			errorChan <- err
		}

		fmt.Println(lat, long)
		result := UserLocationRequestModel{
			Latitude:  lat,
			Longitude: long,
		}

		requestingUserCoordsChan <- result
		//close(requestingUserCoordsChan)
	}(wg)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		var recommendedUserLocation LocationModel
		err := s.Con.QueryRow(context.Background(), `
			select latitude, longitude from authuser where id = $1
		`, recommendedUserId).Scan(&recommendedUserLocation.Latitude, &recommendedUserLocation.Longitude)

		if err != nil {
			fmt.Println("err at wg 2 in calc dist func", err)
			errorChan <- err
		}
		lat, err := strconv.ParseFloat(recommendedUserLocation.Latitude, 64)

		if err != nil {
			fmt.Println("error conv str to float in wg2", err)
			errorChan <- err
		}

		long, err := strconv.ParseFloat(recommendedUserLocation.Longitude, 64)

		if err != nil {
			fmt.Println("error conv str to float in wg2", err)
			errorChan <- err
		}

		fmt.Println(lat, long)

		result := UserLocationRequestModel{
			Latitude:  lat,
			Longitude: long,
		}

		recommendedUserCoordsChan <- result
		//close(recommendedUserCoordsChan)
	}(wg)
	requestingUserCoords := <-requestingUserCoordsChan
	recommendedUserCoords := <-recommendedUserCoordsChan
	fmt.Println("before wait")
	wg.Wait()
	fmt.Println("after wait")

	//close(errorChan)

	fmt.Println("finished wait group execution.... at dist calc")

	if len(errorChan) > 0 {
		potentialError := <-errorChan

		return -1, potentialError
	}

	distance := lib.GetDistanceFromCoords(requestingUserCoords.Latitude, requestingUserCoords.Longitude, recommendedUserCoords.Latitude, recommendedUserCoords.Latitude)

	fmt.Println(distance)

	return distance, nil
}

func (s *UserRecommendationService) CalcDistanceBetweenUsersGPT(requestingUserId string, recommendedUserId string) (float64, error) {
	var requestingUserCoords UserLocationRequestModel
	var recommendedUserCoords UserLocationRequestModel
	var requestingUserLocation LocationModel
	var recommendedUserLocation LocationModel
	var requestingUserErr error
	var recommendedUserErr error

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		requestingUserErr = s.Con.QueryRow(context.Background(), `
            SELECT latitude, longitude FROM authuser WHERE id = $1
        `, requestingUserId).Scan(&requestingUserLocation.Latitude, &requestingUserLocation.Longitude)
	}()

	go func() {
		defer wg.Done()
		recommendedUserErr = s.Con.QueryRow(context.Background(), `
            SELECT latitude, longitude FROM authuser WHERE id = $1
        `, recommendedUserId).Scan(&recommendedUserLocation.Latitude, &recommendedUserLocation.Longitude)
	}()

	wg.Wait()

	if requestingUserErr != nil {
		return -1, requestingUserErr
	}

	if recommendedUserErr != nil {
		return -1, recommendedUserErr
	}

	requestingUserCoords.Latitude, _ = strconv.ParseFloat(requestingUserLocation.Latitude, 64)
	requestingUserCoords.Longitude, _ = strconv.ParseFloat(requestingUserLocation.Longitude, 64)
	recommendedUserCoords.Latitude, _ = strconv.ParseFloat(recommendedUserLocation.Latitude, 64)
	recommendedUserCoords.Longitude, _ = strconv.ParseFloat(recommendedUserLocation.Longitude, 64)

	distance := lib.GetDistanceFromCoords(requestingUserCoords.Latitude, requestingUserCoords.Longitude, recommendedUserCoords.Latitude, recommendedUserCoords.Longitude)

	fmt.Println(distance)

	return distance, nil
}
