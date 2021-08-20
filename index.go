package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strings"
)

func getMyIPV6() string {
	s, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range s {
		i := regexp.MustCompile(`(\w+:){7}\w+`).FindString(a.String())
		if strings.Count(i, ":") == 7 {
			return i
		}
	}
	return ""
}

type Config struct {
	Name  string
	Value string
}

func readJson(jsonPath string) {
	jsonFile, err := os.Open(jsonPath)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var config map[string]interface{}
	json.Unmarshal([]byte(byteValue), &config)
	fmt.Println(config)
}

// 对用户的命令行参数进行处理
func commandLineProcess(args []string) {
	arglen := len(args)
	if arglen > 1 {
		commands := args[1:]
		for i := 0; i < (arglen - 1); i++ {
			spliceCommand := commands[i]
			switch spliceCommand {
			case "-service":
				serviceName := commands[i+1]
				fmt.Println(serviceName)
				i = i + 1
				break
			}
		}
	} else {
		userChoose()
	}
}

func userChoose() {

}

func main() {
	fmt.Println("欢迎使用DDNS Tools")
	ipv6 := getMyIPV6()
	fmt.Println("您目前的ipv6:" + ipv6)
	//readJson("config.json")
	commandLineProcess(os.Args)
}
