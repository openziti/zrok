import AccountDetail from "./AccountDetail";
import ShareDetail from "./ShareDetail";
import EnvironmentDetail from "./environment/EnvironmentDetail";

const Detail = (props) => {
    let detailComponent = <h1>{props.selection.id} ({props.selection.type})</h1>;

    switch(props.selection.type) {
        case "account":
            detailComponent = <AccountDetail user={props.user} />;
            break;

        case "environment":
            detailComponent = <EnvironmentDetail selection={props.selection} />;
            break;

        case "share":
            detailComponent = <ShareDetail selection={props.selection} />;
    }

    return (
        <div className={"detail-container"}>{detailComponent}</div>
    );
};

export default Detail;