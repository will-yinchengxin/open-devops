package rpc

import (
	"fmt"
)

// Todo 初始化, 测试使用, 无实际意义
func (*Server) Ping(input string, output *string) error {
	fmt.Println(input)
	*output = "get it"
	return nil
}

