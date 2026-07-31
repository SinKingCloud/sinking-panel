import {useLocation, useModel, useSelectedRoutes} from "umi";
import defaultSettings from "../../../../config/defaultSettings";
import {useEffect} from "react";

const Title = () => {
    const web = useModel("web");//网站信息
    const match = useSelectedRoutes();
    const location = useLocation();

    const initWeb = () => {
        const route = match?.at(-1)?.route as any;
        if (route && route?.title) {
            document.title = route?.title + " - " + (web?.info?.title || defaultSettings?.title);
        }
    }

    useEffect(() => {
        initWeb();
    }, [web?.info?.title, location]);

    return undefined;
};

export default Title;
