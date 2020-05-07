package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/PratikMahajan/OC-Cluster-Automator/config"
	"github.com/PratikMahajan/OC-Cluster-Automator/models"
	"go.uber.org/zap"
)

type Flags struct {
	// Creating Cluster
	Create bool
	// Provide the cluster to destroy
	Destroy string
	// Dry run of Algo
	DryRun bool
	// Platform to create cluster on
	Platform string
}

func main() {

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()

	// graceful exit
	_, cancelCtx := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	go func() {
		<-sigs
		cancelCtx()
		logger.Info("received interrupt signal, will exit gracefully at the " +
			"end of the control loop execution")
	}()

	// Load configuration for cluster
	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal("failed to load configuration: ", zap.String("err", err.Error()))
	}

	logger.Debug("loaded configuration:")

	// Command line arguments
	flags := Flags{}
	flag.BoolVar(&flags.Create, "create", false, "Create a new cluster")
	flag.StringVar(&flags.Destroy, "destroy", "nil", "Destroy the cluster")
	flag.BoolVar(&flags.DryRun, "dryrun", false, "Dry run the program")
	flag.StringVar(&flags.Platform, "platform", "", "Platform to create cluster (aws/azure)")
	flag.Parse()

	// Find auxiliary scripts
	cwd, err := os.Getwd()
	if err != nil {
		logger.Fatal("failed to get working directory: ", zap.String("err", err.Error()))
	}

	platform := ""
	switch flags.Platform {
	case "aws":
		platform = "aws"
	case "azure":
		platform = "azure"
	case "":
		logger.Fatal("Please enter a platform to create the cluster on")
	default:
		log.Fatalf("Platform '%s' is not supported. Try aws/azure", flags.Platform)
	}

	// run-openshift-install.sh script
	runOpenShiftInstallScript := filepath.Join(cwd,
		"scripts/run-openshift-install.sh")
	if _, err := os.Stat(runOpenShiftInstallScript); err != nil {
		logger.Fatal("failed to stat scripts/run-openshift-install.sh: ",
			zap.String("err", err.Error()))
	}

	storePath := cfg.OCStorePath + "/OCClusterAutomator"
	_, err = os.Stat(storePath)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(storePath, 0755)
		if errDir != nil {
			logger.Fatal("Unable to create directory", zap.String("directory", storePath), zap.Error(err))
		}

	}

	if flags.Create {
		logger.Info("execute OpenShift install create")

		suffix := randomString(5) //time.Now().UTC().Format("0102405")
		// configure the cluster
		cluster := models.Cluster{
			Name:     fmt.Sprintf("%s-%s-%s", cfg.ClusterNamePrefix, platform, suffix),
			Dir:      storePath,
			Platform: platform,
		}

		// Dry run
		if flags.DryRun {
			logger.Info("would exec ", zap.String("Script", runOpenShiftInstallScript),
				zap.String("-s", cluster.Dir), zap.String("-a create -n ", cluster.Name), zap.String("-p ", cluster.Platform))
			return
		}

		// Create cluster
		cmd := exec.Command(runOpenShiftInstallScript,
			"-s", cluster.Dir,
			"-a", "create",
			"-n", cluster.Name,
			"-p", cluster.Platform)
		err := runCmd(logger, cmd)
		if err != nil {
			logger.Fatal("failed to create cluster",
				zap.String(cluster.Name, err.Error()))
		}

		logger.Info("created cluster", zap.String("Name", cluster.Name))

		err = saveClusterInfo(cluster, storePath)
		if err != nil {
			logger.Fatal("Unable to create/append json file", zap.String("ClusterName", cluster.Name), zap.Error(err))
		}

	}
	if flags.Destroy != "nil" {
		logger.Info("execute OpenShift install destroy")
		clusterStore, err := getSavedClusterInfo(storePath)
		if err != nil {
			logger.Fatal("clusterStore file error", zap.Error(err))
		}
		var cluster models.Cluster
		found := false
		logger.Info("Printing Current clusters on Platform", zap.String("Platform", platform))
		for _, value := range clusterStore.Clusters[platform] {
			logger.Info("Cluster", zap.String("Name", value.Name))
			if value.Name == flags.Destroy {
				cluster = value
				found = true
			}
		}

		if !found {
			logger.Info("cluster not found", zap.String("name", flags.Destroy))
			logger.Info("trying to delete cluster in default directory", zap.String("name", flags.Destroy))
			cluster = models.Cluster{Name: flags.Destroy, Dir: storePath, Platform: platform}
		}

		// Dry run
		if flags.DryRun {
			logger.Info("would exec Script %s", zap.String("Script", runOpenShiftInstallScript),
				zap.String("-s", cluster.Dir), zap.String("-a delete -n ", cluster.Name), zap.String("-p ", cluster.Platform))
			return
		}

		// Create cluster
		cmd := exec.Command(runOpenShiftInstallScript,
			"-s", cluster.Dir,
			"-a", "delete",
			"-n", cluster.Name,
			"-p", cluster.Platform)
		err = runCmd(logger, cmd)
		if err != nil {
			logger.Fatal("failed to destroy cluster",
				zap.String(cluster.Name, err.Error()))
		}

		logger.Info("destroyed cluster", zap.String("Name", cluster.Name))
		err = removeClusterInfo(cluster, storePath)
		if err != nil {
			logger.Fatal("Unable to remove data from json file", zap.String("ClusterName", cluster.Name), zap.Error(err))
		}
	}

}

// runCmd runs a command as a subprocess, handles printing out stdout and stderr
func runCmd(stdoutLogger *zap.Logger, cmd *exec.Cmd) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %s", err.Error())
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			stdoutLogger.Info("Install Script", zap.String("StdOut", scanner.Text()))
		}
	}()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %s", err.Error())
	}

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			stdoutLogger.Info("Install Script", zap.String("StdErr", scanner.Text()))
		}
	}()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %s", err.Error())
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to wait for command to complete: %s", err.Error())
	}

	return nil
}

func saveClusterInfo(cluster models.Cluster, storePath string) error {
	jsonFile := storePath + "/clusterinfo.json"
	var js []byte
	_, err := os.Stat(jsonFile)
	data := models.ClusterStore{}
	if os.IsNotExist(err) {
		acc := make(map[string][]models.Cluster)
		acc[cluster.Platform] = append(acc[cluster.Platform], cluster)
		data = models.ClusterStore{Clusters: acc}
	} else {
		file, err := ioutil.ReadFile(jsonFile)
		if err != nil {
			return fmt.Errorf("unable to read file %s:%v", jsonFile, err)
		}
		err = json.Unmarshal(file, &data)
		if err != nil {
			return fmt.Errorf("unable to get cluster data from file %s:%v", jsonFile, err)
		}
		data.Clusters[cluster.Platform] = append(data.Clusters[cluster.Platform], cluster)
	}
	js, err = json.Marshal(data)
	if err != nil {
		return fmt.Errorf("unable to create json file for clusters : %v", err)
	}
	err = ioutil.WriteFile(jsonFile, js, 0644)
	if err != nil {
		return fmt.Errorf("error writing json file to system %v", err)
	}
	return nil
}

func removeClusterInfo(cluster models.Cluster, storePath string) error {
	jsonFile := storePath + "/clusterinfo.json"
	var js []byte
	_, err := os.Stat(jsonFile)
	data := models.ClusterStore{}
	if os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist %v", jsonFile, err)
	} else {
		file, err := ioutil.ReadFile(jsonFile)
		if err != nil {
			return fmt.Errorf("unable to read file %s:%v", jsonFile, err)
		}
		err = json.Unmarshal(file, &data)
		if err != nil {
			return fmt.Errorf("unable to get cluster data from file %s:%v", jsonFile, err)
		}
		clusterIndex := indexOf(cluster, data.Clusters[cluster.Platform])
		data.Clusters[cluster.Platform] = removeIndex(data.Clusters[cluster.Platform], clusterIndex)
	}
	js, err = json.Marshal(data)
	if err != nil {
		return fmt.Errorf("unable to create json file for clusters : %v", err)
	}
	err = ioutil.WriteFile(jsonFile, js, 0644)
	if err != nil {
		return fmt.Errorf("error writing json file to system %v", err)
	}
	return nil
}

func indexOf(element models.Cluster, data []models.Cluster) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1 //not found.
}

func removeIndex(s []models.Cluster, index int) []models.Cluster {
	return append(s[:index], s[index+1:]...)
}

func getSavedClusterInfo(storePath string) (*models.ClusterStore, error) {
	jsonFile := storePath + "/clusterinfo.json"
	_, err := os.Stat(jsonFile)
	data := models.ClusterStore{}
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s does not exist", jsonFile)
	}
	file, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %s:%v", jsonFile, err)
	}
	err = json.Unmarshal(file, &data)
	if err != nil {
		return nil, fmt.Errorf("unable to get cluster data from file %s:%v", jsonFile, err)
	}
	return &data, nil
}

//randomString generates random string of given length.
// ex: for n = 5 it generates 6p7l0
func randomString(n int) string {
	tm := fmt.Sprintf("%d", time.Now().UnixNano())
	var letter = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	var timeS = []rune(tm)
	b := make([]rune, n)
	for i := range b {
		if i%2 == 1 {
			b[i] = letter[rand.Intn(len(letter))]

		} else {
			b[i] = timeS[rand.Intn(len(timeS))]
		}
	}
	return string(b)
}
