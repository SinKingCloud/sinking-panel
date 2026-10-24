import DatabaseManager from "../shared/manager";
import Panel from "./panel";
import type {DatabaseManagerProps} from "../shared/entities";

export type {DatabaseManagerProps} from "../shared/entities";

const engine = {
    module: "postgres",
    query: "SELECT version();",
    objects: [{key: "view", label: "视图与物化视图", icon: "EyeOutlined"}, {key: "routine", label: "函数与过程", icon: "FunctionOutlined"}, {key: "trigger", label: "触发器", icon: "ThunderboltOutlined"}],
};

export default (props: DatabaseManagerProps) => <DatabaseManager {...props} type={3} panel={Panel} engine={engine}/>;
