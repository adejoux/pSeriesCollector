package main

import (
	"flag"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"

	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/adejoux/pSeriesCollector/pkg/agent"
	"github.com/adejoux/pSeriesCollector/pkg/agent/bus"
	"github.com/adejoux/pSeriesCollector/pkg/agent/devices/hmc"
	"github.com/adejoux/pSeriesCollector/pkg/agent/output"
	"github.com/adejoux/pSeriesCollector/pkg/agent/selfmon"
	"github.com/adejoux/pSeriesCollector/pkg/config"
	"github.com/adejoux/pSeriesCollector/pkg/data/hmcpcm"
	"github.com/adejoux/pSeriesCollector/pkg/data/impexp"
	"github.com/adejoux/pSeriesCollector/pkg/webui"
)

var (
	log        = logrus.New()
	quit       = make(chan struct{})
	startTime  = time.Now()
	getversion bool
	httpPort   = 8080
	appdir     = os.Getenv("PWD")
	homeDir    string
	pidFile    string
	logDir     = filepath.Join(appdir, "log")
	confDir    = filepath.Join(appdir, "conf")
	dataDir    = confDir
	configFile = filepath.Join(confDir, "pseriescollector.toml")
)

func writePIDFile() {
	if pidFile == "" {
		return
	}

	// Ensure the required directory structure exists.
	err := os.MkdirAll(filepath.Dir(pidFile), 0700)
	if err != nil {
		log.Fatal(3, "Failed to verify pid directory", err)
	}

	// Retrieve the PID and write it.
	pid := strconv.Itoa(os.Getpid())
	if err := ioutil.WriteFile(pidFile, []byte(pid), 0644); err != nil {
		log.Fatal(3, "Failed to write pidfile", err)
	}
}

func flags() *flag.FlagSet {
	var f flag.FlagSet
	f.BoolVar(&getversion, "version", getversion, "display de version")
	f.StringVar(&configFile, "config", configFile, "config file")
	f.IntVar(&httpPort, "http", httpPort, "http port")
	f.StringVar(&logDir, "logs", logDir, "log directory")
	f.StringVar(&homeDir, "home", homeDir, "home directory")
	f.StringVar(&dataDir, "data", dataDir, "Data directory")
	f.StringVar(&pidFile, "pidfile", pidFile, "path to pid file")
	f.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		f.VisitAll(func(flag *flag.Flag) {
			format := "%10s: %s\n"
			fmt.Fprintf(os.Stderr, format, "-"+flag.Name, flag.Usage)
		})
		fmt.Fprintf(os.Stderr, "\nAll settings can be set in config file: %s\n", configFile)
		os.Exit(1)

	}
	return &f
}

func init() {
	//Log format
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.Formatter = customFormatter
	customFormatter.FullTimestamp = true

	// parse first time to see if config file is being specified
	f := flags()
	f.Parse(os.Args[1:])

	if getversion {
		t, _ := strconv.ParseInt(agent.BuildStamp, 10, 64)
		fmt.Printf("pseriescollector v%s (git: %s ) built at [%s]\n", agent.Version, agent.Commit, time.Unix(t, 0).Format("2006-01-02 15:04:05"))
		os.Exit(0)
	}

	// now load up config settings
	if _, err := os.Stat(configFile); err == nil {
		viper.SetConfigFile(configFile)
		confDir = filepath.Dir(configFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("/etc/pseriescollector/")
		viper.AddConfigPath("/opt/pseriescollector/conf/")
		viper.AddConfigPath("./conf/")
		viper.AddConfigPath(".")
	}
	err := viper.ReadInConfig()
	if err != nil {
		log.Errorf("Fatal error config file: %s \n", err)
		os.Exit(1)
	}
	err = viper.Unmarshal(&agent.MainConfig)
	if err != nil {
		log.Errorf("Fatal error config file: %s \n", err)
		os.Exit(1)
	}
	cfg := &agent.MainConfig

	if len(cfg.General.LogDir) > 0 {
		logDir = cfg.General.LogDir
		os.Mkdir(logDir, 0755)
		//Log output
		f, _ := os.OpenFile(logDir+"/pseriescollector.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		log.Out = f
	}
	if len(cfg.General.LogLevel) > 0 {
		l, _ := logrus.ParseLevel(cfg.General.LogLevel)
		log.Level = l
	}
	if len(cfg.General.DataDir) > 0 {
		dataDir = cfg.General.DataDir
	}
	if len(cfg.General.HomeDir) > 0 {
		homeDir = cfg.General.HomeDir
	}
	//check if exist public dir in home
	if _, err := os.Stat(filepath.Join(homeDir, "public")); err != nil {
		log.Warnf("There is no public (www) directory on [%s] directory", homeDir)
		if len(homeDir) == 0 {
			homeDir = appdir
		}
	}

	log.Debug("AGENT MAINCONFIG LOAD  %#+v", cfg)
	//needed to create SQLDB when SQLite and debug log
	config.SetLogger(log)
	config.SetDirs(dataDir, logDir, confDir)
	output.SetLogger(log)
	selfmon.SetLogger(log)

	hmc.SetDBConfig(&agent.DBConfig)
	hmc.SetLogDir(logDir)

	webui.SetLogger(log)
	webui.SetLogDir(logDir)
	webui.SetConfDir(confDir)
	agent.SetLogger(log)

	impexp.SetLogger(log)
	bus.SetLogger(log)
	hmcpcm.SetLogger(log)

	//
	log.Infof("Set Default directories : \n   - Exec: %s\n   - Config: %s\n   -Logs: %s\n -Home: %s\n", appdir, confDir, logDir, homeDir)
}

func main() {

	defer func() {
		//errorLog.Close()
	}()
	writePIDFile()
	//Init BD config

	agent.MainConfig.Database.InitDB()
	impexp.SetDB(&agent.MainConfig.Database)

	agent.LoadConf()

	agent.DeviceProcessStart()

	webui.WebServer(filepath.Join(homeDir, "public"), httpPort, &agent.MainConfig.HTTP, agent.MainConfig.General.InstanceID)

}
