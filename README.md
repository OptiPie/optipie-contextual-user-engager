# optipie-contextual-user-engager
Automatically engage with users, based on the context of their last tweet

## Execution Steps

1. Runs on EC2 machine which Scheduled by AWS EventBridge Scheduler
2. Executable runs on WinStartup
3. Engagement completes
4. EC2 stopper lambda called at the end, instance stops until further schedule to avoid recurring costs

## Important Troubleshooting
- Once saved session fails, change the session name(increment) ENV var($BROWSER_USER_DATA_DIR) 
to start off a new chrome profile and complete the manual login. 
- From "C:\temp\fresh-chrome-{X}" to "C:\temp\fresh-chrome-{X+1}"