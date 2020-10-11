package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"jjz.io/xscp/auth"
	"path"
	"path/filepath"
	"runtime"

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
	for scanner.Scan() {
		host := scanner.Text()
		// 跳过空行和注释
		if host == "" || strings.HasPrefix(host, "#") {
			continue
		}
		if len(strings.Split(host, ":")) == 1 {
			host = host + ":22"
		}

		if isCli {
			result, err = doShell(host, &clientConfig, shell)
		} else {
			err = doScp(host, &clientConfig, flag.Arg(0), flag.Arg(1))
		}

		if err == nil {
			fmt.Printf("[SUCCESS] %s \n", host)
			if len(result) > 0 {
				fmt.Println(result)
			}
		} else {
			fmt.Printf("[FAILURE] %s exited with %s \n", host, err.Error())
		}
	}

}

func parseArgs() {
	hostFileEnv := os.Getenv("XSCP_HOST_FILE")
	flag.StringVar(&hostFile, "f", hostFileEnv, "hosts file")

	/*host := flag.String("h", "", "host")
	overwrite := flag.Bool("o", false, "overwrite if exist")
	copyDir := flag.Bool("r", false, "recusively client directory")*/

	priKeyEnv := os.Getenv("XSCP_PRI_KEY")
	flag.StringVar(&priKey, "k", priKeyEnv, "private key")

	usernameEnv := os.Getenv("XSCP_USERNAME")
	flag.StringVar(&username, "u", usernameEnv, "username")

	flag.BoolVar(&isCli, "c", false, "exec shell")

	flag.Parse()

	if !isCli && flag.NArg() != 2 {
		flag.PrintDefaults()
		return
	}

	if isCli {
		shell = strings.Join(flag.Args(), " ")
	}

}

func doScp(host string, clientConfig *ssh.ClientConfig, localTarget string, remotePath string) error {
	client := auth.NewClient(host, clientConfig)
	defer client.Close()

	err := client.Connect()
	if err != nil {
		return err
	}

	f, err := os.Open(localTarget)
	if err != nil {
		return err
	}
	defer f.Close()

	var filename string
	switch runtime.GOOS {
	case "windows":
		filename = path.Base(filepath.ToSlash(localTarget))
	case "linux":
		filename = path.Base(localTarget)
	}

	// todo 目前只支持向linux服务器拷贝
	if []rune(remotePath)[len(remotePath)-1] != '/' {
		remotePath = remotePath + "/"
	}

	// todo per
	err = client.CopyFile(f, remotePath+filename, "0655")
	if err != nil {
		return err
	}

	return nil
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
