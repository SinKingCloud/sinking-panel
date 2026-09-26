import {memo, useState} from "react";
import {useTheme} from "sinking-antd";
import TerminalSession from "@/pages/components/terminal/session";
import TerminalModal, {terminalBackground} from "@/pages/components/terminal/modal";

interface ContainerTerminalProps {
    id: string;
    title?: string;
    open: boolean;
    onClose: () => void;
}

// 与文件终端共用弹窗布局，容器工具栏放在标题右侧。
const ContainerTerminal = ({id, title = "容器终端", open, onClose}: ContainerTerminalProps) => {
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const dark = Boolean(theme?.isDarkMode?.() || theme?.isDarkTheme?.());
    const [toolbarContainer, setToolbarContainer] = useState<HTMLDivElement | null>(null);

    return <TerminalModal title={title} open={open && Boolean(id)} onClose={onClose}
        destroyOnHidden closable={!toolbarContainer} toolbarRef={setToolbarContainer}>
        {open && id && <TerminalSession
            key={id}
            path="/container/terminal"
            params={{id}}
            title={title}
            compact={compact}
            showHeader={false}
            toolbarContainer={toolbarContainer}
            showStatus={false}
            showConnect={false}
            showClear
            showReconnect
            showFullscreen
            showClose
            terminalBackground={terminalBackground}
            terminalAccent={dark ? "#c3ccd6" : undefined}
            acknowledge
            autoConnect
            onClose={onClose}/>}
    </TerminalModal>;
};

export default memo(ContainerTerminal);
