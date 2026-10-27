let compiler: Promise<typeof import("@babel/standalone")> | undefined;
const compiled = new Map<string, {source: string; code: string; imports: string[]; size: number}>();
let cacheSize = 0;

// 只缓存转换结果，不保留组件状态、运行实例或插件闭包。
async function compile(path: string, source: string, scope: string) {
    const key = `${scope}/${path}`;
    const cached = compiled.get(key);
    if (cached?.source === source) {
        compiled.delete(key);
        compiled.set(key, cached);
        return cached;
    }
    compiler ??= import("@babel/standalone").catch((error) => {
        compiler = undefined;
        throw error;
    });
    const babel = await compiler;
    const result = babel.transform(source, {
        filename: key,
        sourceFileName: key,
        sourceMaps: "inline",
        ast: true,
        presets: [
            ["react", {runtime: "automatic"}],
            ...(/\.tsx?$/.test(path) ? [["typescript", {onlyRemoveTypeImports: true}]] : []),
        ],
        plugins: [
            ({types}: any) => ({visitor: {
                ImportExpression(node: any) {
                    node.replaceWith(types.callExpression(types.identifier("__import"), [node.node.source]));
                },
                CallExpression(node: any) {
                    if (node.node.callee.type === "Import") node.node.callee = types.identifier("__import");
                },
            }}),
            "transform-modules-commonjs",
        ],
    });
    const imports = new Set<string>();
    const visit = (node: any) => {
        if (!node || typeof node !== "object") return;
        if (node.type === "CallExpression" && node.callee?.type === "Identifier" && node.callee.name === "require"
            && node.arguments[0]?.type === "StringLiteral") imports.add(node.arguments[0].value);
        for (const value of Object.values(node)) {
            if (Array.isArray(value)) value.forEach(visit);
            else if (value && typeof value === "object") visit(value);
        }
    };
    visit(result.ast);
    const code = `${result.code || ""}\n//# sourceURL=software/${key}`;
    const output = {source, code, imports: [...imports], size: (source.length + code.length) * 2};
    const previous = compiled.get(key);
    if (previous) {
        cacheSize -= previous.size;
        compiled.delete(key);
    }
    // 跨弹窗复用转换缓存，限制常驻内存，源码更新时自然替换旧结果。
    if (output.size <= 16 * 1024 * 1024) {
        compiled.set(key, output);
        cacheSize += output.size;
    }
    while (cacheSize > 16 * 1024 * 1024 || compiled.size > 128) {
        const oldest = compiled.keys().next().value!;
        cacheSize -= compiled.get(oldest)!.size;
        compiled.delete(oldest);
    }
    return output;
}

// 浏览器转换源码并链接模块；公共依赖复用面板实例，动态 import 在使用时才加载。
export default function createSoftwareModules(entry: string, read: (path: string) => Promise<{path: string; source: string}>, shared: Record<string, any>, scope = "") {
    const files = new Map<string, Promise<any>>();
    const resolutions = new Map<string, Promise<any>>();
    const native = new Map<string, any>();
    let disposed = false;
    let loaded: Promise<any> | undefined;
    const checkOpen = () => {
        if (disposed) throw new Error("软件页面已关闭");
    };
    const dispose = () => {
        disposed = true;
        files.clear();
        resolutions.clear();
        native.clear();
    };
    const resolve = (name: string, parent?: string): Promise<any> => {
        checkOpen();
        if (typeof name !== "string" || !name) throw new Error("模块路径不能为空");
        if (parent && !name.startsWith("./") && !name.startsWith("../")) {
            if (!Object.prototype.hasOwnProperty.call(shared, name)) throw new Error(`插件未提供依赖 ${name}，请使用面板公共库或插件内的相对路径`);
            if (!native.has(name)) native.set(name, {path: name, imports: [], module: {exports: shared[name]}, state: 2});
            return Promise.resolve(native.get(name));
        }
        if (/[\\%?#:\u0000-\u001f]/.test(name)) throw new Error(`模块路径不合法：${name}`);
        const parts = parent ? parent.split("/").slice(0, -1) : [];
        for (const part of name.split("/")) {
            if (part === ".") continue;
            if (part === "..") {
                if (parts.length <= 1) throw new Error(`模块路径不能超出 web 目录：${name}`);
                parts.pop();
            } else {
                if (!part) throw new Error(`模块路径不合法：${name}`);
                parts.push(part);
            }
        }
        const path = parts.join("/");
        if (parts[0] !== "web" || parts.length < 2) throw new Error(`模块必须位于 web 目录：${path}`);
        const existing = resolutions.get(path) || files.get(path);
        if (existing) return existing;
        const promise = (async () => {
            // 由服务端一次解析后缀和目录入口，避免浏览器逐个请求不存在的文件。
            const page = await read(path);
            const actual = page?.path, source = page?.source;
            checkOpen();
            if (typeof actual !== "string" || !actual.startsWith("web/") || !/\.(tsx?|jsx?|mjs|json)$/.test(actual)) throw new Error(`不支持的模块路径：${actual}`);
            if (typeof source !== "string") throw new Error(`软件模块内容不合法：${actual}`);
            let file = files.get(actual);
            if (!file) {
                file = (async () => {
                    if (actual.endsWith(".json")) return {
                        path: actual, imports: [], module: {exports: JSON.parse(source)}, state: 2,
                    };
                    const output = await compile(actual, source, scope);
                    checkOpen();
                    return {path: actual, code: output.code, imports: output.imports, dependencies: new Map(), module: {exports: {}}, state: 0};
                })().catch((error) => {
                    files.delete(actual);
                    throw error;
                });
                files.set(actual, file);
            }
            return file;
        })().catch((error) => {
            resolutions.delete(path);
            throw error;
        });
        resolutions.set(path, promise);
        return promise;
    };
    const evaluate = (record: any): any => {
        checkOpen();
        if (record.state) return record.module.exports;
        record.state = 1;
        try {
            const execute = new Function("module", "exports", "require", "__import", record.code);
            execute(record.module, record.module.exports, (name: string) => {
                const dependency = record.dependencies.get(name);
                if (!dependency) throw new Error(`无法同步加载 ${name}，请使用 import()`);
                return evaluate(dependency);
            }, async (name: string) => {
                const exports = await load(name, record.path);
                return exports?.__esModule ? exports : {...exports, default: exports};
            });
            record.state = 2;
            return record.module.exports;
        } catch (error: any) {
            record.state = 0;
            record.module.exports = {};
            throw new Error(`${record.path}：${error?.message || "模块执行失败"}`);
        }
    };
    const load = async (name: string, parent?: string) => {
        const root = await resolve(name, parent);
        const visited = new Set<string>();
        let pending = [root];
        // 分层获取静态依赖，支持循环引用和并行读取，避免递归等待相互阻塞。
        while (pending.length) {
            const next: any[] = [];
            await Promise.all(pending.map(async (record) => {
                if (visited.has(record.path)) return;
                visited.add(record.path);
                const dependencies = await Promise.all(record.imports.map(async (value: string) => [value, await resolve(value, record.path)]));
                for (const [value, dependency] of dependencies) {
                    record.dependencies.set(value, dependency);
                    next.push(dependency);
                }
            }));
            pending = next;
        }
        return evaluate(root);
    };
    return {
        load: async () => {
            checkOpen();
            return loaded ??= load(entry).catch((error) => { dispose(); throw error; });
        },
        dispose,
    };
}
