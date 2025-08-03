// simple_test.go
package main

/*
#cgo LDFLAGS: -lodbc
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sql.h>
#include <sqlext.h>

int test_connection_and_insert() {
    SQLHENV env;
    SQLHDBC dbc;
    SQLHSTMT stmt;
    SQLRETURN ret;
    
    printf("开始ODBC连接测试...\n");
    
    // 初始化ODBC
    SQLAllocHandle(SQL_HANDLE_ENV, SQL_NULL_HANDLE, &env);
    SQLSetEnvAttr(env, SQL_ATTR_ODBC_VERSION, (void*)SQL_OV_ODBC3, 0);
    SQLAllocHandle(SQL_HANDLE_DBC, env, &dbc);
    
    // 连接数据库 - 使用正确的连接字符串
    char connStr[] = "Driver={IBM INFORMIX ODBC DRIVER};Servername=informix;Database=testgo;UID=informix;PWD=in4mix";
    ret = SQLDriverConnect(dbc, NULL, (SQLCHAR*)connStr, SQL_NTS, NULL, 0, NULL, SQL_DRIVER_NOPROMPT);
    
    if (ret != SQL_SUCCESS && ret != SQL_SUCCESS_WITH_INFO) {
        printf("连接失败\n");
        return 0;
    }
    
    printf("连接成功！\n");
    
    SQLAllocHandle(SQL_HANDLE_STMT, dbc, &stmt);
    
    // 创建表
    char createSQL[] = "CREATE TABLE IF NOT EXISTS test_table (id SERIAL PRIMARY KEY, name VARCHAR(50), message VARCHAR(200))";
    SQLExecDirect(stmt, (SQLCHAR*)createSQL, SQL_NTS);
    
    // 插入数据
    char insertSQL[] = "INSERT INTO test_table (name, message) VALUES ('Go测试', 'Go+ODBC+Informix连接成功！')";
    ret = SQLExecDirect(stmt, (SQLCHAR*)insertSQL, SQL_NTS);
    
    if (ret == SQL_SUCCESS || ret == SQL_SUCCESS_WITH_INFO) {
        printf("数据插入成功！\n");
        
        // 查询验证
        char querySQL[] = "SELECT id, name, message FROM test_table ORDER BY id DESC LIMIT 1";
        ret = SQLExecDirect(stmt, (SQLCHAR*)querySQL, SQL_NTS);
        
        if (ret == SQL_SUCCESS || ret == SQL_SUCCESS_WITH_INFO) {
            ret = SQLFetch(stmt);
            if (ret == SQL_SUCCESS || ret == SQL_SUCCESS_WITH_INFO) {
                SQLINTEGER id;
                SQLCHAR name[100];
                SQLCHAR message[300];
                SQLLEN idLen, nameLen, messageLen;
                
                SQLGetData(stmt, 1, SQL_C_SLONG, &id, 0, &idLen);
                SQLGetData(stmt, 2, SQL_C_CHAR, name, sizeof(name), &nameLen);
                SQLGetData(stmt, 3, SQL_C_CHAR, message, sizeof(message), &messageLen);
                
                printf("插入的数据: ID=%d, 名称=%s, 消息=%s\n", id, name, message);
            }
        }
        
        SQLFreeHandle(SQL_HANDLE_STMT, stmt);
        SQLDisconnect(dbc);
        SQLFreeHandle(SQL_HANDLE_DBC, dbc);
        SQLFreeHandle(SQL_HANDLE_ENV, env);
        return 1;
    } else {
        printf("数据插入失败\n");
        SQLFreeHandle(SQL_HANDLE_STMT, stmt);
        SQLDisconnect(dbc);
        SQLFreeHandle(SQL_HANDLE_DBC, dbc);
        SQLFreeHandle(SQL_HANDLE_ENV, env);
        return 0;
    }
}
*/
import "C"
import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("🎯 Go+ODBC+Informix 测试")
	fmt.Println("======================")
	
	// 设置环境变量
	os.Setenv("INFORMIXDIR", "/opt/IBM/informix")
	os.Setenv("INFORMIXSERVER", "informix")
	os.Setenv("INFORMIXSQLHOSTS", "/opt/IBM/informix/etc/sqlhosts")
	os.Setenv("LD_LIBRARY_PATH", "/opt/IBM/informix/lib:/opt/IBM/informix/lib/esql:/opt/IBM/informix/lib/cli:"+os.Getenv("LD_LIBRARY_PATH"))
	os.Setenv("CLIENT_LOCALE", "en_US.utf8")
	os.Setenv("DB_LOCALE", "en_US.utf8")
	
	result := C.test_connection_and_insert()
	
	if result > 0 {
		fmt.Println("\n✅ 成功！Go+ODBC+Informix 连接和数据插入完全成功！")
	} else {
		fmt.Println("\n❌ 失败")
	}
}