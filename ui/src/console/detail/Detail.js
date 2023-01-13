import AccountDetail from "./account/AccountDetail";
import ShareDetail from "./share/ShareDetail";
import EnvironmentDetail from "./environment/EnvironmentDetail";

const Detail = (props) => {
    let detailComponent = <h1>{props.selection.id} ({props.selection.type})</h1>;

    switch(props.selection.type) {
        case "environment":
            detailComponent = <EnvironmentDetail selection={props.selection} />;
            break;

        case "share":
            detailComponent = <ShareDetail selection={props.selection} />;
            break;

        default:
            detailComponent = <AccountDetail user={props.user} />;
            break;
    }

    return (
        <div className={"detail-container"}>{detailComponent}</div>
    );
};

export default Detail;