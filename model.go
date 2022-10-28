package  main
//https://blog.csdn.net/pengpengzhou/article/details/105385666?spm=1001.2101.3001.6650.2&utm_medium=distribute.pc_relevant.none-task-blog-2~default~CTRLIST~default-2.no_search_link&depth_1-utm_source=distribute.pc_relevant.none-task-blog-2~default~CTRLIST~default-2.no_search_link
//https://gitee.com/medivh-liu/sessredistore/blob/master/sessredistore.go

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base32"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dchest/captcha"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type RedisConfig struct {
	Network string `yaml:"network"`
	Addr string `yaml:"addr"`
	Port string `yaml:"port"`
	Password string `yaml:"password"`
	Db int `yaml:"db"`
	Pools int `yaml:"pools"`
	MinConns int `yaml:"min_conns"`
}
type EMailConfig struct {
	Vendor string `yaml:"vendor"`
	Host string `yaml:"host"`
	Port int `yaml:"port"`
	Sender string `yaml:"sender"`
	Password string `yaml:"password"`
	Nice string `yaml:"nice"`
	CC string `yaml:"cc"`
}
type MessageConfig struct {
	Vendor string `yaml:"vendor"`
	TokenUrl string `yaml:"token_url"`
	ClientId string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	Retry int `yaml:"retry"`
	Token string `yaml:"token"`
	SendUrl string `yaml:"send_url"`
	Tid string `yaml:"tid"`
	Expires int64 `yaml:"expires"`
}
type GlobalConfig struct {
    Redis RedisConfig `yaml:"redis"`
	EMail EMailConfig `yaml:"email"`
	Message MessageConfig `yaml:"message"`
}

type Settings struct{
	AllowTcpFallbackRelay bool `json:"allowTcpFallbackRelay"`
	PortMappingEnabled 	bool   `json:"portMappingEnabled"`
	PrimaryPort int `json:"primaryPort"`
	SoftwareUpdate string `json:"softwareUpdate"`
	SoftwareUpdateChannel string `json:"softwareUpdateChannel"`
	AllowManagementFrom []string `json:"allowManagementFrom"`
}
type Physical struct {

}
type Virtual struct {

}
type Zerotire struct{
	Physical Physical 	`json:"physical"`
	Virtual	 Virtual	`json:"virtual"`
	Settings Settings `json:"settings"`
}
type AuthToken struct {
	AuthToken string `json:"auth_token"`
}

type BoxStatus struct {
	HostName string
}
type RedisStore struct {
	redisClient *redis.Client
    ctx context.Context
	sessionsStore sessions.Store
	config GlobalConfig
	boxStatus BoxStatus
}
type StatusRespone struct {
	Address string `json:address`
	Clock   int64  `json:clock`
	Config  struct {
		Physical struct{} `json:physical`
		Settings struct {
			AllowTcpFallbackRelay bool   `json:allowTcpFallbackRelay`
			PortMappingEnabled    bool   `json:portMappingEnabled`
			PrimaryPort           int    `json:primaryPort`
			SoftwareUpdate        string `json:softwareUpdate`
			SoftwareUpdateChannel string `json:softwareUpdateChannel`
		} `json:settings`
	} `json:config`
	Online               bool   `json:online`
	PlanetWorldId        int64  `json:planetWorldId`
	PlanetWorldTimestamp int64  `json:planetWorldTimestamp`
	PublicIdentity       string `json:publicIdentity`
	TcpFallbackActive    bool   `json:tcpFallbackActive`
	Version              string `json:version`
	VersionBuild         int    `json:versionBuild`
	VersionMajor         int    `json:versionMajor`
	VersionMinor         int    `json:versionMinor`
	VersionRev           int    `json:versionRev`
}

var Redis *redis.Client
var CS *RedisStore

func GetConfig() GlobalConfig {
	data, err := ioutil.ReadFile("./app.yaml")
	if err != nil {
		eLogger.Fatal(err)
	}
	var config GlobalConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		eLogger.Errorf("配置文件加载错误: ", err)
		eLogger.Fatal(err)
	}

	return config
}
func SetConfig(config GlobalConfig) {
	//转换成yaml字符串类型
	d, err := yaml.Marshal(&config)
	if err != nil {
		eLogger.Errorf("配置文件marshal加载错误: ", err)
		eLogger.Fatal(err)
	}
	//log.Printf("--- config dump:\n%s\n\n", string(d))
	//string(d)

	err = ioutil.WriteFile("./app.yaml",[]byte(d),0777)
	if err != nil {
		eLogger.Errorf("配置文件写入错误: ", err)
		eLogger.Fatal(err)
	}
}
//从 /etc/config/zero/  下读取
func ReadFromJson(src string){
	data,err:= ioutil.ReadFile(src)
	if err != nil {
		eLogger.Errorf("local.conf读取错误: ", err)
		eLogger.Fatal(err)
	}
	newZerotire := &Zerotire{}
	err = json.Unmarshal(data,&newZerotire)
	if err != nil {
		eLogger.Errorf("local.conf反序列化错误: ", err)
		eLogger.Fatal(err)
	}
	fmt.Println(*newZerotire)
}
//写入到 /etc/config/zero/  下
func WriteToJson(src string, ip string){
	zerotire := &Zerotire{
		Physical: Physical{},
		Virtual:  Virtual{},
		Settings: Settings{true, true, 9993,
			"apply", "release", []string{"192.168.216.0/24", "127.0.0.1","192.168.195.0/24"},
		},
	}

	data,err := json.MarshalIndent(zerotire,"","	")	// 第二个表示每行的前缀，这里不用，第三个是缩进符号，这里用tab
	if err != nil {
		eLogger.Errorf("local.conf序列化错误: ", err)
		eLogger.Fatal(err)
	}

	err = ioutil.WriteFile(src,data,0777)
	if err != nil {
		eLogger.Errorf("local.conf写入错误: ", err)
		eLogger.Fatal(err)
	}
	// 在盒子中执行/etc/init.d/zerotier restart
	cmd := exec.Command("/etc/init.d/zerotier", "restart")
	cmd.Run()
}

func WriteRc(){
	//这里路径是"/etc/rc.local"
	filePath := "./conf/rc.local"
	file, err := os.OpenFile(filePath, os.O_CREATE | os.O_TRUNC | os.O_WRONLY, 0777)
	if err != nil {
		fmt.Printf("open file error : %v\n", err)
	}
	defer file.Close()
	var str [4]string
	str[0] = "# Put your custom commands here that should be executed once\r\n"
	str[1] = "# the system init finished. By default this file does nothing.\r\n"
	str[2] = "fsck.ext4 -y /dev/mmcblk0p2\r\n"
	str[3] = "exit 0\r\n"
	writer := bufio.NewWriter(file)

	for _ , str1 := range str{
		writer.WriteString(str1)
	}
	writer.Flush()
}

func rstat_handler(c echo.Context) error{
	Redis.Ping(CS.ctx).Result()
	printRedisPool(Redis.PoolStats())
	fmt.Fprintf(c.Response().Writer, "Hello")
	return nil
}

func printRedisPool(stats *redis.PoolStats) {
	eLogger.Printf("Hits=%d Misses=%d Timeouts=%d TotalConns=%d IdleConns=%d StaleConns=%d\n",
		stats.Hits, stats.Misses, stats.Timeouts, stats.TotalConns, stats.IdleConns, stats.StaleConns)
}

func printRedisOption(opt *redis.Options) {
	eLogger.Printf("Network=%v\n", opt.Network)
	eLogger.Printf("Addr=%v\n", opt.Addr)
	eLogger.Printf("Password=%v\n", opt.Password)
	eLogger.Printf("DB=%v\n", opt.DB)
	eLogger.Printf("MaxRetries=%v\n", opt.MaxRetries)
	eLogger.Printf("MinRetryBackoff=%v\n", opt.MinRetryBackoff)
	eLogger.Printf("MaxRetryBackoff=%v\n", opt.MaxRetryBackoff)
	eLogger.Printf("DialTimeout=%v\n", opt.DialTimeout)
	eLogger.Printf("ReadTimeout=%v\n", opt.ReadTimeout)
	eLogger.Printf("WriteTimeout=%v\n", opt.WriteTimeout)
	eLogger.Printf("PoolSize=%v\n", opt.PoolSize)
	eLogger.Printf("MinIdleConns=%v\n", opt.MinIdleConns)
	eLogger.Printf("MaxConnAge=%v\n", opt.MaxConnAge)
	eLogger.Printf("PoolTimeout=%v\n", opt.PoolTimeout)
	eLogger.Printf("IdleTimeout=%v\n", opt.IdleTimeout)
	eLogger.Printf("IdleCheckFrequency=%v\n", opt.IdleCheckFrequency)
	eLogger.Printf("TLSConfig=%v\n", opt.TLSConfig)

}

func InitRedisStore() *RedisStore {
	config := GetConfig()

	ctx := context.Background()
	Redis = redis.NewClient(&redis.Options{
		Network:  config.Redis.Network,                  //网络类型，tcp or unix，默认tcp
			Addr: config.Redis.Addr + ":" + config.Redis.Port,
			Password: config.Redis.Password,
			DB: config.Redis.Db,

		//连接池容量及闲置连接数量
		PoolSize:     config.Redis.Pools, // 连接池最大socket连接数，默认为4倍CPU数， 4 * runtime.NumCPU
		MinIdleConns: config.Redis.MinConns, //在启动阶段创建指定数量的Idle连接，并长期维持idle状态的连接数不少于指定数量；

		//超时
		DialTimeout:  5 * time.Second, //连接建立超时时间，默认5秒。
		ReadTimeout:  3 * time.Second, //读超时，默认3秒， -1表示取消读超时
		WriteTimeout: 3 * time.Second, //写超时，默认等于读超时
		PoolTimeout:  4 * time.Second, //当所有连接都处在繁忙状态时，客户端等待可用连接的最大等待时长，默认为读超时+1秒。

		//闲置连接检查包括IdleTimeout，MaxConnAge
		IdleCheckFrequency: 60 * time.Second, //闲置连接检查的周期，默认为1分钟，-1表示不做周期性检查，只在客户端获取连接时对闲置连接进行处理。
		IdleTimeout:        5 * time.Minute,  //闲置超时，默认5分钟，-1表示取消闲置超时检查
		MaxConnAge:         0 * time.Second,  //连接存活时长，从创建开始计时，超过指定时长则关闭连接，默认为0，即不关闭存活时长较长的连接

		//命令执行失败时的重试策略
		MaxRetries:      0,                      // 命令执行失败时，最多重试多少次，默认为0即不重试
		MinRetryBackoff: 8 * time.Millisecond,   //每次计算重试间隔时间的下限，默认8毫秒，-1表示取消间隔
		MaxRetryBackoff: 512 * time.Millisecond, //每次计算重试间隔时间的上限，默认512毫秒，-1表示取消间隔

		//可自定义连接函数
		Dialer: func(ctx context.Context, addr string, port string) (net.Conn, error) {
			netDialer := &net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Minute,
			}
			return netDialer.Dial(config.Redis.Network, config.Redis.Addr + ":" + config.Redis.Port)
		},

		//钩子函数
		OnConnect: func(ctx context.Context,conn *redis.Conn) error { //仅当客户端执行命令时需要从连接池获取连接时，如果连接池需要新建连接时则会调用此钩子函数
			fmt.Printf("conn=%v\n", conn)
			return nil
		},

	})
	// 连接测活
	_, err := Redis.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("连接Redis成功")
	printRedisOption(Redis.Options())
	printRedisPool(Redis.PoolStats())
	sstore := NewRedisStore(Redis, []byte("secret-key"))
	host_name,_ := os.Hostname()
	var boxs = BoxStatus{host_name}
	CS = &RedisStore{Redis,ctx, sstore,config, boxs}
	captcha.SetCustomStore(CS)
	return CS
}

func (cs *RedisStore) Set(id string, value []byte) {
	err := cs.redisClient.Set(cs.ctx,id, value, time.Minute*2).Err()
	if err != nil {
		eLogger.Errorf(err.Error())
	}
}

func (cs *RedisStore) Get(id string, clear bool) []byte {
	val, err := cs.redisClient.Get(cs.ctx,id).Result()
	if err != nil {
		eLogger.Errorf(err.Error())
		return []byte("")
	}
	if clear {
		err := cs.redisClient.Del(cs.ctx,id).Err()
		if err != nil {
			eLogger.Errorf(err.Error())
			return []byte("")
		}
	}
	return []byte(val)
}

// Redis Sessions Store

// Amount of time for cookies/redis keys to expire.
var sessionExpire = 86400 * 7

// SessionSerializer provides an interface hook for alternative serializers
type SessionSerializer interface {
	Deserialize(d []byte, ss *sessions.Session) error
	Serialize(ss *sessions.Session) ([]byte, error)
}

// JSONSerializer encode the session map to JSON.
type JSONSerializer struct{}

// Serialize to JSON. Will err if there are unmarshalable key values
func (s JSONSerializer) Serialize(ss *sessions.Session) ([]byte, error) {
	m := make(map[string]interface{}, len(ss.Values))
	for k, v := range ss.Values {
		ks, ok := k.(string)
		if !ok {
			err := fmt.Errorf("Non-string key value, cannot serialize session to JSON: %v", k)
			fmt.Printf("redistore.JSONSerializer.serialize() Error: %v", err)
			return nil, err
		}
		m[ks] = v
	}
	return json.Marshal(m)
}

// Deserialize back to map[string]interface{}
func (s JSONSerializer) Deserialize(d []byte, ss *sessions.Session) error {
	m := make(map[string]interface{})
	err := json.Unmarshal(d, &m)
	if err != nil {
		fmt.Printf("redistore.JSONSerializer.deserialize() Error: %v", err)
		return err
	}
	for k, v := range m {
		ss.Values[k] = v
	}
	return nil
}

// GobSerializer uses gob package to encode the session map
type GobSerializer struct{}

// Serialize using gob
func (s GobSerializer) Serialize(ss *sessions.Session) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(ss.Values)
	if err == nil {
		return buf.Bytes(), nil
	}
	return nil, err
}

// Deserialize back to map[interface{}]interface{}
func (s GobSerializer) Deserialize(d []byte, ss *sessions.Session) error {
	dec := gob.NewDecoder(bytes.NewBuffer(d))
	return dec.Decode(&ss.Values)
}

// RedisStore stores sessions in a redis backend.
type RedisSessionsStore struct {
	clt           *redis.Client
	Codecs        []securecookie.Codec
	Options       *sessions.Options // default configuration
	DefaultMaxAge int               // default Redis TTL for a MaxAge == 0 session
	maxLength     int
	keyPrefix     string
	serializer    SessionSerializer
}

// SetMaxLength sets RediStore.maxLength if the `l` argument is greater or equal 0
// maxLength restricts the maximum length of new sessions to l.
// If l is 0 there is no limit to the size of a session, use with caution.
// The default for a new RediStore is 4096. Redis allows for max.
// value sizes of up to 512MB (http://redis.io/topics/data-types)
// Default: 4096,
func (s *RedisSessionsStore) SetMaxLength(l int) {
	if l >= 0 {
		s.maxLength = l
	}
}

// SetKeyPrefix set the prefix
func (s *RedisSessionsStore) SetKeyPrefix(p string) {
	s.keyPrefix = p
}

// SetSerializer sets the serializer
func (s *RedisSessionsStore) SetSerializer(ss SessionSerializer) {
	s.serializer = ss
}

// SetMaxAge restricts the maximum age, in seconds, of the session record
// both in database and a browser. This is to change session storage configuration.
// If you want just to remove session use your session `s` object and change it's
// `Options.MaxAge` to -1, as specified in
//    http://godoc.org/github.com/gorilla/sessions#Options
//
// Default is the one provided by this package value - `sessionExpire`.
// Set it to 0 for no restriction.
// Because we use `MaxAge` also in SecureCookie crypting algorithm you should
// use this function to change `MaxAge` value.
func (s *RedisSessionsStore) SetMaxAge(v int) {
	var c *securecookie.SecureCookie
	var ok bool
	s.Options.MaxAge = v
	for i := range s.Codecs {
		if c, ok = s.Codecs[i].(*securecookie.SecureCookie); ok {
			c.MaxAge(v)
		} else {
			fmt.Printf("Can't change MaxAge on codec %v\n", s.Codecs[i])
		}
	}
}

// NewRedisStore returns a new RedisStore.
// size: maximum number of idle connections.
func NewRedisStore(conn *redis.Client, keyPairs ...[]byte) *RedisSessionsStore {
	return &RedisSessionsStore{
		clt:    conn,
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: sessionExpire,
		},
		DefaultMaxAge: 60 * 20, // 20 minutes seems like a reasonable default
		maxLength:     4096,
		keyPrefix:     "session_",
		serializer:    GobSerializer{},
	}
}

// Get returns a session for the given name after adding it to the registry.
//
// See gorilla/sessions FilesystemStore.Get().
func (s *RedisSessionsStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a session for the given name without adding it to the registry.
//
// See gorilla/sessions FilesystemStore.New().
func (s *RedisSessionsStore) New(r *http.Request, name string) (*sessions.Session, error) {
	var (
		err error
		ok  bool
	)
	session := sessions.NewSession(s, name)
	// make a copy
	options := *s.Options
	session.Options = &options
	session.IsNew = true
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...)
		if err == nil {
			ok, err = s.load(r.Context(), session)
			session.IsNew = !(err == nil && ok) // not new if no error and data available
		}
	}
	return session, err
}

// Save adds a single session to the response.
func (s *RedisSessionsStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	// Marked for deletion.
	if session.Options.MaxAge <= 0 {
		if err := s.delete(r.Context(), session); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
	} else {
		// Build an alphanumeric key for the redis store.
		if session.ID == "" {
			session.ID = strings.TrimRight(base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)), "=")
		}
		if err := s.save(r.Context(), session); err != nil {
			return err
		}
		encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.Codecs...)
		if err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	}
	return nil
}

// Delete removes the session from redis, and sets the cookie to expire.
//
// WARNING: This method should be considered deprecated since it is not exposed via the gorilla/sessions interface.
// Set session.Options.MaxAge = -1 and call Save instead. - July 18th, 2013
func (s *RedisSessionsStore) Delete(ctx context.Context, r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if err := s.clt.Del(ctx, s.keyPrefix+session.ID).Err(); err != nil {
		return err
	}
	// Set cookie to expire.
	options := *session.Options
	options.MaxAge = -1
	http.SetCookie(w, sessions.NewCookie(session.Name(), "", &options))
	// Clear session values.
	for k := range session.Values {
		delete(session.Values, k)
	}
	return nil
}

// save stores the session in redis.
func (s *RedisSessionsStore) save(ctx context.Context, session *sessions.Session) error {
	b, err := s.serializer.Serialize(session)
	if err != nil {
		return err
	}
	if s.maxLength != 0 && len(b) > s.maxLength {
		return errors.New("SessionStore: the value to store is too big")
	}

	age := time.Duration(session.Options.MaxAge) * time.Second
	if age == 0 {
		age = time.Duration(s.DefaultMaxAge) * time.Second
	}

	return s.clt.SetEX(ctx, s.keyPrefix+session.ID, b, age).Err()
}

// load reads the session from redis.
// returns true if there is a session data in DB
func (s *RedisSessionsStore) load(ctx context.Context, session *sessions.Session) (bool, error) {
	data, err := s.clt.Get(ctx, s.keyPrefix+session.ID).Bytes()
	if err != nil {
		return false, err
	}
	if data == nil {
		return false, nil // no data was associated with this key
	}
	return true, s.serializer.Deserialize(data, session)
}

// delete removes keys from redis if MaxAge<0
func (s *RedisSessionsStore) delete(ctx context.Context, session *sessions.Session) error {
	if err := s.clt.Del(ctx, s.keyPrefix+session.ID).Err(); err != nil {
		return err
	}
	return nil
}
