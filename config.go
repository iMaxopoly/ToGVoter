package main
import (
	"os"
	"fmt"
	"strconv"
	"bufio"
	"github.com/fogcreek/mini"
)

var D_PROXY_LIST []string
const D_CAPTCHA_URL = "http://topofgames.com/imageverify.php"
var Target string
var ConfigStruct struct {
	ProxyListFilepath string `json:"proxy_list_filepath"`
	TargetID          int `json:"target_id"`
	WorkerCount       int    `json:"worker_count"`
	DBCUsername       string     `json:"dbc_username"`
	DBCPassword       string     `json:"dbc_password"`
	ProxyType          string   `json:"proxy_type"`
	Timeout       int    `json:"timeout"`
}

func LoadConfig() error {
	//read config
	configFile, err := mini.LoadConfiguration("settings.ini")
	if err != nil {
		return err
	}
	ConfigStruct.ProxyListFilepath = configFile.String("proxy_list_path","proxies.txt")
	ConfigStruct.TargetID = int(configFile.Integer("target_id",80397))
	ConfigStruct.WorkerCount = int(configFile.Integer("worker_count", 20))
	ConfigStruct.DBCUsername = configFile.String("deathbycaptcha_username","username")
	ConfigStruct.DBCPassword = configFile.String("deathbycaptcha_password","password")
	ConfigStruct.ProxyType = configFile.String("proxy_type","http")
	ConfigStruct.Timeout = int(configFile.Integer("timeout", 20))

	Target = "http://topofgames.com/index.php?do=votes&id="+strconv.Itoa(ConfigStruct.TargetID)
	return nil

}

func LoadProxies() error{
	//load proxies
	if _, err := os.Stat(ConfigStruct.ProxyListFilepath); err != nil {
		err := fmt.Errorf("Proxy file parsing error.")
		return err
	}
	proxyFile, err := os.Open(ConfigStruct.ProxyListFilepath)
	if err != nil {
		return err
	}
	defer proxyFile.Close()
	scanner := bufio.NewScanner(proxyFile)
	for scanner.Scan() {
		D_PROXY_LIST = append(D_PROXY_LIST, scanner.Text())
	}
	return nil
}