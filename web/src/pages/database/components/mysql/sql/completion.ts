export interface SqlTable {
    name: string;
    schema?: string;
    type?: string;
}

export interface SqlColumn {
    name: string;
    type?: string;
    comment?: string;
}

export interface SqlReference {
    schema?: string;
    table: string;
    alias?: string;
}

export interface SqlCompletionContext {
    blocked: boolean;
    tablePosition: boolean;
    columnPosition: boolean;
    qualifiers: string[];
    references: SqlReference[];
}

export interface SqlCompletion {
    caption: string;
    value: string;
    meta: string;
    score: number;
    docText?: string;
    cursorBack?: number;
}

type Token = {value: string; start: number; end: number; depth: number; kind: "word" | "symbol" | "string" | "comment"; quoted?: boolean; closed?: boolean};
const keywords = ["SELECT", "FROM", "WHERE", "JOIN", "LEFT JOIN", "RIGHT JOIN", "INNER JOIN", "CROSS JOIN", "ON", "AS", "AND", "OR", "NOT", "IN", "EXISTS", "BETWEEN", "LIKE", "IS NULL", "IS NOT NULL", "GROUP BY", "ORDER BY", "HAVING", "LIMIT", "OFFSET", "DISTINCT", "ASC", "DESC", "UNION ALL", "INSERT INTO", "VALUES", "UPDATE", "SET", "DELETE FROM", "CREATE TABLE", "ALTER TABLE", "CASE", "WHEN", "THEN", "ELSE", "END", "WITH", "NULL", "TRUE", "FALSE"];
const reserved = new Set([...keywords.flatMap((word) => word.split(" ")), "OUTER", "USING", "UNION", "FOR", "LOCK", "PROCEDURE", "PARTITION", "USE", "FORCE", "IGNORE", "INDEX", "WINDOW", "RETURNING", "NATURAL", "STRAIGHT_JOIN", "INTO", "TABLE", "VIEW", "DESCRIBE"]);
const functions: [string, string][] = [["COUNT", "统计数量"], ["SUM", "求和"], ["AVG", "平均值"], ["MIN", "最小值"], ["MAX", "最大值"], ["CONCAT", "拼接字符串"], ["COALESCE", "第一个非 NULL 值"], ["IFNULL", "替换 NULL 值"], ["NULLIF", "相等时返回 NULL"], ["IF", "条件判断"], ["NOW", "当前时间"], ["CURDATE", "当前日期"], ["DATE_FORMAT", "格式化日期"], ["DATE_ADD", "日期加减"], ["DATEDIFF", "日期差值"], ["ROUND", "四舍五入"], ["ABS", "绝对值"], ["LOWER", "转换小写"], ["UPPER", "转换大写"], ["TRIM", "去除首尾空格"], ["LENGTH", "字节长度"], ["CHAR_LENGTH", "字符长度"], ["SUBSTRING", "截取字符串"], ["REPLACE", "替换字符串"], ["JSON_EXTRACT", "读取 JSON 值"], ["JSON_UNQUOTE", "去除 JSON 引号"]];
const isWord = (character: string) => /[\p{L}\p{N}_$]/u.test(character);
const wordIs = (token: Token | undefined, ...values: string[]) => Boolean(token?.kind === "word" && !token.quoted && values.includes(token.value.toUpperCase()));
const identifier = (token: Token | undefined) => token?.kind === "word" && (token.quoted || !reserved.has(token.value.toUpperCase()));
export const quoteSqlIdentifier = (value: string) => `\`${value.replace(/`/g, "``")}\``;

// 保留位置与括号层级，字符串和注释中的 FROM、分号不会参与上下文判断。
function tokenize(sql: string) {
    const tokens: Token[] = [];
    let index = 0, depth = 0;
    while (index < sql.length) {
        if (/\s/.test(sql[index])) { index++; continue; }
        const start = index, character = sql[index];
        if (character === "#" || sql.slice(index, index + 2) === "--" && /\s/.test(sql[index + 2] || " ")) {
            while (index < sql.length && sql[index] !== "\n") index++;
            tokens.push({value: "", start, end: index, depth, kind: "comment", closed: index < sql.length});
        } else if (sql.slice(index, index + 2) === "/*") {
            const end = sql.indexOf("*/", index + 2);
            index = end < 0 ? sql.length : end + 2;
            tokens.push({value: "", start, end: index, depth, kind: "comment", closed: end >= 0});
        } else if (["'", '"', "`"].includes(character)) {
            let value = "", closed = false;
            index++;
            while (index < sql.length) {
                if (sql[index] === character) {
                    if (sql[index + 1] === character) { value += character; index += 2; continue; }
                    index++; closed = true; break;
                }
                if (character !== "`" && sql[index] === "\\" && index + 1 < sql.length) { value += sql[index + 1]; index += 2; }
                else value += sql[index++];
            }
            tokens.push({value, start, end: index, depth, kind: character === "`" ? "word" : "string", quoted: true, closed});
        } else if (isWord(character)) {
            while (index < sql.length && isWord(sql[index])) index++;
            tokens.push({value: sql.slice(start, index), start, end: index, depth, kind: "word"});
        } else {
            if (character === ")") depth = Math.max(0, depth - 1);
            tokens.push({value: character, start, end: ++index, depth, kind: "symbol"});
            if (character === "(") depth++;
        }
    }
    return tokens;
}

function readReferences(tokens: Token[]): SqlReference[] {
    const references: SqlReference[] = [];
    let fromList = false;
    for (let index = 0; index < tokens.length; index++) {
        const token = tokens[index];
        if (wordIs(token, "WHERE", "GROUP", "ORDER", "HAVING", "LIMIT", "SET", "VALUES", "UNION")) fromList = false;
        const startsTable = wordIs(token, "FROM", "JOIN", "UPDATE", "INTO", "STRAIGHT_JOIN") || token.value === "," && fromList;
        if (!startsTable) continue;
        if (wordIs(token, "FROM")) fromList = true;
        let next = index + 1;
        if (!identifier(tokens[next])) continue;
        const first = tokens[next++].value;
        let schema: string | undefined, table = first;
        if (tokens[next]?.value === "." && identifier(tokens[next + 1])) {
            schema = first; table = tokens[next + 1].value; next += 2;
        }
        if (wordIs(tokens[next], "AS")) next++;
        const alias = identifier(tokens[next]) ? tokens[next].value : undefined;
        references.push({schema, table, alias});
    }
    return references;
}

export function sqlCompletionContext(sql: string, offset: number): SqlCompletionContext {
    let all = tokenize(sql);
    // 用户尚未关闭反引号时，光标后面的 FROM 仍需参与字段和别名识别。
    if (all.some((token) => token.kind === "word" && token.quoted && !token.closed && token.start < offset && token.end >= offset)) {
        sql = `${sql.slice(0, offset)}\`${sql.slice(offset)}`;
        all = tokenize(sql);
    }
    const blocked = all.some((token) => ["string", "comment"].includes(token.kind) && token.start < offset
        && (offset < token.end || offset === token.end && !token.closed));
    const start = all.filter((token) => token.kind === "symbol" && token.value === ";" && token.end <= offset).at(-1)?.end || 0;
    const end = all.find((token) => token.kind === "symbol" && token.value === ";" && token.start >= offset)?.start ?? sql.length;
    const tokens = all.filter((token) => token.start >= start && token.end <= end && token.kind !== "comment" && token.kind !== "string");
    const currentWord = tokens.find((token) => token.kind === "word" && token.start < offset && token.end >= offset);
    const before = tokens.filter((token) => token.end <= (currentWord?.start ?? offset));
    let index = before.length - 1;
    const qualifiers: string[] = [];
    while (before[index]?.value === "." && before[index - 1]?.kind === "word") {
        qualifiers.unshift(before[index - 1].value); index -= 2;
    }
    const previous = before[index];
    const depth = before.reduce((depth, token) => token.kind !== "symbol" ? depth : token.value === "(" ? token.depth + 1 : token.value === ")" ? token.depth : depth, 0);
    const clauses = before.slice(0, index + 1).filter((token) => token.depth <= depth
        && wordIs(token, "SELECT", "FROM", "JOIN", "UPDATE", "INTO", "WHERE", "ON", "GROUP", "ORDER", "HAVING", "LIMIT", "SET", "VALUES", "UNION"));
    const clause = clauses.at(-1);
    const fromClause = clauses.filter((token) => token.depth === depth && !wordIs(token, "JOIN", "ON")).at(-1);
    const tablePosition = wordIs(previous, "FROM", "JOIN", "UPDATE", "INTO", "TABLE", "VIEW", "DESCRIBE", "STRAIGHT_JOIN")
        || previous?.value === "," && wordIs(fromClause, "FROM", "UPDATE");
    const insertColumns = wordIs(clause, "INTO") && before.some((token) => token.kind === "symbol" && token.value === "(" && token.depth === clause!.depth && token.start > clause!.end);
    const columnPosition = !tablePosition && (qualifiers.length > 0 || insertColumns || wordIs(clause, "SELECT", "WHERE", "ON", "GROUP", "ORDER", "HAVING", "SET", "VALUES"));
    const frames = [{depth: 0, start, end}];
    for (const token of tokens) {
        if (token.start >= offset) break;
        if (token.kind === "symbol" && token.value === "(") frames.push({depth: token.depth + 1, start: token.end, end});
        else if (token.kind === "symbol" && token.value === ")" && frames.length > 1) frames.pop();
    }
    const references: SqlReference[] = [];
    const aliases = new Set<string>();
    for (const frame of frames.reverse()) {
        const scopeStart = tokens.filter((token) => token.depth === frame.depth && token.start >= frame.start && token.end <= offset && wordIs(token, "UNION")).at(-1)?.end ?? frame.start;
        const scopeEnd = tokens.find((token) => token.start >= offset && (token.value === ")" && token.depth < frame.depth || token.depth === frame.depth && wordIs(token, "UNION")))?.start ?? end;
        for (const reference of readReferences(tokens.filter((token) => token.depth === frame.depth && token.start >= scopeStart && token.end <= scopeEnd))) {
            const key = (reference.alias || [reference.schema, reference.table].filter(Boolean).join(".")).toLowerCase();
            if (!aliases.has(key)) { aliases.add(key); references.push(reference); }
        }
    }
    return {blocked, tablePosition, columnPosition, qualifiers, references};
}

export function completionReferences(context: SqlCompletionContext, tables: SqlTable[], schema: string, selectedTable?: string) {
    const find = (name: string) => tables.find((table) => table.name === name)
        || tables.filter((table) => table.name.toLowerCase() === name.toLowerCase()).filter((_, _index, matches) => matches.length === 1)[0];
    let references = context.references;
    if (context.qualifiers.length) {
        const [first, second] = context.qualifiers;
        const matches = references.filter((reference) => (reference.alias || reference.table).toLowerCase() === first.toLowerCase());
        references = context.qualifiers.length === 1
            ? matches.length > 1 ? [] : [matches[0] || {table: first}]
            : context.qualifiers.length === 2 ? [{schema: first, table: second}] : [];
    } else if (!references.length && selectedTable) references = [{table: selectedTable}];
    return references.flatMap((reference) => {
        // 明确指定其他数据库时，表无需出现在当前数据库的导航列表里。
        if (reference.schema && reference.schema !== schema) return [reference];
        const table = find(reference.table);
        return table ? [{...reference, table: table.name, schema}] : [];
    });
}

export function sqlCompletions(context: SqlCompletionContext, tables: SqlTable[], schema: string, columns: {reference: SqlReference; columns: SqlColumn[]}[]): SqlCompletion[] {
    if (context.blocked) return [];
    if (context.tablePosition) {
        if (context.qualifiers.length && (context.qualifiers.length !== 1 || context.qualifiers[0] !== schema)) return [];
        return tables.map((table) => ({caption: table.name, value: quoteSqlIdentifier(table.name), meta: String(table.type).includes("VIEW") ? "视图" : "表", score: 1000, docText: `${schema}.${table.name}`}));
    }
    const qualified = context.qualifiers.length > 0;
    const ownerParts = (reference: SqlReference) => reference.alias ? [reference.alias]
        : context.references.some((other) => !other.alias && other.table === reference.table && (other.schema || schema) !== (reference.schema || schema))
            ? [reference.schema || schema, reference.table] : [reference.table];
    const items: SqlCompletion[] = (context.columnPosition || qualified ? columns : []).flatMap(({reference, columns}) => columns.map((column) => {
        const parts = ownerParts(reference), owner = parts.join(".");
        const name = !qualified && context.references.length > 1 ? `${parts.map(quoteSqlIdentifier).join(".")}.${quoteSqlIdentifier(column.name)}` : quoteSqlIdentifier(column.name);
        return {caption: !qualified && context.references.length > 1 ? `${column.name} · ${owner}` : column.name, value: name,
            meta: [column.type, owner].filter(Boolean).join(" · "), score: 1200,
            docText: [column.name, column.type, `所属表：${reference.schema || schema}.${reference.table}${reference.alias ? `（${reference.alias}）` : ""}`, column.comment].filter(Boolean).join("\n")};
    }));
    if (!qualified) {
        items.push(...context.references.map((reference) => ({caption: ownerParts(reference).join("."), value: `${ownerParts(reference).map(quoteSqlIdentifier).join(".")}.`, meta: reference.alias ? `别名 · ${reference.table}` : "表", score: 900})));
        items.push(...keywords.map((keyword) => ({caption: keyword, value: `${keyword} `, meta: "关键字", score: 500})));
        items.push(...functions.map(([name, description]) => ({caption: `${name}()`, value: `${name}()`, meta: "函数", score: 600, cursorBack: 1, docText: description})));
    }
    return [...new Map(items.map((item) => [item.value, item])).values()];
}

// 补全整个标识符，兼容已输入的反引号，避免生成重复引号或保留旧字段后缀。
export function sqlCompletionRange(line: string, column: number) {
    const tokens = tokenize(line);
    const word = tokens.find((token) => token.kind === "word" && token.start <= column && token.end >= column);
    const end = word?.quoted && !word.closed ? column + (line.slice(column).match(/^[\p{L}\p{N}_$]*/u)?.[0].length || 0) : word?.end ?? column;
    return {start: word?.start ?? column, end};
}

// 复用已有标点和关键词后缀，支持在标识符或函数名中间确认补全。
export function sqlCompletionEdit(line: string, column: number, item: SqlCompletion) {
    let {start, end} = sqlCompletionRange(line, column);
    let value = item.value, cursor: number | undefined;
    if (item.meta === "关键字") {
        const words = value.trim().split(/\s+/);
        let preceding = 0;
        for (let count = words.length - 1; count > 0; count--) {
            const match = line.slice(0, start).match(new RegExp(`\\b${words.slice(0, count).join("\\s+")}\\s+$`, "i"));
            if (match) { start = match.index!; preceding = count; break; }
        }
        const remaining = words.slice(preceding + 1);
        if (remaining.length) {
            const match = line.slice(end).match(new RegExp(`^\\s+${remaining.join("\\s+")}\\b`, "i"));
            if (match) end += match[0].length;
        }
        if (/^[ \t]/.test(line.slice(end))) { value = value.trimEnd(); cursor = start + value.length + 1; }
    } else if (item.cursorBack && value.endsWith("()")) {
        const opening = line.slice(end).match(/^[ \t]*\(/);
        if (opening) { value = value.slice(0, -2); cursor = start + value.length + opening[0].length; }
        else if (line[end] === ")") { value = value.slice(0, -1); cursor = start + value.length; }
    } else if (value.endsWith(".") && line[end] === ".") {
        value = value.slice(0, -1); cursor = start + value.length + 1;
    } else if (["表", "视图"].includes(item.meta) && !value.endsWith(".") && end === line.length) {
        value += " ";
    }
    return {start, end, value, cursor: cursor ?? start + value.length - (item.cursorBack || 0)};
}
