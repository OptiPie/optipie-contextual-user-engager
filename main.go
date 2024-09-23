package main

import (
	"context"
	"fmt"
	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/fields"
	"github.com/michimani/gotwi/user/userlookup"
	"github.com/michimani/gotwi/user/userlookup/types"
)

func main() {
	in := &gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           "***",
		OAuthTokenSecret:     "***",
	}

	c, err := gotwi.NewClient(in)
	if err != nil {
		fmt.Println(err)
		return
	}

	p := &types.GetByUsernameInput{
		Username: "optipieapp",
		UserFields: fields.UserFieldList{
			fields.UserFieldMostRecentTweetID,
		},
	}

	u, err := userlookup.GetByUsername(context.Background(), c, p)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("optipie related stuff here: \n %v", gotwi.StringValue(u.Data.MostRecentTweetID))
}
