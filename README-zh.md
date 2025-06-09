# Inspect4Oracle - Oracle 数据库巡检利器

[![Go Report Card](https://goreportcard.com/badge/github.com/goodwaysIT/inspect4oracle)](https://goreportcard.com/report/github.com/goodwaysIT/inspect4oracle)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)


**Inspect4Oracle 是一款强大、易用、开源的 Oracle 数据库巡检工具，旨在帮助数据库管理员 (DBA)、开发人员和运维工程师快速、全面地了解 Oracle 数据库的运行状态和健康状况。**

通过直观的 Web 界面，用户可以轻松连接到目标数据库，选择感兴趣的巡检模块，生成包含丰富图表和数据的交互式巡检报告。

## ✨ 项目亮点与优势

*   **全面巡检**: 内置多个核心巡检模块，覆盖数据库基本信息、参数配置、存储空间、对象状态、性能指标、备份恢复以及安全配置等关键领域。
*   **用户友好**: 提供现代化的 Web 用户界面，操作简单直观，无需复杂的命令行操作。
*   **交互式报告**: 生成的报告包含动态图表和可排序表格，方便用户深入分析数据。
*   **一键导出**: 支持将巡检报告导出为 HTML 格式，便于分享和离线查阅。
*   **轻松部署**: 基于 Go 语言开发，编译后为单个可执行文件，内置静态资源，无需额外依赖，部署简单快捷。
*   **跨平台运行**: 支持在 Windows, Linux, macOS 等主流操作系统上运行。
*   **开源免费**: 项目完全开源，您可以自由使用、修改和分发。
*   **高度可扩展**: 清晰的模块化设计，方便社区开发者贡献新的巡检模块和功能。
*   **安全连接**: 支持输入详细的连接信息，巡检过程不存储数据库凭证，保障数据安全。

## 🎯 目标用户

*   **数据库管理员 (DBA)**: 进行日常巡检、故障排查、性能优化和安全审计。
*   **开发人员**: 了解数据库环境配置，分析应用相关的数据库对象和性能。
*   **运维工程师**: 监控数据库状态，确保业务系统的稳定运行。
*   **数据库初学者**: 通过巡检报告学习 Oracle 数据库的内部结构和关键指标。

## 📸 界面截图

清晰的数据库连接界面：
![连接界面截图](./assets/images/connection_ui_zh.png)
直观的巡检报告概览：
![报告概览截图](./assets/images/report_overview_zh.png)
丰富的交互式图表展示：
![图表示例截图](./assets/images/chart_example_zh.gif)
灵活的巡检模块选择与报告生成：
![报告设置截图](./assets/images/report_settings_zh.png)

## 🚀 快速开始

### 1. 获取程序

*   **下载预编译版本 (推荐)**:
    前往本项目的 [GitHub Releases](https://github.com/goodwaysIT/inspect4oracle/releases) 页面下载适用于您操作系统的最新预编译版本。
*   **从源码构建**:
    如果您希望自行构建，请参考项目的 [BUILD-zh.md](./BUILD-zh.md) 构建指南。

### 2. 运行程序

您可以通过以下几种方式运行程序：

*   **通过 `go run` (适用于已配置 Go 开发环境的场景)**:
    ```bash
    go run main.go
    ```
    (请将 `main.go` 替换为实际的项目入口文件名，如果它不同的话)

*   **直接运行预编译的可执行文件**:
    下载或构建完成后，直接运行可执行文件：

*   **Windows**: 双击 `inspect4oracle.exe` 或在命令行运行 `inspect4oracle.exe`。
*   **Linux / macOS**: 在终端运行 `./inspect4oracle`。

程序启动后，会显示监听的 IP 地址和端口号，默认为 `http://0.0.0.0:8080`。

您可以通过 `-h` 或 `--help` 参数查看所有可用的命令行选项，例如：
```bash
# Windows
inspect4oracle.exe -h

# Linux / macOS
./inspect4oracle -h
```
这将显示如何指定不同的监听端口、开启调试模式等。

### 3. 开始巡检

1.  打开您的 Web 浏览器，访问程序启动时提示的地址 (例如 `http://localhost:8080`)。
2.  在首页的连接表单中，输入您的 Oracle 数据库的连接信息 (主机、端口、服务名/SID、用户名、密码)。
3.  点击“验证连接”以确保连接信息正确且用户拥有必要的查询权限。
4.  选择您希望巡检的模块。
5.  点击“开始巡检”按钮。
6.  巡检完成后，系统将自动跳转到生成的巡检报告页面。
7.  您可以浏览报告、与图表交互，并通过报告页面的导出功能将报告保存为 HTML 文件。

> **注意**:
> 为了获取最全面的巡检信息并确保所有模块都能正常工作，建议使用 `SYSTEM` 用户执行巡检。
>
> 如果您希望使用权限受限的普通用户执行巡检，请确保该用户已被授予访问相关数据字典视图和动态性能视图（如 `V$`视图、`DBA_`视图等）的必要查询权限。以下是根据程序内部权限校验列表生成的基础授权SQL脚本示例，您可以根据实际需要巡检的模块和数据库版本进行调整和补充：
```sql
-- 授予查询以下V$视图的权限:
GRANT SELECT ON V_$ACTIVE_SESSION_HISTORY TO YOUR_USER;
GRANT SELECT ON V_$ASM_DISKGROUP TO YOUR_USER; -- 如果使用ASM且需要检查
GRANT SELECT ON V_$DATABASE TO YOUR_USER;
GRANT SELECT ON V_$INSTANCE TO YOUR_USER;
GRANT SELECT ON V_$SESSION TO YOUR_USER;
GRANT SELECT ON V_$SQL TO YOUR_USER;
GRANT SELECT ON V_$SQLAREA TO YOUR_USER;
GRANT SELECT ON V_$SYSMETRIC TO YOUR_USER;
GRANT SELECT ON V_$SYSTEM_PARAMETER TO YOUR_USER;
GRANT SELECT ON V_$TEMP_EXTENT_POOL TO YOUR_USER;
GRANT SELECT ON V_$VERSION TO YOUR_USER;

-- 授予查询以下DBA_视图的权限:
GRANT SELECT ON DBA_DATA_FILES TO YOUR_USER;
GRANT SELECT ON DBA_FREE_SPACE TO YOUR_USER;
GRANT SELECT ON DBA_OBJECTS TO YOUR_USER;
GRANT SELECT ON DBA_ROLES TO YOUR_USER;
GRANT SELECT ON DBA_ROLE_PRIVS TO YOUR_USER;
GRANT SELECT ON DBA_SEGMENTS TO YOUR_USER;
GRANT SELECT ON DBA_SYS_PRIVS TO YOUR_USER;
GRANT SELECT ON DBA_TABLESPACES TO YOUR_USER;
GRANT SELECT ON DBA_TEMP_FILES TO YOUR_USER;
GRANT SELECT ON DBA_USERS TO YOUR_USER;

-- 根据您启用的巡检模块，可能还需要其他权限，例如:
-- GRANT SELECT ON V_$PARAMETER TO YOUR_USER; (替代 V_$SYSTEM_PARAMETER)
-- GRANT SELECT ON DBA_PROFILES TO YOUR_USER; (安全模块)
-- GRANT SELECT ON V_$RMAN_BACKUP_JOB_DETAILS TO YOUR_USER; (备份模块)
-- GRANT SELECT ON V_$FLASHBACK_DATABASE_LOG TO YOUR_USER; (备份模块)
-- GRANT SELECT ON DBA_RECYCLEBIN TO YOUR_USER; (备份模块)
-- GRANT SELECT ON DBA_DATAPUMP_JOBS TO YOUR_USER; (备份模块)
-- GRANT SELECT ON DBA_AUDIT_TRAIL TO YOUR_USER; (如果使用传统审计)
-- ... 请根据实际巡检范围和错误日志补充更多权限 ...
```

## 📦 核心巡检模块

Inspect4Oracle 提供以下核心巡检模块 (部分模块可能仍在开发中，欢迎关注项目进展)：

*   **`dbinfo` (数据库信息)**:
    *   数据库版本、实例信息、启动时间、平台信息等。
    *   NLS 参数设置。
*   **`parameters` (参数配置)**:
    *   非默认数据库参数列表及其值。
    *   重要的隐藏参数 (按需)。
*   **`storage` (存储管理)**:
    *   表空间使用情况 (总量、已用、可用、百分比)。
    *   数据文件信息。
    *   控制文件和重做日志文件状态。
    *   ASM 磁盘组信息 (如果数据库使用 ASM)。
*   **`objects` (对象状态)**:
    *   无效对象列表 (OWNER, OBJECT_NAME, OBJECT_TYPE)。
    *   对象类型统计。
    *   大对象/段信息 (Top Segments by size)。
*   **`performance` (性能分析)**:
    *   关键等待事件。
    *   当前会话信息。
    *   SGA/PGA 内存使用情况。
    *   命中率 (Buffer Cache Hit Ratio, Library Cache Hit Ratio等)。
    *   (更多性能指标正在规划中)
*   **`backup` (备份与恢复)**:
    *   归档模式状态。
    *   最近 RMAN 备份任务记录 (成功/失败)。
    *   闪回数据库状态和空间使用。
    *   回收站对象信息。
    *   Data Pump 任务历史。
*   **`security` (安全审计)**:
    *   非系统用户信息 (状态、锁定/过期、默认表空间、Profile)。
    *   拥有 DBA 等高权限角色的用户。
    *   用户系统权限列表。
    *   Profile 配置 (特别是密码策略相关参数，如 `FAILED_LOGIN_ATTEMPTS`, `PASSWORD_LIFE_TIME`)。
    *   非系统角色列表。
    *   (审计配置等更多安全特性正在规划中)

## 🤝 参与贡献

我们热烈欢迎社区的开发者参与到 Inspect4Oracle 项目的贡献中来！无论是报告 Bug、提出功能建议，还是直接贡献代码，您的帮助对项目都至关重要。

### 如何贡献

1.  **报告问题 (Bugs)**: 如果您在使用过程中发现任何问题，请通过 GitHub Issues 提交详细的 Bug 报告。
2.  **功能建议**: 如果您有新的功能想法或改进建议，也请通过 GitHub Issues 提出。
3.  **贡献代码**:
    *   Fork 本仓库到您的 GitHub 账户。
    *   创建一个新的分支 (例如 `feature/your-new-feature` 或 `fix/issue-number`)。
    *   在您的分支上进行修改和开发。
    *   确保您的代码遵循项目现有的编码风格和规范。
    *   提交您的更改，并推送到您的 Fork 仓库。
    *   创建一个 Pull Request (PR) 到主仓库的 `main` (或 `develop`) 分支，并详细描述您的更改内容。



## 📜 开源许可

本项目基于 [MIT License](LICENSE) 开源。

