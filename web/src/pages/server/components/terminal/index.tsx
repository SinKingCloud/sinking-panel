import {forwardRef, memo, useState} from "react";
import TerminalSession, {type ConnectionStatus, type TerminalRef} from "@/pages/components/terminal/session";

export type {ConnectionStatus, TerminalRef} from "@/pages/components/terminal/session";

// SSH 仅提供服务器信息与认证编辑入口，连接及交互界面由通用终端负责。
const Terminal = forwardRef<TerminalRef, any>(({
    server, initializing, unavailable, resetKey, onStatusChange, onEdit, ...props
}, ref) => {
    const [status, setStatus] = useState<ConnectionStatus>("idle");
    const connected = status === "connected" || status === "connecting";
    const localUnavailable = server.id === 0 && unavailable && !connected;
    const needsConfiguration = !localUnavailable && server.id === 0 && (!server.ip || !server.user);
    const disabledReason = initializing ? "正在加载本机连接"
        : localUnavailable ? "本机连接加载失败"
            : needsConfiguration ? "请先配置本机连接" : undefined;
    const endpoint = initializing ? "正在加载本机连接..."
        : localUnavailable ? "本机连接加载失败"
            : server.ip && server.user ? `${server.user}@${server.ip}:${server.port}`
                : server.id === 0 ? "尚未配置本机 SSH" : `${server.ip}:${server.port}`;

    return <TerminalSession
        {...props}
        ref={ref}
        path="/server/ssh"
        params={{id: server.id}}
        binaryControl
        title={server.name || "终端"}
        endpoint={endpoint}
        loading={initializing}
        disabledReason={disabledReason}
        disabledIcon={localUnavailable ? "WarningOutlined" : "SettingOutlined"}
        resetKey={JSON.stringify([resetKey, server.auth_type, server.ip, server.port, server.user])}
        onStatusChange={(nextStatus: ConnectionStatus) => {
            setStatus(nextStatus);
            onStatusChange?.(server.id, nextStatus);
        }}
        errorAction={onEdit ? {
            label: "修改信息",
            codes: ["credential_invalid", "password_invalid", "private_key_invalid"],
            onClick: () => onEdit(server),
        } : undefined}/>
});

Terminal.displayName = "Terminal";
export default memo(Terminal);
