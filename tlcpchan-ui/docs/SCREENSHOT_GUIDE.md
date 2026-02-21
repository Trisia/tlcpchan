# UI 截图补充指南

当前用户手册的**核心内容和拓扑图已经完成**，但还缺少实际的UI界面截图。本指南说明如何补充这些截图。

## 已完成内容 ✅

- ✅ 完整的用户手册框架和文字内容
- ✅ 4张专业的系统架构/拓扑图
- ✅ 详细的操作步骤说明

## 需要补充的UI截图 📸

以下截图需要补充到 `tlcpchan-ui/docs/img/` 目录：

| 文件名 | 说明 | 关键标注位置 |
|--------|------|-------------|
| `01-dashboard.png` | 仪表盘主界面 | 左侧导航菜单、统计卡片区域 |
| `02-instances.png` | 实例管理界面 | "新建实例"按钮、实例列表 |
| `03-create-tlcp.png` | 创建TLCP实例对话框 | 协议选择下拉框、密钥库选择 |
| `04-create-tls.png` | 创建TLS实例对话框 | 协议选择下拉框、密钥库选择 |
| `05-create-auto.png` | 创建AUTO实例对话框 | 协议选择"AUTO"、双密钥库配置 |
| `06-keystores.png` | 密钥管理界面 | "生成密钥库"按钮、密钥库列表 |
| `07-trusted-certs.png` | 信任证书界面 | "上传证书"按钮、证书列表 |

## 截图步骤

### 方法1：使用浏览器开发者工具（推荐）

1. **启动服务**（如果尚未启动）：
   ```bash
   # 后端已在 target/tlcpchan-test 运行
   # 前端已在 localhost:3000 运行
   ```

2. **访问UI**：打开浏览器访问 `http://localhost:3000`

3. **逐个页面截图**：
   - 导航到每个需要截图的页面
   - 使用浏览器截图功能（Ctrl+Shift+S 或开发者工具截图）
   - 保存为PNG格式

4. **添加红色标注**：
   - 使用图片编辑工具（如Photoshop、GIMP或在线工具）
   - 在关键操作位置添加红色矩形框

### 方法2：使用Python脚本自动化

如果有完整的base64截图数据，可以使用提供的脚本：

```python
from target.process_screenshot import save_base64_screenshot, add_red_boxes

# 保存截图
save_base64_screenshot(base64_data, "tlcpchan-ui/docs/img/01-dashboard.png")

# 添加红色标注（例如标注"新建实例"按钮位置）
add_red_boxes(
    "tlcpchan-ui/docs/img/02-instances.png",
    [(100, 150, 300, 200)]  # x1, y1, x2, y2
)
```

## 当前手册状态

用户手册已位于：`tlcpchan-ui/README.md`

手册包含：
- 快速开始指南
- 三种实例类型（TLCP/TLS/AUTO）的完整教程
- 证书管理操作指南
- 应用场景配置说明
- 完整的拓扑结构说明（带专业生成的图表）

补充UI截图后，手册将更加直观和易用！
