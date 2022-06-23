package mem_index

import (
	"encoding/json"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	ii "github.com/ning1875/inverted-index"
	"github.com/ning1875/inverted-index/labels"
	"openDevops/src/common"
	"openDevops/src/models"
	"strconv"
	"strings"
	"time"
)

type HostIndex struct {
	Ir      *ii.HeadIndexReader
	Logger  log.Logger
	Modulus int // 静态分配的模(在海量数据处理的时候起到一个分片的效果)
	Num     int
}

/*
- 需要注意配置了分片的逻辑，主要的目的是解决数据量大一个实例索引撑不住或慢的问题
  - 先获取总数，即count
  - 模>0 ,分片>0
  - 取模后相等则keep
  - 拼接出属于这个分片应该 in 的ids
  - 没有分片则全量ids
- 然后根据拼接出来的ids查询数据库数据
- 遍历字段，刷索引即可
- 最后的hi.Ir.Reset(actuallyH) 代表索引全量更新，也就是renew map
*/
func (hi *HostIndex) FlushIndex() {
	// 数个数
	start := time.Now()
	r := new(models.ResourceHost)
	total := int(r.Count())


	ids := ""
	//for i := 0; i < total; i++ {
	for i := 1; i < total+1; i++ {
		// 先写单点逻辑
		if hi.Modulus == 0 {
			ids += fmt.Sprintf("%d,", i)
			continue
		}
		// 分片匹配中了 ，keep的逻辑
		if i%hi.Modulus == hi.Num {
			ids += fmt.Sprintf("%d,", i)
			continue
		}
	}


	ids = strings.TrimRight(ids, ",")
	inSql := fmt.Sprintf("id in (%s) ", ids)
	objs, err := models.ResourceHostGetMany(inSql)
	if err != nil {
		return
	}

	// 自动刷node path(stree_path 表)
	thisGPAS := map[string]struct{}{}

	thisH := ii.NewHeadReader()
	for _, item := range objs {
		m := make(map[string]string)
		m["hash"] = item.Hash
		m["uid"] = item.Uid
		m["name"] = item.Name
		m["cloud_provider"] = item.CloudProvider
		m["charging_mode"] = item.ChargingMode
		m["region"] = item.Region
		m["instance_type"] = item.InstanceType
		m["availability_zone"] = item.AvailabilityZone
		m["vpc_id"] = item.VpcId
		m["subnet_id"] = item.SubnetId
		m["status"] = item.Status
		m["account_id"] = strconv.FormatInt(item.AccountId, 10)
		// cpu mem
		m["cpu"] = item.CPU
		m["mem"] = item.Mem
		m["disk"] = item.Disk
		// g.p.a
		m["stree_group"] = item.StreeGroup
		m["stree_product"] = item.StreeProduct
		m["stree_app"] = item.StreeApp
		thisGPAS[fmt.Sprintf("%s.%s.%s", item.StreeGroup, item.StreeProduct, item.StreeApp)] = struct{}{}

		/*
			rand.Seed(time.Now().Unix())
			ips := []string{fmt.Sprintf("8.8.8.%d", rand.Int63n(10))}
			ipJ, _ := json.Marshal(ips)

			prIps := []string{}
			json.Unmarshal([]byte("[\"8.8.8.0\"]"), &prIps)
			fmt.Println(prIps, prIps[0])  // [8.8.8.0] 8.8.8.0
		*/
		// 数组型 内网ips 公网ips 安全组
		prIps := []string{}
		puIps := []string{}
		// json列表型
		json.Unmarshal([]byte(item.PrivateIps), &prIps)
		json.Unmarshal([]byte(item.PublicIps), &puIps)

		// json map型
		tags := make(map[string]string)
		json.Unmarshal([]byte(item.Tags), &tags)

		// 调用倒排索引库刷新索引
		// mapTolsets: map[string]string -> []struct{ Name, Value string}
		thisH.GetOrCreateWithID(uint64(item.Id), item.Hash, mapTolsets(m))
		thisH.GetOrCreateWithID(uint64(item.Id), item.Hash, mapTolsets(tags))

		// 数组型
		for _, i := range prIps {
			mp := map[string]string{
				"private_ip": i,
			}
			thisH.GetOrCreateWithID(uint64(item.Id), item.Hash, mapTolsets(mp))
		}

		for _, i := range puIps {
			mp := map[string]string{
				"private_ip": i,
			}
			thisH.GetOrCreateWithID(uint64(item.Id), item.Hash, mapTolsets(mp))
		}
		for _, i := range prIps {
			mp := map[string]string{
				"public_ip": i,
			}
			thisH.GetOrCreateWithID(uint64(item.Id), item.Hash, mapTolsets(mp))
		}
	}

	hi.Ir.Reset(thisH)
	level.Debug(hi.Logger).Log("msg", "FlushIndex.time.took","took", time.Since(start).Seconds())

	// 自动的将g.p.a 添加到node_path
	go func() {
		level.Info(hi.Logger).Log("msg", "FlushIndex.Add.GPA.To.PATH",
			"num", len(thisGPAS),
		)
		for node := range thisGPAS {
			inputs := common.NodeCommonReq{
				Node: node,
			}
			models.StreePathAddOne(&inputs, hi.Logger)
		}
	}()
}

func mapTolsets(m map[string]string) labels.Labels {
	// type Labels []Label
	// type Label struct {
	//	Name, Value string
	// }
	var lset labels.Labels
	for k, v := range m {
		l := labels.Label{
			Name:  k,
			Value: v,
		}
		lset = append(lset, l)
	}
	return lset
}

func (hi *HostIndex) GetIndexReader() *ii.HeadIndexReader {
	return hi.Ir
}

func (hi *HostIndex) GetLogger() log.Logger {
	return hi.Logger
}