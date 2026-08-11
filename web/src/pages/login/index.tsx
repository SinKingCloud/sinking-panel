import React, {useRef, useState} from "react";
import {App, Button, Form, Input} from "antd";
import {Body, Icon, useTheme} from "sinking-antd";
import {useModel} from "umi";
import Captcha, {CaptchaRef} from "@/components/captcha";
import Settings from "@/../config/defaultSettings";
import {login} from "@/service/auth/login";
import {deleteHeader, loginDevice, setLoginToken} from "@/utils/auth";
import {historyPush} from "@/utils/route";
import useStyles from "./styles";

export default (): React.ReactNode => {
    const {message} = App.useApp();
    const captcha = useRef<CaptchaRef>({});
    const submitting = useRef(false);
    const [loading, setLoading] = useState(false);
    const web = useModel("web");
    const user = useModel("user");
    const theme = useTheme();
    const isCompactMode = theme?.isCompactTheme?.() || false;
    const isDarkMode = Boolean(theme?.isDarkMode?.() || theme?.isDarkTheme?.());
    const {styles} = useStyles({isCompactMode, isDarkMode});
    const name = web?.info?.name || Settings?.name || Settings?.title;

    const finish = () => {
        submitting.current = false;
        setLoading(false);
    };

    const submit = (values: any) => {
        if (submitting.current) {
            return;
        }
        submitting.current = true;
        setLoading(true);
        if (!captcha.current?.Show) {
            finish();
            return;
        }
        captcha.current.Show(async (result) => {
            try {
                const response = await login({
                    body: {
                        account: values.account,
                        password: values.password,
                        device: loginDevice,
                        token: result.token,
                        captcha_x: result.x,
                        captcha_y: result.y,
                    },
                });
                if (response?.code != 200 || !response.data) {
                    message.error(response?.message || "登录失败");
                    return;
                }
                setLoginToken(response.data);
                const data = await user?.getWebUser();
                if (!data) {
                    deleteHeader();
                    message.error("账户信息加载失败，请重试");
                    return;
                }
                user?.setWeb(data);
                message.success(response?.message || "登录成功");
                historyPush("index");
            } finally {
                finish();
            }
        }, finish);
    };

    return (
        <Body className={styles.body}>
            <main className={styles.screen}>
                <Captcha ref={captcha}/>
                <div className={styles.backdrop} aria-hidden="true"/>
                <section className={styles.loginPanel}>
                    <div className="brand">
                        <Icon type="icon-logo"/>
                        <h1 title={name}>{name}</h1>
                    </div>
                    <p className="product">服务器管理面板</p>
                    <Form className={styles.form} layout="vertical" requiredMark={false} onFinish={submit}>
                        <Form.Item
                            className="field-item"
                            name="account"
                            rules={[{required: true, message: "请输入登录账号"}]}
                        >
                            <Input
                                className={styles.input}
                                prefix={<Icon type="UserOutlined"/>}
                                aria-label="登录账号"
                                placeholder="登录账号"
                                autoComplete="username"
                            />
                        </Form.Item>
                        <Form.Item
                            className="field-item"
                            name="password"
                            rules={[{required: true, message: "请输入登录密码"}]}
                        >
                            <Input.Password
                                className={styles.input}
                                prefix={<Icon type="LockOutlined"/>}
                                aria-label="登录密码"
                                placeholder="登录密码"
                                autoComplete="current-password"
                            />
                        </Form.Item>
                        <Form.Item className="submit-item">
                            <Button className={styles.submit} type="primary" htmlType="submit" loading={loading} block>
                                登录
                            </Button>
                        </Form.Item>
                    </Form>
                </section>
            </main>
        </Body>
    );
};
