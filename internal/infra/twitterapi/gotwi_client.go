package twitterapi

import (
	"context"
	"fmt"
	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/fields"
	"github.com/michimani/gotwi/tweet/managetweet"
	tweettypes "github.com/michimani/gotwi/tweet/managetweet/types"
	"github.com/michimani/gotwi/user/userlookup"
	usertypes "github.com/michimani/gotwi/user/userlookup/types"
)

type NewTwitterAPIArgs struct {
	OAuthToken       string
	OAuthTokenSecret string
}

type TwitterAPI struct {
	gotwiClient *gotwi.Client
}

func NewTwitterAPI(args NewTwitterAPIArgs) (*TwitterAPI, error) {
	if args.OAuthToken == "" {
		return nil, fmt.Errorf("oAuthToken can't be nil")
	}
	if args.OAuthTokenSecret == "" {
		return nil, fmt.Errorf("oAuthTokenSecret can't be nil")
	}

	in := &gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           args.OAuthToken,
		OAuthTokenSecret:     args.OAuthTokenSecret,
	}

	gotwiClient, err := gotwi.NewClient(in)
	if err != nil {
		return nil, err
	}

	return &TwitterAPI{
		gotwiClient,
	}, nil
}

func (ta *TwitterAPI) GetMostRecentTweetIDByUsername(ctx context.Context, userName string) (string, error) {
	input := &usertypes.GetByUsernameInput{
		Username: userName,
		UserFields: fields.UserFieldList{
			fields.UserFieldMostRecentTweetID,
		},
	}

	userOutput, err := userlookup.GetByUsername(ctx, ta.gotwiClient, input)
	if err != nil {
		return "", err
	}

	tweetID := gotwi.StringValue(userOutput.Data.MostRecentTweetID)
	return tweetID, nil
}

func (ta *TwitterAPI) PostReplyTweet(ctx context.Context, inReplyToTweetID string, replyTweetText string) (string, error) {
	input := &tweettypes.CreateInput{
		Reply: &tweettypes.CreateInputReply{
			InReplyToTweetID: inReplyToTweetID,
		},
		Text: &replyTweetText,
	}

	replyTweetOutput, err := managetweet.Create(ctx, ta.gotwiClient, input)
	if err != nil {
		return "", err
	}
	repliedTweetId := gotwi.StringValue(replyTweetOutput.Data.ID)

	return repliedTweetId, nil
}
