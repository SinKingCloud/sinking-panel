import {memo, useEffect, useState} from "react";
import {Icon} from "sinking-antd";
import {getSoftwareIcon} from "@/service/api/software";
import useStyles from "../styles";

interface SoftwareIconProps {
    code?: string;
    icon?: string;
    updateTime?: string | number;
}

const SoftwareIcon = memo(function SoftwareIcon({code, icon, updateTime}: SoftwareIconProps) {
    const {styles} = useStyles();
    const [source, setSource] = useState("");
    useEffect(() => {
        setSource("");
        if (!code || !icon) return;
        const controller = new AbortController();
        let objectUrl = "";
        void getSoftwareIcon(code, icon, controller.signal).then((blob) => {
            if (controller.signal.aborted) return;
            objectUrl = URL.createObjectURL(blob);
            setSource(objectUrl);
        }).catch(() => {});
        return () => {
            controller.abort();
            if (objectUrl) URL.revokeObjectURL(objectUrl);
        };
    }, [code, icon, updateTime]);
    return <span className={styles.icon} aria-hidden="true">
        {source ? <img src={source} alt="" onError={() => setSource("")}/> : <Icon type="AppstoreOutlined"/>}
    </span>;
});

export default SoftwareIcon;
