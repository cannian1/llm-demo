package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/prompts"
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
