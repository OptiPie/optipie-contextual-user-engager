package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/OptiPie/optipie-contextual-user-engager/internal/app/prepare"
	dynamodbrepo "github.com/OptiPie/optipie-contextual-user-engager/internal/infra/dynamodb"
	dbmodels "github.com/OptiPie/optipie-contextual-user-engager/internal/infra/dynamodb/models"
)

func main() {
	ctx := context.Background()

	awsCfg, err := prepare.AwsConfig(ctx)
	if err != nil {
		log.Fatalf("prepare aws config error: %v", err)
	}

	svc := prepare.Dynamodb(awsCfg)
	repository := dynamodbrepo.NewRepository(svc, "optipie-cue-users")

	users, err := repository.GetUsers(ctx)
	if err != nil {
		slog.Error("error on getUsers", "err", err)
	}
	for _, user := range users {
		updateUser := dbmodels.UpdateUserArgs{
			IsReplied:            false,
			RepliedTweetCount:    user.RepliedTweetCount,
			LastRepliedTweetTime: time.Now(),
		}
		err = repository.UpdateUser(ctx, user.UserName, updateUser)
		if err != nil {
			slog.Error("error on updateUser",
				"err", err, "userName", user.UserName)
		}
	}
	slog.Info("all cue users are reset")
}
