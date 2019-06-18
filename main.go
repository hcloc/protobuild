package protobuild

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var (
	protoPath 		string
	protoOut		string
)

func init() {
	flag.StringVar(&protoPath, "f", "", "proto源文件路径")
	flag.StringVar(&protoOut, "t", "", "proto文件编译输出路径")
}

func main() {
	flag.Parse()
	if protoPath == "" {
		log.Fatal("proto源文件地址不能为空")
	}
	dirs, err := getAllPkgs(protoPath)
	if err != nil {
		log.Fatalf("读取proto文件子目录发生错误, error:%v", err)
	}
	dirs = append(dirs, protoPath)
	err = compilerProto(dirs)
	if err != nil {
		log.Fatalf("编译proto失败， err:%v", err)
	}
	log.Println("生成完成")
}

func getAllPkgs(pp string) (dirs []string, err error) {
	var (
		dirTemp string
	)
	dir, err := ioutil.ReadDir(pp)
	if err != nil {
		return nil, err
	}
	for _, fi := range dir {
		if fi.IsDir() {
			dirTemp = filepath.Join(pp, fi.Name())
			dirs = append(dirs, dirTemp)
			childDirs, _ := getAllPkgs(dirTemp)
			dirs = append(dirs, childDirs...)
		}
	}
	return dirs, nil
}

func compilerProto(dirs []string) error {
	var (
		goOut string
		errStr string
	)
	for _, dirPath := range dirs {
		dir := strings.Replace(dirPath, protoPath, "", 1)
		if dir != "" {
			goOut = filepath.Join(protoOut, dir[1:])
		}else{
			goOut = protoOut
		}
		_, err := os.Stat(goOut)
		if err != nil {
			if os.IsPermission(err) {
				log.Fatalf("mkdir path:%s, err:%v", goOut, err)
			}
			if os.IsNotExist(err) {
				if err := os.MkdirAll(goOut, os.ModePerm); err != nil{
					log.Fatalf("mkdir path:%s, err:%v", goOut, err)
				}
			}
		}
		if _, err := os.Stat(dirPath); err != nil {
			if os.IsNotExist(err) {
				log.Fatalf("源路径：%s, error:%v", dirPath, err)
			}
		}
		var (
			out bytes.Buffer
			stderr bytes.Buffer
			files []string
		)
		files = getProtoFile(dirPath)
		if len(files) == 0 {
			continue
		}
		protoPathInfo := fmt.Sprintf("-I=%s%s", dirPath, string(os.PathSeparator))
		goOutPath := fmt.Sprintf("--go_out=plugins=grpc:%s", goOut)
		var args = []string{protoPathInfo, goOutPath}
		args = append(args, files...)
		cmd := exec.Command("protoc", args...)
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			errStr = strings.Trim(stderr.String(), " ")
			log.Println(fmt.Sprintf("错误目录:%s 错误描述:%s", protoPath, errStr))
		}
	}
	return nil
}

func getProtoFile(p string) []string {
	var files []string
	var fileName string
	dir, err := ioutil.ReadDir(p)
	if err != nil {
		return files
	}
	for _, fi := range dir {
		if fi.IsDir() {
			continue
		}
		fileName = fi.Name()
		if strings.ToLower(path.Ext(fileName)) != ".proto"  {
			continue
		}
		files = append(files, fileName)
	}
	return files
}

