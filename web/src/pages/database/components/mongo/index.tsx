import DatabaseManager from "../shared/manager";
import type {DatabaseManagerProps} from "../shared/entities";
import Panel from "./panel";

export type {DatabaseManagerProps} from "../shared/entities";

export default (props: DatabaseManagerProps) => <DatabaseManager {...props} type={4} panel={Panel}/>;
