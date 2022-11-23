package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/viper"

	"github.com/sirupsen/logrus"
)

type Config struct {
	DB       DBConfig       `json:"db"`
	Platform PlatformConfig `json:"platform"`
	Arweave  ArweaveConfig  `json:"arweave"`
	Sqs      SqsConfig      `json:"sqs"`
}

type DBConfig struct {
	Host     string `json:"host"`
	Port     uint   `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
	TZ       string `json:"tz"`
}

type PlatformConfig struct {
	Twitter  TwitterPlatformConfig  `json:"twitter"`
	Ethereum EthereumPlatformConfig `json:"ethereum"`
	Discord  DiscordPlatformConfig  `json:"discord"`
}

type ArweaveConfig struct {
	Jwk       string `json:"jwk"`
	ClientUrl string `json:"client_url"`
}

type SqsConfig struct {
	QueueName string `json:"queue_name"`
}

type TwitterPlatformConfig struct {
	AccessToken       string `json:"access_token"`
	AccessTokenSecret string `json:"access_token_secret"`
	ConsumerKey       string `json:"consumer_key"`
	ConsumerSecret    string `json:"consumer_secret"`
}

type InstagramPlatformConfig struct {
         AppSecret string `json:"app_secret"`
         AccessToken       string `json:"graph_access_token"`
	 
}

type EthereumPlatformConfig struct {
	RPCServer string `json:"rpc_server"`
}

type CliConfig struct {
	ServerURL  string `json:"server_url"`
	UploadPath string `json:"upload_url"`
	QueryPath  string `json:"query_url"`
}

type DiscordPlatformConfig struct {
	BotToken             string `json:"bot_token"`
	ProofServerChannelID string `json:"proof_server_channel_id"`
}

var (
	C     *Config = new(Config)
	Viper *viper.Viper
)

func Init(configPath string) {
	if C.DB.Host != "" { // Initialized
		return
	}
	configContent, err := ioutil.ReadFile(configPath)
	if err != nil {
		logrus.Fatalf("Error during opening config file! %v", err)
	}

	err = json.Unmarshal(configContent, C)
	if err != nil {
		logrus.Fatalf("Error duriong unmarshaling config file: %v", err)
	}
}

func InitCliConfig() {
	Viper = viper.New()

	Viper.SetConfigName("cli") // config file name without extension
	Viper.SetConfigType("toml")
	//viper.AddConfigPath(".")
	Viper.AddConfigPath("./config/") // config file path
	//viper.AutomaticEnv()             // read value ENV variable

	err := Viper.ReadInConfig()
	if err != nil {
		fmt.Printf("fatal error config file: cli err:%v \n", err)
		os.Exit(1)
	}
}

func GetDatabaseDSN() string {
	template := "host=%s port=%d user=%s password=%s dbname=%s TimeZone=%s sslmode=disable"
	return fmt.Sprintf(template,
		C.DB.Host,
		C.DB.Port,
		C.DB.User,
		C.DB.Password,
		C.DB.DBName,
		C.DB.TZ,
	)
}
