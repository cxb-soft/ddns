package main

import (
	"fmt"
	"net"
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

func main() {
	fmt.Println("欢迎使用DDNS Tools")
	ipv6 := getMyIPV6()
	fmt.Println("您目前的ipv6:" + ipv6)
}
