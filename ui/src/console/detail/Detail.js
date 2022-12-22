import EnvironmentDetail from "./EnvironmentDetail";
import AccountDetail from "./AccountDetail";
import ServiceDetail from "./ServiceDetail";

const Detail = (props) => {
    let detailComponent = <h1>{props.selection.id} ({props.selection.type})</h1>;

    switch(props.selection.type) {
        case "account":
            detailComponent = <AccountDetail user={props.user} />;
            break;

        case "environment":
            detailComponent = <EnvironmentDetail selection={props.selection} />;
            break;

        case "service":
            detailComponent = <ServiceDetail selection={props.selection} />;
    }

    return (
        <div className={"detail-container"}>{detailComponent}</div>
    );
};

export default Detail;