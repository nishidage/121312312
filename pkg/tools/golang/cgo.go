package golang

import (
	"fmt"
	"strings"

	"arhat.dev/dukkha/pkg/constant"
	"arhat.dev/dukkha/pkg/field"
)

type CGOSepc struct {
	field.BaseField

	// Enable cgo
	Enabled bool `yaml:"enabled"`

	// CGO_CPPFLAGS
	CPPFlags []string `yaml:"cppflags"`

	// CGO_CFLAGS
	CFlags []string `yaml:"cflags"`

	// CGO_CXXFLAGS
	CXXFlags []string `yaml:"cxxflags"`

	// CGO_FFLAGS
	FFlags []string `yaml:"fflags"`

	// CGO_LDFLAGS
	LDFlags []string `yaml:"ldflags"`

	CC string `yaml:"cc"`

	CXX string `yaml:"cxx"`
}

func (c *CGOSepc) getEnv(
	doingCrossCompiling bool,
	mKernel, mArch, hostOS, targetLibc string,
) []string {
	if !c.Enabled {
		return []string{"CGO_ENABLED=0"}
	}

	var ret []string
	ret = append(ret, "CGO_ENABLED=1")

	appendListEnv := func(name string, values, defVals []string) {
		actual := values
		if len(values) == 0 {
			actual = defVals
		}

		if len(actual) != 0 {
			ret = append(ret, fmt.Sprintf("%s=%s", name, strings.Join(actual, " ")))
		}
	}

	appendEnv := func(name, value, defVal string) {
		actual := value
		if len(value) == 0 {
			actual = defVal
		}

		if len(actual) != 0 {
			ret = append(ret, fmt.Sprintf("%s=%s", name, actual))
		}
	}

	var (
		cppflags []string
		cflags   []string
		cxxflags []string
		fflags   []string
		ldflags  []string

		cc  = "gcc"
		cxx = "g++"
	)

	if hostOS == constant.OS_MACOS {
		cc = "clang"
		cxx = "clang++"
	}

	if doingCrossCompiling {
		switch hostOS {
		case constant.OS_DEBIAN,
			constant.OS_UBUNTU:
			var tripleName string
			switch mKernel {
			case constant.KERNEL_LINUX:
				tripleName, _ = constant.GetDebianTripleName(mArch, mKernel, targetLibc)
			case constant.KERNEL_DARWIN:
				// TODO: set darwin version
				tripleName, _ = constant.GetAppleTripleName(mArch, "")
			case constant.KERNEL_WINDOWS:
				tripleName, _ = constant.GetDebianTripleName(mArch, mKernel, targetLibc)
			default:
			}

			cc = tripleName + "-gcc"
			cxx = tripleName + "-g++"
		case constant.OS_ALPINE:
			tripleName, _ := constant.GetAlpineTripleName(mArch)
			cc = tripleName + "-gcc"
			cxx = tripleName + "-g++"
		case constant.OS_MACOS:
			cc = "clang"
			cxx = "clang++"
		}
	}

	// TODO: generate suitable flags
	appendListEnv("CGO_CPPFLAGS", c.CPPFlags, cppflags)
	appendListEnv("CGO_CFLAGS", c.CFlags, cflags)
	appendListEnv("CGO_CXXFLAGS", c.CXXFlags, cxxflags)
	appendListEnv("CGO_FFLAGS", c.FFlags, fflags)
	appendListEnv("CGO_LDFLAGS", c.LDFlags, ldflags)

	appendEnv("CC", c.CC, cc)
	appendEnv("CXX", c.CXX, cxx)

	return ret
}