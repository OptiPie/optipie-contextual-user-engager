package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/OptiPie/optipie-contextual-user-engager/internal/app/prepare"
	dynamodbrepo "github.com/OptiPie/optipie-contextual-user-engager/internal/infra/dynamodb"
)

type config struct {
	cookie        string
	csrfToken     string
	authorization string
}

func loadConfig() config {
	cookie := os.Getenv("X_COOKIE")
	csrfToken := os.Getenv("X_CSRF_TOKEN")
	authorization := os.Getenv("X_AUTHORIZATION")

	if cookie == "" || csrfToken == "" || authorization == "" {
		log.Fatal("missing required env vars: X_COOKIE, X_CSRF_TOKEN, X_AUTHORIZATION")
	}

	return config{cookie: cookie, csrfToken: csrfToken, authorization: authorization}
}

type CheckResult struct {
	Username string
	Reason   string
}

func checkUser(client *http.Client, cfg config, username string) CheckResult {
	variables := fmt.Sprintf(`{"screen_name":"%s","withGrokTranslatedBio":true}`, strings.ToLower(username))
	features := `{"hidden_profile_subscriptions_enabled":true,"profile_label_improvements_pcf_label_in_post_enabled":true,"responsive_web_profile_redirect_enabled":false,"rweb_tipjar_consumption_enabled":false,"verified_phone_label_enabled":false,"subscriptions_verification_info_is_identity_verified_enabled":true,"subscriptions_verification_info_verified_since_enabled":true,"highlights_tweets_tab_ui_enabled":true,"responsive_web_twitter_article_notes_tab_enabled":true,"subscriptions_feature_can_gift_premium":true,"creator_subscriptions_tweet_preview_api_enabled":true,"responsive_web_graphql_skip_user_profile_image_extensions_enabled":false,"responsive_web_graphql_timeline_navigation_enabled":true}`
	fieldToggles := `{"withPayments":false,"withAuxiliaryUserLabels":true}`

	reqURL := fmt.Sprintf("https://x.com/i/api/graphql/IGgvgiOx4QZndDHuD3x9TQ/UserByScreenName?variables=%s&features=%s&fieldToggles=%s",
		url.QueryEscape(variables),
		url.QueryEscape(features),
		url.QueryEscape(fieldToggles),
	)

	req, _ := http.NewRequest("GET", reqURL, nil)
	req.Header.Set("authorization", cfg.authorization)
	req.Header.Set("cookie", cfg.cookie)
	req.Header.Set("x-csrf-token", cfg.csrfToken)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-twitter-auth-type", "OAuth2Session")
	req.Header.Set("x-twitter-active-user", "yes")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return CheckResult{username, "ERROR"}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if strings.Contains(string(body), "Rate limit exceeded") {
		fmt.Printf("%-30s // RATE LIMITED - SKIPPED\n", username)
		time.Sleep(5 * time.Second) // back off a bit
		return CheckResult{username, ""}
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return CheckResult{username, "DEAD"}
	}

	user, ok := data["user"].(map[string]interface{})
	if !ok || user == nil {
		return CheckResult{username, "DEAD"}
	}

	userResult, ok := user["result"].(map[string]interface{})
	if !ok {
		return CheckResult{username, "DEAD"}
	}

	if privacy, ok := userResult["privacy"].(map[string]interface{}); ok {
		if protected, ok := privacy["protected"].(bool); ok && protected {
			return CheckResult{username, "PROTECTED"}
		}
	}

	if rel, ok := userResult["relationship_perspectives"].(map[string]interface{}); ok {
		if blockedBy, ok := rel["blocked_by"].(bool); ok && blockedBy {
			return CheckResult{username, "BLOCKED"}
		}
	}

	return CheckResult{username, ""}
}

func main() {
	cfg := loadConfig()

	ctx := context.Background()
	awsCfg, err := prepare.AwsConfig(ctx)
	if err != nil {
		log.Fatalf("prepare aws config error: %v", err)
	}

	svc := prepare.Dynamodb(awsCfg)
	repository := dynamodbrepo.NewRepository(svc, "optipie-cue-users")

	file, _ := os.Open("results.csv")
	defer file.Close()

	reader := csv.NewReader(file)
	records, _ := reader.ReadAll()

	client := &http.Client{}

	fmt.Println("=== REMOVE USERS ===")
	for i, row := range records {
		if i == 0 {
			continue
		}
		username := row[0]
		result := checkUser(client, cfg, username)
		if result.Reason != "" {
			fmt.Printf("Delete: %-30s // %s\n", result.Username, result.Reason)
			err = repository.DeleteUser(ctx, username)
			if err != nil {
				fmt.Printf("failed to delete user %s: %v", username, err)
			}
		}
		time.Sleep(1 * time.Second)
	}
}
