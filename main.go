package main

import (
	"flag"
	"fmt"
	logSimple "log"
	"math/rand"
	"os"
	"time"

	"github.com/vikebot/vbgs/pkg/ntfydistr"

	"github.com/vikebot/vbcore"
	"github.com/vikebot/vbdb"
	"github.com/vikebot/vbgs/vbge"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log    *zap.Logger
	config *gameserverConfig

	// battle is the game (mapentity with players)
	battle          *vbge.Battle
	envDisableCrypt bool
	dist            ntfydistr.Distributor
)

func gsInit() {
	val, exists := os.LookupEnv("VB_DISABLE_CRYPT")
	if exists && val == "1" {
		envDisableCrypt = true
	}

	log.Info("seeding global PRNG-source")
	noice, err := vbcore.CryptoGenBytes(1)
	if err != nil {
		log.Fatal("failed getting noice value for seeding global PRNG-source", zap.Error(err))
	}
	if noice[0] == 0 {
		noice[0] = 1
	}
	rand.Seed(time.Now().UnixNano() / int64(noice[0]))

	log.Info("init database connections")
	vbdbConfig := &vbdb.Config{
		DbAddr: vbcore.NewEndpointAddr(config.Database.MariaDB.Host),
		DbUser: config.Database.MariaDB.User,
		DbPass: config.Database.MariaDB.Password,
		DbName: config.Database.MariaDB.Name,
	}
	err = vbdb.Init(vbdbConfig, log)
	if err != nil {
		log.Fatal("init failed", zap.Error(err))
	}

	registryInit()
}

func battleInit(joinedPlayers []int) {
	battle = &vbge.Battle{
		// MapSize
		Map:     vbge.NewMapEntity(vbge.MapHeight, vbge.MapWidth),
		Players: make(map[int]*vbge.Player),
	}
	// MapSize
	vbge.SetMapDimensions(vbge.MapHeight, vbge.MapWidth)

	for _, j := range joinedPlayers {
		p, err := vbge.NewPlayerWithSpawn(j, battle.Map)
		if err != nil {
			log.Fatal("failed to init vbge/(*Player) struct", zap.Error(err))
		}
		battle.Players[j] = p
	}

}

func main() {
	conf := flag.String("config", "", "path to config file")
	flag.Parse()

	if conf == nil || *conf == "" {
		logSimple.Fatal("no gameserver config defined")
	}
	config = loadConfig(*conf)

	// init zap logging
	initLog()

	// Prepare basic stuff of the server and init our battle (fetch map)
	gsInit()

	// getAllPlayers
	joinedPlayers := getJoinedPlayers()

	// init the battle
	battleInit(joinedPlayers)

	// init the distributor
	distributorInit(joinedPlayers)
	defer dist.Close()

	// Start and shutdown channels
	startChan := make(chan bool)
	shutdownChan := make(chan bool)

	// Start the network services
	ntcpInit(startChan, shutdownChan)
	nwsInit(startChan, shutdownChan)

	// Sleep till start
	startTime := time.Now().UTC().Add(time.Second * 2)
	sleepDuration := startTime.Sub(time.Now().UTC())
	log.Info("prepared services. sleeping till starttime",
		zap.Time("starttime", startTime),
		zap.Duration("sleeping", sleepDuration))
	time.Sleep(sleepDuration)

	// Activate services that listen on starting channel signal
	startChan <- true
	startChan <- true
	startChan <- true

	// Shutdown services in on hour
	log.Info("started services. sleeping till shutdown",
		zap.Time("shutdowntim", time.Now().UTC().Add(time.Hour*1)))
	time.Sleep(time.Hour * 1) // Time of a game
	shutdownChan <- true
}

func initLog() {
	// Logging server
	enablerFunc := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= config.Log.Level
	})
	var encoder zapcore.Encoder
	switch config.Log.Config {
	case "development", "dev":
		zapConfig := zap.NewDevelopmentConfig()
		if config.Log.Colored {
			zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		encoder = zapcore.NewConsoleEncoder(zapConfig.EncoderConfig)
	case "production", "prod":
		zapConfig := zap.NewProductionConfig()
		encoder = zapcore.NewJSONEncoder(zapConfig.EncoderConfig)
	default:
		fmt.Println("config.log.config is of unknown type. only 'development', 'dev', 'production' and 'prod' are allowed")
		os.Exit(-1)
	}
	core := zapcore.NewTee(zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), enablerFunc))
	log = zap.New(core)
}

func getJoinedPlayers() (joinedPlayers []int) {
	joined, success := vbdb.JoinedUsersCtx(config.Battle.RoundID, log)
	if !success {
		log.Fatal("unable to load users for this round", zap.Int("round_id", config.Battle.RoundID))
	}

	return joined
}

func distributorInit(joinedPlayers []int) {
	// TODO: make global stop chan
	stop := make(chan struct{})
	dist = ntfydistr.NewDistributor(joinedPlayers, stop, log.Named("Distributor"))
}
