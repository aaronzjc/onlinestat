package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aaronzjc/onlinestat/internal"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/cli"
)

var (
	appName = "online-stat"
	usage   = "run online-stat server"
	desc    = `online-stat is a online user counter service`
	version = "1.0"
)

func main() {
	app := *cli.NewApp()
	app.Name = appName
	app.Usage = usage
	app.Description = desc
	app.Version = version
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "config,c",
			Usage: "(config) Load configuration from `FILE`",
		},
	}
	app.Before = setup
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func setup(ctx *cli.Context) error {
	var err error
	if ctx.String("config") == "" {
		return errors.New("invalid config option, use -h get full doc")
	}
	// 初始化项目配置
	configFile := ctx.String("config")
	if _, err = os.Stat(configFile); os.IsNotExist(err) {
		return errors.New("config file not found")
	}
	if err = internal.LoadConfig(configFile); err != nil {
		return err
	}

	// 初始化统计后端
	if err = internal.SetupStater(internal.GetConfig()); err != nil {
		return err
	}
	return nil
}

func run(ctx *cli.Context) error {
	conf := internal.GetConfig()
	var addr string
	if conf.Env != "prod" {
		addr = fmt.Sprintf("127.0.0.1:%d", conf.Http.Port)
	} else {
		addr = fmt.Sprintf(":%d", conf.Http.Port)
	}

	router := httprouter.New()
	internal.RegistRoutes(router)

	// 启动服务器
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 10,
	}
	go server.ListenAndServe()
	log.Printf("[START] server listen at %s", addr)

	// 监听关闭信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGQUIT, os.Interrupt, syscall.SIGTERM)
	<-sig

	// 收到关闭信号，主动回收连接
	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err := server.Shutdown(ctxTimeout); err != nil {
		log.Printf("[STOP] server shutdown error %v", err)
		return err
	}
	log.Printf("[STOP] server shutdown ok")
	return nil
}
