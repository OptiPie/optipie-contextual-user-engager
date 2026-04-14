package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/OptiPie/optipie-contextual-user-engager/config"
	"github.com/OptiPie/optipie-contextual-user-engager/internal/app/prepare"
	dynamodbrepo "github.com/OptiPie/optipie-contextual-user-engager/internal/infra/dynamodb"
	"github.com/OptiPie/optipie-contextual-user-engager/internal/infra/openaiapi"
	"github.com/OptiPie/optipie-contextual-user-engager/internal/infra/twitterapi"
	"github.com/OptiPie/optipie-contextual-user-engager/internal/usecase"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	_ "github.com/michimani/gotwi/user/userlookup"
)

// list of users for initial table creation
// NOT UP-TO-DATE
var _ = []string{
	"rektcapital",
	"TraderLion_",
	"CryptoTony__",
	"IncomeSharks",
	"tradersreality",
	"quantscience_",
	"VentureCoinist",
	"AlgoTradingGuy",
	"CryptoBoss1984",
	"QuintenFrancois",
	"CryptoFaibik",
	"Alejandro_XBT",
	"blackwidowbtc",
	"MrRakun35",
	"GauravGomase7",
	"CryptoDonAlt",
	"BigCheds",
	"TechTradesTT",
	"EWAnalysis",
	"Trader_muru",
	"Prof_heist",
	"ramazan1833853",
	"TriggerTrades",
	"EliteOptions2",
	"traderstewie",
	"ValentinTrades",
	"QuantedTrading",
	"TheLongInvest",
	"BTC_Archive",
	"TRADE_TALK_",
	"Tradermayne",
	"CoinDesk",
	"AltcoinPsycho",
	"CryptoCapo_",
	"StockMKTNewz",
	"meetblossomapp",
	"NebraskanGooner",
	"KioseffTrading",
	"LuxAlgo",
	"Chart_School",
	"trendoscope",
	"thisisorlando",
	"NicTrades",
	"RektProof",
	"AlessioRastani",
	"TradingView",
}

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	appConfig, err := config.GetConfig()
	if err != nil {
		log.Fatalf("config can't be loaded, %v", err)
	}

	awsCfg, err := prepare.AwsConfig(ctx)
	if err != nil {
		log.Fatalf("prepare aws config error: %v", err)
	}

	svc := prepare.Dynamodb(awsCfg)
	repository := dynamodbrepo.NewRepository(svc, "optipie-cue-users")

	lambdaClient := prepare.Lambda(awsCfg)

	defer func() {
		_, err = lambdaClient.Invoke(ctx, &lambda.InvokeInput{
			FunctionName: aws.String("arn:aws:lambda:eu-central-1:257797589448:function:optipie-contextual-user-engager-stoper"),
		})

		if err != nil {
			log.Printf("err on lambda invoker optipie-contextual-user-engager-stoper: %v", err)
		}
	}()

	twitterAPI, err := twitterapi.NewTwitterAPI(twitterapi.NewTwitterAPIArgs{
		OAuthToken:       appConfig.Twitter.OAuthToken,
		OAuthTokenSecret: appConfig.Twitter.OAuthTokenSecret,
	})
	if err != nil {
		log.Fatalf("twitter api can't be initialized, %v", err)
	}

	browserClient, err := twitterapi.NewBrowserClient(twitterapi.BrowserClientArgs{
		Username:    appConfig.Twitter.User,
		Password:    appConfig.Twitter.Password,
		UserDataDir: appConfig.Browser.UserDataDir,
	})
	if err != nil {
		log.Fatalf("browser client can't be initialized, %v", err)
	}
	defer browserClient.Close()

	if err = browserClient.Login(ctx); err != nil {
		browserClient.Close()
		log.Fatalf("browser client login failed, %v", err)
	}

	openaiAPI, err := openaiapi.NewOpenaiAPI(openaiapi.NewOpenaiAPIArgs{
		OpenaiSecretKey: appConfig.Openai.SecretKey,
		SystemMessage: "You are admin of OptiPie TradingView Input Optimizer's Twitter account. " +
			"Reply to given tweets accordingly, promote it nicely and keep it short. Sound like a real person: casual, direct, no buzzwords, no exclamation marks, no filler phrases like 'great point' or 'absolutely'. " +
				"Occasionally use 1-2 relevant hashtags or emojis when they feel natural, but don't force them on every reply. " +
			"If tweet isn't related to OptiPie/finance, " +
			"respond as `isRelated:false` otherwise true along with `reply:'message'`," +
			"put them in json and don't wrap in json markers",
	})

	if err != nil {
		log.Fatalf("openai api can't be initialized, %v", err)
	}

	engager, err := usecase.NewEngager(&usecase.EngagerArgs{
		TwitterAPI:     twitterAPI,
		BrowserClient:  browserClient,
		OpenaiAPI:      openaiAPI,
		DynamoDbClient: repository,
		UserCount:      3, // pick x random users from dynamo user-names table, twitter free api rate limits
	})

	if err != nil {
		log.Fatalf("engager can't be initialized, %v", err)
	}

	err = engager.Engage(ctx)

	if err != nil {
		log.Fatalf("error on Engage %v", err)
	}

	log.Printf("engager completed the cycle")
}
