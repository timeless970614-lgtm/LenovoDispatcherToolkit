# ipf_wrapper DLL 构建说明

## 目录结构

```
backend/
  ipf_wrapper/
    ipf_wrapper.h       ← C 头文件（给 cgo 用）
    ipf_wrapper.cpp     ← C++ 实现
    ipf_wrapper.def     ← DLL 导出表
    BUILD.md            ← 本文件
```

## 依赖

- **源 DLL**（运行时需要，放在 EXE 同目录或指定路径）：
  - `LenovoIPFV2.dll`（优先）或 `LenovoIPF.dll`（备选）
  - 来源：复制自 `C:\LenovoDispatcher\ML_Scenario\LenovoIPFV2\LenovoIPFV2.dll`
    或 `C:\LenovoDispatcher\ML_Scenario\LenovoIPF\LenovoIPF.dll`

## 方法一：Visual Studio（推荐）

### 1. 创建 DLL 项目

1. 打开 Visual Studio → **创建新项目** → 选择 **"动态链接库 (DLL)"**
2. 项目名：`ipf_wrapper`，位置：`C:\LenovoDispatcher\LenovoToolkit\backend\`
3. 删除自动生成的 `pch.h / pch.cpp`，将 `ipf_wrapper.h`、`ipf_wrapper.cpp`、`ipf_wrapper.def` 添加到项目中

### 2. 配置项目属性

| 属性 | 值 |
|------|-----|
| **C/C++ → 预处理器** | 添加 `;_CRT_SECURE_NO_WARNINGS` |
| **C/C++ → 代码生成** | 运行库 = `/MD`（发布）或 `/MDd`（调试） |
| **链接器 → 输入 → 附加依赖项** | `user32.lib; ole32.lib` |
| **链接器 → 高级 → 导入库** | `$(OutDir)ipf_wrapper.lib` |
| **链接器 → 高级 → 无入口点** | `否` |
| **链接器 → 常规 → 输出文件** | `$(OutDir)ipf_wrapper.dll` |

### 3. 移除默认源文件

项目创建时自动生成的 `.cpp` 文件（通常叫 `dllmain.cpp`）会与本项目冲突。在 **解决方案资源管理器** 中删除它，只保留 `ipf_wrapper.cpp`。

### 4. 编译

- **配置**：Release / x64
- 生成的 `ipf_wrapper.dll` 放在 EXE 同目录下：
  `C:\LenovoDispatcher\LenovoToolkit\build\bin\ipf_wrapper.dll`

## 方法二：命令行（MinGW-w64）

如果已安装 MinGW-w64（gcc）：

```bat
:: 在 "x64 Native Tools Command Prompt for VS" 中运行：
cd C:\LenovoDispatcher\LenovoToolkit\backend\ipf_wrapper
g++ -shared -o ipf_wrapper.dll ipf_wrapper.cpp ipf_wrapper.def ^
    -Wall -Wl,--out-implib,ipf_wrapper.lib ^
    -D_CRT_SECURE_NO_WARNINGS
```

## 方法三：Visual Studio 命令行（msvc）

```bat
:: 在 "x64 Native Tools Command Prompt for VS" 中运行：
cd C:\LenovoDispatcher\LenovoToolkit\backend\ipf_wrapper
cl /LD /EHsc /D_CRT_SECURE_NO_WARNINGS ^
    ipf_wrapper.cpp ipf_wrapper.def ^
    /link /DEF:ipf_wrapper.def /OUT:ipf_wrapper.dll
```

## 验证 DLL 导出

```powershell
dumpbin /EXPORTS C:\LenovoDispatcher\LenovoToolkit\build\bin\ipf_wrapper.dll
```

确认导出：
```
IPF_Connect
IPF_GetVersion
IPF_Disconnect
IPF_GetSystemPower_mW
IPF_GetCpuTemp_cK
IPF_GetPL1_mW
IPF_GetPL2_mW
IPF_GetPL4_mW
IPF_GetAllPL_mW
IPF_SetDllPath
```

## IPF DLL 复制

编译成功后，将源 DLL 复制到输出目录：

```powershell
Copy-Item "C:\LenovoDispatcher\ML_Scenario\LenovoIPFV2\LenovoIPFV2.dll" "C:\LenovoDispatcher\LenovoToolkit\build\bin\LenovoIPFV2.dll"
Copy-Item "C:\LenovoDispatcher\ML_Scenario\LenovoIPF\LenovoIPF.dll"       "C:\LenovoDispatcher\LenovoToolkit\build\bin\LenovoIPF.dll"
```

## 调试技巧

如果 `IPF_Connect()` 返回 0：
1. 确认 `LenovoIPFV2.dll` 或 `LenovoIPF.dll` 在 EXE 同目录
2. 用 [Dependencies Walker](https://github.com/lucasg/Dependencies) 或 `dumpbin /dependents` 检查 DLL 依赖
3. 检查 Lenovo Vantage / ITS 服务是否运行
