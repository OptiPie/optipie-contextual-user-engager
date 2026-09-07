package usecase

import (
	"context"
	"log/slog"
	"math/rand"
	"slices"
	"time"

	dbmodels "github.com/OptiPie/optipie-contextual-user-engager/internal/infra/dynamodb/models"
)

// prepareUserList selects 3 random users from dynamodb, also returns if cycle is finished
func (e *engager) prepareUserList(ctx context.Context) ([]dbmodels.User, bool, error) {
	var isCycleFinished bool
	var userCount = e.userCount
	users, err := e.dynamoDbClient.GetUsersToReply(ctx)
	if err != nil {
		return nil, false, err
	}

	// if all users got replied in this cycle, reset users
	if len(users) <= userCount {
		isCycleFinished = true
		userCount = len(users)
	}

	var randomIndexes []int
	// random indexes to pick user names
	for {
		if len(randomIndexes) == userCount {
			break
		}

		randomIndex := rand.Intn(len(users))
		if !slices.Contains(randomIndexes, randomIndex) {
			randomIndexes = append(randomIndexes, randomIndex)
		}
	}

	randomUsers := make([]dbmodels.User, userCount)

	for i := range randomUsers {
		randomUser := users[randomIndexes[i]]
		// make sure that assigned user has id available, if not save it to db
		if randomUser.UserTwitterID == "" {
			userID, err := e.twitterAPI.GetUserIDByUsername(ctx, randomUser.UserName)
			if err != nil {
				slog.Error("error on get user id by username", slog.String("username", randomUser.UserName))
				return nil, false, err
			}
			err = e.dynamoDbClient.UpdateUser(ctx, randomUser.UserName, dbmodels.UpdateUserArgs{
				IsReplied:            randomUser.IsReplied,
				RepliedTweetCount:    randomUser.RepliedTweetCount,
				LastRepliedTweetTime: randomUser.LastRepliedTweetTime,
				UserTwitterID:        userID,
			})
			if err != nil {
				slog.Error("error on update user id by username", slog.String("username", randomUser.UserName))
				return nil, false, err
			}
			// assign twitter user ID back to user
			randomUser.UserTwitterID = userID
		}
		randomUsers[i] = randomUser
	}

	return randomUsers, isCycleFinished, nil
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
