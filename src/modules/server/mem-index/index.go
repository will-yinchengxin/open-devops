package mem_index

import (
	"context"
	"github.com/go-kit/log"
	ii "github.com/ning1875/inverted-index"
	"github.com/ning1875/inverted-index/index"
	"openDevops/src/common"
	"github.com/go-kit/log/level"
	"openDevops/src/modules/server/config"
	"strings"
	"sync"
	"time"
)

type ResourceIndexer interface {
	FlushIndex()                         // 刷新索引的方法
	GetIndexReader() *ii.HeadIndexReader // 获取内置的索引reader
	GetLogger() log.Logger
}

var indexContainer = make(map[string]ResourceIndexer)

func iRegister(name string, ri ResourceIndexer) {
	indexContainer[name] = ri
}

func JudgeResourceIndexExists(name string) bool {
	_, ok := indexContainer[name]
	return ok
}

func Init(logger log.Logger, ims []*config.IndexModuleConf) {
	loadNum := 0
	loadResource := make([]string, 0)
	for _, i := range ims {
		if !i.Enable {

			continue
		}
		level.Info(logger).Log("msg", "mem-index.init", "name", i.ResourceName)
		loadNum += 1
		loadResource = append(loadResource, i.ResourceName)
		switch i.ResourceName {
		case common.RESOURCE_HOST:
			mi := &HostIndex{
				Ir:      ii.NewHeadReader(),
				Logger:  logger,
				Modulus: i.Modulus,
				Num:     i.Num,
			}
			iRegister(i.ResourceName, mi)
		case common.RESOURCE_RDS: // 后续可信根据不同的 共有类型执行不同的对应操作
			mi := &HostIndex{
				Ir:      ii.NewHeadReader(),
				Logger:  logger,
				Modulus: i.Modulus,
				Num:     i.Num,
			}
			iRegister(i.ResourceName, mi)
		}
	}
	level.Info(logger).Log("msg", "mem-index.init.summary", "loadNum", loadNum, "detail", strings.Join(loadResource, " "))
}

func GetResourceIndexReader(name string) (bool, ResourceIndexer) {
	ri, ok := indexContainer[name]
	return ok, ri

}

func GetAllResourceIndexReader() (make map[string]ResourceIndexer) {
	return indexContainer

}

// matchid入口函数
func GetMatchIdsByIndex(req common.ResourceQueryReq) (matchIds []uint64) {
	ri, ok := indexContainer[req.ResourceType]
	if !ok {
		return
	}
	matcher := common.FormatLabelMatcher(req.Labels)

	p, err := ii.PostingsForMatchers(ri.GetIndexReader(), matcher...)
	if err != nil {
		level.Error(ri.GetLogger()).Log("msg", "ii.PostingsForMatchers.error", "ResourceType", req.ResourceType, "err", err)
		return
	}
	matchIds, err = index.ExpandPostings(p)
	if err != nil {
		level.Error(ri.GetLogger()).Log("msg", "index.ExpandPostings.error", "ResourceType", req.ResourceType, "err", err)
		return
	}
	return
}

func RevertedIndexSyncManager(ctx context.Context, logger log.Logger) error {
	level.Info(logger).Log("msg", "RevertedIndexSyncManager.start", "resource_num", len(indexContainer))
	ticker := time.NewTicker(15 * time.Second)
	doIndexFlush()
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			level.Info(logger).Log("msg", "RevertedIndexSyncManager.exit.receive_quit_signal", "resource_num", len(indexContainer))
			return nil
		case <-ticker.C:
			level.Info(logger).Log("msg", "doIndexFlush.cron", "resource_num", len(indexContainer))
			doIndexFlush()
		}
	}
}

func doIndexFlush() {
	var wg sync.WaitGroup
	wg.Add(len(indexContainer))
	for _, ir := range indexContainer {
		ir := ir
		go func() {
			defer wg.Done()
			ir.FlushIndex()
		}()
	}
	wg.Wait()
}