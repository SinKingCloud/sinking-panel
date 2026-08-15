// noinspection JSIncompatibleTypesComparison

import React, {useRef, useEffect, useState, useCallback, useMemo, useLayoutEffect} from 'react';
import Script, {preloadScript, isScriptLoaded} from '@/components/script';

// Ace Editor 配置接口
export interface AceEditorProps {
    value?: string;
    defaultValue?: string;
    mode?: string;
    theme?: string;
    width?: string | number;
    height?: string | number;
    fontSize?: number;
    tabSize?: number;
    readOnly?: boolean;
    showPrintMargin?: boolean;
    showGutter?: boolean;
    highlightActiveLine?: boolean;
    highlightSelectedWord?: boolean;
    wrapEnabled?: boolean;
    autoScrollEditorIntoView?: boolean;
    maxLines?: number;
    minLines?: number;
    placeholder?: string;
    className?: string;
    style?: React.CSSProperties;
    onChange?: (value: string, event?: any) => void;
    onSelectionChange?: (selection: any, event?: any) => void;
    onCursorChange?: (selection: any, event?: any) => void;
    onBlur?: (event?: any) => void;
    onFocus?: (event?: any) => void;
    onLoad?: (editor: any) => void;
    onBeforeLoad?: (ace: any) => void;
    commands?: Array<{
        name: string;
        bindKey: { win: string; mac: string };
        exec: (editor: any) => void;
    }>;
    annotations?: Array<{
        row: number;
        column: number;
        text: string;
        type: 'error' | 'warning' | 'info';
    }>;
    markers?: Array<{
        startRow: number;
        startCol: number;
        endRow: number;
        endCol: number;
        className: string;
        type: string;
    }>;
    enableBasicAutocompletion?: boolean;
    enableLiveAutocompletion?: boolean;
    enableSnippets?: boolean;
    showLineNumbers?: boolean;
    acePath?: string; // 自定义 Ace 资源路径
    loadingContent?: React.ReactNode; // 仅首次加载时的自定义加载内容
    containerStyle?: React.CSSProperties; // 内部稳定容器样式
    onError?: (error: Error) => void; // 错误处理回调
}

const EMPTY_COMMANDS: NonNullable<AceEditorProps['commands']> = [];
const EMPTY_ANNOTATIONS: NonNullable<AceEditorProps['annotations']> = [];
const EMPTY_MARKERS: NonNullable<AceEditorProps['markers']> = [];


const AceEditor: React.FC<AceEditorProps> = ({
                                                 value,
                                                 defaultValue = '',
                                                 mode = 'text',
                                                 theme = 'monokai',
                                                 width = '100%',
                                                 height = 400,
                                                 fontSize = 14,
                                                 tabSize = 4,
                                                 readOnly = false,
                                                 showPrintMargin = true,
                                                 showGutter = true,
                                                 highlightActiveLine = true,
                                                 highlightSelectedWord = true,
                                                 wrapEnabled = false,
                                                 autoScrollEditorIntoView = false,
                                                 maxLines,
                                                 minLines,
                                                 placeholder = '',
                                                 className = '',
                                                 style = {},
                                                 containerStyle,
                                                 loadingContent,
                                                 onChange,
                                                 onSelectionChange,
                                                 onCursorChange,
                                                 onBlur,
                                                 onFocus,
                                                 onLoad,
                                                 onBeforeLoad,
                                                 commands = EMPTY_COMMANDS,
                                                 annotations = EMPTY_ANNOTATIONS,
                                                 markers = EMPTY_MARKERS,
                                                 enableBasicAutocompletion = true,
                                                 enableLiveAutocompletion = true,
                                                 enableSnippets = true,
                                                 showLineNumbers = true,
                                                 acePath = '/ace',
                                                 onError
                                             }) => {
    const containerRef = useRef<HTMLDivElement | null>(null);
    const editorRef = useRef<any>(null);
    const [aceLoaded, setAceLoaded] = useState(false);
    const [hasScriptLoaded, setHasScriptLoaded] = useState(false); // 仅首次加载展示 loading
    const editorValueRef = useRef(value ?? defaultValue);

    // 创建一个稳定的容器元素，避免 React 重新创建
    const stableContainer = useRef<HTMLDivElement | null>(null);
    const mountedRef = useRef(false);
    const [containerMounted, setContainerMounted] = useState(false);
    const outerContainerRef = useRef<any>(null);
    const activeRef = useRef(true);
    const commandNamesRef = useRef<string[]>([]);
    const markerIdsRef = useRef<number[]>([]);
    const modeRequestRef = useRef(0);
    const themeRequestRef = useRef(0);
    const callbacksRef = useRef({
        onChange,
        onSelectionChange,
        onCursorChange,
        onBlur,
        onFocus,
        onLoad,
        onBeforeLoad,
        onError,
    });

    useLayoutEffect(() => {
        callbacksRef.current = {
            onChange,
            onSelectionChange,
            onCursorChange,
            onBlur,
            onFocus,
            onLoad,
            onBeforeLoad,
            onError,
        };
    }, [
        onBeforeLoad,
        onBlur,
        onChange,
        onCursorChange,
        onError,
        onFocus,
        onLoad,
        onSelectionChange,
    ]);

    // 统一应用编辑器可变配置
    const applyEditorOptions = useCallback((ed: any, options: {
        fontSize: number;
        tabSize: number;
        readOnly: boolean;
        showPrintMargin: boolean;
        showGutter: boolean;
        showLineNumbers: boolean;
        highlightActiveLine: boolean;
        highlightSelectedWord: boolean;
        wrapEnabled: boolean;
        autoScrollEditorIntoView: boolean;
        maxLines?: number;
        minLines?: number;
        placeholder?: string;
    }) => {
        if (!ed) return;
        const {
            fontSize,
            tabSize,
            readOnly,
            showPrintMargin,
            showGutter,
            showLineNumbers,
            highlightActiveLine,
            highlightSelectedWord,
            wrapEnabled,
            autoScrollEditorIntoView,
            maxLines,
            minLines,
            placeholder
        } = options;

        ed.setFontSize(fontSize);
        ed.session.setTabSize(tabSize);
        ed.setReadOnly(readOnly);
        ed.setShowPrintMargin(showPrintMargin);
        ed.renderer.setShowGutter(showGutter && showLineNumbers);
        ed.setHighlightActiveLine(highlightActiveLine);
        ed.setHighlightSelectedWord(highlightSelectedWord);
        ed.session.setUseWrapMode(wrapEnabled);
        ed.setAutoScrollEditorIntoView(autoScrollEditorIntoView);
        if (maxLines) ed.setOption('maxLines', maxLines);
        if (minLines) ed.setOption('minLines', minLines);
        if (placeholder) ed.setOption('placeholder', placeholder);
    }, []);

    // 创建稳定容器（只执行一次）
    useLayoutEffect(() => {
        // 创建稳定的容器元素，只创建一次
        if (!stableContainer.current) {
            stableContainer.current = document.createElement('div');
            stableContainer.current.style.overflow = 'hidden';
        }

        return () => {
            // 不主动移除稳定容器，返回页面时可复用，避免丢失
        };
    }, []);

    // 挂载稳定容器到 React 容器中
    useLayoutEffect(() => {
        if (containerRef.current && stableContainer.current && !mountedRef.current) {
            // 设置初始尺寸
            const initialWidth = typeof width === 'number' ? `${width}px` : width;
            const initialHeight = typeof height === 'number' ? `${height}px` : height;

            // 设置外层容器初始尺寸
            if (outerContainerRef.current) {
                outerContainerRef.current.style.width = initialWidth;
                outerContainerRef.current.style.height = initialHeight;
            }

            // 设置稳定容器初始尺寸
            stableContainer.current.style.width = initialWidth;
            stableContainer.current.style.height = initialHeight;

            containerRef.current.appendChild(stableContainer.current);
            mountedRef.current = true;
            setContainerMounted(true);
        } else if (containerRef.current && stableContainer.current && !stableContainer.current.parentNode) {
            // 跨页面返回时，如果稳定容器没有挂载，重新挂载
            containerRef.current.appendChild(stableContainer.current);
            mountedRef.current = true;
            setContainerMounted(true);
        }
    });

    // 只加载核心脚本，扩展按需加载
    const coreScript = useMemo(() => `${acePath}/src-min/ace.js`, [acePath]);
    const isAcePresent = useMemo(() => typeof window !== 'undefined' && (window as any).ace, []);

    // 如果 ace 已经在全局存在（例如返回页面后），直接标记已加载
    useEffect(() => {
        if (isAcePresent && !aceLoaded) setAceLoaded(true);
    }, [isAcePresent, aceLoaded]);

    // React 容器样式 - 使用固定样式，通过稳定容器控制实际尺寸
    const containerBaseStyle = useMemo(() => ({
        width: '100%',
        height: '100%',
        overflow: 'hidden',
        position: 'relative' as const
    }), []);

    // 应用用户传入的 containerStyle 到稳定容器
    useLayoutEffect(() => {
        if (!stableContainer.current) return;
        const el = stableContainer.current as HTMLDivElement;
        el.style.overflow = 'hidden';
        if (containerStyle) {
            try {
                Object.assign(el.style, containerStyle);
            } catch {
            }
        }
    }, [containerStyle]);

    // 动态加载模式文件
    const loadMode = useCallback(async (modeName: string) => {
        if (!modeName || modeName === 'text') return;

        const scriptUrl = `${acePath}/src-min/mode-${modeName}.js`;

        if (isScriptLoaded(scriptUrl)) return;

        await preloadScript(scriptUrl, {
            cache: true,
            timeout: 5000,
            retryCount: 2
        });
    }, [acePath]);

    // 动态加载主题文件
    const loadTheme = useCallback(async (themeName: string) => {
        if (!themeName) return;

        const scriptUrl = `${acePath}/src-min/theme-${themeName}.js`;

        if (isScriptLoaded(scriptUrl)) return;

        await preloadScript(scriptUrl, {
            cache: true,
            timeout: 5000,
            retryCount: 2
        });
    }, [acePath]);


    // 当 Ace 加载完成后初始化编辑器
    useEffect(() => {
        if (!aceLoaded || editorRef.current) return;
        let cancelled = false;

        const initialize = async () => {
            if (!stableContainer.current || !window.ace || !mountedRef.current) return;

            // 预先加载初始的模式和主题
            const modeUrl = `${acePath}/src-min/mode-${mode}.js`;
            const themeUrl = `${acePath}/src-min/theme-${theme}.js`;

            try {
                // 并行加载所有需要的资源
                const loadTasks = [];

                // 加载模式文件
                if (mode !== 'text' && !isScriptLoaded(modeUrl)) {
                    loadTasks.push(preloadScript(modeUrl, {cache: true}));
                }

                // 加载主题文件
                if (!isScriptLoaded(themeUrl)) {
                    loadTasks.push(preloadScript(themeUrl, {cache: true}));
                }

                // 加载扩展文件
                if (enableBasicAutocompletion || enableLiveAutocompletion) {
                    const langToolsUrl = `${acePath}/src-min/ext-language_tools.js`;
                    if (!isScriptLoaded(langToolsUrl)) {
                        loadTasks.push(preloadScript(langToolsUrl, {cache: true}));
                    }
                }

                if (enableSnippets) {
                    const searchboxUrl = `${acePath}/src-min/ext-searchbox.js`;
                    if (!isScriptLoaded(searchboxUrl)) {
                        loadTasks.push(preloadScript(searchboxUrl, {cache: true}));
                    }
                }

                if (loadTasks.length > 0) {
                    await Promise.all(loadTasks);
                }

                if (cancelled || !activeRef.current || !mountedRef.current) {
                    return;
                }

                // 直接初始化编辑器，使用稳定容器
                if (!editorRef.current && stableContainer.current && window.ace) {
                    callbacksRef.current.onBeforeLoad?.(window.ace);

                    const editor = window.ace.edit(stableContainer.current);
                    editorRef.current = editor;

                    // 基础配置
                    editor.setTheme(`ace/theme/${theme}`);
                    editor.session.setMode(`ace/mode/${mode}`);
                    editor.setValue(editorValueRef.current, -1);
                    applyEditorOptions(editor, {
                        fontSize,
                        tabSize,
                        readOnly,
                        showPrintMargin,
                        showGutter,
                        showLineNumbers,
                        highlightActiveLine,
                        highlightSelectedWord,
                        wrapEnabled,
                        autoScrollEditorIntoView,
                        maxLines,
                        minLines,
                        placeholder
                    });

                    // 自动完成配置
                    if (enableBasicAutocompletion || enableLiveAutocompletion || enableSnippets) {
                        window.ace.require('ace/ext/language_tools');
                        editor.setOptions({
                            enableBasicAutocompletion,
                            enableLiveAutocompletion,
                            enableSnippets
                        });
                    }

                    // 事件监听
                    editor.on('change', (event: any) => {
                        const newValue = editor.getValue();
                        editorValueRef.current = newValue;
                        callbacksRef.current.onChange?.(newValue, event);
                    });

                    editor.on('changeSelection', (event: any) => {
                        callbacksRef.current.onSelectionChange?.(editor.getSelection(), event);
                    });
                    editor.on('changeCursor', (event: any) => {
                        callbacksRef.current.onCursorChange?.(editor.getSelection(), event);
                    });
                    editor.on('blur', (event: any) => callbacksRef.current.onBlur?.(event));
                    editor.on('focus', (event: any) => callbacksRef.current.onFocus?.(event));

                    // 添加自定义命令
                    commands.forEach(command => {
                        editor.commands.addCommand(command);
                        commandNamesRef.current.push(command.name);
                    });

                    // 设置注释和标记
                    if (annotations.length > 0) {
                        editor.session.setAnnotations(annotations);
                    }

                    if (markers.length > 0) {
                        const Range = window.ace.require('ace/range').Range as any;
                        markers.forEach(marker => {
                            const range = new Range(marker.startRow, marker.startCol, marker.endRow, marker.endCol);
                            editor.session.addMarker(range, marker.className, marker.type);
                        });
                    }
                    if (!cancelled && activeRef.current) {
                        callbacksRef.current.onLoad?.(editor);
                    }
                }

            } catch (error) {
                if (!cancelled && activeRef.current) {
                    callbacksRef.current.onError?.(error as Error);
                }
            }
        };

        void initialize();
        return () => {
            cancelled = true;
        };
    }, [
        aceLoaded,
        acePath,
        annotations,
        applyEditorOptions,
        autoScrollEditorIntoView,
        commands,
        containerMounted,
        enableBasicAutocompletion,
        enableLiveAutocompletion,
        enableSnippets,
        fontSize,
        highlightActiveLine,
        highlightSelectedWord,
        markers,
        maxLines,
        minLines,
        mode,
        placeholder,
        readOnly,
        showGutter,
        showLineNumbers,
        showPrintMargin,
        tabSize,
        theme,
        wrapEnabled,
    ]);

    // 更新编辑器配置
    useEffect(() => {
        if (value !== undefined) {
            editorValueRef.current = value;
        }
        if (!editorRef.current) return;

        const editor = editorRef.current;

        // 更新值（避免无限循环）
        if (value !== undefined && value !== editor.getValue()) {
            const cursorPosition = editor.getCursorPosition();
            editor.setValue(value, -1);
            editor.moveCursorToPosition(cursorPosition);
            editorValueRef.current = value;
        }

        // 更新其他配置
        applyEditorOptions(editor, {
            fontSize,
            tabSize,
            readOnly,
            showPrintMargin,
            showGutter,
            showLineNumbers,
            highlightActiveLine,
            highlightSelectedWord,
            wrapEnabled,
            autoScrollEditorIntoView,
            maxLines,
            minLines,
            placeholder
        });

    }, [
        value, fontSize, tabSize, readOnly, showPrintMargin, showGutter,
        highlightActiveLine, highlightSelectedWord, wrapEnabled,
        autoScrollEditorIntoView, maxLines, minLines, placeholder, showLineNumbers
    ]);

    // 同步命令（commands）
    useEffect(() => {
        if (!editorRef.current) return;
        const editor = editorRef.current;

        // 移除旧命令
        if (commandNamesRef.current.length) {
            commandNamesRef.current.forEach((name) => {
                try {
                    editor.commands.removeCommand(name);
                } catch {
                }
            });
            commandNamesRef.current = [];
        }

        // 添加新命令
        if (Array.isArray(commands) && commands.length > 0) {
            commands.forEach((cmd) => {
                try {
                    editor.commands.addCommand(cmd);
                    commandNamesRef.current.push(cmd.name);
                } catch {
                }
            });
        }
    }, [commands]);

    // 同步注解（annotations）
    useEffect(() => {
        if (!editorRef.current) return;
        const editor = editorRef.current;
        try {
            editor.session.setAnnotations(Array.isArray(annotations) ? annotations : []);
        } catch {
        }
    }, [annotations]);

    // 同步标记（markers）
    useEffect(() => {
        if (!editorRef.current) return;
        const session = editorRef.current.session;

        // 清理旧的标记
        if (markerIdsRef.current.length) {
            markerIdsRef.current.forEach((id) => {
                try {
                    session.removeMarker(id);
                } catch {
                }
            });
            markerIdsRef.current = [];
        }

        if (Array.isArray(markers) && markers.length > 0) {
            try {
                const aceAny = (window as any).ace;
                const RangeCtor = aceAny?.require?.('ace/range')?.Range as any;
                if (RangeCtor) {
                    markers.forEach((m) => {
                        const range = new RangeCtor(m.startRow, m.startCol, m.endRow, m.endCol);
                        const id = session.addMarker(range, m.className, m.type);
                        markerIdsRef.current.push(id);
                    });
                }
            } catch {
            }
        }
    }, [markers]);

    // 处理尺寸变化，直接更新外层和稳定容器的样式
    useLayoutEffect(() => {
        const newWidth = typeof width === 'number' ? `${width}px` : width;
        const newHeight = typeof height === 'number' ? `${height}px` : height;

        // 更新外层容器尺寸
        if (outerContainerRef.current) {
            outerContainerRef.current.style.width = newWidth;
            outerContainerRef.current.style.height = newHeight;
        }

        // 更新稳定容器尺寸
        if (stableContainer.current) {
            stableContainer.current.style.width = newWidth;
            stableContainer.current.style.height = newHeight;
        }

        // 通知编辑器重新计算大小
        if (editorRef.current) {
            editorRef.current.resize();
        }
    }, [width, height]);

    // 容器可能因弹窗、侧栏或响应式布局变化，需按实际尺寸通知 Ace 重排
    useLayoutEffect(() => {
        const container = outerContainerRef.current;
        if (!container || typeof ResizeObserver === 'undefined') return;

        let resizeFrame: number | undefined;
        const observer = new ResizeObserver(() => {
            if (resizeFrame !== undefined) {
                cancelAnimationFrame(resizeFrame);
            }
            resizeFrame = requestAnimationFrame(() => {
                resizeFrame = undefined;
                editorRef.current?.resize();
            });
        });
        observer.observe(container);

        return () => {
            observer.disconnect();
            if (resizeFrame !== undefined) {
                cancelAnimationFrame(resizeFrame);
            }
        };
    }, []);

    // 处理模式变化
    useEffect(() => {
        const editor = editorRef.current;
        if (!editor || !mode) return;
        const requestId = ++modeRequestRef.current;
        let cancelled = false;

        const updateMode = async () => {
            try {
                await loadMode(mode);
                if (
                    !cancelled &&
                    activeRef.current &&
                    modeRequestRef.current === requestId &&
                    editorRef.current === editor
                ) {
                    editor.session.setMode(`ace/mode/${mode}`);
                }
            } catch (error) {
                if (
                    !cancelled &&
                    activeRef.current &&
                    modeRequestRef.current === requestId &&
                    editorRef.current === editor
                ) {
                    callbacksRef.current.onError?.(error as Error);
                }
            }
        };
        void updateMode();
        return () => {
            cancelled = true;
            if (modeRequestRef.current === requestId) {
                modeRequestRef.current += 1;
            }
        };
    }, [loadMode, mode]);

    // 处理主题变化
    useEffect(() => {
        const editor = editorRef.current;
        if (!editor || !theme) return;
        const requestId = ++themeRequestRef.current;
        let cancelled = false;

        const updateTheme = async () => {
            try {
                await loadTheme(theme);
                if (
                    !cancelled &&
                    activeRef.current &&
                    themeRequestRef.current === requestId &&
                    editorRef.current === editor
                ) {
                    editor.setTheme(`ace/theme/${theme}`);
                }
            } catch (error) {
                if (
                    !cancelled &&
                    activeRef.current &&
                    themeRequestRef.current === requestId &&
                    editorRef.current === editor
                ) {
                    callbacksRef.current.onError?.(error as Error);
                }
            }
        };
        void updateTheme();
        return () => {
            cancelled = true;
            if (themeRequestRef.current === requestId) {
                themeRequestRef.current += 1;
            }
        };
    }, [loadTheme, theme]);

    // 组件卸载时清理
    useEffect(() => {
        activeRef.current = true;
        return () => {
            activeRef.current = false;
            mountedRef.current = false;
            if (editorRef.current) {
                editorRef.current.destroy();
                editorRef.current = null;
            }
        };
    }, []);

    const handleCoreLoad = useCallback(() => {
        if (!activeRef.current) {
            return;
        }
        setAceLoaded(true);
        setHasScriptLoaded(true);
    }, []);

    const handleCoreError = useCallback((error: Error) => {
        if (activeRef.current) {
            callbacksRef.current.onError?.(error);
        }
    }, []);

    const handleCoreTimeout = useCallback(() => {
        if (activeRef.current) {
            callbacksRef.current.onError?.(new Error(`Script load timeout: ${coreScript}`));
        }
    }, [coreScript]);

    return (
        <div className={className} style={style} ref={outerContainerRef}>
            {isAcePresent ? (
                <div ref={containerRef} style={containerBaseStyle as any}/>
            ) : (
                <Script
                    src={coreScript}
                    parallel={false}
                    timeout={10000}
                    retryCount={2}
                    cache={true}
                    onLoad={handleCoreLoad}
                    onError={handleCoreError}
                    onTimeout={handleCoreTimeout}
                    loading={hasScriptLoaded ? null : loadingContent}>
                    <div ref={containerRef} style={containerBaseStyle as any}/>
                </Script>
            )}
        </div>
    );
};

export default AceEditor;

// 声明全局 ace 对象
declare global {
    interface Window {
        ace: any;
    }
}
