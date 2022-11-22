package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/sirupsen/logrus"
)

var helpMsg string = `uamq is a utility to decode and encode UART messages to MQTT messages and vice versa

Usage:
	uamq -h           Displays this help message
	uamq --help       Displays this help message
	uamq config.yaml  Starts the uamq utility`

var k *koanf.Koanf
var logFile *os.File

var uartKill = make(chan bool, 1)
var mqttKill = make(chan bool, 1)

type MqttConf struct {
	host string
	port int
}

type RedisConf struct {
	host string
	port int
}

type UartConf struct {
	name     string
	baudrate int
}

type LogConf struct {
	target string
}

type Config struct {
	mqtt  MqttConf
	redis RedisConf
	uart  UartConf
	log   LogConf
}

type Cli struct {
	help       string
	wg         sync.WaitGroup
	configFile string
	conf       *Config
}

func app() *Cli {
	app := new(Cli)
	app.help = helpMsg

	arg := os.Args[1:]
	if len(arg) == 0 || len(arg) > 1 {
		fmt.Println(app.help)
		os.Exit(1)
	}
	if arg[0] == "" || arg[0] == "-h" || arg[0] == "--help" {
		fmt.Println(app.help)
		os.Exit(0)
	}
	app.configFile = arg[0]
	return app
}

func (c *Cli) loadConfig() {
	conf := new(Config)
	k = koanf.New(".")
	f := file.Provider(c.configFile)
	err := k.Load(f, yaml.Parser())
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	conf.mqtt.host = k.String("mqtt.host")
	conf.mqtt.port = k.Int("mqtt.port")

	conf.redis.host = k.String("redis.host")
	conf.redis.port = k.Int("redis.port")

	conf.uart.name = k.String("uart.name")
	conf.uart.baudrate = k.Int("uart.baudrate")

	conf.log.target = k.String("log.target")

	c.conf = conf

	f.Watch(func(event interface{}, err error) {
		if err != nil {
			Log.WithError(err).Error("watch error")
			return
		}

		// Throw away the old config and load a fresh copy.
		Log.Info("config changed. Reloading ...")
		c.refresh()
		// k = koanf.New(".")
		// k.Load(f, yaml.Parser())
		// k.Print()
	})
}

func (c *Cli) start() {
	c.loadConfig()
	logPath := c.conf.log.target
	err := makeLogDir(filepath.Base(logPath))
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	logFile, err = os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	InitLogger()
	defer logFile.Close()

	Log.Info("uamq started")

	RedisInit()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		uartHandler(c.conf.uart.name, c.conf.uart.baudrate)
	}()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		mqttHandler(c.conf.mqtt.host, c.conf.mqtt.port)
	}()
	c.wg.Wait()
}

func (c *Cli) refresh() {
	Log.Info("Refreshing")
	uartKill <- true
	mqttKill <- true
	c.start()
}

var once sync.Once

var Log *logrus.Logger

func InitLogger() {

	once.Do(func() {
		Log = logrus.New()
		Log.SetFormatter(&logrus.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
			ForceQuote:    true,
		})
		Log.SetReportCaller(true)

		mw := io.MultiWriter(os.Stdout, logFile)
		Log.SetOutput(mw)
	})

}

func makeLogDir(dir string) error {
	err := os.MkdirAll(dir, 0755)
	if err == nil {
		return nil
	}
	if os.IsExist(err) {
		// check that the existing path is a directory
		info, err := os.Stat(dir)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return errors.New("path exists but is not a directory")
		}
		return nil
	}
	return err
}
