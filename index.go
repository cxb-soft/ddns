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
	Config         map[string]interface{} // 配置
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
		config.Config = make(map[string]interface{})
		for i := 0; i < (arglen - 1); i++ {
			spliceCommand := commands[i]
			switch spliceCommand {
			case "-service":
				if i == arglen-2 {
					log.Fatal("命令行参数不正确")
				}
				serviceName := commands[i+1]
				config.ServiceName = serviceName
				config.commandOptions = commands
				resultMap["service"] = serviceName
				fmt.Println("当前服务:" + config.ServiceName)
				i = i + 1
				break
			case "-cfemail":
				if i == arglen-2 {
					log.Fatal("命令行参数不正确")
				}
				commandConfig := commands[i+1]
				config.Config["cfemail"] = commandConfig
				resultMap["config"] = commandConfig
				i = i + 1
				break
			case "-cfapikey":
				if i == arglen-2 {
					log.Fatal("命令行参数不正确")
				}
				commandConfig := commands[i+1]
				config.Config["cfapikey"] = commandConfig
				resultMap["config"] = commandConfig
				i = i + 1
				break
			case "-domainList":
				if i == arglen-2 {
					log.Fatal("命令行参数不正确")
				}
				var commandConfig []interface{}
				commandConfig1 := strings.Split(commands[i+1], ",")
				for i := 0; i < len(commandConfig1); i++ {
					commandConfig = append(commandConfig, commandConfig1[i])
				}
				config.Config["domainList"] = commandConfig
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
	if jsonConfig["cloudflare"] == nil {
		return false
	}
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
		structConfig := config.Config
		localConfig := readJson("config.json")
		configExist := checkCloudflareConfig(localConfig)
		commandOptions := config.commandOptions
		if in("-cfemail", commandOptions) && in("-cfapikey", commandOptions) && in("-domainList", commandOptions) {
			/*email, ok1 := configContent["email"].(string)
			apikey, ok2 := configContent["apikey"].(string)
			if ok1 && ok2 {
				//fmt.Println("合法")
			} else {
				log.Fatal("传入的Config不合法")
			}*/
			email := structConfig["cfemail"].(string)
			apikey := structConfig["cfapikey"].(string)

			target_domain := structConfig["domainList"].([]interface{})

			cloudflareChangeDns(email, apikey, target_domain, getMyIPV6())

		} else {
			if configExist {
				email := localConfig["cloudflare"].(map[string]interface{})["email"].(string)
				apikey := localConfig["cloudflare"].(map[string]interface{})["apikey"].(string)
				target_domain := localConfig["cloudflare"].(map[string]interface{})["domainList"].([]interface{})
				cloudflareChangeDns(email, apikey, target_domain, getMyIPV6())
			}
		}
	}
}

// 列出cloudflare内的所有域名
func cloudflareDomainList(email string, apikey string) []interface{} {
	result1 := request(email, apikey, "zones", "GET", "")["result"]
	if result1 == nil {
		log.Fatal("Cloudflare账户配置错误")
	}
	result := result1.([]interface{})
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
func cloudflareChangeDns(email string, apikey string, targets []interface{}, target_ip string) {
	domainList := cloudflareDomainList(email, apikey)

	for i := 0; i < len(targets); i++ {
		targetDomain := targets[i].(string)
		result := cloudflareCheckChildDomain(email, apikey, targetDomain, domainList)
		_, notFound := result["notFound"]
		if notFound {
			if result["domainId"] == nil {
				log.Fatal("域名获取失败")
			}
			cloudflareAddDNS(email, apikey, targetDomain, target_ip, "AAAA", result["domainId"].(string), "false")
		} else {
			cloudflareChangeDNS(email, apikey, targetDomain, target_ip, "AAAA", result["zone_id"].(string), result["id"].(string), "false")
		}
		fmt.Println("已将 " + targetDomain + " 解析到 " + target_ip)
	}

}

// Cloudflare :: 修改解析
func cloudflareChangeDNS(email string, apikey string, domain string, ip string, domainType string, domainId string, childDomainId string, proxied string) bool {
	requstUrl := "zones/" + domainId + "/dns_records/" + childDomainId
	params := fmt.Sprintf("{\"type\":\"%s\",\"name\":\"%s\",\"content\":\"%s\",\"ttl\":0,\"priority\":10,\"proxied\":%s}", domainType, domain, ip, proxied)
	result := request(email, apikey, requstUrl, "PUT", params)
	if result["success"] == true {
		return true
	} else {
		return false
	}
}

// Cloudflare :: 添加解析
func cloudflareAddDNS(email string, apikey string, domain string, ip string, domainType string, domainId string, proxied string) bool {
	params := fmt.Sprintf("{\"type\":\"%s\",\"name\":\"%s\",\"content\":\"%s\",\"ttl\":0,\"priority\":10,\"proxied\":%s}", domainType, domain, ip, proxied)
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
	fmt.Print("选择服务:\n1.Cloudflare\n输入序号:")
	var number string
	fmt.Scanf("%s", &number)
	localConfig := readJson("config.json")
	switch number {
	case "1":
		fmt.Println("您已选择Cloudflare")
		isConfig := checkCloudflareConfig(localConfig)
		useConfig := "n"
		if isConfig {
			fmt.Print("检测到已有配置,是否使用(y/n):")
			fmt.Scanf("%s", &useConfig)
		}
		if useConfig == "y" {
			email := localConfig["cloudflare"].(map[string]interface{})["email"].(string)
			apikey := localConfig["cloudflare"].(map[string]interface{})["apikey"].(string)
			target_domain := localConfig["cloudflare"].(map[string]interface{})["domainList"].([]interface{})
			cloudflareChangeDns(email, apikey, target_domain, getMyIPV6())
		} else {
			var email string
			var apikey string
			var domainListStr string
			fmt.Print("Cloudflare Email:")
			fmt.Scanf("%s", &email)
			fmt.Print("Cloudflare API KEY:")
			fmt.Scanf("%s", &apikey)
			fmt.Print("域名(如果有多个逗号隔开):")
			fmt.Scanf("%s", &domainListStr)
			var domainList []string
			domainList = strings.Split(domainListStr, ",")
			target_domain := string2interface(domainList)
			cloudflareChangeDns(email, apikey, target_domain, getMyIPV6())
		}
		break
	default:
		fmt.Println("您已选择Cloudflare")
		isConfig := checkCloudflareConfig(localConfig)
		useConfig := "n"
		if isConfig {
			fmt.Print("检测到已有配置,是否使用(y/n):")
			fmt.Scanf("%s", &useConfig)
		}
		if useConfig == "y" {
			email := localConfig["cloudflare"].(map[string]interface{})["email"].(string)
			apikey := localConfig["cloudflare"].(map[string]interface{})["apikey"].(string)
			target_domain := localConfig["cloudflare"].(map[string]interface{})["domainList"].([]interface{})
			cloudflareChangeDns(email, apikey, target_domain, getMyIPV6())
		}
	}
}

// string 转 interface
func string2interface(origin []string) []interface{} {
	var result []interface{}
	for i := 0; i < len(origin); i++ {
		result = append(result, origin[i])
	}
	return result
}

// 检查配置文件
func configCheck() {
	_, err := os.Stat("./config.json")
	if err == nil {

	}
	if os.IsNotExist(err) {
		newFile, err := os.Create("config.json")
		if err != nil {
			log.Fatal(err)
		}
		_ = newFile.Close()
		file, err := os.OpenFile(
			"config.json",
			os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
			0666,
		)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		// 写字节到文件中
		_, err = file.WriteString("{}")

		// 写文件字符串到文件
		//
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	fmt.Println("欢迎使用DDNS Tools")
	ipv6 := getMyIPV6()
	fmt.Println("您目前的ipv6:" + ipv6)
	//readJson("config.json")
	configCheck()
	commandLineProcess(os.Args)
}
