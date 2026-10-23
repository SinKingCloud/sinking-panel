import {get, post} from "@/utils/request";
import {systemTask} from "./system";

/** 数据库实例列表 */
export async function getDatabaseList(params: API.RequestParams = {}) {
    return get("/database/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 数据库实例详情及连接表单 */
export async function getDatabaseInfo(params: API.RequestParams = {}) {
    return get("/database/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 添加数据库实例 */
export async function createDatabase(params: API.RequestParams = {}) {
    return post("/database/create", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改数据库实例 */
export async function updateDatabase(params: API.RequestParams = {}) {
    return post("/database/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 删除连接信息，不删除远端数据库 */
export async function deleteDatabase(params: API.RequestParams = {}) {
    return post("/database/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 使用待保存配置测试连接 */
export async function testDatabase(params: API.RequestParams = {}) {
    return post("/database/test", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取数据库类型对应的连接配置表单 */
export async function getDatabaseConfig(params: API.RequestParams = {}) {
    return post("/database/config", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 实时数据库列表 POST /database/mysql/databases */
export async function getMysqlDatabases(params: API.RequestParams = {}) {
    return post("/database/mysql/databases", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 真实数据库信息 POST /database/mysql/database_info */
export async function getMysqlDatabaseInfo(params: API.RequestParams = {}) {
    return post("/database/mysql/database_info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 可用字符集、排序规则及服务器默认值 */
export async function getMysqlCharsets(params: API.RequestParams = {}) {
    return post("/database/mysql/charsets", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 创建真实数据库 POST /database/mysql/create_database */
export async function createMysqlDatabase(params: API.RequestParams = {}) {
    return post("/database/mysql/create_database", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 删除真实数据库 POST /database/mysql/drop_database */
export async function dropMysqlDatabase(params: API.RequestParams = {}) {
    return post("/database/mysql/drop_database", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 修改数据库属性 POST /database/mysql/update_database */
export async function updateMysqlDatabase(params: API.RequestParams = {}) {
    return post("/database/mysql/update_database", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 版本及可用功能 POST /database/mysql/version */
export async function getMysqlVersion(params: API.RequestParams = {}) {
    return post("/database/mysql/version", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 创建数据表 POST /database/mysql/create_table */
export async function createMysqlTable(params: API.RequestParams = {}) {
    return post("/database/mysql/create_table", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 重命名数据表 POST /database/mysql/rename_table */
export async function renameMysqlTable(params: API.RequestParams = {}) {
    return post("/database/mysql/rename_table", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 添加字段 POST /database/mysql/add_column */
export async function addMysqlColumn(params: API.RequestParams = {}) {
    return post("/database/mysql/add_column", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 修改字段 POST /database/mysql/update_column */
export async function updateMysqlColumn(params: API.RequestParams = {}) {
    return post("/database/mysql/update_column", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 删除字段 POST /database/mysql/drop_column */
export async function dropMysqlColumn(params: API.RequestParams = {}) {
    return post("/database/mysql/drop_column", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 创建索引 POST /database/mysql/create_index */
export async function createMysqlIndex(params: API.RequestParams = {}) {
    return post("/database/mysql/create_index", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 修改索引备注 POST /database/mysql/update_index */
export async function updateMysqlIndex(params: API.RequestParams = {}) {
    return post("/database/mysql/update_index", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 删除索引 POST /database/mysql/drop_index */
export async function dropMysqlIndex(params: API.RequestParams = {}) {
    return post("/database/mysql/drop_index", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 查看建表语句 POST /database/mysql/ddl */
export async function getMysqlDDL(params: API.RequestParams = {}) {
    return post("/database/mysql/ddl", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 导入数据 POST /database/mysql/import */
export async function importMysql(params: API.RequestParams = {}) {
    return post("/database/mysql/import", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 导出数据到服务器 POST /database/mysql/export */
export async function exportMysql(params: API.RequestParams = {}) {
    return post("/database/mysql/export", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 数据库对象 POST /database/mysql/objects */
export async function getMysqlObjects(params: API.RequestParams = {}) {
    return post("/database/mysql/objects", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 查看对象定义 POST /database/mysql/object_definition */
export async function getMysqlObjectDefinition(params: API.RequestParams = {}) {
    return post("/database/mysql/object_definition", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 保存数据库对象 POST /database/mysql/save_object */
export async function saveMysqlObject(params: API.RequestParams = {}) {
    return post("/database/mysql/save_object", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 删除数据库对象 POST /database/mysql/drop_object */
export async function dropMysqlObject(params: API.RequestParams = {}) {
    return post("/database/mysql/drop_object", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 创建外键 POST /database/mysql/create_foreign_key */
export async function createMysqlForeignKey(params: API.RequestParams = {}) {
    return post("/database/mysql/create_foreign_key", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 修改外键 POST /database/mysql/update_foreign_key */
export async function updateMysqlForeignKey(params: API.RequestParams = {}) {
    return post("/database/mysql/update_foreign_key", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 删除外键 POST /database/mysql/drop_foreign_key */
export async function dropMysqlForeignKey(params: API.RequestParams = {}) {
    return post("/database/mysql/drop_foreign_key", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 数据表列表 POST /database/mysql/tables */
export async function getMysqlTables(params: API.RequestParams = {}) {
    return post("/database/mysql/tables", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 字段信息 POST /database/mysql/columns */
export async function getMysqlColumns(params: API.RequestParams = {}) {
    return post("/database/mysql/columns", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 索引信息 POST /database/mysql/indexes */
export async function getMysqlIndexes(params: API.RequestParams = {}) {
    return post("/database/mysql/indexes", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 分页数据 POST /database/mysql/rows */
export async function getMysqlRows(params: API.RequestParams = {}) {
    return post("/database/mysql/rows", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 执行语句 POST /database/mysql/execute */
export async function executeMysql(params: API.RequestParams = {}) {
    return post("/database/mysql/execute", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 添加数据 POST /database/mysql/insert */
export async function insertMysql(params: API.RequestParams = {}) {
    return post("/database/mysql/insert", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 修改数据 POST /database/mysql/update */
export async function updateMysql(params: API.RequestParams = {}) {
    return post("/database/mysql/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 删除数据 POST /database/mysql/delete */
export async function deleteMysql(params: API.RequestParams = {}) {
    return post("/database/mysql/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 账户权限 POST /database/mysql/users */
export async function manageMysqlUsers(params: API.RequestParams = {}) {
    return post("/database/mysql/users", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mysql 维护工具 POST /database/mysql/maintenance */
export async function maintainMysql(params: API.RequestParams = {}) {
    return post("/database/mysql/maintenance", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 实时数据库列表 POST /database/postgres/databases */
export async function getPostgresDatabases(params: API.RequestParams = {}) {
    return post("/database/postgres/databases", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 真实数据库信息 POST /database/postgres/database_info */
export async function getPostgresDatabaseInfo(params: API.RequestParams = {}) {
    return post("/database/postgres/database_info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 创建真实数据库 POST /database/postgres/create_database */
export async function createPostgresDatabase(params: API.RequestParams = {}) {
    return post("/database/postgres/create_database", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 删除真实数据库 POST /database/postgres/drop_database */
export async function dropPostgresDatabase(params: API.RequestParams = {}) {
    return post("/database/postgres/drop_database", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 重命名数据库 POST /database/postgres/rename_database */
export async function renamePostgresDatabase(params: API.RequestParams = {}) {
    return post("/database/postgres/rename_database", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 版本及可用功能 POST /database/postgres/version */
export async function getPostgresVersion(params: API.RequestParams = {}) {
    return post("/database/postgres/version", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 架构列表 POST /database/postgres/schemas */
export async function getPostgresSchemas(params: API.RequestParams = {}) {
    return post("/database/postgres/schemas", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 创建数据表 POST /database/postgres/create_table */
export async function createPostgresTable(params: API.RequestParams = {}) {
    return post("/database/postgres/create_table", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 删除数据表 POST /database/postgres/drop_table */
export async function dropPostgresTable(params: API.RequestParams = {}) {
    return post("/database/postgres/drop_table", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 重命名数据表 POST /database/postgres/rename_table */
export async function renamePostgresTable(params: API.RequestParams = {}) {
    return post("/database/postgres/rename_table", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 添加字段 POST /database/postgres/add_column */
export async function addPostgresColumn(params: API.RequestParams = {}) {
    return post("/database/postgres/add_column", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 修改字段 POST /database/postgres/update_column */
export async function updatePostgresColumn(params: API.RequestParams = {}) {
    return post("/database/postgres/update_column", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 删除字段 POST /database/postgres/drop_column */
export async function dropPostgresColumn(params: API.RequestParams = {}) {
    return post("/database/postgres/drop_column", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 创建索引 POST /database/postgres/create_index */
export async function createPostgresIndex(params: API.RequestParams = {}) {
    return post("/database/postgres/create_index", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 删除索引 POST /database/postgres/drop_index */
export async function dropPostgresIndex(params: API.RequestParams = {}) {
    return post("/database/postgres/drop_index", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 查看建表语句 POST /database/postgres/ddl */
export async function getPostgresDDL(params: API.RequestParams = {}) {
    return post("/database/postgres/ddl", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 创建视图 POST /database/postgres/create_view */
export async function createPostgresView(params: API.RequestParams = {}) {
    return post("/database/postgres/create_view", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 导入数据 POST /database/postgres/import */
export async function importPostgres(params: API.RequestParams = {}) {
    return post("/database/postgres/import", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 导出数据到服务器 POST /database/postgres/export */
export async function exportPostgres(params: API.RequestParams = {}) {
    return post("/database/postgres/export", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 数据库对象 POST /database/postgres/objects */
export async function getPostgresObjects(params: API.RequestParams = {}) {
    return post("/database/postgres/objects", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 查看对象定义 POST /database/postgres/object_definition */
export async function getPostgresObjectDefinition(params: API.RequestParams = {}) {
    return post("/database/postgres/object_definition", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 保存数据库对象 POST /database/postgres/save_object */
export async function savePostgresObject(params: API.RequestParams = {}) {
    return post("/database/postgres/save_object", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 删除数据库对象 POST /database/postgres/drop_object */
export async function dropPostgresObject(params: API.RequestParams = {}) {
    return post("/database/postgres/drop_object", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 创建外键 POST /database/postgres/create_foreign_key */
export async function createPostgresForeignKey(params: API.RequestParams = {}) {
    return post("/database/postgres/create_foreign_key", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 修改外键 POST /database/postgres/update_foreign_key */
export async function updatePostgresForeignKey(params: API.RequestParams = {}) {
    return post("/database/postgres/update_foreign_key", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 删除外键 POST /database/postgres/drop_foreign_key */
export async function dropPostgresForeignKey(params: API.RequestParams = {}) {
    return post("/database/postgres/drop_foreign_key", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 数据表列表 POST /database/postgres/tables */
export async function getPostgresTables(params: API.RequestParams = {}) {
    return post("/database/postgres/tables", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 字段信息 POST /database/postgres/columns */
export async function getPostgresColumns(params: API.RequestParams = {}) {
    return post("/database/postgres/columns", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 索引信息 POST /database/postgres/indexes */
export async function getPostgresIndexes(params: API.RequestParams = {}) {
    return post("/database/postgres/indexes", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 分页数据 POST /database/postgres/rows */
export async function getPostgresRows(params: API.RequestParams = {}) {
    return post("/database/postgres/rows", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 执行查询 POST /database/postgres/query */
export async function queryPostgres(params: API.RequestParams = {}) {
    return post("/database/postgres/query", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 执行语句 POST /database/postgres/execute */
export async function executePostgres(params: API.RequestParams = {}) {
    return post("/database/postgres/execute", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 添加数据 POST /database/postgres/insert */
export async function insertPostgres(params: API.RequestParams = {}) {
    return post("/database/postgres/insert", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 修改数据 POST /database/postgres/update */
export async function updatePostgres(params: API.RequestParams = {}) {
    return post("/database/postgres/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 删除数据 POST /database/postgres/delete */
export async function deletePostgres(params: API.RequestParams = {}) {
    return post("/database/postgres/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 账户权限 POST /database/postgres/users */
export async function managePostgresUsers(params: API.RequestParams = {}) {
    return post("/database/postgres/users", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Postgres 维护工具 POST /database/postgres/maintenance */
export async function maintainPostgres(params: API.RequestParams = {}) {
    return post("/database/postgres/maintenance", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 实时数据库列表 POST /database/sqlserver/databases */
export async function getSqlServerDatabases(params: API.RequestParams = {}) {
    return post("/database/sqlserver/databases", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 真实数据库信息 POST /database/sqlserver/database_info */
export async function getSqlServerDatabaseInfo(params: API.RequestParams = {}) {
    return post("/database/sqlserver/database_info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 创建真实数据库 POST /database/sqlserver/create_database */
export async function createSqlServerDatabase(params: API.RequestParams = {}) {
    return post("/database/sqlserver/create_database", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 删除真实数据库 POST /database/sqlserver/drop_database */
export async function dropSqlServerDatabase(params: API.RequestParams = {}) {
    return post("/database/sqlserver/drop_database", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 重命名数据库 POST /database/sqlserver/rename_database */
export async function renameSqlServerDatabase(params: API.RequestParams = {}) {
    return post("/database/sqlserver/rename_database", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 版本及可用功能 POST /database/sqlserver/version */
export async function getSqlServerVersion(params: API.RequestParams = {}) {
    return post("/database/sqlserver/version", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 架构列表 POST /database/sqlserver/schemas */
export async function getSqlServerSchemas(params: API.RequestParams = {}) {
    return post("/database/sqlserver/schemas", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 创建数据表 POST /database/sqlserver/create_table */
export async function createSqlServerTable(params: API.RequestParams = {}) {
    return post("/database/sqlserver/create_table", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 删除数据表 POST /database/sqlserver/drop_table */
export async function dropSqlServerTable(params: API.RequestParams = {}) {
    return post("/database/sqlserver/drop_table", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 重命名数据表 POST /database/sqlserver/rename_table */
export async function renameSqlServerTable(params: API.RequestParams = {}) {
    return post("/database/sqlserver/rename_table", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 添加字段 POST /database/sqlserver/add_column */
export async function addSqlServerColumn(params: API.RequestParams = {}) {
    return post("/database/sqlserver/add_column", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 修改字段 POST /database/sqlserver/update_column */
export async function updateSqlServerColumn(params: API.RequestParams = {}) {
    return post("/database/sqlserver/update_column", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 删除字段 POST /database/sqlserver/drop_column */
export async function dropSqlServerColumn(params: API.RequestParams = {}) {
    return post("/database/sqlserver/drop_column", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 创建索引 POST /database/sqlserver/create_index */
export async function createSqlServerIndex(params: API.RequestParams = {}) {
    return post("/database/sqlserver/create_index", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 删除索引 POST /database/sqlserver/drop_index */
export async function dropSqlServerIndex(params: API.RequestParams = {}) {
    return post("/database/sqlserver/drop_index", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 查看建表语句 POST /database/sqlserver/ddl */
export async function getSqlServerDDL(params: API.RequestParams = {}) {
    return post("/database/sqlserver/ddl", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 创建视图 POST /database/sqlserver/create_view */
export async function createSqlServerView(params: API.RequestParams = {}) {
    return post("/database/sqlserver/create_view", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 导入数据 POST /database/sqlserver/import */
export async function importSqlServer(params: API.RequestParams = {}) {
    return post("/database/sqlserver/import", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 导出数据到服务器 POST /database/sqlserver/export */
export async function exportSqlServer(params: API.RequestParams = {}) {
    return post("/database/sqlserver/export", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 数据库对象 POST /database/sqlserver/objects */
export async function getSqlServerObjects(params: API.RequestParams = {}) {
    return post("/database/sqlserver/objects", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 查看对象定义 POST /database/sqlserver/object_definition */
export async function getSqlServerObjectDefinition(params: API.RequestParams = {}) {
    return post("/database/sqlserver/object_definition", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 保存数据库对象 POST /database/sqlserver/save_object */
export async function saveSqlServerObject(params: API.RequestParams = {}) {
    return post("/database/sqlserver/save_object", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 删除数据库对象 POST /database/sqlserver/drop_object */
export async function dropSqlServerObject(params: API.RequestParams = {}) {
    return post("/database/sqlserver/drop_object", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 创建外键 POST /database/sqlserver/create_foreign_key */
export async function createSqlServerForeignKey(params: API.RequestParams = {}) {
    return post("/database/sqlserver/create_foreign_key", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 修改外键 POST /database/sqlserver/update_foreign_key */
export async function updateSqlServerForeignKey(params: API.RequestParams = {}) {
    return post("/database/sqlserver/update_foreign_key", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 删除外键 POST /database/sqlserver/drop_foreign_key */
export async function dropSqlServerForeignKey(params: API.RequestParams = {}) {
    return post("/database/sqlserver/drop_foreign_key", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 数据表列表 POST /database/sqlserver/tables */
export async function getSqlServerTables(params: API.RequestParams = {}) {
    return post("/database/sqlserver/tables", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 字段信息 POST /database/sqlserver/columns */
export async function getSqlServerColumns(params: API.RequestParams = {}) {
    return post("/database/sqlserver/columns", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 索引信息 POST /database/sqlserver/indexes */
export async function getSqlServerIndexes(params: API.RequestParams = {}) {
    return post("/database/sqlserver/indexes", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 分页数据 POST /database/sqlserver/rows */
export async function getSqlServerRows(params: API.RequestParams = {}) {
    return post("/database/sqlserver/rows", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 执行查询 POST /database/sqlserver/query */
export async function querySqlServer(params: API.RequestParams = {}) {
    return post("/database/sqlserver/query", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 执行语句 POST /database/sqlserver/execute */
export async function executeSqlServer(params: API.RequestParams = {}) {
    return post("/database/sqlserver/execute", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 添加数据 POST /database/sqlserver/insert */
export async function insertSqlServer(params: API.RequestParams = {}) {
    return post("/database/sqlserver/insert", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 修改数据 POST /database/sqlserver/update */
export async function updateSqlServer(params: API.RequestParams = {}) {
    return post("/database/sqlserver/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 删除数据 POST /database/sqlserver/delete */
export async function deleteSqlServer(params: API.RequestParams = {}) {
    return post("/database/sqlserver/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 账户权限 POST /database/sqlserver/users */
export async function manageSqlServerUsers(params: API.RequestParams = {}) {
    return post("/database/sqlserver/users", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** SqlServer 维护工具 POST /database/sqlserver/maintenance */
export async function maintainSqlServer(params: API.RequestParams = {}) {
    return post("/database/sqlserver/maintenance", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 实时数据库列表 POST /database/sqlite/databases */
export async function getSqliteDatabases(params: API.RequestParams = {}) {
    return post("/database/sqlite/databases", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 真实数据库信息 POST /database/sqlite/database_info */
export async function getSqliteDatabaseInfo(params: API.RequestParams = {}) {
    return post("/database/sqlite/database_info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 版本及可用功能 POST /database/sqlite/version */
export async function getSqliteVersion(params: API.RequestParams = {}) {
    return post("/database/sqlite/version", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 架构列表 POST /database/sqlite/schemas */
export async function getSqliteSchemas(params: API.RequestParams = {}) {
    return post("/database/sqlite/schemas", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 创建数据表 POST /database/sqlite/create_table */
export async function createSqliteTable(params: API.RequestParams = {}) {
    return post("/database/sqlite/create_table", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 删除数据表 POST /database/sqlite/drop_table */
export async function dropSqliteTable(params: API.RequestParams = {}) {
    return post("/database/sqlite/drop_table", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 重命名数据表 POST /database/sqlite/rename_table */
export async function renameSqliteTable(params: API.RequestParams = {}) {
    return post("/database/sqlite/rename_table", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 添加字段 POST /database/sqlite/add_column */
export async function addSqliteColumn(params: API.RequestParams = {}) {
    return post("/database/sqlite/add_column", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 修改字段 POST /database/sqlite/update_column */
export async function updateSqliteColumn(params: API.RequestParams = {}) {
    return post("/database/sqlite/update_column", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 删除字段 POST /database/sqlite/drop_column */
export async function dropSqliteColumn(params: API.RequestParams = {}) {
    return post("/database/sqlite/drop_column", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 创建索引 POST /database/sqlite/create_index */
export async function createSqliteIndex(params: API.RequestParams = {}) {
    return post("/database/sqlite/create_index", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 删除索引 POST /database/sqlite/drop_index */
export async function dropSqliteIndex(params: API.RequestParams = {}) {
    return post("/database/sqlite/drop_index", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 查看建表语句 POST /database/sqlite/ddl */
export async function getSqliteDDL(params: API.RequestParams = {}) {
    return post("/database/sqlite/ddl", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 创建视图 POST /database/sqlite/create_view */
export async function createSqliteView(params: API.RequestParams = {}) {
    return post("/database/sqlite/create_view", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 导入数据 POST /database/sqlite/import */
export async function importSqlite(params: API.RequestParams = {}) {
    return post("/database/sqlite/import", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 导出数据到服务器 POST /database/sqlite/export */
export async function exportSqlite(params: API.RequestParams = {}) {
    return post("/database/sqlite/export", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 数据库对象 POST /database/sqlite/objects */
export async function getSqliteObjects(params: API.RequestParams = {}) {
    return post("/database/sqlite/objects", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 查看对象定义 POST /database/sqlite/object_definition */
export async function getSqliteObjectDefinition(params: API.RequestParams = {}) {
    return post("/database/sqlite/object_definition", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 保存数据库对象 POST /database/sqlite/save_object */
export async function saveSqliteObject(params: API.RequestParams = {}) {
    return post("/database/sqlite/save_object", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 删除数据库对象 POST /database/sqlite/drop_object */
export async function dropSqliteObject(params: API.RequestParams = {}) {
    return post("/database/sqlite/drop_object", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 创建外键 POST /database/sqlite/create_foreign_key */
export async function createSqliteForeignKey(params: API.RequestParams = {}) {
    return post("/database/sqlite/create_foreign_key", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 修改外键 POST /database/sqlite/update_foreign_key */
export async function updateSqliteForeignKey(params: API.RequestParams = {}) {
    return post("/database/sqlite/update_foreign_key", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 删除外键 POST /database/sqlite/drop_foreign_key */
export async function dropSqliteForeignKey(params: API.RequestParams = {}) {
    return post("/database/sqlite/drop_foreign_key", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 数据表列表 POST /database/sqlite/tables */
export async function getSqliteTables(params: API.RequestParams = {}) {
    return post("/database/sqlite/tables", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 字段信息 POST /database/sqlite/columns */
export async function getSqliteColumns(params: API.RequestParams = {}) {
    return post("/database/sqlite/columns", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 索引信息 POST /database/sqlite/indexes */
export async function getSqliteIndexes(params: API.RequestParams = {}) {
    return post("/database/sqlite/indexes", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 分页数据 POST /database/sqlite/rows */
export async function getSqliteRows(params: API.RequestParams = {}) {
    return post("/database/sqlite/rows", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 执行查询 POST /database/sqlite/query */
export async function querySqlite(params: API.RequestParams = {}) {
    return post("/database/sqlite/query", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 执行语句 POST /database/sqlite/execute */
export async function executeSqlite(params: API.RequestParams = {}) {
    return post("/database/sqlite/execute", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 添加数据 POST /database/sqlite/insert */
export async function insertSqlite(params: API.RequestParams = {}) {
    return post("/database/sqlite/insert", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 修改数据 POST /database/sqlite/update */
export async function updateSqlite(params: API.RequestParams = {}) {
    return post("/database/sqlite/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 删除数据 POST /database/sqlite/delete */
export async function deleteSqlite(params: API.RequestParams = {}) {
    return post("/database/sqlite/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Sqlite 维护工具 POST /database/sqlite/maintenance */
export async function maintainSqlite(params: API.RequestParams = {}) {
    return post("/database/sqlite/maintenance", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Redis 实时数据库列表 POST /database/redis/databases */
export async function getRedisDatabases(params: API.RequestParams = {}) {
    return post("/database/redis/databases", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Redis 真实数据库信息 POST /database/redis/database_info */
export async function getRedisDatabaseInfo(params: API.RequestParams = {}) {
    return post("/database/redis/database_info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Redis 版本及可用功能 POST /database/redis/version */
export async function getRedisVersion(params: API.RequestParams = {}) {
    return post("/database/redis/version", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Redis 键列表 POST /database/redis/keys */
export async function getRedisKeys(params: API.RequestParams = {}) {
    return post("/database/redis/keys", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Redis 键值详情 POST /database/redis/value */
export async function getRedisValue(params: API.RequestParams = {}) {
    return post("/database/redis/value", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Redis 保存键值 POST /database/redis/save */
export async function saveRedis(params: API.RequestParams = {}) {
    return post("/database/redis/save", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Redis 删除键值 POST /database/redis/delete */
export async function deleteRedis(params: API.RequestParams = {}) {
    return post("/database/redis/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Redis 重命名键 POST /database/redis/rename */
export async function renameRedis(params: API.RequestParams = {}) {
    return post("/database/redis/rename", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Redis 设置有效期 POST /database/redis/expire */
export async function expireRedis(params: API.RequestParams = {}) {
    return post("/database/redis/expire", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Redis 执行命令 POST /database/redis/command */
export async function executeRedisCommand(params: API.RequestParams = {}) {
    return post("/database/redis/command", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Redis 运行信息 POST /database/redis/info */
export async function getRedisInfo(params: API.RequestParams = {}) {
    return post("/database/redis/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Redis 账户权限 POST /database/redis/users */
export async function manageRedisUsers(params: API.RequestParams = {}) {
    return post("/database/redis/users", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Redis 维护工具 POST /database/redis/maintenance */
export async function maintainRedis(params: API.RequestParams = {}) {
    return post("/database/redis/maintenance", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Redis 导入数据 POST /database/redis/import */
export async function importRedis(params: API.RequestParams = {}) {
    return post("/database/redis/import", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Redis 导出数据 POST /database/redis/export */
export async function exportRedis(params: API.RequestParams = {}) {
    return post("/database/redis/export", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 实时数据库列表 POST /database/mongo/databases */
export async function getMongoDatabases(params: API.RequestParams = {}) {
    return post("/database/mongo/databases", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 真实数据库信息 POST /database/mongo/database_info */
export async function getMongoDatabaseInfo(params: API.RequestParams = {}) {
    return post("/database/mongo/database_info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 创建真实数据库 POST /database/mongo/create_database */
export async function createMongoDatabase(params: API.RequestParams = {}) {
    return post("/database/mongo/create_database", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 删除真实数据库 POST /database/mongo/drop_database */
export async function dropMongoDatabase(params: API.RequestParams = {}) {
    return post("/database/mongo/drop_database", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 版本及可用功能 POST /database/mongo/version */
export async function getMongoVersion(params: API.RequestParams = {}) {
    return post("/database/mongo/version", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 集合列表 POST /database/mongo/collections */
export async function getMongoCollections(params: API.RequestParams = {}) {
    return post("/database/mongo/collections", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 索引列表 POST /database/mongo/indexes */
export async function getMongoIndexes(params: API.RequestParams = {}) {
    return post("/database/mongo/indexes", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 文档列表 POST /database/mongo/documents */
export async function getMongoDocuments(params: API.RequestParams = {}) {
    return post("/database/mongo/documents", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 添加文档 POST /database/mongo/insert */
export async function insertMongo(params: API.RequestParams = {}) {
    return post("/database/mongo/insert", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 修改文档 POST /database/mongo/update */
export async function updateMongo(params: API.RequestParams = {}) {
    return post("/database/mongo/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 删除文档 POST /database/mongo/delete */
export async function deleteMongo(params: API.RequestParams = {}) {
    return post("/database/mongo/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 聚合查询 POST /database/mongo/aggregate */
export async function aggregateMongo(params: API.RequestParams = {}) {
    return post("/database/mongo/aggregate", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 执行命令 POST /database/mongo/command */
export async function executeMongoCommand(params: API.RequestParams = {}) {
    return post("/database/mongo/command", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 创建集合 POST /database/mongo/create_collection */
export async function createMongoCollection(params: API.RequestParams = {}) {
    return post("/database/mongo/create_collection", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 删除集合 POST /database/mongo/drop_collection */
export async function dropMongoCollection(params: API.RequestParams = {}) {
    return post("/database/mongo/drop_collection", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 创建索引 POST /database/mongo/create_index */
export async function createMongoIndex(params: API.RequestParams = {}) {
    return post("/database/mongo/create_index", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 删除索引 POST /database/mongo/drop_index */
export async function dropMongoIndex(params: API.RequestParams = {}) {
    return post("/database/mongo/drop_index", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 账户权限 POST /database/mongo/users */
export async function manageMongoUsers(params: API.RequestParams = {}) {
    return post("/database/mongo/users", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 维护工具 POST /database/mongo/maintenance */
export async function maintainMongo(params: API.RequestParams = {}) {
    return post("/database/mongo/maintenance", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 导入数据 POST /database/mongo/import */
export async function importMongo(params: API.RequestParams = {}) {
    return post("/database/mongo/import", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 导出数据 POST /database/mongo/export */
export async function exportMongo(params: API.RequestParams = {}) {
    return post("/database/mongo/export", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 创建视图 POST /database/mongo/create_view */
export async function createMongoView(params: API.RequestParams = {}) {
    return post("/database/mongo/create_view", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 修改视图 POST /database/mongo/update_view */
export async function updateMongoView(params: API.RequestParams = {}) {
    return post("/database/mongo/update_view", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Mongo 视图定义 POST /database/mongo/view_definition */
export async function getMongoViewDefinition(params: API.RequestParams = {}) {
    return post("/database/mongo/view_definition", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Elasticsearch 版本及可用功能 POST /database/elasticsearch/version */
export async function getElasticsearchVersion(params: API.RequestParams = {}) {
    return post("/database/elasticsearch/version", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Elasticsearch 索引列表 POST /database/elasticsearch/indices */
export async function getElasticsearchIndices(params: API.RequestParams = {}) {
    return post("/database/elasticsearch/indices", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Elasticsearch 索引详情 POST /database/elasticsearch/info */
export async function getElasticsearchInfo(params: API.RequestParams = {}) {
    return post("/database/elasticsearch/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Elasticsearch 创建索引 POST /database/elasticsearch/create_index */
export async function createElasticsearchIndex(params: API.RequestParams = {}) {
    return post("/database/elasticsearch/create_index", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Elasticsearch 删除索引 POST /database/elasticsearch/drop_index */
export async function dropElasticsearchIndex(params: API.RequestParams = {}) {
    return post("/database/elasticsearch/drop_index", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Elasticsearch 分页文档 POST /database/elasticsearch/documents */
export async function getElasticsearchDocuments(params: API.RequestParams = {}) {
    return post("/database/elasticsearch/documents", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Elasticsearch 释放文档分页会话 POST /database/elasticsearch/close_search */
export async function closeElasticsearchSearch(params: API.RequestParams = {}) {
    return post("/database/elasticsearch/close_search", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Elasticsearch 添加文档 POST /database/elasticsearch/insert */
export async function insertElasticsearch(params: API.RequestParams = {}) {
    return post("/database/elasticsearch/insert", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Elasticsearch 修改文档 POST /database/elasticsearch/update */
export async function updateElasticsearch(params: API.RequestParams = {}) {
    return post("/database/elasticsearch/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Elasticsearch 删除文档 POST /database/elasticsearch/delete */
export async function deleteElasticsearch(params: API.RequestParams = {}) {
    return post("/database/elasticsearch/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Elasticsearch 字段映射 POST /database/elasticsearch/mapping */
export async function getElasticsearchMapping(params: API.RequestParams = {}) {
    return post("/database/elasticsearch/mapping", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Elasticsearch 分析器选项 POST /database/elasticsearch/analyzers */
export async function getElasticsearchAnalyzers(params: API.RequestParams = {}) {
    return post("/database/elasticsearch/analyzers", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Elasticsearch 修改字段映射 POST /database/elasticsearch/update_mapping */
export async function updateElasticsearchMapping(params: API.RequestParams = {}) {
    return post("/database/elasticsearch/update_mapping", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Elasticsearch 按索引结构更新已有文档 POST /database/elasticsearch/refresh_fields */
export async function refreshElasticsearchFields(params: API.RequestParams = {}) {
    return post("/database/elasticsearch/refresh_fields", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Elasticsearch 索引设置 POST /database/elasticsearch/settings */
export async function getElasticsearchSettings(params: API.RequestParams = {}) {
    return post("/database/elasticsearch/settings", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Elasticsearch 修改索引设置 POST /database/elasticsearch/update_settings */
export async function updateElasticsearchSettings(params: API.RequestParams = {}) {
    return post("/database/elasticsearch/update_settings", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Elasticsearch 集群状态 POST /database/elasticsearch/status */
export async function getElasticsearchStatus(params: API.RequestParams = {}) {
    return post("/database/elasticsearch/status", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** Elasticsearch 执行 REST 请求 POST /database/elasticsearch/request */
export async function executeElasticsearchRequest(params: API.RequestParams = {}) {
    return post("/database/elasticsearch/request", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 查询数据库导入、导出后台任务。 */
export async function getDatabaseTask(params: API.RequestParams = {}) {
    return systemTask(params);
}

// 仅汇总明确的接口函数，组件可按需替换接口实现，不拼接请求地址。
export type DatabaseRequests = Record<string, (params?: API.RequestParams) => ReturnType<typeof post>>;

export const mysqlRequests = {
    databases: getMysqlDatabases,
    database_info: getMysqlDatabaseInfo,
    charsets: getMysqlCharsets,
    create_database: createMysqlDatabase,
    drop_database: dropMysqlDatabase,
    update_database: updateMysqlDatabase,
    version: getMysqlVersion,
    create_table: createMysqlTable,
    rename_table: renameMysqlTable,
    add_column: addMysqlColumn,
    update_column: updateMysqlColumn,
    drop_column: dropMysqlColumn,
    create_index: createMysqlIndex,
    update_index: updateMysqlIndex,
    drop_index: dropMysqlIndex,
    ddl: getMysqlDDL,
    import: importMysql,
    export: exportMysql,
    objects: getMysqlObjects,
    object_definition: getMysqlObjectDefinition,
    save_object: saveMysqlObject,
    drop_object: dropMysqlObject,
    create_foreign_key: createMysqlForeignKey,
    update_foreign_key: updateMysqlForeignKey,
    drop_foreign_key: dropMysqlForeignKey,
    tables: getMysqlTables,
    columns: getMysqlColumns,
    indexes: getMysqlIndexes,
    rows: getMysqlRows,
    execute: executeMysql,
    insert: insertMysql,
    update: updateMysql,
    delete: deleteMysql,
    users: manageMysqlUsers,
    maintenance: maintainMysql,
    "system/task": getDatabaseTask,
} satisfies DatabaseRequests;

export const postgresRequests = {
    databases: getPostgresDatabases,
    database_info: getPostgresDatabaseInfo,
    create_database: createPostgresDatabase,
    drop_database: dropPostgresDatabase,
    rename_database: renamePostgresDatabase,
    version: getPostgresVersion,
    schemas: getPostgresSchemas,
    create_table: createPostgresTable,
    drop_table: dropPostgresTable,
    rename_table: renamePostgresTable,
    add_column: addPostgresColumn,
    update_column: updatePostgresColumn,
    drop_column: dropPostgresColumn,
    create_index: createPostgresIndex,
    drop_index: dropPostgresIndex,
    ddl: getPostgresDDL,
    create_view: createPostgresView,
    import: importPostgres,
    export: exportPostgres,
    objects: getPostgresObjects,
    object_definition: getPostgresObjectDefinition,
    save_object: savePostgresObject,
    drop_object: dropPostgresObject,
    create_foreign_key: createPostgresForeignKey,
    update_foreign_key: updatePostgresForeignKey,
    drop_foreign_key: dropPostgresForeignKey,
    tables: getPostgresTables,
    columns: getPostgresColumns,
    indexes: getPostgresIndexes,
    rows: getPostgresRows,
    query: queryPostgres,
    execute: executePostgres,
    insert: insertPostgres,
    update: updatePostgres,
    delete: deletePostgres,
    users: managePostgresUsers,
    maintenance: maintainPostgres,
    "system/task": getDatabaseTask,
} satisfies DatabaseRequests;

export const sqlserverRequests = {
    databases: getSqlServerDatabases,
    database_info: getSqlServerDatabaseInfo,
    create_database: createSqlServerDatabase,
    drop_database: dropSqlServerDatabase,
    rename_database: renameSqlServerDatabase,
    version: getSqlServerVersion,
    schemas: getSqlServerSchemas,
    create_table: createSqlServerTable,
    drop_table: dropSqlServerTable,
    rename_table: renameSqlServerTable,
    add_column: addSqlServerColumn,
    update_column: updateSqlServerColumn,
    drop_column: dropSqlServerColumn,
    create_index: createSqlServerIndex,
    drop_index: dropSqlServerIndex,
    ddl: getSqlServerDDL,
    create_view: createSqlServerView,
    import: importSqlServer,
    export: exportSqlServer,
    objects: getSqlServerObjects,
    object_definition: getSqlServerObjectDefinition,
    save_object: saveSqlServerObject,
    drop_object: dropSqlServerObject,
    create_foreign_key: createSqlServerForeignKey,
    update_foreign_key: updateSqlServerForeignKey,
    drop_foreign_key: dropSqlServerForeignKey,
    tables: getSqlServerTables,
    columns: getSqlServerColumns,
    indexes: getSqlServerIndexes,
    rows: getSqlServerRows,
    query: querySqlServer,
    execute: executeSqlServer,
    insert: insertSqlServer,
    update: updateSqlServer,
    delete: deleteSqlServer,
    users: manageSqlServerUsers,
    maintenance: maintainSqlServer,
    "system/task": getDatabaseTask,
} satisfies DatabaseRequests;

export const sqliteRequests = {
    databases: getSqliteDatabases,
    database_info: getSqliteDatabaseInfo,
    version: getSqliteVersion,
    schemas: getSqliteSchemas,
    create_table: createSqliteTable,
    drop_table: dropSqliteTable,
    rename_table: renameSqliteTable,
    add_column: addSqliteColumn,
    update_column: updateSqliteColumn,
    drop_column: dropSqliteColumn,
    create_index: createSqliteIndex,
    drop_index: dropSqliteIndex,
    ddl: getSqliteDDL,
    create_view: createSqliteView,
    import: importSqlite,
    export: exportSqlite,
    objects: getSqliteObjects,
    object_definition: getSqliteObjectDefinition,
    save_object: saveSqliteObject,
    drop_object: dropSqliteObject,
    create_foreign_key: createSqliteForeignKey,
    update_foreign_key: updateSqliteForeignKey,
    drop_foreign_key: dropSqliteForeignKey,
    tables: getSqliteTables,
    columns: getSqliteColumns,
    indexes: getSqliteIndexes,
    rows: getSqliteRows,
    query: querySqlite,
    execute: executeSqlite,
    insert: insertSqlite,
    update: updateSqlite,
    delete: deleteSqlite,
    maintenance: maintainSqlite,
    "system/task": getDatabaseTask,
} satisfies DatabaseRequests;

export const redisRequests = {
    databases: getRedisDatabases,
    database_info: getRedisDatabaseInfo,
    version: getRedisVersion,
    keys: getRedisKeys,
    value: getRedisValue,
    save: saveRedis,
    delete: deleteRedis,
    rename: renameRedis,
    expire: expireRedis,
    command: executeRedisCommand,
    info: getRedisInfo,
    users: manageRedisUsers,
    maintenance: maintainRedis,
    import: importRedis,
    export: exportRedis,
    "system/task": getDatabaseTask,
} satisfies DatabaseRequests;

export const mongoRequests = {
    databases: getMongoDatabases,
    database_info: getMongoDatabaseInfo,
    create_database: createMongoDatabase,
    drop_database: dropMongoDatabase,
    version: getMongoVersion,
    collections: getMongoCollections,
    indexes: getMongoIndexes,
    documents: getMongoDocuments,
    insert: insertMongo,
    update: updateMongo,
    delete: deleteMongo,
    aggregate: aggregateMongo,
    command: executeMongoCommand,
    create_collection: createMongoCollection,
    drop_collection: dropMongoCollection,
    create_index: createMongoIndex,
    drop_index: dropMongoIndex,
    users: manageMongoUsers,
    maintenance: maintainMongo,
    import: importMongo,
    export: exportMongo,
    create_view: createMongoView,
    update_view: updateMongoView,
    view_definition: getMongoViewDefinition,
    "system/task": getDatabaseTask,
} satisfies DatabaseRequests;

export const elasticsearchRequests = {
    version: getElasticsearchVersion,
    indices: getElasticsearchIndices,
    info: getElasticsearchInfo,
    create_index: createElasticsearchIndex,
    drop_index: dropElasticsearchIndex,
    documents: getElasticsearchDocuments,
    close_search: closeElasticsearchSearch,
    insert: insertElasticsearch,
    update: updateElasticsearch,
    delete: deleteElasticsearch,
    mapping: getElasticsearchMapping,
    analyzers: getElasticsearchAnalyzers,
    update_mapping: updateElasticsearchMapping,
    refresh_fields: refreshElasticsearchFields,
    settings: getElasticsearchSettings,
    update_settings: updateElasticsearchSettings,
    status: getElasticsearchStatus,
    request: executeElasticsearchRequest,
    "system/task": getDatabaseTask,
} satisfies DatabaseRequests;

export const databaseRequests: Record<number, DatabaseRequests> = {
    1: redisRequests,
    2: mysqlRequests,
    3: postgresRequests,
    4: mongoRequests,
    5: sqlserverRequests,
    6: sqliteRequests,
    7: elasticsearchRequests,
};
