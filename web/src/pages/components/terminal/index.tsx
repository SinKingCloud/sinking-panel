import {
    forwardRef,
    memo,
    useCallback,
    useEffect,
    useImperativeHandle,
    useMemo,
    useRef,
} from "react";
import type {CSSProperties} from "react";
import {theme as antTheme} from "antd";
import {createStyles} from "antd-style";
import {useTheme} from "sinking-antd";
import type {Terminal as XTermInstance} from "@xterm/xterm";
import {FitAddon} from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";

const XTerm = (require("@xterm/xterm/lib/xterm.js") as typeof import("@xterm/xterm")).Terminal;

export interface TerminalSize {
    cols: number;
    rows: number;
}

export interface TerminalRef {
    clear: () => void;
    reset: () => void;
    write: (data: string | Uint8Array) => void;
    focus: () => void;
    fit: (focus?: boolean) => void;
    getSize: () => TerminalSize;
    sendInput: (content: string) => boolean;
    paste: (content: string) => boolean;
}

export interface TerminalProps {
    socket?: WebSocket;
    active?: boolean;
    compact?: boolean;
    background?: string;
    accentColor?: string;
    className?: string;
    style?: CSSProperties;
}

const useStyles = createStyles(({css}: any, props: { compact: boolean; background: string }) => ({
    terminal: css`
        width: 100%;
        height: 100%;
        min-width: 0;
        min-height: 0;
        overflow: hidden;
        flex: 1;
        background: ${props.background};

        .xterm {
            height: 100%;
            padding: ${props.compact ? 10 : 12}px;
            box-sizing: border-box;
        }

        .xterm-viewport {
            background: ${props.background} !important;
            scrollbar-width: thin;
            scrollbar-color: rgba(255, 255, 255, .18) transparent;
        }

        .xterm-viewport::-webkit-scrollbar {
            width: 8px !important;
            height: 8px !important;
        }

        .xterm-viewport::-webkit-scrollbar-thumb {
            border-radius: 4px;
            background: rgba(255, 255, 255, .18);
        }
    `,
}));

const Terminal = forwardRef<TerminalRef, TerminalProps>(({
                                                             socket,
                                                             active = true,
                                                             compact = false,
                                                             background,
                                                             accentColor,
                                                             className,
                                                             style
                                                         }, ref): any => {
    const {token} = antTheme.useToken();
    const appTheme = useTheme();
    const dark = Boolean(appTheme?.isDarkMode?.() || appTheme?.isDarkTheme?.());
    const terminalBackground = background || (dark ? "rgb(15, 15, 15)" : "#000000");
    const terminalAccent = accentColor || (dark ? "#c3ccd6" : token.colorPrimary);
    const selectionBackground = accentColor
        ? "rgba(195, 204, 214, .28)"
        : dark
            ? "rgba(195, 204, 214, .28)"
            : token.colorPrimary;
    const styleProps = useMemo(() => ({compact, background: terminalBackground}), [compact, terminalBackground]);
    const {styles} = useStyles(styleProps);
    const hostRef = useRef<HTMLDivElement | any>(null);
    const terminalRef = useRef<XTermInstance | undefined>(undefined);
    const fitAddonRef = useRef<FitAddon | undefined>(undefined);
    const socketRef = useRef<WebSocket | undefined>(undefined);
    const socketEncoderRef = useRef(new TextEncoder());
    const fitFrameRef = useRef(0);
    const lastSizeRef = useRef("");
    socketRef.current = socket;

    const terminalTheme = useMemo(() => ({
        background: terminalBackground,
        foreground: "#dedede",
        cursor: terminalAccent,
        cursorAccent: terminalBackground,
        selectionBackground,
        selectionForeground: "#ffffff",
        black: "#111111",
        red: "#ff7b72",
        green: "#7ee787",
        yellow: "#e3b341",
        blue: "#79c0ff",
        magenta: "#d2a8ff",
        cyan: "#a5d6ff",
        white: "#ededed",
        brightBlack: "#707070",
        brightRed: "#ffa198",
        brightGreen: "#aff5b4",
        brightYellow: "#f2cc60",
        brightBlue: "#a5d6ff",
        brightMagenta: "#e2c5ff",
        brightCyan: "#b6e3ff",
        brightWhite: "#ffffff",
    }), [selectionBackground, terminalAccent, terminalBackground, token.colorPrimary]);
    const themeRef = useRef(terminalTheme);
    themeRef.current = terminalTheme;

    const sendResize = useCallback(() => {
        const currentSocket = socketRef.current;
        const terminal = terminalRef.current;
        if (!terminal || currentSocket?.readyState !== WebSocket.OPEN) {
            return;
        }
        const size = `${terminal.cols}|${terminal.rows}`;
        if (lastSizeRef.current === size) {
            return;
        }
        lastSizeRef.current = size;
        try {
            currentSocket.send(socketEncoderRef.current.encode(JSON.stringify({event: "resize", content: size})));
        } catch {
            lastSizeRef.current = "";
        }
    }, []);

    const fitNow = useCallback(() => {
        const host = hostRef.current;
        const addon = fitAddonRef.current;
        if (!host || !addon || host.clientWidth <= 0 || host.clientHeight <= 0) {
            return;
        }
        try {
            addon.fit();
            sendResize();
        } catch {
            // Layout transitions can make the host temporarily zero-sized.
        }
    }, [sendResize]);

    const scheduleFit = useCallback((focus = false) => {
        cancelAnimationFrame(fitFrameRef.current);
        fitFrameRef.current = requestAnimationFrame(() => {
            fitFrameRef.current = requestAnimationFrame(() => {
                fitNow();
                if (focus) {
                    terminalRef.current?.focus();
                }
            });
        });
    }, [fitNow]);

    const sendInput = useCallback((content: string) => {
        const currentSocket = socketRef.current;
        if (!content || currentSocket?.readyState !== WebSocket.OPEN) {
            return false;
        }
        try {
            currentSocket.send(socketEncoderRef.current.encode(JSON.stringify({event: "write", content})));
            terminalRef.current?.focus();
            return true;
        } catch {
            return false;
        }
    }, []);

    const paste = useCallback((content: string) => {
        const currentSocket = socketRef.current;
        const terminal = terminalRef.current;
        if (!content || !terminal || currentSocket?.readyState !== WebSocket.OPEN) {
            return false;
        }
        terminal.paste(content);
        terminal.focus();
        return true;
    }, []);

    useImperativeHandle(ref, () => ({
        clear: () => terminalRef.current?.clear(),
        reset: () => terminalRef.current?.reset(),
        write: (data) => terminalRef.current?.write(data),
        focus: () => terminalRef.current?.focus(),
        fit: scheduleFit,
        getSize: () => {
            fitNow();
            return {
                cols: terminalRef.current?.cols || 80,
                rows: terminalRef.current?.rows || 24,
            };
        },
        sendInput,
        paste,
    }), [fitNow, paste, scheduleFit, sendInput]);

    useEffect(() => {
        const host = hostRef.current;
        if (!host) {
            return;
        }
        const terminal = new XTerm({
            cursorBlink: true,
            cursorStyle: "block",
            fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, Liberation Mono, monospace",
            fontSize: 12,
            lineHeight: 1.25,
            logLevel: "off",
            scrollback: 10000,
            theme: themeRef.current,
        });
        const addon = new FitAddon();
        terminal.loadAddon(addon);
        terminal.open(host);
        terminalRef.current = terminal;
        fitAddonRef.current = addon;

        const input = terminal.onData(sendInput);
        const observer = new ResizeObserver(() => scheduleFit());
        observer.observe(host);
        scheduleFit();

        return () => {
            cancelAnimationFrame(fitFrameRef.current);
            observer.disconnect();
            input.dispose();
            terminal.dispose();
            terminalRef.current = undefined;
            fitAddonRef.current = undefined;
        };
    }, []);

    useEffect(() => {
        const terminal = terminalRef.current;
        if (!terminal) {
            return;
        }
        terminal.options.theme = terminalTheme;
        terminal.options.fontSize = 12;
        scheduleFit();
    }, [compact, scheduleFit, terminalTheme]);

    useEffect(() => {
        lastSizeRef.current = "";
        if (!socket) {
            return;
        }
        const handleOpen = () => {
            if (socketRef.current === socket) {
                scheduleFit(active);
            }
        };
        if (socket.readyState === WebSocket.OPEN) {
            handleOpen();
        } else {
            socket.addEventListener("open", handleOpen);
        }
        return () => socket.removeEventListener("open", handleOpen);
    }, [active, scheduleFit, socket]);

    useEffect(() => {
        if (active) {
            scheduleFit(true);
        }
    }, [active, scheduleFit]);

    return (
        <div
            ref={hostRef}
            className={[styles.terminal, className].filter(Boolean).join(" ")}
            style={style}/>
    );
});

Terminal.displayName = "Terminal";

export default memo(Terminal);
