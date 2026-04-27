package twitterapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type BrowserClientArgs struct {
	Username    string
	Password    string
	UserDataDir string
}

type BrowserClient struct {
	username     string
	password     string
	allocCtx     context.Context
	allocCancel  context.CancelFunc
	anchorCtx    context.Context
	anchorCancel context.CancelFunc
}

func NewBrowserClient(args BrowserClientArgs) (*BrowserClient, error) {
	if args.Username == "" {
		return nil, fmt.Errorf("username can't be empty")
	}
	if args.Password == "" {
		return nil, fmt.Errorf("password can't be empty")
	}
	if args.UserDataDir == "" {
		return nil, fmt.Errorf("UserDataDir can't be empty")
	}

	// FALLBACK: If ExecAllocator fails (websocket timeout), launch Chrome manually with debug port
	// and switch to chromedp.NewRemoteAllocator(ctx, "ws://127.0.0.1:9222") instead.
	// PowerShell command to launch Chrome with debug port:
	//   & "C:\Program Files\Google\Chrome\Application\chrome.exe" --remote-debugging-port=9222 --user-data-dir="C:\temp\fresh-chrome" --no-first-run
	// Then verify port is open: netstat -an | findstr 9222
	//
	// IMPORTANT (Windows EC2): UserDataDir MUST be a fresh directory (e.g. C:\temp\fresh-chrome),
	// NOT the existing Chrome user profile (C:\Users\...\AppData\Local\Google\Chrome\User Data).
	// The existing profile has a stale singleton lock that prevents Chrome from binding the
	// remote debugging port — chromedp silently fails with "websocket url timeout" and you'll
	// spend hours debugging a blank page. Fresh dir = no lock = works instantly.
	// Set via env: BROWSER_USER_DATA_DIR=C:\temp\fresh-chrome (use setx to persist across sessions)
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.UserDataDir(args.UserDataDir),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("enable-automation", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.WindowSize(1280, 800),
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// anchor tab keeps Chrome alive so cookies stay in memory across operations
	anchorCtx, anchorCancel := chromedp.NewContext(allocCtx)

	return &BrowserClient{
		username:     args.Username,
		password:     args.Password,
		allocCtx:     allocCtx,
		allocCancel:  allocCancel,
		anchorCtx:    anchorCtx,
		anchorCancel: anchorCancel,
	}, nil
}

// Close shuts down Chrome. Sleeps briefly so Chrome can flush session data to
// the UserDataDir before the process is killed.
func (bc *BrowserClient) Close() {
	time.Sleep(3 * time.Second)
	bc.anchorCancel()
	bc.allocCancel()
}

// newTab creates an isolated tab with its own timeout. Tabs share the session
// since they run in the same Chrome instance (same allocator = same cookie jar).
func (bc *BrowserClient) newTab(timeout time.Duration) (context.Context, context.CancelFunc) {
	tabCtx, tabCancel := chromedp.NewContext(bc.allocCtx)
	timeoutCtx, timeoutCancel := context.WithTimeout(tabCtx, timeout)
	return timeoutCtx, func() {
		timeoutCancel()
		tabCancel()
	}
}

// Login checks if already authenticated via saved session. If not, opens the login
// page and waits up to 3 minutes for manual login.
func (bc *BrowserClient) Login(_ context.Context) error {
	// Step 1: check if already logged in — close tab explicitly before opening the next one
	checkCtx, checkCancel := bc.newTab(30 * time.Second)
	var currentURL string
	_ = chromedp.Run(checkCtx,
		chromedp.Navigate("https://x.com/home"),
		chromedp.Sleep(3*time.Second),
		chromedp.Location(&currentURL),
	)
	checkCancel()

	if strings.Contains(currentURL, "/home") {
		slog.Info("already logged in", "url", currentURL)
		return nil
	}

	// Step 2: open login page and wait for manual login
	slog.Info("not logged in, waiting for manual login in the browser window (3 min timeout)...")
	manualCtx, manualCancel := bc.newTab(3 * time.Minute)
	defer manualCancel()
	if err := chromedp.Run(manualCtx,
		chromedp.Navigate("https://x.com/i/flow/login"),
		chromedp.WaitVisible(`[data-testid="SideNav_AccountSwitcher_Button"]`, chromedp.ByQuery),
		chromedp.Location(&currentURL),
		chromedp.Sleep(5*time.Second),
	); err != nil {
		return fmt.Errorf("manual login timed out or failed: %w", err)
	}

	slog.Info("login completed, session saved", "url", currentURL)
	return errors.New("session saved, please run it again")
}

// GetMostRecentTweetIDByUsername navigates to a user's profile and returns the
// tweet ID of their most recent non-pinned tweet.
func (bc *BrowserClient) GetMostRecentTweetIDByUsername(_ context.Context, username string) (string, error) {
	tabCtx, cancel := bc.newTab(30 * time.Second)
	defer cancel()

	profileURL := fmt.Sprintf("https://x.com/%s", username)
	var tweetID string

	if err := chromedp.Run(tabCtx,
		chromedp.Navigate(profileURL),
		chromedp.WaitVisible(`article[data-testid="tweet"]`, chromedp.ByQuery),
		chromedp.Sleep(3*time.Second),
		chromedp.Evaluate(`(function() {
			const tweets = document.querySelectorAll('article[data-testid="tweet"]');
			for (const tweet of tweets) {
				const cell = tweet.closest('[data-testid="cellInnerDiv"]');
				if (cell && cell.querySelector('[data-testid="socialContext"]')) {
					continue;
				}
				const links = tweet.querySelectorAll('a[href*="/status/"]');
				for (const link of links) {
					const m = link.href.match(/\/status\/(\d+)/);
					if (m) return m[1];
				}
			}
			return '';
		})()`, &tweetID),
	); err != nil {
		return "", fmt.Errorf("failed to get most recent tweet ID for @%s: %w", username, err)
	}

	if tweetID == "" {
		return "", fmt.Errorf("no non-pinned tweet found for @%s", username)
	}

	return tweetID, nil
}

// ScrapeTweetContent navigates to the tweet and returns its text content.
func (bc *BrowserClient) ScrapeTweetContent(_ context.Context, userName, tweetID string) (string, error) {
	tabCtx, cancel := bc.newTab(30 * time.Second)
	defer cancel()

	tweetURL := fmt.Sprintf("https://x.com/%s/status/%s", userName, tweetID)
	var tweetText string
	if err := chromedp.Run(tabCtx,
		chromedp.Navigate(tweetURL),
		chromedp.WaitVisible(`div[data-testid="tweetText"]`, chromedp.ByQuery),
		chromedp.TextContent(`div[data-testid="tweetText"]`, &tweetText, chromedp.NodeVisible),
	); err != nil {
		return "", err
	}

	return strings.Join(strings.Fields(tweetText), " "), nil
}

// PostReplyTweet navigates to the tweet and posts a reply via browser automation.
func (bc *BrowserClient) PostReplyTweet(_ context.Context, tweetID, userName, replyText string) error {
	tabCtx, cancel := bc.newTab(60 * time.Second)
	defer cancel()

	tweetURL := fmt.Sprintf("https://x.com/%s/status/%s", userName, tweetID)
	slog.Info("posting tweet", "url", tweetURL)
	err := chromedp.Run(tabCtx,
		chromedp.Navigate(tweetURL),
		chromedp.Sleep(5*time.Second),
		chromedp.WaitReady(`[data-testid="tweetTextarea_0"]`, chromedp.ByQuery),
		chromedp.Evaluate(`document.querySelector('[data-testid="tweetTextarea_0"]').scrollIntoView()`, nil),
		chromedp.Sleep(3*time.Second),
		chromedp.Click(`[data-testid="tweetTextarea_0"]`, chromedp.ByQuery),
		chromedp.SendKeys(`[data-testid="tweetTextarea_0"]`, replyText, chromedp.ByQuery),
		chromedp.Sleep(5*time.Second),
		chromedp.Evaluate(`document.querySelector('[data-testid="tweetButtonInline"]').click()`, nil),
		chromedp.Sleep(5*time.Second),
	)
	if err != nil {
		return err
	}
	return nil
}
