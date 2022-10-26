// Code generated by optiongen. DO NOT EDIT.
// optiongen: github.com/timestee/optiongen

package gittools

import (
	"log"
	"os"
	"sync/atomic"
	"unsafe"
)

// Config should use NewConfig to initialize it
type Config struct {
	RsaPath   string `xconf:"rsa_path" usage:"rsa 绝对路径或者home目录下相对路径"`
	Logger    Logger `xconf:"logger" usage:"日志输出"`
	UserName  string `xconf:"user_name" usage:"config user.name"`
	UserEmail string `xconf:"user_email" usage:"config user.email"`
	Depth     int    `xconf:"depth" usage:"git depth"`
}

// NewConfig new Config
func NewConfig(opts ...ConfigOption) *Config {
	cc := newDefaultConfig()
	for _, opt := range opts {
		opt(cc)
	}
	if watchDogConfig != nil {
		watchDogConfig(cc)
	}
	return cc
}

// ApplyOption apply multiple new option and return the old ones
// sample:
// old := cc.ApplyOption(WithTimeout(time.Second))
// defer cc.ApplyOption(old...)
func (cc *Config) ApplyOption(opts ...ConfigOption) []ConfigOption {
	var previous []ConfigOption
	for _, opt := range opts {
		previous = append(previous, opt(cc))
	}
	return previous
}

// ConfigOption option func
type ConfigOption func(cc *Config) ConfigOption

// WithRsaPath rsa 绝对路径或者home目录下相对路径
func WithRsaPath(v string) ConfigOption {
	return func(cc *Config) ConfigOption {
		previous := cc.RsaPath
		cc.RsaPath = v
		return WithRsaPath(previous)
	}
}

// WithLogger 日志输出
func WithLogger(v Logger) ConfigOption {
	return func(cc *Config) ConfigOption {
		previous := cc.Logger
		cc.Logger = v
		return WithLogger(previous)
	}
}

// WithUserName config user.name
func WithUserName(v string) ConfigOption {
	return func(cc *Config) ConfigOption {
		previous := cc.UserName
		cc.UserName = v
		return WithUserName(previous)
	}
}

// WithUserEmail config user.email
func WithUserEmail(v string) ConfigOption {
	return func(cc *Config) ConfigOption {
		previous := cc.UserEmail
		cc.UserEmail = v
		return WithUserEmail(previous)
	}
}

// WithDepth git depth
func WithDepth(v int) ConfigOption {
	return func(cc *Config) ConfigOption {
		previous := cc.Depth
		cc.Depth = v
		return WithDepth(previous)
	}
}

// InstallConfigWatchDog the installed func will called when NewConfig  called
func InstallConfigWatchDog(dog func(cc *Config)) { watchDogConfig = dog }

// watchDogConfig global watch dog
var watchDogConfig func(cc *Config)

// newDefaultConfig new default Config
func newDefaultConfig() *Config {
	cc := &Config{}

	for _, opt := range [...]ConfigOption{
		WithRsaPath(".ssh/id_rsa"),
		WithLogger(log.New(os.Stdout, "", log.LstdFlags)),
		WithUserName(""),
		WithUserEmail(""),
		WithDepth(1),
	} {
		opt(cc)
	}

	return cc
}

// AtomicSetFunc used for XConf
func (cc *Config) AtomicSetFunc() func(interface{}) { return AtomicConfigSet }

// atomicConfig global *Config holder
var atomicConfig unsafe.Pointer

// onAtomicConfigSet global call back when  AtomicConfigSet called by XConf.
// use ConfigInterface.ApplyOption to modify the updated cc
// if passed in cc not valid, then return false, cc will not set to atomicConfig
var onAtomicConfigSet func(cc ConfigInterface) bool

// InstallCallbackOnAtomicConfigSet install callback
func InstallCallbackOnAtomicConfigSet(callback func(cc ConfigInterface) bool) {
	onAtomicConfigSet = callback
}

// AtomicConfigSet atomic setter for *Config
func AtomicConfigSet(update interface{}) {
	cc := update.(*Config)
	if onAtomicConfigSet != nil && !onAtomicConfigSet(cc) {
		return
	}
	atomic.StorePointer(&atomicConfig, (unsafe.Pointer)(cc))
}

// AtomicConfig return atomic *ConfigVisitor
func AtomicConfig() ConfigVisitor {
	current := (*Config)(atomic.LoadPointer(&atomicConfig))
	if current == nil {
		defaultOne := newDefaultConfig()
		if watchDogConfig != nil {
			watchDogConfig(defaultOne)
		}
		atomic.CompareAndSwapPointer(&atomicConfig, nil, (unsafe.Pointer)(defaultOne))
		return (*Config)(atomic.LoadPointer(&atomicConfig))
	}
	return current
}

// all getter func
func (cc *Config) GetRsaPath() string   { return cc.RsaPath }
func (cc *Config) GetLogger() Logger    { return cc.Logger }
func (cc *Config) GetUserName() string  { return cc.UserName }
func (cc *Config) GetUserEmail() string { return cc.UserEmail }
func (cc *Config) GetDepth() int        { return cc.Depth }

// ConfigVisitor visitor interface for Config
type ConfigVisitor interface {
	GetRsaPath() string
	GetLogger() Logger
	GetUserName() string
	GetUserEmail() string
	GetDepth() int
}

// ConfigInterface visitor + ApplyOption interface for Config
type ConfigInterface interface {
	ConfigVisitor
	ApplyOption(...ConfigOption) []ConfigOption
}
