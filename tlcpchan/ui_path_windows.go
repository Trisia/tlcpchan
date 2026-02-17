//go:build windows

package main

// getUIExecutableName 返回 Windows 平台下 UI 可执行文件的名称
func getUIExecutableName() string {
	return "tlcpchan-ui.exe"
}
