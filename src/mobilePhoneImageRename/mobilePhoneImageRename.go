package mobilePhoneImageRename

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

type Command struct {
	fs *flag.FlagSet

	name string
}

func Cmd() *Command {
	gc := &Command{
		fs: flag.NewFlagSet("mobilePhoneImageRename", flag.ContinueOnError),
	}

	gc.fs.StringVar(&gc.name, "name", "World", "name of the person to be greeted")

	return gc
}

func (g *Command) Name() string {
	return g.fs.Name()
}

func (g *Command) Init(args []string) error {
	return g.fs.Parse(args)
}

func find(root, ext string) []string {
	var a []string
	filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ext {
			a = append(a, s)
		}
		return nil
	})
	return a
}

func (g *Command) Run() error {
	// fmt.Println("Hello", g.name, "!")

	// Current directory (where the executable is called from)
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(path)

	// Current executable (where the executable is located)
	path2, err2 := os.Executable()
	if err2 != nil {
		log.Println(err2)
	}
	exPath := filepath.Dir(path2)
	fmt.Println(exPath)

	// find all images
	for _, s := range find(path, ".jpg") {
		file_path := filepath.Dir(s)
		file_name_old := filepath.Base(s)

		// split file name
		reg1 := regexp.MustCompile(`^PXL_(?P<Y>\d{4})(?P<M>\d{2})(?P<D>\d{2})_(?P<h>\d{2})(?P<m>\d{2})(?P<s>\d{2})(?P<ms>\d{3})\.(?P<ending>jpg|png|mp4)$`)
		match := reg1.FindStringSubmatch(file_name_old)
		result := make(map[string]string)
		for i, name := range reg1.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}
		// fmt.Printf("by name: %s %s\n", result["M"], result["m"])

		//TODO ask for how much + on hour (summer time +2 vs winter time +1 or completely other time zones)
		//TODO use date time format and add the time bc maybe the day will change

		// increment hour
		i, err := strconv.Atoi(result["h"])
		if err != nil {
			panic(err)
		}
		i = i + 1

		// crete new file name with incremented hour
		file_name_new := fmt.Sprintf("%s-%s-%s_%02d-%s-%s-%s___GooglePixel6.%s", result["Y"], result["M"],
			result["D"], i, result["m"], result["s"], result["ms"], result["ending"])

		// Rename
		e := os.Rename(filepath.Join(file_path, file_name_old), filepath.Join(file_path, file_name_new))
		if e != nil {
			log.Fatal(e)
		}
	}

	return nil
}
