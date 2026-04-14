package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"ditting/internal/core"
)

// Reporter 定义了把扫描结果导出为指定格式文件的接口。
type Reporter interface {
	Generate(secrets []core.Secret, outputPath string) error
}

// JsonReporter 将结果以美化的 JSON 列表形式输出。
type JsonReporter struct{}

func (r *JsonReporter) Generate(secrets []core.Secret, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("无法创建 JSON 报告文件: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // 美化输出
	return encoder.Encode(secrets)
}

// CsvReporter 将结果以标准 CSV 格式输出。
type CsvReporter struct{}

func (r *CsvReporter) Generate(secrets []core.Secret, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("无法创建 CSV 报告文件: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	header := []string{"Severity", "RuleID", "Description", "FilePath", "LineNumber", "Content"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// 写入数据条目
	for _, secret := range secrets {
		row := []string{
			secret.Severity,
			secret.RuleID,
			secret.Description,
			secret.FilePath,
			fmt.Sprintf("%d", secret.LineNumber),
			secret.Content,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
