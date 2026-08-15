import React, {
    forwardRef,
    memo,
    useCallback,
    useImperativeHandle,
    useMemo,
    useRef,
    useState,
} from "react";
import {App, Spin} from "antd";
import {createStyles} from "antd-style";
import {ProModal, Title, useTheme} from "sinking-antd";
import defaultSettings from "@/../config/defaultSettings";
import AceEditor from "@/components/ace-editor";
import {getFileInfo, updateFile} from "@/service/api/file";

const acePath = `${defaultSettings?.basePath || "/"}ace`;
const pageSize = 100000;
const maxContentSize = 8 * 1024 * 1024;

const useStyles = createStyles(({css, token}) => ({
    modal: css`
        .ant-modal {
            max-width: calc(100vw - 24px);
            outline: none;
        }

        .ant-modal-body {
            max-height: calc(100dvh - 168px);
            overflow-y: auto;
            overscroll-behavior: contain;
        }

        @supports not (height: 100dvh) {
            .ant-modal-body {
                max-height: calc(100vh - 168px);
            }
        }
    `,
    fileMeta: css`
        min-width: 0;
        margin-bottom: 12px;
        padding: 8px 10px;
        overflow: hidden;
        border-radius: ${token.borderRadiusSM}px;
        background: ${token.colorFillQuaternary};

        .file-name,
        .file-path {
            display: block;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .file-name {
            color: ${token.colorTextHeading};
            font-size: ${token.fontSize}px;
            font-weight: 500;
            line-height: 20px;
        }

        .file-path {
            color: ${token.colorTextTertiary};
            font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
            font-size: ${token.fontSizeSM}px;
            line-height: 18px;
        }
    `,
    editor: css`
        overflow: hidden;
        border: 1px solid ${token.colorBorderSecondary};
        border-radius: ${token.borderRadius}px;

        .ace_editor {
            border-radius: ${token.borderRadius}px !important;
        }
    `,
    loading: css`
        height: min(420px, 52dvh);
        min-height: 240px;
        display: flex;
        align-items: center;
        justify-content: center;
    `,
}));

interface EditorState {
    generation: number;
    path: string;
    name: string;
}

interface FileEditorProps {
    onSuccess: () => void;
}

export interface FileEditorRef {
    open: (path: string, name: string) => void;
}

const extensionModes: Record<string, string> = {
    bash: "sh",
    bat: "batchfile",
    c: "c_cpp",
    cc: "c_cpp",
    cfg: "ini",
    conf: "ini",
    cpp: "c_cpp",
    cs: "csharp",
    css: "css",
    cts: "typescript",
    cxx: "c_cpp",
    dart: "dart",
    go: "golang",
    graphql: "graphqlschema",
    groovy: "groovy",
    h: "c_cpp",
    hpp: "c_cpp",
    htm: "html",
    html: "html",
    ini: "ini",
    java: "java",
    js: "javascript",
    json: "json",
    jsonc: "json",
    jsx: "jsx",
    kt: "kotlin",
    kts: "kotlin",
    less: "less",
    lua: "lua",
    m: "objectivec",
    markdown: "markdown",
    md: "markdown",
    mjs: "javascript",
    mts: "typescript",
    nginx: "nginx",
    php: "php",
    pl: "perl",
    properties: "properties",
    ps1: "powershell",
    py: "python",
    r: "r",
    rb: "ruby",
    rs: "rust",
    sass: "sass",
    scala: "scala",
    scss: "scss",
    sh: "sh",
    sql: "sql",
    svg: "svg",
    swift: "swift",
    toml: "toml",
    ts: "typescript",
    tsx: "tsx",
    vue: "vue",
    xml: "xml",
    yaml: "yaml",
    yml: "yaml",
    zsh: "sh",
};

const getEditorMode = (name: string) => {
    const normalized = name.toLowerCase();
    if (normalized === "dockerfile" || normalized.startsWith("dockerfile.")) {
        return "dockerfile";
    }
    if (normalized === "makefile" || normalized.startsWith("makefile.")) {
        return "makefile";
    }
    if (normalized === ".gitignore") {
        return "gitignore";
    }
    const extension = normalized.includes(".") ? normalized.split(".").pop() || "" : "";
    return extensionModes[extension] || "text";
};

const contentByteSize = (value: string) => new TextEncoder().encode(value).byteLength;

const FileEditor = forwardRef<FileEditorRef, FileEditorProps>(({onSuccess}, ref) => {
    const {message} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const dark = Boolean(theme?.isDarkMode?.() || theme?.isDarkTheme?.());
    const {styles} = useStyles();
    const detailGenerationRef = useRef(0);
    const submitGenerationRef = useRef(0);
    const submittingRef = useRef(false);
    const editorRef = useRef<EditorState | undefined>(undefined);
    const [editor, setEditor] = useState<EditorState>();
    const [content, setContent] = useState("");
    const [version, setVersion] = useState("");
    const [loading, setLoading] = useState(false);
    const [submitting, setSubmitting] = useState(false);
    editorRef.current = editor;

    const close = useCallback(() => {
        detailGenerationRef.current += 1;
        submitGenerationRef.current += 1;
        submittingRef.current = false;
        editorRef.current = undefined;
        setEditor(undefined);
        setContent("");
        setVersion("");
        setLoading(false);
        setSubmitting(false);
    }, []);

    const load = useCallback(async (generation: number, path: string) => {
        let cursor = 0;
        let fileVersion = "";
        let value = "";
        let loadedBytes = 0;

        try {
            while (true) {
                const response = await getFileInfo({
                    body: {
                        path,
                        read: true,
                        cursor,
                        page_size: pageSize,
                        version: fileVersion,
                    },
                });
                if (detailGenerationRef.current !== generation) {
                    return;
                }
                if (response?.code !== 200 || !response.data) {
                    throw new Error(response?.message || "读取文件内容失败");
                }

                const data = response.data;
                if (data.is_dir) {
                    throw new Error("目录不支持内容编辑");
                }
                if (Number(data.size) > maxContentSize) {
                    throw new Error("文件超过 8 MiB，无法在线编辑");
                }

                const responseVersion = String(data.version || "");
                if (!fileVersion) {
                    if (!responseVersion) {
                        throw new Error("无法获取文件版本，请重新打开");
                    }
                    fileVersion = responseVersion;
                } else if (responseVersion !== fileVersion) {
                    throw new Error("文件内容已发生变化，请重新打开");
                }

                const nextCursor = Number(data.next_cursor);
                const chunk = String(data.content || "");
                if (!Number.isSafeInteger(nextCursor) || nextCursor < cursor) {
                    throw new Error("文件游标异常，请重新打开");
                }
                loadedBytes += contentByteSize(chunk);
                if (nextCursor > maxContentSize || loadedBytes > maxContentSize) {
                    throw new Error("文件超过 8 MiB，无法在线编辑");
                }
                value += chunk;

                if (data.eof) {
                    break;
                }
                if (nextCursor === cursor) {
                    throw new Error("文件读取未能继续，请重新打开");
                }
                cursor = nextCursor;
            }

            if (detailGenerationRef.current !== generation) {
                return;
            }
            setContent(value);
            setVersion(fileVersion);
        } catch (error: unknown) {
            if (detailGenerationRef.current !== generation) {
                return;
            }
            message.error(error instanceof Error ? error.message : "读取文件内容失败");
            close();
        } finally {
            if (detailGenerationRef.current === generation) {
                setLoading(false);
            }
        }
    }, [close, message]);

    const open = useCallback((path: string, name: string) => {
        const generation = ++detailGenerationRef.current;
        submitGenerationRef.current += 1;
        submittingRef.current = false;
        const nextEditor = {
            generation,
            path,
            name: name || path.split(/[\\/]/).pop() || path,
        };
        editorRef.current = nextEditor;
        setEditor(nextEditor);
        setContent("");
        setVersion("");
        setLoading(true);
        setSubmitting(false);
        void load(generation, path);
    }, [load]);

    useImperativeHandle(ref, () => ({open}), [open]);

    const submit = useCallback(async () => {
        if (!editor || loading || submittingRef.current) {
            return;
        }
        if (!version) {
            message.error("文件内容尚未加载完成");
            return;
        }
        if (contentByteSize(content) > maxContentSize) {
            message.error("文件超过 8 MiB，无法在线保存");
            return;
        }

        const submitGeneration = ++submitGenerationRef.current;
        const editorGeneration = editor.generation;
        submittingRef.current = true;
        setSubmitting(true);
        try {
            const response = await updateFile({body: {
                path: editor.path,
                content,
                version,
            }});
            if (response?.code !== 200) {
                message.error(response?.message || "保存文件失败");
                return;
            }

            message.success(response.message || "保存成功");
            onSuccess();
            if (
                submitGenerationRef.current === submitGeneration &&
                editorRef.current?.generation === editorGeneration
            ) {
                close();
            }
        } finally {
            if (submitGenerationRef.current === submitGeneration) {
                submittingRef.current = false;
                setSubmitting(false);
            }
        }
    }, [close, content, editor, loading, message, onSuccess, version]);

    const mode = useMemo(() => getEditorMode(editor?.name || ""), [editor?.name]);

    return (
        <ProModal
            title={<Title>编辑文件</Title>}
            width="900px"
            okText="保存"
            onOk={submit}
            onCancel={close}
            modalProps={{
                rootClassName: styles.modal,
                open: Boolean(editor),
                cancelText: "取消",
                confirmLoading: submitting,
                okButtonProps: {disabled: loading || !version},
                focusable: {focusTriggerAfterClose: false},
                mask: {closable: true},
                forceRender: true,
                style: {top: "clamp(12px, 5vh, 56px)", paddingBottom: 24},
            }}>
            {editor && (loading ? (
                <div className={styles.loading}>
                    <Spin description="加载文件内容..."/>
                </div>
            ) : (
                <>
                    <div className={styles.fileMeta}>
                        <span className="file-name" title={editor.name}>{editor.name}</span>
                        <span className="file-path" title={editor.path}>{editor.path}</span>
                    </div>
                    <AceEditor
                        value={content}
                        mode={mode}
                        theme={dark ? "monokai" : "github"}
                        width="100%"
                        height={compact
                            ? "clamp(240px, 52dvh, 500px)"
                            : "clamp(280px, 56dvh, 560px)"}
                        fontSize={12}
                        showPrintMargin={false}
                        wrapEnabled
                        acePath={acePath}
                        onChange={setContent}
                        onLoad={(instance) => instance.textInput?.getElement?.()
                            ?.setAttribute("aria-label", "文件内容")}
                        className={styles.editor}/>
                </>
            ))}
        </ProModal>
    );
});

FileEditor.displayName = "FileEditor";

export default memo(FileEditor);
