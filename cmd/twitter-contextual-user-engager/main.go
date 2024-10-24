package main

import (
	"context"
	"github.com/OptiPie/optipie-contextual-user-engager/config"
	"github.com/OptiPie/optipie-contextual-user-engager/internal/infra/openaiapi"
	"github.com/OptiPie/optipie-contextual-user-engager/internal/infra/twitterapi"
	"github.com/OptiPie/optipie-contextual-user-engager/internal/usecase"
	_ "github.com/michimani/gotwi/user/userlookup"
	"log"
	"log/slog"
	"os"
)

var userNames = []string{
	"@KevinDaveyAlgo",
	"@tradingQnA",
	"@QuantInsti",
	"@BQuantR",
	"@QuantiCarlo",
	"@daily_quant",
	"@qc_alpha",
	"@alphatrader_ZZ",
	"@sobertrading",
	"@AlphaTrends",
	"@GoNoGoCharts",
	"@TrendSpider",
	"@CryptoHopper",
	"@Pentosh1",
	"@QuantConnect",
	"@AlgoTrader",
	"@ZorroTrader",
	"@TradingviewTeam",
	"@PineScriptArmy",
	"@NinjaTrader",
	"@cryptoquant_com",
	"@Jesse_Livermore",
	"@RT_Tenkan",
	"@Tradingcomposure",
	"@WaveTraders",
	"@Zen_Trading",
	"@BacktestRookies",
	"@PatternProfits",
	"@PullbackAlerts",
	"@Trader_Mader",
	"@Algo_Money",
	"@TraderMJ",
	"@AlpacaHQ",
	"@QuantitativeTrad",
	"@AlgoTradingGuy",
	"@HawkQuant",
	"@robwritestrades",
	"@volumehunter1",
	"@TraderJ_Algo",
	"@AlgoTradingGod",
	"@chrisdcarlson",
	"@AutomatedTrader",
	"@TradeIndicators",
	"@tendstrading",
	"@FiniAlgo",
	"@StratExecution",
	"@BotTraderAI",
	"@NabilFars",
	"@AutofuturesAI",
	"@Quant_Society",
}

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	appConfig, err := config.GetConfig()
	if err != nil {
		log.Fatalf("config can't be loaded, %v", err)
	}

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
		TwitterAPI: twitterAPI,
		OpenaiAPI:  openaiAPI,
		UserNames:  userNames,
	})

	if err != nil {
		log.Fatalf("engager can't be initialized, %v", err)
	}

	engager.Engage(ctx)
}
