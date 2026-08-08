import {App, Button, Form, Input} from 'antd';
import React, {useRef, useState} from 'react';
import {Body, Icon} from 'sinking-antd';
import {useModel} from "umi";
import {login} from "@/service/auth/login";
import {loginDevice, setLoginToken} from "@/utils/auth";
import Captcha, {CaptchaRef} from "@/components/captcha";
import {createStyles} from "antd-style";
import Settings from "@/../config/defaultSettings";
import {historyPush} from "@/utils/route";

const useStyles = createStyles(({css, responsive, token}): any => {
    return {
        container: {
            display: "flex",
            flexDirection: "column",
            height: "100vh",
            backgroundImage: "url('https://gw.alipayobjects.com/zos/rmsportal/TVYTbAXWheQpRcWDaDMu.svg')",
            backgroundRepeat: "no-repeat",
            backgroundPosition: "center 110px",
            backgroundSize: "100%",
        },
        content: {
            flex: 1,
            padding: "120px 0 32px 0"
        },
        main: css`
            width: 328px;
            margin: 0 auto;

            ${responsive.md} {
                width: 95%;
                max-width: 300px;
            }
        `,
        top: {
            textAlign: "center"
        },
        header: {
            height: 40,
            lineHeight: "40px",
            a: {
                textDecoration: "none"
            },
            span: {
                fontSize: "30px",
                fontWeight: "bolder"
            }
        },
        logo: {
            height: 40,
            marginRight: 16,
            verticalAlign: "top",
            fontSize: 27,
            color: token?.colorPrimary,
        },
        desc: {
            marginTop: 12,
            marginBottom: 20,
            color: "@text-color-secondary",
            fontSize: 14,
        },
    };
});

const Login: React.FC = () => {
    const {
        styles: {
            container, content, top, header, logo, desc, main
        }
    } = useStyles();
    const {message} = App.useApp();
    const captcha = useRef<CaptchaRef>({});
    const [isLoading, setIsLoading] = useState(false);
    /**
     * 表单
     */
    const [form] = Form.useForm<{ account: string; password: string }>();
    /**
     * 获取当前用户信息
     */
    const user = useModel("user");
    const web = useModel("web");

    return (
        <Body>
            <div className={container}>
                <Captcha ref={captcha}/>
                <div className={content}>
                    <div className={top}>
                        <div className={header}>
                            <Icon type={"icon-logo"}
                                  className={logo}/>
                            <span>{web?.info?.name || Settings?.title}</span>
                        </div>
                        <div className={desc}>
                            服务器管理面板
                        </div>
                    </div>
                    <div className={main}>
                        <Form form={form} size="large" onFinish={async (values) => {
                            captcha?.current?.Show?.(
                                async (res) => {
                                    setIsLoading(true);
                                    await login({
                                        body: {
                                            account: values.account,
                                            password: values.password,
                                            device: loginDevice,
                                            token: res.token,
                                            captcha_x: res.x,
                                            captcha_y: res.y
                                        },
                                        onSuccess: (r) => {
                                            setLoginToken(r.data);
                                            user?.refreshWebUser(() => {
                                                message?.success(r?.message);
                                                historyPush("index");
                                            });
                                        },
                                        onFail: (r) => {
                                            message?.error(r?.message || "登录失败")
                                        },
                                        onFinally: () => {
                                            setIsLoading(false);
                                        }
                                    });
                                }
                            );
                        }}>
                            <Form.Item name='account' rules={[{required: true, message: '请输入账户'}]}>
                                <Input prefix={<Icon type="UserOutlined" className='site-form-item-icon'/>}
                                       placeholder='请输入账户' size={'large'}/>
                            </Form.Item>
                            <Form.Item name='password' rules={[{required: true, message: '请输入账户密码'}]}>
                                <Input.Password prefix={<Icon type="LockOutlined" className='site-form-item-icon'/>}
                                                size={'large'}
                                                placeholder='请输入账户密码'/>
                            </Form.Item>
                            <Form.Item>
                                <Button type='primary' loading={isLoading} htmlType='submit' size={'large'} block>
                                    登 录
                                </Button>
                            </Form.Item>
                        </Form>
                    </div>
                </div>
            </div>
        </Body>
    );
};

export default Login;
