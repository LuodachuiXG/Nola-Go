package util

import (
	"bytes"
	"nola-go/internal/logger"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"go.uber.org/zap"
)

// MarkdownToHtml 将 Markdown 文本转为 Html
func MarkdownToHtml(markdown string) string {

	mdb := []byte(markdown)

	md := goldmark.New()
	var buf bytes.Buffer

	// 转换 Markdown
	if err := md.Convert(mdb, &buf); err != nil {
		logger.Log.Error("Markdown 转换 Html 失败", zap.Error(err))
		return ""
	}

	return buf.String()
}

// HtmlToPlainText 将 Html 转为纯文本，去掉标记
func HtmlToPlainText(html string) string {
	// 去除 HTML 标签
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(html, "")

	// 解码 HTML 实体
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	return text
}

// MarkdownToPlainText 将 Markdown 转为纯文本
func MarkdownToPlainText(markdown string) string {
	html := MarkdownToHtml(markdown)
	return strings.ReplaceAll(HtmlToPlainText(html), "\n", "")
}
