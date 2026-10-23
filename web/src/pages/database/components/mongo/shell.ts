import {parse} from "acorn";

// 只解析数据和白名单方法，不执行 JavaScript。JSON 始终从语法树生成，保留键顺序与 BSON 精度。
const maxLength = 1024 * 1024;
const expression = (source: string): any => {
    if (!source.trim()) throw new Error("请输入 mongosh 语句");
    if (source.length > maxLength) throw new Error("语句不能超过 1 MB");
    try {
        const program: any = parse(source, {ecmaVersion: 2022});
        if (program.body.length !== 1 || program.body[0].type !== "ExpressionStatement") throw new Error("每次只支持执行一条 mongosh 语句");
        return program.body[0].expression;
    } catch (error: any) {
        if (error instanceof RangeError) throw new Error("表达式嵌套层级过深");
        if (error instanceof SyntaxError) throw new Error(`语法错误：${error.message}`);
        throw error;
    }
};
const property = (node: any): string => {
    if (!node.computed && node.property?.type === "Identifier") return node.property.name;
    if (node.property?.type === "Literal" && typeof node.property.value === "string") return node.property.value;
    throw new Error("不支持动态计算方法或集合名称");
};
const entries = (node: any): [string, any][] => {
    if (node?.type !== "ObjectExpression") throw new Error("此参数需要填写对象，例如 {name: '张三'}");
    const seen = new Set<string>();
    return node.properties.map((item: any) => {
        if (item.type !== "Property" || item.computed || item.method || item.shorthand || item.kind !== "init") throw new Error("对象中不支持变量、展开、方法或计算属性");
        const name = item.key.type === "Identifier" ? item.key.name : String(item.key.value);
        if (name.includes("\0") || seen.has(name)) throw new Error(`字段“${name}”无效或重复`);
        seen.add(name);
        return [name, item.value];
    });
};
const objectJSON = (items: [string, string][]) => `{${items.map(([key, value]) => `${JSON.stringify(key)}:${value}`).join(",")}}`;
const assertArguments = (args: any[], min: number, max = min) => {
    if (args.length < min || args.length > max || args.some((arg) => arg.type === "SpreadElement"))
        throw new Error(`此方法需要 ${min === max ? min : `${min} 至 ${max}`} 个参数，不支持省略或展开参数`);
};
const stringValue = (node: any, label = "名称") => {
    if (node?.type !== "Literal" || typeof node.value !== "string" || node.value.includes("\0")) throw new Error(`${label}需要使用字符串`);
    return node.value as string;
};
const numberText = (node: any): string => {
    if (node?.type === "Literal" && typeof node.value === "number") return node.raw;
    if (node?.type === "UnaryExpression" && ["-", "+"].includes(node.operator) && node.argument.type === "Literal" && typeof node.argument.value === "number") return `${node.operator}${node.argument.raw}`;
    throw new Error("此参数需要数字");
};
const integerValue = (node: any, label = "数量") => {
    const value = Number(numberText(node));
    if (!Number.isSafeInteger(value)) throw new Error(`${label}需要填写安全范围内的整数`);
    return value;
};
const integerBSON = (text: string, wide: boolean) => {
    if (!/^[+-]?\d+$/.test(text)) throw new Error("BSON 整数格式不正确，请用字符串填写完整整数");
    const value = BigInt(text);
    if (value < BigInt(wide ? "-9223372036854775808" : "-2147483648") || value > BigInt(wide ? "9223372036854775807" : "2147483647")) throw new Error("BSON 整数超出范围");
    return JSON.stringify({[wide ? "$numberLong" : "$numberInt"]: value.toString()});
};
const dataJSON = (node: any, depth = 0): string => {
    if (depth > 80) throw new Error("数据嵌套层级过深");
    switch (node?.type) {
        case "ObjectExpression": return objectJSON(entries(node).map(([key, value]) => [key, dataJSON(value, depth + 1)]));
        case "ArrayExpression":
            if (node.elements.some((item: any) => !item || item.type === "SpreadElement")) throw new Error("数组不支持空位或展开表达式");
            return `[${node.elements.map((item: any) => dataJSON(item, depth + 1)).join(",")}]`;
        case "Literal": {
            if (node.regex) {
                if (/[^imsu]/.test(node.regex.flags)) throw new Error("MongoDB 正则仅支持 i、m、s、u 标志，请移除 g、y 或 d");
                return JSON.stringify({$regularExpression: {pattern: node.regex.pattern, options: [...node.regex.flags].sort().join("")}});
            }
            if (node.bigint !== undefined) return integerBSON(node.bigint, true);
            if (typeof node.value === "number") {
                if (!Number.isFinite(node.value)) throw new Error("数字超出范围，请使用 NumberDecimal('…')");
                if (Number.isInteger(node.value) && !Number.isSafeInteger(node.value)) throw new Error("大整数请使用 NumberLong('…')，避免精度丢失");
                return JSON.stringify(node.value);
            }
            return JSON.stringify(node.value);
        }
        case "UnaryExpression": {
            const text = numberText(node);
            const value = Number(text);
            if (!Number.isFinite(value) || Number.isInteger(value) && !Number.isSafeInteger(value)) throw new Error("数字超出安全范围，请使用 NumberLong('…') 或 NumberDecimal('…')");
            return Object.is(value, -0) ? "-0" : JSON.stringify(value);
        }
        case "CallExpression": case "NewExpression": {
            let name = node.callee.type === "Identifier" ? node.callee.name : "";
            if (node.callee.type === "MemberExpression" && !node.callee.computed && node.callee.object.type === "Identifier") {
                const base = node.callee.object.name;
                const method = node.callee.property.name;
                if (["Long", "Int32", "Decimal128", "ObjectId"].includes(base) && ["fromString", "createFromHexString"].includes(method)) name = base;
            }
            if (name === "RegExp" || name === "BSONRegExp") {
                assertArguments(node.arguments, 1, 2);
                const pattern = stringValue(node.arguments[0], "正则内容");
                const flags = node.arguments[1] ? stringValue(node.arguments[1], "正则标志") : "";
                if ((name === "BSONRegExp" ? /[^imsux]/ : /[^imsu]/).test(flags) || new Set(flags).size !== flags.length) throw new Error("不支持此正则标志，请使用 MongoDB 支持的 i、m、s、u、x");
                return JSON.stringify({$regularExpression: {pattern, options: [...flags].sort().join("")}});
            }
            if (!["ObjectId", "ISODate", "Date", "NumberLong", "Long", "NumberInt", "Int32", "NumberDecimal", "Decimal128"].includes(name)) throw new Error("仅支持 BSON 构造方法，不支持执行任意 JavaScript 函数");
            if (name === "Date" && node.type !== "NewExpression") throw new Error("Date() 返回文本；日期值请使用 new Date(...) 或 ISODate(...)");
            assertArguments(node.arguments, name === "Date" || name === "ISODate" ? 0 : 1, 1);
            const arg = node.arguments[0];
            if (name === "Date" || name === "ISODate") {
                const value = !arg ? new Date() : new Date(arg.type === "Literal" && typeof arg.value === "string" ? arg.value : integerValue(arg, "日期时间戳"));
                if (Number.isNaN(value.getTime())) throw new Error("日期时间无效，请使用 ISODate('2026-01-01T00:00:00Z')");
                return JSON.stringify({$date: {$numberLong: String(value.getTime())}});
            }
            const text = arg.type === "Literal" && typeof arg.value === "string" ? arg.value : numberText(arg);
            if (name === "ObjectId") {
                if (typeof arg.value !== "string" || !/^[a-f\d]{24}$/i.test(text)) throw new Error("ObjectId 需要 24 位十六进制字符串");
                return JSON.stringify({$oid: text.toLowerCase()});
            }
            if (["NumberLong", "Long", "NumberInt", "Int32"].includes(name)) return integerBSON(text, name === "NumberLong" || name === "Long");
            if (!["NaN", "Infinity", "-Infinity"].includes(text) && !/^[+-]?(?:\d+(?:\.\d*)?|\.\d+)(?:[eE][+-]?\d+)?$/.test(text)) throw new Error("NumberDecimal 需要有效的数字字符串");
            return JSON.stringify({$numberDecimal: text});
        }
        default: throw new Error("仅支持 mongosh 数据表达式，不支持变量、运算、循环或 JavaScript 脚本");
    }
};
const objectValue = (node: any) => { entries(node); return dataJSON(node); };
const options = (node: any, allowed: string[]) => {
    const result = new Map<string, string>();
    if (!node) return result;
    for (const [key, value] of entries(node)) {
        if (!allowed.includes(key)) throw new Error(`暂不支持选项“${key}”，请使用 db.runCommand 明确指定完整命令`);
        result.set(key, dataJSON(value));
    }
    return result;
};
const database = (node: any): string | undefined => {
    if (node?.type === "Identifier" && node.name === "db") return undefined;
    if (node?.type === "CallExpression" && node.callee.type === "MemberExpression" && property(node.callee) === "getSiblingDB" && node.callee.object.type === "Identifier" && node.callee.object.name === "db") {
        assertArguments(node.arguments, 1);
        const name = stringValue(node.arguments[0], "数据库名称");
        if (!name) throw new Error("数据库名称不能为空");
        return name;
    }
    throw new Error("语句必须使用 db 或 db.getSiblingDB('数据库') 开头");
};
const targetCollection = (node: any) => {
    if (node?.type === "CallExpression" && node.callee.type === "MemberExpression" && property(node.callee) === "getCollection") {
        assertArguments(node.arguments, 1);
        const name = stringValue(node.arguments[0], "集合名称");
        if (!name) throw new Error("集合名称不能为空");
        return {name, database: database(node.callee.object)};
    }
    const parts: string[] = [];
    let current = node;
    while (current?.type === "MemberExpression") { parts.unshift(property(current)); current = current.object; }
    if (!parts.length || parts.some((part) => !part)) throw new Error("请指定集合，例如 db.getCollection('users')");
    return {name: parts.join("."), database: database(current)};
};
const pipelineWrites = (node: any): boolean => {
    if (node?.type !== "ArrayExpression") throw new Error("聚合管道必须是数组");
    return node.elements.some((stage: any) => {
        const fields = entries(stage);
        if (fields.length !== 1) throw new Error("每个聚合步骤只能包含一个操作");
        return ["$out", "$merge"].includes(fields[0][0]);
    });
};

export const parseMongoShell = (source: string): {
    operation: "aggregate" | "command"; body: any; write: boolean; kind: string; collection?: string;
    query?: {filter: string; sort: string; projection: string; skip: number; limit: number; database_name?: string};
} => {
    let node = expression(source);
    const chain: {name: string; args: any[]}[] = [];
    while (node?.type === "CallExpression" && node.callee.type === "MemberExpression" && ["sort", "skip", "limit", "project", "projection", "toArray"].includes(property(node.callee))) {
        chain.unshift({name: property(node.callee), args: node.arguments});
        node = node.callee.object;
    }
    if (chain.some((item, index) => item.name === "toArray" && index !== chain.length - 1)) throw new Error("toArray() 必须放在语句最后，不能在转换为数组后继续设置游标选项");
    if (node?.type !== "CallExpression" || node.callee.type !== "MemberExpression" || node.optional || node.callee.optional) throw new Error("请输入完整方法调用，例如 db.users.find({})");
    const kind = property(node.callee);
    const args = node.arguments;
    if (kind === "runCommand") {
        if (chain.length) throw new Error("runCommand 不支持链式调用");
        assertArguments(args, 1);
        const fields = entries(args[0]);
        if (!fields.length) throw new Error("命令不能为空");
        const name = fields[0][0];
        const readOnly = new Set(["find", "count", "distinct", "ping", "hello", "isMaster", "ismaster", "buildInfo", "buildinfo", "serverStatus", "dbStats", "collStats", "listDatabases", "listCollections", "listIndexes", "usersInfo", "rolesInfo", "connectionStatus", "getParameter"]);
        let write = !readOnly.has(name);
        if (name === "aggregate") {
            const stages = fields.find(([key]) => key === "pipeline")?.[1];
            write = !stages || pipelineWrites(stages);
        }
        return {operation: "command", body: {command: objectValue(args[0]), ...(database(node.callee.object) ? {database_name: database(node.callee.object)} : {})}, write, kind};
    }
    const target = targetCollection(node.callee.object);
    const scope = target.database ? {database_name: target.database} : {};
    const command = (items: [string, string][], write = true): any => ({operation: "command", body: {command: objectJSON(items), ...scope}, write, kind, collection: target.name});
    if (["find", "findOne"].includes(kind)) {
        assertArguments(args, 0, 2);
        let filter = args[0] ? objectValue(args[0]) : "{}";
        let projection = args[1] ? objectValue(args[1]) : "{}";
        let sort = "{}";
        let skip = 0;
        let limit = kind === "findOne" ? 1 : 0;
        const seen = new Set<string>();
        for (const item of chain) {
            if (kind === "findOne") throw new Error("findOne 返回单个文档，不支持游标链式调用");
            const name = item.name === "projection" ? "project" : item.name;
            if (seen.has(name)) throw new Error(`请只设置一次 ${name}`);
            seen.add(name);
            assertArguments(item.args, name === "toArray" ? 0 : 1);
            if (name === "sort") sort = objectValue(item.args[0]);
            if (name === "project") {
                if (args[1]) throw new Error("find 已指定返回字段，请不要重复设置 project");
                projection = objectValue(item.args[0]);
            }
            if (name === "skip") { skip = integerValue(item.args[0]); if (skip < 0) throw new Error("跳过数量不能小于 0"); }
            if (name === "limit") limit = integerValue(item.args[0]);
        }
        const items: [string, string][] = [["find", JSON.stringify(target.name)], ["filter", filter]];
        if (projection !== "{}") items.push(["projection", projection]);
        if (sort !== "{}") items.push(["sort", sort]);
        if (skip) items.push(["skip", String(skip)]);
        if (limit) items.push(["limit", String(Math.abs(limit))]);
        if (limit < 0 || kind === "findOne") items.push(["singleBatch", "true"]);
        return {...command(items, false), query: {filter, projection, sort, skip, limit, ...scope}};
    }
    if (kind === "aggregate") {
        assertArguments(args, 1, 2);
        const write = pipelineWrites(args[0]);
        const stages = args[0].elements.map((item: any) => dataJSON(item));
        for (const item of chain) {
            if (item.name === "toArray") { assertArguments(item.args, 0); continue; }
            assertArguments(item.args, 1);
            const name = item.name === "projection" ? "project" : item.name;
            if (["limit", "skip"].includes(name)) {
                const value = integerValue(item.args[0]);
                if (value < (name === "limit" ? 1 : 0)) throw new Error(`聚合 ${name} 数量无效`);
                stages.push(`{"$${name}":${value}}`);
            } else stages.push(`{"$${name}":${objectValue(item.args[0])}}`);
        }
        if (write && chain.some((item) => item.name !== "toArray")) throw new Error("写入聚合后不能添加游标步骤");
        const pipeline = `[${stages.join(",")}]`;
        const extra = options(args[1], ["allowDiskUse", "maxTimeMS", "collation", "hint", "comment", "let", "bypassDocumentValidation", "readConcern", "writeConcern", "cursor"]);
        if (!write && !extra.size) return {operation: "aggregate", body: {collection: target.name, pipeline, limit: 500, ...scope}, write, kind, collection: target.name};
        return command([["aggregate", JSON.stringify(target.name)], ["pipeline", pipeline], ["cursor", extra.get("cursor") || "{}"], ...[...extra].filter(([key]) => key !== "cursor")], write);
    }
    if (chain.length) throw new Error(`不支持在 ${kind} 后使用游标方法`);
    switch (kind) {
        case "insertOne": case "insertMany": {
            assertArguments(args, 1, 2);
            if (kind === "insertMany" && (args[0].type !== "ArrayExpression" || !args[0].elements.length)) throw new Error("insertMany 需要非空文档数组");
            const documents = kind === "insertOne" ? `[${objectValue(args[0])}]` : `[${args[0].elements.map(objectValue).join(",")}]`;
            return command([["insert", JSON.stringify(target.name)], ["documents", documents], ...options(args[1], ["ordered", "bypassDocumentValidation", "writeConcern", "comment"]) ]);
        }
        case "updateOne": case "updateMany": case "replaceOne": {
            assertArguments(args, 2, 3);
            const query = objectValue(args[0]);
            if (kind === "replaceOne" && entries(args[1]).some(([key]) => key.startsWith("$"))) throw new Error("replaceOne 需要完整文档，不能使用更新操作符");
            if (kind !== "replaceOne" && args[1].type !== "ArrayExpression" && !entries(args[1]).every(([key]) => key.startsWith("$"))) throw new Error("更新需要 $set 等操作符，替换文档请使用 replaceOne");
            const extra = options(args[2], ["upsert", "arrayFilters", "collation", "hint", "bypassDocumentValidation", "writeConcern", "comment", "let", ...(kind === "updateOne" ? ["sort"] : [])]);
            const root = ["bypassDocumentValidation", "writeConcern", "comment", "let"];
            const update = objectJSON([["q", query], ["u", dataJSON(args[1])], ["multi", String(kind === "updateMany")], ...[...extra].filter(([key]) => !root.includes(key))]);
            return command([["update", JSON.stringify(target.name)], ["updates", `[${update}]`], ...[...extra].filter(([key]) => root.includes(key))]);
        }
        case "deleteOne": case "deleteMany": {
            assertArguments(args, 1, 2);
            const extra = options(args[1], ["collation", "hint", "let", "comment", "writeConcern"]);
            const root = ["let", "comment", "writeConcern"];
            const item = objectJSON([["q", objectValue(args[0])], ["limit", kind === "deleteOne" ? "1" : "0"], ...[...extra].filter(([key]) => !root.includes(key))]);
            return command([["delete", JSON.stringify(target.name)], ["deletes", `[${item}]`], ...[...extra].filter(([key]) => root.includes(key))]);
        }
        case "countDocuments": {
            assertArguments(args, 0, 2);
            const extra = options(args[1], ["skip", "limit", "maxTimeMS", "collation", "hint", "comment"]);
            const stages = [`{"$match":${args[0] ? objectValue(args[0]) : "{}"}}`];
            for (const name of ["skip", "limit"]) if (extra.has(name)) {
                const value = Number(extra.get(name));
                if (!Number.isSafeInteger(value) || value < 0) throw new Error(`${name} 需要非负整数`);
                if (value) stages.push(`{"$${name}":${value}}`);
                extra.delete(name);
            }
            stages.push('{"$group":{"_id":null,"count":{"$sum":1}}}');
            return command([["aggregate", JSON.stringify(target.name)], ["pipeline", `[${stages.join(",")}]`], ["cursor", "{}"], ...extra], false);
        }
        case "distinct":
            assertArguments(args, 1, 3);
            return command([["distinct", JSON.stringify(target.name)], ["key", JSON.stringify(stringValue(args[0], "字段名称"))], ["query", args[1] ? objectValue(args[1]) : "{}"], ...options(args[2], ["collation", "maxTimeMS", "comment", "readConcern"])], false);
        case "createIndex": {
            assertArguments(args, 1, 2);
            const keys = entries(args[0]);
            if (!keys.length) throw new Error("请填写索引字段");
            const extra = options(args[1], ["name", "unique", "sparse", "expireAfterSeconds", "partialFilterExpression", "collation", "wildcardProjection", "hidden", "background", "weights", "default_language", "language_override", "textIndexVersion", "2dsphereIndexVersion", "bits", "min", "max", "bucketSize", "storageEngine"]);
            const name = extra.get("name") || JSON.stringify(keys.map(([key, value]) => {
                const direction = JSON.parse(dataJSON(value));
                const text = typeof direction === "object" && direction !== null ? direction.$numberInt ?? direction.$numberLong ?? direction.$numberDouble : direction;
                if (!["1", "-1", "text", "hashed", "2d", "2dsphere", "geoHaystack", "columnstore"].includes(String(text))) throw new Error(`索引字段“${key}”的类型无效`);
                return `${key}_${text}`;
            }).join("_"));
            extra.delete("name");
            return command([["createIndexes", JSON.stringify(target.name)], ["indexes", `[${objectJSON([["key", objectValue(args[0])], ["name", name], ...extra])}]`]]);
        }
        case "dropIndex":
            assertArguments(args, 1, 2);
            return command([["dropIndexes", JSON.stringify(target.name)], ["index", args[0].type === "ObjectExpression" ? objectValue(args[0]) : JSON.stringify(stringValue(args[0], "索引名称"))], ...options(args[1], ["writeConcern", "maxTimeMS", "comment"])]);
        case "drop":
            assertArguments(args, 0, 1);
            return command([["drop", JSON.stringify(target.name)], ...options(args[0], ["writeConcern", "comment"]) ]);
        default: throw new Error(`暂不支持 ${kind} 方法，请使用 db.runCommand({...})；不支持执行任意 JavaScript 脚本`);
    }
};

// 从 JSON 语法树直接格式化，数字 token 和对象键顺序不会经过 JSON.parse 重建。
export const shellValue = (json: string): string => {
    if (json.length > maxLength) throw new Error("数据不能超过 1 MB");
    try { JSON.parse(json); } catch { throw new Error("JSON 格式不正确"); }
    const render = (node: any, depth = 0): string => {
        if (depth > 80) throw new Error("数据嵌套层级过深");
        const indent = "  ".repeat(depth);
        if (node.type === "ObjectExpression") {
            const values = entries(node);
            if (values.length === 1) {
                const [name, value] = values[0];
                const constructors: Record<string, string> = {$oid: "ObjectId", $numberLong: "NumberLong", $numberInt: "NumberInt", $numberDecimal: "NumberDecimal"};
                if (Object.prototype.hasOwnProperty.call(constructors, name) && value.type === "Literal" && typeof value.value === "string") return `${constructors[name]}(${JSON.stringify(value.value)})`;
                if (name === "$date") {
                    let time: string | number | undefined;
                    if (value.type === "Literal") time = value.value;
                    else if (value.type === "ObjectExpression") {
                        const nested = entries(value);
                        if (nested.length === 1 && nested[0][0] === "$numberLong") time = Number(nested[0][1].value);
                    }
                    if (time !== undefined && !Number.isNaN(new Date(time).getTime())) return `ISODate(${JSON.stringify(new Date(time).toISOString())})`;
                }
                if (name === "$regularExpression" && value.type === "ObjectExpression") {
                    const nested = new Map(entries(value));
                    const pattern = nested.get("pattern");
                    const flags = nested.get("options");
                    if (nested.size === 2 && typeof pattern?.value === "string" && typeof flags?.value === "string")
                        return `${flags.value.includes("x") ? "new BSONRegExp" : "RegExp"}(${JSON.stringify(pattern.value)}, ${JSON.stringify(flags.value)})`;
                }
            }
            if (!values.length) return "{}";
            return `{\n${values.map(([key, value]) => `${indent}  ${JSON.stringify(key)}: ${render(value, depth + 1)}`).join(",\n")}\n${indent}}`;
        }
        if (node.type === "ArrayExpression") return node.elements.length ? `[\n${node.elements.map((value: any) => `${indent}  ${render(value, depth + 1)}`).join(",\n")}\n${indent}]` : "[]";
        if (node.type === "Literal") {
            if (typeof node.value === "number") {
                if (!Number.isFinite(node.value) || Number.isInteger(node.value) && !Number.isSafeInteger(node.value)) {
                    const integral = /^\d+$/.test(node.raw) && BigInt(node.raw) <= BigInt("9223372036854775807");
                    return `${integral ? "NumberLong" : "NumberDecimal"}(${JSON.stringify(node.raw)})`;
                }
                return node.raw;
            }
            return JSON.stringify(node.value);
        }
        if (node.type === "UnaryExpression") {
            const text = numberText(node);
            if (!Number.isFinite(Number(text)) || Number.isInteger(Number(text)) && !Number.isSafeInteger(Number(text))) {
                const integral = /^-\d+$/.test(text) && BigInt(text) >= BigInt("-9223372036854775808");
                return `${integral ? "NumberLong" : "NumberDecimal"}(${JSON.stringify(text)})`;
            }
            return text;
        }
        throw new Error("JSON 数据包含不支持的表达式");
    };
    return render(expression(`(${json})`));
};

export const shellFind = (collection: string, query: {filter: string; sort: string; projection: string}): string => {
    let source = `db.getCollection(${JSON.stringify(collection)}).find(${shellValue(query.filter || "{}")}`;
    if ((query.projection || "{}").replace(/\s/g, "") !== "{}") source += `, ${shellValue(query.projection)}`;
    source += ")";
    if ((query.sort || "{}").replace(/\s/g, "") !== "{}") source += `.sort(${shellValue(query.sort)})`;
    return source;
};
