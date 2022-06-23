package cloud_sync

import (
	"github.com/go-kit/log"
	"openDevops/src/common"
	"github.com/go-kit/log/level"
	"sync"
	"time"
	"context"
)

type CloudResource interface {
	sync()
}

// 预留多种云类型
type CloudAlibaba struct {
}
type CloudTencent struct {
}
//type CloudHuWei struct {
//}

// 接口容器
var (
	cloudResourceContainer = make(map[string]CloudResource)
)

// 资源注册
func cRegister(name string, cr CloudResource) {
	cloudResourceContainer[name] = cr
}

// Todo 这里模拟一种公有云的操作
func Init(logger log.Logger) {
	hs := &HostSync{
		TableName: common.RESOURCE_HOST,
		Logger:    logger,
	}
	cRegister(common.RESOURCE_HOST, hs)
}


// --------------------------------------------------------------------------------------------------------------
func CloudSyncManager(ctx context.Context, logger log.Logger) error {
	level.Info(logger).Log("msg", "CloudSyncManager.start", "resource_num", len(cloudResourceContainer))
	ticker := time.NewTicker(5 * time.Second)
	doCloudSync(logger)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			level.Info(logger).Log("msg", "CloudSyncManager.exit.receive_quit_signal", "resource_num", len(cloudResourceContainer))
			return nil
		case <-ticker.C:
			level.Info(logger).Log("msg", "doCloudSync.cron", "resource_num", len(cloudResourceContainer))
			doCloudSync(logger)
		}
	}

}

func doCloudSync(logger log.Logger) {
	var wg sync.WaitGroup
	wg.Add(len(cloudResourceContainer))
	for _, sy := range cloudResourceContainer {
		sy := sy
		go func() {
			defer wg.Done()
			sy.sync()
		}()
	}
	wg.Wait()
}