package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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

func readJson(jsonPath string) map[string]interface{} {
	jsonFile, err := os.Open(jsonPath)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var config map[string]interface{}
	json.Unmarshal([]byte(byteValue), &config)
	return config
}

// 命令行结构体
type Args struct {
	ServiceName    string                 // DDNS服务商
	Config         string                 // 配置
	commandOptions []string               // 命令行参数
	resultMap      map[string]interface{} // 命令行参数变成interface
}

// 对用户的命令行参数进行处理
func commandLineProcess(args []string) {
	arglen := len(args)
	if arglen > 1 {
		var config Args
		commands := args[1:]
		resultMap := make(map[string]interface{})
		for i := 0; i < (arglen - 1); i++ {
			spliceCommand := commands[i]
			switch spliceCommand {
			case "-service":
				serviceName := commands[i+1]
				config.ServiceName = serviceName
				config.commandOptions = commands
				resultMap["service"] = serviceName
				fmt.Println(config.ServiceName)
				i = i + 1
				break
			case "-config":
				commandConfig := commands[i+1]
				resultMap["config"] = commandConfig
				i = i + 1
				break

			}
		}
		config.resultMap = resultMap
		mainProcess(config)
	} else {
		userChoose()
	}
}

//检查cloudflare配置是否存在
func checkCloudflareConfig(jsonConfig map[string]interface{}) bool {
	_, ok1 := jsonConfig["cloudflare"].(map[string]interface{})["email"]
	_, ok2 := jsonConfig["cloudflare"].(map[string]interface{})["apikey"]
	if ok1 && ok2 {
		return true
	} else {
		return false
	}
}

// 判断数组中是否有一个元素
func in(target string, str_array []string) bool {
	for _, element := range str_array {
		if target == element {
			return true
		}
	}
	return false
}

// 处理
func mainProcess(config Args) {
	switch config.ServiceName {
	case "cloudflare":
		localConfig := readJson("config.json")
		configExist := checkCloudflareConfig(localConfig)
		commandOptions := config.commandOptions
		if in("-config", commandOptions) {
			configContentString := config.resultMap["config"].(string)
			var configContent map[string]interface{}
			err := json.Unmarshal([]byte(configContentString), &configContent)
			if err != nil {
				log.Fatal("config.json解析失败")
			}
			_, ok1 := configContent["email"]
			_, ok2 := configContent["apikey"]
			if ok1 && ok2 {
				fmt.Println("合法")
			} else {
				fmt.Println("不合法")
			}
		} else {
			if configExist {
				email := localConfig["cloudflare"].(map[string]interface{})["email"].(string)
				apikey := localConfig["cloudflare"].(map[string]interface{})["apikey"].(string)
				cloudflareChangeDns(email, apikey, commandOptions, "Asd")
			}
		}

		cf_config_string := config.Config
		var cf_config map[string]interface{}
		json.Unmarshal([]byte(cf_config_string), &cf_config)
	}
}

func cloudflareChangeDns(email string, apikey string, targets []string, target_ip string) {
	fmt.Println(email)
	fmt.Println(apikey)
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
