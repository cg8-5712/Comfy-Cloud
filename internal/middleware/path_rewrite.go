package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
)

// PathRewriteMiddleware 路径重写中间件（用户数据隔离）
func PathRewriteMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户 ID（从认证中间件注入）
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		uid := userID.(uint)
		path := c.Request.URL.Path

		// 1. 重写 GET 请求路径（读取文件）
		if c.Request.Method == "GET" {
			if newPath := rewriteGetPath(path, uid); newPath != "" {
				c.Request.URL.Path = newPath
			}
		}

		// 2. 重写 POST 请求中的路径（提交任务）
		if c.Request.Method == "POST" {
			if strings.Contains(path, "/prompt") || strings.Contains(path, "/upload") {
				rewritePostBody(c, uid)
			}
		}

		c.Next()
	}
}

// rewriteGetPath 重写 GET 请求路径
func rewriteGetPath(path string, userID uint) string {
	// /output/xxx → /users/{user_id}/output/xxx
	if strings.HasPrefix(path, "/output/") {
		return fmt.Sprintf("/users/%d/output/%s", userID, strings.TrimPrefix(path, "/output/"))
	}

	// /view → /users/{user_id}/output/xxx (ComfyUI 的图片查看接口)
	if strings.HasPrefix(path, "/view") {
		// 保持原路径，但在查询参数中添加用户前缀
		// 实际处理在 ComfyUI 端或通过 volume 映射
		return ""
	}

	// /workflows/xxx → /users/{user_id}/workflows/xxx
	if strings.HasPrefix(path, "/workflows/") {
		return fmt.Sprintf("/users/%d/workflows/%s", userID, strings.TrimPrefix(path, "/workflows/"))
	}

	// /models/xxx → 需要检查是否是私有模型
	if strings.HasPrefix(path, "/models/") {
		modelPath := strings.TrimPrefix(path, "/models/")
		// 如果是 user_xxx 开头，说明是私有模型
		if strings.HasPrefix(modelPath, fmt.Sprintf("user_%d/", userID)) {
			return fmt.Sprintf("/users/%d/models/%s", userID, strings.TrimPrefix(modelPath, fmt.Sprintf("user_%d/", userID)))
		}
		// 否则是共享模型，不重写
		return ""
	}

	return ""
}

// rewritePostBody 重写 POST 请求体中的路径
func rewritePostBody(c *gin.Context, userID uint) {
	// 读取原始请求体
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return
	}
	defer c.Request.Body.Close()

	// 尝试解析为 JSON
	var data map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		// 不是 JSON，恢复原始 body
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		return
	}

	// 重写路径字段
	modified := false

	// 处理 prompt 请求（ComfyUI workflow）
	if prompt, ok := data["prompt"].(map[string]interface{}); ok {
		if rewriteWorkflowPaths(prompt, userID) {
			modified = true
		}
	}

	// 处理 output_path 字段
	if outputPath, ok := data["output_path"].(string); ok {
		data["output_path"] = fmt.Sprintf("users/%d/output/%s", userID, outputPath)
		modified = true
	}

	// 处理 filename 字段
	if filename, ok := data["filename"].(string); ok {
		if !strings.HasPrefix(filename, fmt.Sprintf("users/%d/", userID)) {
			data["filename"] = fmt.Sprintf("users/%d/workflows/%s", userID, filename)
			modified = true
		}
	}

	// 如果修改了，重新序列化
	if modified {
		newBody, err := json.Marshal(data)
		if err == nil {
			c.Request.Body = io.NopCloser(bytes.NewBuffer(newBody))
			c.Request.ContentLength = int64(len(newBody))
		}
	} else {
		// 恢复原始 body
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}
}

// rewriteWorkflowPaths 重写 workflow JSON 中的路径
func rewriteWorkflowPaths(workflow map[string]interface{}, userID uint) bool {
	modified := false

	// 遍历 workflow 的所有节点
	for _, nodeData := range workflow {
		node, ok := nodeData.(map[string]interface{})
		if !ok {
			continue
		}

		// 获取 inputs
		inputs, ok := node["inputs"].(map[string]interface{})
		if !ok {
			continue
		}

		// 重写常见的路径字段
		pathFields := []string{
			"image",           // LoadImage 节点
			"filename_prefix", // SaveImage 节点
			"output_path",     // 自定义输出路径
			"input_path",      // 自定义输入路径
			"model",           // 模型路径（可能需要重写）
			"lora_name",       // LoRA 名称
			"vae_name",        // VAE 名称
		}

		for _, field := range pathFields {
			if value, ok := inputs[field].(string); ok {
				// 检查是否需要重写
				if newValue := rewriteFieldPath(field, value, userID); newValue != value {
					inputs[field] = newValue
					modified = true
				}
			}
		}
	}

	return modified
}

// rewriteFieldPath 重写特定字段的路径
func rewriteFieldPath(field, value string, userID uint) string {
	switch field {
	case "filename_prefix":
		// SaveImage 的输出前缀
		// "myimage" → "users/123/output/myimage"
		if !strings.HasPrefix(value, fmt.Sprintf("users/%d/", userID)) {
			return fmt.Sprintf("users/%d/output/%s", userID, value)
		}

	case "image":
		// LoadImage 的输入图片
		// "input.png" → "users/123/uploads/input.png"
		if !strings.HasPrefix(value, fmt.Sprintf("users/%d/", userID)) && !strings.HasPrefix(value, "users/") {
			return fmt.Sprintf("users/%d/uploads/%s", userID, value)
		}

	case "model", "lora_name", "vae_name":
		// 模型路径
		// 如果是 user_123/xxx，重写为 users/123/models/xxx
		if strings.HasPrefix(value, fmt.Sprintf("user_%d/", userID)) {
			return fmt.Sprintf("users/%d/models/%s", userID, strings.TrimPrefix(value, fmt.Sprintf("user_%d/", userID)))
		}
		// 否则是共享模型，不重写
	}

	return value
}

// ResponseRewriteMiddleware 重写响应中的路径（可选）
func ResponseRewriteMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户 ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		uid := userID.(uint)

		// 创建自定义 ResponseWriter 来拦截响应
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		// 如果是 JSON 响应，重写路径
		if strings.Contains(c.Writer.Header().Get("Content-Type"), "application/json") {
			var data interface{}
			if err := json.Unmarshal(blw.body.Bytes(), &data); err == nil {
				if rewriteResponsePaths(data, uid) {
					newBody, _ := json.Marshal(data)
					c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(newBody)))
					c.Writer.Write(newBody)
					return
				}
			}
		}

		// 否则直接返回原始响应
		c.Writer.Write(blw.body.Bytes())
	}
}

// bodyLogWriter 用于拦截响应体
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return len(b), nil
}

// rewriteResponsePaths 重写响应中的路径（将内部路径转换为用户可见路径）
func rewriteResponsePaths(data interface{}, userID uint) bool {
	modified := false
	userPrefix := fmt.Sprintf("users/%d/", userID)

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			// 递归处理
			if rewriteResponsePaths(value, userID) {
				modified = true
			}

			// 重写路径字段
			if str, ok := value.(string); ok {
				if strings.HasPrefix(str, userPrefix+"output/") {
					v[key] = strings.TrimPrefix(str, userPrefix+"output/")
					modified = true
				} else if strings.HasPrefix(str, userPrefix+"workflows/") {
					v[key] = strings.TrimPrefix(str, userPrefix+"workflows/")
					modified = true
				}
			}
		}

	case []interface{}:
		for _, item := range v {
			if rewriteResponsePaths(item, userID) {
				modified = true
			}
		}
	}

	return modified
}
