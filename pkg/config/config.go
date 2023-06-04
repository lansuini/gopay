package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"io/ioutil"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var Instance *Config

type Config struct {
	AppName     string `yaml:"AppName"`
	GlobalJsVer string `yaml:"GlobalJsVer"`
	Env         string `yaml:"Env"`        // 环境：prod、dev
	BaseUrl     string `yaml:"BaseUrl"`    // base url
	Port        string `yaml:"Port"`       // 端口
	LogPath     string `yaml:"LogPath"`    // 日志文件
	LogFile     string `yaml:"LogFile"`    // 日志文件
	StaticPath  string `yaml:"StaticPath"` // 静态文件目录
	DataSalt    string `yaml:"DataSalt"`   // 数据加密salt
	DataSaltIV  string `yaml:"DataSaltIV"` // 数据加密iv
	PsSalt      string `yaml:"PsSalt"`     // 密码加密salt
	// 数据库配置
	DB struct {
		Url          string `yaml:"Url"`
		MaxIdleConns int    `yaml:"MaxIdleConns"`
		MaxOpenConns int    `yaml:"MaxOpenConns"`
	} `yaml:"DB"`

	// Github
	Github struct {
		ClientID     string `yaml:"ClientID"`
		ClientSecret string `yaml:"ClientSecret"`
	} `yaml:"Github"`

	// OSChina
	OSChina struct {
		ClientID     string `yaml:"ClientID"`
		ClientSecret string `yaml:"ClientSecret"`
	} `yaml:"OSChina"`

	// QQ登录
	QQConnect struct {
		AppId  string `yaml:"AppId"`
		AppKey string `yaml:"AppKey"`
	} `yaml:"QQConnect"`

	// 阿里云oss配置
	Uploader struct {
		Enable    string `yaml:"Enable"`
		AliyunOss struct {
			Host          string `yaml:"Host"`
			Bucket        string `yaml:"Bucket"`
			Endpoint      string `yaml:"Endpoint"`
			AccessId      string `yaml:"AccessId"`
			AccessSecret  string `yaml:"AccessSecret"`
			StyleSplitter string `yaml:"StyleSplitter"`
			StyleAvatar   string `yaml:"StyleAvatar"`
			StylePreview  string `yaml:"StylePreview"`
			StyleSmall    string `yaml:"StyleSmall"`
			StyleDetail   string `yaml:"StyleDetail"`
		} `yaml:"AliyunOss"`
		Local struct {
			Host string `yaml:"Host"`
			Path string `yaml:"Path"`
		} `yaml:"Local"`
	} `yaml:"Uploader"`

	// 百度SEO相关配置
	// 文档：https://ziyuan.baidu.com/college/courseinfo?id=267&page=2#h2_article_title14
	BaiduSEO struct {
		Site  string `yaml:"Site"`
		Token string `yaml:"Token"`
	} `yaml:"BaiduSEO"`

	// 神马搜索SEO相关
	// 文档：https://zhanzhang.sm.cn/open/mip
	SmSEO struct {
		Site     string `yaml:"Site"`
		UserName string `yaml:"UserName"`
		Token    string `yaml:"Token"`
	} `yaml:"SmSEO"`

	// smtp
	Smtp struct {
		Host     string `yaml:"Host"`
		Port     string `yaml:"Port"`
		Username string `yaml:"Username"`
		Password string `yaml:"Password"`
		SSL      bool   `yaml:"SSL"`
	} `yaml:"Smtp"`

	// es
	Es struct {
		Url   string `yaml:"Url"`
		Index string `yaml:"Index"`
	} `yaml:"Es"`

	// host
	HostSet struct {
		GM_ENABLE     bool   `yaml:"GM_ENABLE"`
		DOMAIN        string `yaml:"DOMAIN"`
		PORT          string `yaml:"PORT"`
		GM_PORT       string `yaml:"GM_PORT"`
		NOTIFY_PORT   string `yaml:"NOTIFY_PORT"`
		MERCHANT_PORT string `yaml:"MERCHANT_PORT"`
		GATE_PORT     string `yaml:"GATE_PORT"`
		CB_PORT       string `yaml:"CB_PORT"`
		GM_IPWHITE    string `yaml:"GM_IPWHITE"`
	} `yaml:"HostSet"`

	// es
	Redis struct {
		NetWork  string `yaml:"NetWork"`
		Addr     string `yaml:"Addr"`
		Password string `yaml:"Password"`
		Db       int    `yaml:"Db"`
	} `yaml:"Redis"`

	SystemConfig struct {
		GATE_IP_PROTECT string `yaml:"GATE_IP_PROTECT"`
		IP_PROTECT_TIME string `yaml:"IP_PROTECT_TIME"`
	} `yaml:"SystemConfig"`

	Crontab struct {
		PAY_NOTIFY_AUTO_PUSH    bool `yaml:"PAY_NOTIFY_AUTO_PUSH"`
		PAY_NOTIFY_TASK         int  `yaml:"PAY_NOTIFY_TASK"`
		SETTLE_NOTIFY_AUTO_PUSH bool `yaml:"SETTLE_NOTIFY_AUTO_PUSH"`
		SETTLE_NOTIFY_TASK      int  `yaml:"SETTLE_NOTIFY_TASK"`
		SETTLE_FETCH_AUTO_PUSH  bool `yaml:"SETTLE_FETCH_AUTO_PUSH"`
		SETTLE_FETCH_TASK       int  `yaml:"SETTLE_FETCH_TASK"`
		SETTLE_FETCH_LIMIT      bool `yaml:"SETTLE_FETCH_LIMIT"`
	} `yaml:"Crontab"`
}

var once sync.Once

// Viper viper global instance
var Viper *viper.Viper

func Init(filename string) *Config {
	Instance = &Config{}
	if yamlFile, err := ioutil.ReadFile(filename); err != nil {
		logrus.Error(err)
	} else if err = yaml.Unmarshal(yamlFile, Instance); err != nil {
		logrus.Error(err)
	}
	once.Do(func() {
		Viper = viper.New()
		// scan the file named config in the root directory
		Viper.AddConfigPath("./")
		Viper.SetConfigName("config")
		Viper.SetConfigType("yaml")
		Viper.Set("verbose", true)
		// read config, if failed, configure by default
		if err := Viper.ReadInConfig(); err == nil {
			logrus.Println("Read config successfully: ", Viper.ConfigFileUsed())
		} else {
			logrus.Printf("Read failed: %s \n", err)
			//panic(err)
		}
	})
	Viper.WatchConfig()
	Viper.OnConfigChange(func(e fsnotify.Event) {
		logrus.Info("Config file changed:", e.Name)
		if yamlFile, err := ioutil.ReadFile(filename); err != nil {
			logrus.Error(err)
		} else if err = yaml.Unmarshal(yamlFile, Instance); err != nil {
			logrus.Error(err)
		}
	})

	return Instance
}
