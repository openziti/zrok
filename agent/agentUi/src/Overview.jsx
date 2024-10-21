import "bootstrap/dist/css/bootstrap.min.css";
import ShareCard from "./ShareCard.jsx";
import AccessCard from "./AccessCard.jsx";
import LanIcon from "@mui/icons-material/Lan";
import ShareIcon from "@mui/icons-material/Share";
import {Card} from "@mui/material";

const Overview = (props) => {
    let cards = [];
    if(props.overview.size > 0) {
        props.overview.forEach((row) => {
            switch(row.type) {
                case "share":
                    cards.push(<ShareCard releaseShare={props.releaseShare} share={row.v} />);
                    break;

                case "access":
                    cards.push(<AccessCard releaseAccess={props.releaseAccess} access={row.v} />);
                    break;
            }
        });
    } else {
        cards.push(<Card key="empty"><h5>zrok Agent is empty! Add a <a href="#" onClick={props.shareClick}>share <ShareIcon /></a> or <a href={"#"} onClick={props.accessClick}>access <LanIcon /></a> share to get started.</h5></Card>);
    }

    return <>{cards}</>;
}

export default Overview;
