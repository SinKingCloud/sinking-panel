import {parseExpressionAt} from "acorn";

let sequence = 0;

export const mongoTypes = [
    ["string", "文本"], ["int32", "整数"], ["int64", "长整数"], ["double", "小数"], ["decimal", "高精度数"],
    ["boolean", "布尔值"], ["date", "日期时间"], ["objectId", "ObjectId"], ["object", "对象"], ["array", "数组"], ["null", "空值"],
].map(([value, label]) => ({value, label}));

export const newMongoField = (type = "string", key = ""): any => ({id: ++sequence, key, type,
    value: type === "boolean" ? false : ["int32", "int64", "double", "decimal"].includes(type) ? "0"
        : type === "date" ? new Date().toISOString() : "", children: []});

const decodeValue = (value: any, key: string): any => {
    const node = newMongoField("string", key);
    node.existing = true;
    if (value === null) return {...node, type: "null"};
    if (Array.isArray(value)) return {...node, type: "array", children: value.map((item, index) => decodeValue(item, String(index)))};
    if (typeof value === "object") {
        const names = Object.keys(value);
        if (names.length === 1) {
            const type = ({$oid: "objectId", $numberInt: "int32", $numberLong: "int64", $numberDouble: "double", $numberDecimal: "decimal"} as any)[names[0]];
            if (type && typeof value[names[0]] === "string") return {...node, type, value: value[names[0]]};
            if (names[0] === "$date") {
                const date = new Date(typeof value.$date === "object" ? Number(value.$date?.$numberLong) : value.$date);
                if (!Number.isNaN(date.getTime())) return {...node, type: "date", value: date.toISOString(), original: value};
            }
        }
        // 不支持的 BSON 类型保留原值，避免二进制、时间戳等被转成普通对象。
        if (names.some((name) => ["$binary", "$regularExpression", "$timestamp", "$minKey", "$maxKey", "$undefined", "$code", "$dbPointer", "$symbol", "$date"].includes(name)))
            return {...node, type: "raw", value};
        return {...node, type: "object", children: Object.entries(value).map(([name, item]) => decodeValue(item, name))};
    }
    if (typeof value === "boolean") return {...node, type: "boolean", value};
    if (typeof value === "number") return {...node, type: Number.isInteger(value) && value >= -2147483648 && value <= 2147483647 ? "int32" : "double", value: String(value)};
    return {...node, value: String(value)};
};

export const mongoFields = (document: any, lockId = false): any[] => Object.entries(document || {}).map(([key, value]) => ({
    ...decodeValue(value, key), locked: lockId && key === "_id",
}));

export const mongoValue = (node: any): any => {
    if (node.type === "raw") return node.value;
    const value = String(node.value ?? "");
    switch (node.type) {
        case "string": return value;
        case "null": return null;
        case "boolean": return Boolean(node.value);
        case "object": return mongoDocument(node.children);
        case "array": return (node.children || []).map(mongoValue);
        case "objectId":
            if (!/^[a-f\d]{24}$/i.test(value)) throw new Error(`字段“${node.key}”需要 24 位 ObjectId`);
            return {$oid: value};
        case "int32":
        case "int64": {
            if (!/^[+-]?\d+$/.test(value)) throw new Error(`字段“${node.key}”需要整数`);
            const number = BigInt(value);
            const min = BigInt(node.type === "int32" ? "-2147483648" : "-9223372036854775808");
            const max = BigInt(node.type === "int32" ? "2147483647" : "9223372036854775807");
            if (number < min || number > max) throw new Error(`字段“${node.key}”超出${node.type === "int32" ? "整数" : "长整数"}范围`);
            return {[node.type === "int32" ? "$numberInt" : "$numberLong"]: number.toString()};
        }
        case "double":
        case "decimal":
            if (!["NaN", "Infinity", "-Infinity"].includes(value) && (!/^[+-]?(?:\d+(?:\.\d*)?|\.\d+)(?:[eE][+-]?\d+)?$/.test(value)
                || (node.type === "double" && !Number.isFinite(Number(value))))) throw new Error(`字段“${node.key}”需要有效的数字`);
            return {[node.type === "decimal" ? "$numberDecimal" : "$numberDouble"]: value};
        case "date": {
            const date = new Date(value);
            if (Number.isNaN(date.getTime())) throw new Error(`请选择字段“${node.key}”的日期时间`);
            return {$date: {$numberLong: String(date.getTime())}};
        }
        default: throw new Error("不支持的字段类型");
    }
};

export const mongoDocument = (fields: any[]): Record<string, any> => {
    const document: Record<string, any> = Object.create(null);
    for (const node of fields || []) {
        if (!node.key && !node.existing) throw new Error("请填写字段名称");
        if (node.key.includes("\0")) throw new Error("字段名称不能包含空字符");
        if (Object.prototype.hasOwnProperty.call(document, node.key)) throw new Error(`字段“${node.key}”重复`);
        document[node.key] = mongoValue(node);
    }
    return document;
};

// 从原始数字 token 生成 Extended JSON，避免 JSON.parse 提前截断长整数。
export const mongoJSON = (text: string): Record<string, any> => {
    if (text.length > 16 << 20 || new TextEncoder().encode(text).byteLength > 16 << 20) throw new Error("文档不能超过 16 MB");
    try { JSON.parse(text); } catch { throw new Error("请输入有效的 JSON 文档"); }
    let root: any;
    try { root = parseExpressionAt(text, 0, {ecmaVersion: 2022}); }
    catch { throw new Error("JSON 文档格式不正确或嵌套层级过深"); }
    if (root.type !== "ObjectExpression") throw new Error("文档必须是 JSON 对象");
    const wrappers = ["$oid", "$numberInt", "$numberLong", "$numberDouble", "$numberDecimal", "$date", "$timestamp", "$binary", "$regularExpression", "$minKey", "$maxKey", "$undefined", "$symbol", "$dbPointer", "$uuid"];
    const read = (node: any, depth = 0, raw = false): any => {
        if (depth > 100) throw new Error("文档嵌套不能超过 100 层");
        if (node.type === "ObjectExpression") {
            const names = node.properties.map((item: any) => String(item.key.value));
            const wrapped = names.length === 1 && wrappers.includes(names[0])
                || names.includes("$code") && names.every((name: string) => ["$code", "$scope"].includes(name))
                || names.includes("$binary") && names.every((name: string) => ["$binary", "$type"].includes(name))
                || names.includes("$regex") && names.every((name: string) => ["$regex", "$options"].includes(name));
            const document: Record<string, any> = Object.create(null);
            node.properties.forEach((item: any, index: number) => {
                const name = names[index];
                if (name.includes("\0")) throw new Error("字段名称不能包含空字符");
                if (Object.prototype.hasOwnProperty.call(document, name)) throw new Error(`字段“${name}”重复`);
                document[name] = read(item.value, depth + 1, raw || wrapped);
            });
            return document;
        }
        if (node.type === "ArrayExpression") return node.elements.map((item: any) => read(item, depth + 1, raw));
        if (node.type === "Literal" && typeof node.value !== "number") return node.value;
        const source = text.slice(node.start, node.end);
        if (raw) {
            const value = Number(source);
            if (!Number.isFinite(value) || Number.isInteger(value) && !Number.isSafeInteger(value))
                throw new Error("Extended JSON 内的裸数字超出安全范围，请使用对应的字符串格式");
            return value;
        }
        if (!/[.eE]/.test(source)) {
            if (source.length > 20) throw new Error("整数超出 64 位范围，请使用 $numberDecimal 字符串");
            const value = BigInt(source);
            if (value < -9223372036854775808n || value > 9223372036854775807n) throw new Error("整数超出 64 位范围，请使用 $numberDecimal 字符串");
            return {[value >= -2147483648n && value <= 2147483647n ? "$numberInt" : "$numberLong"]: value.toString()};
        }
        if (!Number.isFinite(Number(source))) throw new Error("小数超出范围，请使用 $numberDecimal 字符串");
        return {$numberDouble: source};
    };
    return read(root);
};

// 按配置行顺序生成对象，复合索引与排序的字段顺序不能被重排。
export const mongoOrderedFields = (rows: any[], label: string) => {
    const seen = new Set<string>();
    return `{${rows.map((row: any) => {
        const field = row.field?.trim();
        if (!field) throw new Error(`请填写${label}字段`);
        if (field.includes("\0") || seen.has(field)) throw new Error(`${label}字段无效或重复`);
        seen.add(field);
        return `${JSON.stringify(field)}:${JSON.stringify(row.direction)}`;
    }).join(",")}}`;
};

export const mongoQuery = (filters: any[], sort: any[], projection: any[], projectionMode: number, logic = "all") => {
    // 未填写字段的草稿行不参与查询，预览和切换 mongosh 时同样忽略。
    filters = filters.filter((row) => row.field?.trim());
    sort = sort.filter((row) => row.field?.trim());
    projection = projection.filter((row) => row.field?.trim());
    if (!["all", "any"].includes(logic)) throw new Error("请选择筛选条件关系");
    const clauses = filters.map((row: any) => {
        const field = row.field?.trim();
        if (!field || field.startsWith("$") || field.includes("\0")) throw new Error("请填写有效的筛选字段");
        if (row.operator === "exists") return {[field]: {$exists: true}};
        if (row.operator === "missing") return {[field]: {$exists: false}};
        if (row.operator === "null") return {[field]: {$eq: null, $exists: true}};
        if (row.operator === "notNull") return {[field]: {$ne: null, $exists: true}};
        if (!["$eq", "$ne", "$gt", "$gte", "$lt", "$lte", "$in", "$nin", "$exists", "$regex", "contains", "starts", "ends"].includes(row.operator))
            throw new Error("请选择有效的筛选条件");
        let value;
        if (row.node?.json !== undefined) {
            try { value = JSON.parse(row.node.json); }
            catch { throw new Error(`字段“${field}”的 JSON 格式不正确`); }
            if (row.node.type === "object" && (!value || typeof value !== "object" || Array.isArray(value))) throw new Error(`字段“${field}”需要 JSON 对象`);
            if (row.node.type === "array" && !Array.isArray(value)) throw new Error(`字段“${field}”需要 JSON 数组`);
        } else value = mongoValue({...row.node, key: field});
        if (["$in", "$nin"].includes(row.operator) && (!Array.isArray(value) || !value.length)) throw new Error(`字段“${field}”至少需要一个值`);
        if (row.operator === "$exists" && typeof value !== "boolean") throw new Error("“存在”条件需要布尔值");
        if (["$regex", "contains", "starts", "ends"].includes(row.operator) && typeof value !== "string") throw new Error("文本条件需要填写文本值");
        if (["contains", "starts", "ends"].includes(row.operator)) {
            if (!value) throw new Error(`请填写字段“${field}”要查找的文本`);
            // 普通文本按字面匹配，不要求用户了解正则表达式。
            const pattern = value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
            return {[field]: {$regex: `${row.operator === "starts" ? "^" : ""}${pattern}${row.operator === "ends" ? "$" : ""}`}};
        }
        return {[field]: {[row.operator]: value}};
    });
    if (sort.some((row) => ![1, -1].includes(row.direction))) throw new Error("请选择有效的排序方式");
    if (![0, 1].includes(projectionMode)) throw new Error("请选择有效的字段显示方式");
    return {
        filter: JSON.stringify(clauses.length ? {[logic === "any" ? "$or" : "$and"]: clauses} : {}),
        sort: mongoOrderedFields(sort, "排序"),
        projection: mongoOrderedFields(projection.map((row) => ({...row, direction: projectionMode})), "返回"),
    };
};
