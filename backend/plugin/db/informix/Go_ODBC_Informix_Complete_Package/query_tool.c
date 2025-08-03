#include <stdio.h>
#include <sql.h>
#include <sqlext.h>

int main() {
    SQLHENV env;
    SQLHDBC dbc;
    SQLHSTMT stmt;
    SQLRETURN ret;
    
    printf("🔍 手动查询验证工具\n");
    printf("==================\n");
    
    // 初始化
    SQLAllocHandle(SQL_HANDLE_ENV, SQL_NULL_HANDLE, &env);
    SQLSetEnvAttr(env, SQL_ATTR_ODBC_VERSION, (void*)SQL_OV_ODBC3, 0);
    SQLAllocHandle(SQL_HANDLE_DBC, env, &dbc);
    
    // 连接数据库
    char connStr[] = "Driver={IBM INFORMIX ODBC DRIVER};Servername=informix;Database=testgo;UID=informix;PWD=in4mix";
    ret = SQLDriverConnect(dbc, NULL, (SQLCHAR*)connStr, SQL_NTS, NULL, 0, NULL, SQL_DRIVER_NOPROMPT);
    
    if (ret != SQL_SUCCESS && ret != SQL_SUCCESS_WITH_INFO) {
        printf("❌ 连接失败\n");
        return 1;
    }
    
    printf("✅ 连接成功\n");
    
    SQLAllocHandle(SQL_HANDLE_STMT, dbc, &stmt);
    
    // 查询数据
    char querySQL[] = "SELECT id, name, message FROM test_table ORDER BY id";
    ret = SQLExecDirect(stmt, (SQLCHAR*)querySQL, SQL_NTS);
    
    if (ret == SQL_SUCCESS || ret == SQL_SUCCESS_WITH_INFO) {
        printf("📋 查询结果:\n");
        
        int rowNum = 0;
        while (SQLFetch(stmt) == SQL_SUCCESS) {
            SQLINTEGER id;
            SQLCHAR name[100];
            SQLCHAR message[300];
            SQLLEN idLen, nameLen, messageLen;
            
            SQLGetData(stmt, 1, SQL_C_SLONG, &id, 0, &idLen);
            SQLGetData(stmt, 2, SQL_C_CHAR, name, sizeof(name), &nameLen);
            SQLGetData(stmt, 3, SQL_C_CHAR, message, sizeof(message), &messageLen);
            
            rowNum++;
            printf("%d. ID=%d, 名称=%s\n", rowNum, id, name);
            printf("   消息=%s\n", message);
        }
        
        printf("📊 总共 %d 条记录\n", rowNum);
    }
    
    // 清理
    SQLFreeHandle(SQL_HANDLE_STMT, stmt);
    SQLDisconnect(dbc);
    SQLFreeHandle(SQL_HANDLE_DBC, dbc);
    SQLFreeHandle(SQL_HANDLE_ENV, env);
    
    return 0;
}