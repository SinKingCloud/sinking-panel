import {memo, type ReactNode, type Ref} from "react";
import {Grid, Spin} from "antd";
import {createStyles} from "antd-style";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";

interface TerminalModalProps {
    title: string;
    open: boolean;
    onClose: () => void;
    destroyOnHidden?: boolean;
    loading?: boolean;
    closable?: boolean;
    toolbarRef?: Ref<HTMLDivElement>;
    children?: ReactNode;
}

export const terminalBackground = "rgb(15, 15, 15)";

const useStyles = createStyles(({css, token}: any, {compact}: {compact: boolean}) => {
    const headerPadding = compact ? "9px 14px" : "10px 16px";
    const desktopHeight = compact ? "min(560px, calc(100dvh - 140px))" : "min(640px, calc(100dvh - 170px))";
    const mobileHeight = compact ? "calc(100dvh - 88px)" : "calc(100dvh - 100px)";

    return {
        modal: css`
            .ant-modal {
                max-width: calc(100vw - 24px);
            }

            .ant-modal-container {
                padding: 0;
                overflow: hidden;
                background: ${token.colorBgContainer};
            }

            .ant-modal-header {
                margin: 0;
                padding: ${headerPadding};
                border-bottom: 1px solid ${token.colorSplit};
                background: ${token.colorBgContainer};
            }

            .ant-modal-title {
                min-width: 0;
            }
        `,
        titleBar: css`
            width: 100%;
            min-width: 0;
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 12px;
        `,
        title: css`
            flex: 1;
            min-width: 0;
            overflow: hidden;

            > * {
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }
        `,
        toolbar: css`
            display: flex;
            flex: none;
            align-items: center;
            gap: 2px;

            .status-text {
                margin-right: 5px;
                color: ${token.colorTextTertiary};
                font-size: 11px;
                font-weight: normal;
                white-space: nowrap;
            }

            .status-text.connected {
                color: ${token.colorSuccessText};
            }

            .status-text.connecting {
                color: ${token.colorWarningText};
            }

            .status-text.error {
                color: ${token.colorErrorText};
            }

            .ant-btn {
                width: ${compact ? 30 : 32}px;
                height: ${compact ? 30 : 32}px;
                padding: 0;
            }

            @media (max-width: 479px) {
                .status-text {
                    display: none;
                }
            }
        `,
        closeButton: css`
            width: 30px;
            height: 30px;
            flex: none;
            padding: 0;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            border: 0;
            border-radius: ${token.borderRadiusSM}px;
            background: transparent;
            color: ${token.colorTextTertiary};
            cursor: pointer;

            &:hover,
            &:focus-visible {
                background: ${token.colorFillTertiary};
                color: ${token.colorText};
            }

            &:focus-visible {
                outline: 2px solid ${token.colorPrimaryBg};
                outline-offset: 1px;
            }
        `,
        host: css`
            height: ${desktopHeight};
            padding: 0;
            min-width: 0;
            min-height: 0;
            display: flex;
            flex-direction: column;
            overflow: hidden;
            background: ${terminalBackground};
            box-sizing: border-box;

            > * {
                min-height: 0;
                flex: 1;
                overflow: hidden;
                border-radius: 0;
            }

            @media (max-width: 767px) {
                height: ${mobileHeight};
                padding: 0;
            }
        `,
        loading: css`
            width: 100%;
            height: 100%;
            display: flex;
            align-items: center;
            justify-content: center;
            background: ${terminalBackground};
        `,
    };
});

const TerminalModal = ({
    title, open, onClose, destroyOnHidden = false, loading = false, closable = true, toolbarRef, children,
}: TerminalModalProps) => {
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact});

    return <ProModal
        title={(
            <div className={styles.titleBar}>
                <div className={styles.title}><Title size="small">{title}</Title></div>
                {toolbarRef && <div ref={toolbarRef} className={styles.toolbar}/>}
                {closable && <button
                    className={styles.closeButton}
                    type="button"
                    aria-label={`关闭${title}`}
                    onClick={onClose}>
                    <Icon type="CloseOutlined"/>
                </button>}
            </div>
        )}
        width={compact ? "min(960px, calc(100vw - 24px))" : "min(1040px, calc(100vw - 24px))"}
        onCancel={onClose}
        modalProps={{
            open,
            rootClassName: styles.modal,
            destroyOnHidden,
            style: {
                top: screens.md ? 100 : 24,
                paddingBottom: screens.md ? 100 : 24,
            },
            closable: false,
            footer: null,
            mask: {closable: true},
            styles: {
                container: {padding: 0},
                body: {padding: 0},
            },
        }}>
        <div className={styles.host}>
            {loading ? <div className={styles.loading}><Spin size="small"/></div> : children}
        </div>
    </ProModal>;
};

export default memo(TerminalModal);
