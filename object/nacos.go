package object

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/easynet-cn/file-service/util"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	Config *viper.Viper
)

func InitNacos() {
	Config = viper.New()

	flagSet := pflag.NewFlagSet("system", pflag.ContinueOnError)

	flagSet.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{UnknownFlags: true}
	flagSet.SortFlags = false

	flagSet.String("spring.profiles.active", "dev", "active profile")
	flagSet.String("nacos.namespace", "", "nacos namespace")
	flagSet.String("nacos.host", "127.0.0.1", "nacos host")
	flagSet.Int("nacos.port", 8848, "nacos port")
	flagSet.String("nacos.context-path", "nacos", "nacos context path")
	flagSet.String("nacos.group", "group", "nacos group")
	flagSet.String("nacos.username", "", "nacos username")
	flagSet.String("nacos.password", "", "nacos password")
	flagSet.Int("server.port", 6103, "server port")

	flagSet.Parse(os.Args[1:])

	if err := Config.BindPFlags(flagSet); err != nil {
		panic(err)
	}

	Config.SetConfigName("application")
	Config.SetConfigType("yml")
	Config.AddConfigPath("./")

	if err := Config.ReadInConfig(); err != nil {
		panic(err)
	}

	//创建 serverConfig
	serverConfig := []constant.ServerConfig{
		{
			IpAddr:      Config.GetString("nacos.host"),
			Port:        Config.GetUint64("nacos.port"),
			ContextPath: Config.GetString("nacos.context-path"),
		},
	}

	// 创建clientConfig
	clientConfig := constant.ClientConfig{
		NamespaceId:         Config.GetString("nacos.namespace"),
		Username:            Config.GetString("nacos.username"),
		Password:            Config.GetString("nacos.password"),
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogLevel:            "debug",
	}

	// 创建动态配置客户端
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfig,
		},
	)

	if err != nil {
		log.Fatalf("初始化nacos动态配置客户端失败: %s", err.Error())
	}

	if content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: fmt.Sprintf("%s-%s.yml", Config.GetString("spring.application.name"), Config.GetString("spring.profiles.active")),
		Group:  Config.GetString("nacos.group")}); err != nil {
		log.Fatalf("获取配置文件失败: %s", err.Error())
	} else {
		remoteConfig := viper.New()

		remoteConfig.SetConfigType("yml")

		if err := remoteConfig.ReadConfig(bytes.NewBuffer([]byte(content))); err != nil {
			log.Fatalf("获取配置文件失败: %s", err.Error())
		}

		keys := remoteConfig.AllKeys()

		for _, key := range keys {
			Config.Set(key, remoteConfig.Get(key))
		}
	}

	// 创建服务发现客户端
	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfig,
		},
	)

	if err != nil {
		log.Fatalf("初始化nacos服务发现客户端失败: %s", err.Error())
	}

	// 服务注册
	success, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          util.ExternalIP().String(),
		Port:        Config.GetUint64("server.port"),
		ServiceName: Config.GetString("spring.application.name"),
		Weight:      1,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"preserved.register.source": "SPRING_CLOUD"},
	})

	if !success || err != nil {
		log.Fatalf("初始化nacos服务注册失败: %s", err.Error())
	}
}
