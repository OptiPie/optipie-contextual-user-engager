package usecase

import (
	"context"
	dbmodels "github.com/OptiPie/optipie-contextual-user-engager/internal/infra/dynamodb/models"
	"log/slog"
	"math/rand"
	"slices"
	"time"
)

// prepareUserList selects 3 random users from dynamodb, also returns if cycle is finished
func (e *engager) prepareUserList(ctx context.Context) ([]string, bool, error) {
	var isCycleFinished bool
	var userCount = e.userCount
	userNames, err := e.dynamoDbClient.GetUserNamesToReply(ctx)
	if err != nil {
		return nil, isCycleFinished, err
	}

	// if all users got replied in this cycle, reset users
	if len(userNames) <= userCount {
		isCycleFinished = true
		userCount = len(userNames)
	}

	var randomIndexes []int
	// random indexes to pick user names
	for {
		if len(randomIndexes) == userCount {
			break
		}

		randomIndex := rand.Intn(len(userNames))
		if !slices.Contains(randomIndexes, randomIndex) {
			randomIndexes = append(randomIndexes, randomIndex)
		}
	}

	randomUserNames := make([]string, userCount)

	for i := range randomUserNames {
		randomUserNames[i] = userNames[randomIndexes[i]]
	}

	return randomUserNames, isCycleFinished, nil
}

// resetUserList resets user list and prepares it for next cycle
func (e *engager) resetUserList(ctx context.Context) {
	users, err := e.dynamoDbClient.GetUsers(ctx)
	if err != nil {
		slog.Error("error on getUsers", "err", err)
	}
	for _, user := range users {
		updateUser := dbmodels.UpdateUserArgs{
			IsReplied:            false,
			RepliedTweetCount:    user.RepliedTweetCount,
			LastRepliedTweetTime: time.Now(),
		}
		err = e.dynamoDbClient.UpdateUser(ctx, user.UserName, updateUser)
		if err != nil {
			slog.Error("error on updateUser",
				"err", err, "userName", user.UserName)
		}
	}
}
