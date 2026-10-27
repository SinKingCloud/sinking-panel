import {useCallback, useEffect, useRef} from "react";
import type {DatabasePanelRecord, MysqlApi} from "../entities";
import {completionReferences, sqlCompletionContext, sqlCompletionEdit, sqlCompletions} from "./completion";
import type {SqlColumn, SqlCompletion, SqlReference, SqlTable} from "./completion";

interface CompletionOptions {
    api: MysqlApi;
    record: DatabasePanelRecord;
    schema?: string;
    table?: string;
    tables: SqlTable[];
    columns: SqlColumn[];
    revision?: string;
}

const cacheKey = ({record, schema, revision}: CompletionOptions) => JSON.stringify([record.id, record.database_name || record.name, schema, revision]);
type MetadataEntry<T> = {expires: number; promise: Promise<T[]>};

function cachedMetadata<T>(cache: Map<string, MetadataEntry<T>>, key: string, request: () => Promise<unknown>) {
    let entry = cache.get(key);
    if (!entry || entry.expires <= Date.now()) {
        // pending 请求不按 TTL 淘汰；连续输入、同表多个别名共享同一个请求。
        const pending: MetadataEntry<T> = {expires: Infinity, promise: Promise.resolve([])};
        pending.promise = Promise.resolve().then(request).then((data) => {
            if (!Array.isArray(data)) throw new Error("元数据信息不完整");
            pending.expires = Date.now() + 60000;
            return data.filter((item: any) => typeof item?.name === "string") as T[];
        }).catch(() => { pending.expires = Date.now() + 5000; return []; });
        if (cache.size >= 200) {
            const oldest = [...cache].find(([, entry]) => entry.expires !== Infinity)?.[0];
            if (oldest) cache.delete(oldest);
        }
        cache.set(key, pending);
        entry = pending;
    }
    return entry.promise;
}

export default function useSqlCompletion(options: CompletionOptions) {
    const current = useRef(options);
    current.current = options;
    const editorRef = useRef<any>(undefined);
    const cleanup = useRef<(() => void) | undefined>(undefined);
    const cache = useRef({key: "", tables: new Map<string, MetadataEntry<SqlTable>>(), columns: new Map<string, MetadataEntry<SqlColumn>>()});
    const key = cacheKey(options);
    if (cache.current.key !== key) cache.current = {key, tables: new Map(), columns: new Map()};

    const attach = useCallback((editor: any) => {
        cleanup.current?.();
        editorRef.current = editor;
        let disposed = false;
        let requestId = 0;
        const previous = editor.completers;
        const previousCommand = editor.commands.commands?.startAutocomplete;
        const completionCommand = previousCommand && {...previousCommand, bindKey: {win: "Alt-/", mac: "Alt-/"}};
        // 只替换当前 SQL 编辑器的快捷键，避开系统输入法常用的 Ctrl + Space。
        if (completionCommand) editor.commands.addCommand(completionCommand);
        const showSuggestions = () => {
            editor.execCommand("startAutocomplete");
            // 自动弹出的空前缀列表不抢占回车换行；方向键可继续选择候选。
            if (editor.completer) {
                editor.completer.autoSelect = false;
                if (editor.completer.popup) {
                    editor.completer.popup.autoSelect = false;
                    editor.completer.popup.setRow(-1);
                }
            }
        };
        const insertMatch = (editor: any, item: SqlCompletion) => {
            const position = editor.getCursorPosition();
            const edit = sqlCompletionEdit(editor.session.getLine(position.row), position.column, item);
            editor.session.replace({start: {row: position.row, column: edit.start}, end: {row: position.row, column: edit.end}}, edit.value);
            editor.moveCursorTo(position.row, edit.cursor);
            editor.clearSelection();
            if (item.value.endsWith(".") || item.cursorBack || item.meta === "关键字") {
                const session = editor.session, sql = session.getValue(), requestCache = cache.current;
                // ACE 在 insertMatch 返回后关闭旧列表；等其结束后再提示字段/下一段 SQL。
                queueMicrotask(() => {
                    if (disposed || cache.current !== requestCache || editor.session !== session || session.getValue() !== sql) return;
                    const cursor = editor.getCursorPosition();
                    if (cursor.row !== position.row || cursor.column !== edit.cursor) return;
                    const context = sqlCompletionContext(sql, session.doc.positionToIndex(cursor, 0));
                    if (!context.blocked && (context.tablePosition || context.columnPosition)) showSuggestions();
                });
            }
        };
        const provider = {
            id: "mysql-schema",
            identifierRegexps: [/[\p{L}\p{N}_$]/u],
            triggerCharacters: ["."],
            cancel: () => { requestId++; },
            getCompletions: (_: any, session: any, position: {row: number; column: number}, _prefix: string, callback: (error: null, items: SqlCompletion[]) => void) => {
                const id = ++requestId;
                const options = current.current;
                const requestCache = cache.current;
                const sql = session.getValue();
                const offset = session.doc.positionToIndex(position, 0);
                const context = sqlCompletionContext(sql, offset);
                const schema = options.schema ?? options.record.database_name ?? options.record.name;
                const valid = () => !disposed && requestId === id && cache.current === requestCache && current.current.table === options.table
                    && current.current.tables === options.tables && current.current.api === options.api && editor.session === session
                    && session.getValue() === sql && editor.getCursorPosition().row === position.row && editor.getCursorPosition().column === position.column;
                const finish = (columns: {reference: SqlReference; columns: SqlColumn[]}[], tables = options.tables, tableSchema = schema) => {
                    // 旧请求回传空数组也会关闭 ACE 正在等待的新提示，因此过期回调直接丢弃。
                    if (valid()) callback(null, sqlCompletions(context, tables, tableSchema, columns).map((item) => ({...item, completer: {insertMatch}})));
                };
                if (context.blocked) { finish([]); return; }
                if (context.tablePosition) {
                    const targetSchema = context.qualifiers.length === 1 ? context.qualifiers[0] : schema;
                    if (targetSchema === schema || context.qualifiers.length > 1) finish([]);
                    else void cachedMetadata(requestCache.tables, targetSchema, () => options.api.tables(options.record, {database_name: targetSchema, schema: targetSchema}))
                        .then((tables) => finish([], tables, targetSchema));
                    return;
                }
                if (!context.columnPosition) { finish([]); return; }
                const references = completionReferences(context, options.tables, schema, options.table);
                const load = (reference: SqlReference) => {
                    const targetSchema = reference.schema || schema;
                    const promise = targetSchema === schema && reference.table === options.table && options.columns.length
                        ? Promise.resolve(options.columns)
                        : cachedMetadata(requestCache.columns, JSON.stringify([targetSchema, reference.table]),
                            () => options.api.columns(options.record, {database_name: targetSchema, schema: targetSchema, table: reference.table}));
                    return promise.then((columns) => ({reference, columns}));
                };
                void Promise.all(references.map(load)).then(finish);
            },
            insertMatch,
        };
        // 仅替换此 SQL 编辑器的补全器，避免数据库名称进入其他文件编辑器的提示。
        editor.completers = [provider];
        const afterExec = (event: any) => {
            if (event.command?.name !== "insertstring") return;
            if (/[\p{L}\p{N}_$]$/u.test(event.args || "")) {
                // 开始输入标识符后恢复首项选择，可直接按回车确认。
                if (editor.completer?.autoSelect === false) {
                    editor.completer.autoSelect = true;
                    if (editor.completer.popup) {
                        editor.completer.popup.autoSelect = true;
                        editor.completer.popup.setRow(0);
                    }
                }
                return;
            }
            if (![" ", ",", "("].includes(event.args)) return;
            const offset = editor.session.doc.positionToIndex(editor.getCursorPosition(), 0);
            const sql = editor.session.getValue(), context = sqlCompletionContext(sql, offset);
            const startsExpression = event.args !== " " || /(?:\b(?:SELECT|DISTINCT|WHERE|ON|AND|OR|NOT|BY|HAVING|SET|WHEN|THEN|ELSE)|[=<>!+*/-])\s+$/i.test(sql.slice(0, offset));
            if (!context.blocked && (context.tablePosition || context.columnPosition && startsExpression)) showSuggestions();
        };
        editor.commands.on("afterExec", afterExec);
        cleanup.current = () => {
            disposed = true;
            editor.commands.off("afterExec", afterExec);
            if (editor.completers?.includes(provider)) editor.completers = previous;
            if (completionCommand && editor.commands.commands?.startAutocomplete === completionCommand) {
                editor.commands.removeCommand(completionCommand);
                editor.commands.addCommand(previousCommand);
            }
        };
    }, []);

    useEffect(() => {
        // 热更新会重跑 Effect，但已有 ACE 实例不会再次触发 onLoad。
        if (editorRef.current) attach(editorRef.current);
        return () => { cleanup.current?.(); };
    }, [attach]);
    return attach;
}
