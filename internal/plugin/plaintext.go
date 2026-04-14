package plugin

import (
	"bufio"
	"context"
	"os"
	"regexp"
	"strings"
)

// PlainTextParser 实现了针对通用的纯文本文件 (.txt, .log) 的解析逻辑。
// 对齐原版 whispers/plugins/plaintext.py 的实现。
type PlainTextParser struct {
	uriParser *UriParser
	reUri     *regexp.Regexp
}

func NewPlainTextParser() *PlainTextParser {
	return &PlainTextParser{
		uriParser: NewUriParser(),
		// 预编译用于识别 URI 的正则
		reUri: regexp.MustCompile(`(?i)(http|ftp|smtp|scp|ssh|jdbc[:\w\d]*|s3)s?://?.+`),
	}
}

// SupportedExtensions 返回支持的后缀。
func (p *PlainTextParser) SupportedExtensions() []string {
	return []string{".txt", ".log", ".md", ".text"}
}

// Parse 执行纯文本扫描（主要寻找嵌入的 URI 字符串）。
func (p *PlainTextParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results []KeyValue
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// 对齐原版 logic: for value in line.split()
		words := strings.Fields(line)
		for _, word := range words {
			// 如果单词匹配 URI 格式，则作为凭据进行解析
			if p.reUri.MatchString(word) {
				uriKVs := p.uriParser.ParseURI(word)
				for _, kv := range uriKVs {
					kv.Line = lineNum
					kv.Path = "plaintext.uri"
					results = append(results, kv)
				}
			}
		}
	}

	return results, nil
}
