import {useEffect, useState} from "react";
import {getWebInfo} from "@/service/auth/info";
import defaultSettings from "../../config/defaultSettings";

export default () => {
    const defaults: API.WebInfo = {
        title: defaultSettings.title,
        name: defaultSettings.name,
        ui: {
            layout: "left",
            watermark: false,
            theme: "dark",
            compact: false,
            color: defaultSettings.color,
            radius: defaultSettings.radius,
        },
    };
    const [info, setInfo] = useState<API.WebInfo>();
    /**
     * 获取站点信息
     */
    const getInfo = async () => {
        const resp = await getWebInfo();
        if (resp?.code != 200 || !resp.data) {
            return undefined;
        }
        return {
            ...defaults,
            ...resp.data,
            ui: {
                ...defaults.ui,
                ...resp.data.ui,
            },
        };
    }
    /**
     * 刷新站点信息
     */
    const refreshInfo = () => {
        setInfo(undefined);
        getInfo().then((d) => {
            if (d) {
                setInfo(d);
            }
        });
    }

    useEffect(() => {
        refreshInfo();
    }, []);

    return {
        info,
        setInfo,
        getInfo,
        refreshInfo
    };
};
