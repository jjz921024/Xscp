package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"jjz.io/xscp/auth"
	"jjz.io/xscp/utils"
	"path"
	"path/filepath"
	"runtime"
	"sync"

	"golang.org/x/crypto/ssh"
	"os"
	"strings"
)

var (
	hostFile string
	priKey   string
	username string
	isCli    bool
	shell    string
)

func main() {
	// 解析参数
	parseArgs()

	// 打开hosts文件
	hosts, err := os.Open(hostFile)
	if err != nil {
		fmt.Printf("Exit with open host file error: %s\n", err.Error())
		return
	}

	clientConfig, _ := auth.PrivateKey(username, priKey, ssh.InsecureIgnoreHostKey())

	var result string
	scanner := bufio.NewScanner(hosts)
	var hostList []string
	for scanner.Scan() {
		host := scanner.Text()
		// 跳过空行和注释
		if host == "" || strings.HasPrefix(host, "#") {
			continue
		}
		// 默认22端口
		if len(strings.Split(host, ":")) == 1 {
			host = host + ":22"
		}
		hostList = append(hostList, host)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(hostList))

	greed := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	for _, host := range hostList {
		host := host
		go func() {
			if isCli {
				result, err = doShell(host, &clientConfig, shell)
			} else {
				result, err = doScp(host, &clientConfig, flag.Arg(0), flag.Arg(1))
			}

			if err == nil {
				fmt.Printf("%s - %s \n%s\n", greed("[SUCCESS]"), yellow(host), result)
			} else {
				fmt.Printf("%s - %s \nexited with %s\n", red("[FAILURE]"), yellow(host), err.Error())
			}

			wg.Done()
		}()
	}

	wg.Wait()
}

func parseArgs() {
	hostFileEnv := os.Getenv("XSCP_HOST_FILE")
	flag.StringVar(&hostFile, "f", hostFileEnv, "hosts file")

	priKeyEnv := os.Getenv("XSCP_PRI_KEY")
	flag.StringVar(&priKey, "k", priKeyEnv, "private key")

	usernameEnv := os.Getenv("XSCP_USERNAME")
	flag.StringVar(&username, "u", usernameEnv, "username")

	// 执行一条命令
	flag.BoolVar(&isCli, "c", false, "exec shell")

	/*host := flag.String("h", "", "host")
	overwrite := flag.Bool("o", false, "overwrite if exist")*/

	flag.Parse()

	if !isCli && flag.NArg() != 2 {
		flag.PrintDefaults()
		return
	}

	if isCli {
		shell = strings.Join(flag.Args(), " ")
	}

}

func doScp(host string, clientConfig *ssh.ClientConfig, localFile string, remotePath string) (string, error) {
	// 获取文件
	file, err := os.Open(localFile)
	if err != nil {
		return "", err
	}

	stat, err := file.Stat()
	if err != nil {
		return "", err
	}
	if stat.IsDir() {
		return "", errors.New("不支持远程复制目录")
	}

	// ssh连接
	client := auth.NewClient(host, clientConfig)
	defer client.Close()

	err = client.Connect()
	if err != nil {
		return "", err
	}

	// todo: 远程目录不存在则创建
	// todo: 复制目录
	// 获取文件名
	var filename string
	switch runtime.GOOS {
	case "windows":
		filename = path.Base(filepath.ToSlash(localFile))
	case "linux":
		filename = path.Base(localFile)
	}

	// todo 目前只支持向linux服务器拷贝
	if []rune(remotePath)[len(remotePath)-1] != '/' {
		remotePath = remotePath + "/"
	}

	err = client.CopyFile(file, remotePath+filename, utils.ConvertPerm(stat.Mode().String()))
	if err != nil {
		return "", err
	}

	hint := "copy: " + file.Name() + " --> " + remotePath + filename + "\n"
	return hint, nil
}

func doShell(host string, clientConfig *ssh.ClientConfig, shell string) (string, error) {
	client := auth.NewClient(host, clientConfig)
	defer client.Close()

	err := client.Connect()
	if err != nil {
		return "", err
	}

	// todo 执行类似sudo命令时要求输入密码，需提供交互式shell
	buf, err := client.Session.CombinedOutput(shell)
	if err != nil {
		return "", errors.New(string(buf))
	}

	return string(buf), nil
}
