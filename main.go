package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/PratikMahajan/OC-Cluster-Manager/config"
	"go.uber.org/zap"
)

type Flags struct {
	// Creating Cluster
	Create bool
	// Destroy Cluster
	Destroy bool
	// Dry run of Algo
	DryRun bool
	// Platform to create cluster on
	Platform string
}

type Cluster struct {
	Name     string
	Dir      string
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
	flag.BoolVar(&flags.Destroy, "destroy", false, "Destroy the cluster")
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

	currentDate := time.Now().UTC().Format("01-02")
	// configure the cluster
	cluster := Cluster{
		Name:     fmt.Sprintf("%s-%s", cfg.ClusterNamePrefix, currentDate),
		Dir:      cfg.OCStorePath,
		Platform: platform,
	}

	if flags.Create {

		logger.Info("execute OpenShift install create")

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

	}
	if flags.Destroy {
		logger.Info("execute OpenShift install destroy")

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
		err := runCmd(logger, cmd)
		if err != nil {
			logger.Fatal("failed to destroy cluster",
				zap.String(cluster.Name, err.Error()))
		}

		logger.Info("destroyed cluster", zap.String("Name", cluster.Name))
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
