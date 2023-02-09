package utils

import (
	"fmt"

	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter/circleci/config"
)

func ConvertEnvs(in map[string]*config.Xcode) map[string]string {
	env := make(map[string]string)
	for k, v := range in {
		if v.String != nil {
			env[k] = *v.String
		}
		if v.Double != nil {
			env[k] = fmt.Sprintf("%v", *v.Double)
		}
	}
	return env
}

func ReplaceSecret(s string, prefix string) string {
	secret := fmt.Sprintf("replace-%s", prefix)
	if s != "" {
		secret = s
	}
	return fmt.Sprintf("<+secrets.getValue(\"%s\")>", secret)
}

func ReplaceString(s string, prefix string) string {
	val := fmt.Sprintf("replace-%s", prefix)
	if s != "" {
		val = s
	}
	return val
}
