/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package bootstrap

import (
	"github.com/azzill/goze/balancer"
	"github.com/azzill/goze/cache"
	"github.com/azzill/goze/config"
	"github.com/azzill/goze/context"
	"github.com/azzill/goze/log"
	"github.com/azzill/goze/server"
	"github.com/azzill/goze/sql"
	"github.com/azzill/goze/util"
	"net/http"
	"time"
)

const (
	configFileName           = "server.yaml"
	defServerAddr            = ":8080"
	defRedisAddress          = ""
	defRedisNetwork          = "tcp"
	defRedisPassword         = ""
	defRedisWriteTimeout     = 5
	defRedisReadTimeout      = 5
	defRedisConnectTimeout   = 15
	defRedisDb               = 0
	defMicroServiceName      = "Unnamed"
	defMicroServiceWeight    = 0
	defMicroServiceAddress   = ""
	defEnableDiscoverServer  = false
	defEnableDiscoverClient  = false
	defHttpReadTimeout       = 10
	defHttpReadHeaderTimeout = 10
	defHttpIdleTimeout       = 5
	defHttpWriteTimeout      = 10
	defWRRBalancerTimeout    = 10
	defBalancerRule          = balancer.WeightedRoundRobinRule
	defSQLDataSource         = ""
	defSQLDriverName         = ""
)

type Configuration struct {
	Server       RestConfiguration
	Cache        RedisConfiguration
	MicroService MicroServiceConfiguration
	LoadBalance  LoadBalanceConfiguration
	SQL          SQLConfiguration
}

type MicroServiceConfiguration struct {
	EnableDiscoverServer bool
	EnableDiscoverClient bool
	DiscoverServerAddr   string
	ServiceName          string
	//Weighted round robin only
	Weight uint
}

type LoadBalanceConfiguration struct {
	LoadBalanceRule balancer.LoadBalanceRule

	//Timeout for WRR Balancer http client
	WRRBalancerTimeout time.Duration
}

type RedisConfiguration struct {
	Address        string
	Network        string
	Password       string
	WriteTimeout   time.Duration
	ReadTimeout    time.Duration
	ConnectTimeout time.Duration
	Db             int
}

type RestConfiguration struct {
	ServerAddr        string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
}

type SQLConfiguration struct {
	DataSource string
	DriverName string
}

var logger = log.NewLogger("Loader")

func StartGozeApplication(components ...interface{}) {
	logger.Info("Goze loader is starting...")
	ctx := Bootstrap(configFileName)
	for _, comp := range components {
		ctx.With(comp)
	}
	ctx.Inject()
	ctx.GetComponent("RestServer").(*server.RestServer).StartServer()
}

func Bootstrap(configPath string) *context.ApplicationContext {
	commonCfg := config.NewCommonConfiguration(configPath)
	cfg := loadConfiguration(commonCfg)
	ctx := &context.ApplicationContext{Components: make(map[string]interface{}), Configuration: commonCfg}

	httpConfig := &server.HttpConfig{}
	e := util.Map(&cfg.Server, httpConfig)
	if e != nil {
		panic(e.Error())
	}
	//port, e := strconv.Atoi(strings.Split(cfg.Server.ServerAddr, ":")[1])
	//if e != nil {
	//	panic(e.Error())
	//}

	restServer := server.NewRestServer(cfg.Server.ServerAddr, httpConfig)
	//microService := discover.NewWeightedMicroService(cfg.MicroService.ServiceName,
	//	uint(port), cfg.MicroService.Weight)
	redis := cache.NewRedisClient(cfg.Cache.Network, cfg.Cache.Address, cfg.Cache.Password, cfg.Cache.WriteTimeout,
		cfg.Cache.ReadTimeout, cfg.Cache.ConnectTimeout, cfg.Cache.Db)

	ctx.With(restServer).With(redis) //TODO ++ MicroService Instance Manager or LoadBalancer?
	if cfg.SQL.DataSource != "" {
		ctx.With(sql.NewSQL(cfg.SQL.DataSource, cfg.SQL.DriverName))
	}

	return ctx
}

func loadConfiguration(cfg *config.CommonConfiguration) *Configuration {

	configs := &Configuration{
		//Server:       RestConfiguration{},
		//Cache:        RedisConfiguration{},
		//MicroService: MicroServiceConfiguration{},
		//LoadBalance:  LoadBalanceConfiguration{},
	}
	//Rest server
	configs.Server.ServerAddr = cfg.DefaultGet("goze.server.address", defServerAddr).(string)
	configs.Server.ReadTimeout = time.Duration(cfg.DefaultGet("goze.server.read-timeout", defHttpReadTimeout).(int)) * time.Second
	configs.Server.ReadHeaderTimeout = time.Duration(cfg.DefaultGet("goze.server.read-header-timeout", defHttpReadHeaderTimeout).(int)) * time.Second
	configs.Server.WriteTimeout = time.Duration(cfg.DefaultGet("goze.server.write-timeout", defHttpWriteTimeout).(int)) * time.Second
	configs.Server.IdleTimeout = time.Duration(cfg.DefaultGet("goze.server.idle-timeout", defHttpIdleTimeout).(int)) * time.Second
	configs.Server.MaxHeaderBytes = cfg.DefaultGet("goze.server.max-header-bytes", http.DefaultMaxHeaderBytes).(int)

	//Cache
	configs.Cache.Address = cfg.DefaultGet("goze.cache.redis.address", defRedisAddress).(string)
	configs.Cache.Network = cfg.DefaultGet("goze.cache.redis.network", defRedisNetwork).(string)
	configs.Cache.Db = cfg.DefaultGet("goze.cache.redis.db", defRedisDb).(int)
	configs.Cache.ReadTimeout = time.Duration(cfg.DefaultGet("goze.cache.redis.read-timeout", defRedisReadTimeout).(int)) * time.Second
	configs.Cache.WriteTimeout = time.Duration(cfg.DefaultGet("goze.cache.redis.write-timeout", defRedisWriteTimeout).(int)) * time.Second
	configs.Cache.ConnectTimeout = time.Duration(cfg.DefaultGet("goze.cache.redis.connect-timeout", defRedisConnectTimeout).(int)) * time.Second
	configs.Cache.Password = cfg.DefaultGet("goze.cache.redis.password", defRedisPassword).(string)

	//Micro service
	configs.MicroService.ServiceName = cfg.DefaultGet("goze.micro-service.name", defMicroServiceName).(string)
	configs.MicroService.DiscoverServerAddr = cfg.DefaultGet("goze.micro-service.address", defMicroServiceAddress).(string)
	configs.MicroService.Weight = uint(cfg.DefaultGet("goze.micro-service.weight", defMicroServiceWeight).(int))
	configs.MicroService.EnableDiscoverClient = cfg.DefaultGet("goze.micro-service.enable-client",
		defEnableDiscoverClient).(bool)
	configs.MicroService.EnableDiscoverClient = cfg.DefaultGet("goze.micro-service.enable-server",
		defEnableDiscoverServer).(bool)
	//TODO ++ MS Config (Heartbeat, Timeout etc.)
	//Load balancer
	configs.LoadBalance.LoadBalanceRule = balancer.LoadBalanceRule(cfg.DefaultGet("goze.balancer.rule", defBalancerRule).(string))
	configs.LoadBalance.WRRBalancerTimeout = time.Duration(cfg.DefaultGet("goze.balancer.wrr.client.timeout", defWRRBalancerTimeout).(int)) * time.Second

	//SQL
	configs.SQL.DataSource = cfg.DefaultGet("goze.sql.datasource", defSQLDataSource).(string)
	configs.SQL.DriverName = cfg.DefaultGet("goze.sql.driver", defSQLDriverName).(string)

	return configs
}
