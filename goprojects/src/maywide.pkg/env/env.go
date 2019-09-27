// env
package env

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/axgle/mahonia"
	"github.com/larspensjo/config"
)

var __Father__ string
var __Child__ []*os.Process
var ChildCount int
var iniconf *config.Config

func init() {
	// if child, do not create child again.
	__Father__ = os.Getenv("__FATHER__")
	myfather := fmt.Sprint(os.Getppid())
	if __Father__ == "" {
		__Father__ = fmt.Sprint(os.Getpid())
		os.Setenv("__FATHER__", __Father__)
		if len(os.Args) > 2 {
			pStr := ""
			l := len(os.Args) - 1
			for i := 0; i < len(os.Args[l]); i++ {
				c := os.Args[l][i : i+1]
				if c >= "0" && c <= "9" {
					pStr = pStr + c
				} else {
					os.Args[l] = os.Args[l][i:]
					break
				}
			}
			if pStr == "" {
				ChildCount = 0
			} else {
				ChildCount, _ = strconv.Atoi(pStr)
			}

			for i := 0; i < ChildCount; i++ {
				NewChild()
			}
		}
	} else {
		if myfather != __Father__ {
			os.Exit(0)
		}
	}
}

func NewChild() (*os.Process, error) {
	mypid := fmt.Sprint(os.Getpid())
	if __Father__ != mypid {
		return nil, nil
	}

	ChildAttr := &os.ProcAttr{
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}

	// fork child after parameters checked.
	Child, err := os.StartProcess(os.Args[0], os.Args, ChildAttr)
	if err != nil {
		return nil, err
	}

	__Child__ = append(__Child__, Child)
	return Child, nil
}

func KillChild() {
	for _, v := range __Child__ {
		v.Kill()
		v.Release()
	}
}

func IsChild() bool {
	mypid := fmt.Sprint(os.Getpid())
	if __Father__ != mypid {
		return true
	}
	return false
}

func IsFather() bool {
	mypid := fmt.Sprint(os.Getpid())
	if __Father__ == mypid {
		return true
	}
	return false
}

func Iputenv(initfile string, section string) (cfg *config.Config, err error) {
	cfg, err = config.ReadDefault(initfile)
	if err != nil {
		cfg, err = config.ReadDefault("/usr/lib/lines/" + initfile)
		if err != nil {
			fmt.Printf("%s not found: %s\n", initfile, err.Error())
			os.Exit(-1)
		}
	}
	iniconf = cfg

	//加载环境变量
	sec, err := cfg.SectionOptions(section)
	if err == nil {
		for _, v := range sec {
			options, err := cfg.String(section, v)
			if err != nil {
				fmt.Printf("read %s %s for env failed\n", section, v)
				os.Exit(-1)
			} else {
				newv := mahonia.NewDecoder("gbk").ConvertString(v)
				if runtime.GOOS == "windows" {
					os.Setenv(newv, options)
				} else {
					os.Setenv(v, options)
				}
			}
		}
	}
	return
}

func Ivalue(cfg *config.Config, section string, options string) (value string, err error) {
	value, err = cfg.String(section, options)
	if err != nil {
		fmt.Printf("Cannot get node : %s.%s\n", section, options)
		os.Exit(-1)
	}

	if strings.Contains(value, "$(ivalue") {
		tmpstr := strings.Replace(value, "$(ivalue ", "", -1)
		tmpstr = strings.Replace(tmpstr, ")", "", -1)
		cmdarg := strings.Split(tmpstr, " ")
		cmd := exec.Command("ivalue", cmdarg...)
		buf, err := cmd.Output()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(-1)
		}
		value = strings.TrimSpace(string(buf))
	}
	return
}

func Open(initfile string) (cfg *config.Config) {
	cfg, err := config.ReadDefault(initfile)
	if err != nil {
		cfg, err = config.ReadDefault("/usr/lib/lines/" + initfile)
		if err != nil {
			fmt.Printf("%s not found: %s\n", initfile, err.Error())
			os.Exit(-1)
		}
	}
	iniconf = cfg
	return
}

func Value(section string, options string) string {
	if iniconf == nil {
		fmt.Printf("<inifile> was not been Iputenv(file,section) or Iopen(file)\n")
		os.Exit(-1)
	}

	value, err := iniconf.String(section, options)
	if err != nil {
		fmt.Printf("Cannot get node : %s.%s\n", section, options)
		os.Exit(-1)
	}

	if strings.Contains(value, "$(ivalue") {
		tmpstr := strings.Replace(value, "$(ivalue ", "", -1)
		tmpstr = strings.Replace(tmpstr, ")", "", -1)
		cmdarg := strings.Split(tmpstr, " ")
		cmd := exec.Command("ivalue", cmdarg...)
		buf, err := cmd.Output()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(-1)
		}
		value = strings.TrimSpace(string(buf))
	}
	return value
}

func Exists(section string, options string) bool {
	if iniconf == nil {
		fmt.Printf("<inifile> was not been Iputenv(file,section) or Iopen(file)\n")
		return false
	}

	_, err := iniconf.String(section, options)
	if err != nil {
		return false
	}

	return true
}
