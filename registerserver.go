package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	types "github.com/NpoolAccounting/service-register/types"
	"github.com/NpoolDevOps/fbc-auth-service/authapi"
	authTypes "github.com/NpoolDevOps/fbc-auth-service/types"
	etcdcli "github.com/NpoolDevOps/fbc-license-service/etcdcli"
	httpdaemon "github.com/NpoolRD/http-daemon"
	"github.com/google/uuid"
	"io/ioutil"
	"net/http"
	"strings"
	_ "strings"
	_ "time"
)

// etcd key
const accountingDomain = "accounting.npool.top"

const prometheusDomain = "prometheus.npool.top"

type RegisterConfig struct {
	Port int
}

type RegisterServer struct {
	config RegisterConfig
}

func NewRegisterServer(configFile string) *RegisterServer {

	buf, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot read file %v: %v", configFile, err)
		return nil
	}

	config := RegisterConfig{}
	err = json.Unmarshal(buf, &config)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot parse file %v: %v", configFile, err)
		return nil
	}

	server := &RegisterServer{
		config: config,
	}

	log.Infof(log.Fields{}, "successful to create service register server")

	return server
}

func (s *RegisterServer) Run() error {

	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: types.ServiceRegisterAPI,
		Method:   "POST",
		Handler: func(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
			return s.ServiceRegisterRequest(w, req)
		},
	})

	log.Infof(log.Fields{}, "start http daemon at %v", s.config.Port)
	httpdaemon.Run(s.config.Port)
	return nil
}

func (s *RegisterServer) ServiceRegisterRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {

	ip := req.RemoteAddr
	ip = ip[0:strings.LastIndex(ip, ":")]
	fmt.Printf("ip:%v", ip)

	// 解析请求参数
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err.Error(), -1
	}
	input := types.ServiceRegisterInput{}
	err = json.Unmarshal(b, &input)
	fmt.Println(input)
	if err != nil {
		return nil, err.Error(), -1
	}
	sha256Password := sha256.Sum256([]byte(input.Password))
	password := fmt.Sprintf("%v", sha256Password)
	password = password[0:12]
	// 登录
	// username password  login
	userLoginInput := authTypes.UserLoginInput{
		Username:  input.UserName,
		Password:  password,
		AppId:     uuid.MustParse("00000003-0003-0003-0003-000000000003"),
		TargetUrl: "",
	}
	_, err = authapi.Login(userLoginInput)
	if err != nil {
		return nil, err.Error(), -1
	}

	// 判断 域名是否在域名数组里面
	domainArr := []string{accountingDomain, prometheusDomain}
	result := in(input.DomainName, domainArr)
	// 不存在
	if !result {
		return nil, "domainName is permission denied", -1
	}
	resp, err := etcdcli.Get(input.DomainName)
	if err != nil && resp != nil {
		log.Errorf(log.Fields{}, "cannot get %v: %v", accountingDomain, err)
		return "", err.Error(), -1
	}

	if resp != nil {
		s2 := types.ServiceRegisterOutput{
			IP:   ip,
			Port: input.Port,
		}
		s2info, _ := json.Marshal(s2) //转换成JSON返回的是byte[]
		vals := append(resp, s2info)
		strs := ""
		for i, v := range vals {
			if 0 < i {
				strs = fmt.Sprintf("%v,", strs)
			}
			strs = fmt.Sprintf("%v%v", strs, string(v))
		}
		// put json string
		info, err := etcdcli.Put(input.DomainName, strs)
		if err != nil {
			return nil, err.Error(), -1
		}
		return info, "success", 0
	} else {
		// put server & port
		servcerInfo := types.ServiceRegisterOutput{
			IP:   input.IP,
			Port: input.Port,
		}
		jsons, _ := json.Marshal(servcerInfo) //转换成JSON返回的是byte[]

		info, err := etcdcli.Put(input.DomainName, string(jsons))
		if err != nil {
			return nil, err.Error(), -1
		}
		return info, "success", 0
	}

}

func in(target string, str_array []string) bool {
	for _, element := range str_array {
		if target == element {
			return true
		}
	}
	return false
}
