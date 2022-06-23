package main

import (
	"context"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	_ "github.com/go-sql-driver/mysql"
	"github.com/oklog/run"
	"github.com/prometheus/common/promlog"
	promlogflag "github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"math/rand"
	"openDevops/src/models"
	"openDevops/src/modules/server/cloud_sync"
	"openDevops/src/modules/server/config"
	mem_index "openDevops/src/modules/server/mem-index"
	"openDevops/src/modules/server/rpc"
	"openDevops/src/modules/server/web/route"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var (
	/*
	https://pkg.go.dev/gopkg.in/alecthomas/kingpin.v2
	https://blog.csdn.net/yang_kaiyue/article/details/122620794

		kingpin.Arg(): 创建固定参数(按顺序传入, 不需要 --flag 指定)
		kingpin.Flag(): 创建可选参数

		kingpin.Parse(): 用法同 flag
		kingpin.MustParse(): Parse() 底层调用的它
	*/
	// 命令行解析
	app = kingpin.New(filepath.Base(os.Args[0]), "The Open-Devops-Server")
	// 指定配置文件, 启动文件的时候可以 -c 传入
	configFile = app.Flag("config.file", "open-devops-server configuration file path").Short('c').Default("open-devops-server.yml").String()
)

var logger log.Logger

// go build -o server.exe  -ldflags "-X 'github.com/prometheus/common/version.BuildUser=root@n9e'  -X 'github.com/prometheus/common/version.BuildDate=\`date\`' -X 'github.com/prometheus/common/version.Version=\`cat VERSION\`'" src/modules/server/server.go
func main() {
	// 版本信息
	app.Version(version.Print("open-devops-server"))
	// 帮助信息
	app.HelpFlag.Short('h')

	promlogConfig := promlog.Config{}

	promlogflag.AddFlags(app, &promlogConfig)
	// 强制解析
	kingpin.MustParse(app.Parse(os.Args[1:]))


	/*
		加入参数以json格式输出日志信息 --log.format=json

	 	go run src/modules/server/server.go --log.format=json
	*/
	logger = func(config *promlog.Config) log.Logger {
		var (
			l  log.Logger
			le level.Option
		)
		if config.Format.String() == "logfmt" {
			l = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
		} else {
			l = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
		}

		switch config.Level.String() {
		case "debug":
			le = level.AllowDebug()
		case "info":
			le = level.AllowInfo()
		case "warn":
			le = level.AllowWarn()
		case "error":
			le = level.AllowError()
		}
		l = level.NewFilter(l, le)
		l = log.With(l, "ts", log.TimestampFormat(
			func() time.Time { return time.Now().Local() },
			"2006-01-02T15:04:05.000Z07:00",
		), "caller", log.DefaultCaller)
		return l
	}(&promlogConfig)
	level.Info(logger).Log("msg", "using config.file", "file.path", *configFile)

	// 初始化配置文件
	sConfig, err := config.LoadFile(*configFile)
	if err != nil {
		level.Error(logger).Log("msg", "config.LoadFile Error,Exiting ...", "error", err)
		return
	}
	level.Info(logger).Log("msg", "load.config.success", "file.path", *configFile, "content.mysql.num", len(sConfig.MysqlS))

	rand.Seed(time.Now().UnixNano())

	// 初始化数据库
	models.InitMySQL(sConfig.MysqlS)
	level.Info(logger).Log("msg", "load.mysql.success", "db.num", len(models.DB))

	// 多个goroutine 协作，共同进退(prometheus 采用的)
	// 编排开始
	var g run.Group
	ctxAll, cancelAll := context.WithCancel(context.Background())
	fmt.Println(ctxAll, "this is server") // context.Background.WithCancel


	// 处理信号退出的handler
	{
		term := make(chan os.Signal, 1)
		signal.Notify(term, os.Interrupt, syscall.SIGTERM)
		cancelC := make(chan struct{})
		g.Add(
			func() error {
				select {
				case <-term:
					level.Warn(logger).Log("msg", "Receive SIGTERM ,exiting gracefully....")
					cancelAll()
					return nil
				case <-cancelC:
					level.Warn(logger).Log("msg", "other cancel exiting")
					return nil
				}
			},
			func(err error) {
				close(cancelC)
			},
		)
	}

	// rpc server
	{
		g.Add(func() error {
			errChan := make(chan error, 1)
			go func() {
				errChan <- rpc.Start(sConfig.RpcAddr, logger)
			}()
			select {
			case err = <-errChan:
				level.Error(logger).Log("msg", "rpc server error", "err", err)
				return err
			case <-ctxAll.Done():
				level.Info(logger).Log("msg", "receive_quit_signal_rpc_server_exit")
				return nil
			}

		}, func(err error) {
			cancelAll()
		},
		)
	}

	// http server
	{
		g.Add(func() error {
			errChan := make(chan error, 1)
			go func() {
				errChan <- route.StartGin(sConfig.HttpAddr, logger)
			}()
			select {
			case err = <-errChan:
				level.Error(logger).Log("msg", "web server error", "err", err)
				return err
			case <-ctxAll.Done():
				level.Info(logger).Log("msg", "receive_quit_signal_web_server_exit")
				return nil
			}
		}, func(err error) { cancelAll()},
		)
	}

	// 公有云同步
	{
		// 开启公有云才执行
		if sConfig.PCC.Enable {
			cloud_sync.Init(logger)
			g.Add(func() error {
				err = cloud_sync.CloudSyncManager(ctxAll, logger)
				if err != nil {
					level.Error(logger).Log("msg", "cloudsync.CloudSyncManager.error", "err", err)

				}
				return err

			}, func(err error) {cancelAll()},
			)
		}
	}

	// 刷新倒排索引
	{
		// 初始化倒排索引模块
		mem_index.Init(logger, sConfig.IndexModules)
		level.Info(logger).Log("msg", "load.inverted-index.success")
		g.Add(func() error {
			err = mem_index.RevertedIndexSyncManager(ctxAll, logger)
			if err != nil {
				level.Error(logger).Log("msg", "mem_index.RevertedIndexSyncManager.error", "err", err)
			}
			return err
		}, func(err error) { cancelAll() },
		)
	}

	g.Run()
}