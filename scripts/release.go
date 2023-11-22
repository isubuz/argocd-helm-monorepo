package main

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

var ENVS = []string{"dev", "staging", "qa"}

func fileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	return !os.IsNotExist(err)
}

type conf struct {
	Name      string  `yaml:"name"`
	Service   service `yaml:"service"`
	ReleaseId int     `yaml:"releaseId"`
}

type service struct {
	Port int `yaml:"port"`
}

func main() {
	releaseIdEnvVar := strings.TrimSpace(os.Getenv("RELEASE_ID"))
	if releaseIdEnvVar == "" {
		panic("RELEASE_ID env var must be set")
	}
	numAppsEnvVar := strings.TrimSpace(os.Getenv("NUM_APPS"))
	if numAppsEnvVar == "" {
		panic("NUM_APPS env var must be set")
	}
	numApps, err := strconv.Atoi(numAppsEnvVar)
	check(err)

	var syncEnvs []string
	releaseEnvsVar := strings.TrimSpace(os.Getenv("RELEASE_ENVS"))
	if len(releaseEnvsVar) == 0 {
		syncEnvs = []string{}
	} else {
		syncEnvs = strings.Split(releaseEnvsVar, ",")
	}

	currReleaseId, err := strconv.Atoi(releaseIdEnvVar)
	fmt.Printf("Found current release id (%d)\n", currReleaseId)
	check(err)

	nextReleaseId := currReleaseId + 1
	fmt.Printf("Generating release for id (%d)\n", nextReleaseId)

	for _, env := range ENVS {
		for i := 1; i <= numApps; i++ {
			envDir := fmt.Sprintf("../helm-values/%s", env)
			err = os.MkdirAll(envDir, os.ModePerm)
			check(err)

			app := fmt.Sprintf("gb-%d", i)
			valuesFilePath := fmt.Sprintf("%s/%s.yaml", envDir, app)
			if fileExists(valuesFilePath) {
				if len(syncEnvs) > 0 && !slices.Contains(syncEnvs, env) {
					continue
				}
				valuesFile, err := os.ReadFile(valuesFilePath)
				check(err)

				var c conf
				err = yaml.Unmarshal(valuesFile, &c)
				check(err)

				// set next release id
				c.ReleaseId = nextReleaseId

				// write file
				d, err := yaml.Marshal(&c)
				check(err)

				fmt.Printf("Writing next release id for app (%s) and env (%s)\n", app, env)
				err = os.WriteFile(valuesFilePath, d, 0644)
				check(err)

			} else {
				fmt.Printf("Generating new values file for app (%s) and env (%s)\n", app, env)
				c := conf{
					Name: app,
					Service: service{
						Port: 30000 + i,
					},
					ReleaseId: nextReleaseId,
				}
				d, err := yaml.Marshal(&c)
				check(err)
				err = os.WriteFile(valuesFilePath, d, 0644)
				check(err)
			}
		}

	}
}
