import SqlHost from "../shared/sql/host";
import Accounts from "./accounts";
import type {SqlHostProps} from "../shared/sql/entities";

export default (props: SqlHostProps) => <SqlHost {...props} accounts={Accounts}/>;
