package config

import (
	"ReadySetGo/util"
	"fmt"
	"github.com/alexflint/go-arg"
	"net"
	"os"
	"strconv"
)

func init() {
	loadConfig()
	if err := validateConfig(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func GetBinPath() string {
	return conf.BinPath
}

func GetPort() int {
	return conf.Port
}

var conf *config

type config struct {
	BinPath string `arg:"env:RSG_BIN_PATH,required" help:"Where to store uploaded binaries"`
	Port    int    `arg:"env:RSG_PORT,required" help:"Port where the web interface will be available"`
}

func loadConfig() {
	cfg := &config{}
	arg.MustParse(cfg)
	conf = cfg
}

func validateConfig() error {
	if err := validateBinPath(); err != nil {
		return err
	}
	if err := validatePort(); err != nil {
		return err
	}
	return nil
}

func validateBinPath() error {
	isDir, err := util.IsDirectory(conf.BinPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("path to bin directory does not exist")
	}
	if err != nil {
		return fmt.Errorf("unable to validate bin path: %w", err)
	}
	if !isDir {
		return fmt.Errorf("given bin directory path is not a directory")
	}
	isWritable, err := util.IsWritable(conf.BinPath)
	if err != nil {
		return fmt.Errorf("unable to validate bin path: %w", err)
	}
	if !isWritable {
		return fmt.Errorf("cannot write to bin directory")
	}

	return nil
}

func validatePort() error {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(conf.Port))

	if err != nil {
		return fmt.Errorf("unable to listen to port %d: %w", conf.Port, err)
	}
	_ = ln.Close()
	return nil
}
