import {useState} from "react";
import {getAccountInfo} from "@/service/api/system";

export default () => {
    const [web, setWeb] = useState<API.UserInfo>();

    /**
     * 获取网站用户信息
     */
    const getWebUser = async () => {
        const resp = await getAccountInfo();
        if (resp?.code != 200 || !resp.data) {
            return undefined;
        }
        return resp.data;
    }

    /**
     * 刷新网站用户信息
     */
    const refreshWebUser = (callback?: (data?: API.UserInfo) => void) => {
        getWebUser().then((data) => {
            setWeb(data);
            callback?.(data);
        });
    }

    return {
        web,
        setWeb,
        getWebUser,
        refreshWebUser,
    };
};
