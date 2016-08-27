package realize

import (
	"errors"
	"fmt"
	"gopkg.in/urfave/cli.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Config struct contains the general informations about a project
type Config struct {
	file     string
	Version  string `yaml:"version,omitempty"`
	Projects []Project
}

// New method puts the cli params in the struct
func New(params *cli.Context) *Config {
	return &Config{
		file:    AppFile,
		Version: AppVersion,
		Projects: []Project{
			{
				Name:   nameFlag(params),
				Path:   filepath.Clean(params.String("path")),
				Build:  params.Bool("build"),
				Bin:    boolFlag(params.Bool("no-bin")),
				Run:    boolFlag(params.Bool("no-run")),
				Fmt:    boolFlag(params.Bool("no-fmt")),
				Test:   params.Bool("test"),
				Params: argsParam(params),
				Watcher: Watcher{
					Paths:  watcherPaths,
					Ignore: watcherIgnores,
					Exts:   watcherExts,
				},
			},
		},
	}
}

// argsParam parse one by one the given argumentes
func argsParam(params *cli.Context) []string {
	argsN := params.NArg()
	if argsN > 0 {
		var args []string
		for i := 0; i <= argsN-1; i++ {
			args = append(args, params.Args().Get(i))
		}
		return args
	}
	return nil
}

// NameParam check the project name presence. If empty takes the working directory name
func nameFlag(params *cli.Context) string {
	var name string
	if params.String("name") == "" && params.String("path") == "" {
		return WorkingDir()
	} else if params.String("path") != "/" {
		name = filepath.Base(params.String("path"))
	} else {
		name = params.String("name")
	}
	return name
}

// BoolParam is used to check the presence of a bool flag
func boolFlag(b bool) bool {
	if b {
		return false
	}
	return true
}

// WorkingDir returns the last element of the working dir path
func WorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(Red(err))
	}
	return filepath.Base(dir)
}

// Duplicates check projects with same name or same combinations of main/path
func Duplicates(value Project, arr []Project) (Project, error) {
	for _, val := range arr {
		if value.Path == val.Path || value.Name == val.Name {
			return val, errors.New("There is a duplicate of '" + val.Name + "'. Check your config file!")
		}
	}
	return Project{}, nil
}

// Clean duplicate projects
func (h *Config) Clean() {
	arr := h.Projects
	for key, val := range arr {
		if _, err := Duplicates(val, arr[key+1:]); err != nil {
			h.Projects = append(arr[:key], arr[key+1:]...)
			break
		}
	}
}

// Read, Check and remove duplicates from the config file
func (h *Config) Read() error {
	_, err := os.Stat(h.file)
	if err == nil {
		file, err := ioutil.ReadFile(h.file)
		if err == nil {
			if len(h.Projects) > 0 {
				err = yaml.Unmarshal(file, h)
				if err == nil {
					h.Clean()
				}
				return err
			}
			return errors.New("There are no projects")
		}
		return h.Create()
	}
	return err
}

// Create and unmarshal yaml config file
func (h *Config) Create() error {
	y, err := yaml.Marshal(h)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(h.file, y, 0655)
}

// Add another project
func (h *Config) Add(params *cli.Context) error {
	err := h.Read()
	if err == nil {
		new := Project{
			Name:   nameFlag(params),
			Path:   filepath.Clean(params.String("path")),
			Build:  params.Bool("build"),
			Bin:    boolFlag(params.Bool("no-bin")),
			Run:    boolFlag(params.Bool("no-run")),
			Fmt:    boolFlag(params.Bool("no-fmt")),
			Test:   params.Bool("test"),
			Params: argsParam(params),
			Watcher: Watcher{
				Paths:  watcherPaths,
				Exts:   watcherExts,
				Ignore: watcherIgnores,
			},
		}
		if _, err := Duplicates(new, h.Projects); err != nil {
			return err
		}
		h.Projects = append(h.Projects, new)
		err = h.Create()
		if err == nil {
			fmt.Println(Green("Your project was successfully added"))
		}
		return err
	}
	err = h.Create()
	if err == nil {
		fmt.Println(Green("The config file was successfully created"))
	}
	return err
}

// Remove a project in list
func (h *Config) Remove(params *cli.Context) error {
	err := h.Read()
	if err == nil {
		for key, val := range h.Projects {
			if params.String("name") == val.Name {
				h.Projects = append(h.Projects[:key], h.Projects[key+1:]...)
				err = h.Create()
				if err == nil {
					fmt.Println(Green("Your project was successfully removed"))
				}
				return err
			}
		}
		return errors.New("No project found")
	}
	return err
}

// List of projects
func (h *Config) List() error {
	err := h.Read()
	if err == nil {
		for _, val := range h.Projects {
			fmt.Println(Blue("|"), Blue(strings.ToUpper(val.Name)))
			fmt.Println(MagentaS("|"), "\t", Yellow("Base Path"), ":", MagentaS(val.Path))
			fmt.Println(MagentaS("|"), "\t", Yellow("Run"), ":", MagentaS(val.Run))
			fmt.Println(MagentaS("|"), "\t", Yellow("Build"), ":", MagentaS(val.Build))
			fmt.Println(MagentaS("|"), "\t", Yellow("Install"), ":", MagentaS(val.Bin))
			fmt.Println(MagentaS("|"), "\t", Yellow("Fmt"), ":", MagentaS(val.Fmt))
			fmt.Println(MagentaS("|"), "\t", Yellow("Params"), ":", MagentaS(val.Params))
			fmt.Println(MagentaS("|"), "\t", Yellow("Watcher"), ":")
			fmt.Println(MagentaS("|"), "\t\t", Yellow("After"), ":", MagentaS(val.Watcher.After))
			fmt.Println(MagentaS("|"), "\t\t", Yellow("Before"), ":", MagentaS(val.Watcher.Before))
			fmt.Println(MagentaS("|"), "\t\t", Yellow("Extensions"), ":", MagentaS(val.Watcher.Exts))
			fmt.Println(MagentaS("|"), "\t\t", Yellow("Paths"), ":", MagentaS(val.Watcher.Paths))
			fmt.Println(MagentaS("|"), "\t\t", Yellow("Paths ignored"), ":", MagentaS(val.Watcher.Ignore))
			fmt.Println(MagentaS("|"), "\t\t", Yellow("Watch preview"), ":", MagentaS(val.Watcher.Preview))
		}
		return nil
	}
	return err
}
