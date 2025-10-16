# STools - Simple Tools Manager

stools 是一个简单的工具管理器，用于管理和执行各种工具脚本。它提供了一个统一的接口来添加、删除和执行工具。

## 功能特性

- **工具管理**：添加和删除工具
- **工具执行**：统一执行已安装的工具
- **配置管理**：使用 YAML 配置文件管理工具信息
- **自动补全**：显示所有可用工具及其描述


## 使用方法

### 基本语法

```bash
./stools <command> [parameters]
```

### 命令说明

#### 1. 执行工具
```bash
./stools <toolname> [parameters]
```
执行指定的工具，可以传递参数给工具。

#### 2. 添加工具
```bash
./stools add <toolpath> <description>
```
- `toolpath`：工具的路径（目录）
- `description`：工具的描述信息

**示例：**
```bash
./stools add /path/to/my-tool "My custom tool for data processing"
```

#### 3. 删除工具
```bash
./stools rm <toolname>
```
删除指定的工具及其配置。

**示例：**
```bash
./stools rm my-tool
```

#### 4. 查看帮助
```bash
./stools
```
显示使用说明和所有可用工具列表。

## 工具结构

每个工具应该包含以下结构：

```
module/
├── tool1/
│   ├── tool          # 可执行脚本（入口文件）
│   └── tool.py       # 实际的工具脚本
├── tool2/
│   ├── tool          # 可执行脚本（入口文件）
│   └── script.R       # 实际的工具脚本
└── ...
```

### 工具目录详细说明

#### 1. 工具目录结构
每个工具目录必须包含：
- **`tool`**：可执行入口文件（必需）
- **其他文件**：实际的工具脚本、配置文件等（可选）

#### 2. `tool` 入口文件
`tool` 文件是工具的入口点，固定名称，必须具有可执行权限。它负责：
- 设置环境变量
- 调用实际的工具脚本
- 传递命令行参数

**示例 `tool` 文件（Shell 脚本）：**
```bash
#!/bin/bash
export scriptpath=$(cd `dirname $0`; pwd)/tool.py
/path/python3 $scriptpath "$@"
```

tool文件里请包含python、Rscript、make或perl这种语言解释器，`"$@"`这种写法会把脚本需要的参数做正确的传递。

#### 3. 实际工具脚本
实际的工具逻辑可以写在任何脚本文件中，常见的有：
- `tool.py`: Python 脚本
- `tool.R`: R 脚本
- `tool.pl`: perl 脚本
- `tool.mk`: makefile 脚本，默认调用makefile脚本的**main**模块
- 其他任何可执行脚本

#### 4. 工具目录示例

## 配置文件

程序使用 `conf.yaml` 文件来存储工具信息：

```yaml
tool1: "Description of tool1"
tool2: "Description of tool2"
```

配置文件会自动创建和更新，无需手动编辑。

## 示例

### 添加一个 R 脚本工具
```
deseq2/
├── tool              # 入口文件
└── tool.R            # Rscript 脚本
```

1. 创建工具目录：
**目录名称即是工具名称**

```bash
mkdir -p deseq2
```

2. 创建可执行脚本：tool
```bash
#!/bin/bash
# deseq2/tool
export scriptpath=$(cd `dirname $0`; pwd)/tool.R
/path/Rscript $scriptpath "$@"
```

3. 编写tool.R：
```R
args = commandArgs(T)

countsData = read.csv(args[1], sep="\t", header=T, row.names=1)
design = read.table(args[2], header=T, row.names=1)
cmp_raw = unlist(strsplit(args[3], ','))

library(DESeq2)
ddsFullCountTable <- DESeqDataSetFromMatrix(countData = countsData, colData = design , design = ~condition)
dds = DESeq(ddsFullCountTable)
res = results(dds , contrast = c("condition",cmp[1], cmp[2]))

out_file_name = paste(cmp_raw[1], cmp_raw[2], "transcript.xls", sep="_")
write.table(res, file=out_file_name, sep="\t", quote = FALSE, row.names = TRUE)
```

4. 添加到 stools：
```bash
./stools add ./deseq2 "A simple R tool"
```

5. 执行工具：
```bash
./stools deseq2
./stools deseq2 counts.tsv desisn.txt Case,Ctl
# 其实相当于执行了 Rscript tool.R counts.tsv desisn.txt Case,Ctl
```

### 添加一个 makefile 脚本工具

1. 创建工具目录：
```bash
mkdir -p my-make-tool
```

2. 创建可执行脚本：tool
```bash
#!/bin/bash
# my-shell-tool/tool
export scriptpath=$(cd `dirname $0`; pwd)/tool.mk
make -f $scriptpath "$@"
```

3. 编写tool.mk：
```makefile
help:
    @echo -e "Makefile 功能:"
    @echo -e "\tSTO lasso analysis"
    @echo
    @echo -e "Targets:"
    @echo -e "\thelp - 显示此帮助信息"
    @echo -e "\tSPI - 子项目编号, 用于获取内部gef数据"
    @echo -e "\tSAMPLE - 对应子项目的样本名称，用于获取内部gef数据"
    @echo -e "\toutdir - 输出目录"
    @echo

checkPara:
    # 创建分析目录
    @[ -d $(outdir) ] && echo $(outdir) exist || mkdir -p $(outdir)
    
    # 记录用户参数
    @echo "SPI: " $(SPI) >>$(AnalysisDir)/jobs/$(jobId)/log.txt
    @echo "SAMPLE: " $(SAMPLE) >>$(AnalysisDir)/jobs/$(jobId)/log.txt
    @echo "outdir: " $(outdir) >>$(AnalysisDir)/jobs/$(jobId)/log.txt
    
makeshell:
    python3 $(BIN)/generate_shell.py -SPI $(SPI) -SAMPLE $(SAMPLE) -OUTDIR $(outdir)
    # generate work.sh

dowork:
    sh $(outdir)/work.sh
 
Main:checkPara makeshell dowork
```

4. 添加到 stools：
```bash
./stools add ./my-shell-tool "A simple Shell tool"
```

5. 执行工具：
```bash
./stools my-shell-tool
./stools my-shell-tool SPI=pm-xxx-01 SAMPLE=Case outdir=/path/outdir
```

## 注意事项

1. **工具脚本权限**：确保工具脚本具有执行权限
2. **工具名称唯一性**：工具名称必须唯一，添加同名工具会覆盖现有工具
3. **路径安全**：工具路径不应包含特殊字符，建议使用简单的英文名称
4. **配置文件**：`conf.yaml` 文件会自动创建在程序同目录下

## 许可证

本项目采用 MIT 许可证。

## 作者
- yfinddream@gmail.com
- author: 苑赞
- 版本：1.0.9
- Jun 19 2025

## 更新版本
```bash
git add -A
git commit -m "mes"
git tag v1.1.1
git push origin main
git push origin v1.1.1
```