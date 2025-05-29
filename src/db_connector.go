// Copyright 2025

package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pingcap/errors"
	log "github.com/sirupsen/logrus"
)

// ConnectionManager 负责管理数据库连接池
type ConnectionManager struct {
	connections map[string]*Conn
	currentConn *Conn
	defaultPort      string
	defaultParams    string
	retryConnCount   int
	defaultTimeZone  string
	allowAllFiles    bool
}

// NewConnectionManager 创建一个新的连接管理器实例
func NewConnectionManager(defaultPort string, defaultParams string, retryConnCount int) *ConnectionManager {
	return &ConnectionManager{
		connections:     make(map[string]*Conn),
		defaultPort:     defaultPort,
		defaultParams:   defaultParams,
		retryConnCount:  retryConnCount,
		defaultTimeZone: "Asia/Shanghai",
		allowAllFiles:   true,
	}
}

// GetConnection 获取指定名称的连接，如果不存在则返回nil
func (cm *ConnectionManager) GetConnection(connName string) *Conn {
	conn, exists := cm.connections[connName]
	if !exists {
		return nil
	}
	return conn
}

// GetCurrentConnection 获取当前活动的连接
func (cm *ConnectionManager) GetCurrentConnection() *Conn {
	return cm.currentConn
}

// SetCurrentConnection 设置当前活动的连接
func (cm *ConnectionManager) SetCurrentConnection(conn *Conn) {
	cm.currentConn = conn
}

// AddConnection 添加一个新的数据库连接
func (cm *ConnectionManager) AddConnection(connName, hostName, userName, password, db string, expectedErrs []string) (*Conn, error) {
	var (
		mdb *sql.DB
		err error
	)

	
	if cm.currentConn != nil &&
		cm.currentConn.hostName == hostName &&
		cm.currentConn.userName == userName &&
		cm.currentConn.password == password &&
		expectedErrs == nil {
		
		mdb = cm.currentConn.mdb
	} else {
		
		dsn := cm.buildDSN(userName, password, hostName, db)
		
		
		retryCount := cm.retryConnCount
		if expectedErrs != nil {
			retryCount = 1
		}
		
		
		mdb, err = cm.openDBWithRetry("mysql", dsn, retryCount)
	}

	if err != nil {
		if expectedErrs == nil {
			log.Fatalf("Open db err %v", err)
		}
		return nil, err
	}

	
	conn, err := cm.initConn(mdb, userName, password, hostName, db)
	if err != nil {
		if expectedErrs == nil {
			log.Fatalf("Init conn err %v", err)
		}
		return nil, err
	}

	
	cm.connections[connName] = conn
	cm.currentConn = conn
	return conn, nil
}

// SwitchConnection 切换到指定名称的连接
func (cm *ConnectionManager) SwitchConnection(connName string) (*Conn, error) {
	conn, ok := cm.connections[connName]
	if !ok {
		return nil, fmt.Errorf("connection %s not found", connName)
	}
	cm.currentConn = conn
	return conn, nil
}

// DisconnectConnection 断开指定的连接
func (cm *ConnectionManager) DisconnectConnection(connName string) error {
	conn, ok := cm.connections[connName]
	if !ok {
		return fmt.Errorf("connection %s not found", connName)
	}

	
	if cm.currentConn == conn {
		cm.currentConn = nil
	}

	
	if conn.conn != nil {
		if err := conn.conn.Close(); err != nil {
			return err
		}
		conn.conn = nil
	}

	
	delete(cm.connections, connName)
	return nil
}

// CloseAllConnections 关闭所有连接
func (cm *ConnectionManager) CloseAllConnections() {
	for _, conn := range cm.connections {
		if conn.conn != nil {
			conn.conn.Close()
		}
	}
	cm.connections = make(map[string]*Conn)
	cm.currentConn = nil
}

// buildDSN 构建数据库连接字符串
func (cm *ConnectionManager) buildDSN(userName, password, hostName, db string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?time_zone=%%27%s%%27&allowAllFiles=%t%s",
		userName, password, hostName, cm.defaultPort, db, cm.defaultTimeZone, cm.allowAllFiles, cm.defaultParams)
}

// openDBWithRetry 打开数据库连接并在失败时进行重试
func (cm *ConnectionManager) openDBWithRetry(driverName, dataSourceName string, retryCount int) (mdb *sql.DB, err error) {
	startTime := time.Now()
	sleepTime := time.Millisecond * 500
	
	for i := 0; i < retryCount; i++ {
		mdb, err = sql.Open(driverName, dataSourceName)
		if err != nil {
			log.Warnf("open db failed, retry count %d (remain %d) err %v", i, retryCount-i, err)
			time.Sleep(sleepTime)
			continue
		}
		err = mdb.Ping()
		if err == nil {
			break
		}
		log.Warnf("ping db failed, retry count %d (remain %d) err %v", i, retryCount-i, err)
		mdb.Close()
		time.Sleep(sleepTime)
	}
	if err != nil {
		log.Errorf("open db failed %v, take time %v", err, time.Since(startTime))
		return nil, errors.Trace(err)
	}

	return
}

// initConn 初始化数据库连接
func (cm *ConnectionManager) initConn(mdb *sql.DB, userName, password, hostName, dbName string) (*Conn, error) {
	conn := &Conn{
		mdb:      mdb,
		hostName: hostName,
		userName: userName,
		password: password,
		db:       dbName,
	}

	
	sqlConn, err := mdb.Conn(context.Background())
	if err != nil {
		return nil, err
	}
	conn.conn = sqlConn

	return conn, nil
}
