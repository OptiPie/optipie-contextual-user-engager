package usecase

import (
	"context"
	"fmt"
	dynamodbrepo "github.com/OptiPie/optipie-contextual-user-engager/internal/infra/dynamodb"
	dbmodels "github.com/OptiPie/optipie-contextual-user-engager/internal/infra/dynamodb/models"
	"github.com/OptiPie/optipie-contextual-user-engager/internal/infra/openaiapi"
	"github.com/OptiPie/optipie-contextual-user-engager/internal/infra/twitterapi"
	"log"
	"log/slog"
	"strings"
	"time"
)

var _ Engager = &engager{}

type EngagerArgs struct {
	TwitterAPI     *twitterapi.TwitterAPI
	OpenaiAPI      *openaiapi.OpenaiAPI
	DynamoDbClient *dynamodbrepo.Client
	UserCount      int
}

func NewEngager(args *EngagerArgs) (Engager, error) {
	if args.TwitterAPI == nil {
		return nil, fmt.Errorf("twitterAPI can't be nil")
	}
	if args.OpenaiAPI == nil {
		return nil, fmt.Errorf("openaiAPI can't be nil")
	}
	if args.DynamoDbClient == nil {
		return nil, fmt.Errorf("dynamoDbClient can't be nil")
	}
	if args.UserCount == 0 {
		return nil, fmt.Errorf("userCount can't be 0")
	}

	return &engager{
		twitterAPI:     args.TwitterAPI,
		openaiAPI:      args.OpenaiAPI,
		dynamoDbClient: args.DynamoDbClient,
		userCount:      args.UserCount,
	}, nil
}

type engager struct {
	twitterAPI     *twitterapi.TwitterAPI
	openaiAPI      *openaiapi.OpenaiAPI
	dynamoDbClient *dynamodbrepo.Client
	userCount      int
}

type Engager interface {
	Engage(ctx context.Context) error
}

func (e *engager) Engage(ctx context.Context) error {
	randomUserNames, isCycleFinished, err := e.prepareUserList(ctx)
	if err != nil {
		slog.Error("error on prepareUserList", "err", err)
	}

	// if cycle is finished, reset user list for next cycle
	if isCycleFinished {
		defer e.resetUserList(ctx)
	}

	slog.Info("randomUserNames are",
		"names", randomUserNames, "isCycleFinished", isCycleFinished)

	for _, userName := range randomUserNames {
		tweetID, err := e.twitterAPI.GetMostRecentTweetIDByUsername(ctx, userName)
		if err != nil {
			slog.Error("failed to run getMostRecentTweetIDByUsername",
				"err", err,
				"username", userName)
			continue
		}

		tweetContent, err := scrapeTweetContent(ctx, userName, tweetID)
		if err != nil {
			slog.Error("error on scrapeTweetContent", "err", err)
			continue
		}
		if countWords(tweetContent) < 3 {
			slog.Warn("tweetContent is too short", "content", tweetContent)
			continue
		}
		replyTweetContent, err := e.openaiAPI.CreateChat(ctx, tweetContent)
		if err != nil {
			slog.Error("error on CreateChat", "err", err)
			continue
		}

		if replyTweetContent == "" {
			slog.Warn("tweet is not related", "content", tweetContent)
			continue
		}

		log.Printf("%v", replyTweetContent)

		repliedTweetId, err := e.twitterAPI.PostReplyTweet(ctx, tweetID, replyTweetContent)
		if err != nil {
			slog.Error("error on postReplyTweet", "error", err)
			continue
		}

		err = e.dynamoDbClient.UpdateUser(ctx, userName, dbmodels.UpdateUserArgs{
			IsReplied:            true,
			LastRepliedTweetTime: time.Now(),
		})
		if err != nil {
			slog.Error("error on updateUser", "error", err)
			return err
		}

		log.Printf("replied tweet for user, tweetId: %v", repliedTweetId)
	}

	return nil
}

func countWords(s string) int {
	return len(strings.Fields(s))
}
