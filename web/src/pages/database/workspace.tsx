import {useRef} from "react";
import {Redis, Mysql, Postgres, Mongo, SqlServer, Sqlite, Elasticsearch} from "./components";
import ActionDropdown from "@/pages/components/action-dropdown";
import FilePicker from "@/pages/components/file-picker";
import AceEditor from "@/components/ace-editor";
import {databaseRequests} from "@/service/api/database";

const components = {ActionDropdown, FilePicker, Editor: AceEditor};
const managers: Record<number, any> = {1: Redis, 2: Mysql, 3: Postgres, 4: Mongo, 5: SqlServer, 6: Sqlite, 7: Elasticsearch};

export default (props: any) => {
    // 关闭时保留当前管理器，让原窗口完成退场动画再清理自身状态。
    const type = useRef(props.record?.type);
    if (props.record) type.current = props.record.type;
    const Manager = managers[Number(type.current)];
    return Manager ? <Manager {...props} requests={databaseRequests[Number(type.current)]} components={components}/> : null;
};
