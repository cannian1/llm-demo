package api

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/prompts"
	"log"
	"net/http"
)

func Translator(c *gin.Context) {

	var reqData struct {
		OutputLang string `json:"outputLang" binding:"required"`
		Text       string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON " + err.Error()})
		return
	}

	// 创建 prompt
	prompt := prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
		prompts.NewSystemMessagePromptTemplate("你是一个只能翻译文本的翻译引擎，不需要进行解释。", nil),
		prompts.NewHumanMessagePromptTemplate(`翻译这段文本到{{.outputLang}}: {{.text}}`,
			[]string{"outputLang", "text"}),
	})

	// 填充 prompt
	values := map[string]any{
		"outputLang": reqData.OutputLang,
		"text":       reqData.Text,
	}

	messages, err := prompt.FormatMessages(values)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to format prompt" + err.Error()})
		return
	}

	// 连接到 ollama，指定使用的模型和服务器地址(这个改成单例更好), 如果目标机器的内存不够，可以考虑使用小一些的模型
	llm, err := ollama.New(ollama.WithModel("qwen:72b"), ollama.WithServerURL("http://10.101.5.11:11434"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load model"})
	}

	content := []llms.MessageContent{
		llms.TextParts(messages[0].GetType(), messages[0].GetContent()),
		llms.TextParts(messages[1].GetType(), messages[1].GetContent()),
	}

	res, err := llm.GenerateContent(context.Background(), content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call llm"})
	}
	c.JSON(http.StatusOK, gin.H{"response": res.Choices[0].Content})
}

func GenerateResponse(c *gin.Context) {
	var reqData struct {
		Prompt string `json:"prompt" binding:"required"`
	}

	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON " + err.Error()})
		return
	}
	llm, err := ollama.New(ollama.WithModel("llama3:70b"), ollama.WithServerURL("http://10.101.5.11:11434"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load model"})
	}

	res, err := llm.Call(context.Background(), reqData.Prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call llm"})
	}

	c.JSON(http.StatusOK, gin.H{"response": res})
}

func StreamResponse(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	llm, err := ollama.New(ollama.WithModel("qwen:72b"), ollama.WithServerURL("http://10.101.5.11:11434"))
	if err != nil {
		log.Println("Failed to load model")
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", "Failed to load model")
		c.Writer.Flush()
	}

	content, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, "夜晚的水母为什么不会游泳"),
		},
		// 流式响应可以让用户更快地获取到回答的部分内容
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			for _, line := range bytes.Split(chunk, []byte("\n")) {
				if len(line) > 0 {
					fmt.Fprintf(c.Writer, "data: %s\n\n", string(line))
					fmt.Println(string(line))
					c.Writer.Flush()
				}
			}
			return nil
		}),
	)
	if err != nil {
		// 发送错误事件到客户端
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", "Failed to call llm")
		c.Writer.Flush()
	}
	if content.Choices[0].FuncCall != nil {
		fmt.Fprintf(c.Writer, "data: Function call :%v\n\n", content.Choices[0].FuncCall)
		c.Writer.Flush()
	}
}
