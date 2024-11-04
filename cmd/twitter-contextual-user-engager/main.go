package main

import (
	"context"
	"github.com/OptiPie/optipie-contextual-user-engager/config"
	"github.com/OptiPie/optipie-contextual-user-engager/internal/app/prepare"
	dynamodbrepo "github.com/OptiPie/optipie-contextual-user-engager/internal/infra/dynamodb"
	"github.com/OptiPie/optipie-contextual-user-engager/internal/infra/openaiapi"
	"github.com/OptiPie/optipie-contextual-user-engager/internal/infra/twitterapi"
	"github.com/OptiPie/optipie-contextual-user-engager/internal/usecase"
	_ "github.com/michimani/gotwi/user/userlookup"
	"log"
	"log/slog"
	"os"
)

// list of users for initial table creation
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
	"CryptoMagnified",
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
	"TrendSpider",
	"TriggerTrades",
	"EliteOptions2",
	"traderstewie",
	"ValentinTrades",
	"QuantedTrading",
	"TheLongInvest",
	"BTC_Archive",
	"TRADE_TALK_",
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

	twitterAPI, err := twitterapi.NewTwitterAPI(twitterapi.NewTwitterAPIArgs{
		OAuthToken:       appConfig.Twitter.OAuthToken,
		OAuthTokenSecret: appConfig.Twitter.OAuthTokenSecret,
	})

	if err != nil {
		log.Fatalf("twitter api can't be initialized, %v", err)
	}

	openaiAPI, err := openaiapi.NewOpenaiAPI(openaiapi.NewOpenaiAPIArgs{
		OpenaiSecretKey: appConfig.Openai.SecretKey,
		SystemMessage: "You are admin of OptiPie TradingView Input Optimizer's Twitter account. " +
			"Reply to given tweets accordingly, promote it nicely and dont sound fake. If tweet isn't related to OptiPie/finance, " +
			"respond as `isRelated:false` otherwise true along with `reply:'message'`," +
			"put them in json and don't wrap in json markers",
	})

	if err != nil {
		log.Fatalf("openai api can't be initialized, %v", err)
	}

	engager, err := usecase.NewEngager(&usecase.EngagerArgs{
		TwitterAPI:     twitterAPI,
		OpenaiAPI:      openaiAPI,
		DynamoDbClient: repository,
		UserCount:      3, // pick 3 random users from dynamo user-names table
	})

	if err != nil {
		log.Fatalf("engager can't be initialized, %v", err)
	}

	engager.Engage(ctx)

}
