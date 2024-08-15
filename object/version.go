package object

import (
	"github.com/golang-module/carbon/v2"
)

const (
	Version = "0.0.1"
)

type SystemVersion struct {
	ApplicationName string `json:"applicationName"`
	Version         string `json:"version"`
	Profile         string `json:"profile"`
	GoVersion       string `json:"goVersion"`
	CurrentTime     string `json:"currentTime"`
}

func NewSystemVersion() *SystemVersion {
	return &SystemVersion{
		ApplicationName: Nacos.GetConfig().GetString("spring.application.name"),
		Version:         Version,
		Profile:         Nacos.GetConfig().GetString("spring.profiles.active"),
		GoVersion:       Version,
		CurrentTime:     carbon.Now().ToDateTimeString(),
	}
}
