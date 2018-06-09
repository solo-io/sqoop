package install

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"net/http"
	"io"
	"github.com/pkg/errors"
	"github.com/solo-io/gloo/pkg/bootstrap"
	"github.com/solo-io/glooctl/pkg/config"
)

const (
	successMessage = `QLoo installed successfully.
Please switch to directory '%s', and run "docker-compose up"
to start QLoo.

`
)

var installDockerCmd = &cobra.Command{
	Use:   "docker [folder]",
	Short: "install QLoo with Docker and file-based storage",
	Long: `
Installs gloo to run with Docker Compose in the given install folder.
If the folder doesn't exist glooctl will create it.

Once installed you can go to the install folder and run:
	docker-compose up 
to start gloo.

Glooctl will configure itself to use this instance of gloo.`,
	Args: cobra.ExactArgs(1),
	Run: func(c *cobra.Command, a []string) {
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Unable to get current directory", err)
			os.Exit(1)
		}
		installDir := filepath.Join(pwd, a[0])
		err = dockerInstall(installDir)
		if err != nil {
			fmt.Printf("Unable to install gloo to %s: %q\n", installDir, err)
			os.Exit(1)
		}
		fmt.Printf(successMessage, installDir)
	},
}
}

const (
	envoyYamlURL         = "https://raw.githubusercontent.com/solo-io/qloo/master/install/docker-compose/envoy-config.yaml"
	dockerComposeYamlURL = "https://raw.githubusercontent.com/solo-io/qloo/master/install/docker-compose/docker-compose.yaml"
)

func dockerInstall(folder string) error {
	err := createInstallFolder(folder)
	if err != nil {
		return err
	}
	err = download(dockerComposeYamlURL, filepath.Join(folder, "docker-compose.yaml"))
	if err != nil {
		return err
	}
	err = download(envoyYamlURL, filepath.Join(folder, "envoy-config.yaml"))
	if err != nil {
		return err
	}

	err = createStorageFolders(folder)
	if err != nil {
		return err
	}

	return updateQLooctlConfig(folder)
}

func createInstallFolder(folder string) error {
	stat, err := os.Stat(folder)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "unable to setup install directory")
		}
		err = os.MkdirAll(folder, 0755)
		if err != nil {
			return errors.Wrap(err, "unable to create directory")
		}
		return nil
	}

	if !stat.IsDir() {
		return errors.Errorf("%s already exists and isn't a directory", folder)
	}

	return nil
}

func createStorageFolders(folder string) error {
	// _gloo_config/*
	for _, f := range []string{
		"upstreams",
		"virtualservices",
		"roles",
		"secrets",
		"files",
		"schemas",
		"resolvermaps",
	} {
		err := os.MkdirAll(filepath.Join(folder, "_gloo_config", f), 0755)
		if err != nil {
			return errors.Wrap(err, "unable to create storage directory"+f)
		}
	}

	return nil
}

func updateQLooctlConfig(folder string) error {
	opts := &bootstrap.Options{}

	opts.ConfigStorageOptions.Type = "file"
	opts.FileStorageOptions.Type = "file"
	opts.SecretStorageOptions.Type = "file"

	opts.FileOptions.ConfigDir = filepath.Join(folder, "_gloo_config")
	opts.FileOptions.FilesDir = filepath.Join(folder, "_gloo_config/files")
	opts.FileOptions.SecretDir = filepath.Join(folder, "_gloo_config/secrets")

	err := config.SaveConfig(opts)
	if err != nil {
		return errors.Wrap(err, "unable to configure glooctl")
	}

	return nil
}

func download(src, dst string) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	resp, err := http.Get(src)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}
