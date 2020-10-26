package copy

import (
	"context"
	"fmt"
)

type Option func(c *ServiceInfo)

type ConfigInfo struct {
	Routes []Route
}

type ServiceInfo struct {
	Name     string               `json:"name"`
	AppID    string               `json:"appId"`
	Scheme   string               `json:"scheme"`
	Address  string               `json:"address"`
	Weight   string               `json:"weight"`
	Enable   string               `json:"enable"`
	Healthy  string               `json:"healthy"`
	Metadata map[string]string    `json:"metadata"`
	Region   string               `json:"region"`
	Zone     string               `json:"zone"`
	Kind     constant.ServiceKind `json:"kind"`
	// Deployment 部署组：不同组的流量隔离
	// 比如某些服务给内部调用和第三方调用，可以配置不同的deployment，进行流量隔离
	Deployment string `json:"deployment"`
	//Group 流量组：流量在Group之间进行负载均衡
	Group    string              `json:"group"`
	Services map[string]*Service `json:"services" toml:"services"`
}

type Service struct {
	Namespace string            `json:"namespace" toml:"namespace"`
	Name      string            `json:"name" toml:"name"`
	Labels    map[string]string `json:"labels" toml:"labels"`
	Methods   []string          `json:"methods" toml:"methods"`
}

func (si ServiceInfo) Label() string {
	return fmt.Sprintf("%s://%s", si.Scheme, si.Address)
}

type Server interface {
	Serve() error
	Stop() error
	GracefulStop(ctx context.Context) error
	Info() *ServiceInfo
}

type Route struct {
	//权重组
	WeightGroups []WeightGroup
	//方法名
	Method string
}

type WeightGroup struct {
	Group  string
	Weight int
}

func ApplyOptions(options ...Option)ServiceInfo {
}

func defultServiceInfo()ServiceInfo  {
	si:=ServiceInfo{
		Name: pkg.
	}
}
