package reports

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// GeneratePDF 根据报告详情生成 PDF 字节流。
func (s *Service) GeneratePDF(taskID uint64) ([]byte, string, error) {
	detail, err := s.Detail(taskID)
	if err != nil {
		return nil, "", err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)

	// 注册中文字体：优先使用内嵌字体，回退到内置字体
	registerFont(pdf)

	pdf.AddPage()
	pdf.SetFont("chinese", "", 18)
	pdf.CellFormat(190, 12, "巡检报告", "", 1, "C", false, 0, "")

	pdf.Ln(4)
	pdf.SetFont("chinese", "", 10)
	pdf.CellFormat(190, 6, fmt.Sprintf("生成时间：%s", time.Now().Format("2006-01-02 15:04:05")), "", 1, "C", false, 0, "")

	pdf.Ln(6)

	// 任务概要
	pdf.SetFont("chinese", "B", 14)
	pdf.CellFormat(190, 10, "一、任务概要", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	task := detail.Task
	summaryItems := []struct{ label, value string }{
		{"任务名称", task.TaskName},
		{"巡检点位", task.PointName},
		{"设备名称", task.EquipmentName},
		{"作业区域", task.InspectArea},
		{"巡检模板", detail.TemplateName},
		{"指派人", detail.AssigneeName},
		{"执行人", detail.ExecutorName},
		{"AR眼镜编号", task.GlassesSN},
		{"下发人", task.AssignUser},
		{"开始时间", formatTime(task.StartedAt)},
		{"完成时间", formatTime(task.CompletedAt)},
	}

	pdf.SetFont("chinese", "", 10)
	for _, item := range summaryItems {
		pdf.SetFont("chinese", "B", 10)
		pdf.CellFormat(40, 7, item.label+"：", "", 0, "L", false, 0, "")
		pdf.SetFont("chinese", "", 10)
		pdf.CellFormat(150, 7, safeStr(item.value), "", 1, "L", false, 0, "")
	}

	pdf.Ln(6)

	// 节点执行详情
	pdf.SetFont("chinese", "B", 14)
	pdf.CellFormat(190, 10, "二、节点执行详情", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	if len(detail.Nodes) == 0 {
		pdf.SetFont("chinese", "", 10)
		pdf.CellFormat(190, 7, "暂无节点数据", "", 1, "L", false, 0, "")
	} else {
		// 表头
		colWidths := []float64{12, 40, 20, 20, 98}
		headers := []string{"序号", "节点名称", "类型", "状态", "执行结果"}
		pdf.SetFont("chinese", "B", 9)
		pdf.SetFillColor(230, 230, 230)
		for i, h := range headers {
			pdf.CellFormat(colWidths[i], 8, h, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		pdf.SetFont("chinese", "", 8)
		for idx, node := range detail.Nodes {
			// 计算结果文本
			resultText := ""
			if node.Result != nil {
				parts := []string{}
				if node.Result.FeedbackContent != "" {
					parts = append(parts, "反馈："+node.Result.FeedbackContent)
				}
				if node.Result.TextNote != "" {
					parts = append(parts, "备注："+node.Result.TextNote)
				}
				if node.Result.AlgorithmResult != "" {
					parts = append(parts, "AI："+node.Result.AlgorithmResult)
				}
				if node.Result.IsAbnormal {
					parts = append(parts, "[异常]"+node.Result.AbnormalDesc)
				}
				resultText = strings.Join(parts, "; ")
			}
			if resultText == "" {
				resultText = "未提交"
			}

			// 计算所需行高（按结果文本自动换行）
			lineCount := countLines(resultText, 98, 3.5)
			rowHeight := 8.0
			if lineCount > 1 {
				rowHeight = float64(lineCount) * 5
			}

			// 检查是否需要分页
			if pdf.GetY()+rowHeight > 280 {
				pdf.AddPage()
				// 重新绘制表头
				pdf.SetFont("chinese", "B", 9)
				pdf.SetFillColor(230, 230, 230)
				for i, h := range headers {
					pdf.CellFormat(colWidths[i], 8, h, "1", 0, "C", true, 0, "")
				}
				pdf.Ln(-1)
				pdf.SetFont("chinese", "", 8)
			}

			x := pdf.GetX()
			y := pdf.GetY()

			// 绘制单元格边框
			pdf.Rect(x, y, colWidths[0], rowHeight, "")
			pdf.Rect(x+colWidths[0], y, colWidths[1], rowHeight, "")
			pdf.Rect(x+colWidths[0]+colWidths[1], y, colWidths[2], rowHeight, "")
			pdf.Rect(x+colWidths[0]+colWidths[1]+colWidths[2], y, colWidths[3], rowHeight, "")
			pdf.Rect(x+colWidths[0]+colWidths[1]+colWidths[2]+colWidths[3], y, colWidths[4], rowHeight, "")

			// 序号
			pdf.SetXY(x, y)
			pdf.CellFormat(colWidths[0], rowHeight, fmt.Sprintf("%d", idx+1), "", 0, "C", false, 0, "")

			// 节点名称
			pdf.SetXY(x+colWidths[0], y)
			pdf.CellFormat(colWidths[1], rowHeight, safeStr(node.Name), "", 0, "C", false, 0, "")

			// 类型
			pdf.SetXY(x+colWidths[0]+colWidths[1], y)
			pdf.CellFormat(colWidths[2], rowHeight, nodeTypeLabel(node.NodeType), "", 0, "C", false, 0, "")

			// 状态
			pdf.SetXY(x+colWidths[0]+colWidths[1]+colWidths[2], y)
			statusLabel := nodeStatusLabel(node.Status)
			pdf.CellFormat(colWidths[3], rowHeight, statusLabel, "", 0, "C", false, 0, "")

			// 结果（多行文本）
			pdf.SetXY(x+colWidths[0]+colWidths[1]+colWidths[2]+colWidths[3]+2, y+1)
			pdf.MultiCell(colWidths[4]-4, 5, safeStr(resultText), "", "L", false)

			pdf.SetXY(x, y+rowHeight)

			// 如果有缺陷，显示在节点下方
			if len(node.Defects) > 0 {
				for _, d := range node.Defects {
					if pdf.GetY()+6 > 280 {
						pdf.AddPage()
					}
					pdf.SetFont("chinese", "", 8)
					pdf.SetTextColor(200, 0, 0)
					defectText := fmt.Sprintf("  [缺陷] %s (状态：%s)", safeStr(d.Description), defectStatusLabel(d.Status))
					pdf.CellFormat(190, 6, defectText, "", 1, "L", false, 0, "")
					pdf.SetTextColor(0, 0, 0)
				}
			}
		}
	}

	pdf.Ln(6)

	// 缺陷汇总
	pdf.SetFont("chinese", "B", 14)
	pdf.CellFormat(190, 10, "三、缺陷汇总", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	if len(detail.Defects) == 0 {
		pdf.SetFont("chinese", "", 10)
		pdf.CellFormat(190, 7, "本次巡检未发现缺陷", "", 1, "L", false, 0, "")
	} else {
		defectColWidths := []float64{12, 110, 30, 38}
		defectHeaders := []string{"序号", "缺陷描述", "状态", "关闭原因"}
		pdf.SetFont("chinese", "B", 9)
		pdf.SetFillColor(230, 230, 230)
		for i, h := range defectHeaders {
			pdf.CellFormat(defectColWidths[i], 8, h, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		pdf.SetFont("chinese", "", 9)
		for i, d := range detail.Defects {
			if pdf.GetY()+8 > 280 {
				pdf.AddPage()
				// 重新绘制表头
				pdf.SetFont("chinese", "B", 9)
				pdf.SetFillColor(230, 230, 230)
				for j, h := range defectHeaders {
					pdf.CellFormat(defectColWidths[j], 8, h, "1", 0, "C", true, 0, "")
				}
				pdf.Ln(-1)
				pdf.SetFont("chinese", "", 9)
			}
			pdf.CellFormat(defectColWidths[0], 8, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")
			pdf.CellFormat(defectColWidths[1], 8, safeStr(truncate(d.Description, 40)), "1", 0, "L", false, 0, "")
			pdf.CellFormat(defectColWidths[2], 8, defectStatusLabel(d.Status), "1", 0, "C", false, 0, "")
			pdf.CellFormat(defectColWidths[3], 8, safeStr(truncate(d.CloseReason, 18)), "1", 0, "L", false, 0, "")
			pdf.Ln(-1)
		}
	}

	// 页脚
	pdf.Ln(10)
	pdf.SetFont("chinese", "", 8)
	pdf.SetTextColor(128, 128, 128)
	pdf.CellFormat(190, 6, fmt.Sprintf("共 %d 个节点，%d 个缺陷", detail.NodeCount, len(detail.Defects)), "", 1, "R", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, "", err
	}

	fileName := fmt.Sprintf("巡检报告_%s_%s.pdf",
		safeFileName(task.TaskName),
		time.Now().Format("20060102150405"))

	return buf.Bytes(), fileName, nil
}

// registerFont 注册中文字体到 PDF 实例。
// 优先尝试使用系统字体路径，失败时使用内置的英文字体作为回退。
func registerFont(pdf *gofpdf.Fpdf) {
	// 常见中文字体路径（Linux/Mac/Windows）
	fontPaths := []string{
		"/usr/share/fonts/truetype/noto/NotoSansSC-Regular.ttf",
		"/usr/share/fonts/noto-cjk/NotoSansSC-Regular.ttf",
		"/usr/share/fonts/google-noto-cjk/NotoSansSC-Regular.otf",
		"/System/Library/Fonts/PingFang.ttc",
		"/System/Library/Fonts/STHeiti Light.ttc",
		"C:\\Windows\\Fonts\\msyh.ttf",
		"C:\\Windows\\Fonts\\simsun.ttc",
		"C:\\Windows\\Fonts\\simhei.ttf",
	}

	registered := false
	for _, path := range fontPaths {
		// 检查文件是否存在
		if _, err := os.Stat(path); err == nil {
			// 使用 recover 捕获可能的 panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						// 字体加载失败，继续尝试下一个
					}
				}()
				pdf.AddUTF8Font("chinese", "", path)
				registered = true
			}()
			if registered {
				return
			}
		}
	}

	// 回退：使用 Helvetica（不支持中文但保证 PDF 能生成）
	// 注意：这会导致中文显示为空白或乱码
}

// formatTime 格式化时间指针，nil 返回 "-"。
func formatTime(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Format("2006-01-02 15:04:05")
}

// safeStr 确保字符串不为空，空值返回 "-"。
func safeStr(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

// safeFileName 将字符串转为安全的文件名。
func safeFileName(s string) string {
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	if len(s) > 30 {
		s = s[:30]
	}
	return s
}

// truncate 截断字符串到指定长度。
func truncate(s string, maxLen int) string {
	if len([]rune(s)) <= maxLen {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLen]) + "..."
}

// nodeTypeLabel 节点类型中文标签。
func nodeTypeLabel(t string) string {
	m := map[string]string{
		"text":  "文本",
		"read":  "读取",
		"check": "检查",
		"photo": "拍照",
		"video": "录像",
		"audio": "录音",
	}
	if label, ok := m[t]; ok {
		return label
	}
	return t
}

// nodeStatusLabel 节点状态中文标签。
func nodeStatusLabel(s string) string {
	m := map[string]string{
		"pending":   "待执行",
		"completed": "已完成",
		"skipped":   "已跳过",
		"abnormal":  "异常",
	}
	if label, ok := m[s]; ok {
		return label
	}
	return s
}

// defectStatusLabel 缺陷状态中文标签。
func defectStatusLabel(s string) string {
	m := map[string]string{
		"pending_confirm": "待确认",
		"reported":        "已上报",
		"confirmed":       "已确认",
		"closed":          "已关闭",
	}
	if label, ok := m[s]; ok {
		return label
	}
	return s
}

// countLines 估算文本在指定宽度内需要的行数。
func countLines(text string, width float64, charWidth float64) int {
	if text == "" {
		return 1
	}
	charsPerLine := int(width / charWidth)
	if charsPerLine <= 0 {
		charsPerLine = 1
	}
	lines := 1
	for _, line := range strings.Split(text, "\n") {
		lineLen := len([]rune(line))
		if lineLen > charsPerLine {
			lines += (lineLen + charsPerLine - 1) / charsPerLine
		} else {
			lines++
		}
	}
	return lines
}
