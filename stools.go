package main

import (
	"bytes"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)


func samestringlen(inst string, lens int) string {
	minus := lens - len(inst)
	if minus > 0 {
		inst = inst + strings.Repeat(" ", minus)
	}
	return inst
}

func checkErr(err error) {
	if err != nil {
		//panic(err)
		log.Fatal(err)
	}
}

func usage() {
	toolName := filepath.Base(os.Args[0])
	fmt.Println(fmt.Sprintf("version: 1.0.9"))
	fmt.Println(fmt.Sprintf("Usage:   %s  <tool> [parameters]", toolName))
	fmt.Println(fmt.Sprintf("         %s  rm <toolname>", toolName))
	fmt.Println(fmt.Sprintf("         %s  add <toolpath> <description>", toolName))
	
	// 读取 conf.yaml 文件
	file, _ := exec.LookPath(os.Args[0])
	filepaths, _ := filepath.Abs(file)
	bin := filepath.Dir(filepaths)
	confPath := filepath.Join(bin, "conf.yaml")
	
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		os.Exit(1)
	}
	fmt.Println("\nAvailable tools:")

	// 读取 YAML 文件
	data, err := os.ReadFile(confPath)
	if err != nil {
		fmt.Printf("Error reading conf.yaml: %v\n", err)
		os.Exit(1)
	}
	
	// 解析 YAML 为 Node 以保持顺序
	var node yaml.Node
	err = yaml.Unmarshal(data, &node)
	if err != nil {
		fmt.Printf("Error parsing conf.yaml: %v\n", err)
		os.Exit(1)
	}
	
	// 打印工具信息（按 YAML 文件中的顺序）
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		rootNode := node.Content[0]
		if rootNode.Kind == yaml.MappingNode {
			for i := 0; i < len(rootNode.Content); i += 2 {
				if i+1 < len(rootNode.Content) {
					tool := rootNode.Content[i].Value
					description := rootNode.Content[i+1].Value
					fmt.Println(fmt.Sprintf("\t%s\t%s", samestringlen(tool, 21), description))
				}
			}
		}
	}
	os.Exit(1)
}

func spaceString(raw string) (out string) {
	reg := regexp.MustCompile(`\s`)
	if len(reg.FindAllString(raw, -1)) > 0 {
		out = "\"" + raw + "\""
	} else {
		out = raw
	}
	return
}

func script_command(bin string, flags []string) {
	tool_path, err := filepath.Abs(fmt.Sprintf("%s/module/%s/tool", bin, flags[0]))
	if err != nil {
		log.Fatal("Error getting absolute path:", err)
	}

	_, err = os.Stat(tool_path)
	if os.IsNotExist(err) {
		usage()
	}

	command_line := fmt.Sprintf("%s ", tool_path)
	if len(os.Args) > 2 {
		for _, k := range os.Args[2:] {
			command_line = command_line + " " + spaceString(k)
		}

	}

	fmt.Println(command_line)
	run_cmd(command_line)
}

func run_cmd(command_line string) {
	args := strings.Fields(command_line)

	cmd := exec.Command(args[0], args[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Start()
	if err != nil {
		log.Fatal(string(stderr.Bytes()), err)
	}
	//log.Printf("Waiting for command to finish...")
	// Start
	err = cmd.Wait()
	fmt.Println(string(stdout.Bytes()))

	if err != nil {
		log.Fatal(string(stderr.Bytes()), err)
	}
}

func rmtool(bin, toolname string) {
	// 删除工具目录
	os.RemoveAll(fmt.Sprintf("%s/module/%s", bin, toolname))
	
	// 从 conf.yaml 中删除工具及其描述
	confPath := filepath.Join(bin, "conf.yaml")
	
	// 检查文件是否存在
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		// 文件不存在，直接返回
		return
	}
	
	// 读取现有的 conf.yaml 文件
	data, err := os.ReadFile(confPath)
	if err != nil {
		log.Fatal("Error reading conf.yaml:", err)
	}
	
	// 解析 YAML
	var tools map[string]string
	err = yaml.Unmarshal(data, &tools)
	if err != nil {
		log.Fatal("Error parsing conf.yaml:", err)
	}
	
	// 删除指定工具
	delete(tools, toolname)
	
	// 将更新后的内容写回文件
	data, err = yaml.Marshal(tools)
	if err != nil {
		log.Fatal("Error marshaling yaml:", err)
	}
	
	err = os.WriteFile(confPath, data, 0644)
	if err != nil {
		log.Fatal("Error writing conf.yaml:", err)
	}
}

func addtool(bin, toolpath, description string) {

	tool_file := filepath.Join(toolpath, "tool")
	if _, err := os.Stat(tool_file); os.IsNotExist(err) {
		log.Fatal("tool file not exists")
	}

	_, err := os.Stat(bin)
	if os.IsNotExist(err) {
		os.MkdirAll(bin, 0755)
	}
	_, err = os.Stat(toolpath)
	if os.IsNotExist(err) {
		log.Fatal("toolpath not exists")
	}
	
	_, err = os.Stat(fmt.Sprintf("%s/module", bin))
	if os.IsNotExist(err) {
		os.MkdirAll(fmt.Sprintf("%s/module", bin), 0755)
	}

	toolname := filepath.Base(toolpath)
	if _, err := os.Stat(fmt.Sprintf("%s/module/%s", bin, toolname)); err == nil {
		os.RemoveAll(fmt.Sprintf("%s/module/%s", bin, toolname))
	}
	
	cp_cmd := fmt.Sprintf("cp -rL %s %s/module/%s", toolpath, bin, toolname)
	run_cmd(cp_cmd)

	// 给 tool 文件增加可执行权限
	toolFile := filepath.Join(bin, "module", toolname, "tool")
	err = os.Chmod(toolFile, 0755)
	if err != nil {
		log.Printf("Warning: Failed to set executable permission on %s: %v", toolFile, err)
	}

	// 读取现有的 conf.yaml 文件
	confPath := filepath.Join(bin, "conf.yaml")
	var tools map[string]string
	
	// 检查文件是否存在
	if _, err := os.Stat(confPath); err == nil {
		// 文件存在，读取现有内容
		data, err := os.ReadFile(confPath)
		if err != nil {
			log.Fatal("Error reading conf.yaml:", err)
		}
		
		err = yaml.Unmarshal(data, &tools)
		if err != nil {
			log.Fatal("Error parsing conf.yaml:", err)
		}
	} else {
		// 文件不存在，创建新的 map
		tools = make(map[string]string)
	}
	
	// 添加新工具到 map 中
	tools[toolname] = description
	
	// 将更新后的内容写回文件
	data, err := yaml.Marshal(tools)
	if err != nil {
		log.Fatal("Error marshaling yaml:", err)
	}
	
	err = os.WriteFile(confPath, data, 0644)
	if err != nil {
		log.Fatal("Error writing conf.yaml:", err)
	}
}

func main() {
	flag.Parse()
	file, _ := exec.LookPath(os.Args[0])
	filepaths, err := filepath.Abs(file)
	if err != nil {
		log.Fatal("Error getting absolute path:", err)
	}
	bin := filepath.Dir(filepaths)
	
	switch len(flag.Args()) {
	case 0:
		usage()
	case 1:
		switch flag.Args()[0] {
		case "at", "rm":
			usage()
		default:
			script_command(bin, flag.Args())
		}
	case 2:
		switch flag.Args()[0] {
		case "rm":
			rmtool(bin, flag.Args()[1])
		default:
			script_command(bin, flag.Args())
		}
	case 3:
		switch flag.Args()[0] {
		case "add":
			addtool(bin, flag.Args()[1], flag.Args()[2])
		default:
			script_command(bin, flag.Args())
		}

	default:
		script_command(bin, flag.Args())
	}
}


