package main

import (
	"github.com/gin-gonic/gin"
	"llm_demo/api"
)

func main() {
	r := gin.Default()
	v1 := r.Group("/api/v1")

	v1.POST("translate", api.Translator)
	v1.POST("generate", api.GenerateResponse)
	r.Run(":8080")
}
