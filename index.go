package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
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

// 读取json的函数
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
			email, ok1 := configContent["email"].(string)
			apikey, ok2 := configContent["apikey"].(string)
			if ok1 && ok2 {
				//fmt.Println("合法")
			} else {
				log.Fatal("传入的Config不合法")
			}
			cloudflareDomainList(email, apikey)

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

// 列出cloudflare内的所有域名
func cloudflareDomainList(email string, apikey string) []interface{} {
	result := request(email, apikey, "zones", "GET", "")["result"].([]interface{})
	return result
}

// cloudflare 请求封装
func request(email string, apikey string, api string, method string, params string) map[string]interface{} {
	url := "https://api.cloudflare.com/client/v4/" + api

	client := &http.Client{}

	payload := strings.NewReader(params)

	req, err := http.NewRequest(method, url, payload)

	if err != nil {

		return request(email, apikey, api, method, params)
	}
	req.Header.Add("X-Auth-Email", email)
	req.Header.Add("X-Auth-Key", apikey)
	res, err := client.Do(req)
	if err != nil {
		return request(email, apikey, api, method, params)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return request(email, apikey, api, method, params)
	}
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return result
}

// 修改cloudflare解析记录
func cloudflareChangeDns(email string, apikey string, targets []string, target_ip string) {
	domainList := cloudflareDomainList(email, apikey)
	for i := 0; i < len(targets); i++ {
		targetDomain := targets[i]
		result := cloudflareCheckChildDomain(email, apikey, "", domainList)
		_, notFound := result["notFound"]
		if notFound {
			cloudflareAddDNS(email, apikey, targetDomain, target_ip, "AAAA", result["domainId"].(string), "false")
		} else {
			fmt.Println(result["id"])
		}
	}

}

// Cloudflare :: 添加解析
func cloudflareAddDNS(email string, apikey string, domain string, ip string, domainType string, domainId string, proxied string) bool {
	params := fmt.Sprintf("{\"type\":\"%s\",\"name\":\"%s\",\"content\":\"%s\",\"ttl\":120,\"priority\":10,\"proxied\":%s}", domainType, domain, ip, proxied)
	result := request(email, apikey, "zones/"+domainId+"/dns_records", "POST", params)
	if result["success"] == true {
		return true
	} else {
		return false
	}
}

// Cloudflare :: 检查是否有子域名的解析存在
func cloudflareCheckChildDomain(email string, apikey string, childDomain string, domains []interface{}) map[string]interface{} {

	for i := 0; i < len(domains); i++ {
		domainName := domains[i].(map[string]interface{})["name"].(string)
		if strings.Contains(childDomain, domainName) {
			domainId := domains[i].(map[string]interface{})["id"].(string)
			result := clouodflareGetChildDomain(email, apikey, domainId)
			for i := 0; i < len(result); i++ {
				itemDomain := result[i].(map[string]interface{})
				childDomainItem := itemDomain["name"]
				if childDomain == childDomainItem {
					return itemDomain
				}
			}
			resultdomain := make(map[string]interface{})
			resultdomain["notFound"] = true
			resultdomain["domainId"] = domainId
			return resultdomain
		}
	}
	result := make(map[string]interface{})
	result["notFound"] = true

	return result
}

// Cloudflare :: 获取子域名
func clouodflareGetChildDomain(email string, apikey string, domainId string) []interface{} {
	requestUrl := "zones/" + domainId + "/dns_records"
	result := request(email, apikey, requestUrl, "GET", "")["result"].([]interface{})
	return result
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
