import {forwardRef, useState, useImperativeHandle, useRef} from "react";
import {Spin, App} from "antd";
import {ProModal} from "sinking-antd";
import GoCaptcha from "go-captcha-react";
import {getCaptcha} from "@/service/common/captcha";
import {getRandStr} from "@/utils/string";
import {createStyles} from "antd-style";

/**
 * 验证码组件接口
 */
export interface CaptchaRef {
    Show?: (onSuccess?: (res: any) => void, onClose?: () => void) => void;
}

/**
 * 验证码响应数据接口
 */
interface CaptchaResponse {
    token: string;
    x: number;
    y: number;
}

/**
 * 样式配置
 */
const useStyles: any = createStyles(({isDarkMode, token}) => {
    return {
        modal: {
            ".ant-modal": {
                width: "min(326px, calc(100vw - 24px)) !important",
                maxWidth: "calc(100vw - 24px)",
            },
            ".ant-modal-container": {
                padding: "0 !important",
                backgroundColor: "transparent !important",
                boxShadow: "none !important",
            },
            ".ant-modal-body> div": isDarkMode ? {
                "--go-captcha-theme-text-color": "#8f94a7 !important",
                "--go-captcha-theme-bg-color": "#18181a !important",
                "--go-captcha-theme-btn-color": "#ffffff !important",
                "--go-captcha-theme-btn-bg-color": token?.colorPrimary + " !important",
                "--go-captcha-theme-btn-border-color": token?.colorPrimary + " !important",
                "--go-captcha-theme-active-color": token?.colorPrimary + " !important",
                "--go-captcha-theme-border-color": "#3c3f44 !important",
                "--go-captcha-theme-icon-color": "#696d7b !important",
                "--go-captcha-theme-drag-bar-color": "#3c3f44 !important",
                "--go-captcha-theme-drag-bg-color": token?.colorPrimary + " !important",
                "--go-captcha-theme-drag-icon-color": "#ffffff !important",
                "--go-captcha-theme-round-color": "#3c3f44 !important",
                "--go-captcha-theme-loading-icon-color": token?.colorPrimary + " !important",
                "--go-captcha-theme-body-bg-color": "#34383e !important",
                "--go-captcha-theme-dot-color": "#cedffe !important",
                "--go-captcha-theme-dot-bg-color": token?.colorPrimary + " !important",
                "--go-captcha-theme-dot-border-color": "#f7f9fb !important",
            } : {
                "--go-captcha-theme-btn-color": token?.colorPrimary + " !important",
                "--go-captcha-theme-btn-bg-color": token?.colorPrimary + " !important",
                "--go-captcha-theme-drag-bg-color": token?.colorPrimary + " !important",
            },
        },
    };
});

const Captcha = forwardRef<CaptchaRef>((_, ref): any => {
    const {styles} = useStyles();
    const {message} = App.useApp();

    const slideRef = useRef<any>(null);
    const [visible, setVisible] = useState(false);
    const [loading, setLoading] = useState(true);
    const [data, setData] = useState<any>(null);

    /**
     * 回调函数引用
     */
    const token = useRef<string>("");
    const request = useRef(0);
    const successCallback = useRef<((res: CaptchaResponse) => void) | null>(null);
    const closeCallback = useRef<(() => void) | null>(null);

    /**
     * 关闭验证码
     */
    const close = (notify = true) => {
        request.current += 1;
        setVisible(false);
        setData(null);
        setLoading(true);
        if (notify && closeCallback.current) {
            closeCallback.current();
        }
        successCallback.current = null;
        closeCallback.current = null;
    };

    /**
     * 刷新验证码
     */
    const refresh = async () => {
        const requestId = ++request.current;
        const currentToken = getRandStr(16);
        try {
            setLoading(true);
            token.current = currentToken;
            await getCaptcha({
                body: {token: currentToken},
                onSuccess: (res) => {
                    if (requestId !== request.current) {
                        return;
                    }
                    const d = res?.data;
                    if (d) {
                        setData({
                            key: d.key,
                            image: d.image_base64,
                            thumb: d.tile_base64,
                            thumbWidth: d.tile_width,
                            thumbHeight: d.tile_height,
                            tileX: d.tile_x,
                            tileY: d.tile_y,
                            width: d.width,
                            height: d.height,
                        });
                    }
                },
                onFail: (error) => {
                    if (requestId !== request.current) {
                        return;
                    }
                    message.error(error?.message || "验证码加载失败");
                    close();
                },
                onFinally: () => {
                    if (requestId === request.current) {
                        setLoading(false);
                    }
                }
            });
        } catch (error) {
            if (requestId !== request.current) {
                return;
            }
            message.error("验证码加载失败");
            setLoading(false);
            close();
        }
    };

    /**
     * 显示验证码
     * @param onSuccess 成功回调函数
     * @param onClose 关闭回调函数
     */
    const show = async (onSuccess?: (res: CaptchaResponse) => void, onClose?: () => void) => {
        // 设置回调函数
        successCallback.current = onSuccess || null;
        closeCallback.current = onClose || null;
        // 显示模态框并刷新验证码
        setVisible(true);
        await refresh();
    };

    /**
     * 方法挂载
     */
    useImperativeHandle(ref, () => ({
        Show: show,
    }));

    return (
        <ProModal
            onCancel={() => close()}
            modalProps={{
                open: visible,
                destroyOnHidden: true,
                footer: null,
                closable: false,
                mask: {closable: false},
                keyboard: false,
                rootClassName: styles?.modal,
            }}>
            <Spin spinning={loading} size={"large"}>
                <GoCaptcha.Slide
                    ref={slideRef}
                    data={{
                        image: data?.image || "",
                        thumb: data?.thumb || "",
                        thumbX: 0,
                        thumbY: data?.tileY || 0,
                        thumbWidth: data?.thumbWidth || 0,
                        thumbHeight: data?.thumbHeight || 0,
                    }}
                    config={{
                        width: data?.width,
                        height: data?.height,
                        scope: true,
                    }}
                    events={{
                        confirm: (point) => {
                            const result: CaptchaResponse = {
                                token: token.current,
                                x: point?.x || 0,
                                y: point?.y || 0
                            };
                            if (successCallback.current) {
                                successCallback.current(result);
                            }
                            close(false);
                        },
                        refresh: refresh,
                        close: () => close(),
                    }}
                />
            </Spin>
        </ProModal>
    );
});

export default Captcha;
