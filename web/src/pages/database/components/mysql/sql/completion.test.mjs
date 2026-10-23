import assert from "node:assert/strict";
import {test} from "node:test";
import {completionReferences, quoteSqlIdentifier, sqlCompletionContext, sqlCompletionRange, sqlCompletions} from "./completion.ts";

const tables = [{name: "users", type: "BASE TABLE"}, {name: "orders", type: "BASE TABLE"}, {name: "active_users", type: "VIEW"}, {name: "order details"}];
const context = (marked) => {
    const offset = marked.indexOf("|");
    assert.notEqual(offset, -1);
    return sqlCompletionContext(marked.replace("|", ""), offset);
};

test("FROM、JOIN 以及逗号连接提示表名，限定当前数据库", () => {
    for (const sql of ["SELECT * FROM |", "SELECT * FROM us|", "SELECT * FROM users u JOIN |", "SELECT * FROM users, |", "UPDATE |", "INSERT INTO |", "SELECT * FROM app.|"]) {
        const parsed = context(sql);
        assert.equal(parsed.tablePosition, true, sql);
        const suggestions = sqlCompletions(parsed, tables, "app", []);
        assert.equal(suggestions.find((item) => item.caption === "active_users")?.meta, "视图");
        assert.equal(suggestions.find((item) => item.caption === "order details")?.value, "`order details`");
        assert.ok(suggestions.every((item) => ["表", "视图"].includes(item.meta)));
    }
    assert.deepEqual(sqlCompletions(context("SELECT * FROM other.|"), tables, "app", []), []);
});

test("读取光标后的 FROM，识别显式和隐式别名", () => {
    for (const sql of ["SELECT u.| FROM users u", "SELECT u.na| FROM users AS u", "SELECT `u`.| FROM `app`.`users` AS `u`"]) {
        const parsed = context(sql);
        assert.deepEqual(parsed.qualifiers, ["u"]);
        assert.deepEqual(completionReferences(parsed, tables, "app"), [{schema: "app", table: "users", alias: "u"}]);
    }
});

test("函数参数中仍能读取外层查询的字段", () => {
    const parsed = context("SELECT COUNT(u.|) FROM users u");
    assert.equal(parsed.tablePosition, false);
    assert.equal(completionReferences(parsed, tables, "app")[0]?.table, "users");
});

test("尚未关闭的反引号不吞掉光标后的表名和别名", () => {
    const parsed = context("SELECT u.`us|\nFROM users u");
    assert.deepEqual(parsed.qualifiers, ["u"]);
    assert.equal(completionReferences(parsed, tables, "app")[0]?.table, "users");
    assert.equal(context("SELECT * FROM `us|").tablePosition, true);
});

test("多条语句隔离上下文，字符串内的分号和注释不参与解析", () => {
    const parsed = context("SELECT * FROM orders; SELECT u.|, ';', 'FROM orders' FROM users u -- JOIN orders o\n;");
    assert.deepEqual(parsed.references, [{schema: undefined, table: "users", alias: "u"}]);
    assert.equal(context("SELECT ';', u.| FROM users u").references[0]?.table, "users");
});

test("字符串与三种注释内不弹出补全", () => {
    for (const sql of ["SELECT 'na|me' FROM users", 'SELECT "na|me" FROM users', "SELECT 'na|", "SELECT 1 -- na|", "SELECT 1 # na|", "SELECT /* na|me */ 1", "SELECT /* na|"]) {
        const parsed = context(sql);
        assert.equal(parsed.blocked, true, sql);
        assert.deepEqual(sqlCompletions(parsed, tables, "app", []), []);
    }
    assert.equal(context("SELECT 'name' |FROM users").blocked, false);
});

test("简单子查询的别名按作用域匹配，不泄露兄弟子查询的别名", () => {
    const parsed = context("SELECT (SELECT u.| FROM users u), (SELECT * FROM orders u) FROM orders o");
    assert.equal(completionReferences(parsed, tables, "app")[0]?.table, "users");
    assert.equal(parsed.references.filter((item) => item.alias === "u").length, 1);
});

test("未引用的未知别名不请求字段，切换表后使用新表", () => {
    assert.deepEqual(completionReferences(context("SELECT x.| FROM users u"), tables, "app", "orders"), []);
    assert.equal(completionReferences(context("SELECT |"), tables, "app", "orders")[0]?.table, "orders");
    assert.deepEqual(completionReferences(context("SELECT u.| FROM other.users u"), tables, "app"), [{schema: "other", table: "users", alias: "u"}]);
    assert.equal(completionReferences(context("SELECT app.users.|"), tables, "app")[0]?.table, "users");
});

test("多表字段包含所属表、类型和备注，插入时消除歧义", () => {
    const parsed = context("SELECT | FROM users u JOIN orders o ON u.id = o.user_id");
    const references = completionReferences(parsed, tables, "app");
    const suggestions = sqlCompletions(parsed, tables, "app", references.map((reference) => ({reference, columns: [{name: "id", type: "bigint", comment: "主键"}]})));
    assert.ok(suggestions.some((item) => item.value === "`u`.`id`" && item.meta === "bigint · u" && item.docText?.includes("主键")));
    assert.ok(suggestions.some((item) => item.value === "`o`.`id`"));
    assert.ok(!suggestions.some((item) => item.value === "`id`"));
});

test("点号后只提示字段，反引号完整转义", () => {
    const parsed = context("SELECT u.| FROM users u");
    const suggestions = sqlCompletions(parsed, tables, "app", [{reference: {table: "users", alias: "u"}, columns: [{name: "odd`name", type: "text"}]}]);
    assert.equal(suggestions.length, 1);
    assert.equal(suggestions[0].value, "`odd``name`");
    assert.equal(quoteSqlIdentifier("order details"), "`order details`");
});

test("基础提示包括关键字和函数，函数插入后光标进入括号", () => {
    const suggestions = sqlCompletions(context("SEL|"), tables, "app", []);
    assert.ok(suggestions.some((item) => item.caption === "SELECT" && item.value === "SELECT "));
    assert.ok(suggestions.some((item) => item.caption === "COUNT()" && item.value === "COUNT()" && item.cursorBack === 1));
});

test("替换完整标识符，保留点号并避免重复反引号", () => {
    for (const [line, column, start, end] of [["u.na", 4, 2, 4], ["u.`na", 5, 2, 5], ["u.`name`", 5, 2, 8], ["u.name", 4, 2, 6], ["u.`first name`", 10, 2, 14], ["u.", 2, 2, 2]]) {
        assert.deepEqual(sqlCompletionRange(line, column), {start, end}, line);
    }
});
