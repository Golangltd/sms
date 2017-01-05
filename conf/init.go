package conf

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"git.kissdata.com/ycfm/common/utils"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/smallnest/rpcx"
	"github.com/smallnest/rpcx/clientselector"
	"github.com/smallnest/rpcx/codec"
)

var (
	// rpcx
	OfficialAccountClient, AccountClient *rpcx.Client
	Servers                              []string
	RpcAddr, EtcdAddr                    string

	// db::mysql
	DBHost    string
	DBPort    int
	DBUser    string
	DBPawd    string
	DBName    string
	DBCharset string
	DBTimeLoc string
	DBMaxIdle int
	DBMaxConn int
	DBDebug   bool
)

func initRpcEnv() {
	EtcdAddr = strings.TrimSpace(beego.AppConfig.String("etcd::address"))
	RpcAddr = strings.TrimSpace(beego.AppConfig.String("rpc::address"))
	if EtcdAddr == "" || RpcAddr == "" {
		panic("params `etcd::address || rpc::address` empty")
	}
	serverTemp := beego.AppConfig.String("rpc::servers")
	Servers = strings.Split(serverTemp, ",")
	return
}

func connRpcClient(appName string) (client *rpcx.Client) {
	s := clientselector.NewEtcdClientSelector([]string{EtcdAddr}, fmt.Sprintf("/%s/%s/%s", beego.BConfig.RunMode, "rpcx", appName), time.Minute, rpcx.RandomSelect, time.Minute)
	client = rpcx.NewClient(s)
	client.FailMode = rpcx.Failover
	client.ClientCodecFunc = codec.NewProtobufClientCodec
	return
}

func init() {
	var (
		err   error
		exist bool = false
		name  string
	)

	// 初始化rpc短信服务
	initRpcEnv()
	if name, exist = utils.FindServer("accounts", Servers); !exist {
		panic("params `account` service not exist")
	}
	AccountClient = connRpcClient(name)
	if name, exist = utils.FindServer("official_accounts", Servers); !exist {
		panic("params `official_accounts` service not exist")
	}
	OfficialAccountClient = connRpcClient(name)

	// 1. connect mysql
	DBHost = strings.TrimSpace(beego.AppConfig.String("db::host"))
	if "" == DBHost {
		panic("app parameter `db::host` empty")
	}

	DBPort, err = beego.AppConfig.Int("db::port")
	if err != nil {
		panic("app parameter `db::port` error")
	}
	DBUser = strings.TrimSpace(beego.AppConfig.String("db::user"))
	if "" == DBUser {
		panic("app parameter `db::user` empty")
	}

	DBPawd = strings.TrimSpace(beego.AppConfig.String("db::pawd"))
	if "" == DBPawd {
		panic("app parameter `db::pawd` empty")
	}

	DBName = strings.TrimSpace(beego.AppConfig.String("db::name"))
	if "" == DBName {
		panic("app parameter `db::name` empty")
	}

	DBCharset = strings.TrimSpace(beego.AppConfig.String("db::charset"))
	if "" == DBCharset {
		panic("app parameter `db::charset` empty")
	}

	DBTimeLoc = strings.TrimSpace(beego.AppConfig.String("db::time_loc"))
	if "" == DBTimeLoc {
		panic("app parameter `db::time_loc` empty")
	}

	DBMaxIdle, err = beego.AppConfig.Int("db::max_idle")
	if err != nil {
		panic("app parameter `db::max_idle` error")
	}

	DBMaxConn, err = beego.AppConfig.Int("db::max_conn")
	if err != nil {
		panic("app parameter `db::max_conn` error")
	}
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&loc=%s", DBUser, DBPawd, DBHost, DBPort, DBName, DBCharset, url.QueryEscape(DBTimeLoc))

	err = orm.RegisterDataBase("default", "mysql", dataSourceName, DBMaxIdle, DBMaxConn)
	if err != nil {
		panic(err)
	}
	// orm debug
	DBDebug, err = beego.AppConfig.Bool("dev::debug")
	if err != nil {
		panic("app parameter `dev::debug` error:" + err.Error())
	}
	if DBDebug {
		orm.Debug = true
	}
	return
}
