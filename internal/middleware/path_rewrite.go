package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strings"

	"github.com/gin-gonic/gin"
)

// userPrefix returns the per-user directory prefix, e.g. "user_123"
func userPrefix(uid uint) string {
	return fmt.Sprintf("user_%d", uid)
}

// PathRewriteMiddleware 路径重写中间件（用户数据隔离）
//
// 隔离策略：
//   - 工作流/设置：通过 Comfy-User header，利用 ComfyUI 原生多用户支持
//   - 生成图片：改写 SaveImage 节点的 filename_prefix，输出到 output/user_{id}/
//   - 查看图片：校验 /view 请求的 subfolder 属于当前用户
//   - 上传图片：注入 subfolder=user_{id}，上传到 input/user_{id}/
//   - 模型：共享，不改写
func PathRewriteMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		uid := userID.(uint)
		prefix := userPrefix(uid)
		path := c.Request.URL.Path

		// 1. 注入 Comfy-User header → ComfyUI 原生 userdata 隔离
		c.Request.Header.Set("Comfy-User", prefix)

		// 2. /view 请求：校验 subfolder 防止跨用户访问
		if path == "/view" || path == "/comfy/view" {
			rewriteViewRequest(c, prefix)
		}

		// 3. /prompt 请求：改写 SaveImage 节点的 filename_prefix
		if c.Request.Method == "POST" && (path == "/prompt" || path == "/comfy/prompt") {
			rewritePromptBody(c, prefix)
		}

		// 4. /upload/image 请求：注入 subfolder
		if c.Request.Method == "POST" && (strings.HasSuffix(path, "/upload/image") || strings.HasSuffix(path, "/upload/mask")) {
			rewriteUploadRequest(c, prefix)
		}

		c.Next()
	}
}

// rewriteViewRequest 校验并改写 /view 请求的 subfolder
func rewriteViewRequest(c *gin.Context, prefix string) {
	viewType := c.Query("type")
	subfolder := c.Query("subfolder")

	// output 和 input 类型需要隔离
	if viewType == "output" || viewType == "input" {
		// 强制 subfolder 以 user_{id} 开头
		if !strings.HasPrefix(subfolder, prefix) {
			q := c.Request.URL.Query()
			if subfolder == "" {
				q.Set("subfolder", prefix)
			} else {
				q.Set("subfolder", prefix+"/"+subfolder)
			}
			c.Request.URL.RawQuery = q.Encode()
		}
	}
	// temp 类型不限制
}

// rewritePromptBody 改写 /prompt 请求体中 SaveImage 等节点的 filename_prefix
func rewritePromptBody(c *gin.Context, prefix string) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return
	}
	c.Request.Body.Close()

	var data map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		return
	}

	modified := false
	if prompt, ok := data["prompt"].(map[string]interface{}); ok {
		modified = rewritePromptNodes(prompt, prefix)
	}

	if modified {
		newBody, err := json.Marshal(data)
		if err == nil {
			c.Request.Body = io.NopCloser(bytes.NewBuffer(newBody))
			c.Request.ContentLength = int64(len(newBody))
			return
		}
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
}

// rewritePromptNodes 遍历 workflow 节点，改写输出路径
func rewritePromptNodes(prompt map[string]interface{}, prefix string) bool {
	modified := false

	// 需要改写 filename_prefix 的节点类型
	outputNodeTypes := map[string]bool{
		"SaveImage":        true,
		"SaveAnimatedWEBP": true,
		"SaveAnimatedPNG":  true,
		"SaveLatent":       true,
	}

	for _, nodeData := range prompt {
		node, ok := nodeData.(map[string]interface{})
		if !ok {
			continue
		}

		classType, _ := node["class_type"].(string)
		inputs, ok := node["inputs"].(map[string]interface{})
		if !ok {
			continue
		}

		// 改写输出节点的 filename_prefix
		if outputNodeTypes[classType] {
			if fp, ok := inputs["filename_prefix"].(string); ok {
				if !strings.HasPrefix(fp, prefix+"/") {
					inputs["filename_prefix"] = prefix + "/" + fp
					modified = true
				}
			}
		}
	}

	return modified
}

// rewriteUploadRequest 改写上传请求，注入用户 subfolder
func rewriteUploadRequest(c *gin.Context, prefix string) {
	contentType := c.GetHeader("Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return
	}
	c.Request.Body.Close()

	reader := multipart.NewReader(bytes.NewReader(body), params["boundary"])

	// 重建 multipart body，注入/覆盖 subfolder 字段
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	hasSubfolder := false
	for {
		part, err := reader.NextPart()
		if err != nil {
			break
		}

		fieldName := part.FormName()
		partBody, _ := io.ReadAll(part)

		if fieldName == "subfolder" {
			// 覆盖 subfolder：加上用户前缀
			original := strings.TrimSpace(string(partBody))
			var newVal string
			if original == "" {
				newVal = prefix
			} else {
				newVal = prefix + "/" + original
			}
			writeFormField(writer, "subfolder", newVal)
			hasSubfolder = true
		} else if part.FileName() != "" {
			// 文件字段
			pw, _ := writer.CreatePart(part.Header)
			pw.Write(partBody)
		} else {
			writeFormField(writer, fieldName, string(partBody))
		}
	}

	if !hasSubfolder {
		writeFormField(writer, "subfolder", prefix)
	}

	writer.Close()

	c.Request.Body = io.NopCloser(&buf)
	c.Request.ContentLength = int64(buf.Len())
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
}

func writeFormField(w *multipart.Writer, fieldName, value string) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, fieldName))
	p, _ := w.CreatePart(h)
	p.Write([]byte(value))
}
