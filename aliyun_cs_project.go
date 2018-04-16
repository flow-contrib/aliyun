package aliyun

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chr4/pwgen"
	"github.com/denverdino/aliyungo/cs"
	"github.com/gogap/logrus"
	"github.com/howeyc/gopass"
)

type EnvFunc func(code, key string, query url.Values) (ret string, err error)

var (
	EnvFuncs = map[string]EnvFunc{
		"pwgen":    envFuncPWGEN,
		"readline": envFuncReadline,
	}
)

type DockerProjectCreationArgs struct {
	*DockerProjectClient

	WaitProjects []string
	CreationArgs *cs.ProjectCreationArgs
}

func (p *DockerProjectCreationArgs) Create() error {
	return p.Client.CreateProject(p.CreationArgs)
}

func (p *DockerProjectCreationArgs) Wait() error {

	if len(p.WaitProjects) == 0 {
		return nil
	}

	wg := &sync.WaitGroup{}

	errChan := make(chan error, 1)

	for _, projName := range p.WaitProjects {

		if projName == p.CreationArgs.Name {
			continue
		}

		wg.Add(1)

		go func(name string) {

			defer wg.Done()

			for {
				proj, err := p.Client.GetProject(name)
				if err != nil {
					select {
					case errChan <- err:
					default:
					}
					return
				}
				if cs.ClusterState(proj.CurrentState) == cs.Running {
					return
				}

				time.Sleep(time.Second * 5)
			}

		}(projName)
	}

	wg.Wait()

	var err error

	select {
	case err = <-errChan:
	default:
	}

	return err

}

type DockerProject struct {
	Client  *DockerProjectClient
	Project *cs.Project
}

func (p *Aliyun) ListDockerProjects(clusterNames ...string) (dockerPojects map[string][]*DockerProject, err error) {

	clients, err := p.GetDockerClusterProjectClient(clusterNames...)
	if err != nil {
		return
	}

	if len(clients) == 0 {
		return
	}

	var clusterProjects = make(map[string][]*DockerProject)

	for clusterName, client := range clients {

		if _, exist := clusterProjects[clusterName]; !exist {
			clusterProjects[clusterName] = nil
		}

		var projects cs.GetProjectsResponse
		projects, err = client.Client.GetProjects("", true, true)
		if err != nil {
			return
		}

		for i := 0; i < len(projects); i++ {
			clusterProjects[clusterName] = append(
				clusterProjects[clusterName],
				&DockerProject{
					Client:  clients[clusterName],
					Project: &projects[i],
				},
			)
		}

	}

	dockerPojects = clusterProjects

	return
}

func (p *Aliyun) CreateDockerProjectArgs() (createArgs []*DockerProjectCreationArgs, err error) {

	csConfig := p.Config.GetConfig("aliyun.cs.swarm")

	if csConfig.IsEmpty() {
		return
	}

	clusterProjects, err := p.ListDockerProjects(csConfig.Keys()...)
	if err != nil {
		return
	}

	var args []*DockerProjectCreationArgs

	for _, clusterName := range csConfig.Keys() {
		clusterProjectConfigs := csConfig.GetConfig(clusterName + ".projects")

		projects, exist := clusterProjects[clusterName]
		if !exist {
			err = fmt.Errorf("cluster %s not exist", clusterName)
			return
		}

		var existProjects = make(map[string]bool)

		for _, proj := range projects {
			existProjects[proj.Project.Name] = true
		}

		for _, needCreateProjectName := range clusterProjectConfigs.Keys() {

			if existProjects[needCreateProjectName] {

				logrus.WithField("CODE", p.Code).
					WithField("CLUSTER-NAME", clusterName).
					WithField("PROJECT-NAME", needCreateProjectName).
					Warn("project already created")

				continue
			}

			projectConf := clusterProjectConfigs.GetConfig(needCreateProjectName)

			if projectConf.IsEmpty() {

				logrus.WithField("CODE", p.Code).
					WithField("DOCKER-CLUSTER-NAME", clusterName).
					WithField("DOCKER-PROJECT-NAME", needCreateProjectName).
					Warn("project's config is empty, ignore to create")

				continue
			}

			templateFile := projectConf.GetString("template")

			if len(templateFile) == 0 {
				err = fmt.Errorf("template file not set, cluster: %s, project: %s", clusterName, needCreateProjectName)
				return
			}

			var tmplData []byte
			tmplData, err = ioutil.ReadFile(templateFile)

			if err != nil {
				return
			}

			envsConf := projectConf.GetConfig("environment")

			mapENVs := map[string]string{}

			for _, envKey := range envsConf.Keys() {
				var v, fnName string
				name := fmt.Sprintf("cs.swarm.%s.%s.environment.%s", clusterName, needCreateProjectName, envKey)
				v, fnName, err = p.tryInvokeEnvFunc(name, envsConf.GetString(envKey))

				if err != nil {
					return
				}

				if len(fnName) > 0 {
					logrus.WithField("CODE", p.Code).
						WithField("FUNC", fnName).
						WithField("DOCKER-CLUSTER-NAME", clusterName).
						WithField("PROJECT", needCreateProjectName).
						WithField("NAME", name).
						WithField("VALUE", v).Debugln("Environment func invoked")
				}

				mapENVs[envKey] = v
			}

			var projsClient map[string]*DockerProjectClient
			projsClient, err = p.GetDockerClusterProjectClient(clusterName)
			if err != nil {
				return
			}

			cli, exist := projsClient[clusterName]
			if !exist {
				err = fmt.Errorf("get docker cluster %s project's of %s client failure", clusterName, needCreateProjectName)
				return
			}

			waitProjects := projectConf.GetStringList("wait.projects")

			arg := &DockerProjectCreationArgs{
				WaitProjects:        waitProjects,
				DockerProjectClient: cli,
				CreationArgs: &cs.ProjectCreationArgs{
					Name:        needCreateProjectName,
					Description: projectConf.GetString("description"),
					Template:    string(tmplData),
					Version:     projectConf.GetString("version", "1.0.0"),
					LatestImage: projectConf.GetBoolean("latest-image", true),
					Environment: mapENVs,
				},
			}

			args = append(args, arg)
		}
	}

	createArgs = args

	return
}

func (p *Aliyun) DeleteDockerProjectArgs() (dockerPojects map[string][]*DockerProject, err error) {
	csConfig := p.Config.GetConfig("aliyun.cs.swarm")

	if csConfig.IsEmpty() {
		return
	}

	clusterProjects, err := p.ListDockerProjects(csConfig.Keys()...)
	if err != nil {
		return
	}

	retDockerPojects := map[string][]*DockerProject{}

	for _, clusterName := range csConfig.Keys() {
		clusterProjectConfigs := csConfig.GetConfig(clusterName + ".projects")

		projects, exist := clusterProjects[clusterName]
		if !exist {
			continue
		}

		var existProjects = make(map[string]*DockerProject)

		for i, proj := range projects {
			existProjects[proj.Project.Name] = projects[i]
		}

		for _, needDeleteProjectName := range clusterProjectConfigs.Keys() {

			proj, exist := existProjects[needDeleteProjectName]
			if !exist {
				continue
			}

			retDockerPojects[clusterName] = append(retDockerPojects[clusterName], proj)
		}
	}

	dockerPojects = retDockerPojects

	return
}

func (p *Aliyun) tryInvokeEnvFunc(name, expr string) (string, string, error) {

	if len(expr) == 0 {
		return expr, "", nil
	}

	if !strings.HasPrefix(expr, "func://") {
		return expr, "", nil
	}

	funcUrl, err := url.Parse(expr)
	if err != nil {
		return expr, "", nil
	}

	if funcUrl.Scheme != "func" {
		return expr, "", nil
	}

	fnName := funcUrl.Hostname()

	fn, exist := EnvFuncs[fnName]

	if !exist {
		return expr, "", nil
	}

	ret, err := fn(p.Code, name, funcUrl.Query())

	return ret, fnName, err
}

func envFuncReadline(code, key string, query url.Values) (ret string, err error) {
	prompt := query.Get("prompt")
	typ := query.Get("type")
	strConfirm := query.Get("confirm")
	hashAlog := strings.ToLower(query.Get("hash"))
	env := query.Get("set_env")

	if len(prompt) == 0 {
		prompt = fmt.Sprintf("[KEY:%s]", key)
	} else {
		prompt = fmt.Sprintf("[KEY:%s] %s", key, prompt)
	}

	var input []byte

	confirm, _ := strconv.ParseBool(strConfirm)

	if typ == "password" || typ == "pwd" || typ == "pass" {
		input, err = gopass.GetPasswdPrompt(prompt+":", true, os.Stdin, os.Stdout)
		if err != nil {
			return
		}

		if confirm {
			var input2 []byte
			input2, err = gopass.GetPasswdPrompt(prompt+"[CONFIRM]:", true, os.Stdin, os.Stdout)
			if err != nil {
				return
			}

			if string(input) != string(input2) {
				err = fmt.Errorf("twice input did not same, key: %s", key)
				return
			}
		}
	} else {
		fmt.Printf("%s:", prompt)
		input = readLine(os.Stdin)

		if confirm {
			fmt.Printf("%s[CONFIRM]:", prompt)
			var input2 []byte
			input2 = readLine(os.Stdin)

			if string(input) != string(input2) {
				err = fmt.Errorf("twice input did not same")
				return
			}
		}
	}

	if hashAlog == "md5" {
		h := md5.Sum(input)
		input = []byte(fmt.Sprintf("%0x", h))
	} else if hashAlog == "sha256" {
		h := sha256.Sum256(input)
		input = []byte(fmt.Sprintf("%0x", h))
	}

	ret = string(input)

	if len(env) > 0 {
		os.Setenv(env, string(input))
	}

	return
}

func envFuncPWGEN(code, key string, query url.Values) (ret string, err error) {

	strLen := query.Get("len")
	env := query.Get("set_env")

	genLen, err := strconv.ParseInt(strLen, 10, 64)
	if err != nil {
		return
	}

	ret = pwgen.AlphaNum(int(genLen))

	if len(env) > 0 {
		os.Setenv(env, ret)
	}

	return
}

func readLine(reader io.Reader) []byte {
	buf := bufio.NewReader(reader)
	line, err := buf.ReadBytes('\n')

	for err == nil {
		line = bytes.TrimRight(line, "\n")
		if len(line) > 0 {
			if line[len(line)-1] == 13 { //'\r'
				line = bytes.TrimRight(line, "\r")
			}
			return line
		}
		line, err = buf.ReadBytes('\n')
	}

	if len(line) > 0 {
		return line
	}

	return nil
}
