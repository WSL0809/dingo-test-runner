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
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pingcap/errors"
	log "github.com/sirupsen/logrus"
)

// OpenDBWithRetry opens a database specified by its database driver name and a
// driver-specific data source name. And it will do some retries if the connection fails.
func OpenDBWithRetry(driverName, dataSourceName string, retryCount int) (mdb *sql.DB, err error) {
	startTime := time.Now()
	sleepTime := time.Millisecond * 500
	// Parse DSN to extract connection info for better error messages
	var host, port, dbname string
	if parts := strings.Split(dataSourceName, "@tcp("); len(parts) > 1 {
		if hostParts := strings.Split(parts[1], ")"); len(hostParts) > 0 {
			hostPort := strings.Split(hostParts[0], ":")
			if len(hostPort) >= 2 {
				host = hostPort[0]
				port = hostPort[1]
			}
		}
	}
	if parts := strings.Split(dataSourceName, ")/"); len(parts) > 1 {
		dbParts := strings.Split(parts[1], "?")
		if len(dbParts) > 0 {
			dbname = dbParts[0]
		}
	}

	// The max retry interval is 60 s.
	for i := 0; i < retryCount; i++ {
		mdb, err = sql.Open(driverName, dataSourceName)
		if err != nil {
			log.Warnf("第 %d 次尝试连接数据库失败，还剩 %d 次重试，错误：%v", i+1, retryCount-i-1, err)
			time.Sleep(sleepTime)
			continue
		}
		err = mdb.Ping()
		if err == nil {
			break
		}
		log.Warnf("第 %d 次尝试ping数据库失败，还剩 %d 次重试，错误：%v", i+1, retryCount-i-1, err)
		mdb.Close()
		time.Sleep(sleepTime)
	}
	if err != nil {
		// Provide user-friendly error message
		var friendlyMsg strings.Builder
		friendlyMsg.WriteString("数据库连接失败，请检查以下配置：\n")
		if host != "" && port != "" {
			friendlyMsg.WriteString(fmt.Sprintf("  - 主机地址: %s\n", host))
			friendlyMsg.WriteString(fmt.Sprintf("  - 端口: %s\n", port))
		}
		if dbname != "" {
			friendlyMsg.WriteString(fmt.Sprintf("  - 数据库名: %s\n", dbname))
		}
		friendlyMsg.WriteString("  - 请确认数据库服务是否启动\n")
		friendlyMsg.WriteString("  - 请确认网络连接是否正常\n")
		friendlyMsg.WriteString("  - 请确认用户名和密码是否正确\n")
		friendlyMsg.WriteString("  - 请确认防火墙设置是否阻止连接\n")
		friendlyMsg.WriteString(fmt.Sprintf("\n总耗时: %v，重试次数: %d\n", time.Since(startTime), retryCount))
		friendlyMsg.WriteString(fmt.Sprintf("底层错误: %v", err))

		log.Errorf(friendlyMsg.String())
		return nil, errors.Trace(errors.Errorf("数据库连接失败: %v", err))
	}

	return
}

func processEscapes(str string) string {
	escapeMap := map[string]string{
		`\n`: "\n",
		`\t`: "\t",
		`\r`: "\r",
		`\/`: "/",
		`\\`: "\\", // better be the last one
	}

	for escape, replacement := range escapeMap {
		str = strings.ReplaceAll(str, escape, replacement)
	}

	return str
}

func ParseReplaceRegex(originalString string) ([]*ReplaceRegex, error) {
	var begin, middle, end, cnt int
	ret := make([]*ReplaceRegex, 0)
	for i, c := range originalString {
		if c != '/' {
			continue
		}
		if i != 0 && originalString[i-1] == '\\' {
			continue
		}
		cnt++
		switch cnt % 3 {
		// The first '/'
		case 1:
			begin = i
		// The second '/'
		case 2:
			middle = i
		// The last '/', we could compile regex and process replace string
		case 0:
			end = i
			reg, err := regexp.Compile(originalString[begin+1 : middle])
			if err != nil {
				return nil, err
			}
			ret = append(ret, &ReplaceRegex{
				regex:   reg,
				replace: processEscapes(originalString[middle+1 : end]),
			})
		}
	}
	if cnt%3 != 0 {
		return nil, errors.Errorf("Could not parse regex in --replace_regex: sql:%v", originalString)
	}
	return ret, nil
}
