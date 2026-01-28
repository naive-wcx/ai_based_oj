package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// getIntQuery 获取 int 类型的查询参数
func getIntQuery(c *gin.Context, key string, defaultVal int) int {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}

// getUintQuery 获取 uint 类型的查询参数
func getUintQuery(c *gin.Context, key string) uint {
	val := c.Query(key)
	if val == "" {
		return 0
	}
	intVal, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return 0
	}
	return uint(intVal)
}

// getUintParam 获取 uint 类型的路径参数
func getUintParam(c *gin.Context, key string) uint {
	val := c.Param(key)
	if val == "" {
		return 0
	}
	intVal, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return 0
	}
	return uint(intVal)
}

// getIntFormValue 获取 int 类型的表单值
func getIntFormValue(c *gin.Context, key string, defaultVal int) int {
	val := c.PostForm(key)
	if val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}
