import AccountDetail from "./AccountDetail";
import ShareDetail from "./ShareDetail";
import Environment from "./environment/Environment";

const Detail = (props) => {
    let detailComponent = <h1>{props.selection.id} ({props.selection.type})</h1>;

    switch(props.selection.type) {
        case "account":
            detailComponent = <AccountDetail user={props.user} />;
            break;

        case "environment":
            detailComponent = <Environment selection={props.selection} />;
            break;

        case "service":
            detailComponent = <ShareDetail selection={props.selection} />;
    }

    return (
        <div className={"detail-container"}>{detailComponent}</div>
    );
};

export default Detail;