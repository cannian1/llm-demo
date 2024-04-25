package main

import (
	"github.com/gin-gonic/gin"
	"llm_demo/api"
)

func main() {
	r := gin.Default()
	v1 := r.Group("/api/v1")

	v1.POST("translate", api.Translator)           // 翻译
	v1.POST("generate", api.GenerateResponse)      // 用 prompt 生成回答
	v1.POST("stream_response", api.StreamResponse) // 流式响应结合SSE
	r.Run(":8080")
}
