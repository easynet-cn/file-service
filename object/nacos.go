package object

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/easynet-cn/file-service/util"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/common/file"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	Config *viper.Viper
)

func InitNacos() {
	Config = viper.New()

	if flagSet, err := getFlatSet(); err != nil {
		panic(err)
	} else if err := Config.BindPFlags(flagSet); err != nil {
		panic(err)
	}

	Config.SetConfigType("yml")

	configPath, configName := getLocalConfigPathAndName(Config)

	Config.AddConfigPath(configPath)
	Config.SetConfigName(configName)
	Config.SetConfigType("yml")

	if err := Config.ReadInConfig(); err != nil {
		panic(err)
	}

	serverConfig, clientConfig := getNacosConfig(Config)

	// Create config client and get remote config
	if configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfig,
		},
	); err != nil {
		log.Fatalf("Failed to create config client: %v", err)
	} else if remoteConfig, err := getRemoteConfig(configClient, Config); err != nil {
		log.Fatalf("Failed to get remote config: %v", err)
	} else {
		keys := remoteConfig.AllKeys()

		for _, key := range keys {
			Config.Set(key, remoteConfig.Get(key))
		}
	}

	// Create naming client and register service
	if namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfig,
		},
	); err != nil {
		log.Fatalf("Failed to create naming client: %v", err)
	} else if success, err := registerService(namingClient, Config); err != nil {
		log.Fatalf("Failed to register service: %v", err)
	} else if !success {
		log.Fatalf("Failed to register service")
	}

	registerNacoseServices()
}

func getFlatSet() (*pflag.FlagSet, error) {
	flagSet := pflag.NewFlagSet("system", pflag.ContinueOnError)

	flagSet.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{UnknownFlags: true}
	flagSet.SortFlags = false

	flagSet.String("spring.config.location", "", "config location")
	flagSet.String("spring.profiles.active", "dev", "active profile")
	flagSet.String("nacos.namespace", "", "nacos namespace")
	flagSet.String("nacos.host", "127.0.0.1", "nacos host")
	flagSet.Int("nacos.port", 8848, "nacos port")
	flagSet.String("nacos.context-path", "nacos", "nacos context path")
	flagSet.String("nacos.group", "group", "nacos group")
	flagSet.String("nacos.username", "", "nacos username")
	flagSet.String("nacos.password", "", "nacos password")
	flagSet.Int("server.port", 6103, "server port")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	return flagSet, nil
}

func getNacosConfig(config *viper.Viper) ([]constant.ServerConfig, constant.ClientConfig) {
	serverConfig := []constant.ServerConfig{
		{
			IpAddr:      config.GetString("nacos.host"),
			Port:        config.GetUint64("nacos.port"),
			ContextPath: config.GetString("nacos.context-path"),
		},
	}

	clientConfig := constant.ClientConfig{
		NamespaceId:         config.GetString("nacos.namespace"),
		Username:            config.GetString("nacos.username"),
		Password:            config.GetString("nacos.password"),
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogLevel:            "warn",
	}

	return serverConfig, clientConfig
}

func getLocalConfigPathAndName(config *viper.Viper) (string, string) {
	configPath := "./"
	configName := "application"

	configLocation := config.GetString("spring.config.location")

	if configLocation != "" {
		if file.IsExistFile(configLocation) {
			configPath = filepath.Dir(configLocation)
			configName = strings.TrimSuffix(filepath.Base(configLocation), filepath.Ext(configLocation))
		}
	} else {
		activeProfile := config.GetString("spring.profiles.active")

		if activeProfile != "" && file.IsExistFile(path.Join("./", fmt.Sprintf("application-%s.yml", activeProfile))) {
			configName = fmt.Sprintf("application-%s", activeProfile)
		}
	}

	return configPath, configName
}

func getRemoteConfig(configClient config_client.IConfigClient, config *viper.Viper) (*viper.Viper, error) {
	remoteConfig := viper.New()

	remoteConfig.SetConfigType("yml")

	if content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: fmt.Sprintf("%s-%s.yml", config.GetString("spring.application.name"), config.GetString("spring.profiles.active")),
		Group:  config.GetString("nacos.group")}); err != nil {

		return nil, err
	} else if err := remoteConfig.ReadConfig(bytes.NewBuffer([]byte(content))); err != nil {
		return nil, err
	}

	return remoteConfig, nil
}

func registerService(namingClient naming_client.INamingClient, config *viper.Viper) (bool, error) {
	serviceName := config.GetString("nacos.service-name")

	if serviceName == "" {
		serviceName = config.GetString("spring.application.name")
	}

	return namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          util.ExternalIP().String(),
		Port:        config.GetUint64("server.port"),
		ServiceName: serviceName,
		Weight:      1,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata: map[string]string{
			"preserved.register.source": "SPRING_CLOUD",
			"goVersion":                 runtime.Version(),
			"version":                   Version,
		},
	})
}

func registerNacoseServices() {
	for k := range Config.GetStringMap("nacos.services") {
		//创建 serverConfig
		serverConfig := []constant.ServerConfig{
			{
				IpAddr:      Config.GetString(fmt.Sprintf("nacos.services.%s.host", k)),
				Port:        Config.GetUint64(fmt.Sprintf("nacos.services.%s.port", k)),
				ContextPath: Config.GetString(fmt.Sprintf("nacos.services.%s.context-path", k)),
			},
		}

		// 创建clientConfig
		clientConfig := constant.ClientConfig{
			NamespaceId:         Config.GetString(fmt.Sprintf("nacos.services.%s.namespace", k)),
			Username:            Config.GetString(fmt.Sprintf("nacos.services.%s.username", k)),
			Password:            Config.GetString(fmt.Sprintf("nacos.services.%s.password", k)),
			TimeoutMs:           5000,
			NotLoadCacheAtStart: true,
			LogLevel:            "warn",
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

		serviceName := Config.GetString(fmt.Sprintf("nacos.services.%s.service-name", k))

		if serviceName == "" {
			serviceName = Config.GetString("spring.application.name")
		}

		// 服务注册
		success, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
			Ip:          util.ExternalIP().String(),
			Port:        Config.GetUint64("server.port"),
			ServiceName: serviceName,
			Weight:      1,
			Enable:      true,
			Healthy:     true,
			Ephemeral:   true,
			Metadata: map[string]string{
				"preserved.register.source": "SPRING_CLOUD",
				"goVersion":                 runtime.Version(),
				"version":                   Version,
			},
		})

		if !success || err != nil {
			log.Fatalf("初始化nacos服务注册失败: %s", err.Error())
		}
	}
}
