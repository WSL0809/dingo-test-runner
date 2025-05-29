// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// XUnitTestSuites is a set of mysqltest suite.
type XUnitTestSuites struct {
	XMLName xml.Name `xml:"testsuites"`
	Suites  []XUnitTestSuite
}

// XUnitTestSuite is a single mysqltest suite which may contain many
// testcases in a directory
type XUnitTestSuite struct {
	XMLName    xml.Name        `xml:"testsuite"`
	Tests      int             `xml:"tests,attr"`
	Failures   int             `xml:"failures,attr"`
	Name       string          `xml:"name,attr"`
	Time       string          `xml:"time,attr"`
	Timestamp  string          `xml:"timestamp,attr"`
	Hostname   string          `xml:"hostname,attr"`
	Properties []XUnitProperty `xml:"properties>property,omitempty"`
	TestCases  []XUnitTestCase
}

// XUnitTestCase is a single test case with its result.
type XUnitTestCase struct {
	XMLName     xml.Name          `xml:"testcase"`
	Classname   string            `xml:"classname,attr"`
	Name        string            `xml:"name,attr"`
	Time        string            `xml:"time,attr"`
	QueryCount  int               `xml:"query-count,attr"`
	Status      string            `xml:"status,attr,omitempty"`
	Failure     *XUnitFailure     `xml:"failure,omitempty"`
	Properties  []XUnitProperty   `xml:"properties>property,omitempty"`
	Attachments []XUnitAttachment `xml:"attachments>attachment,omitempty"`
}

// XUnitFailure represents a test failure
type XUnitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

// XUnitAttachment represents a file attachment for a test case
type XUnitAttachment struct {
	Name   string `xml:"name,attr"`
	Source string `xml:"source,attr"`
	Type   string `xml:"type,attr"`
}

// XUnitProperty represents a key/value pair used to define properties.
type XUnitProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

func Write(out io.Writer, testSuite XUnitTestSuite) error {
	// 添加Allure所需的时间戳和主机名
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}
	testSuite.Timestamp = time.Now().Format(time.RFC3339)
	testSuite.Hostname = hostname
	
	// 为测试用例添加状态、属性和处理失败信息
	now := time.Now().Format(time.RFC3339)
	for i := range testSuite.TestCases {
		// 将原来的字符串失败信息转换为结构化失败对象
		if testSuite.TestCases[i].Failure == nil {
			if testSuite.TestCases[i].Status == "" {
				testSuite.TestCases[i].Status = "passed"
			}
		} else {
			testSuite.TestCases[i].Status = "failed"
		}
		
		// 添加测试用例的属性信息
		testDuration := 0
		if testSuite.TestCases[i].Time != "" {
			dur, err := time.ParseDuration(strings.TrimSuffix(testSuite.TestCases[i].Time, "s") + "s")
			if err == nil {
				testDuration = int(dur.Milliseconds())
			}
		}
		
		testSuite.TestCases[i].Properties = []XUnitProperty{
			{Name: "test_framework", Value: "mysql-tester"},
			{Name: "test_started_at", Value: now},
			{Name: "test_ended_at", Value: now},
			{Name: "test_duration_ms", Value: fmt.Sprintf("%d", testDuration)},
		}
	}
	
	testSuites := XUnitTestSuites{
		Suites: make([]XUnitTestSuite, 0),
	}
	testSuites.Suites = append(testSuites.Suites, testSuite)
	_, err := out.Write([]byte(xml.Header))
	if err != nil {
		log.Error("write xunit file fail:", err)
		return err
	}
	doc, err := xml.MarshalIndent(testSuites, "", "\t")
	if err != nil {
		return err
	}
	_, err = out.Write(doc)
	return err
}

// goVersion returns the version as reported by the go binary in PATH. This
// version will not be the same as runtime.Version, which is always the version
// of go used to build the gotestsum binary.
//
// To skip the os/exec call set the GOVERSION environment variable to the
// desired value.
func goVersion() string {
	if version, ok := os.LookupEnv("GOVERSION"); ok {
		return version
	}
	cmd := exec.Command("go", "version")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), "go version ")
}
