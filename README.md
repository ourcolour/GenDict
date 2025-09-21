# 多数据库字典生成工具
# Multi-DB Data Dictionary Generator

[![Go Version](https://img.shields.io/badge/Go-1.19%2B-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/your-username/your-repo-name/pulls)

一个使用Go语言开发的可视化数据字典生成工具，支持多种关系型数据库。该工具能够自动扫描数据库结构，生成详细的数据字典文档，并支持多种输出格式。

![screen](screen.png)

## 功能特性

- **多数据库支持**: 支持 MySQL、PostgreSQL、Oracle、SQLite 等多种主流关系型数据库
- **自动化扫描**: 自动解析数据库结构，提取表、视图、字段等元数据信息（暂不支持索引）
- **多种输出格式**: 支持将数据字典导出为 Excel、Markdown 等多种格式，方便查阅和集成

## 安装说明

### 前提条件

-   **Go 1.21** 或更高版本
-   需要访问的目标数据库（如 MySQL、PostgreSQL）

### 安装方式

1. **从源码安装**

```shell
# 克隆项目代码

git clone https://github.com/ourcolour/GenDict.git
cd GenDict

# 编译并安装
go mod download
go build -o gen-dict main.go

# 将可执行文件移动到 PATH 环境变量包含的目录中，例如
sudo mv gen-dict /usr/local/bin/
```

2. **直接下载二进制文件**

你可以从项目的 Release 页面下载预编译的二进制文件，适用于 Windows、Linux 和 macOS。

3. **支持的数据库**

当前工具支持以下数据库类型。其元数据提取基于各数据库的 information_schema或系统目录表。

| 数据库 | 支持状态 | 备注 |
| ------- | ------ | ---- |
| MySQL | ✅ 支持 | 默认 |
| PostgreSQL | ✅ 支持 | 默认 |
| Oracle | ✅ 支持 | 默认 |
| SQLite | ✅ 支持 | 默认 |

## 贡献指南

我们欢迎任何形式的贡献！包括但不限于：

- 报告 Bug
- 提出新功能或改进建议
- 提交 Pull Request

1. Fork 本项目
2. 创建你的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交你的更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开一个 Pull Request

## 许可证

本项目采用 Apache License 2.0 开源许可证。详情请参阅 [LICENSE](LICENSE) 文件。

## 致谢

- 感谢所有为这个项目做出贡献的开发者们。

---

**注意**: 使用此工具时，请确保你拥有连接目标数据库的合法权限，并遵守相关的数据安全和隐私规定。