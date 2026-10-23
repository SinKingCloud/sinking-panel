import DatabaseManager from "../shared/manager";
import type {DatabaseManagerProps} from "../shared/entities";
import Panel from "./panel";
import Settings from "./settings";

export type {DatabaseManagerProps} from "../shared/entities";

export default (props: DatabaseManagerProps) => <DatabaseManager {...props} type={7} panel={Panel} editForm={Settings}/>;
