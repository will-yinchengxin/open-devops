package test

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"openDevops/src/common"
	"openDevops/src/models"
	"strings"
)

func StreePathAddTest(logger log.Logger) {
	ns := []string{
		"inf.monitor.thanos",
		"inf.monitor.kafka",
		"inf.monitor.prometheus",
		"inf.monitor.m3db",
		"inf.cicd.gray",
		"inf.cicd.deploy",
		"inf.cicd.jenkins",
		"waimai.qiangdan.queue",
		"waimai.qiangdan.worker",
		"waimai.ditu.kafka",
		"waimai.ditu.es",
		"waimai.qiangdan.es",
	}
	for _, n := range ns {
		req := &common.NodeCommonReq{
			Node: n,
		}
		models.StreePathAddOne(req, logger)
	}
}

// 编写查询node的测试函数
func StreePathQueryTest1(logger log.Logger) {
	ns := []string{
		"a",
		"b",
		"c",
		"inf",
		"waimai",
	}
	for _, n := range ns {
		req := &common.NodeCommonReq{
			Node:      n,
			QueryType: 1,
		}
		res := models.StreePathQuery(req, logger)
		level.Info(logger).Log("msg", "StreePathQuery.res", "req.node", n, "num", len(res), "details", strings.Join(res, ","))
	}
}
// 编写查询node的测试函数 type=2
func StreePathQueryTest2(logger log.Logger) {
	ns := []string{
		"a",
		"b",
		"c",
		"inf",
		"waimai",
	}
	for _, n := range ns {
		req := &common.NodeCommonReq{
			Node:      n,
			QueryType: 2,
		}
		res := models.StreePathQuery(req, logger)
		level.Info(logger).Log("msg", "StreePathQuery.res", "req.node", n, "num", len(res), "details", strings.Join(res, ","))
	}
}
// 编写查询node的测试函数 type=3
func StreePathQueryTest3(logger log.Logger) {
	ns := []string{
		"a.b",
		"b.a",
		"c.d",
		"inf.cicd",
		"inf.monitor",
		"waimai.ditu",
		"waimai.monitor",
		"waimai.qiangdan",
	}
	for _, n := range ns {
		req := &common.NodeCommonReq{
			Node:      n,
			QueryType: 3,
		}
		res := models.StreePathQuery(req, logger)
		level.Info(logger).Log("msg", "StreePathQuery.res", "req.node", n, "num", len(res), "details", strings.Join(res, ","))
	}
}

// 编写删除node的测试函数
func StreePathDelTest(logger log.Logger) {
	ns := []string{
		"inf.cicd.jenkins",
		"inf.cicd",
		"inf",
	}
	for _, n := range ns {
		req := &common.NodeCommonReq{
			Node: n,
		}
		res := models.StreePathDelete(req, logger)
		level.Info(logger).Log("msg", "StreePathDelete.res", "req.node", n, "del_num", res)
	}
}
