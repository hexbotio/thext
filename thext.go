package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
)

func main() {

	fmt.Println("THEXT - The Hex Testing Tool")
	fmt.Print("Loading... ")

	// take in the test file as argument
	if len(os.Args) != 2 {
		log.Fatal("ERROR: must supply test file")
	}
	testFile := os.Args[1]
	if _, err := os.Stat(testFile); err != nil {
		if os.IsNotExist(err) {
			log.Fatal("ERROR: test file does not exist ", err)
		}
	}

	// load the test file into a struct
	file, err := ioutil.ReadFile(testFile)
	if err != nil {
		log.Fatal("ERROR: cannot read test file ", err)
	}
	var testConfig TestConfig
	err = json.Unmarshal(file, &testConfig)
	if err != nil {
		log.Fatal("ERROR: cannot read test file ", err)
	}

	// validate hex path
	if _, err := os.Stat(testConfig.HexPath); err != nil {
		if os.IsNotExist(err) {
			log.Fatal("ERROR: hex path does not exist ", err)
		}
	}

	// validate plugins dir
	if _, err := os.Stat(testConfig.PluginsDir); err != nil {
		if os.IsNotExist(err) {
			log.Fatal("ERROR: plugins dir does not exist ", testConfig.PluginsDir, " ", err)
		}
	}

	// tracking
	fmt.Print("Starting...\n\n")
	begin := time.Now().Unix()
	passes := 0
	failes := 0

	// loop over test file
	for _, test := range testConfig.Tests {

		// timer
		start := time.Now().Unix()

		// construct the command and env
		cmd := testConfig.HexPath + " -quiet"
		if test.RulePath != "" {
			cmd = cmd + " -rule-path \"" + test.RulePath + "\""
		}
		cmd = cmd + " -command \"" + test.Command + "\""
		c := exec.Command("/bin/sh", "-c", cmd)
		if len(test.Env) > 0 {
			for key, value := range test.Env {
				c.Env[len(c.Env)] = key + "=" + value
			}
		}

		// setup vars and buffers
		output := ""
		success := true
		var o bytes.Buffer
		var e bytes.Buffer

		// connect to stdout/err
		c.Stdout = &o
		c.Stderr = &e

		// run and collect results
		err := c.Run()
		output = o.String()
		if err != nil {
			output = output + "\n" + err.Error() + "\n" + e.String()
			success = false
		}

		// test output
		response := true
		if test.Response != "" {
			response = strings.Contains(output, test.Response)
		}

		// timer
		total := time.Now().Unix() - start

		// output results
		if success == test.Success && response {
			color.Set(color.FgGreen)
			fmt.Print("[PASS] ")
			color.Unset()
			fmt.Printf("(%d) %s %s\n", total, test.Command, test.RulePath)
			passes++
		} else {
			color.Set(color.FgRed)
			fmt.Print("[FAIL] ")
			color.Unset()
			fmt.Printf("(%d) %s %s\n", total, test.Command, test.RulePath)
			fmt.Printf("  CMD: %s\n", cmd)
			fmt.Printf("  OUT: %s\n", output)
			failes++
		}

	}

	totalTests := time.Now().Unix() - begin
	fmt.Print("\nCOMPLETED: ")
	color.Set(color.FgGreen)
	fmt.Printf("%d Passed ", passes)
	color.Set(color.FgRed)
	fmt.Printf("%d Failed ", failes)
	color.Unset()
	fmt.Printf("in %d sec\n", totalTests)

	if failes > 0 {
		os.Exit(1)
	}
	os.Exit(0)

}

type TestConfig struct {
	HexPath    string `json:"hex_path"`
	PluginsDir string `json:"plugins_dir"`
	Tests      []Test `json:"tests"`
}

type Test struct {
	Command  string            `json:"command"`
	RulePath string            `json:"rule_path"`
	Success  bool              `json:"success"`
	Response string            `json:"response"`
	Env      map[string]string `json:"env"`
}
