package cli

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

var Viper *viper.Viper

func InitConfig() {
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
