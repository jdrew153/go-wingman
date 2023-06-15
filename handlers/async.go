package handlers

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/jdrew153/services"
	"sync"
)

type AsyncHandler struct {
	AService *services.AsyncService
	SService *services.UserStorage
}

func NewAsyncHandler(s *services.AsyncService, u *services.UserStorage) *AsyncHandler {
	return &AsyncHandler{
		AService: s,
		SService: u,
	}
}

func (h *AsyncHandler) TestableAsyncFunction(ctx *fiber.Ctx) error {
	userId := ctx.Params("userId")

	wg := sync.WaitGroup{}

	// error chan
	errChan := make(chan error, 2)

	listAllUsersChan := make(chan []services.NewUser, 1)

	wg.Add(1)
	go func() {
		fmt.Println("hitting first go routine")
		defer wg.Done()
		listUsers, err := services.CreateLargeListOfUsers(userId, h.SService.Con)

		if err != nil {
			fmt.Println("err at first go func", err)
			errChan <- err
			return
		}
		listAllUsersChan <- listUsers
		fmt.Println("should be finishing first go routine...")
		close(listAllUsersChan)
	}()

	listToFilterChan := make(chan map[string]string, 1)

	wg.Add(1)
	go func() {
		fmt.Println("hitting second go routine")
		defer wg.Done()
		listToFilter, err := h.AService.UserIdsWhoUserHasSentMatchRequestTo(context.Background(), userId)

		if err != nil {
			fmt.Println("err at second go func", err)
			errChan <- err
		}
		fmt.Println("should be finishing second go routine...")
		listToFilterChan <- listToFilter
		close(listToFilterChan)
	}()

	settledMatchRequestChan := make(chan map[string]string, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		settled, err := h.AService.FilterSettledMatchRequestsAsync(context.Background(), userId)

		if err != nil {
			fmt.Println("err at second go func", err)
			errChan <- err
		}
		fmt.Println("should be finishing third go routine...")

		settledMatchRequestChan <- settled
		close(settledMatchRequestChan)
	}()

	list1 := <-listAllUsersChan
	list2 := <-listToFilterChan
	list3 := <-settledMatchRequestChan

	wg.Wait()
	close(errChan)

	fmt.Println("wg has completed...")

	for err := range errChan {
		if err != nil {
			return ctx.Status(500).SendString("Something went wrong... " + err.Error())
		}
	}

	returnList := make(map[string]services.NewUser)

	if len(list2) > 0 {
		for _, user := range list1 {
			found := false
			if _, ok := list2[user.Id]; ok {
				found = true
			}

			if !found {
				returnList[user.Id] = user
			}
		}
	} else {
		fmt.Println("list2 has no results...")
	}

	fmt.Println("list after first filter", returnList)

	if len(list3) > 0 {
		for _, user := range returnList {
			found := false
			if _, ok := list3[user.Id]; ok {
				println("found user in list..")
				found = true
			}
			if found {
				delete(returnList, user.Id)
			}
		}
	} else {
		fmt.Println("list 3 has no results..")
	}

	var hydratedUserList []services.UserContext

	for _, v := range returnList {
		user, err := h.AService.CreateUserContextAsync(context.Background(), v)

		if err != nil {
			fmt.Println("error occurred while hydrating user...")
			return ctx.Status(500).SendString("Something went wrong..")
		}

		fmt.Println("potential match check in flight")

		referencedUser, err := services.PotentialMatchFlagCheck(&user, userId, h.SService.Con)

		if err != nil {
			fmt.Println("error occurred while hydrating user...")
			return ctx.Status(500).SendString("Something went wrong..")
		}

		hydratedUserList = append(hydratedUserList, *referencedUser)
	}

	return ctx.JSON(hydratedUserList)

}

func (h *AsyncHandler) TestableAsyncFunctionTwo(ctx *fiber.Ctx) error {
	userId := ctx.Params("userId")

	listUsers, err := services.CreateLargeListOfUsers(userId, h.SService.Con)

	if err != nil {
		return ctx.Status(500).SendString("Something went wrong... " + err.Error())
	}

	listToFilter, err := h.AService.UserIdsWhoUserHasSentMatchRequestTo(context.Background(), userId)

	if err != nil {
		return ctx.Status(500).SendString("Something went wrong... " + err.Error())
	}

	fmt.Println(listUsers)
	fmt.Println(listToFilter)
	return ctx.SendString("hi")

}
