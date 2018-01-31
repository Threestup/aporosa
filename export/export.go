package export

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/Threestup/aporosa/cmd"
)

const (
	modeCSV  = "CSV"
	modeJSON = "JSON"
)

var (
	mode = modeJSON
)

func Init(newMode string) error {
	if mode == "" {
		// default mode
		return nil
	}
	if mode != modeCSV && mode != modeJSON {
		return fmt.Errorf("unknow serialize mode: %v", mode)
	}
	mode = newMode
	return nil
}

func fileExists(s string) (bool, error) {
	_, err := os.Stat(s)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, fmt.Errorf("%s: no such file or directory", s)
	}
	return true, err
}

type jsonBaseFile struct {
	Data []map[string]string `json:"data"`
}

func loadExistingJSON(path string) (jsonBaseFile, error) {
	jbf := jsonBaseFile{
		Data: []map[string]string{},
	}

	if ok, _ := fileExists(path); ok {
		// load things from json
		f, err := os.Open(path)
		if err != nil {
			return jbf, fmt.Errorf("error opening %v: %v", path, err.Error())
		}

		jp := json.NewDecoder(f)
		if err = jp.Decode(&jbf); err != nil {
			return jbf, fmt.Errorf("error deserializing %v: %v", path, err.Error())
		}
	}

	return jbf, nil
}

func saveJSON(formName string, newValue map[string]string) error {
	// make json
	filePath := path.Join(cmd.OutDir, formName+".json")
	values, err := loadExistingJSON(filePath)
	if err != nil {
		return err
	}

	newValue["__ts"] = fmt.Sprintf("%v", time.Now().Unix())
	values.Data = append(values.Data, newValue)

	b, err := json.Marshal(values)
	if err != nil {
		return fmt.Errorf("unable to marshal contact infos: %v", err)
	}

	// writing file
	err = ioutil.WriteFile(filePath, b, 0644)
	if err != nil {
		return fmt.Errorf("unable to write file: %v", err)
	}
	return nil
}

func Save(formName string, values map[string]string) error {
	fmt.Printf("new contact infos: %v\n", values)

	if mode == modeJSON {
		return saveJSON(formName, values)
	} else if mode == modeCSV {

	}

	return nil
}
