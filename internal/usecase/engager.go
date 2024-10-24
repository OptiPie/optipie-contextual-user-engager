package usecase

import (
	"context"
	"fmt"
	"github.com/OptiPie/optipie-contextual-user-engager/internal/infra/openaiapi"
	"github.com/OptiPie/optipie-contextual-user-engager/internal/infra/twitterapi"
	"log"
	"log/slog"
	"strings"
)

var _ Engager = &engager{}

type EngagerArgs struct {
	TwitterAPI *twitterapi.TwitterAPI
	OpenaiAPI  *openaiapi.OpenaiAPI
	UserNames  []string
}

func NewEngager(args *EngagerArgs) (Engager, error) {
	if args.TwitterAPI == nil {
		return nil, fmt.Errorf("twitterAPI can't be nil")
	}
	if args.OpenaiAPI == nil {
		return nil, fmt.Errorf("openaiAPI can't be nil")
	}
	if len(args.UserNames) == 0 {
		return nil, fmt.Errorf("userNames can't be empty")
	}
	return &engager{
		twitterAPI: args.TwitterAPI,
		openaiAPI:  args.OpenaiAPI,
		userNames:  args.UserNames,
	}, nil
}

type engager struct {
	twitterAPI *twitterapi.TwitterAPI
	openaiAPI  *openaiapi.OpenaiAPI
	userNames  []string
}

type Engager interface {
	Engage(ctx context.Context)
}

func (e *engager) Engage(ctx context.Context) {
	for _, userName := range e.userNames {
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
		}

		if replyTweetContent == "" {
			slog.Warn("tweet is not related", "content", tweetContent)
			continue
		}

		log.Printf("%v", replyTweetContent)

		repliedTweetId, err := e.twitterAPI.PostReplyTweet(ctx, tweetID, replyTweetContent)
		if err != nil {
			slog.Error("error on postReplyTweet", "error", err)
		}

		log.Printf("replied tweet for user, tweetId: %v", repliedTweetId)
	}
}

func countWords(s string) int {
	return len(strings.Fields(s))
}
