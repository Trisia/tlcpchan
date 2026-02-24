package commands

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"
)

func systemLogsList(args []string) error {
	resp, err := cli.ListLogs()
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(resp)
	}

	if len(resp.Files) == 0 {
		fmt.Println("没有日志文件")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "文件名\t大小\t修改时间\t当前\n")
	fmt.Fprintf(w, "------\t----\t--------\t----\n")
	for _, f := range resp.Files {
		size := formatBytes(f.Size)
		modTime, _ := time.Parse(time.RFC3339, f.ModTime)
		modTimeStr := modTime.Format("2006-01-02 15:04:05")
		current := ""
		if f.Current {
			current = "*"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", f.Name, size, modTimeStr, current)
	}
	w.Flush()

	return nil
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func systemLogsContent(args []string) error {
	fs := flagSet("content")
	file := fs.String("file", "", "日志文件名，默认读取当前日志文件")
	fs.StringVar(file, "f", "", "日志文件名(缩写)")
	lines := fs.Int("lines", 500, "读取行数，默认500行，最大2000行")
	fs.IntVar(lines, "n", 500, "读取行数(缩写)")
	level := fs.String("level", "", "日志级别过滤：debug、info、warn、error")
	fs.StringVar(level, "l", "", "日志级别(缩写)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *lines > 2000 {
		*lines = 2000
	}

	resp, err := cli.GetLogContent(*file, *lines, *level)
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(resp)
	}

	fmt.Printf("文件: %s\n", resp.File)
	fmt.Printf("总行数: %d, 返回: %d\n", resp.Total, resp.Returned)
	fmt.Println("---")
	for _, line := range resp.Lines {
		fmt.Println(line)
	}

	return nil
}

func systemLogsDownload(args []string) error {
	fs := flagSet("download")
	output := fs.String("output", "", "输出文件路径，默认保存到当前目录")
	fs.StringVar(output, "o", "", "输出路径(缩写)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if len(args) == 0 {
		fmt.Println("请指定要下载的日志文件名")
		fs.Usage()
		return fmt.Errorf("文件名不能为空")
	}

	filename := args[0]

	data, err := cli.DownloadLog(filename)
	if err != nil {
		return err
	}

	outputPath := *output
	if outputPath == "" {
		outputPath = filename
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}

	fmt.Printf("已下载: %s -> %s (%d bytes)\n", filename, outputPath, len(data))
	return nil
}

func systemLogsDownloadAll(args []string) error {
	fs := flagSet("download-all")
	output := fs.String("output", "", "输出文件路径，默认保存到当前目录")
	fs.StringVar(output, "o", "", "输出路径(缩写)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	data, err := cli.DownloadAllLogs()
	if err != nil {
		return err
	}

	defaultName := fmt.Sprintf("tlcpchan-logs-%s.zip", time.Now().Format("20060102-150405"))
	outputPath := *output
	if outputPath == "" {
		outputPath = defaultName
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}

	fmt.Printf("已下载所有日志: %s (%d bytes)\n", outputPath, len(data))
	return nil
}
