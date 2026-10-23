import DatabaseManager from "../shared/manager";
import Panel from "./panel";
import type {DatabaseManagerProps} from "../shared/entities";

export type {DatabaseManagerProps} from "../shared/entities";

const engine = {
    query: "SELECT @@VERSION AS version;",
    objects: [{key: "view", label: "视图", icon: "EyeOutlined"}, {key: "routine", label: "过程与函数", icon: "FunctionOutlined"}, {key: "trigger", label: "触发器", icon: "ThunderboltOutlined"}],
};

export default (props: DatabaseManagerProps) => <DatabaseManager {...props} type={5} panel={Panel} engine={engine}/>;
