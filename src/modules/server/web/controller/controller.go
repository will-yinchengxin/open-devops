package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/log"
	"math"
	"openDevops/src/common"
	"openDevops/src/models"
	mem_index "openDevops/src/modules/server/mem-index"
	"strconv"
	"strings"
)

func ResourceQuery(c *gin.Context) {
	var inputs common.ResourceQueryReq
	if err := c.BindJSON(&inputs); err != nil {
		common.JSONR(c, 400, err)
		return
	}
	ok := mem_index.JudgeResourceIndexExists(inputs.ResourceType)
	if !ok {
		common.JSONR(c, 400, fmt.Errorf("ResourceType_not_exists:%v", inputs.ResourceType))
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "100"))
	if err != nil {
		common.JSONR(c, 400, fmt.Errorf("invalid_page_size"))
		return
	}
	currentPage, err := strconv.Atoi(c.DefaultQuery("current_page", "1"))
	if err != nil {
		common.JSONR(c, 400, fmt.Errorf("invalid current_page"))
		return
	}

	offset := 0
	limit := 0
	limit = pageSize
	if currentPage > 1 {
		offset = (currentPage - 1) * limit
	}

	//matchIds := []uint64{1,2,3}
	matchIds := mem_index.GetMatchIdsByIndex(inputs)

	totalCount := len(matchIds)
	logger := c.MustGet("logger").(log.Logger)

	pageCount := int(math.Ceil(float64(totalCount) / float64(limit)))
	resp := common.QueryResponse{
		Code:        200,
		CurrentPage: currentPage,
		PageSize:    pageSize,
		PageCount:   pageCount,
		TotalCount:  totalCount,
	}
	res, err := models.ResourceQuery(inputs.ResourceType, matchIds, logger, limit, offset)
	if err != nil {
		resp.Code = 500
		resp.Result = err
	}
	resp.Result = res
	common.JSONR(c, resp)
}


func ResourceMount(c *gin.Context) {
	var inputs common.ResourceMountReq
	if err := c.BindJSON(&inputs); err != nil {
		common.JSONR(c, 400, err)
		return
	}
	logger := c.MustGet("logger").(log.Logger)

	// 校验 资源的名
	//ok := models.CheckResources(inputs.ResourceType)
	//if !ok {
	//	common.JSONR(c, 400, fmt.Errorf("resource_node_exist:%v", inputs.ResourceType))
	//	return
	//}

	// 校验g.p.a是否存在
	qReq := &common.NodeCommonReq{
		Node:      inputs.TargetPath,
		QueryType: 4,
	}

	gpa := models.StreePathQuery(qReq, logger)
	if len(gpa) == 0 {
		common.JSONR(c, 400, fmt.Errorf("target_path_not_exist:%v", inputs.TargetPath))
		return
	}

	// 绑定的动作
	rowsAff, err := models.ResourceMount(&inputs, logger)
	if err != nil {
		common.JSONR(c, 500, err)
		return
	}

	common.JSONR(c, 200, fmt.Sprintf("rowsAff:%d", rowsAff))
	return

}

func ResourceUnMount(c *gin.Context) {

	var inputs common.ResourceMountReq
	if err := c.BindJSON(&inputs); err != nil {
		common.JSONR(c, 400, err)
		return
	}
	logger := c.MustGet("logger").(log.Logger)

	// 校验 资源的名
	ok := models.CheckResources(inputs.ResourceType)
	if !ok {
		common.JSONR(c, 400, fmt.Errorf("resource_type_not_exist:%v", inputs.ResourceType))
		return
	}

	// 校验g.p.a是否存在
	qReq := &common.NodeCommonReq{
		Node:      inputs.TargetPath,
		QueryType: 4,
	}

	gpa := models.StreePathQuery(qReq, logger)
	if len(gpa) == 0 {
		common.JSONR(c, 400, fmt.Errorf("target_path_not_exist:%v", inputs.TargetPath))
		return
	}
	// 解绑的动作
	rowsAff, err := models.ResourceUnMount(&inputs, logger)
	if err != nil {
		common.JSONR(c, 500, err)
		return
	}

	common.JSONR(c, 200, fmt.Sprintf("rowsAff:%d", rowsAff))
	return
}

func NodePathAdd(c *gin.Context) {
	var inputs common.NodeCommonReq
	if err := c.Bind(&inputs); err != nil {
		common.JSONR(c, 400, err)
		return
	}
	logger := c.MustGet("logger").(log.Logger)

	res := strings.Split(inputs.Node, ".")
	if len(res) != 3 {
		common.JSONR(c, 400, fmt.Errorf("path_invalidate:%v", inputs.Node))
		return
	}
	err := models.StreePathAddOne(&inputs, logger)

	if err != nil {
		common.JSONR(c, 500, err)
		return
	}
	common.JSONR(c, 200, "path_add_success")
}

func NodePathQuery(c *gin.Context) {
	var inputs common.NodeCommonReq
	if err := c.BindJSON(&inputs); err != nil {
		common.JSONR(c, 400, err)
		return
	}
	logger := c.MustGet("logger").(log.Logger)

	if inputs.QueryType == 3 {
		if len(strings.Split(inputs.Node, ".")) != 2 {
			common.JSONR(c, 400, fmt.Errorf("query_type=3 path should be a.b:%v", inputs.Node))
			return
		}
	}
	res := models.StreePathQuery(&inputs, logger)
	common.JSONR(c, res)
}
