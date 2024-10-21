import ShareIcon from "@mui/icons-material/Share";
import DeleteIcon from "@mui/icons-material/Delete";
import {Card} from "@mui/material";

const ShareCard = (props) => {
    let frontends = [];
    props.share.frontendEndpoint.map((fe) => {
        frontends.push(<a href={fe.toString()} target={"_"}>{fe}</a>);
    })

    const releaseClicked = () => {
        props.releaseShare({token: props.share.token}, (err, data) => { console.log("releaseClicked", data); });
    }

    return (
        <Card>
            <h2>{props.share.token} [<ShareIcon />]</h2>
            <p>({props.share.shareMode}, {props.share.backendMode})</p>
            <p>
                {props.share.backendEndpoint} &rarr; {frontends}
            </p>
            <p><DeleteIcon onClick={releaseClicked}/></p>
        </Card>
    );
}

export default ShareCard;