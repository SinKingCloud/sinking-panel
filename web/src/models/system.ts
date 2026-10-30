import {useCallback, useState} from "react";
import {getSystemInfo} from "@/service/api/system";

export default () => {
    const [info, setInfo] = useState<any>();

    /** 获取系统信息。 */
    const getInfo = useCallback(async () => {
        const resp = await getSystemInfo();
        if (resp?.code !== 200 || !resp?.data) {
            return undefined;
        }
        return resp.data;
    }, []);

    /** 刷新全局系统信息。 */
    const refreshInfo = useCallback(async () => {
        const data = await getInfo();
        setInfo(data);
        return data;
    }, [getInfo]);

    return {
        info,
        setInfo,
        getInfo,
        refreshInfo,
    };
};
