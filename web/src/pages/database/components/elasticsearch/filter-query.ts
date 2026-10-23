const numberTypes = new Set(["byte", "short", "integer", "long", "unsigned_long", "half_float", "float", "double", "scaled_float", "token_count"]);
const integerTypes = new Set(["byte", "short", "integer", "long", "unsigned_long", "token_count"]);
const textTypes = new Set(["text", "string", "match_only_text", "search_as_you_type"]);
const keywordTypes = new Set(["keyword", "constant_keyword", "wildcard", "version"]);

// 同时接受 properties、完整 mapping 和旧版按文档类型返回的 mapping。
export const getFilterFields = (mapping: any, fallback: string[] = []) => {
    const fields = new Map<string, any>();
    const visit = (properties: any, prefix = "", nested: string[] = []) => {
        Object.entries(properties || {}).forEach(([key, value]: any) => {
            if (!value || typeof value !== "object" || Array.isArray(value)) return;
            const name = prefix ? `${prefix}.${key}` : key;
            const type = value.type || (value.properties ? "object" : "unknown");
            const paths = type === "nested" ? [...nested, name] : nested;
            const exactName = Object.keys(value.fields || {}).find((key) => keywordTypes.has(value.fields[key]?.type) || value.fields[key]?.index === "not_analyzed");
            const exact = textTypes.has(type) ? value.index === "not_analyzed" ? name : exactName ? `${name}.${exactName}` : undefined : name;
            const item = {name, type, nested: paths, exact};
            const previous = fields.get(name);
            if (!previous) fields.set(name, item);
            else if (previous.type !== type || previous.exact !== exact || previous.nested.join("\0") !== paths.join("\0")) {
                fields.set(name, {...item, type: "conflict", exact: undefined, ambiguous: true});
            }
            if (value.properties) visit(value.properties, name, paths);
            if (value.fields) visit(value.fields, name, paths);
        });
    };
    const roots = (value: any, wrapped = false) => {
        if (!value || typeof value !== "object" || Array.isArray(value)) return;
        if (value.properties) { visit(value.properties); return; }
        if (value.mappings) { roots(value.mappings, true); return; }
        const children = Object.values(value) as any[];
        if (children.some((child) => child?.mappings)) children.forEach((child) => roots(child?.mappings, true));
        else if (wrapped) children.forEach((child) => { if (child?.properties) visit(child.properties); });
        else visit(value);
    };
    roots(mapping);
    fallback.forEach((name) => { if (!fields.has(name)) fields.set(name, {name, type: "unknown", nested: [], exact: name}); });
    return [...fields.values()];
};

export const getFilterOperators = (field?: any) => {
    if (field?.ambiguous) return [];
    const type = field?.type || "unknown";
    const labels: Record<string, string> = {match: "匹配文本", eq: "等于", neq: "不等于", gt: "大于", gte: "大于等于", lt: "小于", lte: "小于等于", between: "介于", exists: "存在", missing: "不存在", prefix: "开头是"};
    let values: string[] = [];
    if (numberTypes.has(type) || ["date", "date_nanos"].includes(type)) values = ["eq", "neq", "gt", "gte", "lt", "lte", "between"];
    else if (["boolean", "ip"].includes(type)) values = ["eq", "neq"];
    else if (textTypes.has(type)) values = field?.exact ? ["match", "eq", "neq", "prefix"] : ["match"];
    else if (keywordTypes.has(type)) values = ["match", "eq", "neq", "prefix"];
    else if (type === "unknown") values = ["match", "eq", "neq", "gt", "gte", "lt", "lte", "between", "prefix"];
    return [...values, "exists", "missing"].map((value) => ({value, label: labels[value]}));
};

export const buildFilterQuery = (form: any, fields: any[], options: {legacy?: boolean} = {}): any => {
    const logic = form?.logic === "any" ? "any" : "all";
    const combine = (clauses: any[]) => clauses.length === 1 ? clauses[0] : logic === "any" ? {bool: {should: clauses, minimum_should_match: 1}} : {bool: {must: clauses}};
    const clauses = (form?.conditions || []).map((condition: any, index: number) => {
        const name = String(condition.field || "").trim();
        const label = `第 ${index + 1} 条条件`;
        if (!name) throw new Error(`${label}请选择或填写字段`);
        const field = fields.find((item) => item.name === name) || {name, type: "unknown", nested: [], exact: name};
        const operator = condition.operator;
        if (!getFilterOperators(field).some((item) => item.value === operator)) throw new Error(`${label}的字段不支持此条件，请使用 DSL`);
        const value = (input: any): any => {
            if (input === undefined || input === null || input === "") throw new Error(`${label}请输入条件值`);
            if (!["string", "number", "boolean"].includes(typeof input)) throw new Error(`${label}的条件值不正确`);
            if (typeof input === "number" && (!Number.isFinite(input) || (Number.isInteger(input) && !Number.isSafeInteger(input)))) throw new Error(`${label}请用完整文本填写大整数，避免精度丢失`);
            if (field.type === "boolean") {
                if (input === true || input === "true") return true;
                if (input === false || input === "false") return false;
                throw new Error(`${label}请选择是或否`);
            }
            if (numberTypes.has(field.type)) {
                const text = String(input).trim();
                if (!(integerTypes.has(field.type) ? /^[+-]?\d+$/ : /^[+-]?(?:\d+(?:\.\d*)?|\.\d+)(?:e[+-]?\d+)?$/i).test(text) || !Number.isFinite(Number(text))) throw new Error(`${label}请输入有效${integerTypes.has(field.type) ? "整数" : "数字"}`);
                // ES 接受数值文本；超出 JS 安全整数范围时保留原文。
                return Math.abs(Number(text)) > Number.MAX_SAFE_INTEGER ? text : Number(text);
            }
            return input;
        };
        let query: any;
        if (["exists", "missing"].includes(operator)) {
            query = field.type === "nested" ? {match_all: {}} : {exists: {field: name}};
            if (options.legacy && field.type !== "nested") query = {constant_score: {filter: query}};
        } else {
            const input = value(condition.value);
            const target = ["eq", "neq", "prefix"].includes(operator) ? field.exact || name : name;
            if (operator === "match") query = {match: {[name]: input}};
            else if (operator === "eq" || operator === "neq") query = {term: {[target]: input}};
            else if (operator === "prefix") query = {prefix: {[target]: String(input)}};
            else if (operator === "between") {
                const end = value(condition.valueTo);
                if (numberTypes.has(field.type) && (integerTypes.has(field.type) ? BigInt(String(input)) > BigInt(String(end)) : Number(input) > Number(end))) throw new Error(`${label}的起始值不能大于结束值`);
                query = {range: {[name]: {gte: input, lte: end}}};
            } else query = {range: {[name]: {[operator]: input}}};
        }
        return {query, nested: field.nested, negative: operator === "neq" || operator === "missing"};
    });
    if (!clauses.length) return {};
    // 同一 nested 路径的正向条件必须匹配同一个嵌套对象。
    const group = (items: any[], depth = 0): any => {
        const direct: any[] = [];
        const nested = new Map<string, any[]>();
        items.forEach((item) => {
            const path = item.nested[depth];
            if (!path) direct.push(item.query);
            else { if (!nested.has(path)) nested.set(path, []); nested.get(path)!.push(item); }
        });
        nested.forEach((children, path) => direct.push({nested: {path, query: group(children, depth + 1)}}));
        return combine(direct);
    };
    const positive = clauses.filter((item: any) => !item.negative);
    const negative = clauses.filter((item: any) => item.negative).map((item: any) => ({bool: {must_not: [group([item])]}}));
    // 否定 nested 条件置于外层，避免被同一文档中的另一个嵌套对象误匹配。
    if (!negative.length) return {query: group(positive)};
    const query = positive.length ? group(positive) : undefined;
    const key = logic === "any" ? "should" : "must";
    const direct = query?.bool?.[key] && Object.keys(query.bool).every((name) => [key, "minimum_should_match"].includes(name)) ? query.bool[key] : query ? [query] : [];
    return {query: combine([...direct, ...negative])};
};

// 只回填可以完整重建的查询；额外搜索参数、分组和高级 DSL 留在 DSL 模式。
export const parseFilterQuery = (input: any, fields: any[], options: {legacy?: boolean} = {}): any | null => {
    try {
        const body = typeof input === "string" ? JSON.parse(input) : input;
        if (!body || typeof body !== "object" || Array.isArray(body)) return null;
        if (!Object.keys(body).length) return {logic: "all", conditions: []};
        if (Object.keys(body).length !== 1 || !body.query) return null;
        if (JSON.stringify(body.query) === '{"match_all":{}}') return {logic: "all", conditions: []};
        const queryLogic = (query: any): string => query?.nested ? queryLogic(query.nested.query) : query?.bool?.should ? "any" : "all";
        const logic = queryLogic(body.query);
        const conditions: any[] = [];
        const read = (query: any, paths: string[] = [], negative = false): boolean => {
            if (!query || typeof query !== "object" || Array.isArray(query) || Object.keys(query).length !== 1) return false;
            if (query.nested) {
                const {path, query: child, ...rest} = query.nested;
                return typeof path === "string" && !Object.keys(rest).length && read(child, [...paths, path], negative);
            }
            if (query.bool) {
                const bool = query.bool;
                if (bool.must_not && !negative && !Object.keys(bool).some((key) => key !== "must_not")) {
                    const items = Array.isArray(bool.must_not) ? bool.must_not : [bool.must_not];
                    return !paths.length && items.length === 1 && read(items[0], paths, true);
                }
                const key = logic === "any" ? "should" : "must";
                return !negative && Array.isArray(bool[key]) && bool[key].length > 0 && Object.keys(bool).every((name) => name === key || (key === "should" && name === "minimum_should_match" && bool[name] === 1)) && bool[key].every((child: any) => read(child, paths));
            }
            if (query.constant_score) {
                if (!options.legacy || Object.keys(query.constant_score).length !== 1 || !query.constant_score.filter?.exists) return false;
                return read(query.constant_score.filter, paths, negative);
            }
            let condition: any;
            if (query.match_all && !Object.keys(query.match_all).length && paths.length) condition = {field: paths[paths.length - 1], operator: negative ? "missing" : "exists"};
            else if (query.exists && Object.keys(query.exists).length === 1 && typeof query.exists.field === "string") condition = {field: query.exists.field, operator: negative ? "missing" : "exists"};
            else {
                const type = Object.keys(query)[0];
                if (!["match", "term", "prefix", "range"].includes(type) || (negative && type !== "term")) return false;
                const entries = Object.entries(query[type] || {});
                if (entries.length !== 1) return false;
                const [field, value]: any = entries[0];
                if (type === "range") {
                    if (!value || typeof value !== "object" || Array.isArray(value)) return false;
                    const keys = Object.keys(value);
                    if (keys.length === 2 && keys.includes("gte") && keys.includes("lte")) condition = {field, operator: "between", value: value.gte, valueTo: value.lte};
                    else if (keys.length === 1 && ["gt", "gte", "lt", "lte"].includes(keys[0])) condition = {field, operator: keys[0], value: value[keys[0]]};
                    else return false;
                } else condition = {field, operator: type === "term" ? negative ? "neq" : "eq" : type, value};
            }
            const field = fields.find((item) => item.name === condition.field);
            if (JSON.stringify(field?.nested || []) !== JSON.stringify(paths)) return false;
            conditions.push(condition);
            return true;
        };
        if (!read(body.query)) return null;
        const form = {logic, conditions};
        const stable = (value: any): string => Array.isArray(value) ? `[${value.map(stable).join(",")}]` : value && typeof value === "object" ? `{${Object.keys(value).sort().map((key) => `${JSON.stringify(key)}:${stable(value[key])}`).join(",")}}` : JSON.stringify(value);
        return stable(buildFilterQuery(form, fields, options)) === stable(body) ? form : null;
    } catch { return null; }
};
