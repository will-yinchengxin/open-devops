package common

// 请求参数, 可以使用validate 组件进行参数校验

// 操作树结构的通用对象
// 新增 删除 修改
type NodeCommonReq struct {
	Node        string `json:"node"`         // 服务节点名称 ：可以一段式 也可以是两段式 inf inf.mon
	QueryType   int    `json:"query_type"`   // 查询模式 1，2，3
	ForceDelete bool   `json:"force_delete"` //子节点强制删除
}
