// 数字始终按文本处理，直到生成原始 JSON，避免 bigint / decimal 经过 JS Number。
export function encodeSqlValue(kind: string, value: any): string {
    if (kind === "null") return "null";
    if (kind === "number") {
        const match = String(value ?? "").trim().match(/^([+-]?)(?:(\d+)(?:\.(\d*))?|\.(\d+))(?:[eE]([+-]?\d+))?$/);
        if (!match) throw new Error("请输入有效数字");
        const integer = (match[2] || "0").replace(/^0+(?=\d)/, "");
        const fraction = match[3] || match[4];
        return `${match[1] === "-" ? "-" : ""}${integer}${fraction ? `.${fraction}` : ""}${match[5] ? `e${match[5]}` : ""}`;
    }
    if (kind === "boolean") {
        if (value === true || value === "true") return "true";
        if (value === false || value === "false") return "false";
        throw new Error("请选择布尔值");
    }
    if (kind === "binary") {
        const data = String(value ?? "").replace(/\s/g, "");
        if (!/^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$/.test(data)) throw new Error("请输入有效的 Base64 内容");
        return JSON.stringify({encoding: "base64", data});
    }
    return JSON.stringify(String(value ?? ""));
}

export function sqlColumnKind(column: any) {
    const type = String(column.type || "").trim().toLowerCase();
    if (type.endsWith("[]")) return "text";
    if (/\b(blob|binary|varbinary|bytea|image)\b|blob$/.test(type)) return "binary";
    if (/^(bool|boolean|bit)(\b|\()/.test(type) && !/^bit\s*\(/.test(type)) return "boolean";
    if (/^(tinyint|smallint|mediumint|int|integer|bigint|int2|int4|int8|decimal|numeric|number|real|float|double|smallserial|serial|bigserial)(\b|\()/.test(type)) return "number";
    return "text";
}

export function sqlFieldValue(column: any, value: any, editing: boolean) {
    const kind = value?.encoding === "base64" ? "binary" : sqlColumnKind(column);
    const content = kind === "binary" ? value?.data || "" : kind === "boolean" && value != null
        ? String(value === true || value === "true" || value === 1 || value === "1") : value == null ? "" : String(value);
    const mode = editing ? value == null ? "null" : "value"
        : column.generated || column.auto_increment || column.default != null || column.nullable ? "default" : "value";
    return {mode, kind, value: content};
}

export function encodeSqlFields(columns: any[], fields: any[], original?: any[]) {
    const changes: string[] = [];
    columns.forEach((column, index) => {
        const field = fields[index];
        if (!field || column.generated || field.mode === "default") return;
        if (original && field.mode === original[index].mode && field.value === original[index].value && field.kind === original[index].kind) return;
        if (field.mode === "null" && !column.nullable) throw new Error(`${column.name} 不允许 NULL`);
        try {
            changes.push(`${JSON.stringify(column.name)}:${encodeSqlValue(field.mode === "null" ? "null" : field.kind, field.value)}`);
        } catch (reason: any) { throw new Error(`${column.name}：${reason.message}`); }
    });
    return changes.length ? `{${changes.join(",")}}` : "";
}

export function encodeSqlFilters(filters: any[], columns: any[]) {
    if (filters.length > 64) throw new Error("筛选条件不能超过 64 个");
    const values = filters.map((filter, index) => {
        const column = columns.find((item) => item.name === filter.column);
        if (!column) throw new Error(`请选择第 ${index + 1} 个筛选字段`);
        const kind = filter.operator === "like" ? "text" : sqlColumnKind(column);
        let value = "null";
        try {
            if (filter.operator === "in") {
                if (!filter.value?.length) throw new Error("请至少输入一个值");
                if (filter.value.length > 1000) throw new Error("最多输入 1000 个值");
                value = `[${filter.value.map((item: string) => encodeSqlValue(kind, item)).join(",")}]`;
            } else if (filter.operator !== "null" && filter.operator !== "not_null") value = encodeSqlValue(kind, filter.value);
        } catch (reason: any) { throw new Error(`筛选条件 ${index + 1}：${reason.message}`); }
        return `{"column":${JSON.stringify(column.name)},"operator":${JSON.stringify(filter.operator)},"value":${value}}`;
    });
    return values.length ? `[${values.join(",")}]` : "";
}
