import EnvironmentDetail from "./EnvironmentDetail";
import AccountDetail from "./AccountDetail";

const Detail = (props) => {
    let detailComponent = <h1>{props.selection.id} ({props.selection.type})</h1>;

    switch(props.selection.type) {
        case "account":
            detailComponent = <AccountDetail user={props.user} />;
            break;

        case "environment":
            detailComponent = <EnvironmentDetail selection={props.selection} />;
            break;
    }

    return (
        <div className={"detail-container"}>{detailComponent}</div>
    );
};

export default Detail;