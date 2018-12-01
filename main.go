package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/eapache/queue"
	"github.com/vikebot/vbcore"
	"github.com/vikebot/vbdb"
	"github.com/vikebot/vbgs/vbge"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logctx *zap.Logger
	config *gameserverConfig

	// battle is the game (mapentity with players)
	battle          *vbge.Battle
	updateDist      *updateDistributor
	envDisableCrypt bool
)

func gsInit() {
	val, exists := os.LookupEnv("VB_DISABLE_CRYPT")
	if exists && val == "1" {
		envDisableCrypt = true
	}

	log.Println("seeding global PRNG-source")
	noice, err := vbcore.CryptoGenBytes(1)
	if err != nil {
		log.Fatalf("failed getting noice value for seeding global PRNG-source: %s\n", err.Error())
	}
	if noice[0] == 0 {
		noice[0] = 1
	}
	rand.Seed(time.Now().UnixNano() / int64(noice[0]))

	// Init notification distributation network
	updateDist = &updateDistributor{
		History: queue.New(),
	}

	// Logging server
	priority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.DebugLevel
	})
	console := zapcore.Lock(os.Stdout)
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	logCore := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, console, priority),
	)
	logctx = zap.New(logCore)

	log.Println("init database connections")
	vbdbConfig := &vbdb.Config{
		DbAddr: vbcore.NewEndpointAddr(config.Database.MariaDB.Host),
		DbUser: config.Database.MariaDB.User,
		DbPass: config.Database.MariaDB.Password,
		DbName: config.Database.MariaDB.Name,
	}
	err = vbdb.Init(vbdbConfig, logctx)
	if err != nil {
		logctx.Fatal("init failed", zap.Error(err))
	}

	registryInit()
}

func battleInit() {
	battle = &vbge.Battle{
		// MapSize
		Map:     vbge.NewMapEntity(vbge.MapHeight, vbge.MapWidth),
		Players: make(map[int]*vbge.Player),
	}
	// MapSize
	vbge.SetMapDimensions(vbge.MapHeight, vbge.MapWidth)

	joined, success := vbdb.JoinedUsersCtx(config.RoundID, logctx)
	if !success {
		logctx.Fatal("unable to load users for this round", zap.Int("round_id", config.RoundID))
	}

	for _, j := range joined {
		p, err := vbge.NewPlayerWithSpawn(j, battle.Map)
		if err != nil {
			logctx.Fatal("failed to init vbge/(*Player) struct", zap.Error(err))
		}
		battle.Players[j] = p
	}
}

func main() {
	log.Println("defining flags")
	conf := flag.String("config", "", "")

	log.Println("parsing flags")
	flag.Parse()

	if conf == nil || *conf == "" {
		log.Fatal("no gameserver config defined")
	}
	config = loadConfig(*conf)

	// Prepare basic stuff of the server and init our battle (fetch map)
	gsInit()
	battleInit()

	// Start and shutdown channels
	startChan := make(chan bool)
	shutdownChan := make(chan bool)

	// Start the network services
	ntcpInit(startChan, shutdownChan)
	nwsInit(startChan, shutdownChan)

	// Start the log writter
	err := updateDist.InitHistoryWorker(startChan, shutdownChan)
	if err != nil {
		logctx.Fatal("unable to init history log writter. aborting ...", zap.Error(err))
	}

	// Sleep till start
	startTime := time.Now().UTC().Add(time.Second * 2)
	sleepDuration := startTime.Sub(time.Now().UTC())
	logctx.Info("prepared services. sleeping till starttime",
		zap.Time("starttime", startTime),
		zap.Duration("sleeping", sleepDuration))
	time.Sleep(sleepDuration)

	// Activate services that listen on starting channel signal
	startChan <- true
	startChan <- true
	startChan <- true

	// Shutdown services in on hour
	logctx.Info("started services. sleeping till shutdown",
		zap.Time("shutdowntim", time.Now().UTC().Add(time.Hour*1)))
	time.Sleep(time.Hour * 1) // Time of a game
	shutdownChan <- true
}
