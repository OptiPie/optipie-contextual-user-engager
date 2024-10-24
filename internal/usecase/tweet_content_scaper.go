package usecase

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"strings"
	"time"
)

func scrapeTweetContent(ctx context.Context, userName, tweetID string) (string, error) {
	// Set up options with a custom user agent and run in non-headless mode if needed
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.WindowSize(1280, 800),
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	defer timeoutCancel()

	// URL of the tweet
	url := fmt.Sprintf("https://x.com/%s/status/%s", userName, tweetID)

	var tweetText string

	// Run chromedp tasks
	err := chromedp.Run(timeoutCtx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`div[data-testid="tweetText"]`, chromedp.ByQuery), // Give time for the page to fully load
		chromedp.TextContent(`div[data-testid="tweetText"]`, &tweetText, chromedp.NodeVisible),
	)
	if err != nil {
		return "", err
	}

	tweetText = strings.Replace(tweetText, "\n", "", -1)
	tweetText = strings.Join(strings.Fields(tweetText), " ")

	return tweetText, nil
}
